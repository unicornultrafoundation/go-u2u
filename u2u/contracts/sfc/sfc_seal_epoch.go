package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/log"
)

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
	// Get the current epoch (corresponds to snapshot in Solidity)
	currentEpochBigInt, epochGasUsed, err := getCurrentEpoch(evm)
	gasUsed += epochGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Get the epoch snapshot slot for the current epoch
	// This is equivalent to "snapshot" in the Solidity code
	currentEpochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(currentEpochBigInt)
	gasUsed += slotGasUsed

	// Get the validator IDs for the current epoch
	// For a dynamic array in a struct, we first get the length from the slot
	validatorIDsOffsetBig := GetBigInt().SetInt64(validatorIDsOffset)
	validatorIDsSlot := GetBigInt().Add(currentEpochSnapshotSlot, validatorIDsOffsetBig)
	defer PutBigInt(validatorIDsSlot)
	validatorIDsLengthHash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorIDsSlot))
	gasUsed += SloadGasCost
	PutBigInt(validatorIDsOffsetBig)

	// Convert the length hash to a big.Int
	validatorIDsLengthBig := GetBigInt().SetBytes(validatorIDsLengthHash.Bytes())
	validatorIDsLength := validatorIDsLengthBig.Uint64()
	defer PutBigInt(validatorIDsLengthBig)

	// Calculate the base slot for the array elements
	// The array elements start at keccak256(slot)
	validatorIDsBaseSlotBytes := CachedKeccak256Hash(common.BigToHash(validatorIDsSlot).Bytes()).Bytes()
	gasUsed += HashGasCost
	validatorIDsBaseSlot := GetBigInt().SetBytes(validatorIDsBaseSlotBytes)
	defer PutBigInt(validatorIDsBaseSlot)

	// Read each validator ID from storage
	validatorIDs := make([]*big.Int, 0, validatorIDsLength)
	for i := uint64(0); i < validatorIDsLength; i++ {
		// Calculate the slot for this array element: baseSlot + i
		elementSlot := GetBigInt().Add(validatorIDsBaseSlot, big.NewInt(int64(i)))

		// Get the validator ID from storage
		validatorIDHash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(elementSlot))
		gasUsed += SloadGasCost
		PutBigInt(elementSlot)

		// Convert the hash to a big.Int and add it to the list
		validatorID := new(big.Int).SetBytes(validatorIDHash.Bytes())
		validatorIDs = append(validatorIDs, validatorID)
	}

	// Call _sealEpoch_offline
	offlineGasUsed, err := _sealEpoch_offline(evm, validatorIDs, offlineTimes, offlineBlocks, currentEpochSnapshotSlot)
	gasUsed += offlineGasUsed
	if err != nil {
		return nil, gasUsed, err
	}

	// Get the previous epoch (corresponds to prevSnapshot in Solidity)
	// In Solidity, this is "currentSealedEpoch"
	currentSealedEpochHash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)))
	gasUsed += SloadGasCost
	prevEpochBigInt := GetBigInt().SetBytes(currentSealedEpochHash.Bytes())
	defer PutBigInt(prevEpochBigInt)

	// Get the epoch snapshot slot for the previous epoch
	// This is equivalent to "prevSnapshot" in the Solidity code
	prevEpochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(prevEpochBigInt)
	gasUsed += slotGasUsed

	// Get the end time of the previous epoch
	prevEndTimeSlot := GetBigInt().Add(prevEpochSnapshotSlot, big.NewInt(endTimeOffset))
	prevEndTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(prevEndTimeSlot))
	gasUsed += SloadGasCost
	PutBigInt(prevEndTimeSlot)
	prevEndTimeBigInt := GetBigInt().SetBytes(prevEndTime.Bytes())

	// Calculate epoch duration
	epochDuration := GetBigInt().SetInt64(1) // Default to 1 if current time <= prevEndTime
	if evm.Context.Time.Cmp(prevEndTimeBigInt) > 0 {
		epochDuration = epochDuration.Sub(evm.Context.Time, prevEndTimeBigInt)
	}
	PutBigInt(prevEndTimeBigInt)

	// Call _sealEpoch_rewards
	// In Solidity: _sealEpoch_rewards(epochDuration, snapshot, prevSnapshot, validatorIDs, uptimes, originatedTxsFee)
	rewardsGasUsed, err := _sealEpoch_rewards(evm, epochDuration, currentEpochBigInt, prevEpochBigInt, validatorIDs, uptimes, originatedTxsFee)
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

	// Update epoch snapshot end time (snapshot.endTime = _now() in Solidity)
	endTimeOffsetBig := GetBigInt().SetInt64(endTimeOffset)
	endTimeSlot := GetBigInt().Add(currentEpochSnapshotSlot, endTimeOffsetBig)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(endTimeSlot), common.BigToHash(evm.Context.Time))
	gasUsed += SstoreGasCost
	PutBigInt(endTimeOffsetBig)
	PutBigInt(endTimeSlot)

	// Get the base reward per second from the constants manager
	baseRewardPerSecond := getConstantsManagerVariable("baseRewardPerSecond")

	// Update epoch snapshot base reward per second (snapshot.baseRewardPerSecond = c.baseRewardPerSecond() in Solidity)
	baseRewardPerSecondOffsetBig := GetBigInt().SetInt64(baseRewardPerSecondOffset)
	baseRewardPerSecondSlot := GetBigInt().Add(currentEpochSnapshotSlot, baseRewardPerSecondOffsetBig)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(baseRewardPerSecondSlot), common.BigToHash(baseRewardPerSecond))
	gasUsed += SstoreGasCost
	PutBigInt(baseRewardPerSecondOffsetBig)
	PutBigInt(baseRewardPerSecondSlot)

	// Get the total supply
	totalSupplySlotBig := GetBigInt().SetInt64(totalSupplySlot)
	totalSupply := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(totalSupplySlotBig))
	gasUsed += SloadGasCost
	totalSupplyBigInt := GetBigInt().SetBytes(totalSupply.Bytes())

	// Update epoch snapshot total supply (snapshot.totalSupply = totalSupply in Solidity)
	totalSupplyOffsetBig := GetBigInt().SetInt64(totalSupplyOffset)
	totalSupplySnapshotSlot := GetBigInt().Add(currentEpochSnapshotSlot, totalSupplyOffsetBig)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(totalSupplySnapshotSlot), common.BigToHash(totalSupplyBigInt))
	gasUsed += SstoreGasCost
	PutBigInt(totalSupplyBigInt)
	PutBigInt(totalSupplySlotBig)
	PutBigInt(totalSupplyOffsetBig)
	PutBigInt(totalSupplySnapshotSlot)
	PutBigInt(epochDuration)

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

	// Get the existing total stake for the snapshot
	totalStakeSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(totalStakeOffset))
	totalStakeHash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(totalStakeSlot))
	gasUsed += SloadGasCost
	totalStake := new(big.Int).SetBytes(totalStakeHash.Bytes())

	// Fill data for the next snapshot
	// This corresponds to the loop in the Solidity implementation that sets receivedStake and adds to totalStake
	for _, validatorID := range nextValidatorIDs {
		// Get the validator's received stake from getValidator[validatorID].receivedStake
		validatorReceivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(validatorID)
		gasUsed += slotGasUsed

		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot))
		gasUsed += SloadGasCost
		receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

		// Set the received stake for this validator in the epoch snapshot (snapshot.receivedStake[validatorID] = receivedStake)
		validatorReceivedStakeEpochSlot, slotGasUsed := getEpochValidatorReceivedStakeSlot(currentEpochBigInt, validatorID)
		gasUsed += slotGasUsed

		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorReceivedStakeEpochSlot), common.BigToHash(receivedStakeBigInt))
		gasUsed += SstoreGasCost

		// Add to total stake (snapshot.totalStake = snapshot.totalStake.add(receivedStake))
		totalStake = new(big.Int).Add(totalStake, receivedStakeBigInt)
	}

	// Set the validator IDs for the epoch snapshot
	// For a dynamic array in a struct, we first set the length at the slot
	validatorIDsSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(validatorIDsOffset))
	validatorIDsLength := big.NewInt(int64(len(nextValidatorIDs)))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(validatorIDsSlot), common.BigToHash(validatorIDsLength))
	gasUsed += SstoreGasCost

	// Calculate the base slot for the array elements
	// The array elements start at keccak256(slot)
	validatorIDsBaseSlotBytes := crypto.Keccak256(common.BigToHash(validatorIDsSlot).Bytes())
	gasUsed += HashGasCost
	validatorIDsBaseSlot := new(big.Int).SetBytes(validatorIDsBaseSlotBytes)

	// Store each validator ID in the validatorIDs array of the epoch snapshot
	// This corresponds to `snapshot.validatorIDs = nextValidatorIDs` in the Solidity implementation
	for i, validatorID := range nextValidatorIDs {
		// Calculate the slot for this array element: baseSlot + i
		elementSlot := new(big.Int).Add(validatorIDsBaseSlot, big.NewInt(int64(i)))

		// Store the validator ID
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(elementSlot), common.BigToHash(validatorID))
		gasUsed += SstoreGasCost
	}

	// Set the updated total stake for the epoch snapshot
	// We've already calculated the totalStakeSlot above and updated totalStake in the loop
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(totalStakeSlot), common.BigToHash(totalStake))
	gasUsed += SstoreGasCost

	// Update the minimum gas price in the node
	// This corresponds to `node.updateMinGasPrice(minGasPrice)` in the Solidity implementation
	minGasPrice := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)))
	gasUsed += SloadGasCost
	minGasPriceBigInt := new(big.Int).SetBytes(minGasPrice.Bytes())

	// Get the node driver auth address to call updateMinGasPrice
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
	result, _, err := evm.Call(vm.AccountRef(ContractAddress), nodeDriverAuthAddr, data, 50000, big.NewInt(0))
	if err != nil {
		reason, _ := abi.UnpackRevert(result)
		log.Error("SFC: Error calling NodeDriverAuth method", "method", "updateMinGasPrice", "err", err, "reason", reason)
		return nil, gasUsed, err
	}

	return nil, gasUsed, nil
}

// _sealEpoch_offline is an internal function to seal offline validators in an epoch
func _sealEpoch_offline(evm *vm.EVM, validatorIDs []*big.Int, offlineTimes []*big.Int, offlineBlocks []*big.Int, currentEpochSnapshotSlot *big.Int) (uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Get the offline penalty thresholds from the constants manager
	offlinePenaltyThresholdBlocksNum, thresholdGasUsed, err := getOfflinePenaltyThresholdBlocksNum(evm)
	gasUsed += thresholdGasUsed
	if err != nil {
		return gasUsed, err
	}

	offlinePenaltyThresholdTime, thresholdGasUsed, err := getOfflinePenaltyThresholdTime(evm)
	gasUsed += thresholdGasUsed
	if err != nil {
		return gasUsed, err
	}

	// Iterate through validators
	for i, validatorID := range validatorIDs {
		// Check if the validator exceeds the offline thresholds
		if offlineBlocks[i].Cmp(offlinePenaltyThresholdBlocksNum) > 0 && offlineTimes[i].Cmp(offlinePenaltyThresholdTime) >= 0 {
			// Deactivate the validator with OFFLINE_BIT
			deactivateGasUsed, err := _setValidatorDeactivated(evm, validatorID, OFFLINE_BIT)
			gasUsed += deactivateGasUsed
			if err != nil {
				return gasUsed, err
			}

			// Sync the validator
			syncGasUsed, err := _syncValidator(evm, validatorID, false)
			gasUsed += syncGasUsed
			if err != nil {
				return gasUsed, err
			}
		}

		// Add the offset for the offlineTime mapping within the struct
		mappingSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(offlineTimeOffset))

		// Then, calculate the slot for the specific key using our helper function
		// Use CreateValidatorMappingHashInput to create the hash input
		outerHashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)
		// Use cached hash calculation
		offlineTimeSlotHash := CachedKeccak256Hash(outerHashInput)
		offlineTimeSlot := new(big.Int).SetBytes(offlineTimeSlotHash.Bytes())
		gasUsed += HashGasCost

		// Set the value in the state
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(offlineTimeSlot), common.BigToHash(offlineTimes[i]))
		gasUsed += SstoreGasCost

		// Store offline blocks in the epoch snapshot (snapshot.offlineBlocks[validatorID] = offlineBlocks[i])
		// For a mapping within a struct, we need to calculate the slot as:
		// keccak256(key . (struct_slot + offset))
		// Add the offset for the offlineBlocks mapping within the struct
		blocksMappingSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(offlineBlocksOffset))

		// Then, calculate the slot for the specific key using our helper function
		// Use CreateValidatorMappingHashInput to create the hash input
		blockHashInput := CreateValidatorMappingHashInput(validatorID, blocksMappingSlot)
		// Use cached hash calculation
		offlineBlocksSlotHash := CachedKeccak256Hash(blockHashInput)
		offlineBlocksSlot := new(big.Int).SetBytes(offlineBlocksSlotHash.Bytes())
		gasUsed += HashGasCost

		// Set the value in the state
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(offlineBlocksSlot), common.BigToHash(offlineBlocks[i]))
		gasUsed += SstoreGasCost
	}

	return gasUsed, nil
}

// _sealEpoch_rewards is an internal function to seal rewards in an epoch
func _sealEpoch_rewards(evm *vm.EVM, epochDuration *big.Int, currentEpoch *big.Int, prevEpoch *big.Int,
	validatorIDs []*big.Int, uptimes []*big.Int, accumulatedOriginatedTxsFee []*big.Int) (uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Declare variables to avoid redeclaration issues
	var innerHash []byte
	var outerHashInput []byte
	var outerHash []byte

	// Pre-calculate the epoch snapshot base slots for current and previous epochs
	currentEpochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(currentEpoch)
	gasUsed += slotGasUsed

	prevEpochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(prevEpoch)
	gasUsed += slotGasUsed

	// Initialize context for rewards calculation
	baseRewardWeights := make([]*big.Int, len(validatorIDs))
	totalBaseRewardWeight := big.NewInt(0)
	txRewardWeights := make([]*big.Int, len(validatorIDs))
	totalTxRewardWeight := big.NewInt(0)
	epochFee := big.NewInt(0)

	// Calculate tx reward weights and epoch fee
	for i, validatorID := range validatorIDs {
		// Get previous accumulated originated txs fee
		// For a mapping within a struct, we need to calculate the slot as:
		// keccak256(abi.encode(validatorID, keccak256(abi.encode(accumulatedOriginatedTxsFeeOffset, prevEpochSnapshotSlot))))
		// Use our helper function to create the hash input from offset and slot
		innerHash = CreateAndHashOffsetSlot(accumulatedOriginatedTxsFeeOffset, prevEpochSnapshotSlot)
		gasUsed += HashGasCost

		// Use our helper function to create a nested hash input
		outerHashInput = CreateNestedHashInput(validatorID, innerHash)
		// Use cached hash calculation
		outerHash = CachedKeccak256(outerHashInput)
		gasUsed += HashGasCost

		prevAccumulatedTxsFeeSlot := new(big.Int).SetBytes(outerHash)

		prevAccumulatedTxsFee := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(prevAccumulatedTxsFeeSlot))
		gasUsed += SloadGasCost
		prevAccumulatedTxsFeeBigInt := new(big.Int).SetBytes(prevAccumulatedTxsFee.Bytes())

		// Calculate originated txs fee for this epoch
		originatedTxsFee := big.NewInt(0)
		if accumulatedOriginatedTxsFee[i].Cmp(prevAccumulatedTxsFeeBigInt) > 0 {
			originatedTxsFee = new(big.Int).Sub(accumulatedOriginatedTxsFee[i], prevAccumulatedTxsFeeBigInt)
		}

		// Calculate tx reward weight: originatedTxsFee * uptime / epochDuration
		txRewardWeight := new(big.Int).Mul(originatedTxsFee, uptimes[i])
		txRewardWeight = new(big.Int).Div(txRewardWeight, epochDuration)
		txRewardWeights[i] = txRewardWeight

		// Update total tx reward weight
		totalTxRewardWeight = new(big.Int).Add(totalTxRewardWeight, txRewardWeight)

		// Update epoch fee
		epochFee = new(big.Int).Add(epochFee, originatedTxsFee)
	}

	// Calculate base reward weights
	for i, validatorID := range validatorIDs {
		// Get validator's received stake
		receivedStakeSlot, slotGasUsed := getEpochValidatorReceivedStakeSlot(currentEpoch, validatorID)
		gasUsed += slotGasUsed

		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(receivedStakeSlot))
		gasUsed += SloadGasCost
		receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

		// Calculate base reward weight: (stake * uptime / epochDuration) * (uptime / epochDuration)
		term1 := new(big.Int).Mul(receivedStakeBigInt, uptimes[i])
		term1 = new(big.Int).Div(term1, epochDuration)
		term2 := new(big.Int).Mul(term1, uptimes[i])
		baseRewardWeight := new(big.Int).Div(term2, epochDuration)
		baseRewardWeights[i] = baseRewardWeight

		// Update total base reward weight
		totalBaseRewardWeight = new(big.Int).Add(totalBaseRewardWeight, baseRewardWeight)
	}

	baseRewardPerSecond := getConstantsManagerVariable("baseRewardPerSecond")
	validatorCommission := getConstantsManagerVariable("validatorCommission")
	burntFeeShare := getConstantsManagerVariable("burntFeeShare")
	treasuryFeeShare := getConstantsManagerVariable("treasuryFeeShare")

	// Calculate rewards for each validator
	for i, validatorID := range validatorIDs {
		// Calculate raw base reward
		rawBaseReward := big.NewInt(0)
		if baseRewardWeights[i].Cmp(big.NewInt(0)) > 0 {
			totalReward := new(big.Int).Mul(epochDuration, baseRewardPerSecond)
			rawBaseReward = new(big.Int).Mul(totalReward, baseRewardWeights[i])
			rawBaseReward = new(big.Int).Div(rawBaseReward, totalBaseRewardWeight)
		}

		// Calculate raw tx reward
		rawTxReward := big.NewInt(0)
		if txRewardWeights[i].Cmp(big.NewInt(0)) > 0 {
			// Calculate fee reward except burntFeeShare and treasuryFeeShare
			txReward := new(big.Int).Mul(epochFee, txRewardWeights[i])
			txReward = new(big.Int).Div(txReward, totalTxRewardWeight)

			// Subtract burnt and treasury shares
			shareToSubtract := new(big.Int).Add(burntFeeShare, treasuryFeeShare)
			shareToKeep := new(big.Int).Sub(unit, shareToSubtract)

			rawTxReward = new(big.Int).Mul(txReward, shareToKeep)
			rawTxReward = new(big.Int).Div(rawTxReward, unit)
		}

		// Calculate total raw reward
		rawReward := new(big.Int).Add(rawBaseReward, rawTxReward)

		// Get validator auth address
		validatorAuthSlot, slotGasUsed := getValidatorAuthSlot(validatorID)
		gasUsed += slotGasUsed

		validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorAuthSlot))
		gasUsed += SloadGasCost
		validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

		// Calculate validator's commission
		commissionRewardFull := new(big.Int).Mul(rawReward, validatorCommission)
		commissionRewardFull = new(big.Int).Div(commissionRewardFull, unit)

		// Get validator's self-stake
		selfStakeSlot, slotGasUsed := getStakeSlot(validatorAuthAddr, validatorID)
		gasUsed += slotGasUsed

		selfStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(selfStakeSlot))
		gasUsed += SloadGasCost
		selfStakeBigInt := new(big.Int).SetBytes(selfStake.Bytes())

		// Process commission reward if self-stake is not zero
		if selfStakeBigInt.Cmp(big.NewInt(0)) != 0 {
			// Get locked stake
			lockedStakeSlot, slotGasUsed := getLockedStakeSlot(validatorAuthAddr, validatorID)
			gasUsed += slotGasUsed

			lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockedStakeSlot))
			gasUsed += SloadGasCost
			lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())

			// Calculate locked and unlocked commission rewards
			lCommissionRewardFull := new(big.Int).Mul(commissionRewardFull, lockedStakeBigInt)
			lCommissionRewardFull = new(big.Int).Div(lCommissionRewardFull, selfStakeBigInt)

			// Unused in current implementation, but kept for future use
			_ = new(big.Int).Sub(commissionRewardFull, lCommissionRewardFull)

			// Get lockup duration
			lockupDurationSlot, slotGasUsed := getLockupDurationSlot(validatorAuthAddr, validatorID)
			gasUsed += slotGasUsed

			lockupDuration := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupDurationSlot))
			gasUsed += SloadGasCost
			lockupDurationBigInt := new(big.Int).SetBytes(lockupDuration.Bytes())

			// Scale lockup rewards
			reward, scaleGasUsed, err := _scaleLockupReward(evm, lCommissionRewardFull, lockupDurationBigInt)
			gasUsed += scaleGasUsed
			if err != nil {
				return gasUsed, err
			}

			// Scale lockup reward for unlocked commission
			uCommissionRewardFull := new(big.Int).Sub(commissionRewardFull, lCommissionRewardFull)
			uReward, scaleGasUsed, err := _scaleLockupReward(evm, uCommissionRewardFull, big.NewInt(0))
			gasUsed += scaleGasUsed
			if err != nil {
				return gasUsed, err
			}

			// Get current rewards stash
			// The Rewards struct has three fields stored at consecutive slots
			rewardsStashSlot, slotGasUsed := getRewardsStashSlot(validatorAuthAddr, validatorID)
			gasUsed += slotGasUsed

			// Get lockupExtraReward (first field)
			lockupExtraReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(rewardsStashSlot))
			gasUsed += SloadGasCost

			// Get lockupBaseReward (second field)
			lockupBaseRewardSlot := new(big.Int).Add(rewardsStashSlot, big.NewInt(1))
			lockupBaseReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(lockupBaseRewardSlot))
			gasUsed += SloadGasCost

			// Get unlockedReward (third field)
			unlockedRewardSlot := new(big.Int).Add(rewardsStashSlot, big.NewInt(2))
			unlockedReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(unlockedRewardSlot))
			gasUsed += SloadGasCost

			// Convert the rewards stash to a Rewards struct
			currentRewardsStash := Rewards{
				LockupExtraReward: new(big.Int).SetBytes(lockupExtraReward.Bytes()),
				LockupBaseReward:  new(big.Int).SetBytes(lockupBaseReward.Bytes()),
				UnlockedReward:    new(big.Int).SetBytes(unlockedReward.Bytes()),
			}

			// Use sumRewards to add the rewards
			newRewardsStash := sumRewards(currentRewardsStash, reward, uReward)

			// Store each field of the Rewards struct separately
			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(rewardsStashSlot), common.BigToHash(newRewardsStash.LockupExtraReward))
			gasUsed += SstoreGasCost

			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(lockupBaseRewardSlot), common.BigToHash(newRewardsStash.LockupBaseReward))
			gasUsed += SstoreGasCost

			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(unlockedRewardSlot), common.BigToHash(newRewardsStash.UnlockedReward))
			gasUsed += SstoreGasCost

			// Update stashed lockup rewards
			// The Rewards struct has three fields stored at consecutive slots
			stashedLockupRewardsSlot, slotGasUsed := getStashedLockupRewardsSlot(validatorAuthAddr, validatorID)
			gasUsed += slotGasUsed

			// Get lockupExtraReward (first field)
			stashedLockupExtraReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stashedLockupRewardsSlot))
			gasUsed += SloadGasCost

			// Get lockupBaseReward (second field)
			stashedLockupBaseRewardSlot := new(big.Int).Add(stashedLockupRewardsSlot, big.NewInt(1))
			stashedLockupBaseReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stashedLockupBaseRewardSlot))
			gasUsed += SloadGasCost

			// Get unlockedReward (third field)
			stashedUnlockedRewardSlot := new(big.Int).Add(stashedLockupRewardsSlot, big.NewInt(2))
			stashedUnlockedReward := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(stashedUnlockedRewardSlot))
			gasUsed += SloadGasCost

			// Convert the stashed lockup rewards to a Rewards struct
			currentStashedLockupRewards := Rewards{
				LockupExtraReward: new(big.Int).SetBytes(stashedLockupExtraReward.Bytes()),
				LockupBaseReward:  new(big.Int).SetBytes(stashedLockupBaseReward.Bytes()),
				UnlockedReward:    new(big.Int).SetBytes(stashedUnlockedReward.Bytes()),
			}

			// Use sumRewards to add the rewards
			newStashedLockupRewards := sumRewards(currentStashedLockupRewards, reward, uReward)

			// Store each field of the Rewards struct separately
			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stashedLockupRewardsSlot), common.BigToHash(newStashedLockupRewards.LockupExtraReward))
			gasUsed += SstoreGasCost

			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stashedLockupBaseRewardSlot), common.BigToHash(newStashedLockupRewards.LockupBaseReward))
			gasUsed += SstoreGasCost

			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(stashedUnlockedRewardSlot), common.BigToHash(newStashedLockupRewards.UnlockedReward))
			gasUsed += SstoreGasCost
		}

		// Calculate delegators' reward
		delegatorsReward := new(big.Int).Sub(rawReward, commissionRewardFull)

		// Get validator's received stake
		receivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(validatorID)
		gasUsed += slotGasUsed

		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(receivedStakeSlot))
		gasUsed += SloadGasCost
		receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

		// Calculate reward per token
		rewardPerToken := big.NewInt(0)
		if receivedStakeBigInt.Cmp(big.NewInt(0)) != 0 {
			rewardPerToken = new(big.Int).Mul(delegatorsReward, unit)
			rewardPerToken = new(big.Int).Div(rewardPerToken, receivedStakeBigInt)
		}

		// Update accumulated reward per token
		// For a mapping within a struct, we need to calculate the slot as:
		// keccak256(key . (struct_slot + offset))
		// Add the offset for the accumulatedRewardPerToken mapping within the struct
		mappingSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(accumulatedRewardPerTokenOffset))

		// Then, calculate the slot for the specific key using our helper function
		// Use CreateValidatorMappingHashInput to create the hash input
		// Declare outerHashInput at the beginning of the function
		outerHashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)
		// Use cached hash calculation
		accumulatedRewardPerTokenSlotHash := CachedKeccak256Hash(outerHashInput)
		accumulatedRewardPerTokenSlot := new(big.Int).SetBytes(accumulatedRewardPerTokenSlotHash.Bytes())
		gasUsed += HashGasCost

		// For the previous epoch
		// For a mapping within a struct, we need to calculate the slot as:
		// keccak256(key . (struct_slot + offset))
		// Add the offset for the accumulatedRewardPerToken mapping within the struct
		prevMappingSlot := new(big.Int).Add(prevEpochSnapshotSlot, big.NewInt(accumulatedRewardPerTokenOffset))

		// Then, calculate the slot for the specific key using our helper function
		// Use CreateValidatorMappingHashInput to create the hash input
		outerHashInput = CreateValidatorMappingHashInput(validatorID, prevMappingSlot)
		// Use cached hash calculation
		prevAccumulatedRewardPerTokenSlotHash := CachedKeccak256Hash(outerHashInput)
		prevAccumulatedRewardPerTokenSlot := new(big.Int).SetBytes(prevAccumulatedRewardPerTokenSlotHash.Bytes())
		gasUsed += HashGasCost

		prevAccumulatedRewardPerToken := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(prevAccumulatedRewardPerTokenSlot))
		gasUsed += SloadGasCost
		prevAccumulatedRewardPerTokenBigInt := new(big.Int).SetBytes(prevAccumulatedRewardPerToken.Bytes())

		newAccumulatedRewardPerToken := new(big.Int).Add(prevAccumulatedRewardPerTokenBigInt, rewardPerToken)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(accumulatedRewardPerTokenSlot), common.BigToHash(newAccumulatedRewardPerToken))
		gasUsed += SstoreGasCost

		// Update accumulated originated txs fee (snapshot.accumulatedOriginatedTxsFee[validatorID] = accumulatedOriginatedTxsFee[i])
		// Use our helper function to create the hash input from offset and slot
		innerHash = CreateAndHashOffsetSlot(accumulatedOriginatedTxsFeeOffset, currentEpochSnapshotSlot)
		gasUsed += HashGasCost

		// Use our helper function to create a nested hash input
		outerHashInput = CreateNestedHashInput(validatorID, innerHash)
		// Use cached hash calculation
		outerHash = CachedKeccak256(outerHashInput)
		gasUsed += HashGasCost

		// Update accumulated originated txs fee (snapshot.accumulatedOriginatedTxsFee[validatorID] = accumulatedOriginatedTxsFee[i])
		// For a mapping within a struct, we need to calculate the slot as:
		// keccak256(key . (struct_slot + offset))
		// Add the offset for the accumulatedOriginatedTxsFee mapping within the struct
		originatedTxsFeeSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(accumulatedOriginatedTxsFeeOffset))

		// Then, calculate the slot for the specific key using our helper function
		// Use CreateValidatorMappingHashInput to create the hash input
		outerHashInput = CreateValidatorMappingHashInput(validatorID, originatedTxsFeeSlot)
		// Use cached hash calculation
		accumulatedOriginatedTxsFeeSlotHash := CachedKeccak256Hash(outerHashInput)
		accumulatedOriginatedTxsFeeSlot := new(big.Int).SetBytes(accumulatedOriginatedTxsFeeSlotHash.Bytes())
		gasUsed += HashGasCost

		// Set the value in the state
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(accumulatedOriginatedTxsFeeSlot), common.BigToHash(accumulatedOriginatedTxsFee[i]))
		gasUsed += SstoreGasCost

		// Update accumulated uptime
		// For a mapping within a struct, we need to calculate the slot as:
		// keccak256(key . (struct_slot + offset))
		// Add the offset for the accumulatedUptime mapping within the struct
		uptimeMappingSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(accumulatedUptimeOffset))

		// Then, calculate the slot for the specific key using our helper function
		// Use CreateValidatorMappingHashInput to create the hash input
		outerHashInput = CreateValidatorMappingHashInput(validatorID, uptimeMappingSlot)
		// Use cached hash calculation
		accumulatedUptimeSlotHash := CachedKeccak256Hash(outerHashInput)
		accumulatedUptimeSlot := new(big.Int).SetBytes(accumulatedUptimeSlotHash.Bytes())
		gasUsed += HashGasCost

		// For the previous epoch
		// For a mapping within a struct, we need to calculate the slot as:
		// keccak256(key . (struct_slot + offset))
		// Add the offset for the accumulatedUptime mapping within the struct
		prevUptimeMappingSlot := new(big.Int).Add(prevEpochSnapshotSlot, big.NewInt(accumulatedUptimeOffset))

		// Then, calculate the slot for the specific key using our helper function
		// Use CreateValidatorMappingHashInput to create the hash input
		outerHashInput = CreateValidatorMappingHashInput(validatorID, prevUptimeMappingSlot)
		// Use cached hash calculation
		prevAccumulatedUptimeSlotHash := CachedKeccak256Hash(outerHashInput)
		prevAccumulatedUptimeSlot := new(big.Int).SetBytes(prevAccumulatedUptimeSlotHash.Bytes())
		gasUsed += HashGasCost

		prevAccumulatedUptime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(prevAccumulatedUptimeSlot))
		gasUsed += SloadGasCost
		prevAccumulatedUptimeBigInt := new(big.Int).SetBytes(prevAccumulatedUptime.Bytes())

		newAccumulatedUptime := new(big.Int).Add(prevAccumulatedUptimeBigInt, uptimes[i])
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(accumulatedUptimeSlot), common.BigToHash(newAccumulatedUptime))
		gasUsed += SstoreGasCost
	}
	// Update epoch fee
	epochFeeSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(epochFeeOffset))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(epochFeeSlot), common.BigToHash(epochFee))
	gasUsed += SstoreGasCost

	// Update total base reward weight
	totalBaseRewardSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(totalBaseRewardOffset))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(totalBaseRewardSlot), common.BigToHash(totalBaseRewardWeight))
	gasUsed += SstoreGasCost

	// Update total tx reward weight
	totalTxRewardSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(totalTxRewardOffset))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(totalTxRewardSlot), common.BigToHash(totalTxRewardWeight))
	gasUsed += SstoreGasCost

	// Update total supply
	totalSupply := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)))
	gasUsed += SloadGasCost
	totalSupplyBigInt := new(big.Int).SetBytes(totalSupply.Bytes())

	// Subtract epoch fee from total supply
	if totalSupplyBigInt.Cmp(epochFee) > 0 {
		totalSupplyBigInt = new(big.Int).Sub(totalSupplyBigInt, epochFee)
	} else {
		totalSupplyBigInt = big.NewInt(0)
	}

	// Update total supply
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)), common.BigToHash(totalSupplyBigInt))
	gasUsed += SstoreGasCost

	// Transfer 10% of fees to treasury if treasury address is set
	treasuryAddress := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(treasuryAddressSlot)))
	gasUsed += SloadGasCost
	treasuryAddressBytes := treasuryAddress.Bytes()

	// Check if treasury address is not zero
	emptyAddr := common.Address{}
	treasuryAddr := common.BytesToAddress(treasuryAddressBytes)
	if treasuryAddr.Cmp(emptyAddr) != 0 {
		// Calculate fee share
		feeShare := new(big.Int).Mul(epochFee, treasuryFeeShare)
		feeShare = new(big.Int).Div(feeShare, unit)

		// First mint native token to the contract itself
		// This matches the Solidity code: _mintNativeToken(feeShare);
		mintGasUsed, err := _mintNativeToken(evm, ContractAddress, feeShare)
		gasUsed += mintGasUsed
		if err != nil {
			return gasUsed, err
		}

		// Then make a call to transfer the tokens to the treasury address
		// This simulates the Solidity code: treasuryAddress.call.value(feeShare)("");
		callData := []byte{} // Empty call data
		_, _, err = evm.Call(
			vm.AccountRef(ContractAddress), // Caller
			treasuryAddr,                   // Target address
			callData,                       // Call data (empty)
			21000,                          // Gas limit for a simple transfer
			feeShare,                       // Value to transfer
		)
		if err != nil {
			return gasUsed, err
		}
		gasUsed += 21000 // Add gas for the transfer
	}

	return gasUsed, nil
}

// _sealEpoch_minGasPrice is an internal function to seal minimum gas price in an epoch
func _sealEpoch_minGasPrice(evm *vm.EVM, epochDuration *big.Int, epochGas *big.Int) (uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Get the target gas power per second from the constants manager
	targetGasPowerPerSecond := getConstantsManagerVariable("targetGasPowerPerSecond")

	// Calculate target epoch gas
	targetEpochGas := new(big.Int).Mul(epochDuration, targetGasPowerPerSecond)
	targetEpochGas.Add(targetEpochGas, big.NewInt(1)) // Add 1 to avoid division by zero

	// Get the decimal unit (1e18) using the helper function
	decimalUnitBigInt := getDecimalUnit()

	// Calculate gas price delta ratio
	gasPriceDeltaRatio := new(big.Int).Mul(epochGas, decimalUnitBigInt)
	gasPriceDeltaRatio.Div(gasPriceDeltaRatio, targetEpochGas)

	// Get the gas price balancing counterweight from the constants manager
	counterweight := getConstantsManagerVariable("gasPriceBalancingCounterweight")

	// Scale down the change speed
	// gasPriceDeltaRatio = (epochDuration * gasPriceDeltaRatio + counterweight * Decimal.unit()) / (epochDuration + counterweight)
	term1 := new(big.Int).Mul(epochDuration, gasPriceDeltaRatio)
	term2 := new(big.Int).Mul(counterweight, decimalUnitBigInt)
	numerator := new(big.Int).Add(term1, term2)
	denominator := new(big.Int).Add(epochDuration, counterweight)
	gasPriceDeltaRatio = new(big.Int).Div(numerator, denominator)

	// Limit the max/min possible delta in one epoch using the trimGasPriceChangeRatio helper function
	gasPriceDeltaRatio = trimGasPriceChangeRatio(gasPriceDeltaRatio)

	// Get the current min gas price
	minGasPrice := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)))
	gasUsed += SloadGasCost
	minGasPriceBigInt := new(big.Int).SetBytes(minGasPrice.Bytes())

	// Apply the ratio
	newMinGasPrice := new(big.Int).Mul(minGasPriceBigInt, gasPriceDeltaRatio)
	newMinGasPrice.Div(newMinGasPrice, decimalUnitBigInt)

	// Limit the max/min possible minGasPrice using the trimMinGasPrice helper function
	newMinGasPrice = trimMinGasPrice(newMinGasPrice)

	// Apply new minGasPrice
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)), common.BigToHash(newMinGasPrice))
	gasUsed += SstoreGasCost

	return gasUsed, nil
}
