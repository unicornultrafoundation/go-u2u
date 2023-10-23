// Copyright 2019 The go-ethereum Authors
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

package vm

import (
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/holiman/uint256"
	"github.com/unicornultrafoundation/go-u2u/libs/params"
)

var activators = map[int]func(*JumpTable){
	3529: enable3529,
	3198: enable3198,
	2938: enable2938,
	2929: enable2929,
	2200: enable2200,
	1884: enable1884,
	1344: enable1344,
}

// EnableEIP enables the given EIP on the config.
// This operation writes in-place, and callers need to ensure that the globally
// defined jump tables are not polluted.
func EnableEIP(eipNum int, jt *JumpTable) error {
	enablerFn, ok := activators[eipNum]
	if !ok {
		return fmt.Errorf("undefined eip %d", eipNum)
	}
	enablerFn(jt)
	return nil
}

func ValidEip(eipNum int) bool {
	_, ok := activators[eipNum]
	return ok
}
func ActivateableEips() []string {
	var nums []string
	for k := range activators {
		nums = append(nums, fmt.Sprintf("%d", k))
	}
	sort.Strings(nums)
	return nums
}

// enable1884 applies EIP-1884 to the given jump table:
// - Increase cost of BALANCE to 700
// - Increase cost of EXTCODEHASH to 700
// - Increase cost of SLOAD to 800
// - Define SELFBALANCE, with cost GasFastStep (5)
func enable1884(jt *JumpTable) {
	// Gas cost changes
	jt[SLOAD].constantGas = params.SloadGasEIP1884
	jt[BALANCE].constantGas = params.BalanceGasEIP1884
	jt[EXTCODEHASH].constantGas = params.ExtcodeHashGasEIP1884

	// New opcode
	jt[SELFBALANCE] = &operation{
		execute:     opSelfBalance,
		constantGas: GasFastStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}
}

func opSelfBalance(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	balance, _ := uint256.FromBig(interpreter.evm.StateDB.GetBalance(scope.Contract.Address()))
	scope.Stack.push(balance)
	return nil, nil
}

// enable1344 applies EIP-1344 (ChainID Opcode)
// - Adds an opcode that returns the current chain’s EIP-155 unique identifier
func enable1344(jt *JumpTable) {
	// New opcode
	jt[CHAINID] = &operation{
		execute:     opChainID,
		constantGas: GasQuickStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}
}

// opChainID implements CHAINID opcode
func opChainID(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	chainId, _ := uint256.FromBig(interpreter.evm.chainConfig.ChainID)
	scope.Stack.push(chainId)
	return nil, nil
}

// enable2200 applies EIP-2200 (Rebalance net-metered SSTORE)
func enable2200(jt *JumpTable) {
	jt[SLOAD].constantGas = params.SloadGasEIP2200
	jt[SSTORE].dynamicGas = gasSStoreEIP2200
}

// enable2929 enables "EIP-2929: Gas cost increases for state access opcodes"
// https://eips.ethereum.org/EIPS/eip-2929
func enable2929(jt *JumpTable) {
	jt[SSTORE].dynamicGas = gasSStoreEIP2929

	jt[SLOAD].constantGas = 0
	jt[SLOAD].dynamicGas = gasSLoadEIP2929

	jt[EXTCODECOPY].constantGas = params.WarmStorageReadCostEIP2929
	jt[EXTCODECOPY].dynamicGas = gasExtCodeCopyEIP2929

	jt[EXTCODESIZE].constantGas = params.WarmStorageReadCostEIP2929
	jt[EXTCODESIZE].dynamicGas = gasEip2929AccountCheck

	jt[EXTCODEHASH].constantGas = params.WarmStorageReadCostEIP2929
	jt[EXTCODEHASH].dynamicGas = gasEip2929AccountCheck

	jt[BALANCE].constantGas = params.WarmStorageReadCostEIP2929
	jt[BALANCE].dynamicGas = gasEip2929AccountCheck

	jt[CALL].constantGas = params.WarmStorageReadCostEIP2929
	jt[CALL].dynamicGas = gasCallEIP2929

	jt[CALLCODE].constantGas = params.WarmStorageReadCostEIP2929
	jt[CALLCODE].dynamicGas = gasCallCodeEIP2929

	jt[STATICCALL].constantGas = params.WarmStorageReadCostEIP2929
	jt[STATICCALL].dynamicGas = gasStaticCallEIP2929

	jt[DELEGATECALL].constantGas = params.WarmStorageReadCostEIP2929
	jt[DELEGATECALL].dynamicGas = gasDelegateCallEIP2929

	// This was previously part of the dynamic cost, but we're using it as a constantGas
	// factor here
	jt[SELFDESTRUCT].constantGas = params.SelfdestructGasEIP150
	jt[SELFDESTRUCT].dynamicGas = gasSelfdestructEIP2929
}

// enable3529 enabled "EIP-3529: Reduction in refunds":
// - Removes refunds for selfdestructs
// - Reduces refunds for SSTORE
// - Reduces max refunds to 20% gas
func enable3529(jt *JumpTable) {
	jt[SSTORE].dynamicGas = gasSStoreEIP3529
	jt[SELFDESTRUCT].dynamicGas = gasSelfdestructEIP3529
}

// enable3198 applies EIP-3198 (BASEFEE Opcode)
// - Adds an opcode that returns the current block's base fee.
func enable3198(jt *JumpTable) {
	// New opcode
	jt[BASEFEE] = &operation{
		execute:     opBaseFee,
		constantGas: GasQuickStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}
}

func revertBeforePaygas(opcode string, execution executionFunc) executionFunc {
	return func(pc *uint64, interpreter *EVMInterpreter, callContext *ScopeContext) ([]byte, error) {
		if !interpreter.evm.TransactionFeePaid {
			error := errors.New("opcode " + opcode + " must be executed after paygas opcode")
			return nil, error
		}

		return execution(pc, interpreter, callContext)
	}
}

// enable2938 enabled "EIP-2938: Account abstraction"
// - Adds an opcode that returns current transaction NONCE
// - Adds an PAYGAS opcode
// - Changes behaviour of some other opcodes
func enable2938(jt *JumpTable) {
	jt[NONCE] = &operation{
		execute:     opNonce,
		constantGas: GasQuickStep,
		minStack:    minStack(0, 1),
		maxStack:    maxStack(0, 1),
	}
	jt[PAYGAS] = &operation{
		execute:     opPaygas,
		constantGas: GasQuickStep,
		minStack:    minStack(2, 0),
		maxStack:    maxStack(2, 0),
	}
	jt[BALANCE].execute = revertBeforePaygas(BALANCE.String(), opBalance)
	jt[BLOCKHASH].execute = revertBeforePaygas(BLOCKHASH.String(), opBlockhash)
	jt[CALLCODE].execute = revertBeforePaygas(CALLCODE.String(), opCallCode)
	jt[CALLCODE].execute = revertBeforePaygas(CALLCODE.String(), opCallCode)
	jt[CALL].execute = revertBeforePaygas(CALL.String(), opCall)
	jt[COINBASE].execute = revertBeforePaygas(COINBASE.String(), opCoinbase)
	jt[CREATE].execute = revertBeforePaygas(CREATE.String(), opCreate)
	jt[CREATE2].execute = revertBeforePaygas(CREATE2.String(), opCreate2)
	jt[DELEGATECALL].execute = revertBeforePaygas(DELEGATECALL.String(), opDelegateCall)
	jt[DIFFICULTY].execute = revertBeforePaygas(DIFFICULTY.String(), opDifficulty)
	jt[EXTCODECOPY].execute = revertBeforePaygas(EXTCODECOPY.String(), opExtCodeCopy)
	jt[EXTCODEHASH].execute = revertBeforePaygas(EXTCODEHASH.String(), opExtCodeHash)
	jt[EXTCODESIZE].execute = revertBeforePaygas(EXTCODESIZE.String(), opExtCodeSize)
	jt[GASLIMIT].execute = revertBeforePaygas(GASLIMIT.String(), opGasLimit)
	jt[NUMBER].execute = revertBeforePaygas(NUMBER.String(), opNumber)
	jt[STATICCALL].execute = revertBeforePaygas(STATICCALL.String(), opStaticCall)
	jt[TIMESTAMP].execute = revertBeforePaygas(TIMESTAMP.String(), opTimestamp)
}

// opBaseFee implements BASEFEE opcode
func opBaseFee(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	baseFee, _ := uint256.FromBig(interpreter.evm.Context.BaseFee)
	scope.Stack.push(baseFee)
	return nil, nil
}

// opNonce implements NONCE opcode
func opNonce(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	nonce, _ := uint256.FromBig(interpreter.evm.Nonce)
	scope.Stack.push(nonce)
	return nil, nil
}

// opPaygas implements PAYGAS opcode
func opPaygas(pc *uint64, interpreter *EVMInterpreter, scope *ScopeContext) ([]byte, error) {
	gasPrice, gasLimit := scope.Stack.pop(), scope.Stack.pop()

	if interpreter.evm.TransactionFeePaid {
		return nil, nil
	}

	mgval := new(big.Int).Set(gasPrice.ToBig())
	mgval = mgval.Mul(mgval, gasLimit.ToBig())

	address := scope.Contract.Address()
	balance := interpreter.evm.StateDB.GetBalance(address)

	if balance.Cmp(mgval) < 0 {
		return nil, ErrInsufficientBalance
	}

	interpreter.evm.GasPrice = gasPrice.ToBig()
	interpreter.evm.GasLimit = gasLimit.ToBig()
	interpreter.evm.TransactionFeePaid = true
	interpreter.evm.StateDB.SubBalance(address, mgval)

	return nil, nil
}
