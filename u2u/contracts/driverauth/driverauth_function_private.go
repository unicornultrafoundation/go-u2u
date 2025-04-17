package driverauth

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// handleFallback handles the fallback function (when no method is specified)
func handleFallback(evm *vm.EVM, caller common.Address, args []interface{}, input []byte) ([]byte, uint64, error) {
	// The NodeDriverAuth contract doesn't have a fallback function that does anything
	// Just return empty result with no error
	return nil, 0, nil
}

// handleSetGenesisValidator sets a genesis validator
func handleSetGenesisValidator(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the driver contract
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())
	if caller != driverAddr {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the arguments
	if len(args) != 8 {
		return nil, 0, vm.ErrExecutionReverted
	}
	auth, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	pubkey, ok := args[2].([]byte)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	status, ok := args[3].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	createdEpoch, ok := args[4].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	createdTime, ok := args[5].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	deactivatedEpoch, ok := args[6].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	deactivatedTime, ok := args[7].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the SFC contract to set the genesis validator
	sfcAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("setGenesisValidator", auth, validatorID, pubkey, status, createdEpoch, createdTime, deactivatedEpoch, deactivatedTime)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the SFC contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), sfcAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleSetGenesisDelegation sets a genesis delegation
func handleSetGenesisDelegation(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the driver contract
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())
	if caller != driverAddr {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the arguments
	if len(args) != 9 {
		return nil, 0, vm.ErrExecutionReverted
	}
	delegator, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	toValidatorID, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	stake, ok := args[2].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	lockedStake, ok := args[3].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	lockupFromEpoch, ok := args[4].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	lockupEndTime, ok := args[5].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	lockupDuration, ok := args[6].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	earlyUnlockPenalty, ok := args[7].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	rewards, ok := args[8].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the SFC contract to set the genesis delegation
	sfcAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("setGenesisDelegation", delegator, toValidatorID, stake, lockedStake, lockupFromEpoch, lockupEndTime, lockupDuration, earlyUnlockPenalty, rewards)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the SFC contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), sfcAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleDeactivateValidator deactivates a validator
func handleDeactivateValidator(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the driver contract
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())
	if caller != driverAddr {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the arguments
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	status, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the SFC contract to deactivate the validator
	sfcAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("deactivateValidator", validatorID, status)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the SFC contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), sfcAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleSealEpochValidators seals the epoch validators
func handleSealEpochValidators(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the driver contract
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())
	if caller != driverAddr {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	nextValidatorIDs, ok := args[0].([]*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the SFC contract to seal the epoch validators
	sfcAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("sealEpochValidators", nextValidatorIDs)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the SFC contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), sfcAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleSealEpoch seals the epoch
func handleSealEpoch(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the driver contract
	driverAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot))).Bytes())
	if caller != driverAddr {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the arguments
	if len(args) != 5 {
		return nil, 0, vm.ErrExecutionReverted
	}
	offlineTimes, ok := args[0].([]*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	offlineBlocks, ok := args[1].([]*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	uptimes, ok := args[2].([]*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	originatedTxsFee, ok := args[3].([]*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	usedGas, ok := args[4].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the SFC contract to seal the epoch
	sfcAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAuthAbi.Pack("sealEpoch", offlineTimes, offlineBlocks, uptimes, originatedTxsFee, usedGas)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the SFC contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), sfcAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}
