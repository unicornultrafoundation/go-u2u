package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/params"
)

// handleDelegate delegates stake to a validator
func handleDelegate(evm *vm.EVM, caller common.Address, args []interface{}, value *big.Int) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	toValidatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call handleInternalDelegate which implements the _delegate function logic
	result, delegateGasUsed, err := handleInternalDelegate(evm, caller, toValidatorID, value)
	if err != nil {
		return result, gasUsed + delegateGasUsed, err
	}
	gasUsed += delegateGasUsed

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
	// Create arguments for handle_stashRewards
	stashRewardsArgs := []interface{}{caller, toValidatorID}
	// Call handle_stashRewards
	result, stashGasUsed, err := handle_stashRewards(evm, stashRewardsArgs)
	if err != nil {
		return result, gasUsed + stashGasUsed, err
	}

	// Add the gas used by handle_stashRewards
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
	unlockedStakeValues, err := SfcLibAbi.Methods["getUnlockedStake"].Outputs.Unpack(unlockedStakeResult)
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
	withdrawalRequest := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(withdrawalRequestSlot))
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
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(withdrawalRequestAmountSlot), common.BigToHash(amount))

	withdrawalRequestEpochSlot, epochSlotGasUsed := getWithdrawalRequestEpochSlot(caller, toValidatorID, wrID)
	gasUsed += epochSlotGasUsed
	currentEpochBigInt, _, err := getCurrentEpoch(evm)
	if err != nil {
		return nil, 0, err
	}
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(withdrawalRequestEpochSlot), common.BigToHash(currentEpochBigInt))

	withdrawalRequestTimeSlot, timeSlotGasUsed := getWithdrawalRequestTimeSlot(caller, toValidatorID, wrID)
	gasUsed += timeSlotGasUsed
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(withdrawalRequestTimeSlot), common.BigToHash(evm.Context.Time))

	// Sync validator
	result, syncGasUsed, err := handleSyncValidator(evm, toValidatorID)
	if err != nil {
		return result, gasUsed + syncGasUsed, err
	}

	// Add the gas used by handleSyncValidator
	gasUsed += syncGasUsed

	// Emit Undelegated event
	topics := []common.Hash{
		SfcLibAbi.Events["Undelegated"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
		common.BigToHash(wrID),                                      // indexed parameter (wrID)
	}
	data, err := SfcLibAbi.Events["Undelegated"].Inputs.NonIndexed().Pack(
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
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorAuthSlot))
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

// handleInternalDelegate implements the internal _delegate function logic
func handleInternalDelegate(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int, amount *big.Int) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Check that the validator exists
	revertData, checkGasUsed, err := checkValidatorExists(evm, toValidatorID, "_delegate")
	gasUsed += checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Check that the validator is active
	validatorStatusSlot, slotGasUsed := getValidatorStatusSlot(toValidatorID)
	gasUsed += slotGasUsed
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorStatusSlot))
	gasUsed += SloadGasCost
	statusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())

	// Check if validator is not deactivated
	if statusBigInt.Bit(0) == 1 { // WITHDRAWN_BIT
		revertData, err := encodeRevertReason("_delegate", "validator is deactivated")
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}

	// Check if validator is not offline
	if statusBigInt.Bit(3) == 1 { // OFFLINE_BIT
		revertData, err := encodeRevertReason("_delegate", "validator is offline")
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}

	// Check if validator is not a cheater
	if statusBigInt.Bit(7) == 1 { // DOUBLESIGN_BIT
		revertData, err := encodeRevertReason("_delegate", "validator is a cheater")
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}

	// Call _rawDelegate with strict=true
	result, rawGasUsed, err := handleRawDelegate(evm, delegator, toValidatorID, amount, true)
	if err != nil {
		return result, gasUsed + rawGasUsed, err
	}
	gasUsed += rawGasUsed

	// Check delegated stake limit
	withinLimit, checkGasUsed, err := checkDelegatedStakeLimit(evm, toValidatorID)
	gasUsed += checkGasUsed
	if err != nil {
		return nil, gasUsed, err
	}
	if !withinLimit {
		revertData, err := encodeRevertReason("_delegate", "validator's delegations limit is exceeded")
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}

	return nil, gasUsed, nil
}

// handleRawUndelegate implements the _rawUndelegate function logic
func handleRawUndelegate(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int, amount *big.Int, strict bool) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// Update the stake
	stakeSlot, slotGasUsed := getStakeSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stakeSlot))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	stakeBigInt := new(big.Int).SetBytes(stake.Bytes())
	newStake := new(big.Int).Sub(stakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stakeSlot), common.BigToHash(newStake))
	gasUsed += params.SstoreSetGasEIP2200 // Add gas for SSTORE

	// Update the validator's received stake
	validatorReceivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(toValidatorID)
	gasUsed += slotGasUsed
	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())
	newReceivedStake := new(big.Int).Sub(receivedStakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot), common.BigToHash(newReceivedStake))
	gasUsed += params.SstoreSetGasEIP2200 // Add gas for SSTORE

	// Update the total stake
	totalStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)))
	totalStakeBigInt := new(big.Int).SetBytes(totalStakeState.Bytes())
	newTotalStake := new(big.Int).Sub(totalStakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)), common.BigToHash(newTotalStake))

	// Update the total active stake if the validator is active
	validatorStatusSlot, _ := getValidatorStatusSlot(toValidatorID)
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorStatusSlot))
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())
	if validatorStatusBigInt.Cmp(big.NewInt(0)) == 0 { // OK_STATUS
		totalActiveStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
		totalActiveStakeBigInt := new(big.Int).SetBytes(totalActiveStakeState.Bytes())
		newTotalActiveStake := new(big.Int).Sub(totalActiveStakeBigInt, amount)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)), common.BigToHash(newTotalActiveStake))
	}

	// Get the self-stake
	// Create arguments for handleGetSelfStake
	args := []interface{}{toValidatorID}
	// Call handleGetSelfStake
	result, selfStakeGasUsed, err := handleGetSelfStake(evm, args)
	if err != nil {
		return nil, gasUsed + selfStakeGasUsed, err
	}

	// Add the gas used by handleGetSelfStake
	gasUsed += selfStakeGasUsed

	// Unpack the result
	selfStakeValues, err := SfcLibAbi.Methods["getSelfStake"].Outputs.Unpack(result)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// The result should be a single *big.Int value
	if len(selfStakeValues) != 1 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	selfStake, ok := selfStakeValues[0].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Check if the validator should be deactivated
	if selfStake.Cmp(big.NewInt(0)) == 0 {
		// Set the validator as deactivated
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorStatusSlot), common.BigToHash(big.NewInt(1))) // WITHDRAWN_BIT
	} else if validatorStatusBigInt.Cmp(big.NewInt(0)) == 0 { // OK_STATUS
		// Check that the self-stake is at least the minimum self-stake
		minSelfStakeBigInt := getConstantsManagerVariable("minSelfStake")
		if selfStake.Cmp(minSelfStakeBigInt) < 0 {
			// Set the validator as deactivated
			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorStatusSlot), common.BigToHash(big.NewInt(1))) // WITHDRAWN_BIT
		} else {
			// Check that the delegated stake is within the limit
			withinLimit, err := handleCheckDelegatedStakeLimit(evm, toValidatorID)
			if err != nil {
				return nil, gasUsed, err
			}
			if !withinLimit {
				// Set the validator as deactivated
				evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorStatusSlot), common.BigToHash(big.NewInt(1))) // WITHDRAWN_BIT
			}
		}
	}

	// Get the validator auth address
	validatorAuthSlot, slotGasUsed := getValidatorAuthSlot(toValidatorID)
	gasUsed += slotGasUsed
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorAuthSlot))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

	// Recount votes
	_, recountGasUsed, err := handleRecountVotes(evm, delegator, validatorAuthAddr, strict)
	if err != nil && strict {
		return nil, gasUsed + recountGasUsed, err
	}

	// Add the gas used by handleRecountVotes
	gasUsed += recountGasUsed

	return nil, gasUsed, nil
}

// handleRawDelegate implements the _rawDelegate function logic from SFCLib.sol
func handleRawDelegate(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int, amount *big.Int, strict bool) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Check that the amount is greater than 0
	if amount.Cmp(big.NewInt(0)) <= 0 {
		revertData, err := encodeRevertReason("_rawDelegate", "zero amount")
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}

	// Stash rewards
	result, stashGasUsed, err := handle_stashRewards(evm, []interface{}{delegator, toValidatorID})
	if err != nil {
		return result, gasUsed + stashGasUsed, err
	}
	gasUsed += stashGasUsed

	// Update the stake
	stakeSlot, slotGasUsed := getStakeSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stakeSlot))
	gasUsed += SloadGasCost
	stakeBigInt := new(big.Int).SetBytes(stake.Bytes())
	newStake := new(big.Int).Add(stakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stakeSlot), common.BigToHash(newStake))
	gasUsed += SstoreGasCost

	// Update the validator's received stake
	validatorReceivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(toValidatorID)
	gasUsed += slotGasUsed
	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot))
	gasUsed += SloadGasCost
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())
	origStake := new(big.Int).Set(receivedStakeBigInt) // Save original stake for _syncValidator
	newReceivedStake := new(big.Int).Add(receivedStakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot), common.BigToHash(newReceivedStake))
	gasUsed += SstoreGasCost

	// Update the total stake
	totalStakeSlot := common.BigToHash(big.NewInt(totalStakeSlot))
	totalStake := evm.SfcStateDB.GetState(ContractAddress, totalStakeSlot)
	gasUsed += SloadGasCost
	totalStakeBigInt := new(big.Int).SetBytes(totalStake.Bytes())
	newTotalStake := new(big.Int).Add(totalStakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, totalStakeSlot, common.BigToHash(newTotalStake))
	gasUsed += SstoreGasCost

	// Update the total active stake if the validator is active
	validatorStatusSlot, slotGasUsed := getValidatorStatusSlot(toValidatorID)
	gasUsed += slotGasUsed
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorStatusSlot))
	gasUsed += SloadGasCost
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())
	if validatorStatusBigInt.Cmp(big.NewInt(0)) == 0 { // OK_STATUS
		totalActiveStakeSlot := common.BigToHash(big.NewInt(totalActiveStakeSlot))
		totalActiveStake := evm.SfcStateDB.GetState(ContractAddress, totalActiveStakeSlot)
		gasUsed += SloadGasCost
		totalActiveStakeBigInt := new(big.Int).SetBytes(totalActiveStake.Bytes())
		newTotalActiveStake := new(big.Int).Add(totalActiveStakeBigInt, amount)
		evm.SfcStateDB.SetState(ContractAddress, totalActiveStakeSlot, common.BigToHash(newTotalActiveStake))
		gasUsed += SstoreGasCost
	}

	// Sync validator
	result, syncGasUsed, err := handleSyncValidator(evm, toValidatorID, origStake.Cmp(big.NewInt(0)) == 0)
	if err != nil {
		return result, gasUsed + syncGasUsed, err
	}
	gasUsed += syncGasUsed

	// Emit Delegated event
	topics := []common.Hash{
		SfcLibAbi.Events["Delegated"].ID,
		common.BytesToHash(common.LeftPadBytes(delegator.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                                // indexed parameter (toValidatorID)
	}
	data := common.BigToHash(amount).Bytes() // amount

	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	// Get the validator auth address
	validatorAuthSlot, slotGasUsed := getValidatorAuthSlot(toValidatorID)
	gasUsed += slotGasUsed
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorAuthSlot))
	gasUsed += SloadGasCost
	validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

	// Recount votes
	result, recountGasUsed, err := handleRecountVotes(evm, delegator, validatorAuthAddr, strict)
	if err != nil {
		return result, gasUsed + recountGasUsed, err
	}
	gasUsed += recountGasUsed

	return nil, gasUsed, nil
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
	log.Info("handleLockStake done parsing args", "args", args)
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
			return nil, 0, err
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the delegation is not already locked up
	lockupFromEpochSlot, getGasUsed := getLockupFromEpochSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockupFromEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupFromEpochSlot))
	lockupFromEpochBigInt := new(big.Int).SetBytes(lockupFromEpoch.Bytes())

	lockupEndTimeSlot, getGasUsed := getLockupEndTimeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockupEndTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupEndTimeSlot))
	lockupEndTimeBigInt := new(big.Int).SetBytes(lockupEndTime.Bytes())

	if lockupFromEpochBigInt.Cmp(big.NewInt(0)) != 0 && lockupEndTimeBigInt.Cmp(evm.Context.Time) > 0 {
		revertData, err := encodeRevertReason("lockStake", "already locked up")
		if err != nil {
			return nil, 0, err
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the lockup duration is valid
	minLockupDurationBigInt := getConstantsManagerVariable("minLockupDuration")
	maxLockupDurationBigInt := getConstantsManagerVariable("maxLockupDuration")

	if lockupDuration.Cmp(minLockupDurationBigInt) < 0 || lockupDuration.Cmp(maxLockupDurationBigInt) > 0 {
		revertData, err := encodeRevertReason("lockStake", "incorrect duration")
		if err != nil {
			return nil, 0, err
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the validator's lockup period will not end earlier
	validatorAuthSlot, getGasUsed := getValidatorAuthSlot(toValidatorID)
	gasUsed += getGasUsed
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorAuthSlot))
	validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

	endTime := new(big.Int).Add(evm.Context.Time, lockupDuration)

	if caller.Cmp(validatorAuthAddr) != 0 {
		validatorLockupEndTimeSlot, getGasUsed := getLockupEndTimeSlot(validatorAuthAddr, toValidatorID)
		gasUsed += getGasUsed
		validatorLockupEndTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorLockupEndTimeSlot))
		validatorLockupEndTimeBigInt := new(big.Int).SetBytes(validatorLockupEndTime.Bytes())

		if validatorLockupEndTimeBigInt.Cmp(endTime) < 0 {
			revertData, err := encodeRevertReason("lockStake", "validator lockup period will end earlier")
			if err != nil {
				return nil, 0, err
			}
			return revertData, 0, vm.ErrExecutionReverted
		}
	}

	// Stash rewards
	// Create arguments for handle_stashRewards
	stashRewardsArgs := []interface{}{caller, toValidatorID}
	// Call handle_stashRewards
	revertData, _, err = handle_stashRewards(evm, stashRewardsArgs)
	if err != nil {
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the lockup duration is not decreasing
	lockupDurationSlot, getGasUsed := getLockupDurationSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockupDurationState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupDurationSlot))
	lockupDurationStateBigInt := new(big.Int).SetBytes(lockupDurationState.Bytes())

	if lockupDuration.Cmp(lockupDurationStateBigInt) < 0 {
		revertData, err := encodeRevertReason("lockStake", "lockup duration cannot decrease")
		if err != nil {
			return nil, 0, err
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the amount is not greater than the unlocked stake
	// Create arguments for handleGetUnlockedStake
	getUnlockedStakeArgs := []interface{}{caller, toValidatorID}
	// Call handleGetUnlockedStake
	unlockedStakeResult, _, err := handleGetUnlockedStake(evm, getUnlockedStakeArgs)
	if err != nil {
		log.Error("lockStake: handleGetUnlockedStake failed", "err", err)
		return unlockedStakeResult, 0, vm.ErrExecutionReverted
	}

	// Unpack the result
	unlockedStakeValues, err := SfcLibAbi.Methods["getUnlockedStake"].Outputs.Unpack(unlockedStakeResult)
	if err != nil {
		log.Error("lockStake: unpack getUnlockedStake failed", "err", err)
		return nil, 0, err
	}

	unlockedStake, ok := unlockedStakeValues[0].(*big.Int)
	if !ok {
		log.Error("lockStake: unpack unlockedStake failed", "err", err)
		return nil, 0, err
	}

	if amount.Cmp(unlockedStake) > 0 {
		revertData, err := encodeRevertReason("lockStake", "not enough stake")
		if err != nil {
			return nil, 0, err
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the locked stake
	lockedStakeSlot, getGasUsed := getLockedStakeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockedStakeSlot))
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())
	newLockedStake := new(big.Int).Add(lockedStakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockedStakeSlot), common.BigToHash(newLockedStake))

	// Update the lockup info
	currentEpochBigInt, _, err := getCurrentEpoch(evm)
	if err != nil {
		return nil, 0, err
	}
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupFromEpochSlot), common.BigToHash(currentEpochBigInt))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupEndTimeSlot), common.BigToHash(endTime))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupDurationSlot), common.BigToHash(lockupDuration))

	// Emit LockedUpStake event
	topics := []common.Hash{
		SfcLibAbi.Events["LockedUpStake"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
	}
	data, err := SfcLibAbi.Events["LockedUpStake"].Inputs.NonIndexed().Pack(
		lockupDuration,
		amount,
	)
	if err != nil {
		return nil, 0, err
	}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}

// handleRelockStake implements the relockStake function from SFCLib.sol
func handleRelockStake(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
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

	// Call handleLockStake directly without the additional checks that are in lockStake
	result, lockGasUsed, err := handleLockStake(evm, caller, []interface{}{toValidatorID, lockupDuration, amount})
	if err != nil {
		return result, gasUsed + lockGasUsed, err
	}
	gasUsed += lockGasUsed

	return nil, gasUsed, nil
}

// _popDelegationUnlockPenalty calculates and updates the penalty for unlocking a delegation
func _popDelegationUnlockPenalty(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int, unlockAmount *big.Int, totalAmount *big.Int) (*big.Int, uint64, error) {
	var gasUsed uint64 = 0

	// Get the stashed lockup extra reward
	extraRewardSlot, slotGasUsed := getStashedLockupExtraRewardSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	extraReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(extraRewardSlot))
	gasUsed += SloadGasCost

	// Get the stashed lockup base reward
	baseRewardSlot, slotGasUsed := getStashedLockupBaseRewardSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	baseReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(baseRewardSlot))
	gasUsed += SloadGasCost

	// Calculate shares
	extraRewardShare := new(big.Int).Mul(new(big.Int).SetBytes(extraReward.Bytes()), unlockAmount)
	extraRewardShare = extraRewardShare.Div(extraRewardShare, totalAmount)

	baseRewardShare := new(big.Int).Mul(new(big.Int).SetBytes(baseReward.Bytes()), unlockAmount)
	baseRewardShare = baseRewardShare.Div(baseRewardShare, totalAmount)

	// Calculate penalty
	penalty := new(big.Int).Add(extraRewardShare, new(big.Int).Div(baseRewardShare, big.NewInt(2)))

	// Update stashed rewards
	newExtraReward := new(big.Int).Sub(new(big.Int).SetBytes(extraReward.Bytes()), extraRewardShare)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(extraRewardSlot), common.BigToHash(newExtraReward))
	gasUsed += SstoreGasCost

	newBaseReward := new(big.Int).Sub(new(big.Int).SetBytes(baseReward.Bytes()), baseRewardShare)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(baseRewardSlot), common.BigToHash(newBaseReward))
	gasUsed += SstoreGasCost

	// If penalty >= unlockAmount, set penalty = unlockAmount
	if penalty.Cmp(unlockAmount) >= 0 {
		penalty = new(big.Int).Set(unlockAmount)
	}

	return penalty, gasUsed, nil
}

// handleUnlockStake implements the unlockStake function from SFCLib.sol
func handleUnlockStake(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Parse arguments
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

	// Check amount > 0
	if amount.Cmp(big.NewInt(0)) <= 0 {
		revertData, err := encodeRevertReason("unlockStake", "zero amount")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check if stake is locked up
	isLockedUpArgs := []interface{}{caller, toValidatorID}
	isLockedUpResult, _, err := handleIsLockedUp(evm, isLockedUpArgs)
	if err != nil {
		log.Error("unlockStake: handleIsLockedUp failed", "err", err)
		return isLockedUpResult, 0, err
	}
	isLockedUpValues, err := SfcAbi.Methods["isLockedUp"].Outputs.Unpack(isLockedUpResult)
	if err != nil {
		log.Error("unlockStake: unpack isLockedUp failed", "err", err)
		return nil, 0, err
	}
	isLocked, ok := isLockedUpValues[0].(bool)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}
	if !isLocked {
		revertData, err := encodeRevertReason("unlockStake", "not locked up")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Get locked stake info
	lockedStakeSlot, getGasUsed := getLockedStakeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockedStakeSlot))
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())

	// Check amount <= lockedStake
	if amount.Cmp(lockedStakeBigInt) > 0 {
		revertData, err := encodeRevertReason("unlockStake", "not enough locked stake")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check allowed to withdraw
	allowed, err := handleCheckAllowedToWithdraw(evm, caller, toValidatorID)
	if err != nil {
		log.Error("unlockStake: handleCheckAllowedToWithdraw failed", "err", err)
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
	stashRewardsArgs := []interface{}{caller, toValidatorID}
	revertData, _, err := handle_stashRewards(evm, stashRewardsArgs)
	if err != nil {
		log.Error("unlockStake: handle_stashRewards failed", "err", err)
		return revertData, 0, err
	}

	// Calculate penalty
	penalty, penaltyGasUsed, err := _popDelegationUnlockPenalty(evm, caller, toValidatorID, amount, lockedStakeBigInt)
	if err != nil {
		log.Error("unlockStake: _popDelegationUnlockPenalty failed", "err", err)
		return nil, gasUsed + penaltyGasUsed, err
	}
	gasUsed += penaltyGasUsed

	// Check if was locked up before rewards reduction
	lockupEndTimeSlot, getGasUsed := getLockupEndTimeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockupEndTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupEndTimeSlot))
	lockupEndTimeBigInt := new(big.Int).SetBytes(lockupEndTime.Bytes())

	lockupDurationSlot, getGasUsed := getLockupDurationSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockupDuration := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupDurationSlot))
	lockupDurationBigInt := new(big.Int).SetBytes(lockupDuration.Bytes())

	rewardsReductionTime := big.NewInt(1665146565)
	if lockupEndTimeBigInt.Cmp(new(big.Int).Add(lockupDurationBigInt, rewardsReductionTime)) < 0 {
		// If was locked up before rewards reduction, no penalty
		penalty = big.NewInt(0)
	}

	// Update locked stake
	newLockedStake := new(big.Int).Sub(lockedStakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockedStakeSlot), common.BigToHash(newLockedStake))
	gasUsed += SstoreGasCost

	// Apply penalty if any
	if penalty.Cmp(big.NewInt(0)) > 0 {
		// Undelegate the penalty amount
		revertData, _, err = handleRawUndelegate(evm, caller, toValidatorID, penalty, true)
		if err != nil {
			log.Error("unlockStake: handleRawUndelegate failed", "err", err)
			return revertData, gasUsed, err
		}

		// Burn the penalty
		burnU2UArgs := []interface{}{penalty}
		data, burnGasUsed, err := handleBurnU2U(evm, burnU2UArgs)
		if err != nil {
			log.Error("unlockStake: handleBurnU2U failed", "err", err)
			return data, gasUsed + burnGasUsed, err
		}
		gasUsed += burnGasUsed
	}

	// Emit UnlockedStake event
	topics := []common.Hash{
		SfcLibAbi.Events["UnlockedStake"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (validatorID)
	}
	data, err := SfcLibAbi.Events["UnlockedStake"].Inputs.NonIndexed().Pack(
		amount,
		penalty,
	)
	if err != nil {
		log.Error("unlockStake: pack UnlockedStake failed", "err", err)
		return nil, 0, err
	}
	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return penalty.Bytes(), gasUsed, nil
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

	// Check that the withdrawal request exists by checking the epoch field (first field)
	withdrawalRequestEpochSlot, getGasUsed := getWithdrawalRequestEpochSlot(caller, toValidatorID, wrID)
	gasUsed += getGasUsed
	withdrawalRequestEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(withdrawalRequestEpochSlot))
	gasUsed += SloadGasCost
	withdrawalRequestEpochBigInt := new(big.Int).SetBytes(withdrawalRequestEpoch.Bytes())
	if withdrawalRequestEpochBigInt.Cmp(big.NewInt(0)) == 0 {
		revertData, err := encodeRevertReason("withdraw", "request doesn't exist")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that the delegator is allowed to withdraw
	allowed, err := handleCheckAllowedToWithdraw(evm, caller, toValidatorID)
	if err != nil {
		log.Error("withdraw: handleCheckAllowedToWithdraw failed", "err", err)
		return nil, gasUsed, vm.ErrExecutionReverted
	}
	if !allowed {
		revertData, err := encodeRevertReason("withdraw", "outstanding sU2U balance")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Get the request time
	withdrawalRequestTimeSlot, timeGasUsed := getWithdrawalRequestTimeSlot(caller, toValidatorID, wrID)
	gasUsed += timeGasUsed
	withdrawalRequestTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(withdrawalRequestTimeSlot))
	gasUsed += SloadGasCost
	withdrawalRequestTimeBigInt := new(big.Int).SetBytes(withdrawalRequestTime.Bytes())

	// Check if the validator is deactivated
	validatorDeactivatedTimeSlot, getGasUsed := getValidatorDeactivatedTimeSlot(toValidatorID)
	gasUsed += getGasUsed
	validatorDeactivatedTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorDeactivatedTimeSlot))
	validatorDeactivatedTimeBigInt := new(big.Int).SetBytes(validatorDeactivatedTime.Bytes())

	validatorDeactivatedEpochSlot, getGasUsed := getValidatorDeactivatedEpochSlot(toValidatorID)
	gasUsed += getGasUsed
	validatorDeactivatedEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorDeactivatedEpochSlot))
	validatorDeactivatedEpochBigInt := new(big.Int).SetBytes(validatorDeactivatedEpoch.Bytes())

	requestTime := withdrawalRequestTimeBigInt
	requestEpoch := withdrawalRequestEpochBigInt
	if validatorDeactivatedTimeBigInt.Cmp(big.NewInt(0)) != 0 && validatorDeactivatedTimeBigInt.Cmp(withdrawalRequestTimeBigInt) < 0 {
		requestTime = validatorDeactivatedTimeBigInt
		requestEpoch = validatorDeactivatedEpochBigInt
	}

	// Check that enough time has passed
	withdrawalPeriodTimeBigInt := getConstantsManagerVariable("withdrawalPeriodTime")

	if evm.Context.Time.Cmp(new(big.Int).Add(requestTime, withdrawalPeriodTimeBigInt)) < 0 {
		revertData, err := encodeRevertReason("withdraw", "not enough time passed")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Check that enough epochs have passed
	withdrawalPeriodEpochsBigInt := getConstantsManagerVariable("withdrawalPeriodEpochs")

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
	withdrawalRequestAmountState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(withdrawalRequestAmountSlot))
	gasUsed += SloadGasCost
	amount := new(big.Int).SetBytes(withdrawalRequestAmountState.Bytes())

	// Check if the validator is slashed
	validatorStatusSlot, getGasUsed := getValidatorStatusSlot(toValidatorID)
	gasUsed += getGasUsed
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorStatusSlot))
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())
	isCheater := (validatorStatusBigInt.Bit(7) == 1) // DOUBLESIGN_BIT

	// Calculate the penalty using getSlashingPenalty
	penalty, penaltyGasUsed, err := getSlashingPenalty(evm, amount, isCheater, toValidatorID)
	gasUsed += penaltyGasUsed
	if err != nil {
		log.Error("withdraw: getSlashingPenalty failed", "err", err)
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Delete the withdrawal request
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(withdrawalRequestAmountSlot), common.BigToHash(big.NewInt(0)))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(withdrawalRequestEpochSlot), common.BigToHash(big.NewInt(0)))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(withdrawalRequestTimeSlot), common.BigToHash(big.NewInt(0)))

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
	data, err := SfcLibAbi.Events["Withdrawn"].Inputs.NonIndexed().Pack(
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

// handleIsLockedUp implements the isLockedUp function from SFCLib.sol
func handleIsLockedUp(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
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

	// Get lockup info
	lockupEndTimeSlot, getGasUsed := getLockupEndTimeSlot(delegator, toValidatorID)
	gasUsed += getGasUsed
	lockupEndTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupEndTimeSlot))
	lockupEndTimeBigInt := new(big.Int).SetBytes(lockupEndTime.Bytes())
	gasUsed += SloadGasCost

	// Get locked stake
	lockedStakeSlot, getGasUsed := getLockedStakeSlot(delegator, toValidatorID)
	gasUsed += getGasUsed
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockedStakeSlot))
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())
	gasUsed += SloadGasCost

	// Check conditions exactly as in Solidity:
	// 1. endTime != 0
	// 2. lockedStake != 0
	// 3. _now() <= endTime
	isLocked := lockupEndTimeBigInt.Cmp(big.NewInt(0)) != 0 &&
		lockedStakeBigInt.Cmp(big.NewInt(0)) != 0 &&
		lockupEndTimeBigInt.Cmp(evm.Context.Time) >= 0
	log.Info("isLockedUp result", "isLocked", isLocked, "lockupEndTimeBigInt", lockupEndTimeBigInt, "evm.Context.Time", evm.Context.Time,
		"lockupEndTimeBigInt.Cmp(big.NewInt(0)) != 0", lockupEndTimeBigInt.Cmp(big.NewInt(0)) != 0,
		"lockedStakeBigInt.Cmp(big.NewInt(0)) != 0", lockedStakeBigInt.Cmp(big.NewInt(0)) != 0,
		"lockupEndTimeBigInt.Cmp(evm.Context.Time) >= 0", lockupEndTimeBigInt.Cmp(evm.Context.Time) >= 0)

	// Pack result
	result, err := SfcAbi.Methods["isLockedUp"].Outputs.Pack(isLocked)
	if err != nil {
		log.Error("isLockedUp: pack isLockedUp failed", "err", err)
		return nil, 0, err
	}

	log.Info("isLockedUp", "isLocked", isLocked, "result", result)

	return result, gasUsed, nil
}
