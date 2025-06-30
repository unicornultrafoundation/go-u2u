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
		BlockNumber: evm.Context.BlockNumber.Uint64(),
		Address:     ContractAddress,
		Topics:      topics,
		Data:        data,
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

	// Call the internal _createValidator function
	newValidatorID, createValidatorGasUsed, err := handleInternalCreateValidator(evm, caller, pubkey)
	gasUsed += createValidatorGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

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
		BlockNumber: evm.Context.BlockNumber.Uint64(),
		Address:     ContractAddress,
		Topics:      topics,
		Data:        data,
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
		BlockNumber: evm.Context.BlockNumber.Uint64(),
		Address:     ContractAddress,
		Topics:      topics,
		Data:        data,
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
	var gasUsed uint64 = 0
	// Get the constants manager address from storage
	constantsManager := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(constantsManagerSlot)))
	gasUsed += SloadGasCost
	constantsManagerAddr := common.BytesToAddress(constantsManager.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["constsAddress"].Outputs.Pack(constantsManagerAddr)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// GetEpochValidatorIDs returns the validator IDs for a given epoch
func handleGetEpochValidatorIDs(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Parse arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	epoch, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// The validatorIDs field is at offset 6 within the EpochSnapshot struct
	validatorIDsSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(validatorIDsOffset))

	// Read the length of the validatorIDs array
	validatorIDsLengthHash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorIDsSlot))
	gasUsed += SloadGasCost
	validatorIDsLength := validatorIDsLengthHash.Big().Uint64()

	// If no validators for this epoch, return empty array
	if validatorIDsLength == 0 {
		result, err := SfcAbi.Methods["getEpochValidatorIDs"].Outputs.Pack([]*big.Int{})
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return result, gasUsed, nil
	}

	// Calculate the base slot for the array elements: keccak256(validatorIDsSlot)
	validatorIDsBaseSlotBytes := CachedKeccak256(common.BigToHash(validatorIDsSlot).Bytes())
	gasUsed += HashGasCost
	validatorIDsBaseSlot := new(big.Int).SetBytes(validatorIDsBaseSlotBytes)

	// Read each validator ID from storage
	validatorIDs := make([]*big.Int, 0, validatorIDsLength)
	for i := uint64(0); i < validatorIDsLength; i++ {
		elementSlot := new(big.Int).Add(validatorIDsBaseSlot, big.NewInt(int64(i)))
		validatorIDHash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(elementSlot))
		gasUsed += SloadGasCost
		validatorID := new(big.Int).SetBytes(validatorIDHash.Bytes())
		validatorIDs = append(validatorIDs, validatorID)
	}

	// Pack the result
	result, err := SfcAbi.Methods["getEpochValidatorIDs"].Outputs.Pack(validatorIDs)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// GetEpochReceivedStake returns the received stake for a validator in a given epoch
func handleGetEpochReceivedStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Parse arguments
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	epoch, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the epoch received stake slot for this validator in this epoch
	receivedStakeSlot, slotGasUsed := getEpochValidatorReceivedStakeSlot(epoch, validatorID)
	gasUsed += slotGasUsed

	// Read the received stake from storage
	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(receivedStakeSlot))
	gasUsed += SloadGasCost
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["getEpochReceivedStake"].Outputs.Pack(receivedStakeBigInt)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// GetEpochAccumulatedRewardPerToken returns the accumulated reward per token for a validator in a given epoch
func handleGetEpochAccumulatedRewardPerToken(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Parse arguments
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	epoch, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// The accumulatedRewardPerToken mapping is at offset 1 within the EpochSnapshot struct
	mappingSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(accumulatedRewardPerTokenOffset))

	// Calculate the slot for the specific validatorID in this mapping
	hashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)
	accumulatedRewardPerTokenSlotHash := CachedKeccak256Hash(hashInput)
	gasUsed += HashGasCost
	accumulatedRewardPerTokenSlot := new(big.Int).SetBytes(accumulatedRewardPerTokenSlotHash.Bytes())

	// Read the accumulated reward per token from storage
	accumulatedRewardPerToken := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(accumulatedRewardPerTokenSlot))
	gasUsed += SloadGasCost
	accumulatedRewardPerTokenBigInt := new(big.Int).SetBytes(accumulatedRewardPerToken.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["getEpochAccumulatedRewardPerToken"].Outputs.Pack(accumulatedRewardPerTokenBigInt)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// GetEpochAccumulatedUptime returns the accumulated uptime for a validator in a given epoch
func handleGetEpochAccumulatedUptime(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Parse arguments
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	epoch, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// The accumulatedUptime mapping is at offset 2 within the EpochSnapshot struct
	mappingSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(accumulatedUptimeOffset))

	// Calculate the slot for the specific validatorID in this mapping
	hashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)
	accumulatedUptimeSlotHash := CachedKeccak256Hash(hashInput)
	gasUsed += HashGasCost
	accumulatedUptimeSlot := new(big.Int).SetBytes(accumulatedUptimeSlotHash.Bytes())

	// Read the accumulated uptime from storage
	accumulatedUptime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(accumulatedUptimeSlot))
	gasUsed += SloadGasCost
	accumulatedUptimeBigInt := new(big.Int).SetBytes(accumulatedUptime.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["getEpochAccumulatedUptime"].Outputs.Pack(accumulatedUptimeBigInt)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// GetEpochAccumulatedOriginatedTxsFee returns the accumulated originated txs fee for a validator in a given epoch
func handleGetEpochAccumulatedOriginatedTxsFee(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Parse arguments
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	epoch, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// The accumulatedOriginatedTxsFee mapping is at offset 3 within the EpochSnapshot struct
	mappingSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(accumulatedOriginatedTxsFeeOffset))

	// Calculate the slot for the specific validatorID in this mapping
	hashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)
	accumulatedTxsFeeSlotHash := CachedKeccak256Hash(hashInput)
	gasUsed += HashGasCost
	accumulatedTxsFeeSlot := new(big.Int).SetBytes(accumulatedTxsFeeSlotHash.Bytes())

	// Read the accumulated originated txs fee from storage
	accumulatedTxsFee := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(accumulatedTxsFeeSlot))
	gasUsed += SloadGasCost
	accumulatedTxsFeeBigInt := new(big.Int).SetBytes(accumulatedTxsFee.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["getEpochAccumulatedOriginatedTxsFee"].Outputs.Pack(accumulatedTxsFeeBigInt)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// GetEpochOfflineTime returns the offline time for a validator in a given epoch
func handleGetEpochOfflineTime(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Parse arguments
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	epoch, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// The offlineTime mapping is at offset 4 within the EpochSnapshot struct
	mappingSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(offlineTimeOffset))

	// Calculate the slot for the specific validatorID in this mapping
	hashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)
	offlineTimeSlotHash := CachedKeccak256Hash(hashInput)
	gasUsed += HashGasCost
	offlineTimeSlot := new(big.Int).SetBytes(offlineTimeSlotHash.Bytes())

	// Read the offline time from storage
	offlineTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(offlineTimeSlot))
	gasUsed += SloadGasCost
	offlineTimeBigInt := new(big.Int).SetBytes(offlineTime.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["getEpochOfflineTime"].Outputs.Pack(offlineTimeBigInt)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// GetEpochOfflineBlocks returns the offline blocks for a validator in a given epoch
func handleGetEpochOfflineBlocks(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Parse arguments
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	epoch, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the epoch snapshot slot
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(epoch)
	gasUsed += slotGasUsed

	// The offlineBlocks mapping is at offset 5 within the EpochSnapshot struct
	mappingSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(offlineBlocksOffset))

	// Calculate the slot for the specific validatorID in this mapping
	hashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)
	offlineBlocksSlotHash := CachedKeccak256Hash(hashInput)
	gasUsed += HashGasCost
	offlineBlocksSlot := new(big.Int).SetBytes(offlineBlocksSlotHash.Bytes())

	// Read the offline blocks from storage
	offlineBlocks := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(offlineBlocksSlot))
	gasUsed += SloadGasCost
	offlineBlocksBigInt := new(big.Int).SetBytes(offlineBlocks.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["getEpochOfflineBlocks"].Outputs.Pack(offlineBlocksBigInt)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// RewardsStash returns the rewards stash for a delegator and validator
func handleRewardsStash(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Parse arguments
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	delegator, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the rewards stash slot for this delegator and validator
	rewardsStashSlot, slotGasUsed := getRewardsStashSlot(delegator, validatorID)
	gasUsed += slotGasUsed

	// Read all three slots of the rewards stash (Rewards struct has 3 fields)
	packedRewardsStash := make([][]byte, 3)
	for i := 0; i < 3; i++ {
		slot := new(big.Int).Add(rewardsStashSlot, big.NewInt(int64(i)))
		value := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slot))
		packedRewardsStash[i] = value.Bytes()
		gasUsed += SloadGasCost
	}

	// Unpack the rewards stash
	rewards := unpackRewards(packedRewardsStash)

	// Calculate the total rewards (lockupBaseReward + lockupExtraReward + unlockedReward)
	totalRewards := new(big.Int).Add(rewards.LockupBaseReward, rewards.LockupExtraReward)
	totalRewards = new(big.Int).Add(totalRewards, rewards.UnlockedReward)

	// Pack the result
	result, err := SfcAbi.Methods["rewardsStash"].Outputs.Pack(totalRewards)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// GetLockedStake returns the locked stake for a delegator and validator
// This is a port of the getLockedStake function from SFCBase.sol
func handleGetLockedStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Parse arguments
	if len(args) != 2 {
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

	// Check if the delegation is locked up using the existing helper function
	// This matches the Solidity: if (!isLockedUp(delegator, toValidatorID)) { return 0; }
	isLockedUpArgs := []interface{}{delegator, toValidatorID}
	isLockedUpResult, lockedGasUsed, err := handleIsLockedUp(evm, isLockedUpArgs)
	gasUsed += lockedGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Unpack the isLockedUp result
	isLockedUpValues, err := SfcAbi.Methods["isLockedUp"].Outputs.Unpack(isLockedUpResult)
	if err != nil {
		return nil, gasUsed, err
	}

	isLocked, ok := isLockedUpValues[0].(bool)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// If not locked up, return 0
	if !isLocked {
		result, err := SfcAbi.Methods["getLockedStake"].Outputs.Pack(big.NewInt(0))
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return result, gasUsed, nil
	}

	// If locked up, get the locked stake from getLockupInfo[delegator][toValidatorID].lockedStake
	// This matches the Solidity: return getLockupInfo[delegator][toValidatorID].lockedStake;
	lockedStakeSlot, slotGasUsed := getLockedStakeSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed

	// Load the locked stake from storage
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockedStakeSlot))
	gasUsed += SloadGasCost
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())

	// Pack and return the result
	result, err := SfcAbi.Methods["getLockedStake"].Outputs.Pack(lockedStakeBigInt)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
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
	var gasUsed uint64 = 0

	// Parse arguments: stashRewards(address delegator, uint256 toValidatorID)
	if len(args) != 2 {
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

	// Call the internal stash rewards function
	result, stashGasUsed, err := handleInternalStashRewards(evm, []interface{}{delegator, toValidatorID})
	gasUsed += stashGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Check if anything was stashed (similar to Solidity: require(_stashRewards(delegator, toValidatorID), "nothing to stash"))
	if len(result) == 32 {
		updated := new(big.Int).SetBytes(result)
		if updated.Cmp(big.NewInt(0)) == 0 {
			revertData, err := encodeRevertReason("stashRewards", "nothing to stash")
			if err != nil {
				return nil, gasUsed, vm.ErrExecutionReverted
			}
			return revertData, gasUsed, vm.ErrExecutionReverted
		}
	}

	// Return empty bytes for successful execution (no return value for this function)
	return nil, gasUsed, nil
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
