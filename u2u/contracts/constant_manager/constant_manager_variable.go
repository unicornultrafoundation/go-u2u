package constant_manager

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// handleMinSelfStake returns the minimum amount of stake for a validator
func handleMinSelfStake(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(minSelfStakeSlot)))
	result, err := ConstantManagerAbi.Methods["minSelfStake"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleMaxDelegatedRatio returns the maximum ratio of delegations a validator can have
func handleMaxDelegatedRatio(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(maxDelegatedRatioSlot)))
	result, err := ConstantManagerAbi.Methods["maxDelegatedRatio"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleValidatorCommission returns the commission fee percentage a validator gets from delegations
func handleValidatorCommission(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(validatorCommissionSlot)))
	result, err := ConstantManagerAbi.Methods["validatorCommission"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleBurntFeeShare returns the percentage of fees to burn
func handleBurntFeeShare(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(burntFeeShareSlot)))
	result, err := ConstantManagerAbi.Methods["burntFeeShare"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleTreasuryFeeShare returns the percentage of fees to transfer to treasury
func handleTreasuryFeeShare(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(treasuryFeeShareSlot)))
	result, err := ConstantManagerAbi.Methods["treasuryFeeShare"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleUnlockedRewardRatio returns the ratio of reward rate at base rate (no lock)
func handleUnlockedRewardRatio(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(unlockedRewardRatioSlot)))
	result, err := ConstantManagerAbi.Methods["unlockedRewardRatio"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleMinLockupDuration returns the minimum duration of stake/delegation lockup
func handleMinLockupDuration(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(minLockupDurationSlot)))
	result, err := ConstantManagerAbi.Methods["minLockupDuration"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleMaxLockupDuration returns the maximum duration of stake/delegation lockup
func handleMaxLockupDuration(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(maxLockupDurationSlot)))
	result, err := ConstantManagerAbi.Methods["maxLockupDuration"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleWithdrawalPeriodEpochs returns the number of epochs undelegated stake is locked for
func handleWithdrawalPeriodEpochs(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalPeriodEpochsSlot)))
	result, err := ConstantManagerAbi.Methods["withdrawalPeriodEpochs"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleWithdrawalPeriodTime returns the number of seconds undelegated stake is locked for
func handleWithdrawalPeriodTime(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalPeriodTimeSlot)))
	result, err := ConstantManagerAbi.Methods["withdrawalPeriodTime"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleBaseRewardPerSecond returns the base reward per second
func handleBaseRewardPerSecond(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(baseRewardPerSecondSlot)))
	result, err := ConstantManagerAbi.Methods["baseRewardPerSecond"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleOfflinePenaltyThresholdBlocksNum returns the threshold for offline penalty in blocks
func handleOfflinePenaltyThresholdBlocksNum(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(offlinePenaltyThresholdBlocksNumSlot)))
	result, err := ConstantManagerAbi.Methods["offlinePenaltyThresholdBlocksNum"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleOfflinePenaltyThresholdTime returns the threshold for offline penalty in time
func handleOfflinePenaltyThresholdTime(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(offlinePenaltyThresholdTimeSlot)))
	result, err := ConstantManagerAbi.Methods["offlinePenaltyThresholdTime"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleTargetGasPowerPerSecond returns the target gas power per second
func handleTargetGasPowerPerSecond(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(targetGasPowerPerSecondSlot)))
	result, err := ConstantManagerAbi.Methods["targetGasPowerPerSecond"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleGasPriceBalancingCounterweight returns the gas price balancing counterweight
func handleGasPriceBalancingCounterweight(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(gasPriceBalancingCounterweightSlot)))
	result, err := ConstantManagerAbi.Methods["gasPriceBalancingCounterweight"].Outputs.Pack(val.Big())
	return result, 0, err
}

// handleOwner returns the owner of the contract
func handleOwner(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	result, err := ConstantManagerAbi.Methods["owner"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
	return result, 0, err
}
