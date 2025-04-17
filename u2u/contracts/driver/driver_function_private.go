package driver

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// handleFallback handles the fallback function (when no method is specified)
func handleFallback(evm *vm.EVM, caller common.Address, args []interface{}, input []byte) ([]byte, uint64, error) {
	// The NodeDriver contract doesn't have a fallback function that does anything
	// Just return empty result with no error
	return nil, 0, nil
}

// handleSetGenesisValidator sets a genesis validator
func handleSetGenesisValidator(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is address(0) (onlyNode modifier)
	revertData, err := checkOnlyNode(evm, caller, "setGenesisValidator")
	if err != nil {
		return revertData, 0, err
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

	// Call the backend contract to set the genesis validator
	backendAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(backendSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAbi.Pack("setGenesisValidator", auth, validatorID, pubkey, status, createdEpoch, createdTime, deactivatedEpoch, deactivatedTime)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the backend contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), backendAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleSetGenesisDelegation sets a genesis delegation
func handleSetGenesisDelegation(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is address(0) (onlyNode modifier)
	revertData, err := checkOnlyNode(evm, caller, "setGenesisDelegation")
	if err != nil {
		return revertData, 0, err
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

	// Call the backend contract to set the genesis delegation
	backendAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(backendSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAbi.Pack("setGenesisDelegation", delegator, toValidatorID, stake, lockedStake, lockupFromEpoch, lockupEndTime, lockupDuration, earlyUnlockPenalty, rewards)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the backend contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), backendAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleDeactivateValidator deactivates a validator
func handleDeactivateValidator(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is address(0) (onlyNode modifier)
	revertData, err := checkOnlyNode(evm, caller, "deactivateValidator")
	if err != nil {
		return revertData, 0, err
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

	// Call the backend contract to deactivate the validator
	backendAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(backendSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAbi.Pack("deactivateValidator", validatorID, status)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the backend contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), backendAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleSealEpochValidators seals the epoch validators
func handleSealEpochValidators(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is address(0) (onlyNode modifier)
	revertData, err := checkOnlyNode(evm, caller, "sealEpochValidators")
	if err != nil {
		return revertData, 0, err
	}

	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	nextValidatorIDs, ok := args[0].([]*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the backend contract to seal the epoch validators
	backendAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(backendSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAbi.Pack("sealEpochValidators", nextValidatorIDs)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the backend contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), backendAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleSealEpoch seals the epoch
func handleSealEpoch(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is address(0) (onlyNode modifier)
	revertData, err := checkOnlyNode(evm, caller, "sealEpoch")
	if err != nil {
		return revertData, 0, err
	}

	// Get the arguments
	if len(args) != 4 {
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

	// Call the backend contract to seal the epoch with a fixed gas value (841669690)
	backendAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(backendSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAbi.Pack("sealEpoch", offlineTimes, offlineBlocks, uptimes, originatedTxsFee, big.NewInt(841669690))
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the backend contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), backendAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}

// handleSealEpochV1 seals the epoch with a custom gas value
func handleSealEpochV1(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is address(0) (onlyNode modifier)
	revertData, err := checkOnlyNode(evm, caller, "sealEpochV1")
	if err != nil {
		return revertData, 0, err
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

	// Call the backend contract to seal the epoch with the provided gas value
	backendAddr := common.BytesToAddress(evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(backendSlot))).Bytes())

	// Pack the function call data
	data, err := DriverAbi.Pack("sealEpoch", offlineTimes, offlineBlocks, uptimes, originatedTxsFee, usedGas)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call the backend contract
	_, _, err = evm.Call(vm.AccountRef(ContractAddress), backendAddr, data, 50000, big.NewInt(0))
	if err != nil {
		return nil, 0, err
	}

	return nil, 0, nil
}
