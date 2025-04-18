package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/log"
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
	minSelfStakeValues, _, err := callConstantManagerMethod(evm, "minSelfStake")
	if err != nil {
		return nil, 0, err
	}

	// The result should be a single *big.Int value
	if len(minSelfStakeValues) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}

	minSelfStakeBigInt, ok := minSelfStakeValues[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the value is at least the minimum self-stake
	if value.Cmp(minSelfStakeBigInt) < 0 {
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
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)), common.BigToHash(big.NewInt(0))) // OK_STATUS

	// Set the validator created epoch
	validatorCreatedEpochSlot, _ := getValidatorCreatedEpochSlot(newValidatorID)
	currentEpochBigInt, _, err := getCurrentEpoch(evm)
	if err != nil {
		return nil, 0, err
	}
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorCreatedEpochSlot)), common.BigToHash(currentEpochBigInt))

	// Set the validator created time
	validatorCreatedTimeSlot, _ := getValidatorCreatedTimeSlot(newValidatorID)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorCreatedTimeSlot)), common.BigToHash(evm.Context.Time))

	// Set the validator deactivated epoch
	validatorDeactivatedEpochSlot, _ := getValidatorDeactivatedEpochSlot(newValidatorID)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorDeactivatedEpochSlot)), common.BigToHash(big.NewInt(0)))

	// Set the validator deactivated time
	validatorDeactivatedTimeSlot, _ := getValidatorDeactivatedTimeSlot(newValidatorID)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorDeactivatedTimeSlot)), common.BigToHash(big.NewInt(0)))

	// Set the validator auth
	validatorAuthSlot, _ := getValidatorAuthSlot(newValidatorID)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorAuthSlot)), common.BytesToHash(caller.Bytes()))

	// Set the validator pubkey
	validatorPubkeySlot, _ := getValidatorPubkeySlot(newValidatorID)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorPubkeySlot)), common.BytesToHash(pubkey))

	// Emit CreatedValidator event
	topics := []common.Hash{
		SfcAbi.Events["CreatedValidator"].ID,
		common.BigToHash(newValidatorID),                            // indexed parameter (validatorID)
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (auth)
	}
	currentEpochBigInt, _, err = getCurrentEpoch(evm)
	if err != nil {
		return nil, 0, err
	}
	data, err := SfcAbi.Events["CreatedValidator"].Inputs.NonIndexed().Pack(
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

// handleDelegate delegates stake to a validator
func handleDelegate(evm *vm.EVM, caller common.Address, args []interface{}, value *big.Int) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	toValidatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Stash rewards
	// Create arguments for handleStashRewards
	stashRewardsArgs := []interface{}{caller, toValidatorID}
	// Call handleStashRewards
	result, stashGasUsed, err := handleStashRewards(evm, stashRewardsArgs)
	if err != nil {
		return result, gasUsed + stashGasUsed, err
	}

	// Add the gas used by handleStashRewards
	gasUsed += stashGasUsed

	// Call the internal _delegate function
	result, delegateGasUsed, err := handleInternalDelegate(evm, caller, toValidatorID, value)
	if err != nil {
		return result, gasUsed + delegateGasUsed, err
	}

	// Add the gas used by handleInternalDelegate
	gasUsed += delegateGasUsed

	// Sync validator
	result, syncGasUsed, err := handleSyncValidator(evm, toValidatorID)
	if err != nil {
		return result, gasUsed + syncGasUsed, err
	}

	// Add the gas used by handleSyncValidator
	gasUsed += syncGasUsed

	// Emit Delegated event
	topics := []common.Hash{
		SfcAbi.Events["Delegated"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
	}
	data, err := SfcAbi.Events["Delegated"].Inputs.NonIndexed().Pack(
		value,
	)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	// Recount votes
	// Get the validator auth address
	validatorAuthSlot, _ := getValidatorAuthSlot(toValidatorID)
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorAuthSlot)))
	validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

	// Call handleRecountVotes with strict=false
	result, recountGasUsed, err := handleRecountVotes(evm, caller, validatorAuthAddr, false)
	if err != nil {
		// We don't return an error here because recountVotes is not critical
		// and we don't want to revert the transaction if it fails
	}

	// Add the gas used by handleRecountVotes
	gasUsed += recountGasUsed

	return nil, gasUsed, nil
}

// handleUndelegate undelegates stake from a validator
func handleUndelegate(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// Get the arguments
	if len(args) != 3 {
		return nil, 0, vm.ErrExecutionReverted
	}
	toValidatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	wrID, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	amount, ok := args[2].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the validator exists
	revertData, checkGasUsed, err := checkValidatorExists(evm, toValidatorID, "undelegate")
	gasUsed += checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Check that the delegation exists
	revertData, checkGasUsed, err = checkDelegationExists(evm, caller, toValidatorID, "undelegate")
	gasUsed += checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Stash rewards
	// Create arguments for handleStashRewards
	stashRewardsArgs := []interface{}{caller, toValidatorID}
	// Call handleStashRewards
	result, stashGasUsed, err := handleStashRewards(evm, stashRewardsArgs)
	if err != nil {
		return result, gasUsed + stashGasUsed, err
	}

	// Add the gas used by handleStashRewards
	gasUsed += stashGasUsed

	// Check that the amount is greater than 0
	if amount.Cmp(big.NewInt(0)) <= 0 {
		revertData, err := encodeRevertReason("undelegate", "zero amount")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the amount is less than or equal to the unlocked stake
	// Create arguments for handleGetUnlockedStake
	getUnlockedStakeArgs := []interface{}{caller, toValidatorID}
	// Call handleGetUnlockedStake
	unlockedStakeResult, unlockGasUsed, err := handleGetUnlockedStake(evm, getUnlockedStakeArgs)
	if err != nil {
		return unlockedStakeResult, gasUsed + unlockGasUsed, err
	}

	// Add the gas used by handleGetUnlockedStake
	gasUsed += unlockGasUsed

	// Unpack the result
	unlockedStakeValues, err := SfcAbi.Methods["getUnlockedStake"].Outputs.Unpack(unlockedStakeResult)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// The result should be a single *big.Int value
	if len(unlockedStakeValues) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}

	unlockedStake, ok := unlockedStakeValues[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	if amount.Cmp(unlockedStake) > 0 {
		revertData, err := encodeRevertReason("undelegate", "not enough unlocked stake")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the delegator is allowed to withdraw
	allowed, err := handleCheckAllowedToWithdraw(evm, caller, toValidatorID)
	if err != nil {
		// This is a direct call, not through a handler, so we don't have a revert reason
		return nil, gasUsed, err
	}
	if !allowed {
		revertData, err := encodeRevertReason("undelegate", "outstanding sU2U balance")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the withdrawal request ID doesn't already exist
	withdrawalRequestSlot, _ := getWithdrawalRequestSlot(caller, toValidatorID, wrID)
	withdrawalRequest := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalRequestSlot)))
	withdrawalRequestAmount := new(big.Int).SetBytes(withdrawalRequest.Bytes())
	if withdrawalRequestAmount.Cmp(big.NewInt(0)) != 0 {
		revertData, err := encodeRevertReason("undelegate", "wrID already exists")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Raw undelegate
	result, rawGasUsed, err := handleRawUndelegate(evm, caller, toValidatorID, amount, true)
	if err != nil {
		return result, gasUsed + rawGasUsed, err
	}

	// Add the gas used by handleRawUndelegate
	gasUsed += rawGasUsed

	// Set the withdrawal request
	withdrawalRequestAmountSlot, slotGasUsed := getWithdrawalRequestAmountSlot(caller, toValidatorID, wrID)
	gasUsed += slotGasUsed
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalRequestAmountSlot)), common.BigToHash(amount))

	withdrawalRequestEpochSlot, epochSlotGasUsed := getWithdrawalRequestEpochSlot(caller, toValidatorID, wrID)
	gasUsed += epochSlotGasUsed
	currentEpochBigInt, _, err := getCurrentEpoch(evm)
	if err != nil {
		return nil, 0, err
	}
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalRequestEpochSlot)), common.BigToHash(currentEpochBigInt))

	withdrawalRequestTimeSlot, timeSlotGasUsed := getWithdrawalRequestTimeSlot(caller, toValidatorID, wrID)
	gasUsed += timeSlotGasUsed
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalRequestTimeSlot)), common.BigToHash(evm.Context.Time))

	// Sync validator
	result, syncGasUsed, err := handleSyncValidator(evm, toValidatorID)
	if err != nil {
		return result, gasUsed + syncGasUsed, err
	}

	// Add the gas used by handleSyncValidator
	gasUsed += syncGasUsed

	// Emit Undelegated event
	topics := []common.Hash{
		SfcAbi.Events["Undelegated"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
		common.BigToHash(wrID),                                      // indexed parameter (wrID)
	}
	data, err := SfcAbi.Events["Undelegated"].Inputs.NonIndexed().Pack(
		amount,
	)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	// Recount votes
	// Get the validator auth address
	validatorAuthSlot, _ := getValidatorAuthSlot(toValidatorID)
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorAuthSlot)))
	validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

	// Call handleRecountVotes with strict=true
	result, recountGasUsed, err := handleRecountVotes(evm, caller, validatorAuthAddr, true)
	if err != nil {
		return result, gasUsed + recountGasUsed, err
	}

	// Add the gas used by handleRecountVotes
	gasUsed += recountGasUsed

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
	// Get the total stake
	totalStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)))
	totalStakeBigInt := new(big.Int).SetBytes(totalStake.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["getTotalStake"].Outputs.Pack(totalStakeBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// handleGetTotalActiveStake returns the total active stake
func handleGetTotalActiveStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the total active stake
	totalActiveStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
	totalActiveStakeBigInt := new(big.Int).SetBytes(totalActiveStake.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["getTotalActiveStake"].Outputs.Pack(totalActiveStakeBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

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
	// Get the current sealed epoch
	currentSealedEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)))
	currentSealedEpochBigInt := new(big.Int).SetBytes(currentSealedEpoch.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["currentSealedEpoch"].Outputs.Pack(currentSealedEpochBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// handleGetLastValidatorID returns the last validator ID
func handleGetLastValidatorID(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the last validator ID
	lastValidatorID := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lastValidatorIDSlot)))
	lastValidatorIDBigInt := new(big.Int).SetBytes(lastValidatorID.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["lastValidatorID"].Outputs.Pack(lastValidatorIDBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

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

// handleClaimRewards claims the rewards for a delegator
func handleClaimRewards(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	toValidatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the validator exists
	revertData, checkGasUsed, err := checkValidatorExists(evm, toValidatorID, "claimRewards")
	gasUsed = checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Check that the delegator is allowed to withdraw
	allowed, err := handleCheckAllowedToWithdraw(evm, caller, toValidatorID)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}
	if !allowed {
		revertData, err := encodeRevertReason("claimRewards", "outstanding sU2U balance")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Stash the rewards
	// Create arguments for handleStashRewards
	stashRewardsArgs := []interface{}{caller, toValidatorID}
	// Call handleStashRewards
	_, _, err = handleStashRewards(evm, stashRewardsArgs)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the rewards
	rewardsStashSlot, getGasUsed := getRewardsStashSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	rewardsStash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(rewardsStashSlot)))
	rewardsStashBigInt := new(big.Int).SetBytes(rewardsStash.Bytes())

	// Check that the rewards are not zero
	if rewardsStashBigInt.Cmp(big.NewInt(0)) == 0 {
		revertData, err := encodeRevertReason("claimRewards", "zero rewards")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Clear the rewards stash
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(rewardsStashSlot)), common.BigToHash(big.NewInt(0)))

	// Mint the native token
	// TODO: Implement _mintNativeToken

	// Transfer the rewards to the delegator
	evm.SfcStateDB.AddBalance(caller, rewardsStashBigInt)

	// Emit ClaimedRewards event
	// TODO: Split the rewards into lockupExtraReward, lockupBaseReward, and unlockedReward
	topics := []common.Hash{
		SfcAbi.Events["ClaimedRewards"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
	}
	data, err := SfcAbi.Events["ClaimedRewards"].Inputs.NonIndexed().Pack(
		big.NewInt(0),      // lockupExtraReward
		big.NewInt(0),      // lockupBaseReward
		rewardsStashBigInt, // unlockedReward
	)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleRestakeRewards restakes the rewards for a delegator
func handleRestakeRewards(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	toValidatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the validator exists
	revertData, checkGasUsed, err := checkValidatorExists(evm, toValidatorID, "restakeRewards")
	gasUsed = checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Check that the delegator is allowed to withdraw
	allowed, err := handleCheckAllowedToWithdraw(evm, caller, toValidatorID)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}
	if !allowed {
		revertData, err := encodeRevertReason("restakeRewards", "outstanding sU2U balance")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Stash the rewards
	// Create arguments for handleStashRewards
	stashRewardsArgs := []interface{}{caller, toValidatorID}
	// Call handleStashRewards
	_, _, err = handleStashRewards(evm, stashRewardsArgs)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the rewards
	rewardsStashSlot, getGasUsed := getRewardsStashSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	rewardsStash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(rewardsStashSlot)))
	rewardsStashBigInt := new(big.Int).SetBytes(rewardsStash.Bytes())

	// Check that the rewards are not zero
	if rewardsStashBigInt.Cmp(big.NewInt(0)) == 0 {
		revertData, err := encodeRevertReason("restakeRewards", "zero rewards")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Clear the rewards stash
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(rewardsStashSlot)), common.BigToHash(big.NewInt(0)))

	// Mint the native token
	// TODO: Implement _mintNativeToken

	// Delegate the rewards
	// Get the delegation stake
	stakeSlot, getGasUsed := getStakeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlot)))
	stakeBigInt := new(big.Int).SetBytes(stake.Bytes())

	// Update the stake
	newStake := new(big.Int).Add(stakeBigInt, rewardsStashBigInt)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlot)), common.BigToHash(newStake))

	// Update the validator's received stake
	validatorReceivedStakeSlot, getGasUsed := getValidatorReceivedStakeSlot(toValidatorID)
	gasUsed += getGasUsed
	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)))
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())
	newReceivedStake := new(big.Int).Add(receivedStakeBigInt, rewardsStashBigInt)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)), common.BigToHash(newReceivedStake))

	// Update the total stake
	totalStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)))
	totalStakeBigInt := new(big.Int).SetBytes(totalStakeState.Bytes())
	newTotalStake := new(big.Int).Add(totalStakeBigInt, rewardsStashBigInt)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)), common.BigToHash(newTotalStake))

	// Update the total active stake if the validator is active
	validatorStatusSlot, getGasUsed := getValidatorStatusSlot(toValidatorID)
	gasUsed += getGasUsed
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)))
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())
	if validatorStatusBigInt.Cmp(big.NewInt(0)) == 0 { // OK_STATUS
		totalActiveStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
		totalActiveStakeBigInt := new(big.Int).SetBytes(totalActiveStakeState.Bytes())
		newTotalActiveStake := new(big.Int).Add(totalActiveStakeBigInt, rewardsStashBigInt)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)), common.BigToHash(newTotalActiveStake))
	}

	// Update the locked stake
	// TODO: Split the rewards into lockupExtraReward, lockupBaseReward, and unlockedReward
	lockedStakeSlot, getGasUsed := getLockedStakeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lockedStakeSlot)))
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())
	newLockedStake := new(big.Int).Add(lockedStakeBigInt, rewardsStashBigInt)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(lockedStakeSlot)), common.BigToHash(newLockedStake))

	// Emit RestakedRewards event
	topics := []common.Hash{
		SfcAbi.Events["RestakedRewards"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
	}
	data, err := SfcAbi.Events["RestakedRewards"].Inputs.NonIndexed().Pack(
		big.NewInt(0),      // lockupExtraReward
		big.NewInt(0),      // lockupBaseReward
		rewardsStashBigInt, // unlockedReward
	)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleLockStake locks a delegation
func handleLockStake(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the arguments
	if len(args) != 3 {
		return nil, 0, vm.ErrExecutionReverted
	}
	toValidatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	lockupDuration, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	amount, ok := args[2].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the validator exists
	revertData, checkGasUsed, err := checkValidatorExists(evm, toValidatorID, "lockStake")
	gasUsed = checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Check that the validator is active
	revertData, checkGasUsed, err = checkValidatorActive(evm, toValidatorID, "lockStake")
	gasUsed += checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Check that the amount is greater than 0
	if amount.Cmp(big.NewInt(0)) <= 0 {
		revertData, err := encodeRevertReason("lockStake", "zero amount")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the delegation is not already locked up
	lockupFromEpochSlot, getGasUsed := getLockupFromEpochSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockupFromEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lockupFromEpochSlot)))
	lockupFromEpochBigInt := new(big.Int).SetBytes(lockupFromEpoch.Bytes())

	lockupEndTimeSlot, getGasUsed := getLockupEndTimeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockupEndTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lockupEndTimeSlot)))
	lockupEndTimeBigInt := new(big.Int).SetBytes(lockupEndTime.Bytes())

	if lockupFromEpochBigInt.Cmp(big.NewInt(0)) != 0 && lockupEndTimeBigInt.Cmp(evm.Context.Time) > 0 {
		revertData, err := encodeRevertReason("lockStake", "already locked up")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the lockup duration is valid
	minLockupDurationBigInt, _, err := getMinLockupDuration(evm)
	if err != nil {
		return nil, 0, err
	}

	maxLockupDurationBigInt, _, err := getMaxLockupDuration(evm)
	if err != nil {
		return nil, 0, err
	}

	if lockupDuration.Cmp(minLockupDurationBigInt) < 0 || lockupDuration.Cmp(maxLockupDurationBigInt) > 0 {
		revertData, err := encodeRevertReason("lockStake", "incorrect duration")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the validator's lockup period will not end earlier
	validatorAuthSlot, getGasUsed := getValidatorAuthSlot(toValidatorID)
	gasUsed += getGasUsed
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorAuthSlot)))
	validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

	endTime := new(big.Int).Add(evm.Context.Time, lockupDuration)

	if caller.Cmp(validatorAuthAddr) != 0 {
		validatorLockupEndTimeSlot, getGasUsed := getLockupEndTimeSlot(validatorAuthAddr, toValidatorID)
		gasUsed += getGasUsed
		validatorLockupEndTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorLockupEndTimeSlot)))
		validatorLockupEndTimeBigInt := new(big.Int).SetBytes(validatorLockupEndTime.Bytes())

		if validatorLockupEndTimeBigInt.Cmp(endTime) < 0 {
			revertData, err := encodeRevertReason("lockStake", "validator lockup period will end earlier")
			if err != nil {
				return nil, 0, vm.ErrExecutionReverted
			}
			return revertData, 0, vm.ErrExecutionReverted
		}
	}

	// Stash rewards
	// Create arguments for handleStashRewards
	stashRewardsArgs := []interface{}{caller, toValidatorID}
	// Call handleStashRewards
	_, _, err = handleStashRewards(evm, stashRewardsArgs)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the lockup duration is not decreasing
	lockupDurationSlot, getGasUsed := getLockupDurationSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockupDurationState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lockupDurationSlot)))
	lockupDurationStateBigInt := new(big.Int).SetBytes(lockupDurationState.Bytes())

	if lockupDuration.Cmp(lockupDurationStateBigInt) < 0 {
		revertData, err := encodeRevertReason("lockStake", "lockup duration cannot decrease")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the amount is not greater than the unlocked stake
	// Create arguments for handleGetUnlockedStake
	getUnlockedStakeArgs := []interface{}{caller, toValidatorID}
	// Call handleGetUnlockedStake
	unlockedStakeResult, _, err := handleGetUnlockedStake(evm, getUnlockedStakeArgs)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Unpack the result
	unlockedStakeValues, err := SfcAbi.Methods["getUnlockedStake"].Outputs.Unpack(unlockedStakeResult)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// The result should be a single *big.Int value
	if len(unlockedStakeValues) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}

	unlockedStake, ok := unlockedStakeValues[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	if amount.Cmp(unlockedStake) > 0 {
		revertData, err := encodeRevertReason("lockStake", "not enough stake")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the locked stake
	lockedStakeSlot, getGasUsed := getLockedStakeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lockedStakeSlot)))
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())
	newLockedStake := new(big.Int).Add(lockedStakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(lockedStakeSlot)), common.BigToHash(newLockedStake))

	// Update the lockup info
	currentEpochBigInt, _, err := getCurrentEpoch(evm)
	if err != nil {
		return nil, 0, err
	}
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(lockupFromEpochSlot)), common.BigToHash(currentEpochBigInt))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(lockupEndTimeSlot)), common.BigToHash(endTime))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(lockupDurationSlot)), common.BigToHash(lockupDuration))

	// Emit LockedUpStake event
	topics := []common.Hash{
		SfcAbi.Events["LockedUpStake"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
	}
	data, err := SfcAbi.Events["LockedUpStake"].Inputs.NonIndexed().Pack(
		lockupDuration,
		amount,
	)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleWithdraw withdraws a delegation
func handleWithdraw(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the arguments
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	toValidatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	wrID, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the validator exists
	revertData, checkGasUsed, err := checkValidatorExists(evm, toValidatorID, "withdraw")
	gasUsed = checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Check that the withdrawal request exists
	withdrawalRequestSlot, getGasUsed := getWithdrawalRequestSlot(caller, toValidatorID, wrID)
	gasUsed += getGasUsed
	withdrawalRequest := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalRequestSlot)))
	withdrawalRequestAmount := new(big.Int).SetBytes(withdrawalRequest.Bytes())
	if withdrawalRequestAmount.Cmp(big.NewInt(0)) == 0 {
		revertData, err := encodeRevertReason("withdraw", "request doesn't exist")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the delegator is allowed to withdraw
	// TODO: Implement _checkAllowedToWithdraw

	// Get the request time and epoch
	withdrawalRequestTimeSlot, getGasUsed := getWithdrawalRequestTimeSlot(caller, toValidatorID, wrID)
	gasUsed += getGasUsed
	withdrawalRequestTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalRequestTimeSlot)))
	withdrawalRequestTimeBigInt := new(big.Int).SetBytes(withdrawalRequestTime.Bytes())

	withdrawalRequestEpochSlot, getGasUsed := getWithdrawalRequestEpochSlot(caller, toValidatorID, wrID)
	gasUsed += getGasUsed
	withdrawalRequestEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalRequestEpochSlot)))
	withdrawalRequestEpochBigInt := new(big.Int).SetBytes(withdrawalRequestEpoch.Bytes())

	// Check if the validator is deactivated
	validatorDeactivatedTimeSlot, getGasUsed := getValidatorDeactivatedTimeSlot(toValidatorID)
	gasUsed += getGasUsed
	validatorDeactivatedTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorDeactivatedTimeSlot)))
	validatorDeactivatedTimeBigInt := new(big.Int).SetBytes(validatorDeactivatedTime.Bytes())

	validatorDeactivatedEpochSlot, getGasUsed := getValidatorDeactivatedEpochSlot(toValidatorID)
	gasUsed += getGasUsed
	validatorDeactivatedEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorDeactivatedEpochSlot)))
	validatorDeactivatedEpochBigInt := new(big.Int).SetBytes(validatorDeactivatedEpoch.Bytes())

	requestTime := withdrawalRequestTimeBigInt
	requestEpoch := withdrawalRequestEpochBigInt
	if validatorDeactivatedTimeBigInt.Cmp(big.NewInt(0)) != 0 && validatorDeactivatedTimeBigInt.Cmp(withdrawalRequestTimeBigInt) < 0 {
		requestTime = validatorDeactivatedTimeBigInt
		requestEpoch = validatorDeactivatedEpochBigInt
	}

	// Check that enough time has passed
	withdrawalPeriodTimeBigInt, _, err := getWithdrawalPeriodTime(evm)
	if err != nil {
		return nil, 0, err
	}

	if evm.Context.Time.Cmp(new(big.Int).Add(requestTime, withdrawalPeriodTimeBigInt)) < 0 {
		revertData, err := encodeRevertReason("withdraw", "not enough time passed")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that enough epochs have passed
	withdrawalPeriodEpochsBigInt, _, err := getWithdrawalPeriodEpochs(evm)
	if err != nil {
		return nil, 0, err
	}

	currentEpochBigInt, _, err := getCurrentEpoch(evm)
	if err != nil {
		return nil, 0, err
	}

	if currentEpochBigInt.Cmp(new(big.Int).Add(requestEpoch, withdrawalPeriodEpochsBigInt)) < 0 {
		revertData, err := encodeRevertReason("withdraw", "not enough epochs passed")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Get the amount
	withdrawalRequestAmountSlot, slotGasUsed := getWithdrawalRequestAmountSlot(caller, toValidatorID, wrID)
	gasUsed += slotGasUsed
	withdrawalRequestAmount = evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalRequestAmountSlot))).Big()
	gasUsed += SloadGasCost
	amount := new(big.Int).SetBytes(withdrawalRequestAmount.Bytes())

	// Check if the validator is slashed
	validatorStatusSlot, getGasUsed := getValidatorStatusSlot(toValidatorID)
	gasUsed += getGasUsed
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)))
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())
	isCheater := (validatorStatusBigInt.Bit(7) == 1) // DOUBLESIGN_BIT

	// Calculate the penalty
	penalty := big.NewInt(0)
	if isCheater {
		// TODO: Implement getSlashingPenalty
		penalty = new(big.Int).Div(amount, big.NewInt(2)) // Placeholder: 50% penalty
	}

	// Delete the withdrawal request
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalRequestAmountSlot)), common.BigToHash(big.NewInt(0)))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalRequestEpochSlot)), common.BigToHash(big.NewInt(0)))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalRequestTimeSlot)), common.BigToHash(big.NewInt(0)))

	// Update the total slashed stake
	totalSlashedStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalSlashedStakeSlot)))
	totalSlashedStakeBigInt := new(big.Int).SetBytes(totalSlashedStakeState.Bytes())
	newTotalSlashedStake := new(big.Int).Add(totalSlashedStakeBigInt, penalty)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalSlashedStakeSlot)), common.BigToHash(newTotalSlashedStake))

	// Check that the stake is not fully slashed
	if amount.Cmp(penalty) <= 0 {
		revertData, err := encodeRevertReason("withdraw", "stake is fully slashed")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Transfer the amount minus the penalty to the delegator
	amountToTransfer := new(big.Int).Sub(amount, penalty)
	evm.SfcStateDB.AddBalance(caller, amountToTransfer)

	// Burn the penalty
	// TODO: Implement _burnU2U

	// Emit Withdrawn event
	topics := []common.Hash{
		SfcAbi.Events["Withdrawn"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
		common.BigToHash(wrID),                                      // indexed parameter (wrID)
	}
	data, err := SfcAbi.Events["Withdrawn"].Inputs.NonIndexed().Pack(
		amount,
	)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, gasUsed, nil
}

// CurrentEpoch returns the current epoch
func handleCurrentEpoch(evm *vm.EVM) ([]byte, uint64, error) {
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

// IsLockedUp returns whether a delegator's stake is locked up for a validator
func handleIsLockedUp(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement isLockedUp handler
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

// MintU2U mints U2U tokens
func handleMintU2U(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement mintU2U handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// BurnU2U burns U2U tokens
func handleBurnU2U(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement burnU2U handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// SealEpoch seals the current epoch
func handleSealEpoch(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Check if caller is the NodeDriverAuth contract (onlyDriver modifier)
	revertData, checkGasUsed, err := checkOnlyDriver(evm, caller, "sealEpoch")
	gasUsed += checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Get the arguments
	if len(args) != 5 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	offlineTimes, ok := args[0].([]*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	offlineBlocks, ok := args[1].([]*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	uptimes, ok := args[2].([]*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	originatedTxsFee, ok := args[3].([]*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	epochGas, ok := args[4].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}
	// Get the current epoch
	currentEpochBigInt, epochGasUsed, err := getCurrentEpoch(evm)
	gasUsed += epochGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Get the epoch snapshot for the current epoch
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(currentEpochBigInt)
	gasUsed += slotGasUsed

	// Get the validator IDs for the current epoch
	// For a dynamic array in a struct, we first get the length from the slot
	validatorIDsSlot := epochSnapshotSlot + validatorIDsOffset
	validatorIDsLengthHash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorIDsSlot)))
	gasUsed += SloadGasCost

	// Convert the length hash to a big.Int
	validatorIDsLength := new(big.Int).SetBytes(validatorIDsLengthHash.Bytes()).Uint64()

	// Calculate the base slot for the array elements
	// The array elements start at keccak256(slot)
	validatorIDsBaseSlotBytes := crypto.Keccak256(common.BigToHash(big.NewInt(validatorIDsSlot)).Bytes())
	gasUsed += HashGasCost
	validatorIDsBaseSlot := new(big.Int).SetBytes(validatorIDsBaseSlotBytes)

	// Read each validator ID from storage
	validatorIDs := make([]*big.Int, 0, validatorIDsLength)
	for i := uint64(0); i < validatorIDsLength; i++ {
		// Calculate the slot for this array element: baseSlot + i
		elementSlot := new(big.Int).Add(validatorIDsBaseSlot, big.NewInt(int64(i)))

		// Get the validator ID from storage
		validatorIDHash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(elementSlot))
		gasUsed += SloadGasCost

		// Convert the hash to a big.Int and add it to the list
		validatorID := new(big.Int).SetBytes(validatorIDHash.Bytes())
		validatorIDs = append(validatorIDs, validatorID)
	}

	// Call _sealEpoch_offline
	offlineGasUsed, err := _sealEpoch_offline(evm, validatorIDs, offlineTimes, offlineBlocks, currentEpochBigInt)
	gasUsed += offlineGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Get the previous epoch snapshot
	currentSealedEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)))
	gasUsed += SloadGasCost
	currentSealedEpochBigInt := new(big.Int).SetBytes(currentSealedEpoch.Bytes())
	prevEpochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(currentSealedEpochBigInt)
	gasUsed += slotGasUsed

	// Get the end time of the previous epoch
	prevEndTimeSlot := prevEpochSnapshotSlot + endTimeOffset
	prevEndTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(prevEndTimeSlot)))
	gasUsed += SloadGasCost
	prevEndTimeBigInt := new(big.Int).SetBytes(prevEndTime.Bytes())

	// Calculate epoch duration
	epochDuration := big.NewInt(1) // Default to 1 if current time <= prevEndTime
	if evm.Context.Time.Cmp(prevEndTimeBigInt) > 0 {
		epochDuration = new(big.Int).Sub(evm.Context.Time, prevEndTimeBigInt)
	}

	// Call _sealEpoch_rewards
	rewardsGasUsed, err := _sealEpoch_rewards(evm, epochDuration, currentEpochBigInt, currentSealedEpochBigInt, validatorIDs, uptimes, originatedTxsFee)
	gasUsed += rewardsGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Call _sealEpoch_minGasPrice
	minGasPriceGasUsed, err := _sealEpoch_minGasPrice(evm, epochDuration, epochGas)
	gasUsed += minGasPriceGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Update currentSealedEpoch
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)), common.BigToHash(currentEpochBigInt))
	gasUsed += SstoreGasCost

	// Update epoch snapshot end time
	endTimeSlot := epochSnapshotSlot + endTimeOffset
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(endTimeSlot)), common.BigToHash(evm.Context.Time))
	gasUsed += SstoreGasCost

	// Get the base reward per second from constants manager
	baseRewardPerSecond, cmGasUsed, err := callConstantManagerMethod(evm, "baseRewardPerSecond")
	gasUsed += cmGasUsed
	if err != nil || len(baseRewardPerSecond) == 0 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}
	baseRewardPerSecondBigInt, ok := baseRewardPerSecond[0].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Update epoch snapshot base reward per second
	baseRewardPerSecondSlot := epochSnapshotSlot + baseRewardPerSecondOffset
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(baseRewardPerSecondSlot)), common.BigToHash(baseRewardPerSecondBigInt))
	gasUsed += SstoreGasCost

	// Get the total supply
	totalSupply := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)))
	gasUsed += SloadGasCost
	totalSupplyBigInt := new(big.Int).SetBytes(totalSupply.Bytes())

	// Update epoch snapshot total supply
	totalSupplySnapshotSlot := epochSnapshotSlot + totalSupplyOffset
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySnapshotSlot)), common.BigToHash(totalSupplyBigInt))
	gasUsed += SstoreGasCost

	return nil, gasUsed, nil
}

// SealEpochValidators seals the validators for the current epoch
func handleSealEpochValidators(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Check if caller is the NodeDriverAuth contract (onlyDriver modifier)
	revertData, checkGasUsed, err := checkOnlyDriver(evm, caller, "sealEpochValidators")
	gasUsed += checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Get the arguments
	if len(args) != 1 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	nextValidatorIDs, ok := args[0].([]*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Get the current epoch
	currentEpochBigInt, epochGasUsed, err := getCurrentEpoch(evm)
	gasUsed += epochGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Get the epoch snapshot slot for the current epoch
	epochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(currentEpochBigInt)
	gasUsed += slotGasUsed

	// Initialize total stake for the snapshot
	totalStake := big.NewInt(0)

	// Fill data for the next snapshot
	for _, validatorID := range nextValidatorIDs {
		// Get the validator's received stake
		validatorReceivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(validatorID)
		gasUsed += slotGasUsed

		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)))
		gasUsed += SloadGasCost
		receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

		// Set the received stake for this validator in the epoch snapshot
		validatorReceivedStakeEpochSlot, slotGasUsed := getEpochValidatorReceivedStakeSlot(currentEpochBigInt, validatorID)
		gasUsed += slotGasUsed

		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeEpochSlot)), common.BigToHash(receivedStakeBigInt))
		gasUsed += SstoreGasCost

		// Add to total stake
		totalStake = new(big.Int).Add(totalStake, receivedStakeBigInt)
	}

	// Set the validator IDs for the epoch snapshot
	// For a dynamic array in a struct, we first set the length at the slot
	validatorIDsSlot := epochSnapshotSlot + validatorIDsOffset
	validatorIDsLength := big.NewInt(int64(len(nextValidatorIDs)))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorIDsSlot)), common.BigToHash(validatorIDsLength))
	gasUsed += SstoreGasCost

	// Calculate the base slot for the array elements
	// The array elements start at keccak256(slot)
	validatorIDsBaseSlotBytes := crypto.Keccak256(common.BigToHash(big.NewInt(validatorIDsSlot)).Bytes())
	gasUsed += HashGasCost
	validatorIDsBaseSlot := new(big.Int).SetBytes(validatorIDsBaseSlotBytes)

	// Store each validator ID in storage
	for i, validatorID := range nextValidatorIDs {
		// Calculate the slot for this array element: baseSlot + i
		elementSlot := new(big.Int).Add(validatorIDsBaseSlot, big.NewInt(int64(i)))

		// Store the validator ID
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(elementSlot), common.BigToHash(validatorID))
		gasUsed += SstoreGasCost
	}

	// Set the total stake for the epoch snapshot
	totalStakeSlot := epochSnapshotSlot + totalStakeOffset
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)), common.BigToHash(totalStake))
	gasUsed += SstoreGasCost

	// Update the minimum gas price in the node
	minGasPrice := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)))
	gasUsed += SloadGasCost
	minGasPriceBigInt := new(big.Int).SetBytes(minGasPrice.Bytes())

	// Call the node to update the minimum gas price
	nodeDriverAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(nodeDriverAuthSlot)))
	gasUsed += SloadGasCost
	nodeDriverAuthAddr := common.BytesToAddress(nodeDriverAuth.Bytes())

	// Pack the function call data
	data, err := NodeDriverAuthAbi.Pack("updateMinGasPrice", minGasPriceBigInt)
	if err != nil {
		log.Error("SFC: Error packing updateMinGasPrice call data", "err", err)
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Call the node driver
	log.Info("SFC: Calling NodeDriverAuth updateMinGasPrice", "minGasPrice", minGasPriceBigInt)
	result, _, err := evm.Call(vm.AccountRef(ContractAddress), nodeDriverAuthAddr, data, 50000, big.NewInt(0))
	if err != nil {
		reason, _ := abi.UnpackRevert(result)
		log.Error("SFC: Error calling NodeDriverAuth method", "method", "updateMinGasPrice", "err", err, "reason", reason)
		return nil, gasUsed, err
	}

	return nil, gasUsed, nil
}

// RelockStake relocks stake for a validator
func handleRelockStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement relockStake handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// UnlockStake unlocks stake for a validator
func handleUnlockStake(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the arguments
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	toValidatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	amount, ok := args[1].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the validator exists
	revertData, checkGasUsed, err := checkValidatorExists(evm, toValidatorID, "unlockStake")
	gasUsed = checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Check that the amount is greater than 0
	if amount.Cmp(big.NewInt(0)) <= 0 {
		revertData, err := encodeRevertReason("unlockStake", "zero amount")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the delegation is locked up
	lockupFromEpochSlot, getGasUsed := getLockupFromEpochSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockupFromEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lockupFromEpochSlot)))
	lockupFromEpochBigInt := new(big.Int).SetBytes(lockupFromEpoch.Bytes())

	lockupEndTimeSlot, getGasUsed := getLockupEndTimeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockupEndTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lockupEndTimeSlot)))
	lockupEndTimeBigInt := new(big.Int).SetBytes(lockupEndTime.Bytes())

	if lockupFromEpochBigInt.Cmp(big.NewInt(0)) == 0 || lockupEndTimeBigInt.Cmp(evm.Context.Time) <= 0 {
		revertData, err := encodeRevertReason("unlockStake", "not locked up")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the amount is not greater than the locked stake
	lockedStakeSlot, getGasUsed := getLockedStakeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lockedStakeSlot)))
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())

	if amount.Cmp(lockedStakeBigInt) > 0 {
		revertData, err := encodeRevertReason("unlockStake", "not enough locked stake")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the delegator is allowed to withdraw
	allowed, err := handleCheckAllowedToWithdraw(evm, caller, toValidatorID)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}
	if !allowed {
		revertData, err := encodeRevertReason("unlockStake", "outstanding sU2U balance")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Stash rewards
	// Create arguments for handleStashRewards
	stashRewardsArgs := []interface{}{caller, toValidatorID}
	// Call handleStashRewards
	_, _, err = handleStashRewards(evm, stashRewardsArgs)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Calculate the penalty
	penalty := big.NewInt(0)
	lockupDurationSlot, getGasUsed := getLockupDurationSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockupDuration := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lockupDurationSlot)))
	lockupDurationBigInt := new(big.Int).SetBytes(lockupDuration.Bytes())

	// Check if the lockup was created before rewards were reduced
	if lockupEndTimeBigInt.Cmp(new(big.Int).Add(lockupDurationBigInt, big.NewInt(1665146565))) < 0 {
		// If it was locked up before rewards have been reduced, then allow to unlock without penalty
		penalty = big.NewInt(0)
	} else {
		// Calculate the penalty
		// TODO: Implement _popDelegationUnlockPenalty
		penalty = new(big.Int).Div(amount, big.NewInt(10)) // Placeholder: 10% penalty
	}

	// Update the locked stake
	newLockedStake := new(big.Int).Sub(lockedStakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(lockedStakeSlot)), common.BigToHash(newLockedStake))

	// Apply the penalty
	if penalty.Cmp(big.NewInt(0)) != 0 {
		// Undelegate the penalty
		// TODO: Implement _rawUndelegate

		// Burn the penalty
		// TODO: Implement _burnU2U
	}

	// Emit UnlockedStake event
	topics := []common.Hash{
		SfcAbi.Events["UnlockedStake"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
	}
	data, err := SfcAbi.Events["UnlockedStake"].Inputs.NonIndexed().Pack(
		amount,
		penalty,
	)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	// Return the penalty
	result, err := SfcAbi.Methods["unlockStake"].Outputs.Pack(penalty)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
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
	// TODO: Implement fallback function handler
	// For empty input (pure native token transfer), we should reject the transaction
	// For non-empty input, we should delegate the call to libAddress

	// In the SFC contract, the fallback function requires msg.data to be non-empty:
	// function() payable external {
	//     require(msg.data.length != 0, "transfers not allowed");
	//     _delegate(libAddress);
	// }

	return nil, 0, vm.ErrSfcFunctionNotImplemented
}
