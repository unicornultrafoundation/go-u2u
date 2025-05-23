package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
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
	// Create arguments for handle_stashRewards
	stashRewardsArgs := []interface{}{caller, toValidatorID}
	// Call handle_stashRewards
	_, stashGasUsed, err := handle_stashRewards(evm, stashRewardsArgs)
	gasUsed += stashGasUsed
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Get the rewards from the stash
	// In the storage, rewards are stored as a struct with three fields
	rewardsStashSlot, getGasUsed := getRewardsStashSlot(caller, toValidatorID)
	gasUsed += getGasUsed

	// Get lockupExtraReward
	lockupExtraRewardSlot := rewardsStashSlot
	lockupExtraReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupExtraRewardSlot))
	lockupExtraRewardBigInt := new(big.Int).SetBytes(lockupExtraReward.Bytes())
	gasUsed += SloadGasCost

	// Get lockupBaseReward
	lockupBaseRewardSlot := new(big.Int).Add(rewardsStashSlot, big.NewInt(1))
	lockupBaseReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupBaseRewardSlot))
	lockupBaseRewardBigInt := new(big.Int).SetBytes(lockupBaseReward.Bytes())
	gasUsed += SloadGasCost

	// Get unlockedReward
	unlockedRewardSlot := new(big.Int).Add(rewardsStashSlot, big.NewInt(2))
	unlockedReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(unlockedRewardSlot))
	unlockedRewardBigInt := new(big.Int).SetBytes(unlockedReward.Bytes())
	gasUsed += SloadGasCost

	// Create the rewards struct
	rewards := Rewards{
		LockupExtraReward: lockupExtraRewardBigInt,
		LockupBaseReward:  lockupBaseRewardBigInt,
		UnlockedReward:    unlockedRewardBigInt,
	}

	// Calculate total reward
	totalReward := new(big.Int).Add(rewards.UnlockedReward, new(big.Int).Add(rewards.LockupBaseReward, rewards.LockupExtraReward))

	// Check that the rewards are not zero
	if totalReward.Cmp(big.NewInt(0)) == 0 {
		revertData, err := encodeRevertReason("claimRewards", "zero rewards")
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}

	// Clear the rewards stash (delete _rewardsStash[delegator][toValidatorID])
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupExtraRewardSlot), common.Hash{})
	gasUsed += SstoreGasCost
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupBaseRewardSlot), common.Hash{})
	gasUsed += SstoreGasCost
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(unlockedRewardSlot), common.Hash{})
	gasUsed += SstoreGasCost

	// Mint the native token to the contract itself
	// It's important that we mint after erasing (protection against Re-Entrancy)
	mintArgs := []interface{}{
		ContractAddress, // Mint to the contract itself
		totalReward,     // Mint the total reward amount
	}
	_, mintGasUsed, err := handle_mintNativeToken(evm, mintArgs)
	gasUsed += mintGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Transfer the rewards to the delegator
	// It's important that we transfer after erasing (protection against Re-Entrancy)
	evm.SfcStateDB.AddBalance(caller, totalReward)
	evm.SfcStateDB.SubBalance(ContractAddress, totalReward)

	// Emit ClaimedRewards event
	topics := []common.Hash{
		SfcAbi.Events["ClaimedRewards"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
	}

	// Pack the non-indexed parameters manually
	// The event definition is:
	// event ClaimedRewards(address indexed delegator, uint256 indexed toValidatorID, uint256 lockupExtraReward, uint256 lockupBaseReward, uint256 unlockedReward)

	// Create a buffer to hold the packed data
	data := make([]byte, 0, 96) // 3 * 32 bytes for the three uint256 values

	// Pack each uint256 value (32 bytes each)
	lockupExtraRewardBytes := common.BigToHash(rewards.LockupExtraReward).Bytes()
	lockupBaseRewardBytes := common.BigToHash(rewards.LockupBaseReward).Bytes()
	unlockedRewardBytes := common.BigToHash(rewards.UnlockedReward).Bytes()

	// Append all bytes to the data buffer
	data = append(data, lockupExtraRewardBytes...)
	data = append(data, lockupBaseRewardBytes...)
	data = append(data, unlockedRewardBytes...)

	// No error can occur with this manual packing
	err = error(nil)

	if err != nil {
		log.Error("SFC: Error packing ClaimedRewards event data", "err", err)
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, gasUsed, nil
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
	// Create arguments for handle_stashRewards
	stashRewardsArgs := []interface{}{caller, toValidatorID}
	// Call handle_stashRewards
	_, _, err = handle_stashRewards(evm, stashRewardsArgs)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the rewards from the stash
	// In the storage, rewards are stored as a struct with three fields
	rewardsStashSlot, getGasUsed := getRewardsStashSlot(caller, toValidatorID)
	gasUsed += getGasUsed

	// Get lockupExtraReward
	lockupExtraRewardSlot := rewardsStashSlot
	lockupExtraReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupExtraRewardSlot))
	lockupExtraRewardBigInt := new(big.Int).SetBytes(lockupExtraReward.Bytes())
	gasUsed += SloadGasCost

	// Get lockupBaseReward
	lockupBaseRewardSlot := new(big.Int).Add(rewardsStashSlot, big.NewInt(1))
	lockupBaseReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupBaseRewardSlot))
	lockupBaseRewardBigInt := new(big.Int).SetBytes(lockupBaseReward.Bytes())
	gasUsed += SloadGasCost

	// Get unlockedReward
	unlockedRewardSlot := new(big.Int).Add(rewardsStashSlot, big.NewInt(2))
	unlockedReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(unlockedRewardSlot))
	unlockedRewardBigInt := new(big.Int).SetBytes(unlockedReward.Bytes())
	gasUsed += SloadGasCost

	// Create the rewards struct
	rewards := Rewards{
		LockupExtraReward: lockupExtraRewardBigInt,
		LockupBaseReward:  lockupBaseRewardBigInt,
		UnlockedReward:    unlockedRewardBigInt,
	}

	// Calculate total reward
	totalReward := new(big.Int).Add(rewards.UnlockedReward, new(big.Int).Add(rewards.LockupBaseReward, rewards.LockupExtraReward))

	// Check that the rewards are not zero
	if totalReward.Cmp(big.NewInt(0)) == 0 {
		revertData, err := encodeRevertReason("restakeRewards", "zero rewards")
		if err != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}

	// Clear the rewards stash (delete _rewardsStash[delegator][toValidatorID])
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupExtraRewardSlot), common.Hash{})
	gasUsed += SstoreGasCost
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupBaseRewardSlot), common.Hash{})
	gasUsed += SstoreGasCost
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(unlockedRewardSlot), common.Hash{})
	gasUsed += SstoreGasCost

	// Mint the native token to the contract itself
	// It's important that we mint after erasing (protection against Re-Entrancy)
	mintGasUsed, err := _mintNativeToken(evm, ContractAddress, totalReward)
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
	newStake := new(big.Int).Add(stakeBigInt, totalReward)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stakeSlot), common.BigToHash(newStake))

	// Update the validator's received stake
	validatorReceivedStakeSlot, getGasUsed := getValidatorReceivedStakeSlot(toValidatorID)
	gasUsed += getGasUsed
	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot))
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())
	newReceivedStake := new(big.Int).Add(receivedStakeBigInt, totalReward)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot), common.BigToHash(newReceivedStake))

	// Update the total stake
	totalStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)))
	totalStakeBigInt := new(big.Int).SetBytes(totalStakeState.Bytes())
	newTotalStake := new(big.Int).Add(totalStakeBigInt, totalReward)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)), common.BigToHash(newTotalStake))

	// Update the total active stake if the validator is active
	validatorStatusSlot, getGasUsed := getValidatorStatusSlot(toValidatorID)
	gasUsed += getGasUsed
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorStatusSlot))
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())
	if validatorStatusBigInt.Cmp(big.NewInt(0)) == 0 { // OK_STATUS
		totalActiveStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
		totalActiveStakeBigInt := new(big.Int).SetBytes(totalActiveStakeState.Bytes())
		newTotalActiveStake := new(big.Int).Add(totalActiveStakeBigInt, totalReward)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)), common.BigToHash(newTotalActiveStake))
	}

	// Get the lockup duration for the delegation
	lockedStakeSlot, getGasUsed := getLockedStakeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockedStakeSlot))
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())

	// Calculate the lockup reward (sum of lockupExtraReward and lockupBaseReward)
	lockupReward := new(big.Int).Add(rewards.LockupExtraReward, rewards.LockupBaseReward)

	// Update the locked stake with the lockup reward
	newLockedStake := new(big.Int).Add(lockedStakeBigInt, lockupReward)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockedStakeSlot), common.BigToHash(newLockedStake))

	// Emit RestakedRewards event
	topics := []common.Hash{
		SfcAbi.Events["RestakedRewards"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
	}

	// Pack the non-indexed parameters manually
	// The event definition is:
	// event RestakedRewards(address indexed delegator, uint256 indexed toValidatorID, uint256 lockupExtraReward, uint256 lockupBaseReward, uint256 unlockedReward)

	// Create a buffer to hold the packed data
	data := make([]byte, 0, 96) // 3 * 32 bytes for the three uint256 values

	// Pack each uint256 value (32 bytes each)
	lockupExtraRewardBytes := common.BigToHash(rewards.LockupExtraReward).Bytes()
	lockupBaseRewardBytes := common.BigToHash(rewards.LockupBaseReward).Bytes()
	unlockedRewardBytes := common.BigToHash(rewards.UnlockedReward).Bytes()

	// Append all bytes to the data buffer
	data = append(data, lockupExtraRewardBytes...)
	data = append(data, lockupBaseRewardBytes...)
	data = append(data, unlockedRewardBytes...)

	evm.SfcStateDB.AddLog(&types.Log{
		Address: ContractAddress,
		Topics:  topics,
		Data:    data,
	})

	return nil, 0, nil
}
