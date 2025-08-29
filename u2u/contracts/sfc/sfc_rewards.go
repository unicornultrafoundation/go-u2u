package sfc

import (
	"fmt"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
)

// handleInternalClaimRewards implements the _claimRewards function from Solidity
func handleInternalClaimRewards(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int) (Rewards, uint64, error) {
	var gasUsed uint64 = 0

	// Check that the delegator is allowed to withdraw
	allowed, err := handleCheckAllowedToWithdraw(evm, delegator, toValidatorID)
	if err != nil {
		log.Error("handleInternalClaimRewards: handleCheckAllowedToWithdraw failed", "err", err)
		return Rewards{}, 0, err
	}
	if !allowed {
		log.Error("handleInternalClaimRewards: not allowed to withdraw")
		return Rewards{}, 0, fmt.Errorf("outstanding sU2U balance")
	}

	// Stash the rewards
	stashRewardsArgs := []interface{}{delegator, toValidatorID}
	_, stashGasUsed, err := handleInternalStashRewards(evm, stashRewardsArgs)
	gasUsed += stashGasUsed
	if err != nil {
		log.Error("handleInternalClaimRewards: handle_stashRewards failed", "err", err)
		return Rewards{}, gasUsed, vm.ErrExecutionReverted
	}

	// Get the base slot for the rewards stash
	rewardsStashSlot, slotGasUsed := getRewardsStashSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed

	// Read all three slots of the Rewards struct
	packedRewards := make([][]byte, 3)
	for i := 0; i < 3; i++ {
		slot := new(big.Int).Add(rewardsStashSlot, big.NewInt(int64(i)))
		value := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slot))
		packedRewards[i] = value.Bytes()
		gasUsed += SloadGasCost
	}

	// Unpack the rewards
	rewards := unpackRewards(packedRewards)

	// Calculate total reward
	totalReward := new(big.Int).Add(rewards.UnlockedReward, new(big.Int).Add(rewards.LockupBaseReward, rewards.LockupExtraReward))

	// Check that the rewards are not zero
	if totalReward.Cmp(big.NewInt(0)) == 0 {
		log.Error("handleInternalClaimRewards: zero rewards")
		return Rewards{}, gasUsed, fmt.Errorf("zero rewards")
	}

	// Clear the rewards stash
	zeroRewards := Rewards{
		LockupExtraReward: big.NewInt(0),
		LockupBaseReward:  big.NewInt(0),
		UnlockedReward:    big.NewInt(0),
	}
	packedZeroRewards := packRewards(zeroRewards)

	// Write zero to all three slots
	for i := 0; i < 3; i++ {
		slot := new(big.Int).Add(rewardsStashSlot, big.NewInt(int64(i)))
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(slot), common.BytesToHash(packedZeroRewards[i]))
		gasUsed += SstoreGasCost
	}

	// Mint the native token to the contract itself
	// It's important that we mint after erasing (protection against Re-Entrancy)
	mintArgs := []interface{}{
		ContractAddress, // Mint to the contract itself
		totalReward,     // Mint the total reward amount
	}
	_, mintGasUsed, err := handle_mintNativeToken(evm, mintArgs)
	gasUsed += mintGasUsed
	if err != nil {
		log.Error("handleInternalClaimRewards: handle_mintNativeToken failed", "err", err)
		return Rewards{}, gasUsed, err
	}

	// Return the actual rewards struct with the original breakdown
	return rewards, gasUsed, nil
}

// handleClaimRewards implements the claimRewards function from Solidity
func handleClaimRewards(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	toValidatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call handleInternalClaimRewards to get the rewards
	rewards, claimGasUsed, err := handleInternalClaimRewards(evm, caller, toValidatorID)
	gasUsed += claimGasUsed
	if err != nil {
		revertData, encodeErr := encodeRevertReason("claimRewards", "zero rewards")
		if encodeErr != nil {
			return nil, gasUsed, vm.ErrExecutionReverted
		}
		return revertData, gasUsed, vm.ErrExecutionReverted
	}

	// Calculate total reward
	totalReward := new(big.Int).Add(rewards.UnlockedReward, new(big.Int).Add(rewards.LockupBaseReward, rewards.LockupExtraReward))

	// Transfer the rewards to the caller
	if SfcPrecompiles[caller] {
		evm.SfcStateDB.AddBalance(caller, totalReward)
	}
	evm.SfcStateDB.SubBalance(ContractAddress, totalReward)

	// Emit ClaimedRewards event
	topics := []common.Hash{
		SfcLibAbi.Events["ClaimedRewards"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)),
		common.BigToHash(toValidatorID),
	}
	// Pack the three reward values using ABI encoding
	data, err := SfcLibAbi.Events["ClaimedRewards"].Inputs.NonIndexed().Pack(
		rewards.LockupExtraReward,
		rewards.LockupBaseReward,
		rewards.UnlockedReward,
	)
	if err != nil {
		return nil, gasUsed, err
	}
	evm.SfcStateDB.AddLog(&types.Log{
		BlockNumber: evm.Context.BlockNumber.Uint64(),
		Address:     ContractAddress,
		Topics:      topics,
		Data:        data,
	})

	return nil, gasUsed, nil
}

// handleRestakeRewards implements the restakeRewards function from Solidity
func handleRestakeRewards(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	toValidatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Call internal claim rewards (equivalent to _claimRewards in Solidity)
	rewards, claimGasUsed, err := handleInternalClaimRewards(evm, caller, toValidatorID)
	gasUsed += claimGasUsed
	if err != nil {
		resultData, _ := encodeRevertReason("restakeRewards", err.Error())
		return resultData, gasUsed, err
	}

	// Calculate lockup reward and total reward
	lockupReward := new(big.Int).Add(rewards.LockupExtraReward, rewards.LockupBaseReward)
	totalReward := new(big.Int).Add(lockupReward, rewards.UnlockedReward)

	// Delegate the rewards
	result, delegateGasUsed, err := handleInternalDelegate(evm, caller, toValidatorID, totalReward)
	if err != nil {
		log.Error("handleRestakeRewards: handleInternalDelegate failed", "err", err, "result", common.Bytes2Hex(result))
		return result, gasUsed + delegateGasUsed, err
	}
	gasUsed += delegateGasUsed

	// Update locked stake
	lockedStakeSlot, getGasUsed := getLockedStakeSlot(caller, toValidatorID)
	gasUsed += getGasUsed
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockedStakeSlot))
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())
	gasUsed += SloadGasCost

	newLockedStake := new(big.Int).Add(lockedStakeBigInt, lockupReward)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockedStakeSlot), common.BigToHash(newLockedStake))
	gasUsed += SstoreGasCost

	// Emit RestakeRewards event
	topics := []common.Hash{
		SfcLibAbi.Events["RestakedRewards"].ID,
		common.BytesToHash(common.LeftPadBytes(caller.Bytes(), 32)), // indexed parameter (delegator)
		common.BigToHash(toValidatorID),                             // indexed parameter (toValidatorID)
	}

	data := make([]byte, 0, 96)
	data = append(data, common.BigToHash(rewards.LockupExtraReward).Bytes()...)
	data = append(data, common.BigToHash(rewards.LockupBaseReward).Bytes()...)
	data = append(data, common.BigToHash(rewards.UnlockedReward).Bytes()...)

	evm.SfcStateDB.AddLog(&types.Log{
		BlockNumber: evm.Context.BlockNumber.Uint64(),
		Address:     ContractAddress,
		Topics:      topics,
		Data:        data,
	})

	return nil, gasUsed, nil
}
