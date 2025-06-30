package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// Handler functions for SFC contract variables
// This file contains handlers for variable getters (as opposed to function methods)

func handleOwner(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	result, err := SfcAbi.Methods["owner"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
	return result, 0, err
}

func handleCurrentSealedEpoch(evm *vm.EVM) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Get the current sealed epoch slot using a cached constant
	currentSealedEpochSlotHash := common.BigToHash(big.NewInt(currentSealedEpochSlot))

	// Get the current sealed epoch
	val := evm.SfcStateDB.GetState(ContractAddress, currentSealedEpochSlotHash)
	gasUsed += SloadGasCost

	// Use the big.Int pool
	currentSealedEpochBigInt := GetBigInt().SetBytes(val.Bytes())

	result, err := SfcAbi.Methods["currentSealedEpoch"].Outputs.Pack(currentSealedEpochBigInt)

	// Return the big.Int to the pool
	PutBigInt(currentSealedEpochBigInt)

	return result, gasUsed, err
}

func handleLastValidatorID(evm *vm.EVM) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Get the last validator ID slot using a cached constant
	lastValidatorIDSlotHash := common.BigToHash(big.NewInt(lastValidatorIDSlot))

	// Get the last validator ID
	val := evm.SfcStateDB.GetState(ContractAddress, lastValidatorIDSlotHash)
	gasUsed += SloadGasCost

	// Use the big.Int pool
	lastValidatorIDBigInt := GetBigInt().SetBytes(val.Bytes())

	// Pack the result using the cached ABI packing function
	result, err := CachedAbiPack(SfcAbiType, "lastValidatorID", lastValidatorIDBigInt)

	// Return the big.Int to the pool
	PutBigInt(lastValidatorIDBigInt)

	return result, gasUsed, err
}

func handleTotalStake(evm *vm.EVM) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Get the total stake slot using a cached constant
	totalStakeSlotHash := common.BigToHash(big.NewInt(totalStakeSlot))

	// Get the total stake
	val := evm.SfcStateDB.GetState(ContractAddress, totalStakeSlotHash)
	gasUsed += SloadGasCost

	// Use the big.Int pool
	totalStakeBigInt := GetBigInt().SetBytes(val.Bytes())

	result, err := SfcAbi.Methods["totalStake"].Outputs.Pack(totalStakeBigInt)

	// Return the big.Int to the pool
	PutBigInt(totalStakeBigInt)

	return result, gasUsed, err
}

func handleTotalActiveStake(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
	result, err := SfcAbi.Methods["totalActiveStake"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleTotalSlashedStake(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalSlashedStakeSlot)))
	result, err := SfcAbi.Methods["totalSlashedStake"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleTotalSupply(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)))
	result, err := SfcAbi.Methods["totalSupply"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleStakeTokenizerAddress(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeTokenizerAddressSlot)))
	result, err := SfcAbi.Methods["stakeTokenizerAddress"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
	return result, 0, err
}

func handleMinGasPrice(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)))
	result, err := SfcAbi.Methods["minGasPrice"].Outputs.Pack(val.Big())
	return result, 0, err
}

func handleTreasuryAddress(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(treasuryAddressSlot)))
	result, err := SfcAbi.Methods["treasuryAddress"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
	return result, 0, err
}

func handleVoteBookAddress(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(voteBookAddressSlot)))
	result, err := SfcAbi.Methods["voteBookAddress"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
	return result, 0, err
}

func handleGetValidator(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID := args[0].(*big.Int)

	// Calculate the validator slot using our cached function
	slotBigInt, slotGasUsed := getValidatorStatusSlot(validatorID)
	gasUsed += slotGasUsed

	// Get the validator status
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotBigInt))
	gasUsed += SloadGasCost

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["getValidator"].Outputs.Pack(val.Big())
	return result, gasUsed, err
}

func handleGetValidatorID(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr := args[0].(common.Address)

	// Calculate the validator ID slot using our cached function
	slotBigInt, slotGasUsed := getValidatorIDSlot(addr)
	gasUsed += slotGasUsed

	// Get the validator ID
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotBigInt))
	gasUsed += SloadGasCost

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["getValidatorID"].Outputs.Pack(val.Big())
	return result, gasUsed, err
}

func handleGetValidatorPubkey(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID := args[0].(*big.Int)

	// Calculate the validator pubkey slot using our cached function
	slotBigInt, slotGasUsed := getValidatorPubkeySlot(validatorID)
	gasUsed += slotGasUsed

	// Get the validator pubkey (dynamic bytes)
	pubkeyBytes, readBytesGasUsed, err := readDynamicBytes(evm, slotBigInt)
	if err != nil {
		return nil, gasUsed, err
	}
	gasUsed += readBytesGasUsed

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["getValidatorPubkey"].Outputs.Pack(pubkeyBytes)
	return result, gasUsed, err
}

func handleStashedRewardsUntilEpoch(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr := args[0].(common.Address)
	validatorID := args[1].(*big.Int)

	// Calculate the stashed rewards until epoch slot using our cached function
	slotBigInt, slotGasUsed := getStashedRewardsUntilEpochSlot(addr, validatorID)
	gasUsed += slotGasUsed

	// Get the stashed rewards until epoch
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotBigInt))
	gasUsed += SloadGasCost

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["stashedRewardsUntilEpoch"].Outputs.Pack(val.Big())
	return result, gasUsed, err
}

func handleGetWithdrawalRequest(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	if len(args) != 3 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr := args[0].(common.Address)
	validatorID := args[1].(*big.Int)
	wrID := args[2].(*big.Int)

	// Calculate the withdrawal request slot using our cached function
	slotBigInt, slotGasUsed := getWithdrawalRequestSlot(addr, validatorID, wrID)
	gasUsed += slotGasUsed

	// Get the withdrawal request fields
	epoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotBigInt))
	gasUsed += SloadGasCost

	// Use the big.Int pool for offset calculations
	offset1 := GetBigInt().SetInt64(1)
	slot1 := GetBigInt().Add(slotBigInt, offset1)
	time := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slot1))
	gasUsed += SloadGasCost

	offset2 := GetBigInt().SetInt64(2)
	slot2 := GetBigInt().Add(slotBigInt, offset2)
	amount := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slot2))
	gasUsed += SloadGasCost

	// Return temporary big.Ints to the pool
	PutBigInt(offset1)
	PutBigInt(offset2)
	PutBigInt(slot1)
	PutBigInt(slot2)

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["getWithdrawalRequest"].Outputs.Pack(
		epoch.Big(),
		time.Big(),
		amount.Big(),
	)
	return result, gasUsed, err
}

func handleGetStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr := args[0].(common.Address)
	validatorID := args[1].(*big.Int)

	// Calculate the stake slot using our cached function
	slotBigInt, slotGasUsed := getStakeSlot(addr, validatorID)
	gasUsed += slotGasUsed

	// Get the stake
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotBigInt))
	gasUsed += SloadGasCost

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["getStake"].Outputs.Pack(val.Big())
	return result, gasUsed, err
}

func handleGetLockupInfo(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr := args[0].(common.Address)
	validatorID := args[1].(*big.Int)

	// Calculate the lockup info slot using our cached function
	slotBigInt, slotGasUsed := getLockedStakeSlot(addr, validatorID)
	gasUsed += slotGasUsed

	// Get the lockup info fields
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotBigInt))
	gasUsed += SloadGasCost

	// Use the big.Int pool for offset calculations
	offset1 := GetBigInt().SetInt64(1)
	slot1 := GetBigInt().Add(slotBigInt, offset1)
	fromEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slot1))
	gasUsed += SloadGasCost

	offset2 := GetBigInt().SetInt64(2)
	slot2 := GetBigInt().Add(slotBigInt, offset2)
	endTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slot2))
	gasUsed += SloadGasCost

	offset3 := GetBigInt().SetInt64(3)
	slot3 := GetBigInt().Add(slotBigInt, offset3)
	duration := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slot3))
	gasUsed += SloadGasCost

	// Return temporary big.Ints to the pool
	PutBigInt(offset1)
	PutBigInt(offset2)
	PutBigInt(offset3)
	PutBigInt(slot1)
	PutBigInt(slot2)
	PutBigInt(slot3)

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["getLockupInfo"].Outputs.Pack(
		lockedStake.Big(),
		fromEpoch.Big(),
		endTime.Big(),
		duration.Big(),
	)
	return result, gasUsed, err
}

func handleGetStashedLockupRewards(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	if len(args) != 2 {
		return nil, 0, vm.ErrExecutionReverted
	}
	addr := args[0].(common.Address)
	validatorID := args[1].(*big.Int)

	// Calculate the base slot for stashed lockup rewards
	stashedLockupRewardsSlot, slotGasUsed := getStashedLockupRewardsSlot(addr, validatorID)
	gasUsed += slotGasUsed

	// Read all three slots of the stashed lockup rewards
	packedRewards := make([][]byte, 3)
	for i := 0; i < 3; i++ {
		slot := new(big.Int).Add(stashedLockupRewardsSlot, big.NewInt(int64(i)))
		value := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slot))
		packedRewards[i] = value.Bytes()
		gasUsed += SloadGasCost
	}

	// Unpack the rewards
	rewards := unpackRewards(packedRewards)

	// Pack the result using ABI
	result, err := SfcAbi.Methods["getStashedLockupRewards"].Outputs.Pack(
		rewards.LockupBaseReward,
		rewards.LockupExtraReward,
		rewards.UnlockedReward,
	)
	return result, gasUsed, err
}

func handleSlashingRefundRatio(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID := args[0].(*big.Int)

	// Create the hash input using cached padded values
	hashInput := CreateHashInput(validatorID, slashingRefundRatioSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a slot
	slot := common.BytesToHash(hash)

	// Get the slashing refund ratio
	val := evm.SfcStateDB.GetState(ContractAddress, slot)
	gasUsed += SloadGasCost

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["slashingRefundRatio"].Outputs.Pack(val.Big())
	return result, gasUsed, err
}

func handleGetEpochSnapshot(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	epoch := args[0].(*big.Int)

	// Create the hash input using cached padded values
	hashInput := CreateHashInput(epoch, epochSnapshotSlot)

	// Calculate the hash - add gas cost for hashing
	hash := CachedKeccak256(hashInput)
	gasUsed += HashGasCost

	// Convert the hash to a slot
	slotBigInt := GetBigInt().SetBytes(hash)

	// Use the big.Int pool for offset calculations
	offsetBigInt := GetBigInt()
	slotWithOffset := GetBigInt()

	// Get all the fixed-size fields in the order they appear in the struct
	// endTime
	offsetBigInt.SetInt64(endTimeOffset)
	slotWithOffset.Add(slotBigInt, offsetBigInt)
	endTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotWithOffset))
	gasUsed += SloadGasCost

	// epochFee
	offsetBigInt.SetInt64(epochFeeOffset)
	slotWithOffset.Set(slotBigInt).Add(slotWithOffset, offsetBigInt)
	epochFee := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotWithOffset))
	gasUsed += SloadGasCost

	// totalBaseRewardWeight
	offsetBigInt.SetInt64(totalBaseRewardOffset)
	slotWithOffset.Set(slotBigInt).Add(slotWithOffset, offsetBigInt)
	totalBaseRewardWeight := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotWithOffset))
	gasUsed += SloadGasCost

	// totalTxRewardWeight
	offsetBigInt.SetInt64(totalTxRewardOffset)
	slotWithOffset.Set(slotBigInt).Add(slotWithOffset, offsetBigInt)
	totalTxRewardWeight := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotWithOffset))
	gasUsed += SloadGasCost

	// baseRewardPerSecond
	offsetBigInt.SetInt64(baseRewardPerSecondOffset)
	slotWithOffset.Set(slotBigInt).Add(slotWithOffset, offsetBigInt)
	baseRewardPerSecond := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotWithOffset))
	gasUsed += SloadGasCost

	// totalStake
	offsetBigInt.SetInt64(totalStakeOffset)
	slotWithOffset.Set(slotBigInt).Add(slotWithOffset, offsetBigInt)
	totalStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotWithOffset))
	gasUsed += SloadGasCost

	// totalSupply
	offsetBigInt.SetInt64(totalSupplyOffset)
	slotWithOffset.Set(slotBigInt).Add(slotWithOffset, offsetBigInt)
	totalSupply := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(slotWithOffset))
	gasUsed += SloadGasCost

	// Return temporary big.Ints to the pool
	PutBigInt(slotBigInt)
	PutBigInt(offsetBigInt)
	PutBigInt(slotWithOffset)

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["getEpochSnapshot"].Outputs.Pack(
		endTime.Big(),
		epochFee.Big(),
		totalBaseRewardWeight.Big(),
		totalTxRewardWeight.Big(),
		baseRewardPerSecond.Big(),
		totalStake.Big(),
		totalSupply.Big(),
	)
	return result, gasUsed, err
}
