package vm

import (
	"errors"
	"math/big"
	"time"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/params"
)

func (evm *EVM) CallSFC(caller ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	if evm.Config.NoRecursion && evm.depth > 0 {
		log.Error("CallSFC: NoRecursion failed", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(), "to", addr.Hex(), "value", value.String())
		return nil, gas, nil
	}
	// Fail if we're trying to execute above the call depth limit
	if evm.depth > int(params.CallCreateDepth) {
		log.Error("CallSFC: CallCreateDepth failed", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(), "to", addr.Hex(), "value", value.String())
		return nil, gas, ErrDepth
	}
	// Fail if we're trying to transfer more than the available balance
	// TODO(trinhdn97): must double check after totally get rid of EVM flow
	// if value.Sign() != 0 && !evm.Context.CanTransfer(evm.StateDB, caller.Address(), value) {
	// 	log.Error("CallSFC: CanTransfer failed", "height", evm.Context.BlockNumber,
	// 		"caller", caller.Address().Hex(), "to", addr.Hex(), "value", value.String())
	// 	return nil, gas, ErrInsufficientBalance
	// }

	snapshot := evm.SfcStateDB.Snapshot()
	if !evm.SfcStateDB.Exist(addr) {
		log.Debug("SFC precompiled account not exist, creating new account",
			"height", evm.Context.BlockNumber, "addr", addr.Hex())
		evm.SfcStateDB.CreateAccount(addr)
	}
	log.Info("SFC precompiled balance transfer check", "height", evm.Context.BlockNumber,
		"caller", caller.Address().Hex(), "to", addr.Hex(), "value", value.String())
	// Only transfer balance if value is not zero
	if value.Sign() != 0 {
		if _, isSfcPrecompile := evm.SfcPrecompile(caller.Address()); isSfcPrecompile {
			log.Info("SFC precompiled account subtracting balance", "height", evm.Context.BlockNumber,
				"caller", caller.Address().Hex(), "value", value.String())
			evm.SfcStateDB.SubBalance(caller.Address(), value)
		}
		if _, isSfcPrecompile := evm.SfcPrecompile(addr); isSfcPrecompile {
			log.Info("SFC precompiled account adding balance", "height", evm.Context.BlockNumber,
				"caller", caller.Address().Hex(), "to", addr.Hex(), "value", value.String())
			evm.SfcStateDB.AddBalance(addr, value)
		}
	}

	// Handle evmWriter calls from NodeDriver contract
	if sp, isStatePrecompile := evm.statePrecompile(addr); isStatePrecompile {
		log.Info("EvmWriter precompiled calling", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(),
			"to", addr.Hex())
		start := time.Now()
		ret, gas, err = sp.Run(evm.SfcStateDB, evm.Context, evm.TxContext, caller.Address(), input, gas)
		TotalSfcExecutionElapsed += time.Since(start)
		if err != nil {
			log.Error("EvmWriter precompiled calling failed", "height", evm.Context.BlockNumber,
				"caller", caller.Address().Hex(),
				"to", addr.Hex(), "err", err)
		}
	} else {
		sp, isSfcPrecompile := evm.SfcPrecompile(addr)
		if !isSfcPrecompile {
			return nil, 0, nil
		}
		// Run SFC precompiled
		log.Info("SFC precompiled calling", "action", "call", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(),
			"to", addr.Hex())
		start := time.Now()
		ret, _, err = sp.Run(evm, caller.Address(), input, gas, value)
		// TODO(trinhdn97): compared sfc state precompiled gas used/output/error with the correct execution from smc
		// as well for call code, delegate and static calls.
		TotalSfcExecutionElapsed += time.Since(start)
	}
	// When an error was returned by the SFC precompiles or when setting the creation code
	// above, we revert to the snapshot and consume any gas remaining.
	if err != nil {
		evm.SfcStateDB.RevertToSnapshot(snapshot)
		if !errors.Is(err, ErrExecutionReverted) {
			// TODO(trinhdn97): try to consume all remaining gas here, in case this is a valid revert.
		}
	}
	return ret, gas, err
}

func (evm *EVM) CallCodeSFC(caller ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	// Check recursion and depth limits similar to regular CallCode
	if evm.Config.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}

	sp, isSfcPrecompile := evm.SfcPrecompile(addr)
	if isSfcPrecompile && evm.SfcStateDB != nil {
		// Create a snapshot for potential revert
		snapshot := evm.SfcStateDB.Snapshot()

		log.Debug("SFC precompiled calling", "action", "callcode", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(), "to", addr.Hex())

		// Track execution time
		start := time.Now()

		// Run the precompiled contract with the caller's address
		// In callcode, we use the code from the callee but the context (storage) from the caller
		ret, remainingGas, err := sp.Run(evm, caller.Address(), input, gas, value)

		TotalSfcExecutionElapsed += time.Since(start)

		// Handle errors and revert if needed
		if err != nil {
			evm.SfcStateDB.RevertToSnapshot(snapshot)
			if !errors.Is(err, ErrExecutionReverted) {
				gas = 0
			}
		} else {
			// Update gas with the remaining gas from the execution
			gas = remainingGas
		}

		return ret, gas, err
	}

	// If not an SFC precompile or SfcStateDB is nil, return with no changes
	return nil, gas, nil
}

func (evm *EVM) DelegateCallSFC(caller ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	// Check recursion and depth limits similar to regular DelegateCall
	if evm.Config.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}

	sp, isSfcPrecompile := evm.SfcPrecompile(addr)
	if isSfcPrecompile && evm.SfcStateDB != nil {
		// Create a snapshot for potential revert
		snapshot := evm.SfcStateDB.Snapshot()

		log.Debug("SFC precompiled calling", "action", "delegatecall", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(), "to", addr.Hex())

		// In a delegate call, we need to preserve the caller context
		// For SFC precompiled contracts, we need to pass the caller's address
		// but execute in the context of the caller

		// Track execution time
		start := time.Now()

		// Run the precompiled contract with the caller's address
		// This simulates executing the code in the caller's context
		ret, remainingGas, err := sp.Run(evm, caller.Address(), input, gas, value)

		TotalSfcExecutionElapsed += time.Since(start)

		// Handle errors and revert if needed
		if err != nil {
			evm.SfcStateDB.RevertToSnapshot(snapshot)
			if !errors.Is(err, ErrExecutionReverted) {
				gas = 0
			}
		} else {
			// Update gas with the remaining gas from the execution
			gas = remainingGas
		}

		return ret, gas, err
	}

	// If not an SFC precompile or SfcStateDB is nil, return with no changes
	return nil, gas, nil
}

func (evm *EVM) StaticCallSFC(caller ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	// Check recursion and depth limits similar to regular StaticCall
	if evm.Config.NoRecursion && evm.depth > 0 {
		return nil, gas, nil
	}
	if evm.depth > int(params.CallCreateDepth) {
		return nil, gas, ErrDepth
	}

	sp, isSfcPrecompile := evm.SfcPrecompile(addr)
	if isSfcPrecompile && evm.SfcStateDB != nil {
		// Create a snapshot for potential revert
		snapshot := evm.SfcStateDB.Snapshot()

		log.Debug("SFC precompiled calling", "action", "staticcall", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(), "to", addr.Hex())

		// Track execution time
		start := time.Now()

		// Run the precompiled contract with the caller's address
		// For static calls, we should ensure no state modifications
		ret, remainingGas, err := sp.Run(evm, caller.Address(), input, gas, value)

		TotalSfcExecutionElapsed += time.Since(start)

		// Handle errors and revert if needed
		if err != nil {
			evm.SfcStateDB.RevertToSnapshot(snapshot)
			if !errors.Is(err, ErrExecutionReverted) {
				gas = 0
			}
		} else {
			// Update gas with the remaining gas from the execution
			gas = remainingGas
		}

		return ret, gas, err
	}

	// If not an SFC precompile or SfcStateDB is nil, return with no changes
	return nil, gas, nil
}
