package constant_manager

import (
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// Handler functions for ConstantManager contract public functions

// handleUpdateMinSelfStake updates the minimum amount of stake for a validator
func handleUpdateMinSelfStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateMinSelfStake handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateMaxDelegatedRatio updates the maximum ratio of delegations a validator can have
func handleUpdateMaxDelegatedRatio(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateMaxDelegatedRatio handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateValidatorCommission updates the commission fee percentage a validator gets from delegations
func handleUpdateValidatorCommission(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateValidatorCommission handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateBurntFeeShare updates the percentage of fees to burn
func handleUpdateBurntFeeShare(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateBurntFeeShare handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateTreasuryFeeShare updates the percentage of fees to transfer to treasury
func handleUpdateTreasuryFeeShare(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateTreasuryFeeShare handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateUnlockedRewardRatio updates the ratio of reward rate at base rate (no lock)
func handleUpdateUnlockedRewardRatio(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateUnlockedRewardRatio handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateMinLockupDuration updates the minimum duration of stake/delegation lockup
func handleUpdateMinLockupDuration(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateMinLockupDuration handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateMaxLockupDuration updates the maximum duration of stake/delegation lockup
func handleUpdateMaxLockupDuration(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateMaxLockupDuration handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateWithdrawalPeriodEpochs updates the number of epochs undelegated stake is locked for
func handleUpdateWithdrawalPeriodEpochs(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateWithdrawalPeriodEpochs handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateWithdrawalPeriodTime updates the number of seconds undelegated stake is locked for
func handleUpdateWithdrawalPeriodTime(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateWithdrawalPeriodTime handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateBaseRewardPerSecond updates the base reward per second
func handleUpdateBaseRewardPerSecond(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateBaseRewardPerSecond handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateOfflinePenaltyThresholdTime updates the threshold for offline penalty in time
func handleUpdateOfflinePenaltyThresholdTime(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateOfflinePenaltyThresholdTime handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateOfflinePenaltyThresholdBlocksNum updates the threshold for offline penalty in blocks
func handleUpdateOfflinePenaltyThresholdBlocksNum(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateOfflinePenaltyThresholdBlocksNum handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateTargetGasPowerPerSecond updates the target gas power per second
func handleUpdateTargetGasPowerPerSecond(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateTargetGasPowerPerSecond handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleUpdateGasPriceBalancingCounterweight updates the gas price balancing counterweight
func handleUpdateGasPriceBalancingCounterweight(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement updateGasPriceBalancingCounterweight handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleInitialize initializes the ConstantManager contract
func handleInitialize(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement initialize handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleTransferOwnership transfers ownership of the contract
func handleTransferOwnership(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement transferOwnership handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

// handleRenounceOwnership renounces ownership of the contract
func handleRenounceOwnership(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// TODO: Implement renounceOwnership handler
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}
