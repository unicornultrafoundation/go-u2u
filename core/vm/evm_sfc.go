package vm

import (
	"errors"
	"math/big"
	"time"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/log"
)

func (evm *EVM) CallSFC(caller ContractRef, addr common.Address, input []byte, gas uint64, value *big.Int) (ret []byte, leftOverGas uint64, err error) {
	sp, isSfcPrecompile := evm.SfcPrecompile(addr)
	if !isSfcPrecompile {
		return nil, 0, nil
	}
	snapshot := evm.SfcStateDB.Snapshot()
	// Create a state object if not exist, then transfer any value
	if !evm.SfcStateDB.Exist(addr) {
		log.Debug("SFC precompiled account not exist, creating new account", "height", evm.Context.BlockNumber,
			"to", addr.Hex())
		evm.SfcStateDB.CreateAccount(addr)
	}
	evm.Context.Transfer(evm.SfcStateDB, caller.Address(), addr, value)
	// Run SFC precompiled
	log.Info("SFC precompiled calling", "action", "call", "height", evm.Context.BlockNumber,
		"caller", caller.Address().Hex(),
		"to", addr.Hex())
	start := time.Now()
	ret, _, err = sp.Run(evm, caller.Address(), input, gas)
	// TODO(trinhdn97): compared sfc state precompiled gas used/output/error with the correct execution from smc
	// as well for call code, delegate and static calls.
	TotalSfcExecutionElapsed += time.Since(start)

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

func (evm *EVM) CallCodeSFC(caller ContractRef, addr common.Address, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	sp, isSfcPrecompile := evm.SfcPrecompile(addr)
	if isSfcPrecompile && evm.SfcStateDB != nil {
		snapshot := evm.SfcStateDB.Snapshot()
		log.Debug("SFC precompiled calling", "action", "callcode", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(), "to", addr.Hex())
		start := time.Now()
		ret, _, err = sp.Run(evm, caller.Address(), input, gas)
		TotalSfcExecutionElapsed += time.Since(start)
		if err != nil {
			evm.SfcStateDB.RevertToSnapshot(snapshot)
			if !errors.Is(err, ErrExecutionReverted) {
			}
		}
	}
	return ret, gas, err
}

func (evm *EVM) DelegateCallSFC(caller ContractRef, addr common.Address, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	sp, isSfcPrecompile := evm.SfcPrecompile(addr)
	if isSfcPrecompile && evm.SfcStateDB != nil {
		snapshot := evm.SfcStateDB.Snapshot()
		log.Debug("SFC precompiled calling", "action", "delegatecall", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(), "to", addr.Hex())
		start := time.Now()
		ret, _, err = sp.Run(evm, caller.Address(), input, gas)
		TotalSfcExecutionElapsed += time.Since(start)
		if err != nil {
			evm.SfcStateDB.RevertToSnapshot(snapshot)
			if !errors.Is(err, ErrExecutionReverted) {
			}
		}
	}
	return ret, gas, err
}

func (evm *EVM) StaticCallSFC(caller ContractRef, addr common.Address, input []byte, gas uint64) (ret []byte, leftOverGas uint64, err error) {
	sp, isSfcPrecompile := evm.SfcPrecompile(addr)
	if isSfcPrecompile && evm.SfcStateDB != nil {
		snapshot := evm.SfcStateDB.Snapshot()
		log.Debug("SFC precompiled calling", "action", "staticcall", "height", evm.Context.BlockNumber,
			"caller", caller.Address().Hex(),
			"to", addr.Hex())
		start := time.Now()
		ret, _, err = sp.Run(evm, caller.Address(), input, gas)
		TotalSfcExecutionElapsed += time.Since(start)
		if err != nil {
			evm.SfcStateDB.RevertToSnapshot(snapshot)
			if !errors.Is(err, ErrExecutionReverted) {
			}
		}
	}
	return ret, gas, err
}
