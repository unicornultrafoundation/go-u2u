package sfc

import (
	"math/big"
	"strings"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/gossip/contract/sfc100"
	"github.com/unicornultrafoundation/go-u2u/gossip/contract/sfclib100"
	"github.com/unicornultrafoundation/go-u2u/log"
)

var (
	SfcAbi            abi.ABI
	SfcLibAbi         abi.ABI
	CMAbi             abi.ABI
	NodeDriverAbi     abi.ABI
	NodeDriverAuthAbi abi.ABI
	StakeTokenizerAbi abi.ABI
)

func init() {
	SfcAbi, _ = abi.JSON(strings.NewReader(sfc100.ContractMetaData.ABI))
	SfcLibAbi, _ = abi.JSON(strings.NewReader(sfclib100.ContractMetaData.ABI))
	CMAbi, _ = abi.JSON(strings.NewReader(ConstantManagerABIStr))
	NodeDriverAbi, _ = abi.JSON(strings.NewReader(NodeDriverABIStr))
	NodeDriverAuthAbi, _ = abi.JSON(strings.NewReader(NodeDriverAuthABIStr))
	StakeTokenizerAbi, _ = abi.JSON(strings.NewReader(StakeTokenizerABIStr))
}

// SfcPrecompile implements PrecompiledSfcContract interface
type SfcPrecompile struct{}

// parseABIInput parses the input data and returns the method and unpacked parameters
func parseABIInput(smcAbi abi.ABI, input []byte) (*abi.Method, []interface{}, error) {
	// Need at least 4 bytes for function signature
	if len(input) < 4 {
		return nil, nil, vm.ErrExecutionReverted
	}

	// Get function signature from first 4 bytes
	methodID := input[:4]
	method, err := smcAbi.MethodById(methodID)
	if err != nil {
		return nil, nil, vm.ErrExecutionReverted
	}

	// Parse input arguments
	args := []interface{}{}
	if len(input) > 4 {
		args, err = method.Inputs.Unpack(input[4:])
		if err != nil {
			return nil, nil, vm.ErrExecutionReverted
		}
	}

	return method, args, nil
}

// Run runs the precompiled contract
func (p *SfcPrecompile) Run(evm *vm.EVM, caller common.Address, input []byte, suppliedGas uint64, value *big.Int) ([]byte, uint64, error) {
	// We'll use evm.SfcStateDB directly in the handler functions
	// Parse the input to get method and arguments
	method, args, err := parseABIInput(SfcAbi, input)
	if err != nil {
		log.Error("SFCPrecompile.Run: Error parsing input", "err", err)
		method = &abi.Method{
			Name:   "",
			Inputs: abi.Arguments{},
		}
	}

	var (
		result  []byte
		gasUsed uint64
	)

	log.Info("SFC Precompiled: Calling function", "function", method.Name,
		"caller", caller.Hex(), "input", common.Bytes2Hex(input))

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

	case "sealEpoch":
		result, gasUsed, err = handleSealEpoch(evm, caller, args)

	case "sealEpochValidators":
		result, gasUsed, err = handleSealEpochValidators(evm, caller, args)

	case "initialize":
		result, gasUsed, err = handleInitialize(evm, caller, args)

	case "sumRewards":
		result, gasUsed, err = handleSumRewards(evm, args)

	case "getSlashingPenalty":
		result, gasUsed, err = handleGetSlashingPenalty(evm, args)

	// Fallback function
	case "":
		result, gasUsed, err = handleFallback(evm, caller, args, input, value)

	default:
		log.Error("SFC Precompiled: Unknown function", "function", method.Name)
		return nil, 0, vm.ErrSfcFunctionNotImplemented
	}
	if err != nil {
		reason, _ := abi.UnpackRevert(result)
		log.Error("SFC Precompiled: Revert", "function", method.Name, "err", err, "reason", reason, "result", common.Bytes2Hex(result))
		return result, 0, vm.ErrExecutionReverted
	}

	if suppliedGas < gasUsed {
		log.Error("SFC Precompiled: Out of gas", "function", method.Name, "suppliedGas", suppliedGas, "gasUsed", gasUsed)
		// TODO(trinhdn97): temporarily disable gas check here to use the EVM gas for now.
		// Will re-enable this after tweaking gas cost of all handlers.
		// return result, 0, vm.ErrOutOfGas
	}

	return result, gasUsed, nil
}
