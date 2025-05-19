package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
)

// SealEpoch seals the current epoch
func handleSealEpoch(evm *vm.EVM, caller common.Address, args []interface{}) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Initialize epoch cache
	var epochCache *EpochStateCache
	if evm.EpochCache == nil {
		evm.EpochCache = NewEpochStateCache(evm.SfcStateDB)
	}
	epochCache = evm.EpochCache.(*EpochStateCache)

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

	// Initialize epoch cache if not exists
	epochCache.Initialize(currentEpochBigInt)

	// Get the epoch snapshot slot for the current epoch
	// This is equivalent to "snapshot" in the Solidity code
	currentEpochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(currentEpochBigInt)
	gasUsed += slotGasUsed

	// Get the validator IDs for the current epoch
	// For a dynamic array in a struct, we first get the length from the slot
	validatorIDsOffsetBig := GetBigInt().SetInt64(validatorIDsOffset)
	validatorIDsSlot := GetBigInt().Add(currentEpochSnapshotSlot, validatorIDsOffsetBig)
	defer PutBigInt(validatorIDsSlot)
	validatorIDsLengthHash := epochCache.GetState(ContractAddress, common.BigToHash(validatorIDsSlot))
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
		validatorIDHash := epochCache.GetState(ContractAddress, common.BigToHash(elementSlot))
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
	currentSealedEpochHash := epochCache.GetState(ContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)))
	gasUsed += SloadGasCost
	prevEpochBigInt := GetBigInt().SetBytes(currentSealedEpochHash.Bytes())
	defer PutBigInt(prevEpochBigInt)

	// Get the epoch snapshot slot for the previous epoch
	// This is equivalent to "prevSnapshot" in the Solidity code
	prevEpochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(prevEpochBigInt)
	gasUsed += slotGasUsed

	// Get the end time of the previous epoch
	prevEndTimeSlot := GetBigInt().Add(prevEpochSnapshotSlot, big.NewInt(endTimeOffset))
	prevEndTime := epochCache.GetState(ContractAddress, common.BigToHash(prevEndTimeSlot))
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
	epochCache.SetState(ContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)), common.BigToHash(currentEpochBigInt))
	gasUsed += SstoreGasCost

	// Update epoch snapshot end time (snapshot.endTime = _now() in Solidity)
	endTimeOffsetBig := GetBigInt().SetInt64(endTimeOffset)
	endTimeSlot := GetBigInt().Add(currentEpochSnapshotSlot, endTimeOffsetBig)
	epochCache.SetState(ContractAddress, common.BigToHash(endTimeSlot), common.BigToHash(evm.Context.Time))
	gasUsed += SstoreGasCost
	PutBigInt(endTimeOffsetBig)
	PutBigInt(endTimeSlot)

	// Get the base reward per second from the constants manager
	baseRewardPerSecond := getConstantsManagerVariable("baseRewardPerSecond")

	// Update epoch snapshot base reward per second (snapshot.baseRewardPerSecond = c.baseRewardPerSecond() in Solidity)
	baseRewardPerSecondOffsetBig := GetBigInt().SetInt64(baseRewardPerSecondOffset)
	baseRewardPerSecondSlot := GetBigInt().Add(currentEpochSnapshotSlot, baseRewardPerSecondOffsetBig)
	epochCache.SetState(ContractAddress, common.BigToHash(baseRewardPerSecondSlot), common.BigToHash(baseRewardPerSecond))
	gasUsed += SstoreGasCost
	PutBigInt(baseRewardPerSecondOffsetBig)
	PutBigInt(baseRewardPerSecondSlot)

	// Get the total supply
	totalSupplySlotBig := GetBigInt().SetInt64(totalSupplySlot)
	totalSupply := epochCache.GetState(ContractAddress, common.BigToHash(totalSupplySlotBig))
	gasUsed += SloadGasCost
	totalSupplyBigInt := GetBigInt().SetBytes(totalSupply.Bytes())

	// Update epoch snapshot total supply (snapshot.totalSupply = totalSupply in Solidity)
	totalSupplyOffsetBig := GetBigInt().SetInt64(totalSupplyOffset)
	totalSupplySnapshotSlot := GetBigInt().Add(currentEpochSnapshotSlot, totalSupplyOffsetBig)
	epochCache.SetState(ContractAddress, common.BigToHash(totalSupplySnapshotSlot), common.BigToHash(totalSupplyBigInt))
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

	// Initialize epoch cache
	var epochCache *EpochStateCache
	if evm.EpochCache == nil {
		evm.EpochCache = NewEpochStateCache(evm.SfcStateDB)
	}
	epochCache = evm.EpochCache.(*EpochStateCache)

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
	totalStakeHash := epochCache.GetState(ContractAddress, common.BigToHash(totalStakeSlot))
	gasUsed += SloadGasCost
	totalStake := new(big.Int).SetBytes(totalStakeHash.Bytes())

	// Fill data for the next snapshot
	// This corresponds to the loop in the Solidity implementation that sets receivedStake and adds to totalStake
	for _, validatorID := range nextValidatorIDs {
		// Get the validator's received stake from getValidator[validatorID].receivedStake
		validatorReceivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(validatorID)
		gasUsed += slotGasUsed

		receivedStake := epochCache.GetState(ContractAddress, common.BigToHash(validatorReceivedStakeSlot))
		gasUsed += SloadGasCost
		receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

		// Set the received stake for this validator in the epoch snapshot (snapshot.receivedStake[validatorID] = receivedStake)
		validatorReceivedStakeEpochSlot, slotGasUsed := getEpochValidatorReceivedStakeSlot(currentEpochBigInt, validatorID)
		gasUsed += slotGasUsed

		epochCache.SetState(ContractAddress, common.BigToHash(validatorReceivedStakeEpochSlot), common.BigToHash(receivedStakeBigInt))
		gasUsed += SstoreGasCost

		// Add to total stake (snapshot.totalStake = snapshot.totalStake.add(receivedStake))
		totalStake = new(big.Int).Add(totalStake, receivedStakeBigInt)
	}

	// Set the validator IDs for the epoch snapshot
	// For a dynamic array in a struct, we first set the length at the slot
	validatorIDsSlot := new(big.Int).Add(epochSnapshotSlot, big.NewInt(validatorIDsOffset))
	validatorIDsLength := big.NewInt(int64(len(nextValidatorIDs)))
	epochCache.SetState(ContractAddress, common.BigToHash(validatorIDsSlot), common.BigToHash(validatorIDsLength))
	gasUsed += SstoreGasCost

	// Calculate the base slot for the array elements
	// The array elements start at keccak256(slot)
	validatorIDsBaseSlotBytes := CachedKeccak256Hash(common.BigToHash(validatorIDsSlot).Bytes()).Bytes()
	gasUsed += HashGasCost
	validatorIDsBaseSlot := new(big.Int).SetBytes(validatorIDsBaseSlotBytes)

	// Store each validator ID in the validatorIDs array of the epoch snapshot
	// This corresponds to `snapshot.validatorIDs = nextValidatorIDs` in the Solidity implementation
	for i, validatorID := range nextValidatorIDs {
		// Calculate the slot for this array element: baseSlot + i
		elementSlot := new(big.Int).Add(validatorIDsBaseSlot, big.NewInt(int64(i)))

		// Store the validator ID
		epochCache.SetState(ContractAddress, common.BigToHash(elementSlot), common.BigToHash(validatorID))
		gasUsed += SstoreGasCost
	}

	// Set the updated total stake for the epoch snapshot
	// We've already calculated the totalStakeSlot above and updated totalStake in the loop
	epochCache.SetState(ContractAddress, common.BigToHash(totalStakeSlot), common.BigToHash(totalStake))
	gasUsed += SstoreGasCost

	// Update the minimum gas price in the node
	// This corresponds to `node.updateMinGasPrice(minGasPrice)` in the Solidity implementation
	minGasPrice := epochCache.GetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)))
	gasUsed += SloadGasCost
	minGasPriceBigInt := new(big.Int).SetBytes(minGasPrice.Bytes())

	// Get the node driver auth address to call updateMinGasPrice
	nodeDriverAuth := epochCache.GetState(ContractAddress, common.BigToHash(big.NewInt(nodeDriverAuthSlot)))
	gasUsed += SloadGasCost
	nodeDriverAuthAddr := common.BytesToAddress(nodeDriverAuth.Bytes())

	// Pack the function call data
	data, err := NodeDriverAuthAbi.Pack("updateMinGasPrice", minGasPriceBigInt)
	if err != nil {
		log.Error("SFC: Error packing updateMinGasPrice call data", "err", err)
		return nil, gasUsed, vm.ErrExecutionReverted
	}

	// Call the node driver
	result, _, err := evm.CallSFC(
		vm.AccountRef(ContractAddress), // Caller
		nodeDriverAuthAddr,             // Target address
		data,                           // Call data (empty)
		50000,                          // Gas limit for a simple transfer
		big.NewInt(0),                  // Value to transfer
	)
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

	// Initialize epoch cache
	var epochCache *EpochStateCache
	if evm.EpochCache == nil {
		evm.EpochCache = NewEpochStateCache(evm.SfcStateDB)
	}
	epochCache = evm.EpochCache.(*EpochStateCache)

	// Get the offline penalty thresholds from the constants manager
	offlinePenaltyThresholdBlocksNum := getConstantsManagerVariable("offlinePenaltyThresholdBlocksNum")
	offlinePenaltyThresholdTime := getConstantsManagerVariable("offlinePenaltyThresholdTime")

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
		epochCache.SetState(ContractAddress, common.BigToHash(offlineTimeSlot), common.BigToHash(offlineTimes[i]))
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
		epochCache.SetState(ContractAddress, common.BigToHash(offlineBlocksSlot), common.BigToHash(offlineBlocks[i]))
		gasUsed += SstoreGasCost
	}

	return gasUsed, nil
}

// _sealEpoch_rewards is an optimized internal function to seal rewards in an epoch
func _sealEpoch_rewards(evm *vm.EVM, epochDuration *big.Int, currentEpoch *big.Int, prevEpoch *big.Int,
	validatorIDs []*big.Int, uptimes []*big.Int, accumulatedOriginatedTxsFee []*big.Int) (uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Initialize epoch cache
	var epochCache *EpochStateCache
	if evm.EpochCache == nil {
		evm.EpochCache = NewEpochStateCache(evm.SfcStateDB)
	}
	epochCache = evm.EpochCache.(*EpochStateCache)

	// Cache frequently used constants
	constants := struct {
		baseRewardPerSecond *big.Int
		validatorCommission *big.Int
		burntFeeShare       *big.Int
		treasuryFeeShare    *big.Int
		decimalUnit         *big.Int
	}{
		baseRewardPerSecond: getConstantsManagerVariable("baseRewardPerSecond"),
		validatorCommission: getConstantsManagerVariable("validatorCommission"),
		burntFeeShare:       getConstantsManagerVariable("burntFeeShare"),
		treasuryFeeShare:    getConstantsManagerVariable("treasuryFeeShare"),
		decimalUnit:         getDecimalUnit(),
	}

	// Pre-calculate epoch snapshot slots
	currentEpochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(currentEpoch)
	gasUsed += slotGasUsed
	prevEpochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(prevEpoch)
	gasUsed += slotGasUsed

	// Pre-fetch all validator data sequentially
	type ValidatorData struct {
		receivedStake  *big.Int
		selfStake      *big.Int
		lockedStake    *big.Int
		lockupDuration *big.Int
		authAddress    common.Address
	}

	validatorDataCache := make(map[*big.Int]*ValidatorData)
	for _, validatorID := range validatorIDs {
		// Get validator's received stake
		receivedStakeSlot, _ := getEpochValidatorReceivedStakeSlot(currentEpoch, validatorID)
		receivedStake := epochCache.GetState(ContractAddress, common.BigToHash(receivedStakeSlot))
		receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

		// Get validator auth address
		validatorAuthSlot, _ := getValidatorAuthSlot(validatorID)
		validatorAuth := epochCache.GetState(ContractAddress, common.BigToHash(validatorAuthSlot))
		validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

		// Get validator's self-stake
		selfStakeSlot, _ := getStakeSlot(validatorAuthAddr, validatorID)
		selfStake := epochCache.GetState(ContractAddress, common.BigToHash(selfStakeSlot))
		selfStakeBigInt := new(big.Int).SetBytes(selfStake.Bytes())

		// Get locked stake
		lockedStakeSlot, _ := getLockedStakeSlot(validatorAuthAddr, validatorID)
		lockedStake := epochCache.GetState(ContractAddress, common.BigToHash(lockedStakeSlot))
		lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())

		// Get lockup duration
		lockupDurationSlot, _ := getLockupDurationSlot(validatorAuthAddr, validatorID)
		lockupDuration := epochCache.GetState(ContractAddress, common.BigToHash(lockupDurationSlot))
		lockupDurationBigInt := new(big.Int).SetBytes(lockupDuration.Bytes())

		validatorDataCache[validatorID] = &ValidatorData{
			receivedStake:  receivedStakeBigInt,
			selfStake:      selfStakeBigInt,
			lockedStake:    lockedStakeBigInt,
			lockupDuration: lockupDurationBigInt,
			authAddress:    validatorAuthAddr,
		}
	}
	gasUsed += uint64(len(validatorIDs)) * 5 * SloadGasCost // 5 SLOADs per validator

	// Calculate rewards sequentially
	type RewardData struct {
		validatorID           *big.Int
		baseRewardWeight      *big.Int
		txRewardWeight        *big.Int
		originatedTxsFee      *big.Int
		prevAccumulatedTxsFee *big.Int
		index                 int
	}

	rewardDataList := make([]RewardData, 0, len(validatorIDs))
	var totalBaseRewardWeight, totalTxRewardWeight, epochFee big.Int

	for i, validatorID := range validatorIDs {
		// Get previous accumulated originated txs fee
		mappingSlot := new(big.Int).Add(prevEpochSnapshotSlot, big.NewInt(accumulatedOriginatedTxsFeeOffset))
		outerHashInput := CreateNestedHashInput(validatorID, mappingSlot.Bytes())
		outerHash := CachedKeccak256(outerHashInput)
		prevAccumulatedTxsFeeSlot := new(big.Int).SetBytes(outerHash)
		prevAccumulatedTxsFee := epochCache.GetState(ContractAddress, common.BigToHash(prevAccumulatedTxsFeeSlot))
		prevAccumulatedTxsFeeBigInt := new(big.Int).SetBytes(prevAccumulatedTxsFee.Bytes())

		// Calculate originated txs fee for this epoch
		originatedTxsFee := big.NewInt(0)
		if accumulatedOriginatedTxsFee[i].Cmp(prevAccumulatedTxsFeeBigInt) > 0 {
			originatedTxsFee = new(big.Int).Sub(accumulatedOriginatedTxsFee[i], prevAccumulatedTxsFeeBigInt)
		}

		// Calculate tx reward weight
		txRewardWeight := new(big.Int).Mul(originatedTxsFee, uptimes[i])
		txRewardWeight = new(big.Int).Div(txRewardWeight, epochDuration)

		// Calculate base reward weight
		validatorData := validatorDataCache[validatorID]
		term1 := new(big.Int).Mul(validatorData.receivedStake, uptimes[i])
		term1 = new(big.Int).Div(term1, epochDuration)
		term2 := new(big.Int).Mul(term1, uptimes[i])
		baseRewardWeight := new(big.Int).Div(term2, epochDuration)

		rewardData := RewardData{
			validatorID:           validatorID,
			baseRewardWeight:      baseRewardWeight,
			txRewardWeight:        txRewardWeight,
			originatedTxsFee:      originatedTxsFee,
			prevAccumulatedTxsFee: prevAccumulatedTxsFeeBigInt,
			index:                 i,
		}

		rewardDataList = append(rewardDataList, rewardData)
		totalBaseRewardWeight.Add(&totalBaseRewardWeight, baseRewardWeight)
		totalTxRewardWeight.Add(&totalTxRewardWeight, txRewardWeight)
		epochFee.Add(&epochFee, originatedTxsFee)
	}

	// Create reward data map after all calculations are complete
	rewardDataMap := make(map[*big.Int]RewardData)
	for _, data := range rewardDataList {
		rewardDataMap[data.validatorID] = data
	}

	// Prepare batch updates
	updates := make([]struct {
		slot  common.Hash
		value common.Hash
	}, 0, len(validatorIDs)*10) // Pre-allocate for efficiency

	// Process rewards for each validator
	for _, validatorID := range validatorIDs {
		rewardData := rewardDataMap[validatorID]
		validatorData := validatorDataCache[validatorID]

		// Calculate raw base reward
		rawBaseReward := big.NewInt(0)
		if rewardData.baseRewardWeight.Sign() > 0 {
			totalReward := new(big.Int).Mul(epochDuration, constants.baseRewardPerSecond)
			rawBaseReward = new(big.Int).Mul(totalReward, rewardData.baseRewardWeight)
			rawBaseReward = new(big.Int).Div(rawBaseReward, &totalBaseRewardWeight)
		}

		// Calculate raw tx reward
		rawTxReward := big.NewInt(0)
		if rewardData.txRewardWeight.Sign() > 0 {
			txReward := new(big.Int).Mul(&epochFee, rewardData.txRewardWeight)
			txReward = new(big.Int).Div(txReward, &totalTxRewardWeight)

			shareToSubtract := new(big.Int).Add(constants.burntFeeShare, constants.treasuryFeeShare)
			shareToKeep := new(big.Int).Sub(constants.decimalUnit, shareToSubtract)

			rawTxReward = new(big.Int).Mul(txReward, shareToKeep)
			rawTxReward = new(big.Int).Div(rawTxReward, constants.decimalUnit)
		}

		// Calculate total raw reward
		rawReward := new(big.Int).Add(rawBaseReward, rawTxReward)

		// Calculate validator's commission
		commissionRewardFull := new(big.Int).Mul(rawReward, constants.validatorCommission)
		commissionRewardFull = new(big.Int).Div(commissionRewardFull, constants.decimalUnit)

		// Process commission reward if self-stake is not zero
		if validatorData.selfStake.Sign() != 0 {
			// Calculate locked and unlocked commission rewards
			lCommissionRewardFull := new(big.Int).Mul(commissionRewardFull, validatorData.lockedStake)
			lCommissionRewardFull = new(big.Int).Div(lCommissionRewardFull, validatorData.selfStake)

			// Scale lockup rewards
			reward, scaleGasUsed, err := _scaleLockupReward(evm, lCommissionRewardFull, validatorData.lockupDuration)
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

			// Get rewards stash slots
			rewardsStashSlot, _ := getRewardsStashSlot(validatorData.authAddress, validatorID)
			lockupBaseRewardSlot := new(big.Int).Add(rewardsStashSlot, big.NewInt(1))
			unlockedRewardSlot := new(big.Int).Add(rewardsStashSlot, big.NewInt(2))

			// Get current rewards stash
			rewardsStash := epochCache.GetState(ContractAddress, common.BigToHash(rewardsStashSlot))
			lockupBaseReward := epochCache.GetState(ContractAddress, common.BigToHash(lockupBaseRewardSlot))
			unlockedReward := epochCache.GetState(ContractAddress, common.BigToHash(unlockedRewardSlot))

			// Convert to Rewards struct
			currentRewardsStash := Rewards{
				LockupExtraReward: new(big.Int).SetBytes(rewardsStash.Bytes()),
				LockupBaseReward:  new(big.Int).SetBytes(lockupBaseReward.Bytes()),
				UnlockedReward:    new(big.Int).SetBytes(unlockedReward.Bytes()),
			}

			// Calculate new rewards
			newRewardsStash := sumRewards(currentRewardsStash, reward, uReward)

			// Add rewards stash updates
			updates = append(updates,
				struct {
					slot  common.Hash
					value common.Hash
				}{common.BigToHash(rewardsStashSlot), common.BigToHash(newRewardsStash.LockupExtraReward)},
				struct {
					slot  common.Hash
					value common.Hash
				}{common.BigToHash(lockupBaseRewardSlot), common.BigToHash(newRewardsStash.LockupBaseReward)},
				struct {
					slot  common.Hash
					value common.Hash
				}{common.BigToHash(unlockedRewardSlot), common.BigToHash(newRewardsStash.UnlockedReward)},
			)

			// Get stashed lockup rewards slots
			stashedLockupRewardsSlot, _ := getStashedLockupRewardsSlot(validatorData.authAddress, validatorID)
			stashedLockupBaseRewardSlot := new(big.Int).Add(stashedLockupRewardsSlot, big.NewInt(1))
			stashedUnlockedRewardSlot := new(big.Int).Add(stashedLockupRewardsSlot, big.NewInt(2))

			// Get current stashed lockup rewards
			stashedLockupRewards := epochCache.GetState(ContractAddress, common.BigToHash(stashedLockupRewardsSlot))
			stashedLockupBaseReward := epochCache.GetState(ContractAddress, common.BigToHash(stashedLockupBaseRewardSlot))
			stashedUnlockedReward := epochCache.GetState(ContractAddress, common.BigToHash(stashedUnlockedRewardSlot))

			// Convert to Rewards struct
			currentStashedLockupRewards := Rewards{
				LockupExtraReward: new(big.Int).SetBytes(stashedLockupRewards.Bytes()),
				LockupBaseReward:  new(big.Int).SetBytes(stashedLockupBaseReward.Bytes()),
				UnlockedReward:    new(big.Int).SetBytes(stashedUnlockedReward.Bytes()),
			}

			// Calculate new stashed rewards
			newStashedLockupRewards := sumRewards(currentStashedLockupRewards, reward, uReward)

			// Add stashed rewards updates
			updates = append(updates,
				struct {
					slot  common.Hash
					value common.Hash
				}{common.BigToHash(stashedLockupRewardsSlot), common.BigToHash(newStashedLockupRewards.LockupExtraReward)},
				struct {
					slot  common.Hash
					value common.Hash
				}{common.BigToHash(stashedLockupBaseRewardSlot), common.BigToHash(newStashedLockupRewards.LockupBaseReward)},
				struct {
					slot  common.Hash
					value common.Hash
				}{common.BigToHash(stashedUnlockedRewardSlot), common.BigToHash(newStashedLockupRewards.UnlockedReward)},
			)
		}

		// Calculate delegators' reward
		delegatorsReward := new(big.Int).Sub(rawReward, commissionRewardFull)

		// Calculate reward per token
		rewardPerToken := big.NewInt(0)
		if validatorData.receivedStake.Sign() != 0 {
			rewardPerToken = new(big.Int).Mul(delegatorsReward, constants.decimalUnit)
			rewardPerToken = new(big.Int).Div(rewardPerToken, validatorData.receivedStake)
		}

		// Update accumulated reward per token
		mappingSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(accumulatedRewardPerTokenOffset))
		outerHashInput := CreateValidatorMappingHashInput(validatorID, mappingSlot)
		accumulatedRewardPerTokenSlotHash := CachedKeccak256Hash(outerHashInput)
		accumulatedRewardPerTokenSlot := new(big.Int).SetBytes(accumulatedRewardPerTokenSlotHash.Bytes())

		prevMappingSlot := new(big.Int).Add(prevEpochSnapshotSlot, big.NewInt(accumulatedRewardPerTokenOffset))
		outerHashInput = CreateValidatorMappingHashInput(validatorID, prevMappingSlot)
		prevAccumulatedRewardPerTokenSlotHash := CachedKeccak256Hash(outerHashInput)
		prevAccumulatedRewardPerTokenSlot := new(big.Int).SetBytes(prevAccumulatedRewardPerTokenSlotHash.Bytes())

		prevAccumulatedRewardPerToken := epochCache.GetState(ContractAddress, common.BigToHash(prevAccumulatedRewardPerTokenSlot))
		prevAccumulatedRewardPerTokenBigInt := new(big.Int).SetBytes(prevAccumulatedRewardPerToken.Bytes())

		newAccumulatedRewardPerToken := new(big.Int).Add(prevAccumulatedRewardPerTokenBigInt, rewardPerToken)
		updates = append(updates, struct {
			slot  common.Hash
			value common.Hash
		}{
			common.BigToHash(accumulatedRewardPerTokenSlot),
			common.BigToHash(newAccumulatedRewardPerToken),
		})

		// Update accumulated originated txs fee
		originatedTxsFeeSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(accumulatedOriginatedTxsFeeOffset))
		outerHashInput = CreateValidatorMappingHashInput(validatorID, originatedTxsFeeSlot)
		accumulatedOriginatedTxsFeeSlotHash := CachedKeccak256Hash(outerHashInput)
		accumulatedOriginatedTxsFeeSlot := new(big.Int).SetBytes(accumulatedOriginatedTxsFeeSlotHash.Bytes())

		updates = append(updates, struct {
			slot  common.Hash
			value common.Hash
		}{
			common.BigToHash(accumulatedOriginatedTxsFeeSlot),
			common.BigToHash(accumulatedOriginatedTxsFee[rewardData.index]),
		})

		// Update accumulated uptime
		uptimeMappingSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(accumulatedUptimeOffset))
		outerHashInput = CreateValidatorMappingHashInput(validatorID, uptimeMappingSlot)
		accumulatedUptimeSlotHash := CachedKeccak256Hash(outerHashInput)
		accumulatedUptimeSlot := new(big.Int).SetBytes(accumulatedUptimeSlotHash.Bytes())

		prevUptimeMappingSlot := new(big.Int).Add(prevEpochSnapshotSlot, big.NewInt(accumulatedUptimeOffset))
		outerHashInput = CreateValidatorMappingHashInput(validatorID, prevUptimeMappingSlot)
		prevAccumulatedUptimeSlotHash := CachedKeccak256Hash(outerHashInput)
		prevAccumulatedUptimeSlot := new(big.Int).SetBytes(prevAccumulatedUptimeSlotHash.Bytes())

		prevAccumulatedUptime := epochCache.GetState(ContractAddress, common.BigToHash(prevAccumulatedUptimeSlot))
		prevAccumulatedUptimeBigInt := new(big.Int).SetBytes(prevAccumulatedUptime.Bytes())

		newAccumulatedUptime := new(big.Int).Add(prevAccumulatedUptimeBigInt, uptimes[rewardData.index])
		updates = append(updates, struct {
			slot  common.Hash
			value common.Hash
		}{
			common.BigToHash(accumulatedUptimeSlot),
			common.BigToHash(newAccumulatedUptime),
		})
	}

	// Calculate slots for epoch summary data
	epochFeeSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(epochFeeOffset))
	totalBaseRewardSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(totalBaseRewardOffset))
	totalTxRewardSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(totalTxRewardOffset))

	// Get total supply
	totalSupply := epochCache.GetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)))
	totalSupplyBigInt := new(big.Int).SetBytes(totalSupply.Bytes())

	// Subtract epoch fee from total supply
	if totalSupplyBigInt.Cmp(&epochFee) > 0 {
		totalSupplyBigInt = new(big.Int).Sub(totalSupplyBigInt, &epochFee)
	} else {
		totalSupplyBigInt = big.NewInt(0)
	}

	// Add epoch summary updates
	updates = append(updates,
		struct {
			slot  common.Hash
			value common.Hash
		}{common.BigToHash(epochFeeSlot), common.BigToHash(&epochFee)},
		struct {
			slot  common.Hash
			value common.Hash
		}{common.BigToHash(totalBaseRewardSlot), common.BigToHash(&totalBaseRewardWeight)},
		struct {
			slot  common.Hash
			value common.Hash
		}{common.BigToHash(totalTxRewardSlot), common.BigToHash(&totalTxRewardWeight)},
		struct {
			slot  common.Hash
			value common.Hash
		}{common.BigToHash(big.NewInt(totalSupplySlot)), common.BigToHash(totalSupplyBigInt)},
	)

	// Batch update state
	epochCache.BatchSetState(ContractAddress, updates)
	gasUsed += uint64(len(updates)) * SstoreGasCost

	// Transfer fees to treasury if address is set
	treasuryAddress := epochCache.GetState(ContractAddress, common.BigToHash(big.NewInt(treasuryAddressSlot)))
	treasuryAddressBytes := treasuryAddress.Bytes()

	// Check if treasury address is not zero
	emptyAddr := common.Address{}
	treasuryAddr := common.BytesToAddress(treasuryAddressBytes)
	if treasuryAddr.Cmp(emptyAddr) != 0 {
		// Calculate fee share
		feeShare := new(big.Int).Mul(&epochFee, constants.treasuryFeeShare)
		feeShare = new(big.Int).Div(feeShare, constants.decimalUnit)

		// First mint native token to the contract itself
		mintGasUsed, err := _mintNativeToken(evm, ContractAddress, feeShare)
		gasUsed += mintGasUsed
		if err != nil {
			return gasUsed, err
		}

		// Then make a call to transfer the tokens to the treasury address
		callData := []byte{} // Empty call data
		_, _, err = evm.CallSFC(
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

// _sealEpoch_minGasPrice is an internal function to update the minimum gas price
func _sealEpoch_minGasPrice(
	evm *vm.EVM,
	epochDuration *big.Int,
	epochGas *big.Int,
) (uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Initialize epoch cache
	var epochCache *EpochStateCache
	if evm.EpochCache == nil {
		evm.EpochCache = NewEpochStateCache(evm.SfcStateDB)
	}
	epochCache = evm.EpochCache.(*EpochStateCache)

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
	minGasPrice := epochCache.GetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)))
	gasUsed += SloadGasCost
	minGasPriceBigInt := new(big.Int).SetBytes(minGasPrice.Bytes())

	// Apply the ratio
	newMinGasPrice := new(big.Int).Mul(minGasPriceBigInt, gasPriceDeltaRatio)
	newMinGasPrice.Div(newMinGasPrice, decimalUnitBigInt)

	// Limit the max/min possible minGasPrice using the trimMinGasPrice helper function
	newMinGasPrice = trimMinGasPrice(newMinGasPrice)

	// Apply new minGasPrice
	epochCache.SetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)), common.BigToHash(newMinGasPrice))
	gasUsed += SstoreGasCost

	return gasUsed, nil
}
