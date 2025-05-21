package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/params"
)

// Handler functions for SFC contract internal and private functions

// _calcRawValidatorEpochBaseReward is an internal function to calculate raw validator epoch base reward
func handle_calcRawValidatorEpochBaseReward(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _calcRawValidatorEpochBaseReward handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _calcRawValidatorEpochTxReward is an internal function to calculate raw validator epoch transaction reward
func handle_calcRawValidatorEpochTxReward(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _calcRawValidatorEpochTxReward handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _calcValidatorCommission is an internal function to calculate validator commission
func handle_calcValidatorCommission(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _calcValidatorCommission handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _scaleLockupReward is an internal function to scale lockup reward
func handle_scaleLockupReward(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Get the arguments
	if len(args) != 2 {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	fullReward, ok := args[0].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	lockupDuration, ok := args[1].(*big.Int)
	if !ok {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Call the _scaleLockupReward helper function
	reward, scaleGasUsed, err := _scaleLockupReward(evm, fullReward, lockupDuration)
	gasUsed += scaleGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["_scaleLockupReward"].Outputs.Pack(
		reward.LockupBaseReward,
		reward.LockupExtraReward,
		reward.UnlockedReward,
	)
	if err != nil {
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	return result, gasUsed, nil
}

// _setValidatorDeactivated is an internal function to set a validator as deactivated
func handle_setValidatorDeactivated(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _setValidatorDeactivated handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _syncValidator is an internal function to sync validator data
func handle_syncValidator(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _syncValidator handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _validatorExists is an internal function to check if a validator exists
func handle_validatorExists(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _validatorExists handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _now is an internal function to get the current time
func handle_now(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _now handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// getSlashingPenalty is an internal function to get the slashing penalty
func handleGetSlashingPenalty(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement getSlashingPenalty handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleIsNode checks if the caller is the node (address(0))
func handleIsNode(evm *vm.EVM, caller common.Address) (bool, error) {
	// Check if caller is address(0)
	emptyAddr := common.Address{}
	return caller.Cmp(emptyAddr) == 0, nil
}

// handleGetUnlockedStake returns the unlocked stake of a delegator
func handleGetUnlockedStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// Get the arguments
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

	// Get the delegation stake
	stakeSlot, slotGasUsed := getStakeSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stakeSlot))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	stakeBigInt := new(big.Int).SetBytes(stake.Bytes())

	// Get the delegation locked stake
	lockedStakeSlot, slotGasUsed := getLockedStakeSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockedStakeSlot))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())

	// Calculate the unlocked stake
	unlockedStake := new(big.Int).Sub(stakeBigInt, lockedStakeBigInt)
	if unlockedStake.Cmp(big.NewInt(0)) < 0 {
		unlockedStake = big.NewInt(0)
	}

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["getUnlockedStake"].Outputs.Pack(unlockedStake)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	return result, 0, nil
}

// handleCheckAllowedToWithdraw checks if a delegator is allowed to withdraw
func handleCheckAllowedToWithdraw(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int) (bool, error) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// Get the validator status
	validatorStatusSlot, slotGasUsed := getValidatorStatusSlot(toValidatorID)
	gasUsed += slotGasUsed
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorStatusSlot))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())

	// Check if the validator is deactivated
	isDeactivated := (validatorStatusBigInt.Bit(0) == 1) // WITHDRAWN_BIT

	// Get the validator auth
	validatorAuthSlot, slotGasUsed := getValidatorAuthSlot(toValidatorID)
	gasUsed += slotGasUsed
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorAuthSlot))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

	// Check if the delegator is the validator auth
	isAuth := (delegator.Cmp(validatorAuthAddr) == 0)

	// Get the stakeTokenizerAddress
	stakeTokenizerAddressState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeTokenizerAddressSlot)))
	stakeTokenizerAddressBytes := stakeTokenizerAddressState.Bytes()

	// Check if stakeTokenizerAddress is zero
	isZeroAddress := true
	for _, b := range stakeTokenizerAddressBytes {
		if b != 0 {
			isZeroAddress = false
			break
		}
	}

	if isZeroAddress {
		// If stakeTokenizerAddress is zero, a delegator is allowed to withdraw if the validator is deactivated or if the delegator is the validator auth
		return isDeactivated || isAuth, nil
	}

	// Call the allowedToWithdrawStake function on the StakeTokenizer contract
	stakeTokenizerAddr := common.BytesToAddress(stakeTokenizerAddressBytes)

	// Pack the function call data for allowedToWithdrawStake(address,uint256)
	methodID := []byte{0x4d, 0x31, 0x52, 0x9d} // keccak256("allowedToWithdrawStake(address,uint256)")[:4]

	// Use our helper function to create the hash input with parameters
	delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)
	toValidatorIDBytes := common.LeftPadBytes(toValidatorID.Bytes(), 32)

	// Use the byte slice pool for the result
	data := GetByteSlice()
	if cap(data) < len(methodID)+len(delegatorBytes)+len(toValidatorIDBytes) {
		// If the slice from the pool is too small, allocate a new one
		data = make([]byte, 0, len(methodID)+len(delegatorBytes)+len(toValidatorIDBytes))
	}

	// Combine the bytes
	data = append(data, methodID...)
	data = append(data, delegatorBytes...)
	data = append(data, toValidatorIDBytes...)

	// Make the call to the StakeTokenizer contract
	result, _, err := evm.CallSFC(vm.AccountRef(ContractAddress), stakeTokenizerAddr, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return false, err
	}

	// The result is a bool, which is a uint8 in the ABI
	if len(result) < 32 {
		return false, vm.ErrExecutionReverted
	}

	// Check the result (last byte of the 32-byte value)
	allowed := result[31] != 0

	// A delegator is allowed to withdraw if the validator is deactivated, if the delegator is the validator auth, or if the StakeTokenizer allows it
	return isDeactivated || isAuth || allowed, nil
}

// handleCheckDelegatedStakeLimit checks if a validator's delegated stake is within the limit
func handleCheckDelegatedStakeLimit(evm *vm.EVM, validatorID *big.Int) (bool, error) {
	// Get the self-stake
	// Create arguments for handleGetSelfStake
	args := []interface{}{validatorID}
	// Call handleGetSelfStake
	result, _, err := handleGetSelfStake(evm, args)
	if err != nil {
		return false, err
	}

	// Unpack the result
	// Note: We don't cache unpacking operations as they're more complex and less frequent
	selfStakeValues, err := SfcAbi.Methods["getSelfStake"].Outputs.Unpack(result)
	if err != nil {
		return false, vm.ErrExecutionReverted
	}

	// The result should be a single *big.Int value
	if len(selfStakeValues) != 1 {
		return false, vm.ErrExecutionReverted
	}

	selfStake, ok := selfStakeValues[0].(*big.Int)
	if !ok {
		return false, vm.ErrExecutionReverted
	}

	// Get the validator's received stake
	validatorReceivedStakeSlot, _ := getValidatorReceivedStakeSlot(validatorID)
	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot))
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

	// Get the max delegated ratio
	maxDelegatedRatioBigInt := getConstantsManagerVariable("maxDelegatedRatio")

	// Calculate the delegated stake
	delegatedStake := new(big.Int).Sub(receivedStakeBigInt, selfStake)
	if delegatedStake.Cmp(big.NewInt(0)) < 0 {
		delegatedStake = big.NewInt(0)
	}

	// Calculate the maximum allowed delegated stake
	maxDelegatedStake := new(big.Int).Mul(selfStake, maxDelegatedRatioBigInt)
	maxDelegatedStake = new(big.Int).Div(maxDelegatedStake, getDecimalUnit()) // Divide by Decimal.unit()

	// Check if the delegated stake is within the limit
	return delegatedStake.Cmp(maxDelegatedStake) <= 0, nil
}

// handleRecountVotes implements the _recountVotes function logic
func handleRecountVotes(evm *vm.EVM, delegator common.Address, validatorAuth common.Address, strict bool) ([]byte, uint64, error) {
	// Get the voteBookAddress
	voteBookAddress := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(voteBookAddressSlot)))
	voteBookAddressBytes := voteBookAddress.Bytes()

	// Check if voteBookAddress is not zero
	isZeroAddress := true
	for _, b := range voteBookAddressBytes {
		if b != 0 {
			isZeroAddress = false
			break
		}
	}

	if !isZeroAddress {
		// Pack the function call data for recountVotes(address,address)
		methodID := []byte{0x71, 0x7a, 0x68, 0x5d} // keccak256("recountVotes(address,address)")[:4]

		// Use our helper function to create the hash input with parameters
		delegatorBytes := common.LeftPadBytes(delegator.Bytes(), 32)
		validatorAuthBytes := common.LeftPadBytes(validatorAuth.Bytes(), 32)

		// Use the byte slice pool for the result
		data := GetByteSlice()
		if cap(data) < len(methodID)+len(delegatorBytes)+len(validatorAuthBytes) {
			// If the slice from the pool is too small, allocate a new one
			data = make([]byte, 0, len(methodID)+len(delegatorBytes)+len(validatorAuthBytes))
		}

		// Combine the bytes
		data = append(data, methodID...)
		data = append(data, delegatorBytes...)
		data = append(data, validatorAuthBytes...)

		// Make the call to the voteBook contract with gas limit of 8000000
		voteBookAddr := common.BytesToAddress(voteBookAddressBytes)
		_, leftOverGas, err := evm.CallSFC(vm.AccountRef(ContractAddress), voteBookAddr, data, 8000000, big.NewInt(0))

		// Check if the call was successful
		if err != nil && strict {
			return nil, 8000000 - leftOverGas, err
		}
	}

	return nil, 0, nil
}

// callSFCLibDelegate calls the _delegate function in the SFCLib contract
func callSFCLibDelegate(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int, amount *big.Int) ([]byte, uint64, error) {
	// Get the SFCLib contract address
	sfcLibAddr := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(libAddressSlot)))
	sfcLibAddress := common.BytesToAddress(sfcLibAddr.Bytes())

	// Pack the function call data
	// The function signature is _delegate(address,uint256,uint256)
	methodID := []byte{0x9d, 0x11, 0xb4, 0x2d} // keccak256("_delegate(address,uint256,uint256)")[:4]
	data := methodID

	// Encode the parameters
	// address delegator
	data = append(data, common.LeftPadBytes(delegator.Bytes(), 32)...)
	// uint256 toValidatorID
	data = append(data, common.LeftPadBytes(toValidatorID.Bytes(), 32)...)
	// uint256 amount
	data = append(data, common.LeftPadBytes(amount.Bytes(), 32)...)

	// Make the call to the SFCLib contract
	result, leftOverGas, err := evm.CallSFC(vm.AccountRef(ContractAddress), sfcLibAddress, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, defaultGasLimit - leftOverGas, err
	}

	return result, defaultGasLimit - leftOverGas, nil
}

// handleGetSelfStake returns the self-stake of a validator
func handleGetSelfStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the validator auth
	validatorAuthSlot, slotGasUsed := getValidatorAuthSlot(validatorID)
	gasUsed += slotGasUsed
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorAuthSlot))
	gasUsed += SloadGasCost
	validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

	// Get the self-stake
	stakeSlot, slotGasUsed := getStakeSlot(validatorAuthAddr, validatorID)
	gasUsed += slotGasUsed
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stakeSlot))
	gasUsed += SloadGasCost

	// Use the big.Int pool
	stakeBigInt := GetBigInt().SetBytes(stake.Bytes())

	// Don't use cache for ABI packing with parameters
	result, err := SfcAbi.Methods["getSelfStake"].Outputs.Pack(stakeBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Return the big.Int to the pool
	PutBigInt(stakeBigInt)

	return result, gasUsed, nil
}

// handle_stashRewards stashes the rewards for a delegator
func handle_stashRewards(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0
	// Get the arguments
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

	// Calculate nonStashedReward using _newRewards logic
	nonStashedReward, newRewardsGasUsed, err := _newRewards(evm, delegator, toValidatorID)
	if err != nil {
		log.Error("handle_stashRewards _newRewards", "err", err)
		return nil, gasUsed, err
	}
	gasUsed += newRewardsGasUsed

	// Get the highest payable epoch for the validator
	highestPayableEpoch, epochGasUsed, err := _highestPayableEpoch(evm, toValidatorID)
	if err != nil {
		log.Error("handle_stashRewards _highestPayableEpoch", "err", err)
		return nil, gasUsed, err
	}
	gasUsed += epochGasUsed

	// Update stashedRewardsUntilEpoch
	stashedRewardsUntilEpochSlot, slotGasUsed := getStashedRewardsUntilEpochSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stashedRewardsUntilEpochSlot), common.BigToHash(highestPayableEpoch))
	gasUsed += SstoreGasCost

	// Get current rewards stash
	rewardsStashSlot, slotGasUsed := getRewardsStashSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed

	// Get lockupExtraReward from current stash
	lockupExtraRewardSlot := rewardsStashSlot
	lockupExtraReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupExtraRewardSlot))
	lockupExtraRewardBigInt := new(big.Int).SetBytes(lockupExtraReward.Bytes())
	gasUsed += SloadGasCost

	// Get lockupBaseReward from current stash
	lockupBaseRewardSlot := new(big.Int).Add(rewardsStashSlot, big.NewInt(1))
	lockupBaseReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupBaseRewardSlot))
	lockupBaseRewardBigInt := new(big.Int).SetBytes(lockupBaseReward.Bytes())
	gasUsed += SloadGasCost

	// Get unlockedReward from current stash
	unlockedRewardSlot := new(big.Int).Add(rewardsStashSlot, big.NewInt(2))
	unlockedReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(unlockedRewardSlot))
	unlockedRewardBigInt := new(big.Int).SetBytes(unlockedReward.Bytes())
	gasUsed += SloadGasCost

	// Create the current rewards stash struct
	currentRewardsStash := Rewards{
		LockupExtraReward: lockupExtraRewardBigInt,
		LockupBaseReward:  lockupBaseRewardBigInt,
		UnlockedReward:    unlockedRewardBigInt,
	}

	// Sum the rewards
	newRewardsStash := sumRewards(currentRewardsStash, nonStashedReward, Rewards{
		LockupExtraReward: big.NewInt(0),
		LockupBaseReward:  big.NewInt(0),
		UnlockedReward:    big.NewInt(0),
	})

	// Update the rewards stash
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupExtraRewardSlot), common.BigToHash(newRewardsStash.LockupExtraReward))
	gasUsed += SstoreGasCost
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupBaseRewardSlot), common.BigToHash(newRewardsStash.LockupBaseReward))
	gasUsed += SstoreGasCost
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(unlockedRewardSlot), common.BigToHash(newRewardsStash.UnlockedReward))
	gasUsed += SstoreGasCost

	// Get stashed lockup rewards
	stashedLockupRewardsSlot, slotGasUsed := getStashedLockupRewardsSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed

	// Get lockupExtraReward from stashed lockup rewards
	stashedLockupExtraRewardSlot := stashedLockupRewardsSlot
	stashedLockupExtraReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stashedLockupExtraRewardSlot))
	stashedLockupExtraRewardBigInt := new(big.Int).SetBytes(stashedLockupExtraReward.Bytes())
	gasUsed += SloadGasCost

	// Get lockupBaseReward from stashed lockup rewards
	stashedLockupBaseRewardSlot := new(big.Int).Add(stashedLockupRewardsSlot, big.NewInt(1))
	stashedLockupBaseReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stashedLockupBaseRewardSlot))
	stashedLockupBaseRewardBigInt := new(big.Int).SetBytes(stashedLockupBaseReward.Bytes())
	gasUsed += SloadGasCost

	// Get unlockedReward from stashed lockup rewards
	stashedUnlockedRewardSlot := new(big.Int).Add(stashedLockupRewardsSlot, big.NewInt(2))
	stashedUnlockedReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stashedUnlockedRewardSlot))
	stashedUnlockedRewardBigInt := new(big.Int).SetBytes(stashedUnlockedReward.Bytes())
	gasUsed += SloadGasCost

	// Create the stashed lockup rewards struct
	currentStashedLockupRewards := Rewards{
		LockupExtraReward: stashedLockupExtraRewardBigInt,
		LockupBaseReward:  stashedLockupBaseRewardBigInt,
		UnlockedReward:    stashedUnlockedRewardBigInt,
	}

	// Sum the stashed lockup rewards
	newStashedLockupRewards := sumRewards(currentStashedLockupRewards, nonStashedReward, Rewards{
		LockupExtraReward: big.NewInt(0),
		LockupBaseReward:  big.NewInt(0),
		UnlockedReward:    big.NewInt(0),
	})

	// Update the stashed lockup rewards
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stashedLockupExtraRewardSlot), common.BigToHash(newStashedLockupRewards.LockupExtraReward))
	gasUsed += SstoreGasCost
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stashedLockupBaseRewardSlot), common.BigToHash(newStashedLockupRewards.LockupBaseReward))
	gasUsed += SstoreGasCost
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stashedUnlockedRewardSlot), common.BigToHash(newStashedLockupRewards.UnlockedReward))
	gasUsed += SstoreGasCost

	// Check if the delegation is locked up
	isLocked, lockedGasUsed, err := isLockedUp(evm, delegator, toValidatorID)
	if err != nil {
		return nil, gasUsed, err
	}
	gasUsed += lockedGasUsed

	// If not locked up, delete lockup info and stashed lockup rewards
	if !isLocked {
		// Delete lockup info
		lockedStakeSlot, slotGasUsed := getLockedStakeSlot(delegator, toValidatorID)
		gasUsed += slotGasUsed
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockedStakeSlot), common.Hash{})
		gasUsed += SstoreGasCost

		lockupFromEpochSlot, slotGasUsed := getLockupFromEpochSlot(delegator, toValidatorID)
		gasUsed += slotGasUsed
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupFromEpochSlot), common.Hash{})
		gasUsed += SstoreGasCost

		lockupEndTimeSlot, slotGasUsed := getLockupEndTimeSlot(delegator, toValidatorID)
		gasUsed += slotGasUsed
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupEndTimeSlot), common.Hash{})
		gasUsed += SstoreGasCost

		lockupDurationSlot, slotGasUsed := getLockupDurationSlot(delegator, toValidatorID)
		gasUsed += slotGasUsed
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupDurationSlot), common.Hash{})
		gasUsed += SstoreGasCost

		// Delete stashed lockup rewards
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stashedLockupExtraRewardSlot), common.Hash{})
		gasUsed += SstoreGasCost
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stashedLockupBaseRewardSlot), common.Hash{})
		gasUsed += SstoreGasCost
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stashedUnlockedRewardSlot), common.Hash{})
		gasUsed += SstoreGasCost
	}

	// Check if there were any rewards to stash
	updated := nonStashedReward.LockupBaseReward.Cmp(big.NewInt(0)) != 0 ||
		nonStashedReward.LockupExtraReward.Cmp(big.NewInt(0)) != 0 ||
		nonStashedReward.UnlockedReward.Cmp(big.NewInt(0)) != 0
	var result int64
	if updated {
		result = 1
	} else {
		result = 0
	}
	return big.NewInt(result).Bytes(), gasUsed, nil
}

// checkDelegatedStakeLimit checks if the delegated stake is within the limit
// Returns true if the limit is exceeded, false otherwise
func checkDelegatedStakeLimit(evm *vm.EVM, validatorID *big.Int) (bool, uint64, error) {
	var gasUsed uint64 = 0

	// Get the validator's self-stake
	validatorAuthSlot, slotGasUsed := getValidatorAuthSlot(validatorID)
	gasUsed += slotGasUsed
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorAuthSlot))
	gasUsed += SloadGasCost
	validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

	// Get the self-stake (stake of the validator auth address)
	selfStakeSlot, slotGasUsed := getStakeSlot(validatorAuthAddr, validatorID)
	gasUsed += slotGasUsed
	selfStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(selfStakeSlot))
	gasUsed += SloadGasCost
	selfStakeBigInt := new(big.Int).SetBytes(selfStake.Bytes())

	// Get the validator's received stake
	validatorReceivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(validatorID)
	gasUsed += slotGasUsed
	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot))
	gasUsed += SloadGasCost
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

	// Get the max delegated ratio
	maxDelegatedRatioBigInt := getConstantsManagerVariable("maxDelegatedRatio")

	// Calculate the max delegated stake
	maxDelegatedStake := new(big.Int).Mul(selfStakeBigInt, maxDelegatedRatioBigInt)
	maxDelegatedStake = new(big.Int).Div(maxDelegatedStake, getDecimalUnit())

	// Check if the received stake exceeds the self-stake * maxDelegatedRatio
	// In Solidity: return getValidator[validatorID].receivedStake <= getSelfStake(validatorID).mul(c.maxDelegatedRatio()).div(Decimal.unit());
	limitExceeded := receivedStakeBigInt.Cmp(maxDelegatedStake) > 0

	return limitExceeded, gasUsed, nil
}

// handleSyncValidator synchronizes a validator's state
func handleSyncValidator(evm *vm.EVM, validatorID *big.Int, isZeroOrig ...bool) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Check if isZeroOrig parameter was provided
	// This parameter is currently not used in the implementation
	// but is kept for compatibility with the Solidity implementation
	_ = isZeroOrig // Avoid unused variable warning

	// Get the validator status
	validatorStatusSlot, slotGasUsed := getValidatorStatusSlot(validatorID)
	gasUsed += slotGasUsed
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorStatusSlot))
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())

	// Check if the validator is active
	isActive := (validatorStatusBigInt.Cmp(big.NewInt(0)) == 0) // OK_STATUS

	// Get the self-stake
	// Create arguments for handleGetSelfStake
	args := []interface{}{validatorID}
	// Call handleGetSelfStake
	result, _, err := handleGetSelfStake(evm, args)
	if err != nil {
		return nil, 0, err
	}

	// Unpack the result
	selfStakeValues, err := SfcAbi.Methods["getSelfStake"].Outputs.Unpack(result)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	// The result should be a single *big.Int value
	if len(selfStakeValues) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}

	selfStake, ok := selfStakeValues[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the minimum self-stake
	minSelfStakeBigInt := getConstantsManagerVariable("minSelfStake")

	// Check if the self-stake is at least the minimum self-stake
	hasSelfStake := selfStake.Cmp(big.NewInt(0)) > 0
	hasEnoughSelfStake := selfStake.Cmp(minSelfStakeBigInt) >= 0

	// Check if the delegated stake is within the limit
	withinDelegatedLimit, err := handleCheckDelegatedStakeLimit(evm, validatorID)
	if err != nil {
		return nil, 0, err
	}

	// Update the validator status if necessary
	if isActive && (!hasSelfStake || !hasEnoughSelfStake || !withinDelegatedLimit) {
		// Set the validator as deactivated
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorStatusSlot), common.BigToHash(big.NewInt(1))) // WITHDRAWN_BIT

		// Set the validator deactivated epoch
		validatorDeactivatedEpochSlot, slotGasUsed := getValidatorDeactivatedEpochSlot(validatorID)
		gasUsed += slotGasUsed
		currentEpochBigInt, _, err := getCurrentEpoch(evm)
		if err != nil {
			return nil, 0, err
		}
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorDeactivatedEpochSlot), common.BigToHash(currentEpochBigInt))

		// Set the validator deactivated time
		validatorDeactivatedTimeSlot, slotGasUsed := getValidatorDeactivatedTimeSlot(validatorID)
		gasUsed += slotGasUsed
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorDeactivatedTimeSlot), common.BigToHash(evm.Context.Time))

		// Update the total active stake
		validatorReceivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(validatorID)
		gasUsed += slotGasUsed
		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot))
		receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

		totalActiveStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
		totalActiveStakeBigInt := new(big.Int).SetBytes(totalActiveStakeState.Bytes())
		newTotalActiveStake := new(big.Int).Sub(totalActiveStakeBigInt, receivedStakeBigInt)
		if newTotalActiveStake.Cmp(big.NewInt(0)) < 0 {
			newTotalActiveStake = big.NewInt(0)
		}
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)), common.BigToHash(newTotalActiveStake))
	} else if !isActive && hasSelfStake && hasEnoughSelfStake && withinDelegatedLimit {
		// Set the validator as active
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorStatusSlot), common.BigToHash(big.NewInt(0))) // OK_STATUS

		// Clear the validator deactivated epoch
		validatorDeactivatedEpochSlot, _ := getValidatorDeactivatedEpochSlot(validatorID)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorDeactivatedEpochSlot), common.BigToHash(big.NewInt(0)))

		// Clear the validator deactivated time
		validatorDeactivatedTimeSlot, _ := getValidatorDeactivatedTimeSlot(validatorID)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorDeactivatedTimeSlot), common.BigToHash(big.NewInt(0)))

		// Update the total active stake
		validatorReceivedStakeSlot, _ := getValidatorReceivedStakeSlot(validatorID)
		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot))
		receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

		totalActiveStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
		totalActiveStakeBigInt := new(big.Int).SetBytes(totalActiveStakeState.Bytes())
		newTotalActiveStake := new(big.Int).Add(totalActiveStakeBigInt, receivedStakeBigInt)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)), common.BigToHash(newTotalActiveStake))
	}

	return nil, 0, nil
}
