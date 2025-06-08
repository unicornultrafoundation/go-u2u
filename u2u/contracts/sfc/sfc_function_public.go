package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// Handler functions for SFC contract public and external functions

// handleInitialize initializes the SFC contract
func handleInitialize(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if contract is already initialized
	revertData, checkGasUsed, err := checkAlreadyInitialized(evm, "initialize")
	var gasUsed uint64 = checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Get the arguments
	if len(args) != 6 {
		return nil, 0, vm.ErrExecutionReverted
	}
	sealedEpoch, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	_totalSupply, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	nodeDriver, ok := args[2].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	lib, ok := args[3].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	_c, ok := args[4].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	owner, ok := args[5].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the addresses are not zero
	emptyAddr := common.Address{}
	if nodeDriver.Cmp(emptyAddr) == 0 || lib.Cmp(emptyAddr) == 0 || _c.Cmp(emptyAddr) == 0 || owner.Cmp(emptyAddr) == 0 {
		revertData, err := encodeRevertReason("initialize", "zero address")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Set the owner
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)), common.BytesToHash(owner.Bytes()))

	// Set the current sealed epoch
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)), common.BigToHash(sealedEpoch))

	// Set the node driver
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(nodeDriverAuthSlot)), common.BytesToHash(nodeDriver.Bytes()))

	// Set the lib address
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(libAddressSlot)), common.BytesToHash(lib.Bytes()))

	// Set the constants manager
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(constantsManagerSlot)), common.BytesToHash(_c.Bytes()))

	// Set the total supply
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)), common.BigToHash(_totalSupply))

	// Set the min gas price
	initialMinGasPrice := big.NewInt(1000000000) // 1 gwei
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)), common.BigToHash(initialMinGasPrice))

	// Set the epoch snapshot end time
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(epochSnapshotSlot)), common.BigToHash(evm.Context.Time))

	// Set the initialized flag
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(isInitialized)), common.BigToHash(big.NewInt(1)))

	// Emit OwnershipTransferred event
	emptyHash := common.Hash{}
	topics := []common.Hash{
		SfcAbi.Events["OwnershipTransferred"].ID,
		emptyHash, // indexed parameter (previous owner - zero address)
		common.BytesToHash(common.LeftPadBytes(owner.Bytes(), 32)), // indexed parameter (new owner)
	}
	data := []byte{}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// Version returns the version of the SFC contract
func handleVersion(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Return the version as a string
	version := "1.0.0"

	// Pack the version string
	result, err := SfcAbi.Methods["version"].Outputs.Pack(version)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// handleUpdateStakeTokenizerAddress updates the stake tokenizer address
func handleUpdateStakeTokenizerAddress(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, checkGasUsed, err := checkOnlyOwner(evm, caller, "updateStakeTokenizerAddress")
	var gasUsed uint64 = checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Set the stake tokenizer address
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(stakeTokenizerAddressSlot)), common.BytesToHash(addr.Bytes()))

	return nil, 0, nil
}

// handleUpdateLibAddress updates the lib address
func handleUpdateLibAddress(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, checkGasUsed, err := checkOnlyOwner(evm, caller, "updateLibAddress")
	var gasUsed uint64 = checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	v, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Set the lib address
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(libAddressSlot)), common.BytesToHash(v.Bytes()))

	return nil, 0, nil
}

// handleCreateValidator creates a new validator
func handleCreateValidator(evm *vm.EVM, caller common.Address, args []interface{}, value *big.Int) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	pubkey, ok := args[0].([]byte)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the pubkey is not empty
	if len(pubkey) == 0 {
		revertData, err := encodeRevertReason("createValidator", "empty pubkey")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Call the minSelfStake method on the ConstantsManager contract
	minSelfStake := getConstantsManagerVariable("minSelfStake")
	// Check that the value is at least the minimum self-stake
	if value.Cmp(minSelfStake) < 0 {
		revertData, err := encodeRevertReason("createValidator", "insufficient self-stake")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the validator doesn't already exist
	validatorID := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorIDSlot)))
	validatorIDBigInt := new(big.Int).SetBytes(validatorID.Bytes())
	if validatorIDBigInt.Cmp(big.NewInt(0)) != 0 {
		revertData, err := encodeRevertReason("createValidator", "validator already exists")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Get the last validator ID
	lastValidatorID := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lastValidatorIDSlot)))
	lastValidatorIDBigInt := new(big.Int).SetBytes(lastValidatorID.Bytes())

	// Increment the last validator ID
	newValidatorID := new(big.Int).Add(lastValidatorIDBigInt, big.NewInt(1))

	// Set the last validator ID
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(lastValidatorIDSlot)), common.BigToHash(newValidatorID))

	// Set the validator ID for the caller
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorIDSlot)), common.BigToHash(newValidatorID))

	// Set the validator status
	validatorStatusSlot, _ := getValidatorStatusSlot(newValidatorID)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorStatusSlot), common.BigToHash(big.NewInt(0))) // OK_STATUS

	// Set the validator created epoch
	validatorCreatedEpochSlot, _ := getValidatorCreatedEpochSlot(newValidatorID)
	currentEpochBigInt, _, err := getCurrentEpoch(evm)
	if err != nil {
		return nil, 0, err
	}
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorCreatedEpochSlot), common.BigToHash(currentEpochBigInt))

	// Set the validator created time
	validatorCreatedTimeSlot, _ := getValidatorCreatedTimeSlot(newValidatorID)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorCreatedTimeSlot), common.BigToHash(evm.Context.Time))

	// Set the validator deactivated epoch
	validatorDeactivatedEpochSlot, _ := getValidatorDeactivatedEpochSlot(newValidatorID)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorDeactivatedEpochSlot), common.BigToHash(big.NewInt(0)))

	// Set the validator deactivated time
	validatorDeactivatedTimeSlot, _ := getValidatorDeactivatedTimeSlot(newValidatorID)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorDeactivatedTimeSlot), common.BigToHash(big.NewInt(0)))

	// Set the validator auth
	validatorAuthSlot, _ := getValidatorAuthSlot(newValidatorID)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorAuthSlot), common.BytesToHash(caller.Bytes()))

	// Set the validator pubkey
	validatorPubkeySlot, _ := getValidatorPubkeySlot(newValidatorID)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorPubkeySlot), common.BytesToHash(pubkey))

	// Emit CreatedValidator event
	topics := []common.Hash{
		SfcLibAbi.Events["CreatedValidator"].ID,
		common.BigToHash(newValidatorID),                            // indexed parameter (validatorID)
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (auth)
	}
	currentEpochBigInt, _, err = getCurrentEpoch(evm)
	if err != nil {
		return nil, 0, err
	}
	data, err := SfcLibAbi.Events["CreatedValidator"].Inputs.NonIndexed().Pack(
		currentEpochBigInt,
		evm.Context.Time,
	)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	// Delegate the value to the validator
	// This is equivalent to _delegate(msg.sender, lastValidatorID, msg.value)
	result, delegateGasUsed, err := handleInternalDelegate(evm, caller, newValidatorID, value)
	if err != nil {
		return result, gasUsed + delegateGasUsed, err
	}

	// Add the gas used by handleInternalDelegate
	gasUsed += delegateGasUsed

	return nil, gasUsed, nil
}

// handleIsOwner returns whether the given address is the owner of the contract
func handleIsOwner(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the owner
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	ownerAddr := common.BytesToAddress(owner.Bytes())

	// Check if the address is the owner
	isOwner := (addr.Cmp(ownerAddr) == 0)

	// Pack the result
	result, err := SfcAbi.Methods["isOwner"].Outputs.Pack(isOwner)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// handleTransferOwnership transfers ownership of the contract
func handleTransferOwnership(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, checkGasUsed, err := checkOnlyOwner(evm, caller, "transferOwnership")
	var gasUsed uint64 = checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	newOwner, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the new owner is not the zero address
	emptyAddr := common.Address{}
	if newOwner.Cmp(emptyAddr) == 0 {
		revertData, err := encodeRevertReason("transferOwnership", "Ownable: new owner is the zero address")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Get the current owner
	currentOwner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	currentOwnerAddr := common.BytesToAddress(currentOwner.Bytes())

	// Set the new owner
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)), common.BytesToHash(newOwner.Bytes()))

	// Emit OwnershipTransferred event
	topics := []common.Hash{
		SfcAbi.Events["OwnershipTransferred"].ID,
		common.BytesToHash(common.LeftPadBytes(currentOwnerAddr.Bytes(), 32)), // indexed parameter (previous owner)
		common.BytesToHash(common.LeftPadBytes(newOwner.Bytes(), 32)),         // indexed parameter (new owner)
	}
	data := []byte{}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleRenounceOwnership renounces ownership of the contract
func handleRenounceOwnership(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Check if caller is the owner
	revertData, checkGasUsed, err := checkOnlyOwner(evm, caller, "renounceOwnership")
	gasUsed += checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Get the current owner
	currentOwner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	currentOwnerAddr := common.BytesToAddress(currentOwner.Bytes())

	// Set the owner to the zero address
	emptyAddr := common.Address{}
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)), common.BytesToHash(emptyAddr.Bytes()))

	// Emit OwnershipTransferred event
	topics := []common.Hash{
		SfcAbi.Events["OwnershipTransferred"].ID,
		common.BytesToHash(common.LeftPadBytes(currentOwnerAddr.Bytes(), 32)), // indexed parameter (previous owner)
		common.Hash{}, // indexed parameter (new owner - zero address)
	}
	data := []byte{}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, gasUsed, nil
}

// handleGetStakeTokenizerAddress returns the stake tokenizer address
func handleGetStakeTokenizerAddress(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the stake tokenizer address
	stakeTokenizerAddress := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeTokenizerAddressSlot)))
	stakeTokenizerAddr := common.BytesToAddress(stakeTokenizerAddress.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["getStakeTokenizerAddress"].Outputs.Pack(stakeTokenizerAddr)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// handleGetTotalStake returns the total stake
func handleGetTotalStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the total stake slot using a cached constant
	totalStakeSlotHash := common.BigToHash(big.NewInt(totalStakeSlot))

	// Get the total stake
	totalStake := evm.SfcStateDB.GetState(ContractAddress, totalStakeSlotHash)
	gasUsed += SloadGasCost

	// Use the big.Int pool
	totalStakeBigInt := GetBigInt().SetBytes(totalStake.Bytes())

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["getTotalStake"].Outputs.Pack(totalStakeBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Return the big.Int to the pool
	PutBigInt(totalStakeBigInt)

	return result, gasUsed, nil
}

// handleGetTotalActiveStake returns the total active stake
func handleGetTotalActiveStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the total active stake slot using a cached constant
	totalActiveStakeSlotHash := common.BigToHash(big.NewInt(totalActiveStakeSlot))

	// Get the total active stake
	totalActiveStake := evm.SfcStateDB.GetState(ContractAddress, totalActiveStakeSlotHash)
	gasUsed += SloadGasCost

	// Use the big.Int pool
	totalActiveStakeBigInt := GetBigInt().SetBytes(totalActiveStake.Bytes())

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["getTotalActiveStake"].Outputs.Pack(totalActiveStakeBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Return the big.Int to the pool
	PutBigInt(totalActiveStakeBigInt)

	return result, gasUsed, nil
}

// handleGetCurrentEpoch returns the current epoch
func handleGetCurrentEpoch(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the current epoch using the utility function
	currentEpochBigInt, _, err := getCurrentEpoch(evm)
	if err != nil {
		return nil, 0, err
	}

	// Pack the result
	result, err := SfcAbi.Methods["currentEpoch"].Outputs.Pack(currentEpochBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// handleGetCurrentSealedEpoch returns the current sealed epoch
func handleGetCurrentSealedEpoch(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the current sealed epoch slot using a cached constant
	currentSealedEpochSlotHash := common.BigToHash(big.NewInt(currentSealedEpochSlot))

	// Get the current sealed epoch
	currentSealedEpoch := evm.SfcStateDB.GetState(ContractAddress, currentSealedEpochSlotHash)
	gasUsed += SloadGasCost

	// Use the big.Int pool
	currentSealedEpochBigInt := GetBigInt().SetBytes(currentSealedEpoch.Bytes())

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["currentSealedEpoch"].Outputs.Pack(currentSealedEpochBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Return the big.Int to the pool
	PutBigInt(currentSealedEpochBigInt)

	return result, gasUsed, nil
}

// handleGetLastValidatorID returns the last validator ID
func handleGetLastValidatorID(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the last validator ID slot using a cached constant
	lastValidatorIDSlotHash := common.BigToHash(big.NewInt(lastValidatorIDSlot))

	// Get the last validator ID
	lastValidatorID := evm.SfcStateDB.GetState(ContractAddress, lastValidatorIDSlotHash)
	gasUsed += SloadGasCost

	// Use the big.Int pool
	lastValidatorIDBigInt := GetBigInt().SetBytes(lastValidatorID.Bytes())

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["lastValidatorID"].Outputs.Pack(lastValidatorIDBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Return the big.Int to the pool
	PutBigInt(lastValidatorIDBigInt)

	return result, gasUsed, nil
}

// handleGetMinGasPrice returns the minimum gas price
func handleGetMinGasPrice(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the minimum gas price
	minGasPrice := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)))
	minGasPriceBigInt := new(big.Int).SetBytes(minGasPrice.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["minGasPrice"].Outputs.Pack(minGasPriceBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// CurrentEpoch returns the current epoch
func handleCurrentEpoch(evm *vm.EVM) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the current epoch using the utility function
	currentEpochBigInt, epochGasUsed, err := getCurrentEpoch(evm)
	gasUsed += epochGasUsed
	if err != nil {
		return nil, 0, err
	}

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["currentEpoch"].Outputs.Pack(currentEpochBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Return the big.Int to the pool
	PutBigInt(currentEpochBigInt)

	return result, gasUsed, nil
}

// ConstsAddress returns the address of the constants contract
func handleConstsAddress(evm *vm.EVM) ([]byte, uint64, error) {
	// TODO: Implement constsAddress handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// GetEpochValidatorIDs returns the validator IDs for a given epoch
func handleGetEpochValidatorIDs(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement getEpochValidatorIDs handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// GetEpochReceivedStake returns the received stake for a validator in a given epoch
func handleGetEpochReceivedStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement getEpochReceivedStake handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// GetEpochAccumulatedRewardPerToken returns the accumulated reward per token for a validator in a given epoch
func handleGetEpochAccumulatedRewardPerToken(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement getEpochAccumulatedRewardPerToken handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// GetEpochAccumulatedUptime returns the accumulated uptime for a validator in a given epoch
func handleGetEpochAccumulatedUptime(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement getEpochAccumulatedUptime handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// GetEpochAccumulatedOriginatedTxsFee returns the accumulated originated txs fee for a validator in a given epoch
func handleGetEpochAccumulatedOriginatedTxsFee(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement getEpochAccumulatedOriginatedTxsFee handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// GetEpochOfflineTime returns the offline time for a validator in a given epoch
func handleGetEpochOfflineTime(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement getEpochOfflineTime handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// GetEpochOfflineBlocks returns the offline blocks for a validator in a given epoch
func handleGetEpochOfflineBlocks(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement getEpochOfflineBlocks handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// RewardsStash returns the rewards stash for a delegator and validator
func handleRewardsStash(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement rewardsStash handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// GetLockedStake returns the locked stake for a delegator and validator
func handleGetLockedStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement getLockedStake handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// IsSlashed returns whether a validator is slashed
func handleIsSlashed(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement isSlashed handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// PendingRewards returns the pending rewards for a delegator and validator
func handlePendingRewards(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement pendingRewards handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleStashRewards stashes the rewards for a delegator
func handleStashRewards(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement handleStashRewards handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// UpdateConstsAddress updates the address of the constants contract
func handleUpdateConstsAddress(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateConstsAddress handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// UpdateTreasuryAddress updates the address of the treasury
func handleUpdateTreasuryAddress(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateTreasuryAddress handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// UpdateVoteBookAddress updates the address of the vote book
func handleUpdateVoteBookAddress(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateVoteBookAddress handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// DeactivateValidator deactivates a validator
func handleDeactivateValidator(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement deactivateValidator handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// UpdateBaseRewardPerSecond updates the base reward per second
func handleUpdateBaseRewardPerSecond(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateBaseRewardPerSecond handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// UpdateOfflinePenaltyThreshold updates the offline penalty threshold
func handleUpdateOfflinePenaltyThreshold(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateOfflinePenaltyThreshold handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// UpdateSlashingRefundRatio updates the slashing refund ratio
func handleUpdateSlashingRefundRatio(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateSlashingRefundRatio handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// SetGenesisValidator sets a genesis validator
func handleSetGenesisValidator(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement setGenesisValidator handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// SetGenesisDelegation sets a genesis delegation
func handleSetGenesisDelegation(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement setGenesisDelegation handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// SumRewards sums rewards
func handleSumRewards(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement sumRewards handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// Fallback is the payable fallback function that delegates calls to the library
func handleFallback(evm *vm.EVM, args []interface{}, input []byte) ([]byte, uint64, error) {
	//var gasUsed uint64 = 0
	//
	//// Check if input data is empty (pure native token transfer)
	//if len(input) == 0 {
	//	// Return ABI-encoded revert reason: "transfers not allowed"
	//	revertReason := "transfers not allowed"
	//	revertData, err := encodeRevertReason("fallback", revertReason)
	//	if err != nil {
	//		return nil, gasUsed, vm.ErrExecutionReverted
	//	}
	//	return revertData, gasUsed, vm.ErrExecutionReverted
	//}
	//
	//// Create a contract reference for the caller
	//callerRef := vm.AccountRef(evm.TxContext.Origin)
	//
	//// Make the delegate call to the libAddress
	//// This simulates the Solidity _delegate function
	//gas := defaultGasLimit // Use a fixed gas amount for now
	//ret, leftOverGas, err := evm.DelegateCallSFC(callerRef, SfcLibAddr, input, gas)
	//
	//// Calculate gas used
	//gasUsed += gas - leftOverGas
	//
	//// Handle errors similar to the Solidity assembly code:
	//// switch result
	//// case 0 { revert(0, returndatasize) }
	//// default { return (0, returndatasize) }
	//if err != nil {
	//	return ret, gasUsed, err
	//}
	//
	//return ret, gasUsed, nil

	return nil, 0, vm.ErrSfcFunctionNotImplemented
}
