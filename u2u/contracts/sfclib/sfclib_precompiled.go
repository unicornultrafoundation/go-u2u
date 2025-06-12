package sfclib

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

// SfcLibPrecompile implements PrecompiledSfcContract interface
type SfcLibPrecompile struct{}

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
	method, err := SfcLibAbi.MethodById(methodID)
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
func (p *SfcLibPrecompile) Run(evm *vm.EVM, caller common.Address, input []byte, suppliedGas uint64, value *big.Int) ([]byte, uint64, error) {
	// We'll use evm.SfcStateDB directly in the handler functions
	// Parse the input to get method and arguments
	method, args, err := parseABIInput(input)
	if err != nil {
		log.Error("SfcLibPrecompile.Run: Error parsing input", "err", err)
		return nil, 0, err
	}

	var (
		result  []byte
		gasUsed uint64
	)

	log.Info("SFCLib Precompiled: Calling function", "function", method.Name,
		"caller", caller.Hex(), "input", common.Bytes2Hex(input))

	switch method.Name {
	// Fallback function
	case "":
		result, gasUsed, err = handleFallback(evm, args, input)

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
