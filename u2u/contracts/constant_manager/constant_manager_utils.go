package constant_manager

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// checkOnlyOwner checks if the caller is the owner of the contract
// Returns nil if the caller is the owner, otherwise returns an ABI-encoded revert reason
func checkOnlyOwner(evm *vm.EVM, methodName string) ([]byte, error) {
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	if evm.TxContext.Origin.Cmp(common.BytesToAddress(owner.Bytes())) != 0 {
		// Return ABI-encoded revert reason: "Ownable: caller is not the owner"
		revertReason := "Ownable: caller is not the owner"
		revertData, err := ConstantManagerAbi.Methods[methodName].Outputs.Pack()
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
