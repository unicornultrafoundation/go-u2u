package constant_manager

import (
	"strings"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
)

var (
	ConstantManagerAbi abi.ABI
)

func init() {
	var err error
	ConstantManagerAbi, err = abi.JSON(strings.NewReader(ConstantManagerABIStr))
	if err != nil {
		panic(err)
	}
}

// parseABIInput parses the input data and returns the method and unpacked parameters
func parseABIInput(input []byte) (*abi.Method, []interface{}, error) {
	// Handle empty input (native token transfer) - create a dummy method for fallback
	if len(input) == 0 {
		// Create a dummy method with empty name to trigger the fallback function
		dummyMethod := &abi.Method{
			Name:   "",
			Inputs: abi.Arguments{},
		}
		return dummyMethod, []interface{}{}, nil
	}

	// Need at least 4 bytes for function signature
	if len(input) < 4 {
		return nil, nil, vm.ErrExecutionReverted
	}

	// Get function signature from first 4 bytes
	methodID := input[:4]
	method, err := ConstantManagerAbi.MethodById(methodID)
	if err != nil {
		return nil, nil, vm.ErrExecutionReverted
	}

	// Parse input arguments
	var args []interface{}
	if len(input) > 4 {
		args, err = method.Inputs.Unpack(input[4:])
		if err != nil {
			return nil, nil, vm.ErrExecutionReverted
		}
	}

	return method, args, nil
}

// ConstantManagerPrecompile implements PrecompiledSfcContract interface
type ConstantManagerPrecompile struct{}

// Run runs the precompiled contract
func (c *ConstantManagerPrecompile) Run(evm *vm.EVM, caller common.Address, input []byte, suppliedGas uint64) ([]byte, uint64, error) {
	// Initialize/Invalidate the cache
	if cmCache.NeedInvalidating || cmCache.Values == nil || len(cmCache.Values) == 0 {
		InvalidateCmCache(evm)
	}
	// Parse the input to get method and arguments
	method, args, err := parseABIInput(input)
	if err != nil {
		return nil, 0, err
	}

	var (
		result  []byte
		gasUsed uint64
	)

	log.Info("ConstantsManager Precompiled: Calling function", "function", method.Name, "caller", caller.Hex())

	// Dispatch to the appropriate handler based on the method name
	switch method.Name {
	// Getters for public variables
	case "minSelfStake":
		result, gasUsed, err = handleMinSelfStake(evm)
	case "maxDelegatedRatio":
		result, gasUsed, err = handleMaxDelegatedRatio(evm)
	case "validatorCommission":
		result, gasUsed, err = handleValidatorCommission(evm)
	case "burntFeeShare":
		result, gasUsed, err = handleBurntFeeShare(evm)
	case "treasuryFeeShare":
		result, gasUsed, err = handleTreasuryFeeShare(evm)
	case "unlockedRewardRatio":
		result, gasUsed, err = handleUnlockedRewardRatio(evm)
	case "minLockupDuration":
		result, gasUsed, err = handleMinLockupDuration(evm)
	case "maxLockupDuration":
		result, gasUsed, err = handleMaxLockupDuration(evm)
	case "withdrawalPeriodEpochs":
		result, gasUsed, err = handleWithdrawalPeriodEpochs(evm)
	case "withdrawalPeriodTime":
		result, gasUsed, err = handleWithdrawalPeriodTime(evm)
	case "baseRewardPerSecond":
		result, gasUsed, err = handleBaseRewardPerSecond(evm)
	case "offlinePenaltyThresholdBlocksNum":
		result, gasUsed, err = handleOfflinePenaltyThresholdBlocksNum(evm)
	case "offlinePenaltyThresholdTime":
		result, gasUsed, err = handleOfflinePenaltyThresholdTime(evm)
	case "targetGasPowerPerSecond":
		result, gasUsed, err = handleTargetGasPowerPerSecond(evm)
	case "gasPriceBalancingCounterweight":
		result, gasUsed, err = handleGasPriceBalancingCounterweight(evm)
	case "owner":
		result, gasUsed, err = handleOwner(evm)

	// Setter functions (to be implemented in constant_manager_function_public.go)
	case "updateMinSelfStake":
		result, gasUsed, err = handleUpdateMinSelfStake(evm, args)
	case "updateMaxDelegatedRatio":
		result, gasUsed, err = handleUpdateMaxDelegatedRatio(evm, args)
	case "updateValidatorCommission":
		result, gasUsed, err = handleUpdateValidatorCommission(evm, args)
	case "updateBurntFeeShare":
		result, gasUsed, err = handleUpdateBurntFeeShare(evm, args)
	case "updateTreasuryFeeShare":
		result, gasUsed, err = handleUpdateTreasuryFeeShare(evm, args)
	case "updateUnlockedRewardRatio":
		result, gasUsed, err = handleUpdateUnlockedRewardRatio(evm, args)
	case "updateMinLockupDuration":
		result, gasUsed, err = handleUpdateMinLockupDuration(evm, args)
	case "updateMaxLockupDuration":
		result, gasUsed, err = handleUpdateMaxLockupDuration(evm, args)
	case "updateWithdrawalPeriodEpochs":
		result, gasUsed, err = handleUpdateWithdrawalPeriodEpochs(evm, args)
	case "updateWithdrawalPeriodTime":
		result, gasUsed, err = handleUpdateWithdrawalPeriodTime(evm, args)
	case "updateBaseRewardPerSecond":
		result, gasUsed, err = handleUpdateBaseRewardPerSecond(evm, args)
	case "updateOfflinePenaltyThresholdTime":
		result, gasUsed, err = handleUpdateOfflinePenaltyThresholdTime(evm, args)
	case "updateOfflinePenaltyThresholdBlocksNum":
		result, gasUsed, err = handleUpdateOfflinePenaltyThresholdBlocksNum(evm, args)
	case "updateTargetGasPowerPerSecond":
		result, gasUsed, err = handleUpdateTargetGasPowerPerSecond(evm, args)
	case "updateGasPriceBalancingCounterweight":
		result, gasUsed, err = handleUpdateGasPriceBalancingCounterweight(evm, args)
	case "initialize":
		result, gasUsed, err = handleInitialize(evm, args)
	case "transferOwnership":
		result, gasUsed, err = handleTransferOwnership(evm, args)
	case "renounceOwnership":
		result, gasUsed, err = handleRenounceOwnership(evm, args)

	// Fallback function
	case "":
		result, gasUsed, err = handleFallback(evm, args, input)

	default:
		log.Error("CM Precompiled: Unknown function", "function", method.Name)
		return nil, 0, vm.ErrSfcFunctionNotImplemented
	}
	if err != nil {
		reason, _ := abi.UnpackRevert(result)
		log.Error("CM Precompiled: Revert", "function", method.Name, "err", err, "reason", reason)
		return nil, 0, vm.ErrExecutionReverted
	}

	if suppliedGas < gasUsed {
		log.Error("CM Precompiled: Out of gas", "function", method.Name)
		return nil, 0, vm.ErrOutOfGas
	}

	return result, gasUsed, nil
}
