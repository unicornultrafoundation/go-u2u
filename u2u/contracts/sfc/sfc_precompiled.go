package sfc

import (
	"strings"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/gossip/contract/sfc100"
)

var (
	SfcAbi abi.ABI
)

func init() {
	SfcAbi, _ = abi.JSON(strings.NewReader(sfc100.ContractMetaData.ABI))
}

// SfcPrecompile implements PrecompiledSfcContract interface
type SfcPrecompile struct{}

// parseABIInput parses the input data and returns the method and unpacked parameters
func parseABIInput(input []byte) (*abi.Method, []interface{}, error) {
	// Handle empty input (native token transfer) - create a dummy method for fallback
	if len(input) == 0 {
		// Create a dummy method with empty name to trigger the fallback function
		dummyMethod := &abi.Method{
			Name: "",
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
	// Extract stateDB from the EVM
	stateDB := evm.SfcStateDB
	// Parse the input to get method and arguments
	method, args, err := parseABIInput(input)
	if err != nil {
		return nil, 0, err
	}

	var result []byte
	var gasUsed uint64
	switch method.Name {
	case "owner":
		result, gasUsed, err = handleOwner(stateDB)

	case "currentSealedEpoch":
		result, gasUsed, err = handleCurrentSealedEpoch(stateDB)

	case "lastValidatorID":
		result, gasUsed, err = handleLastValidatorID(stateDB)

	case "totalStake":
		result, gasUsed, err = handleTotalStake(stateDB)

	case "totalActiveStake":
		result, gasUsed, err = handleTotalActiveStake(stateDB)

	case "totalSlashedStake":
		result, gasUsed, err = handleTotalSlashedStake(stateDB)

	case "totalSupply":
		result, gasUsed, err = handleTotalSupply(stateDB)

	case "stakeTokenizerAddress":
		result, gasUsed, err = handleStakeTokenizerAddress(stateDB)

	case "minGasPrice":
		result, gasUsed, err = handleMinGasPrice(stateDB)

	case "treasuryAddress":
		result, gasUsed, err = handleTreasuryAddress(stateDB)

	case "voteBookAddress":
		result, gasUsed, err = handleVoteBookAddress(stateDB)

	case "getValidator":
		result, gasUsed, err = handleGetValidator(stateDB, args)

	case "getValidatorID":
		result, gasUsed, err = handleGetValidatorID(stateDB, args)

	case "getValidatorPubkey":
		result, gasUsed, err = handleGetValidatorPubkey(stateDB, args)

	case "stashedRewardsUntilEpoch":
		result, gasUsed, err = handleStashedRewardsUntilEpoch(stateDB, args)

	case "getWithdrawalRequest":
		result, gasUsed, err = handleGetWithdrawalRequest(stateDB, args)

	case "getStake":
		result, gasUsed, err = handleGetStake(stateDB, args)

	case "getLockupInfo":
		result, gasUsed, err = handleGetLockupInfo(stateDB, args)

	case "getStashedLockupRewards":
		result, gasUsed, err = handleGetStashedLockupRewards(stateDB, args)

	case "slashingRefundRatio":
		result, gasUsed, err = handleSlashingRefundRatio(stateDB, args)

	case "getEpochSnapshot":
		result, gasUsed, err = handleGetEpochSnapshot(stateDB, args)

	// Public function handlers - Read-only methods
	case "version":
		result, gasUsed, err = handleVersion(stateDB, args)

	case "currentEpoch":
		result, gasUsed, err = handleCurrentEpoch(stateDB)

	case "constsAddress":
		result, gasUsed, err = handleConstsAddress(stateDB)

	case "getEpochValidatorIDs":
		result, gasUsed, err = handleGetEpochValidatorIDs(stateDB, args)

	case "getEpochReceivedStake":
		result, gasUsed, err = handleGetEpochReceivedStake(stateDB, args)

	case "getEpochAccumulatedRewardPerToken":
		result, gasUsed, err = handleGetEpochAccumulatedRewardPerToken(stateDB, args)

	case "getEpochAccumulatedUptime":
		result, gasUsed, err = handleGetEpochAccumulatedUptime(stateDB, args)

	case "getEpochAccumulatedOriginatedTxsFee":
		result, gasUsed, err = handleGetEpochAccumulatedOriginatedTxsFee(stateDB, args)

	case "getEpochOfflineTime":
		result, gasUsed, err = handleGetEpochOfflineTime(stateDB, args)

	case "getEpochOfflineBlocks":
		result, gasUsed, err = handleGetEpochOfflineBlocks(stateDB, args)

	case "rewardsStash":
		result, gasUsed, err = handleRewardsStash(stateDB, args)

	case "getLockedStake":
		result, gasUsed, err = handleGetLockedStake(stateDB, args)

	case "getSelfStake":
		result, gasUsed, err = handleGetSelfStake(stateDB, args)

	case "isSlashed":
		result, gasUsed, err = handleIsSlashed(stateDB, args)

	case "pendingRewards":
		result, gasUsed, err = handlePendingRewards(stateDB, args)

	case "isLockedUp":
		result, gasUsed, err = handleIsLockedUp(stateDB, args)

	case "getUnlockedStake":
		result, gasUsed, err = handleGetUnlockedStake(stateDB, args)

	case "isOwner":
		result, gasUsed, err = handleIsOwner(stateDB, args)

	// Public function handlers - State-changing methods
	case "renounceOwnership":
		result, gasUsed, err = handleRenounceOwnership(stateDB, args)

	case "transferOwnership":
		result, gasUsed, err = handleTransferOwnership(stateDB, args)

	case "updateConstsAddress":
		result, gasUsed, err = handleUpdateConstsAddress(stateDB, args)

	case "updateLibAddress":
		result, gasUsed, err = handleUpdateLibAddress(stateDB, args)

	case "updateStakeTokenizerAddress":
		result, gasUsed, err = handleUpdateStakeTokenizerAddress(stateDB, args)

	case "updateTreasuryAddress":
		result, gasUsed, err = handleUpdateTreasuryAddress(stateDB, args)

	case "updateVoteBookAddress":
		result, gasUsed, err = handleUpdateVoteBookAddress(stateDB, args)

	case "createValidator":
		result, gasUsed, err = handleCreateValidator(stateDB, args)

	case "delegate":
		result, gasUsed, err = handleDelegate(stateDB, args)

	case "undelegate":
		result, gasUsed, err = handleUndelegate(stateDB, args)

	case "withdraw":
		result, gasUsed, err = handleWithdraw(stateDB, args)

	case "deactivateValidator":
		result, gasUsed, err = handleDeactivateValidator(stateDB, args)

	case "stashRewards":
		result, gasUsed, err = handleStashRewards(stateDB, args)

	case "claimRewards":
		result, gasUsed, err = handleClaimRewards(stateDB, args)

	case "restakeRewards":
		result, gasUsed, err = handleRestakeRewards(stateDB, args)

	case "updateBaseRewardPerSecond":
		result, gasUsed, err = handleUpdateBaseRewardPerSecond(stateDB, args)

	case "updateOfflinePenaltyThreshold":
		result, gasUsed, err = handleUpdateOfflinePenaltyThreshold(stateDB, args)

	case "updateSlashingRefundRatio":
		result, gasUsed, err = handleUpdateSlashingRefundRatio(stateDB, args)

	case "mintU2U":
		result, gasUsed, err = handleMintU2U(stateDB, args)

	case "burnU2U":
		result, gasUsed, err = handleBurnU2U(stateDB, args)

	case "sealEpoch":
		result, gasUsed, err = handleSealEpoch(stateDB, args)

	case "sealEpochValidators":
		result, gasUsed, err = handleSealEpochValidators(stateDB, args)

	case "lockStake":
		result, gasUsed, err = handleLockStake(stateDB, args)

	case "relockStake":
		result, gasUsed, err = handleRelockStake(stateDB, args)

	case "unlockStake":
		result, gasUsed, err = handleUnlockStake(stateDB, args)

	case "initialize":
		result, gasUsed, err = handleInitialize(stateDB, args)

	case "setGenesisValidator":
		result, gasUsed, err = handleSetGenesisValidator(stateDB, args)

	case "setGenesisDelegation":
		result, gasUsed, err = handleSetGenesisDelegation(stateDB, args)

	case "sumRewards":
		result, gasUsed, err = handleSumRewards(stateDB, args)

	// Private function handlers
	case "_delegate":
		result, gasUsed, err = handle_delegate(stateDB, args)

	case "_sealEpoch_offline":
		result, gasUsed, err = handle_sealEpoch_offline(stateDB, args)

	case "_sealEpoch_rewards":
		result, gasUsed, err = handle_sealEpoch_rewards(stateDB, args)

	case "_sealEpoch_minGasPrice":
		result, gasUsed, err = handle_sealEpoch_minGasPrice(stateDB, args)

	case "_calcRawValidatorEpochBaseReward":
		result, gasUsed, err = handle_calcRawValidatorEpochBaseReward(stateDB, args)

	case "_calcRawValidatorEpochTxReward":
		result, gasUsed, err = handle_calcRawValidatorEpochTxReward(stateDB, args)

	case "_calcValidatorCommission":
		result, gasUsed, err = handle_calcValidatorCommission(stateDB, args)

	case "_mintNativeToken":
		result, gasUsed, err = handle_mintNativeToken(stateDB, args)

	case "_scaleLockupReward":
		result, gasUsed, err = handle_scaleLockupReward(stateDB, args)

	case "_setValidatorDeactivated":
		result, gasUsed, err = handle_setValidatorDeactivated(stateDB, args)

	case "_syncValidator":
		result, gasUsed, err = handle_syncValidator(stateDB, args)

	case "_validatorExists":
		result, gasUsed, err = handle_validatorExists(stateDB, args)

	case "_now":
		result, gasUsed, err = handle_now(stateDB, args)

	case "getSlashingPenalty":
		result, gasUsed, err = handleGetSlashingPenalty(stateDB, args)

	// Fallback function
	case "":
		result, gasUsed, err = handleFallback(stateDB, args)

	default:
		return nil, 0, vm.ErrSfcFunctionNotImplemented
	}
	if err != nil {
		return nil, 0, vm.ErrExecutionReverted
	}

	if suppliedGas < gasUsed {
		return nil, 0, vm.ErrOutOfGas
	}
	return result, gasUsed, nil
}
