package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
)

// SfcLibPrecompile implements PrecompiledSfcContract interface
type SfcLibPrecompile struct{}

// Run runs the precompiled contract
func (p *SfcLibPrecompile) Run(evm *vm.EVM, caller common.Address, input []byte, suppliedGas uint64, value *big.Int) ([]byte, uint64, error) {
	// We'll use evm.SfcStateDB directly in the handler functions
	// Parse the input to get method and arguments
	method, args, err := parseABIInput(SfcLibAbi, input)
	if err != nil {
		log.Error("SfcLibPrecompile.Run: Error parsing input", "err", err)
		method = &abi.Method{
			Name:   "",
			Inputs: abi.Arguments{},
		}
	}

	var (
		result  []byte
		gasUsed uint64
	)

	log.Info("SFCLib Precompiled: Calling function", "function", method.Name,
		"caller", caller.Hex(), "input", common.Bytes2Hex(input))

	switch method.Name {
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

	case "recountVotes":
		if len(args) != 4 {
			return nil, 0, vm.ErrExecutionReverted
		}
		delegator, ok := args[0].(common.Address)
		if !ok {
			return nil, 0, vm.ErrExecutionReverted
		}
		validatorAuth, ok := args[1].(common.Address)
		if !ok {
			return nil, 0, vm.ErrExecutionReverted
		}
		strict, ok := args[2].(bool)
		if !ok {
			return nil, 0, vm.ErrExecutionReverted
		}
		gas, ok := args[3].(*big.Int)
		if !ok {
			return nil, 0, vm.ErrExecutionReverted
		}
		result, gasUsed, err = handleRecountVotes(evm, delegator, validatorAuth, strict, gas)

	case "epochEndTime":
		result, gasUsed, err = handleEpochEndTime(evm, args)

	case "createValidator":
		result, gasUsed, err = handleCreateValidator(evm, caller, args, value)

	case "delegate":
		result, gasUsed, err = handleDelegate(evm, caller, args, value)

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
		result, gasUsed, err = handleMintU2U(evm, caller, args)

	case "burnU2U":
		result, gasUsed, err = handleBurnU2U(evm, args)

	case "lockStake":
		result, gasUsed, err = handleLockStake(evm, caller, args)

	case "relockStake":
		result, gasUsed, err = handleRelockStake(evm, caller, args)

	case "unlockStake":
		result, gasUsed, err = handleUnlockStake(evm, caller, args)

	case "setGenesisValidator":
		result, gasUsed, err = handleSetGenesisValidator(evm, args)

	case "setGenesisDelegation":
		result, gasUsed, err = handleSetGenesisDelegation(evm, args)

	default:
		log.Error("SFCLib Precompiled: Unknown function", "function", method.Name)
		return nil, 0, vm.ErrSfcFunctionNotImplemented
	}
	if err != nil {
		reason, _ := abi.UnpackRevert(result)
		log.Error("SFCLib Precompiled: Revert", "function", method.Name, "err", err, "reason", reason, "result", common.Bytes2Hex(result))
		return result, 0, vm.ErrExecutionReverted
	}

	if suppliedGas < gasUsed {
		log.Error("SFCLib Precompiled: Out of gas", "function", method.Name, "suppliedGas", suppliedGas, "gasUsed", gasUsed)
		// TODO(trinhdn97): temporarily disable gas check here to use the EVM gas for now.
		// Will re-enable this after tweaking gas cost of all handlers.
		// return result, 0, vm.ErrOutOfGas
	}

	return result, gasUsed, nil
}
