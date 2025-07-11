package vm

import (
	"errors"
	"math/big"
	"time"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/metrics"
)

// TotalSfcExecutionElapsed is the total execution time of SFC precompiled calls per transaction.
var TotalSfcExecutionElapsed = time.Duration(0)

// CallSFC calls the SFC precompiled contract using the storage from SFC StateDB
// and the logic from SFCPrecompiledContract implementation.
// It also handles any necessary value transfer required and takes
// the necessary steps to create accounts and reverses the state in case of an
// execution error or failed value transfer.
func (evm *EVM) CallSFC(caller ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	if common.IsNilInterface(evm.SfcStateDB) {
		return nil, gas, nil
	}
	// Note that the calls with NoBaseFee config enabled are not included in the total execution time,
	// because these are fake calls from API eth_call or eth_estimateGas, not from a real transaction.
	if metrics.EnabledExpensive && !evm.Config.NoBaseFee {
		defer func(start time.Time) {
			TotalSfcExecutionElapsed += time.Since(start)
		}(time.Now())
	}
	snapshot := evm.SfcStateDB.Snapshot()
	if !evm.SfcStateDB.Exist(addr) {
		log.Debug("SFC precompiled account not exist, creating new account",
			"height", evm.Context.BlockNumber, "addr", addr.Hex())
		evm.SfcStateDB.CreateAccount(addr)
	}
	// Fail if we're trying to transfer more than the available balance
	if _, isSfcPrecompile := evm.SfcPrecompile(caller.Address()); isSfcPrecompile {
		if value.Sign() != 0 && !evm.Context.CanTransfer(evm.SfcStateDB, caller.Address(), value) {
			log.Error("CallSFC: CanTransfer failed", "height", evm.Context.BlockNumber,
				"caller", caller.Address().Hex(), "to", addr.Hex(), "value", value.String())
			return nil, gas, ErrInsufficientBalance
		}
		evm.SfcStateDB.SubBalance(caller.Address(), value)
	}
	if _, isSfcPrecompile := evm.SfcPrecompile(addr); isSfcPrecompile && value.Sign() != 0 {
		evm.SfcStateDB.AddBalance(addr, value)
	}

	// Handle evmWriter calls from NodeDriver contract
	if sp, isStatePrecompile := evm.statePrecompile(addr); isStatePrecompile {
		log.Debug("EvmWriter precompiled calling", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(),
			"to", addr.Hex())
		ret, gas, err = sp.Run(evm.SfcStateDB, evm.Context, evm.TxContext, caller.Address(), input, gas)
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
		log.Debug("CallSFC: SFC precompiled calling", "action", "call", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(), "to", addr.Hex())
		ret, _, err = sp.Run(evm, caller.Address(), input, gas, value)
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
	sp, isSfcPrecompile := evm.SfcPrecompile(addr)
	if isSfcPrecompile && evm.SfcStateDB != nil {
		if metrics.EnabledExpensive && !evm.Config.NoBaseFee {
			defer func(start time.Time) {
				TotalSfcExecutionElapsed += time.Since(start)
			}(time.Now())
		}
		// Create a snapshot for potential revert
		snapshot := evm.SfcStateDB.Snapshot()
		log.Debug("SFC precompiled calling", "action", "callcode", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(), "to", addr.Hex())

		// Run the precompiled contract with the caller's address
		// In callcode, we use the code from the callee but the context (storage) from the caller
		ret, remainingGas, err := sp.Run(evm, caller.Address(), input, gas, value)
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
	sp, isSfcPrecompile := evm.SfcPrecompile(addr)
	if isSfcPrecompile && evm.SfcStateDB != nil {
		if metrics.EnabledExpensive && !evm.Config.NoBaseFee {
			defer func(start time.Time) {
				TotalSfcExecutionElapsed += time.Since(start)
			}(time.Now())
		}
		// Create a snapshot for potential revert
		snapshot := evm.SfcStateDB.Snapshot()
		log.Debug("SFC precompiled calling", "action", "delegatecall", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(), "to", addr.Hex())

		// In a delegate call, we need to preserve the caller context.
		// For SFC precompiled contracts, we need to pass the caller's address,
		// but execute in the context of the caller
		// Run the precompiled contract with the caller's address
		// This simulates executing the code in the caller's context
		ret, remainingGas, err := sp.Run(evm, caller.Address(), input, gas, value)

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
	sp, isSfcPrecompile := evm.SfcPrecompile(addr)
	if isSfcPrecompile && evm.SfcStateDB != nil {
		if metrics.EnabledExpensive && !evm.Config.NoBaseFee {
			defer func(start time.Time) {
				TotalSfcExecutionElapsed += time.Since(start)
			}(time.Now())
		}
		// Create a snapshot for potential revert
		snapshot := evm.SfcStateDB.Snapshot()
		log.Debug("SFC precompiled calling", "action", "staticcall", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(), "to", addr.Hex())

		// Run the precompiled contract with the caller's address
		// For static calls, we should ensure no state modifications
		ret, remainingGas, err := sp.Run(evm, caller.Address(), input, gas, value)
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
