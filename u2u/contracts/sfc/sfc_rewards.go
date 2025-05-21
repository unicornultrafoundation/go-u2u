package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// handleClaimRewards claims the rewards for a delegator
func handleClaimRewards(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
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
	rewardsStash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(rewardsStashSlot))
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
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(rewardsStashSlot), common.BigToHash(big.NewInt(0)))

	// Mint the native token
	result, mintGasUsed, err := handle_mintNativeToken(evm, args)
	gasUsed += mintGasUsed
	if err != nil {
		return result, gasUsed, err
	}

	// Transfer the rewards to the delegator
	evm.SfcStateDB.AddBalance(caller, rewardsStashBigInt)
	evm.SfcStateDB.SubBalance(ContractAddress, rewardsStashBigInt)

	// Get the lockup duration for the delegation
	lockedStakeSlot, getGasUsed := getLockedStakeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockedStakeSlot))
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())

	// Get the lockup duration
	lockupDurationSlot, getGasUsed := getLockupDurationSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockupDuration := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupDurationSlot))
	lockupDurationBigInt := new(big.Int).SetBytes(lockupDuration.Bytes())

	// Scale the rewards based on the lockup duration
	var scaledRewards Rewards
	if lockedStakeBigInt.Cmp(big.NewInt(0)) > 0 {
		// If there's locked stake, use the lockup duration to scale the rewards
		var scaleGasUsed uint64
		var err error
		scaledRewards, scaleGasUsed, err = _scaleLockupReward(evm, rewardsStashBigInt, lockupDurationBigInt)
		gasUsed += scaleGasUsed
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
	} else {
		// If there's no locked stake, all rewards are unlocked
		scaledRewards = Rewards{
			LockupExtraReward: big.NewInt(0),
			LockupBaseReward:  big.NewInt(0),
			UnlockedReward:    rewardsStashBigInt,
		}
	}

	// Emit ClaimedRewards event
	topics := []common.Hash{
		SfcAbi.Events["ClaimedRewards"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
	}
	data, err := SfcAbi.Events["ClaimedRewards"].Inputs.NonIndexed().Pack(
		scaledRewards.LockupExtraReward, // lockupExtraReward
		scaledRewards.LockupBaseReward,  // lockupBaseReward
		scaledRewards.UnlockedReward,    // unlockedReward
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
	rewardsStash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(rewardsStashSlot))
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
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(rewardsStashSlot), common.BigToHash(big.NewInt(0)))

	// Mint the native token
	mintGasUsed, err := _mintNativeToken(evm, ContractAddress, rewardsStashBigInt)
	gasUsed += mintGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Delegate the rewards
	// Get the delegation stake
	stakeSlot, getGasUsed := getStakeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stakeSlot))
	stakeBigInt := new(big.Int).SetBytes(stake.Bytes())

	// Update the stake
	newStake := new(big.Int).Add(stakeBigInt, rewardsStashBigInt)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stakeSlot), common.BigToHash(newStake))

	// Update the validator's received stake
	validatorReceivedStakeSlot, getGasUsed := getValidatorReceivedStakeSlot(toValidatorID)
	gasUsed += getGasUsed
	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot))
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())
	newReceivedStake := new(big.Int).Add(receivedStakeBigInt, rewardsStashBigInt)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot), common.BigToHash(newReceivedStake))

	// Update the total stake
	totalStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)))
	totalStakeBigInt := new(big.Int).SetBytes(totalStakeState.Bytes())
	newTotalStake := new(big.Int).Add(totalStakeBigInt, rewardsStashBigInt)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)), common.BigToHash(newTotalStake))

	// Update the total active stake if the validator is active
	validatorStatusSlot, getGasUsed := getValidatorStatusSlot(toValidatorID)
	gasUsed += getGasUsed
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorStatusSlot))
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())
	if validatorStatusBigInt.Cmp(big.NewInt(0)) == 0 { // OK_STATUS
		totalActiveStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
		totalActiveStakeBigInt := new(big.Int).SetBytes(totalActiveStakeState.Bytes())
		newTotalActiveStake := new(big.Int).Add(totalActiveStakeBigInt, rewardsStashBigInt)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)), common.BigToHash(newTotalActiveStake))
	}

	// Get the lockup duration for the delegation
	lockedStakeSlot, getGasUsed := getLockedStakeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockedStakeSlot))
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())

	// Get the lockup duration
	lockupDurationSlot, getGasUsed := getLockupDurationSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockupDuration := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupDurationSlot))
	lockupDurationBigInt := new(big.Int).SetBytes(lockupDuration.Bytes())

	// Scale the rewards based on the lockup duration
	var scaledRewards Rewards
	if lockedStakeBigInt.Cmp(big.NewInt(0)) > 0 {
		// If there's locked stake, use the lockup duration to scale the rewards
		var scaleGasUsed uint64
		var err error
		scaledRewards, scaleGasUsed, err = _scaleLockupReward(evm, rewardsStashBigInt, lockupDurationBigInt)
		gasUsed += scaleGasUsed
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
	} else {
		// If there's no locked stake, all rewards are unlocked
		scaledRewards = Rewards{
			LockupExtraReward: big.NewInt(0),
			LockupBaseReward:  big.NewInt(0),
			UnlockedReward:    rewardsStashBigInt,
		}
	}

	// Calculate the lockup reward (sum of lockupExtraReward and lockupBaseReward)
	lockupReward := new(big.Int).Add(scaledRewards.LockupExtraReward, scaledRewards.LockupBaseReward)

	// Update the locked stake with the lockup reward
	newLockedStake := new(big.Int).Add(lockedStakeBigInt, lockupReward)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockedStakeSlot), common.BigToHash(newLockedStake))

	// Emit RestakedRewards event
	topics := []common.Hash{
		SfcAbi.Events["RestakedRewards"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
	}
	data, err := SfcAbi.Events["RestakedRewards"].Inputs.NonIndexed().Pack(
		scaledRewards.LockupExtraReward, // lockupExtraReward
		scaledRewards.LockupBaseReward,  // lockupBaseReward
		scaledRewards.UnlockedReward,    // unlockedReward
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
