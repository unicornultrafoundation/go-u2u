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
)

// SetCodeTxPoolValidator provides validation logic for SetCode transactions in the transaction pool
type SetCodeTxPoolValidator struct {
	chainID   *big.Int
	processor *types.SetCodeTxProcessor
	gasCalc   *types.SetCodeTxGasCalculator
}

// NewSetCodeTxPoolValidator creates a new SetCode transaction pool validator
func NewSetCodeTxPoolValidator(chainID *big.Int) *SetCodeTxPoolValidator {
	return &SetCodeTxPoolValidator{
		chainID:   chainID,
		processor: types.NewSetCodeTxProcessor(chainID),
		gasCalc:   types.NewSetCodeTxGasCalculator(chainID),
	}
}

// ValidateSetCodeTransaction validates a SetCode transaction for pool inclusion
func (v *SetCodeTxPoolValidator) ValidateSetCodeTransaction(
	tx *types.Transaction,
	head *EvmHeader,
	signer types.Signer,
	opts *ValidationOptions,
) error {
	// Ensure transaction type is accepted
	if opts.Accept&(1<<tx.Type()) == 0 {
		return fmt.Errorf("%w: tx type %v not supported by this pool", ErrTxTypeNotSupported, tx.Type())
	}

	// Ensure transaction is actually a SetCode transaction
	if tx.Type() != types.SetCodeTxType {
		return fmt.Errorf("not a SetCode transaction: type %d", tx.Type())
	}

	// Get the inner SetCodeTx
	inner := tx.Inner()
	setCodeTx, ok := inner.(*types.SetCodeTx)
	if !ok {
		return fmt.Errorf("failed to extract SetCode transaction data")
	}

	// Ensure the transaction doesn't exceed the current block limit gas
	if head.GasLimit < tx.Gas() {
		return ErrGasLimit
	}

	// Validate for pool inclusion
	if err := v.processor.ValidateForPool(setCodeTx, v.chainID); err != nil {
		return fmt.Errorf("SetCode transaction pool validation failed: %w", err)
	}

	// Validate gas limit
	if err := v.gasCalc.ValidateGasLimit(setCodeTx); err != nil {
		return fmt.Errorf("SetCode transaction gas validation failed: %w", err)
	}

	// Validate EIP-1559 parameters against current base fee
	if head.BaseFee != nil {
		if setCodeTx.GasFeeCap.Cmp(head.BaseFee) < 0 {
			return fmt.Errorf("gas fee cap %v below base fee %v", setCodeTx.GasFeeCap, head.BaseFee)
		}
	}

	// Validate against minimum tip
	if opts.MinTip != nil && setCodeTx.GasTipCap.Cmp(opts.MinTip) < 0 {
		return fmt.Errorf("gas tip cap %v below minimum tip %v", setCodeTx.GasTipCap, opts.MinTip)
	}

	// Validate authorization list size limits
	if len(setCodeTx.AuthorizationList) > MaxAuthorizationListSize {
		return fmt.Errorf("authorization list too large: %d (max %d)", 
			len(setCodeTx.AuthorizationList), MaxAuthorizationListSize)
	}

	return nil
}

// ValidateSetCodeTransactionWithState validates a SetCode transaction with state context
func (v *SetCodeTxPoolValidator) ValidateSetCodeTransactionWithState(
	tx *types.Transaction,
	statedb *state.StateDB,
	head *EvmHeader,
	signer types.Signer,
	basicOpts *ValidationOptions,
	stateOpts *ValidationOptionsWithState,
) error {
	// First do stateless validation
	if err := v.ValidateSetCodeTransaction(tx, head, signer, basicOpts); err != nil {
		return err
	}

	// Extract SetCodeTx inner data
	inner := tx.Inner()
	setCodeTx, ok := inner.(*types.SetCodeTx)
	if !ok {
		return fmt.Errorf("invalid SetCode transaction inner type")
	}

	// Validate sender balance and nonce
	from, err := types.Sender(signer, tx)
	if err != nil {
		return fmt.Errorf("invalid transaction sender: %w", err)
	}

	// Check nonce
	currentNonce := statedb.GetNonce(from)
	if currentNonce != setCodeTx.Nonce {
		return fmt.Errorf("nonce mismatch: account has %d, transaction has %d", currentNonce, setCodeTx.Nonce)
	}

	// Check balance for transaction cost
	balance := statedb.GetBalance(from)
	cost := tx.Cost()
	if balance.Cmp(cost) < 0 {
		return fmt.Errorf("insufficient balance: have %v, need %v", balance, cost)
	}

	// Validate authorization list with state context
	if err := v.validateAuthorizationListWithState(setCodeTx, statedb, head); err != nil {
		return fmt.Errorf("authorization list state validation failed: %w", err)
	}

	return nil
}

// validateAuthorizationListWithState validates authorization list against current state
func (v *SetCodeTxPoolValidator) validateAuthorizationListWithState(
	setCodeTx *types.SetCodeTx,
	statedb *state.StateDB,
	head *EvmHeader,
) error {
	// Create nonce getter function
	getAccountNonce := func(addr common.Address) uint64 {
		return statedb.GetNonce(addr)
	}

	// Process authorization list
	delegations, err := v.processor.ProcessDelegations(
		setCodeTx,
		head.Number,
		getAccountNonce,
	)
	if err != nil {
		return fmt.Errorf("delegation processing failed: %w", err)
	}

	// Validate each delegation target exists and has code
	for authority, codeAddr := range delegations {
		// Check if authority account exists
		if !statedb.Exist(authority) {
			log.Debug("Authority account does not exist", "authority", authority.Hex())
			// This is not necessarily an error for new accounts
		}

		// Check if code address has code (if not zero address)
		if codeAddr != (common.Address{}) {
			code := statedb.GetCode(codeAddr)
			if len(code) == 0 {
				return fmt.Errorf("delegation target %s has no code", codeAddr.Hex())
			}
		}
	}

	return nil
}

// Constants for SetCode transaction validation
const (
	MaxAuthorizationListSize = 256 // Maximum number of authorizations per transaction (EIP-7702 standard)
)

// SetCodeTxPoolManager manages SetCode transactions in the transaction pool
type SetCodeTxPoolManager struct {
	validator     *SetCodeTxPoolValidator
	delegationMap map[common.Address]common.Address // authority -> code address mappings
}

// NewSetCodeTxPoolManager creates a new SetCode transaction pool manager
func NewSetCodeTxPoolManager(chainID *big.Int) *SetCodeTxPoolManager {
	return &SetCodeTxPoolManager{
		validator:     NewSetCodeTxPoolValidator(chainID),
		delegationMap: make(map[common.Address]common.Address),
	}
}

// ProcessSetCodeTransaction processes a SetCode transaction for pool inclusion
func (m *SetCodeTxPoolManager) ProcessSetCodeTransaction(
	tx *types.Transaction,
	statedb *state.StateDB,
	head *EvmHeader,
	signer types.Signer,
	basicOpts *ValidationOptions,
	stateOpts *ValidationOptionsWithState,
) error {
	// Validate the transaction
	if err := m.validator.ValidateSetCodeTransactionWithState(tx, statedb, head, signer, basicOpts, stateOpts); err != nil {
		return err
	}

	// Extract and process delegation mappings
	inner := tx.Inner()
	setCodeTx, ok := inner.(*types.SetCodeTx)
	if !ok {
		return fmt.Errorf("invalid SetCode transaction inner type")
	}

	// Update delegation mappings (for tracking purposes)
	m.updateDelegationMappings(setCodeTx, statedb, head)

	return nil
}

// updateDelegationMappings updates the internal delegation mapping cache
func (m *SetCodeTxPoolManager) updateDelegationMappings(
	setCodeTx *types.SetCodeTx,
	statedb *state.StateDB,
	head *EvmHeader,
) {
	getAccountNonce := func(addr common.Address) uint64 {
		return statedb.GetNonce(addr)
	}

	delegations, err := m.validator.processor.ProcessDelegations(
		setCodeTx,
		head.Number,
		getAccountNonce,
	)
	if err != nil {
		log.Error("Failed to process delegations for mapping update", "error", err)
		return
	}

	// Update internal mappings
	for authority, codeAddr := range delegations {
		m.delegationMap[authority] = codeAddr
	}
}

// GetDelegation returns the code address for a given authority (if any)
func (m *SetCodeTxPoolManager) GetDelegation(authority common.Address) *common.Address {
	if codeAddr, exists := m.delegationMap[authority]; exists {
		return &codeAddr
	}
	return nil
}

// ClearDelegations clears all cached delegation mappings
func (m *SetCodeTxPoolManager) ClearDelegations() {
	m.delegationMap = make(map[common.Address]common.Address)
}

// GetDelegationStats returns statistics about cached delegations
func (m *SetCodeTxPoolManager) GetDelegationStats() (int, []common.Address) {
	authorities := make([]common.Address, 0, len(m.delegationMap))
	for authority := range m.delegationMap {
		authorities = append(authorities, authority)
	}
	return len(m.delegationMap), authorities
}