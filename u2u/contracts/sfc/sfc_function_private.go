package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/params"
)

// Handler functions for SFC contract internal and private functions

// _delegate is an internal function to delegate calls to an implementation address
// This is a Go implementation of the Solidity function:
//
//	function _delegate(address implementation) internal {
//	    assembly {
//	        calldatacopy(0, 0, calldatasize())
//	        let result := delegatecall(gas(), implementation, 0, calldatasize(), 0, 0)
//	        returndatacopy(0, 0, returndatasize())
//	        switch result
//	        case 0 { revert(0, returndatasize()) }
//	        default { return(0, returndatasize()) }
//	    }
//	}
func handle_delegate(evm *vm.EVM, args []interface{}, input []byte) ([]byte, uint64, error) {
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the implementation address from args
	implementation, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// The original calldata is available in the input parameter
	// For the _delegate function, we need to skip the method selector (first 4 bytes)
	// This simulates the Solidity assembly code: calldatacopy(0, 0, calldatasize())
	originalInput := input

	// If the input starts with the _delegate method selector, skip it
	if len(originalInput) >= 4 {
		// Check if the first 4 bytes match the _delegate method selector
		if method, err := SfcAbi.MethodById(originalInput[:4]); err == nil && method.Name == "_delegate" {
			// Skip the method selector and the ABI-encoded implementation address
			// The implementation address is already extracted from args, so we don't need it in the input
			originalInput = []byte{}
		}
	}

	// Create a contract reference for the caller
	callerRef := vm.AccountRef(evm.TxContext.Origin)

	// Make the delegate call
	// This simulates the Solidity assembly code: let result := delegatecall(gas, implementation, 0, calldatasize, 0, 0)
	// Use a fixed gas amount for now
	gas := uint64(1000000)
	ret, leftOverGas, err := evm.DelegateCall(callerRef, implementation, originalInput, gas)

	// Calculate gas used
	gasUsed := gas - leftOverGas

	// Handle errors similar to the Solidity assembly code:
	// switch result
	// case 0 { revert(0, returndatasize) }
	// default { return (0, returndatasize) }
	if err != nil {
		return nil, gasUsed, err
	}

	return ret, gasUsed, nil
}

// _sealEpoch_offline is an internal function to seal offline validators in an epoch
func _sealEpoch_offline(evm *vm.EVM, validatorIDs []*big.Int, offlineTimes []*big.Int, offlineBlocks []*big.Int, currentEpoch *big.Int) (uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Get the epoch snapshot slot (not used directly in this function but needed for gas calculation)
	_, slotGasUsed := getEpochSnapshotSlot(currentEpoch)
	gasUsed += slotGasUsed

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

		// Store offline time in the epoch snapshot
		offlineTimeSlot, slotGasUsed := getEpochValidatorOfflineTimeSlot(currentEpoch, validatorID)
		gasUsed += slotGasUsed

		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(offlineTimeSlot)), common.BigToHash(offlineTimes[i]))
		gasUsed += SstoreGasCost

		// Store offline blocks in the epoch snapshot
		offlineBlocksSlot, slotGasUsed := getEpochValidatorOfflineBlocksSlot(currentEpoch, validatorID)
		gasUsed += slotGasUsed

		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(offlineBlocksSlot)), common.BigToHash(offlineBlocks[i]))
		gasUsed += SstoreGasCost
	}

	return gasUsed, nil
}

// _sealEpoch_rewards is an internal function to seal rewards in an epoch
func _sealEpoch_rewards(evm *vm.EVM, epochDuration *big.Int, currentEpoch *big.Int, prevEpoch *big.Int,
	validatorIDs []*big.Int, uptimes []*big.Int, accumulatedOriginatedTxsFee []*big.Int) (uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Get the epoch snapshot slots
	currentEpochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(currentEpoch)
	gasUsed += slotGasUsed

	// Get previous epoch snapshot slot (used later for calculations)
	_, slotGasUsed = getEpochSnapshotSlot(prevEpoch)
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
		prevAccumulatedTxsFeeSlot, slotGasUsed := getEpochValidatorAccumulatedOriginatedTxsFeeSlot(prevEpoch, validatorID)
		gasUsed += slotGasUsed

		prevAccumulatedTxsFee := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(prevAccumulatedTxsFeeSlot)))
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

		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(receivedStakeSlot)))
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

	// Get the base reward per second from the constants manager
	baseRewardPerSecond, cmGasUsed, err := callConstantManagerMethod(evm, "baseRewardPerSecond")
	gasUsed += cmGasUsed
	if err != nil || len(baseRewardPerSecond) == 0 {
		return gasUsed, err
	}
	baseRewardPerSecondBigInt, ok := baseRewardPerSecond[0].(*big.Int)
	if !ok {
		return gasUsed, vm.ErrExecutionReverted
	}

	// Get the validator commission from the constants manager
	validatorCommission, cmGasUsed, err := callConstantManagerMethod(evm, "validatorCommission")
	gasUsed += cmGasUsed
	if err != nil || len(validatorCommission) == 0 {
		return gasUsed, err
	}
	validatorCommissionBigInt, ok := validatorCommission[0].(*big.Int)
	if !ok {
		return gasUsed, vm.ErrExecutionReverted
	}

	// Get the burnt fee share from the constants manager
	burntFeeShare, cmGasUsed, err := callConstantManagerMethod(evm, "burntFeeShare")
	gasUsed += cmGasUsed
	if err != nil || len(burntFeeShare) == 0 {
		return gasUsed, err
	}
	burntFeeShareBigInt, ok := burntFeeShare[0].(*big.Int)
	if !ok {
		return gasUsed, vm.ErrExecutionReverted
	}

	// Get the treasury fee share from the constants manager
	treasuryFeeShare, cmGasUsed, err := callConstantManagerMethod(evm, "treasuryFeeShare")
	gasUsed += cmGasUsed
	if err != nil || len(treasuryFeeShare) == 0 {
		return gasUsed, err
	}
	treasuryFeeShareBigInt, ok := treasuryFeeShare[0].(*big.Int)
	if !ok {
		return gasUsed, vm.ErrExecutionReverted
	}

	// Get the decimal unit (1e18) using the helper function
	decimalUnitBigInt := getDecimalUnit()
	// Calculate rewards for each validator
	for i, validatorID := range validatorIDs {
		// Calculate raw base reward
		rawBaseReward := big.NewInt(0)
		if baseRewardWeights[i].Cmp(big.NewInt(0)) > 0 {
			totalReward := new(big.Int).Mul(epochDuration, baseRewardPerSecondBigInt)
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
			shareToSubtract := new(big.Int).Add(burntFeeShareBigInt, treasuryFeeShareBigInt)
			shareToKeep := new(big.Int).Sub(decimalUnitBigInt, shareToSubtract)

			rawTxReward = new(big.Int).Mul(txReward, shareToKeep)
			rawTxReward = new(big.Int).Div(rawTxReward, decimalUnitBigInt)
		}

		// Calculate total raw reward
		rawReward := new(big.Int).Add(rawBaseReward, rawTxReward)

		// Get validator auth address
		validatorAuthSlot, slotGasUsed := getValidatorAuthSlot(validatorID)
		gasUsed += slotGasUsed

		validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorAuthSlot)))
		gasUsed += SloadGasCost
		validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

		// Calculate validator's commission
		commissionRewardFull := new(big.Int).Mul(rawReward, validatorCommissionBigInt)
		commissionRewardFull = new(big.Int).Div(commissionRewardFull, decimalUnitBigInt)

		// Get validator's self-stake
		selfStakeSlot, slotGasUsed := getStakeSlot(validatorAuthAddr, validatorID)
		gasUsed += slotGasUsed

		selfStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(selfStakeSlot)))
		gasUsed += SloadGasCost
		selfStakeBigInt := new(big.Int).SetBytes(selfStake.Bytes())

		// Process commission reward if self-stake is not zero
		if selfStakeBigInt.Cmp(big.NewInt(0)) != 0 {
			// Get locked stake
			lockedStakeSlot, slotGasUsed := getLockedStakeSlot(validatorAuthAddr, validatorID)
			gasUsed += slotGasUsed

			lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lockedStakeSlot)))
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

			lockupDuration := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lockupDurationSlot)))
			gasUsed += SloadGasCost
			// Unused in current implementation, but kept for future use
			_ = new(big.Int).SetBytes(lockupDuration.Bytes())

			// Scale lockup rewards
			// TODO: Implement _scaleLockupReward
			// For now, we'll just stash the rewards directly

			// Update rewards stash
			rewardsStashSlot, slotGasUsed := getRewardsStashSlot(validatorAuthAddr, validatorID)
			gasUsed += slotGasUsed

			// TODO: Implement proper rewards stashing with lockup scaling
			// For now, just add the commission reward to the stash
			rewardsStash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(rewardsStashSlot)))
			gasUsed += SloadGasCost
			rewardsStashBigInt := new(big.Int).SetBytes(rewardsStash.Bytes())

			newRewardsStash := new(big.Int).Add(rewardsStashBigInt, commissionRewardFull)
			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(rewardsStashSlot)), common.BigToHash(newRewardsStash))
			gasUsed += SstoreGasCost

			// Update stashed lockup rewards
			stashedLockupRewardsSlot, slotGasUsed := getStashedLockupRewardsSlot(validatorAuthAddr, validatorID)
			gasUsed += slotGasUsed

			stashedLockupRewards := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stashedLockupRewardsSlot)))
			gasUsed += SloadGasCost
			stashedLockupRewardsBigInt := new(big.Int).SetBytes(stashedLockupRewards.Bytes())

			newStashedLockupRewards := new(big.Int).Add(stashedLockupRewardsBigInt, commissionRewardFull)
			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(stashedLockupRewardsSlot)), common.BigToHash(newStashedLockupRewards))
			gasUsed += SstoreGasCost
		}

		// Calculate delegators' reward
		delegatorsReward := new(big.Int).Sub(rawReward, commissionRewardFull)

		// Get validator's received stake
		receivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(validatorID)
		gasUsed += slotGasUsed

		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(receivedStakeSlot)))
		gasUsed += SloadGasCost
		receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

		// Calculate reward per token
		rewardPerToken := big.NewInt(0)
		if receivedStakeBigInt.Cmp(big.NewInt(0)) != 0 {
			rewardPerToken = new(big.Int).Mul(delegatorsReward, decimalUnitBigInt)
			rewardPerToken = new(big.Int).Div(rewardPerToken, receivedStakeBigInt)
		}

		// Update accumulated reward per token
		accumulatedRewardPerTokenSlot, slotGasUsed := getEpochValidatorAccumulatedRewardPerTokenSlot(currentEpoch, validatorID)
		gasUsed += slotGasUsed

		prevAccumulatedRewardPerTokenSlot, slotGasUsed := getEpochValidatorAccumulatedRewardPerTokenSlot(prevEpoch, validatorID)
		gasUsed += slotGasUsed

		prevAccumulatedRewardPerToken := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(prevAccumulatedRewardPerTokenSlot)))
		gasUsed += SloadGasCost
		prevAccumulatedRewardPerTokenBigInt := new(big.Int).SetBytes(prevAccumulatedRewardPerToken.Bytes())

		newAccumulatedRewardPerToken := new(big.Int).Add(prevAccumulatedRewardPerTokenBigInt, rewardPerToken)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(accumulatedRewardPerTokenSlot)), common.BigToHash(newAccumulatedRewardPerToken))
		gasUsed += SstoreGasCost

		// Update accumulated originated txs fee
		accumulatedOriginatedTxsFeeSlot, slotGasUsed := getEpochValidatorAccumulatedOriginatedTxsFeeSlot(currentEpoch, validatorID)
		gasUsed += slotGasUsed

		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(accumulatedOriginatedTxsFeeSlot)), common.BigToHash(accumulatedOriginatedTxsFee[i]))
		gasUsed += SstoreGasCost

		// Update accumulated uptime
		accumulatedUptimeSlot, slotGasUsed := getEpochValidatorAccumulatedUptimeSlot(currentEpoch, validatorID)
		gasUsed += slotGasUsed

		prevAccumulatedUptimeSlot, slotGasUsed := getEpochValidatorAccumulatedUptimeSlot(prevEpoch, validatorID)
		gasUsed += slotGasUsed

		prevAccumulatedUptime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(prevAccumulatedUptimeSlot)))
		gasUsed += SloadGasCost
		prevAccumulatedUptimeBigInt := new(big.Int).SetBytes(prevAccumulatedUptime.Bytes())

		newAccumulatedUptime := new(big.Int).Add(prevAccumulatedUptimeBigInt, uptimes[i])
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(accumulatedUptimeSlot)), common.BigToHash(newAccumulatedUptime))
		gasUsed += SstoreGasCost
	}
	// Update epoch fee
	epochFeeSlot := currentEpochSnapshotSlot + epochFeeOffset
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(epochFeeSlot)), common.BigToHash(epochFee))
	gasUsed += SstoreGasCost

	// Update total base reward weight
	totalBaseRewardSlot := currentEpochSnapshotSlot + totalBaseRewardOffset
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalBaseRewardSlot)), common.BigToHash(totalBaseRewardWeight))
	gasUsed += SstoreGasCost

	// Update total tx reward weight
	totalTxRewardSlot := currentEpochSnapshotSlot + totalTxRewardOffset
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalTxRewardSlot)), common.BigToHash(totalTxRewardWeight))
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
		feeShare := new(big.Int).Mul(epochFee, treasuryFeeShareBigInt)
		feeShare = new(big.Int).Div(feeShare, decimalUnitBigInt)

		// Mint native token
		// TODO: Implement _mintNativeToken
		// For now, we'll just update the total supply
		totalSupplyBigInt = new(big.Int).Add(totalSupplyBigInt, feeShare)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)), common.BigToHash(totalSupplyBigInt))
		gasUsed += SstoreGasCost

		// Transfer to treasury
		// TODO: Implement proper transfer
		// For now, we'll just emit a log
		topics := []common.Hash{
			common.BytesToHash([]byte("TreasuryTransfer")),
			common.BytesToHash(treasuryAddr.Bytes()),
		}
		data := common.BigToHash(feeShare).Bytes()

		evm.SfcStateDB.AddLog(&types.Log{
			Address:     ContractAddress,
			Topics:      topics,
			Data:        data,
			BlockNumber: evm.Context.BlockNumber.Uint64(),
		})
	}

	return gasUsed, nil
}

// _sealEpoch_minGasPrice is an internal function to seal minimum gas price in an epoch
func _sealEpoch_minGasPrice(evm *vm.EVM, epochDuration *big.Int, epochGas *big.Int) (uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0

	// Get the target gas power per second from the constants manager
	targetGasPowerPerSecond, cmGasUsed, err := callConstantManagerMethod(evm, "targetGasPowerPerSecond")
	gasUsed += cmGasUsed
	if err != nil || len(targetGasPowerPerSecond) == 0 {
		return gasUsed, err
	}
	targetGasPowerPerSecondBigInt, ok := targetGasPowerPerSecond[0].(*big.Int)
	if !ok {
		return gasUsed, vm.ErrExecutionReverted
	}

	// Calculate target epoch gas
	targetEpochGas := new(big.Int).Mul(epochDuration, targetGasPowerPerSecondBigInt)
	targetEpochGas = new(big.Int).Add(targetEpochGas, big.NewInt(1)) // Add 1 to avoid division by zero

	// Get the decimal unit (1e18) using the helper function
	decimalUnitBigInt := getDecimalUnit()

	// Calculate gas price delta ratio
	gasPriceDeltaRatio := new(big.Int).Mul(epochGas, decimalUnitBigInt)
	gasPriceDeltaRatio = new(big.Int).Div(gasPriceDeltaRatio, targetEpochGas)

	// Get the gas price balancing counterweight from the constants manager
	counterweight, cmGasUsed, err := callConstantManagerMethod(evm, "gasPriceBalancingCounterweight")
	gasUsed += cmGasUsed
	if err != nil || len(counterweight) == 0 {
		return gasUsed, err
	}
	counterweightBigInt, ok := counterweight[0].(*big.Int)
	if !ok {
		return gasUsed, vm.ErrExecutionReverted
	}

	// Scale down the change speed
	// gasPriceDeltaRatio = (epochDuration * gasPriceDeltaRatio + counterweight * Decimal.unit()) / (epochDuration + counterweight)
	term1 := new(big.Int).Mul(epochDuration, gasPriceDeltaRatio)
	term2 := new(big.Int).Mul(counterweightBigInt, decimalUnitBigInt)
	numerator := new(big.Int).Add(term1, term2)
	denominator := new(big.Int).Add(epochDuration, counterweightBigInt)
	gasPriceDeltaRatio = new(big.Int).Div(numerator, denominator)

	// Limit the max/min possible delta in one epoch using the trimGasPriceChangeRatio helper function
	gasPriceDeltaRatio = trimGasPriceChangeRatio(gasPriceDeltaRatio)

	// Get the current min gas price
	minGasPrice := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)))
	gasUsed += SloadGasCost
	minGasPriceBigInt := new(big.Int).SetBytes(minGasPrice.Bytes())

	// Apply the ratio
	newMinGasPrice := new(big.Int).Mul(minGasPriceBigInt, gasPriceDeltaRatio)
	newMinGasPrice = new(big.Int).Div(newMinGasPrice, decimalUnitBigInt)

	// Limit the max/min possible minGasPrice using the trimMinGasPrice helper function
	newMinGasPrice = trimMinGasPrice(newMinGasPrice)

	// Apply new minGasPrice
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(minGasPriceSlot)), common.BigToHash(newMinGasPrice))
	gasUsed += SstoreGasCost

	return gasUsed, nil
}

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

// _mintNativeToken is an internal function to mint native tokens
func handle_mintNativeToken(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _mintNativeToken handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// _scaleLockupReward is an internal function to scale lockup reward
func handle_scaleLockupReward(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement _scaleLockupReward handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
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
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlot)))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	stakeBigInt := new(big.Int).SetBytes(stake.Bytes())

	// Get the delegation locked stake
	lockedStakeSlot, slotGasUsed := getLockedStakeSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	lockedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(lockedStakeSlot)))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	lockedStakeBigInt := new(big.Int).SetBytes(lockedStake.Bytes())

	// Calculate the unlocked stake
	unlockedStake := new(big.Int).Sub(stakeBigInt, lockedStakeBigInt)
	if unlockedStake.Cmp(big.NewInt(0)) < 0 {
		unlockedStake = big.NewInt(0)
	}

	// Pack the result
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
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())

	// Check if the validator is deactivated
	isDeactivated := (validatorStatusBigInt.Bit(0) == 1) // WITHDRAWN_BIT

	// Get the validator auth
	validatorAuthSlot, slotGasUsed := getValidatorAuthSlot(toValidatorID)
	gasUsed += slotGasUsed
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorAuthSlot)))
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
	data := methodID

	// Encode the parameters
	// address delegator
	data = append(data, common.LeftPadBytes(delegator.Bytes(), 32)...)
	// uint256 toValidatorID
	data = append(data, common.LeftPadBytes(toValidatorID.Bytes(), 32)...)

	// Make the call to the StakeTokenizer contract
	result, _, err := evm.Call(vm.AccountRef(ContractAddress), stakeTokenizerAddr, data, defaultGasLimit, big.NewInt(0))
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
	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)))
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

	// Get the max delegated ratio
	maxDelegatedRatioBigInt, _, err := getMaxDelegatedRatio(evm)
	if err != nil {
		return false, err
	}

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

// handleInternalDelegate implements the internal _delegate function logic
func handleInternalDelegate(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int, amount *big.Int) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// This function can either handle the delegation logic directly or call the SFCLib contract
	// For this implementation, we'll handle it directly, but we could also call the SFCLib contract
	// return callSFCLibDelegate(evm, delegator, toValidatorID, amount)

	// Check that the validator exists
	revertData, checkGasUsed, err := checkValidatorExists(evm, toValidatorID, "_delegate")
	gasUsed += checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Check that the validator is active
	revertData, checkGasUsed, err = checkValidatorActive(evm, toValidatorID, "_delegate")
	gasUsed += checkGasUsed
	if err != nil {
		return revertData, gasUsed, err
	}

	// Check that the amount is greater than 0
	if amount.Cmp(big.NewInt(0)) <= 0 {
		revertData, err := encodeRevertReason("_delegate", "zero amount")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the stake
	stakeSlot, _ := getStakeSlot(delegator, toValidatorID)
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlot)))
	stakeBigInt := new(big.Int).SetBytes(stake.Bytes())
	newStake := new(big.Int).Add(stakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlot)), common.BigToHash(newStake))

	// Update the validator's received stake
	validatorReceivedStakeSlot, _ := getValidatorReceivedStakeSlot(toValidatorID)
	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)))
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())
	newReceivedStake := new(big.Int).Add(receivedStakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)), common.BigToHash(newReceivedStake))

	// Update the total stake
	totalStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)))
	totalStakeBigInt := new(big.Int).SetBytes(totalStake.Bytes())
	newTotalStake := new(big.Int).Add(totalStakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)), common.BigToHash(newTotalStake))

	// Update the total active stake if the validator is active
	validatorStatusSlot, slotGasUsed := getValidatorStatusSlot(toValidatorID)
	gasUsed += slotGasUsed
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())
	if validatorStatusBigInt.Cmp(big.NewInt(0)) == 0 { // OK_STATUS
		totalActiveStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
		totalActiveStakeBigInt := new(big.Int).SetBytes(totalActiveStake.Bytes())
		newTotalActiveStake := new(big.Int).Add(totalActiveStakeBigInt, amount)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)), common.BigToHash(newTotalActiveStake))
	}

	return nil, 0, nil
}

// handleRawUndelegate implements the _rawUndelegate function logic
func handleRawUndelegate(evm *vm.EVM, delegator common.Address, toValidatorID *big.Int, amount *big.Int, strict bool) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// Update the stake
	stakeSlot, slotGasUsed := getStakeSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlot)))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	stakeBigInt := new(big.Int).SetBytes(stake.Bytes())
	newStake := new(big.Int).Sub(stakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlot)), common.BigToHash(newStake))
	gasUsed += params.SstoreSetGasEIP2200 // Add gas for SSTORE

	// Update the validator's received stake
	validatorReceivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(toValidatorID)
	gasUsed += slotGasUsed
	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())
	newReceivedStake := new(big.Int).Sub(receivedStakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)), common.BigToHash(newReceivedStake))
	gasUsed += params.SstoreSetGasEIP2200 // Add gas for SSTORE

	// Update the total stake
	totalStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)))
	totalStakeBigInt := new(big.Int).SetBytes(totalStakeState.Bytes())
	newTotalStake := new(big.Int).Sub(totalStakeBigInt, amount)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalStakeSlot)), common.BigToHash(newTotalStake))

	// Update the total active stake if the validator is active
	validatorStatusSlot, _ := getValidatorStatusSlot(toValidatorID)
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)))
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
	selfStakeValues, err := SfcAbi.Methods["getSelfStake"].Outputs.Unpack(result)
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
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)), common.BigToHash(big.NewInt(1))) // WITHDRAWN_BIT
	} else if validatorStatusBigInt.Cmp(big.NewInt(0)) == 0 { // OK_STATUS
		// Check that the self-stake is at least the minimum self-stake
		minSelfStakeBigInt, minSelfStakeGas, err := getMinSelfStake(evm)
		if err != nil {
			return nil, gasUsed + minSelfStakeGas, err
		}
		gasUsed += minSelfStakeGas
		if selfStake.Cmp(minSelfStakeBigInt) < 0 {
			// Set the validator as deactivated
			evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)), common.BigToHash(big.NewInt(1))) // WITHDRAWN_BIT
		} else {
			// Check that the delegated stake is within the limit
			withinLimit, err := handleCheckDelegatedStakeLimit(evm, toValidatorID)
			if err != nil {
				return nil, gasUsed, err
			}
			if !withinLimit {
				// Set the validator as deactivated
				evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)), common.BigToHash(big.NewInt(1))) // WITHDRAWN_BIT
			}
		}
	}

	// Get the validator auth address
	validatorAuthSlot, slotGasUsed := getValidatorAuthSlot(toValidatorID)
	gasUsed += slotGasUsed
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorAuthSlot)))
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

// handleRecountVotes implements the _recountVotes function logic
func handleRecountVotes(evm *vm.EVM, delegator common.Address, validatorAuth common.Address, strict bool) ([]byte, uint64, error) {
	// Get the voteBookAddress
	voteBookAddressSlot, _ := getVoteBookAddressSlot()
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
		data := methodID

		// Encode the parameters
		// address delegator
		data = append(data, common.LeftPadBytes(delegator.Bytes(), 32)...)
		// address validatorAuth
		data = append(data, common.LeftPadBytes(validatorAuth.Bytes(), 32)...)

		// Make the call to the voteBook contract with gas limit of 8000000
		voteBookAddr := common.BytesToAddress(voteBookAddressBytes)
		_, leftOverGas, err := evm.Call(vm.AccountRef(ContractAddress), voteBookAddr, data, 8000000, big.NewInt(0))

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
	result, leftOverGas, err := evm.Call(vm.AccountRef(ContractAddress), sfcLibAddress, data, defaultGasLimit, big.NewInt(0))
	if err != nil {
		return nil, defaultGasLimit - leftOverGas, err
	}

	return result, defaultGasLimit - leftOverGas, nil
}

// handleGetSelfStake returns the self-stake of a validator
func handleGetSelfStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Get the arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	validatorID, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Get the validator auth
	validatorAuthSlot, _ := getValidatorAuthSlot(validatorID)
	validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorAuthSlot)))
	validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

	// Get the self-stake
	stakeSlot, _ := getStakeSlot(validatorAuthAddr, validatorID)
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlot)))
	stakeBigInt := new(big.Int).SetBytes(stake.Bytes())

	// Pack the result
	result, err := SfcAbi.Methods["getSelfStake"].Outputs.Pack(stakeBigInt)
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	return result, 0, nil
}

// handleStashRewards stashes the rewards for a delegator
func handleStashRewards(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
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
	// Get the current epoch
	currentEpochBigInt, _, err := getCurrentEpoch(evm)
	if err != nil {
		return nil, 0, err
	}

	// Get the stashed rewards until epoch
	stashedRewardsUntilEpochSlot, _ := getStashedRewardsUntilEpochSlot(delegator, toValidatorID)
	stashedRewardsUntilEpoch := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stashedRewardsUntilEpochSlot)))
	stashedRewardsUntilEpochBigInt := new(big.Int).SetBytes(stashedRewardsUntilEpoch.Bytes())

	// Check if rewards are already stashed for the current epoch
	if stashedRewardsUntilEpochBigInt.Cmp(currentEpochBigInt) >= 0 {
		return nil, 0, nil
	}

	// Calculate the rewards using _newRewards logic
	// Get the delegation stake
	stakeSlot, _ := getStakeSlot(delegator, toValidatorID)
	stake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(stakeSlot)))
	stakeBigInt := new(big.Int).SetBytes(stake.Bytes())

	// Get the validator's received stake
	validatorReceivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(toValidatorID)
	gasUsed += slotGasUsed
	receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)))
	gasUsed += params.ColdSloadCostEIP2929 // Add gas for SLOAD
	receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

	// Get the validator status
	validatorStatusSlot, _ := getValidatorStatusSlot(toValidatorID)
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)))
	validatorStatusBigInt := new(big.Int).SetBytes(validatorStatus.Bytes())

	// Check if the validator is active
	isActive := (validatorStatusBigInt.Cmp(big.NewInt(0)) == 0) // OK_STATUS

	// Calculate the rewards
	rewards := big.NewInt(0)
	if isActive && stakeBigInt.Cmp(big.NewInt(0)) > 0 && receivedStakeBigInt.Cmp(big.NewInt(0)) > 0 {
		// Get the validator's auth address
		validatorAuthSlot, slotGasUsed := getValidatorAuthSlot(toValidatorID)
		gasUsed += slotGasUsed
		validatorAuth := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorAuthSlot)))
		validatorAuthAddr := common.BytesToAddress(validatorAuth.Bytes())

		// Check if the delegator is the validator (self-stake)
		isSelfStake := delegator == validatorAuthAddr

		// Get the validator commission
		validatorCommission := big.NewInt(0)
		if !isSelfStake {
			validatorCommissionSlot, slotGasUsed := getValidatorCommissionSlot(toValidatorID)
			gasUsed += slotGasUsed
			commission := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorCommissionSlot)))
			validatorCommission = new(big.Int).SetBytes(commission.Bytes())
		}

		// Calculate the base reward rate
		// In Solidity: uint256 baseRewardRate = _calcValidatorBaseRewardRate(toValidatorID, stashedRewardsUntilEpochBigInt, currentEpochBigInt);
		// For simplicity, we'll use a fixed base reward rate
		baseRewardRate := big.NewInt(1000000) // 0.1% per epoch as an example

		// Calculate the reward weight
		// In Solidity: uint256 weightedStake = (delegationStake * validatorBaseRewardWeight) / validatorTotalStake;
		weightedStake := new(big.Int).Mul(stakeBigInt, big.NewInt(1000000)) // Assuming base weight of 1.0
		weightedStake = new(big.Int).Div(weightedStake, receivedStakeBigInt)

		// Calculate the raw rewards
		// In Solidity: uint256 rawReward = (delegationStake * baseRewardRate * (currentEpochBigInt - stashedRewardsUntilEpochBigInt)) / 1e18;
		epochsDiff := new(big.Int).Sub(currentEpochBigInt, stashedRewardsUntilEpochBigInt)
		rawReward := new(big.Int).Mul(stakeBigInt, baseRewardRate)
		rawReward = new(big.Int).Mul(rawReward, epochsDiff)
		rawReward = new(big.Int).Div(rawReward, big.NewInt(1000000000000000000)) // 1e18

		// Apply commission if not self-stake
		if !isSelfStake && validatorCommission.Cmp(big.NewInt(0)) > 0 {
			commissionReward := new(big.Int).Mul(rawReward, validatorCommission)
			commissionReward = new(big.Int).Div(commissionReward, big.NewInt(1000000)) // Assuming commission is in parts per million
			rawReward = new(big.Int).Sub(rawReward, commissionReward)
		}

		rewards = rawReward
	}

	// Get the current stashed rewards
	rewardsStashSlot, slotGasUsed := getRewardsStashSlot(delegator, toValidatorID)
	gasUsed += slotGasUsed
	rewardsStash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(rewardsStashSlot)))
	rewardsStashBigInt := new(big.Int).SetBytes(rewardsStash.Bytes())

	// Add the rewards to the stash
	newRewardsStash := new(big.Int).Add(rewardsStashBigInt, rewards)
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(rewardsStashSlot)), common.BigToHash(newRewardsStash))

	// Update the stashed rewards until epoch
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(stashedRewardsUntilEpochSlot)), common.BigToHash(currentEpochBigInt))

	return nil, 0, nil
}

// handleSyncValidator synchronizes a validator's state
func handleSyncValidator(evm *vm.EVM, validatorID *big.Int) ([]byte, uint64, error) {
	// Initialize gas used
	var gasUsed uint64 = 0
	// Get the validator status
	validatorStatusSlot, slotGasUsed := getValidatorStatusSlot(validatorID)
	gasUsed += slotGasUsed
	validatorStatus := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)))
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
	minSelfStakeBigInt, _, err := getMinSelfStake(evm)
	if err != nil {
		return nil, 0, err
	}

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
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)), common.BigToHash(big.NewInt(1))) // WITHDRAWN_BIT

		// Set the validator deactivated epoch
		validatorDeactivatedEpochSlot, slotGasUsed := getValidatorDeactivatedEpochSlot(validatorID)
		gasUsed += slotGasUsed
		currentEpochBigInt, _, err := getCurrentEpoch(evm)
		if err != nil {
			return nil, 0, err
		}
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorDeactivatedEpochSlot)), common.BigToHash(currentEpochBigInt))

		// Set the validator deactivated time
		validatorDeactivatedTimeSlot, slotGasUsed := getValidatorDeactivatedTimeSlot(validatorID)
		gasUsed += slotGasUsed
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorDeactivatedTimeSlot)), common.BigToHash(evm.Context.Time))

		// Update the total active stake
		validatorReceivedStakeSlot, slotGasUsed := getValidatorReceivedStakeSlot(validatorID)
		gasUsed += slotGasUsed
		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)))
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
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorStatusSlot)), common.BigToHash(big.NewInt(0))) // OK_STATUS

		// Clear the validator deactivated epoch
		validatorDeactivatedEpochSlot, _ := getValidatorDeactivatedEpochSlot(validatorID)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorDeactivatedEpochSlot)), common.BigToHash(big.NewInt(0)))

		// Clear the validator deactivated time
		validatorDeactivatedTimeSlot, _ := getValidatorDeactivatedTimeSlot(validatorID)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorDeactivatedTimeSlot)), common.BigToHash(big.NewInt(0)))

		// Update the total active stake
		validatorReceivedStakeSlot, _ := getValidatorReceivedStakeSlot(validatorID)
		receivedStake := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorReceivedStakeSlot)))
		receivedStakeBigInt := new(big.Int).SetBytes(receivedStake.Bytes())

		totalActiveStakeState := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)))
		totalActiveStakeBigInt := new(big.Int).SetBytes(totalActiveStakeState.Bytes())
		newTotalActiveStake := new(big.Int).Add(totalActiveStakeBigInt, receivedStakeBigInt)
		evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(totalActiveStakeSlot)), common.BigToHash(newTotalActiveStake))
	}

	return nil, 0, nil
}
