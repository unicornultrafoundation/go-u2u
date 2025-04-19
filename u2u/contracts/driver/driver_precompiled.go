package driver

import (
	"strings"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/gossip/contract/driver100"
	"github.com/unicornultrafoundation/go-u2u/log"
)

var (
	DriverAbi     abi.ABI
	SfcAbi        abi.ABI
	DriverAuthAbi abi.ABI
	EvmWriterAbi  abi.ABI
)

func init() {
	var err error
	DriverAbi, err = abi.JSON(strings.NewReader(driver100.ContractMetaData.ABI))
	if err != nil {
		panic(err)
	}
	SfcAbi, _ = abi.JSON(strings.NewReader(SfcAbiStr))
	DriverAuthAbi, _ = abi.JSON(strings.NewReader(DriverAuthAbiStr))
	EvmWriterAbi, _ = abi.JSON(strings.NewReader(EvmWriterABIStr))
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
	method, err := DriverAbi.MethodById(methodID)
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

// DriverPrecompile implements PrecompiledSfcContract interface
type DriverPrecompile struct{}

// Run runs the precompiled contract
func (p *DriverPrecompile) Run(evm *vm.EVM, caller common.Address, input []byte, suppliedGas uint64) ([]byte, uint64, error) {
	// Parse the input to get method and arguments
	method, args, err := parseABIInput(input)
	if err != nil {
		return nil, 0, err
	}

	var (
		result  []byte
		gasUsed uint64
	)

	log.Info("Driver Precompiled: Calling function", "function", method.Name, "caller", caller.Hex(), "args", args)

	// Dispatch to the appropriate handler based on the method name
	switch method.Name {
	// Getters for variables
	case "backend":
		result, gasUsed, err = handleBackend(evm)
	case "evmWriter":
		result, gasUsed, err = handleEvmWriter(evm)

	// Public function handlers
	case "setBackend":
		result, gasUsed, err = handleSetBackend(evm, caller, args)
	case "initialize":
		result, gasUsed, err = handleInitialize(evm, caller, args)
	case "setBalance":
		result, gasUsed, err = handleSetBalance(evm, caller, args)
	case "copyCode":
		result, gasUsed, err = handleCopyCode(evm, caller, args)
	case "swapCode":
		result, gasUsed, err = handleSwapCode(evm, caller, args)
	case "setStorage":
		result, gasUsed, err = handleSetStorage(evm, caller, args)
	case "incNonce":
		result, gasUsed, err = handleIncNonce(evm, caller, args)
	case "updateNetworkRules":
		result, gasUsed, err = handleUpdateNetworkRules(evm, caller, args)
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
	case "sealEpochV1":
		result, gasUsed, err = handleSealEpochV1(evm, caller, args)

	// Fallback function
	case "":
		result, gasUsed, err = handleFallback(evm, caller, args, input)

	default:
		log.Debug("Driver Precompiled: Unknown function", "function", method.Name)
		return nil, 0, vm.ErrSfcFunctionNotImplemented
	}
	if err != nil {
		reason, _ := abi.UnpackRevert(result)
		log.Error("Driver Precompiled: Revert", "function", method.Name, "err", err, "reason", reason)
		return nil, 0, vm.ErrExecutionReverted
	}

	if suppliedGas < gasUsed {
		log.Error("Driver Precompiled: Out of gas", "function", method.Name)
		return nil, 0, vm.ErrOutOfGas
	}
	log.Debug("Driver Precompiled: Success", "function", method.Name)

	return result, gasUsed, nil
}
