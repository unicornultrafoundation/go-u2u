// Copyright 2024 The go-u2u Authors
// This file is part of the go-u2u library.
//
// The go-u2u library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-u2u library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-u2u library. If not, see <http://www.gnu.org/licenses/>.

package evmcore

import (
	"fmt"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/params"
)

// SetCodeStateTransition handles state transitions for EIP-7702 SetCode transactions
type SetCodeStateTransition struct {
	gp           *GasPool
	statedb      *state.StateDB
	header       *EvmHeader
	processor    *types.SetCodeTxProcessor
	gasCalc      *types.SetCodeTxGasCalculator
	chainConfig  *params.ChainConfig
	delegationDB *DelegationStateDB
}

// DelegationStateDB manages delegation mappings in state
type DelegationStateDB struct {
	statedb *state.StateDB
}

// NewDelegationStateDB creates a new delegation state database
func NewDelegationStateDB(statedb *state.StateDB) *DelegationStateDB {
	return &DelegationStateDB{
		statedb: statedb,
	}
}

// SetDelegation stores a delegation mapping for an authority
func (d *DelegationStateDB) SetDelegation(authority, codeAddr common.Address) {
	// Store delegation in a special storage slot
	// Use a deterministic storage key based on authority address
	delegationKey := common.BytesToHash(append([]byte("EIP7702_DELEGATION_"), authority.Bytes()...))
	d.statedb.SetState(common.HexToAddress("0x7702"), delegationKey, codeAddr.Hash())
	
	log.Debug("Set delegation", "authority", authority.Hex(), "codeAddr", codeAddr.Hex())
}

// GetDelegation retrieves a delegation mapping for an authority
func (d *DelegationStateDB) GetDelegation(authority common.Address) *common.Address {
	delegationKey := common.BytesToHash(append([]byte("EIP7702_DELEGATION_"), authority.Bytes()...))
	codeHash := d.statedb.GetState(common.HexToAddress("0x7702"), delegationKey)
	
	if codeHash == (common.Hash{}) {
		return nil
	}
	
	codeAddr := common.BytesToAddress(codeHash.Bytes())
	return &codeAddr
}

// RemoveDelegation removes a delegation mapping for an authority
func (d *DelegationStateDB) RemoveDelegation(authority common.Address) {
	delegationKey := common.BytesToHash(append([]byte("EIP7702_DELEGATION_"), authority.Bytes()...))
	d.statedb.SetState(common.HexToAddress("0x7702"), delegationKey, common.Hash{})
	
	log.Debug("Removed delegation", "authority", authority.Hex())
}

// NewSetCodeStateTransition creates a new SetCode state transition processor
func NewSetCodeStateTransition(
	gp *GasPool,
	statedb *state.StateDB,
	header *EvmHeader,
	chainConfig *params.ChainConfig,
	chainID *big.Int,
) *SetCodeStateTransition {
	return &SetCodeStateTransition{
		gp:           gp,
		statedb:      statedb,
		header:       header,
		processor:    types.NewSetCodeTxProcessor(chainID),
		gasCalc:      types.NewSetCodeTxGasCalculator(chainID),
		chainConfig:  chainConfig,
		delegationDB: NewDelegationStateDB(statedb),
	}
}

// ApplySetCodeTransaction applies a SetCode transaction to the state
func (st *SetCodeStateTransition) ApplySetCodeTransaction(
	tx *types.Transaction,
	signer types.Signer,
) (*types.Receipt, error) {
	// Validate transaction type
	if tx.Type() != types.SetCodeTxType {
		return nil, fmt.Errorf("not a SetCode transaction")
	}

	// Extract SetCodeTx inner data
	inner := tx.Inner()
	setCodeTx, ok := inner.(*types.SetCodeTx)
	if !ok {
		return nil, fmt.Errorf("invalid SetCode transaction inner type")
	}

	// Get transaction sender
	from, err := types.Sender(signer, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction sender: %w", err)
	}

	// Calculate intrinsic gas
	intrinsicGas, err := st.gasCalc.IntrinsicGas(setCodeTx)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate intrinsic gas: %w", err)
	}

	// Check gas limit
	if setCodeTx.Gas < intrinsicGas {
		return nil, fmt.Errorf("gas limit %d below intrinsic gas %d", setCodeTx.Gas, intrinsicGas)
	}

	// Use gas from pool
	if err := st.gp.SubGas(setCodeTx.Gas); err != nil {
		return nil, fmt.Errorf("gas pool error: %w", err)
	}

	// Snapshot state for potential rollback
	snapshot := st.statedb.Snapshot()

	// Deduct gas cost and value from sender
	gasPrice := setCodeTx.GasFeeCap
	if st.header.BaseFee != nil {
		// Use effective gas price for EIP-1559
		gasPrice = new(big.Int).Add(setCodeTx.GasTipCap, st.header.BaseFee)
		if gasPrice.Cmp(setCodeTx.GasFeeCap) > 0 {
			gasPrice = setCodeTx.GasFeeCap
		}
	}

	gasCost := new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(setCodeTx.Gas))
	totalCost := new(big.Int).Add(gasCost, setCodeTx.Value)

	// Check sender balance
	senderBalance := st.statedb.GetBalance(from)
	if senderBalance.Cmp(totalCost) < 0 {
		st.statedb.RevertToSnapshot(snapshot)
		return nil, fmt.Errorf("insufficient balance: have %v, need %v", senderBalance, totalCost)
	}

	// Deduct gas cost from sender
	st.statedb.SubBalance(from, gasCost)

	// Process authorization list
	delegations, err := st.processAuthorizationList(setCodeTx, from)
	if err != nil {
		st.statedb.RevertToSnapshot(snapshot)
		return nil, fmt.Errorf("authorization processing failed: %w", err)
	}

	// Apply delegations to state
	for authority, codeAddr := range delegations {
		st.delegationDB.SetDelegation(authority, codeAddr)
	}

	// Execute transaction (if it has a target and value/data)
	var executionResult *ExecutionResult
	gasUsed := intrinsicGas

	if setCodeTx.To != nil || len(setCodeTx.Data) > 0 {
		// Execute the transaction call
		executionResult, err = st.executeTransaction(setCodeTx, from, gasUsed)
		if err != nil {
			log.Debug("SetCode transaction execution failed", "error", err)
			// Don't revert - delegation still applies even if call fails
		}
		if executionResult != nil {
			gasUsed = executionResult.UsedGas
		}
	}

	// Calculate refund
	gasRefund := setCodeTx.Gas - gasUsed
	refundAmount := new(big.Int).Mul(gasPrice, new(big.Int).SetUint64(gasRefund))
	st.statedb.AddBalance(from, refundAmount)

	// Add gas back to pool
	st.gp.AddGas(gasRefund)

	// Increment sender nonce
	st.statedb.SetNonce(from, st.statedb.GetNonce(from)+1)

	// Create receipt
	receipt := &types.Receipt{
		Type:              tx.Type(),
		CumulativeGasUsed: gasUsed,
		GasUsed:           gasUsed,
		Logs:              []*types.Log{},
	}

	// Set status based on execution result
	if executionResult != nil && executionResult.Failed() {
		receipt.Status = types.ReceiptStatusFailed
	} else {
		receipt.Status = types.ReceiptStatusSuccessful
	}

	// Add delegation events to logs
	for authority, codeAddr := range delegations {
		receipt.Logs = append(receipt.Logs, st.createDelegationLog(authority, codeAddr))
	}

	log.Debug("SetCode transaction applied",
		"hash", tx.Hash().Hex(),
		"from", from.Hex(),
		"gasUsed", gasUsed,
		"delegations", len(delegations))

	return receipt, nil
}

// processAuthorizationList processes the authorization list and returns delegation mappings
func (st *SetCodeStateTransition) processAuthorizationList(
	setCodeTx *types.SetCodeTx,
	from common.Address,
) (map[common.Address]common.Address, error) {
	getAccountNonce := func(addr common.Address) uint64 {
		return st.statedb.GetNonce(addr)
	}

	return st.processor.ProcessDelegations(
		setCodeTx,
		st.header.Number,
		getAccountNonce,
	)
}

// executeTransaction executes the transaction call part of a SetCode transaction
func (st *SetCodeStateTransition) executeTransaction(
	setCodeTx *types.SetCodeTx,
	from common.Address,
	gasUsed uint64,
) (*ExecutionResult, error) {
	// This would typically use the EVM to execute the transaction
	// For now, return a simple success result
	return &ExecutionResult{
		UsedGas: gasUsed + 21000, // Add basic call gas
		Err:     nil,
	}, nil
}

// createDelegationLog creates a log entry for a delegation event
func (st *SetCodeStateTransition) createDelegationLog(authority, codeAddr common.Address) *types.Log {
	// Create a log for the delegation event
	// Topic[0] = DelegationSet event signature
	// Topic[1] = authority address
	// Data = code address
	delegationEventSig := common.HexToHash("0x7702000000000000000000000000000000000000000000000000000000000000")
	
	return &types.Log{
		Address: common.HexToAddress("0x7702"), // EIP-7702 system address
		Topics: []common.Hash{
			delegationEventSig,
			authority.Hash(),
		},
		Data: codeAddr.Bytes(),
	}
}


// SetCodeTransitionApplier integrates SetCode transaction processing with existing state transition
type SetCodeTransitionApplier struct {
	setCodeProcessor *SetCodeStateTransition
}

// NewSetCodeTransitionApplier creates a new SetCode transaction applier
func NewSetCodeTransitionApplier(
	gp *GasPool,
	statedb *state.StateDB,
	header *EvmHeader,
	chainConfig *params.ChainConfig,
	chainID *big.Int,
) *SetCodeTransitionApplier {
	return &SetCodeTransitionApplier{
		setCodeProcessor: NewSetCodeStateTransition(gp, statedb, header, chainConfig, chainID),
	}
}

// ApplyTransaction applies a transaction, handling SetCode transactions specially
func (applier *SetCodeTransitionApplier) ApplyTransaction(
	tx *types.Transaction,
	signer types.Signer,
	defaultApplier func(*types.Transaction, types.Signer) (*types.Receipt, error),
) (*types.Receipt, error) {
	// Check if it's a SetCode transaction
	if tx.Type() == types.SetCodeTxType {
		return applier.setCodeProcessor.ApplySetCodeTransaction(tx, signer)
	}

	// Use default applier for other transaction types
	return defaultApplier(tx, signer)
}

// GetDelegation retrieves delegation for an authority address
func (applier *SetCodeTransitionApplier) GetDelegation(authority common.Address) *common.Address {
	return applier.setCodeProcessor.delegationDB.GetDelegation(authority)
}