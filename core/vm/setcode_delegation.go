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

package vm

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/params"
)

// DelegationResolver handles EIP-7702 delegation resolution for the EVM
type DelegationResolver struct {
	stateDB StateDB
}

// NewDelegationResolver creates a new delegation resolver for the EVM
func NewDelegationResolver(stateDB StateDB) *DelegationResolver {
	return &DelegationResolver{
		stateDB: stateDB,
	}
}

// ResolveDelegatedCode resolves the actual code address through EIP-7702 delegation chain
func (dr *DelegationResolver) ResolveDelegatedCode(address common.Address) (common.Address, []byte, uint64, error) {
	const maxDelegationDepth = 10
	gasUsed := uint64(0)
	visited := make(map[common.Address]bool)
	currentAddr := address

	for depth := 0; depth < maxDelegationDepth; depth++ {
		// Prevent circular delegation
		if visited[currentAddr] {
			log.Warn("Circular delegation detected", "address", currentAddr.Hex())
			break
		}
		visited[currentAddr] = true

		// Get code from current address
		code := dr.stateDB.GetCode(currentAddr)
		if len(code) == 0 {
			// No code at this address, use it as final address
			break
		}

		// Check if code is a delegation
		if delegatedAddr, isDelegation := types.ParseDelegation(code); isDelegation {
			// This is a delegation, follow the chain
			currentAddr = delegatedAddr
			gasUsed += params.DelegationResolutionGas
			log.Debug("Following delegation", "from", currentAddr.Hex(), "to", delegatedAddr.Hex())
		} else {
			// This is actual code, not a delegation
			break
		}
	}

	// Get final code
	finalCode := dr.stateDB.GetCode(currentAddr)
	
	log.Debug("Delegation resolved", 
		"original", address.Hex(), 
		"final", currentAddr.Hex(), 
		"gasUsed", gasUsed,
		"codeSize", len(finalCode))

	return currentAddr, finalCode, gasUsed, nil
}

// CheckDelegation checks if an address has a delegation
func (dr *DelegationResolver) CheckDelegation(address common.Address) bool {
	code := dr.stateDB.GetCode(address)
	if len(code) == 0 {
		return false
	}
	_, isDelegation := types.ParseDelegation(code)
	return isDelegation
}

// GetDirectDelegation returns the direct delegation for an address (without chain resolution)
func (dr *DelegationResolver) GetDirectDelegation(address common.Address) *common.Address {
	code := dr.stateDB.GetCode(address)
	if len(code) == 0 {
		return nil
	}
	if delegatedAddr, isDelegation := types.ParseDelegation(code); isDelegation {
		return &delegatedAddr
	}
	return nil
}

// EnhancedEVM extends the standard EVM with EIP-7702 delegation support
type EnhancedEVM struct {
	*EVM
	delegationResolver *DelegationResolver
}

// NewEnhancedEVM creates a new EVM with EIP-7702 delegation support
func NewEnhancedEVM(blockCtx BlockContext, txCtx TxContext, stateDB StateDB, chainConfig *params.ChainConfig, vmConfig Config) *EnhancedEVM {
	baseEVM := NewEVM(blockCtx, txCtx, stateDB, stateDB, chainConfig, vmConfig)
	return &EnhancedEVM{
		EVM:                baseEVM,
		delegationResolver: NewDelegationResolver(stateDB),
	}
}

// GetCodeWithDelegation retrieves code with EIP-7702 delegation resolution
func (evm *EnhancedEVM) GetCodeWithDelegation(addr common.Address) []byte {
	// First check for delegation
	if evm.delegationResolver.CheckDelegation(addr) {
		_, code, gasUsed, err := evm.delegationResolver.ResolveDelegatedCode(addr)
		if err != nil {
			log.Error("Delegation resolution failed", "address", addr.Hex(), "error", err)
			// Fall back to normal code lookup
			return evm.StateDB.GetCode(addr)
		}
		// Note: Gas accounting is handled at the opcode level, not here
		_ = gasUsed
		return code
	}
	
	// No delegation, return normal code
	return evm.StateDB.GetCode(addr)
}

// CallWithDelegation performs a call with delegation resolution
func (evm *EnhancedEVM) CallWithDelegation(caller ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	// Resolve delegation if present
	codeAddr := addr
	delegationGas := uint64(0)
	if evm.delegationResolver.CheckDelegation(addr) {
		resolvedAddr, _, gasUsed, err := evm.delegationResolver.ResolveDelegatedCode(addr)
		if err != nil {
			log.Error("Delegation resolution failed in call", "address", addr.Hex(), "error", err)
		} else {
			codeAddr = resolvedAddr
			delegationGas = gasUsed
		}
	}

	// Deduct delegation resolution gas
	if delegationGas > 0 {
		if gas < delegationGas {
			return nil, 0, ErrOutOfGas
		}
		gas -= delegationGas
	}

	// If delegation changed the code address, we need to get the code from the resolved address
	// but execute it in the context of the original address
	if codeAddr != addr {
		return evm.CallWithDelegatedCode(caller, addr, codeAddr, input, gas, value)
	}

	// No delegation, perform normal call
	return evm.Call(caller, addr, input, gas, value)
}

// CallWithDelegatedCode performs a call using delegated code
func (evm *EnhancedEVM) CallWithDelegatedCode(caller ContractRef, contextAddr, codeAddr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	// This is similar to DELEGATECALL but with EIP-7702 semantics
	// Execute code from codeAddr in the context of contextAddr
	
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}

	// Gas limit check
	if !evm.Context.CanTransfer(evm.StateDB, caller.Address(), value) {
		return nil, gas, ErrInsufficientBalance
	}

	var snapshot = evm.StateDB.Snapshot()

	// Perform the transfer
	evm.Context.Transfer(evm.StateDB, caller.Address(), contextAddr, value)

	// Get code from the resolved address
	code := evm.StateDB.GetCode(codeAddr)
	var contract *Contract
	
	if len(code) == 0 {
		ret, err = nil, nil
	} else {
		// Create a contract with context address but code from code address
		addrCopy := contextAddr
		contract = NewContract(caller, AccountRef(addrCopy), value, gas)
		contract.SetCodeOptionalHash(&addrCopy, &codeAndHash{code: code})

		ret, err = evm.interpreter.Run(contract, input, false)
	}

	// Handle execution result
	if err != nil {
		evm.StateDB.RevertToSnapshot(snapshot)
		if err != ErrExecutionReverted {
			gas = 0
		}
		// TODO: consider clearing the return data here if not needed
	}
	
	if contract != nil {
		return ret, contract.Gas, err
	}
	return ret, gas, err
}

// DelegateCallWithDelegation performs a DELEGATECALL with EIP-7702 delegation resolution
func (evm *EnhancedEVM) DelegateCallWithDelegation(caller ContractRef, addr common.Address, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	// Resolve delegation if present
	codeAddr := addr
	delegationGas := uint64(0)
	if evm.delegationResolver.CheckDelegation(addr) {
		resolvedAddr, _, gasUsed, err := evm.delegationResolver.ResolveDelegatedCode(addr)
		if err != nil {
			log.Error("Delegation resolution failed in delegatecall", "address", addr.Hex(), "error", err)
		} else {
			codeAddr = resolvedAddr
			delegationGas = gasUsed
		}
	}

	// Deduct delegation resolution gas
	if delegationGas > 0 {
		if gas < delegationGas {
			return nil, 0, ErrOutOfGas
		}
		gas -= delegationGas
	}

	// If delegation resolved to a different address, get code from there
	if codeAddr != addr {
		// Use the resolved code address for DELEGATECALL
		return evm.DelegateCall(caller, codeAddr, input, gas)
	}

	// No delegation, perform normal DELEGATECALL
	return evm.DelegateCall(caller, addr, input, gas)
}

// StaticCallWithDelegation performs a STATICCALL with EIP-7702 delegation resolution
func (evm *EnhancedEVM) StaticCallWithDelegation(caller ContractRef, addr common.Address, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	// Resolve delegation if present
	codeAddr := addr
	delegationGas := uint64(0)
	if evm.delegationResolver.CheckDelegation(addr) {
		resolvedAddr, _, gasUsed, err := evm.delegationResolver.ResolveDelegatedCode(addr)
		if err != nil {
			log.Error("Delegation resolution failed in staticcall", "address", addr.Hex(), "error", err)
		} else {
			codeAddr = resolvedAddr
			delegationGas = gasUsed
		}
	}

	// Deduct delegation resolution gas
	if delegationGas > 0 {
		if gas < delegationGas {
			return nil, 0, ErrOutOfGas
		}
		gas -= delegationGas
	}

	// If delegation resolved to a different address, get code from there
	if codeAddr != addr {
		// Use the resolved code address for STATICCALL
		return evm.StaticCall(caller, codeAddr, input, gas)
	}

	// No delegation, perform normal STATICCALL
	return evm.StaticCall(caller, addr, input, gas)
}

// GetDelegationResolver returns the delegation resolver for external use
func (evm *EnhancedEVM) GetDelegationResolver() *DelegationResolver {
	return evm.delegationResolver
}

// SetCodeEVMContext provides context for EIP-7702 operations
type SetCodeEVMContext struct {
	chainID           *big.Int
	delegationEnabled bool
}

// NewSetCodeEVMContext creates a new EIP-7702 EVM context
func NewSetCodeEVMContext(chainID *big.Int) *SetCodeEVMContext {
	return &SetCodeEVMContext{
		chainID:           chainID,
		delegationEnabled: true,
	}
}

// IsDelegationEnabled returns whether delegation is enabled
func (ctx *SetCodeEVMContext) IsDelegationEnabled() bool {
	return ctx.delegationEnabled
}

// SetDelegationEnabled enables or disables delegation
func (ctx *SetCodeEVMContext) SetDelegationEnabled(enabled bool) {
	ctx.delegationEnabled = enabled
}

// GetChainID returns the chain ID
func (ctx *SetCodeEVMContext) GetChainID() *big.Int {
	return ctx.chainID
}