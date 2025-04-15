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
	// TODO(trinhdn97): dynamic gas needed for each calls with different methods
	var gasUsed = uint64(3000) // Example fixed gas cost

	// Parse the input to get method and arguments
	method, args, err := parseABIInput(input)
	if err != nil {
		return nil, 0, err
	}

	var result []byte
	switch method.Name {
	case "owner":
		result, err = handleOwner(stateDB)

	case "currentSealedEpoch":
		result, err = handleCurrentSealedEpoch(stateDB)

	case "lastValidatorID":
		result, err = handleLastValidatorID(stateDB)

	case "totalStake":
		result, err = handleTotalStake(stateDB)

	case "totalActiveStake":
		result, err = handleTotalActiveStake(stateDB)

	case "totalSlashedStake":
		result, err = handleTotalSlashedStake(stateDB)

	case "totalSupply":
		result, err = handleTotalSupply(stateDB)

	case "stakeTokenizerAddress":
		result, err = handleStakeTokenizerAddress(stateDB)

	case "minGasPrice":
		result, err = handleMinGasPrice(stateDB)

	case "treasuryAddress":
		result, err = handleTreasuryAddress(stateDB)

	case "voteBookAddress":
		result, err = handleVoteBookAddress(stateDB)

	case "getValidator":
		result, err = handleGetValidator(stateDB, args)
		if err != nil {
			return nil, 0, err
		}

	case "getValidatorID":
		result, err = handleGetValidatorID(stateDB, args)
		if err != nil {
			return nil, 0, err
		}

	case "getValidatorPubkey":
		result, err = handleGetValidatorPubkey(stateDB, args)
		if err != nil {
			return nil, 0, err
		}

	case "stashedRewardsUntilEpoch":
		result, err = handleStashedRewardsUntilEpoch(stateDB, args)
		if err != nil {
			return nil, 0, err
		}

	case "getWithdrawalRequest":
		result, err = handleGetWithdrawalRequest(stateDB, args)
		if err != nil {
			return nil, 0, err
		}

	case "getStake":
		result, err = handleGetStake(stateDB, args)
		if err != nil {
			return nil, 0, err
		}

	case "getLockupInfo":
		result, err = handleGetLockupInfo(stateDB, args)
		if err != nil {
			return nil, 0, err
		}

	case "getStashedLockupRewards":
		result, err = handleGetStashedLockupRewards(stateDB, args)
		if err != nil {
			return nil, 0, err
		}

	case "slashingRefundRatio":
		result, err = handleSlashingRefundRatio(stateDB, args)
		if err != nil {
			return nil, 0, err
		}

	case "getEpochSnapshot":
		result, err = handleGetEpochSnapshot(stateDB, args)
		if err != nil {
			return nil, 0, err
		}

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
