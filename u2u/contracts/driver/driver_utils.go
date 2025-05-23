package driver

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// Gas costs and limits
const (
	defaultGasLimit uint64 = 3000000 // Default gas limit for contract calls

	SloadGasCost  uint64 = 2100  // Cost of SLOAD (GetState) operation (ColdSloadCostEIP2929)
	SstoreGasCost uint64 = 20000 // Cost of SSTORE (SetState) operation (SstoreSetGasEIP2200)
	HashGasCost   uint64 = 30    // Cost of hash operation (Keccak256)
)

// checkOnlyBackend checks if the caller is the backend of the contract
// Returns nil if the caller is the backend, otherwise returns an ABI-encoded revert reason
func checkOnlyBackend(evm *vm.EVM, caller common.Address, methodName string) ([]byte, error) {
	backend := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(backendSlot)))
	backendAddr := common.BytesToAddress(backend.Bytes())
	if caller.Cmp(backendAddr) != 0 {
		// Return ABI-encoded revert reason: "caller is not the backend"
		revertReason := "caller is not the backend"
		revertData, err := DriverAbi.Methods[methodName].Outputs.Pack()
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		// Prepend the error signature: bytes4(keccak256("Error(string)"))
		errorSig := []byte{0x08, 0xc3, 0x79, 0xa0}
		// Pack the revert reason
		packedReason, err := abi.Arguments{{Type: abi.Type{T: abi.StringTy}}}.Pack(revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		// Combine the error signature and packed reason
		revertData = append(errorSig, packedReason...)
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}

// checkOnlyNode checks if the caller is address(0) (the node)
// Returns nil if the caller is address(0), otherwise returns an ABI-encoded revert reason
func checkOnlyNode(evm *vm.EVM, caller common.Address, methodName string) ([]byte, error) {
	emptyAddr := common.Address{}
	if caller.Cmp(emptyAddr) != 0 {
		// Return ABI-encoded revert reason: "caller is not the node"
		revertReason := "caller is not the node"
		revertData, err := DriverAbi.Methods[methodName].Outputs.Pack()
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		// Prepend the error signature: bytes4(keccak256("Error(string)"))
		errorSig := []byte{0x08, 0xc3, 0x79, 0xa0}
		// Pack the revert reason
		packedReason, err := abi.Arguments{{Type: abi.Type{T: abi.StringTy}}}.Pack(revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		// Combine the error signature and packed reason
		revertData = append(errorSig, packedReason...)
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}

// encodeRevertReason encodes a revert reason as an ABI-encoded error
func encodeRevertReason(methodName string, reason string) ([]byte, error) {
	// Prepend the error signature: bytes4(keccak256("Error(string)"))
	errorSig := []byte{0x08, 0xc3, 0x79, 0xa0}
	// Pack the revert reason
	packedReason, err := abi.Arguments{{Type: abi.Type{T: abi.StringTy}}}.Pack(reason)
	if err != nil {
		return nil, err
	}
	// Combine the error signature and packed reason
	revertData := append(errorSig, packedReason...)
	return revertData, nil
}
