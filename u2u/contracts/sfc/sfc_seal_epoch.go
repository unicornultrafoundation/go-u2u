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
	validatorIDsSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(validatorIDsOffset))
	validatorIDsLengthHash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(validatorIDsSlot))
	gasUsed += SloadGasCost

	// Convert the length hash to a big.Int
	validatorIDsLength := new(big.Int).SetBytes(validatorIDsLengthHash.Bytes()).Uint64()

	// Calculate the base slot for the array elements
	// The array elements start at keccak256(slot)
	validatorIDsBaseSlotBytes := crypto.Keccak256(common.BigToHash(validatorIDsSlot).Bytes())
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

	// Get the previous epoch (corresponds to prevSnapshot in Solidity)
	// In Solidity, this is "currentSealedEpoch"
	currentSealedEpochHash := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(currentSealedEpochSlot)))
	gasUsed += SloadGasCost
	prevEpochBigInt := new(big.Int).SetBytes(currentSealedEpochHash.Bytes())

	// Get the epoch snapshot slot for the previous epoch
	// This is equivalent to "prevSnapshot" in the Solidity code
	prevEpochSnapshotSlot, slotGasUsed := getEpochSnapshotSlot(prevEpochBigInt)
	gasUsed += slotGasUsed

	// Get the end time of the previous epoch
	prevEndTimeSlot := new(big.Int).Add(prevEpochSnapshotSlot, big.NewInt(endTimeOffset))
	prevEndTime := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(prevEndTimeSlot))
	gasUsed += SloadGasCost
	prevEndTimeBigInt := new(big.Int).SetBytes(prevEndTime.Bytes())

	// Calculate epoch duration
	epochDuration := big.NewInt(1) // Default to 1 if current time <= prevEndTime
	if evm.Context.Time.Cmp(prevEndTimeBigInt) > 0 {
		epochDuration = new(big.Int).Sub(evm.Context.Time, prevEndTimeBigInt)
	}

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
	endTimeSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(endTimeOffset))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(endTimeSlot), common.BigToHash(evm.Context.Time))
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

	// Update epoch snapshot base reward per second (snapshot.baseRewardPerSecond = c.baseRewardPerSecond() in Solidity)
	baseRewardPerSecondSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(baseRewardPerSecondOffset))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(baseRewardPerSecondSlot), common.BigToHash(baseRewardPerSecondBigInt))
	gasUsed += SstoreGasCost

	// Get the total supply
	totalSupply := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(totalSupplySlot)))
	gasUsed += SloadGasCost
	totalSupplyBigInt := new(big.Int).SetBytes(totalSupply.Bytes())

	// Update epoch snapshot total supply (snapshot.totalSupply = totalSupply in Solidity)
	totalSupplySnapshotSlot := new(big.Int).Add(currentEpochSnapshotSlot, big.NewInt(totalSupplyOffset))
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(totalSupplySnapshotSlot), common.BigToHash(totalSupplyBigInt))
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
