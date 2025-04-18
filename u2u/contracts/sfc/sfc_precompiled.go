package sfc

import (
	"math/big"
	"strings"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/gossip/contract/sfc100"
	"github.com/unicornultrafoundation/go-u2u/log"
)

var (
	SfcAbi            abi.ABI
	CMAbi             abi.ABI
	NodeDriverAbi     abi.ABI
	NodeDriverAuthAbi abi.ABI
)

func init() {
	SfcAbi, _ = abi.JSON(strings.NewReader(sfc100.ContractMetaData.ABI))
	CMAbi, _ = abi.JSON(strings.NewReader(ConstantManagerABIStr))
	NodeDriverAbi, _ = abi.JSON(strings.NewReader(NodeDriverABIStr))
	NodeDriverAuthAbi, _ = abi.JSON(strings.NewReader(NodeDriverAuthABIStr))
}

// SfcPrecompile implements PrecompiledSfcContract interface
type SfcPrecompile struct{}

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
	method, err := SfcAbi.MethodById(methodID)
	if err != nil {
		return nil, nil, vm.ErrExecutionReverted
	}

	// Parse input arguments
	args, err := method.Inputs.Unpack(input[4:])
	if err != nil {
		return nil, nil, vm.ErrExecutionReverted
	}

	return method, args, nil
}

// Run runs the precompiled contract
func (p *SfcPrecompile) Run(evm *vm.EVM, caller common.Address, input []byte, suppliedGas uint64) ([]byte, uint64, error) {
	// We'll use evm.SfcStateDB directly in the handler functions
	// Parse the input to get method and arguments
	method, args, err := parseABIInput(input)
	if err != nil {
		return nil, 0, err
	}

	var result []byte
	var gasUsed uint64
	switch method.Name {
	case "owner":
		result, gasUsed, err = handleOwner(evm)

	case "currentSealedEpoch":
		result, gasUsed, err = handleCurrentSealedEpoch(evm)

	case "lastValidatorID":
		result, gasUsed, err = handleLastValidatorID(evm)

	case "totalStake":
		result, gasUsed, err = handleTotalStake(evm)

	case "totalActiveStake":
		result, gasUsed, err = handleTotalActiveStake(evm)

	case "totalSlashedStake":
		result, gasUsed, err = handleTotalSlashedStake(evm)

	case "totalSupply":
		result, gasUsed, err = handleTotalSupply(evm)

	case "stakeTokenizerAddress":
		result, gasUsed, err = handleStakeTokenizerAddress(evm)

	case "minGasPrice":
		result, gasUsed, err = handleMinGasPrice(evm)

	case "treasuryAddress":
		result, gasUsed, err = handleTreasuryAddress(evm)

	case "voteBookAddress":
		result, gasUsed, err = handleVoteBookAddress(evm)

	case "getValidator":
		result, gasUsed, err = handleGetValidator(evm, args)

	case "getValidatorID":
		result, gasUsed, err = handleGetValidatorID(evm, args)

	case "getValidatorPubkey":
		result, gasUsed, err = handleGetValidatorPubkey(evm, args)

	case "stashedRewardsUntilEpoch":
		result, gasUsed, err = handleStashedRewardsUntilEpoch(evm, args)

	case "getWithdrawalRequest":
		result, gasUsed, err = handleGetWithdrawalRequest(evm, args)

	case "getStake":
		result, gasUsed, err = handleGetStake(evm, args)

	case "getLockupInfo":
		result, gasUsed, err = handleGetLockupInfo(evm, args)

	case "getStashedLockupRewards":
		result, gasUsed, err = handleGetStashedLockupRewards(evm, args)

	case "slashingRefundRatio":
		result, gasUsed, err = handleSlashingRefundRatio(evm, args)

	case "getEpochSnapshot":
		result, gasUsed, err = handleGetEpochSnapshot(evm, args)

	// Public function handlers - Read-only methods
	case "version":
		result, gasUsed, err = handleVersion(evm, args)

	case "currentEpoch":
		result, gasUsed, err = handleCurrentEpoch(evm)

	case "constsAddress":
		result, gasUsed, err = handleConstsAddress(evm)

	case "getEpochValidatorIDs":
		result, gasUsed, err = handleGetEpochValidatorIDs(evm, args)

	case "getEpochReceivedStake":
		result, gasUsed, err = handleGetEpochReceivedStake(evm, args)

	case "getEpochAccumulatedRewardPerToken":
		result, gasUsed, err = handleGetEpochAccumulatedRewardPerToken(evm, args)

	case "getEpochAccumulatedUptime":
		result, gasUsed, err = handleGetEpochAccumulatedUptime(evm, args)

	case "getEpochAccumulatedOriginatedTxsFee":
		result, gasUsed, err = handleGetEpochAccumulatedOriginatedTxsFee(evm, args)

	case "getEpochOfflineTime":
		result, gasUsed, err = handleGetEpochOfflineTime(evm, args)

	case "getEpochOfflineBlocks":
		result, gasUsed, err = handleGetEpochOfflineBlocks(evm, args)

	case "rewardsStash":
		result, gasUsed, err = handleRewardsStash(evm, args)

	case "getLockedStake":
		result, gasUsed, err = handleGetLockedStake(evm, args)

	case "getSelfStake":
		result, gasUsed, err = handleGetSelfStake(evm, args)

	case "isSlashed":
		result, gasUsed, err = handleIsSlashed(evm, args)

	case "pendingRewards":
		result, gasUsed, err = handlePendingRewards(evm, args)

	case "isLockedUp":
		result, gasUsed, err = handleIsLockedUp(evm, args)

	case "getUnlockedStake":
		result, gasUsed, err = handleGetUnlockedStake(evm, args)

	case "isOwner":
		result, gasUsed, err = handleIsOwner(evm, args)

	// Public function handlers - State-changing methods
	case "renounceOwnership":
		result, gasUsed, err = handleRenounceOwnership(evm, caller, args)

	case "transferOwnership":
		result, gasUsed, err = handleTransferOwnership(evm, caller, args)

	case "updateConstsAddress":
		result, gasUsed, err = handleUpdateConstsAddress(evm, args)

	case "updateLibAddress":
		result, gasUsed, err = handleUpdateLibAddress(evm, caller, args)

	case "updateStakeTokenizerAddress":
		result, gasUsed, err = handleUpdateStakeTokenizerAddress(evm, caller, args)

	case "updateTreasuryAddress":
		result, gasUsed, err = handleUpdateTreasuryAddress(evm, args)

	case "updateVoteBookAddress":
		result, gasUsed, err = handleUpdateVoteBookAddress(evm, args)

	case "createValidator":
		// For createValidator, we need to pass a value, but we don't have direct access to it
		// Use a zero value for now, this should be fixed in a future update
		result, gasUsed, err = handleCreateValidator(evm, caller, args, big.NewInt(0))

	case "delegate":
		// For delegate, we need to pass a value, but we don't have direct access to it
		// Use a zero value for now, this should be fixed in a future update
		result, gasUsed, err = handleDelegate(evm, caller, args, big.NewInt(0))

	case "undelegate":
		result, gasUsed, err = handleUndelegate(evm, caller, args)

	case "withdraw":
		result, gasUsed, err = handleWithdraw(evm, caller, args)

	case "deactivateValidator":
		result, gasUsed, err = handleDeactivateValidator(evm, args)

	case "stashRewards":
		result, gasUsed, err = handleStashRewards(evm, args)

	case "claimRewards":
		result, gasUsed, err = handleClaimRewards(evm, caller, args)

	case "restakeRewards":
		result, gasUsed, err = handleRestakeRewards(evm, caller, args)

	case "updateBaseRewardPerSecond":
		result, gasUsed, err = handleUpdateBaseRewardPerSecond(evm, args)

	case "updateOfflinePenaltyThreshold":
		result, gasUsed, err = handleUpdateOfflinePenaltyThreshold(evm, args)

	case "updateSlashingRefundRatio":
		result, gasUsed, err = handleUpdateSlashingRefundRatio(evm, args)

	case "mintU2U":
		result, gasUsed, err = handleMintU2U(evm, args)

	case "burnU2U":
		result, gasUsed, err = handleBurnU2U(evm, args)

	case "sealEpoch":
		result, gasUsed, err = handleSealEpoch(evm, caller, args)

	case "sealEpochValidators":
		result, gasUsed, err = handleSealEpochValidators(evm, caller, args)

	case "lockStake":
		result, gasUsed, err = handleLockStake(evm, caller, args)

	case "relockStake":
		result, gasUsed, err = handleRelockStake(evm, args)

	case "unlockStake":
		result, gasUsed, err = handleUnlockStake(evm, caller, args)

	case "initialize":
		result, gasUsed, err = handleInitialize(evm, caller, args)

	case "setGenesisValidator":
		result, gasUsed, err = handleSetGenesisValidator(evm, args)

	case "setGenesisDelegation":
		result, gasUsed, err = handleSetGenesisDelegation(evm, args)

	case "sumRewards":
		result, gasUsed, err = handleSumRewards(evm, args)

	// Private function handlers
	case "_delegate":
		result, gasUsed, err = handle_delegate(evm, args, input)

	// These internal functions are now implemented directly in the handleSealEpoch function
	// and are no longer called separately
	case "_sealEpoch_offline":
		return nil, 0, vm.ErrSfcFunctionNotImplemented

	case "_sealEpoch_rewards":
		return nil, 0, vm.ErrSfcFunctionNotImplemented

	case "_sealEpoch_minGasPrice":
		return nil, 0, vm.ErrSfcFunctionNotImplemented

	case "_calcRawValidatorEpochBaseReward":
		result, gasUsed, err = handle_calcRawValidatorEpochBaseReward(evm, args)

	case "_calcRawValidatorEpochTxReward":
		result, gasUsed, err = handle_calcRawValidatorEpochTxReward(evm, args)

	case "_calcValidatorCommission":
		result, gasUsed, err = handle_calcValidatorCommission(evm, args)

	case "_mintNativeToken":
		result, gasUsed, err = handle_mintNativeToken(evm, args)

	case "_scaleLockupReward":
		result, gasUsed, err = handle_scaleLockupReward(evm, args)

	case "_setValidatorDeactivated":
		result, gasUsed, err = handle_setValidatorDeactivated(evm, args)

	case "_syncValidator":
		result, gasUsed, err = handle_syncValidator(evm, args)

	case "_validatorExists":
		result, gasUsed, err = handle_validatorExists(evm, args)

	case "_now":
		result, gasUsed, err = handle_now(evm, args)

	case "getSlashingPenalty":
		result, gasUsed, err = handleGetSlashingPenalty(evm, args)

	// Fallback function
	case "":
		result, gasUsed, err = handleFallback(evm, args, input)

	default:
		log.Debug("SFC Precompiled: Unknown function", "function", method.Name)
		return nil, 0, vm.ErrSfcFunctionNotImplemented
	}
	if err != nil {
		reason, _ := abi.UnpackRevert(result)
		log.Error("SFC Precompiled: Revert", "function", method.Name, "err", err, "reason", reason)
		return nil, 0, vm.ErrExecutionReverted
	}

	if suppliedGas < gasUsed {
		log.Error("SFC Precompiled: Out of gas", "function", method.Name, "suppliedGas", suppliedGas, "gasUsed", gasUsed)
		return nil, 0, vm.ErrOutOfGas
	}
	log.Debug("SFC Precompiled: Success", "function", method.Name)

	return result, gasUsed, nil
}
