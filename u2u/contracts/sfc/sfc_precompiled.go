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

// SfcPrecompile implements PrecompiledStateContract interface
type SfcPrecompile struct{}

// parseABIInput parses the input data and returns the method and unpacked parameters
func parseABIInput(input []byte) (*abi.Method, []interface{}, error) {
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
func (p *SfcPrecompile) Run(stateDB vm.StateDB, blockCtx vm.BlockContext, txCtx vm.TxContext, caller common.Address,
	input []byte, suppliedGas uint64) ([]byte, uint64, error) {
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
