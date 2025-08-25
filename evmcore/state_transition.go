// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package evmcore

import (
	"fmt"
	"math"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/params"
)

var emptyCodeHash = crypto.Keccak256Hash(nil)

/*
The State Transitioning Model

A state transition is a change made when a transaction is applied to the current world state
The state transitioning model does all the necessary work to work out a valid new state root.

1) Nonce handling
2) Pre pay gas
3) Create a new state object if the recipient is \0*32
4) Value transfer
== If contract creation ==

	4a) Attempt to run transaction data
	4b) If valid, use result as code for the new state object

== end ==
5) Run Script section
6) Derive new state root
*/
type StateTransition struct {
	gp             *GasPool
	msg            Message
	gas            uint64
	gasPrice       *big.Int
	initialGas     uint64
	value          *big.Int
	data           []byte
	state          vm.StateDB
	consensusState vm.StateDB
	evm            *vm.EVM
}

// Message represents a message sent to a contract.
type Message interface {
	From() common.Address
	To() *common.Address

	GasPrice() *big.Int
	GasFeeCap() *big.Int
	GasTipCap() *big.Int
	Gas() uint64
	Value() *big.Int

	Nonce() uint64
	IsFake() bool
	Data() []byte
	AccessList() types.AccessList
	SetCodeAuthorizations() types.AuthorizationList
}

// ExecutionResult includes all output after executing given an evm
// message no matter the execution itself is successful or not.
type ExecutionResult struct {
	UsedGas     uint64 // Total used gas but include the refunded gas
	RefundedGas uint64 // Total gas refunded after execution
	Err         error  // Any error encountered during the execution (listed in core/vm/errors.go)
	ReturnData  []byte // Returned data from evm(function result or data supplied with revert opcode)
}

// Unwrap returns the internal evm error which allows us for further
// analysis outside.
func (result *ExecutionResult) Unwrap() error {
	return result.Err
}

// Failed returns the indicator whether the execution is successful or not
func (result *ExecutionResult) Failed() bool { return result.Err != nil }

// Return is a helper function to help caller distinguish between revert reason
// and function return. Return returns the data after execution if no error occurs.
func (result *ExecutionResult) Return() []byte {
	if result.Err != nil {
		return nil
	}
	return common.CopyBytes(result.ReturnData)
}

// Revert returns the concrete revert reason if the execution is aborted by `REVERT`
// opcode. Note the reason can be nil if no data supplied with revert opcode.
func (result *ExecutionResult) Revert() []byte {
	if result.Err != vm.ErrExecutionReverted {
		return nil
	}
	return common.CopyBytes(result.ReturnData)
}

// toWordSize returns the ceiled word size required for memory expansion.
func toWordSize(size uint64) uint64 {
	if size > math.MaxUint64-31 {
		return math.MaxUint64/32 + 1
	}
	return (size + 31) / 32
}

// IntrinsicGas computes the 'intrinsic gas' for a message with the given data,
// access list, and authorization list.
func IntrinsicGas(data []byte, accessList types.AccessList, authList types.AuthorizationList, isContractCreation, isHomestead, isEIP2028, isEIP3860 bool) (uint64, error) {
	// Set the starting gas for the raw transaction
	var gas uint64
	if isContractCreation && isHomestead {
		gas = params.TxGasContractCreation
	} else {
		gas = params.TxGas
	}
	
	// Bump the required gas by the amount of transactional data
	dataLen := uint64(len(data))
	if dataLen > 0 {
		// Zero and non-zero bytes are priced differently
		var nz uint64
		for _, byt := range data {
			if byt != 0 {
				nz++
			}
		}
		
		// Make sure we don't exceed uint64 for all data combinations
		var nonZeroGas uint64
		if isEIP2028 {
			nonZeroGas = params.TxDataNonZeroGasEIP2028
		} else {
			nonZeroGas = params.TxDataNonZeroGasFrontier
		}
		
		if (math.MaxUint64-gas)/nonZeroGas < nz {
			return 0, ErrGasUintOverflow
		}
		gas += nz * nonZeroGas
		
		z := dataLen - nz
		if (math.MaxUint64-gas)/params.TxDataZeroGas < z {
			return 0, ErrGasUintOverflow
		}
		gas += z * params.TxDataZeroGas
		
		// EIP-3860: Limit and meter initcode
		if isEIP3860 && isContractCreation {
			if dataLen > params.MaxInitCodeSize {
				return 0, fmt.Errorf("max initcode size exceeded: %v > %v", dataLen, params.MaxInitCodeSize)
			}
			// Calculate initcode word cost
			lenWords := toWordSize(dataLen)
			if (math.MaxUint64-gas)/params.InitCodeWordGas < lenWords {
				return 0, ErrGasUintOverflow
			}
			gas += lenWords * params.InitCodeWordGas
		}
	}
	
	// Calculate access list gas
	if accessList != nil {
		accessListAddressGas := uint64(len(accessList)) * params.TxAccessListAddressGas
		if gas > math.MaxUint64-accessListAddressGas {
			return 0, ErrGasUintOverflow
		}
		gas += accessListAddressGas
		
		accessListStorageGas := uint64(accessList.StorageKeys()) * params.TxAccessListStorageKeyGas
		if gas > math.MaxUint64-accessListStorageGas {
			return 0, ErrGasUintOverflow
		}
		gas += accessListStorageGas
	}
	
	// Calculate authorization list gas
	if authList != nil {
		authGas := uint64(len(authList)) * params.TxAuthTupleGas
		if gas > math.MaxUint64-authGas {
			return 0, ErrGasUintOverflow
		}
		gas += authGas
	}
	
	return gas, nil
}

// NewStateTransition initialises and returns a new state transition object.
func NewStateTransition(evm *vm.EVM, msg Message, gp *GasPool) *StateTransition {
	return &StateTransition{
		gp:       gp,
		evm:      evm,
		msg:      msg,
		gasPrice: msg.GasPrice(),
		value:    msg.Value(),
		data:     msg.Data(),
		state:    evm.StateDB,
	}
}

// ApplyMessage computes the new state by applying the given message
// against the old state within the environment.
//
// ApplyMessage returns the bytes returned by any EVM execution (if it took place),
// the gas used (which includes gas refunds) and an error if it failed. An error always
// indicates a core error meaning that the message would always fail for that particular
// state and would never be accepted within a block.
func ApplyMessage(evm *vm.EVM, msg Message, gp *GasPool) (*ExecutionResult, error) {
	res, err := NewStateTransition(evm, msg, gp).TransitionDb()
	if err != nil {
		log.Debug("Tx skipped", "err", err)
	}
	return res, err
}

// to returns the recipient of the message.
func (st *StateTransition) to() common.Address {
	if st.msg == nil || st.msg.To() == nil /* contract creation */ {
		return common.Address{}
	}
	return *st.msg.To()
}

func (st *StateTransition) buyGas() error {
	mgval := new(big.Int).SetUint64(st.msg.Gas())
	mgval = mgval.Mul(mgval, st.gasPrice)
	// Note: U2U doesn't need to check against gasFeeCap instead of gasPrice, as it's too aggressive in the asynchronous environment
	if have, want := st.state.GetBalance(st.msg.From()), mgval; have.Cmp(want) < 0 {
		return fmt.Errorf("%w: address %v have %v want %v", ErrInsufficientFunds, st.msg.From().Hex(), have, want)
	}
	if err := st.gp.SubGas(st.msg.Gas()); err != nil {
		return err
	}
	st.gas += st.msg.Gas()

	st.initialGas = st.msg.Gas()
	st.state.SubBalance(st.msg.From(), mgval)
	return nil
}

func (st *StateTransition) preCheck() error {
	// Only check transactions that are not fake
	if !st.msg.IsFake() {
		// Make sure this transaction's nonce is correct.
		stNonce := st.state.GetNonce(st.msg.From())
		if msgNonce := st.msg.Nonce(); stNonce < msgNonce {
			return fmt.Errorf("%w: address %v, tx: %d state: %d", ErrNonceTooHigh,
				st.msg.From().Hex(), msgNonce, stNonce)
		} else if stNonce > msgNonce {
			return fmt.Errorf("%w: address %v, tx: %d state: %d", ErrNonceTooLow,
				st.msg.From().Hex(), msgNonce, stNonce)
		}
		// Make sure the sender is an EOA
		if codeHash := st.state.GetCodeHash(st.msg.From()); codeHash != emptyCodeHash && codeHash != (common.Hash{}) {
			return fmt.Errorf("%w: address %v, codehash: %s", ErrSenderNoEOA,
				st.msg.From().Hex(), codeHash)
		}
		
		// EIP-7702: Validate authorization list size limit
		if auths := st.msg.SetCodeAuthorizations(); auths != nil && len(auths) > 256 {
			return ErrAuthorizationListTooLarge
		}
	}
	// Note: U2U doesn't need to check gasFeeCap >= BaseFee, because it's already checked by epochcheck
	return st.buyGas()
}

func (st *StateTransition) internal() bool {
	zeroAddr := common.Address{}
	return st.msg.From() == zeroAddr
}

// TransitionDb will transition the state by applying the current message and
// returning the evm execution result with following fields.
//
//   - used gas:
//     total gas used (including gas being refunded)
//   - returndata:
//     the returned data from evm
//   - concrete execution error:
//     various **EVM** error which aborts the execution,
//     e.g. ErrOutOfGas, ErrExecutionReverted
//
// However if any consensus issue encountered, return the error directly with
// nil evm execution result.
func (st *StateTransition) TransitionDb() (*ExecutionResult, error) {
	// First check this message satisfies all consensus rules before
	// applying the message. The rules include these clauses
	//
	// 1. the nonce of the message caller is correct
	// 2. caller has enough balance to cover transaction fee(gaslimit * gasprice)
	// 3. the amount of gas required is available in the block
	// 4. the purchased gas is enough to cover intrinsic usage
	// 5. there is no overflow when calculating intrinsic gas

	// Note: insufficient balance for **topmost** call isn't a consensus error in U2U, unlike Ethereum
	// Such transaction will revert and consume sender's gas

	// Check clauses 1-3, buy gas if everything is correct
	if err := st.preCheck(); err != nil {
		return nil, err
	}
	msg := st.msg
	sender := vm.AccountRef(msg.From())
	contractCreation := msg.To() == nil

	homestead := st.evm.ChainConfig().IsHomestead(st.evm.Context.BlockNumber)
	istanbul := st.evm.ChainConfig().IsIstanbul(st.evm.Context.BlockNumber)
	london := st.evm.ChainConfig().IsLondon(st.evm.Context.BlockNumber)

	// Check clauses 4-5, subtract intrinsic gas if everything is correct
	gas, err := IntrinsicGas(st.data, st.msg.AccessList(), st.msg.SetCodeAuthorizations(), contractCreation, homestead, istanbul, london)
	if err != nil {
		return nil, err
	}
	if st.gas < gas {
		return nil, fmt.Errorf("%w: have %d, want %d", ErrIntrinsicGas, st.gas, gas)
	}
	st.gas -= gas

	// Set up the initial access list.
	if rules := st.evm.ChainConfig().Rules(st.evm.Context.BlockNumber); rules.IsBerlin {
		st.state.PrepareAccessList(msg.From(), msg.To(), vm.ActivePrecompiles(rules), msg.AccessList())
	}

	// Process EIP-7702 authorizations
	st.processAuthorizations(msg)

	var (
		ret   []byte
		vmerr error // vm errors do not effect consensus and are therefore not assigned to err
	)
	if contractCreation {
		ret, _, st.gas, vmerr = st.evm.Create(sender, st.data, st.gas, st.value)
	} else {
		// Increment the nonce for the next transaction
		st.state.SetNonce(msg.From(), st.state.GetNonce(sender.Address())+1)
		ret, st.gas, vmerr = st.evm.Call(sender, st.to(), st.data, st.gas, st.value)
	}
	// use 10% of not used gas
	if !st.internal() {
		st.gas -= st.gas / 10
	}

	var gasRefund uint64
	if !london {
		// Before EIP-3529: refunds were capped to gasUsed / 2
		gasRefund = st.refundGas(params.RefundQuotient)
	} else {
		// After EIP-3529: refunds are capped to gasUsed / 5
		gasRefund = st.refundGas(params.RefundQuotientEIP3529)
	}

	return &ExecutionResult{
		UsedGas:     st.gasUsed(),
		RefundedGas: gasRefund,
		Err:         vmerr,
		ReturnData:  ret,
	}, nil
}

func (st *StateTransition) refundGas(refundQuotient uint64) uint64 {
	// Apply refund counter, capped to a refund quotient
	refund := st.gasUsed() / refundQuotient
	if refund > st.state.GetRefund() {
		refund = st.state.GetRefund()
	}
	st.gas += refund

	// Return wei for remaining gas, exchanged at the original rate.
	remaining := new(big.Int).Mul(new(big.Int).SetUint64(st.gas), st.gasPrice)
	st.state.AddBalance(st.msg.From(), remaining)

	// Also return remaining gas to the block gas counter so it is
	// available for the next transaction.
	st.gp.AddGas(st.gas)

	return refund
}

// gasUsed returns the amount of gas used up by the state transition.
func (st *StateTransition) gasUsed() uint64 {
	return st.initialGas - st.gas
}

// EIP-7702 Authorization Processing Functions

// validateAuthorization validates an EIP-7702 authorization against the state.
// Returns the authority address if valid, or an error describing the validation failure.
func (st *StateTransition) validateAuthorization(auth *types.AuthorizationTuple) (common.Address, error) {
	// Verify chain ID is null or equal to current chain ID
	if auth.ChainID != nil && auth.ChainID.Cmp(st.evm.ChainConfig().ChainID) != 0 {
		return common.Address{}, ErrAuthorizationWrongChainID
	}
	
	// Limit nonce to 2^64-1 per EIP-2681
	if auth.Nonce+1 < auth.Nonce {
		return common.Address{}, ErrAuthorizationNonceOverflow
	}
	
	// Validate signature values and recover authority
	authority, err := auth.RecoverAuthority()
	if err != nil {
		return common.Address{}, fmt.Errorf("%w: %v", ErrAuthorizationInvalidSignature, err)
	}
	
	// Check the authority account:
	// 1) doesn't have code or has existing delegation
	// 2) matches the auth's nonce
	code := st.state.GetCode(authority)
	if _, ok := types.ParseDelegation(code); len(code) != 0 && !ok {
		return authority, ErrAuthorizationDestinationHasCode
	}
	
	if have := st.state.GetNonce(authority); have != auth.Nonce {
		return authority, ErrAuthorizationNonceMismatch
	}
	
	return authority, nil
}

// applyAuthorization applies an EIP-7702 code delegation to the state.
// This function should only be called after successful validation.
func (st *StateTransition) applyAuthorization(auth *types.AuthorizationTuple) error {
	// Validate authorization first
	authority, err := st.validateAuthorization(auth)
	if err != nil {
		// Note: Per EIP-7702, validation errors don't abort execution
		// Invalid authorizations are simply skipped
		return err
	}
	
	// Update nonce and account code
	st.state.SetNonce(authority, auth.Nonce+1)
	
	if auth.Address == (common.Address{}) {
		// Delegation to zero address means clear existing delegation
		st.state.SetCode(authority, nil)
		return nil
	}
	
	// Otherwise install delegation to auth.Address
	st.state.SetCode(authority, types.AddressToDelegation(auth.Address))
	return nil
}

// processAuthorizations processes all authorizations in the message.
// This function will be integrated into TransitionDb in Step 3.2.
func (st *StateTransition) processAuthorizations(msg Message) {
	auths := msg.SetCodeAuthorizations()
	if auths == nil || len(auths) == 0 {
		return // No authorizations to process
	}
	
	for _, auth := range auths {
		// Note: errors are ignored per EIP-7702 spec
		// Invalid authorizations are simply skipped
		_ = st.applyAuthorization(&auth)
	}
}
