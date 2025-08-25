package constant_manager

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// Handler functions for ConstantManager contract public functions

// handleUpdateMinSelfStake updates the minimum amount of stake for a validator
func handleUpdateMinSelfStake(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateMinSelfStake")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	minValue := new(big.Int).Mul(big.NewInt(100000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	maxValue := new(big.Int).Mul(big.NewInt(10000000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	if value.Cmp(minValue) < 0 {
		// Return ABI-encoded revert reason: "too small value"
		revertData, err := encodeRevertReason("updateMinSelfStake", "too small value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}
	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateMinSelfStake", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the minSelfStake value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(minSelfStakeSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateMaxDelegatedRatio updates the maximum ratio of delegations a validator can have
func handleUpdateMaxDelegatedRatio(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateMaxDelegatedRatio")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	// Decimal.unit() = 1e18 as defined in Decimal.sol
	decimalUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	minValue := new(big.Int).Set(decimalUnit)                 // 1 * Decimal.unit()
	maxValue := new(big.Int).Mul(big.NewInt(31), decimalUnit) // 31 * Decimal.unit()

	if value.Cmp(minValue) < 0 {
		// Return ABI-encoded revert reason: "too small value"
		revertData, err := encodeRevertReason("updateMaxDelegatedRatio", "too small value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}
	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateMaxDelegatedRatio", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the maxDelegatedRatio value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(maxDelegatedRatioSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateValidatorCommission updates the commission fee percentage a validator gets from delegations
func handleUpdateValidatorCommission(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateValidatorCommission")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	// Decimal.unit() = 1e18 as defined in Decimal.sol
	decimalUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	maxValue := new(big.Int).Div(decimalUnit, big.NewInt(2)) // Decimal.unit() / 2

	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateValidatorCommission", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the validatorCommission value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(validatorCommissionSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateBurntFeeShare updates the percentage of fees to burn
func handleUpdateBurntFeeShare(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateBurntFeeShare")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	// Decimal.unit() = 1e18 as defined in Decimal.sol
	decimalUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	maxValue := new(big.Int).Div(decimalUnit, big.NewInt(2)) // Decimal.unit() / 2

	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateBurntFeeShare", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the burntFeeShare value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(burntFeeShareSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateTreasuryFeeShare updates the percentage of fees to transfer to treasury
func handleUpdateTreasuryFeeShare(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateTreasuryFeeShare")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	// Decimal.unit() = 1e18 as defined in Decimal.sol
	decimalUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	maxValue := new(big.Int).Div(decimalUnit, big.NewInt(2)) // Decimal.unit() / 2

	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateTreasuryFeeShare", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the treasuryFeeShare value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(treasuryFeeShareSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateUnlockedRewardRatio updates the ratio of reward rate at base rate (no lock)
func handleUpdateUnlockedRewardRatio(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateUnlockedRewardRatio")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	// Decimal.unit() = 1e18 as defined in Decimal.sol
	decimalUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	minValue := new(big.Int).Mul(big.NewInt(5), new(big.Int).Div(decimalUnit, big.NewInt(100))) // (5 * Decimal.unit()) / 100
	maxValue := new(big.Int).Div(decimalUnit, big.NewInt(2))                                    // Decimal.unit() / 2

	if value.Cmp(minValue) < 0 {
		// Return ABI-encoded revert reason: "too small value"
		revertData, err := encodeRevertReason("updateUnlockedRewardRatio", "too small value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}
	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateUnlockedRewardRatio", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the unlockedRewardRatio value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(unlockedRewardRatioSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateMinLockupDuration updates the minimum duration of stake/delegation lockup
func handleUpdateMinLockupDuration(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateMinLockupDuration")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	minValue := big.NewInt(86400)                                   // 1 day in seconds
	maxValue := new(big.Int).Mul(big.NewInt(86400), big.NewInt(30)) // 30 days in seconds

	if value.Cmp(minValue) < 0 {
		// Return ABI-encoded revert reason: "too small value"
		revertData, err := encodeRevertReason("updateMinLockupDuration", "too small value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}
	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateMinLockupDuration", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the minLockupDuration value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(minLockupDurationSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateMaxLockupDuration updates the maximum duration of stake/delegation lockup
func handleUpdateMaxLockupDuration(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateMaxLockupDuration")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	minValue := new(big.Int).Mul(big.NewInt(86400), big.NewInt(30))   // 30 days in seconds
	maxValue := new(big.Int).Mul(big.NewInt(86400), big.NewInt(1460)) // 1460 days (4 years) in seconds

	if value.Cmp(minValue) < 0 {
		// Return ABI-encoded revert reason: "too small value"
		revertData, err := encodeRevertReason("updateMaxLockupDuration", "too small value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}
	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateMaxLockupDuration", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the maxLockupDuration value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(maxLockupDurationSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateWithdrawalPeriodEpochs updates the number of epochs undelegated stake is locked for
func handleUpdateWithdrawalPeriodEpochs(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateWithdrawalPeriodEpochs")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	minValue := big.NewInt(2)   // Minimum 2 epochs
	maxValue := big.NewInt(100) // Maximum 100 epochs

	if value.Cmp(minValue) < 0 {
		// Return ABI-encoded revert reason: "too small value"
		revertData, err := encodeRevertReason("updateWithdrawalPeriodEpochs", "too small value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}
	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateWithdrawalPeriodEpochs", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the withdrawalPeriodEpochs value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalPeriodEpochsSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateWithdrawalPeriodTime updates the number of seconds undelegated stake is locked for
func handleUpdateWithdrawalPeriodTime(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateWithdrawalPeriodTime")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	minValue := big.NewInt(86400)                                   // 1 day in seconds
	maxValue := new(big.Int).Mul(big.NewInt(86400), big.NewInt(30)) // 30 days in seconds

	if value.Cmp(minValue) < 0 {
		// Return ABI-encoded revert reason: "too small value"
		revertData, err := encodeRevertReason("updateWithdrawalPeriodTime", "too small value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}
	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateWithdrawalPeriodTime", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the withdrawalPeriodTime value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(withdrawalPeriodTimeSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateBaseRewardPerSecond updates the base reward per second
func handleUpdateBaseRewardPerSecond(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateBaseRewardPerSecond")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	// Decimal.unit() = 1e18 as defined in Decimal.sol
	decimalUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	minValue := new(big.Int).Div(decimalUnit, big.NewInt(2))  // 0.5 * Decimal.unit()
	maxValue := new(big.Int).Mul(big.NewInt(32), decimalUnit) // 32 * Decimal.unit()

	if value.Cmp(minValue) < 0 {
		// Return ABI-encoded revert reason: "too small value"
		revertData, err := encodeRevertReason("updateBaseRewardPerSecond", "too small value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}
	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateBaseRewardPerSecond", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the baseRewardPerSecond value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(baseRewardPerSecondSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateOfflinePenaltyThresholdTime updates the threshold for offline penalty in time
func handleUpdateOfflinePenaltyThresholdTime(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateOfflinePenaltyThresholdTime")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	minValue := big.NewInt(86400)                                   // Minimum 1 day in seconds
	maxValue := new(big.Int).Mul(big.NewInt(86400), big.NewInt(10)) // 10 days in seconds

	if value.Cmp(minValue) < 0 {
		// Return ABI-encoded revert reason: "too small value"
		revertData, err := encodeRevertReason("updateOfflinePenaltyThresholdTime", "too small value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}
	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateOfflinePenaltyThresholdTime", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the offlinePenaltyThresholdTime value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(offlinePenaltyThresholdTimeSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateOfflinePenaltyThresholdBlocksNum updates the threshold for offline penalty in blocks
func handleUpdateOfflinePenaltyThresholdBlocksNum(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateOfflinePenaltyThresholdBlocksNum")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	minValue := big.NewInt(100)     // Minimum 100 blocks
	maxValue := big.NewInt(1000000) // Maximum 1,000,000 blocks

	if value.Cmp(minValue) < 0 {
		// Return ABI-encoded revert reason: "too small value"
		revertData, err := encodeRevertReason("updateOfflinePenaltyThresholdBlocksNum", "too small value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}
	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateOfflinePenaltyThresholdBlocksNum", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the offlinePenaltyThresholdBlocksNum value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(offlinePenaltyThresholdBlocksNumSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateTargetGasPowerPerSecond updates the target gas power per second
func handleUpdateTargetGasPowerPerSecond(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateTargetGasPowerPerSecond")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	minValue := big.NewInt(1000000)   // Minimum 1,000,000 gas power per second
	maxValue := big.NewInt(500000000) // Maximum 500,000,000 gas power per second

	if value.Cmp(minValue) < 0 {
		// Return ABI-encoded revert reason: "too small value"
		revertData, err := encodeRevertReason("updateTargetGasPowerPerSecond", "too small value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}
	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateTargetGasPowerPerSecond", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the targetGasPowerPerSecond value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(targetGasPowerPerSecondSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleUpdateGasPriceBalancingCounterweight updates the gas price balancing counterweight
func handleUpdateGasPriceBalancingCounterweight(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "updateGasPriceBalancingCounterweight")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new value from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	value, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Validate the value
	minValue := big.NewInt(100) // Minimum 100
	maxValue := big.NewInt(10 * 86400) // Maximum 864,000

	if value.Cmp(minValue) < 0 {
		// Return ABI-encoded revert reason: "too small value"
		revertData, err := encodeRevertReason("updateGasPriceBalancingCounterweight", "too small value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}
	if value.Cmp(maxValue) > 0 {
		// Return ABI-encoded revert reason: "too large value"
		revertData, err := encodeRevertReason("updateGasPriceBalancingCounterweight", "too large value")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Update the gasPriceBalancingCounterweight value
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(gasPriceBalancingCounterweightSlot)), common.BigToHash(value))

	return nil, 0, nil
}

// handleInitialize initializes the ConstantManager contract
func handleInitialize(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if contract is already initialized
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	emptyHash := common.Hash{}
	if owner.Cmp(emptyHash) != 0 {
		// Return ABI-encoded revert reason: "already initialized"
		revertData, err := encodeRevertReason("initialize", "already initialized")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Get the owner address from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	ownerAddr, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the owner is not the zero address
	emptyAddr := common.Address{}
	if ownerAddr.Cmp(emptyAddr) == 0 {
		// Return ABI-encoded revert reason: "Ownable: new owner is the zero address"
		revertData, err := encodeRevertReason("initialize", "Ownable: new owner is the zero address")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Set the owner
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)), ownerAddr.Hash())

	// Update the cache
	UpdateOwner(ownerAddr)

	// Emit OwnershipTransferred event
	topics := []common.Hash{
		ConstantManagerAbi.Events["OwnershipTransferred"].ID,
		emptyHash, // indexed parameter (previous owner - zero address)
		common.BytesToHash(common.LeftPadBytes(ownerAddr.Bytes(), 32)), // indexed parameter (new owner)
	}
	data := []byte{}
	evm.SfcStateDB.AddLog(&types.Log{
		BlockNumber: evm.Context.BlockNumber.Uint64(),
		Address:     ContractAddress,
		Topics:      topics,
		Data:        data,
	})

	return nil, 0, nil
}

// handleTransferOwnership transfers ownership of the contract
func handleTransferOwnership(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "transferOwnership")
	if err != nil {
		return revertData, 0, err
	}

	// Get the new owner address from args
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	newOwner, ok := args[0].(common.Address)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Check that the new owner is not the zero address
	emptyAddr := common.Address{}
	if newOwner.Cmp(emptyAddr) == 0 {
		// Return ABI-encoded revert reason: "Ownable: new owner is the zero address"
		revertData, err := encodeRevertReason("transferOwnership", "Ownable: new owner is the zero address")
		if err != nil {
			return nil, 0, vm.ErrExecutionReverted
		}
		return revertData, 0, vm.ErrExecutionReverted
	}

	// Get the current owner
	currentOwner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	currentOwnerAddr := common.BytesToAddress(currentOwner.Bytes())

	// Set the new owner
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)), newOwner.Hash())

	// Update the cache
	UpdateOwner(newOwner)

	// Emit OwnershipTransferred event
	topics := []common.Hash{
		ConstantManagerAbi.Events["OwnershipTransferred"].ID,
		common.BytesToHash(common.LeftPadBytes(currentOwnerAddr.Bytes(), 32)), // indexed parameter (previous owner)
		common.BytesToHash(common.LeftPadBytes(newOwner.Bytes(), 32)),         // indexed parameter (new owner)
	}
	data := []byte{}
	evm.SfcStateDB.AddLog(&types.Log{
		BlockNumber: evm.Context.BlockNumber.Uint64(),
		Address:     ContractAddress,
		Topics:      topics,
		Data:        data,
	})

	return nil, 0, nil
}

// handleRenounceOwnership renounces ownership of the contract
func handleRenounceOwnership(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	// Check if caller is the owner
	revertData, err := checkOnlyOwner(evm, "renounceOwnership")
	if err != nil {
		return revertData, 0, err
	}

	// Get the current owner
	currentOwner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	currentOwnerAddr := common.BytesToAddress(currentOwner.Bytes())

	// Set the owner to the zero address
	emptyHash := common.Hash{}
	evm.SfcStateDB.SetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)), emptyHash)

	// Update the cache
	UpdateOwner(common.Address{})

	// Emit OwnershipTransferred event
	topics := []common.Hash{
		ConstantManagerAbi.Events["OwnershipTransferred"].ID,
		common.BytesToHash(common.LeftPadBytes(currentOwnerAddr.Bytes(), 32)), // indexed parameter (previous owner)
		emptyHash, // indexed parameter (new owner - zero address)
	}
	data := []byte{}
	evm.SfcStateDB.AddLog(&types.Log{
		BlockNumber: evm.Context.BlockNumber.Uint64(),
		Address:     ContractAddress,
		Topics:      topics,
		Data:        data,
	})

	return nil, 0, nil
}
