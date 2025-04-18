package driverauth

import (
	"strings"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/gossip/contract/driverauth100"
	"github.com/unicornultrafoundation/go-u2u/log"
)

var (
	DriverAuthAbi abi.ABI
)

func init() {
	var err error
	DriverAuthAbi, err = abi.JSON(strings.NewReader(driverauth100.ContractMetaData.ABI))
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
	method, err := DriverAuthAbi.MethodById(methodID)
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

// DriverAuthPrecompile implements PrecompiledSfcContract interface
type DriverAuthPrecompile struct{}

// Run runs the precompiled contract
func (c *DriverAuthPrecompile) Run(evm *vm.EVM, caller common.Address, input []byte, suppliedGas uint64) ([]byte, uint64, error) {
	// Parse the input to get method and arguments
	method, args, err := parseABIInput(input)
	if err != nil {
		return nil, 0, err
	}

	var result []byte
	var gasUsed uint64

	// Dispatch to the appropriate handler based on the method name
	switch method.Name {
	// Getters for variables
	case "owner":
		result, gasUsed, err = handleOwner(evm)

	case "isOwner":
		result, gasUsed, err = handleIsOwner(evm, caller)

	case "sfc":
		result, gasUsed, err = handleSfc(evm)

	case "driver":
		result, gasUsed, err = handleDriver(evm)

	// Public function handlers
	case "initialize":
		result, gasUsed, err = handleInitialize(evm, caller, args)

	case "migrateTo":
		result, gasUsed, err = handleMigrateTo(evm, caller, args)

	case "execute":
		result, gasUsed, err = handleExecute(evm, caller, args)

	case "mutExecute":
		result, gasUsed, err = handleMutExecute(evm, caller, args)

	case "incBalance":
		result, gasUsed, err = handleIncBalance(evm, caller, args)

	case "upgradeCode":
		result, gasUsed, err = handleUpgradeCode(evm, caller, args)

	case "copyCode":
		result, gasUsed, err = handleCopyCode(evm, caller, args)

	case "incNonce":
		result, gasUsed, err = handleIncNonce(evm, caller, args)

	case "updateNetworkRules":
		result, gasUsed, err = handleUpdateNetworkRules(evm, caller, args)

	case "updateMinGasPrice":
		result, gasUsed, err = handleUpdateMinGasPrice(evm, caller, args)

	case "updateNetworkVersion":
		result, gasUsed, err = handleUpdateNetworkVersion(evm, caller, args)

	case "advanceEpochs":
		result, gasUsed, err = handleAdvanceEpochs(evm, caller, args)

	case "updateValidatorWeight":
		result, gasUsed, err = handleUpdateValidatorWeight(evm, caller, args)

	case "updateValidatorPubkey":
		result, gasUsed, err = handleUpdateValidatorPubkey(evm, caller, args)

	case "setGenesisValidator":
		result, gasUsed, err = handleSetGenesisValidator(evm, caller, args)

	case "setGenesisDelegation":
		result, gasUsed, err = handleSetGenesisDelegation(evm, caller, args)

	case "deactivateValidator":
		result, gasUsed, err = handleDeactivateValidator(evm, caller, args)

	case "sealEpochValidators":
		result, gasUsed, err = handleSealEpochValidators(evm, caller, args)

	case "sealEpoch":
		result, gasUsed, err = handleSealEpoch(evm, caller, args)

	case "transferOwnership":
		result, gasUsed, err = handleTransferOwnership(evm, caller, args)

	case "renounceOwnership":
		result, gasUsed, err = handleRenounceOwnership(evm, caller)

	// Fallback function
	case "":
		result, gasUsed, err = handleFallback(evm, caller, args, input)

	default:
		log.Debug("DriverAuth Precompiled: Unknown function", "function", method.Name)
		return nil, 0, vm.ErrSfcFunctionNotImplemented
	}
	if err != nil {
		reason, _ := abi.UnpackRevert(result)
		log.Error("DriverAuth Precompiled: Revert", "function", method.Name, "reason", reason)
		return nil, 0, vm.ErrExecutionReverted
	}

	if suppliedGas < gasUsed {
		log.Error("DriverAuth Precompiled: Out of gas", "function", method.Name)
		return nil, 0, vm.ErrOutOfGas
	}
	log.Debug("DriverAuth Precompiled: Success", "function", method.Name)

	return result, gasUsed, nil
}
