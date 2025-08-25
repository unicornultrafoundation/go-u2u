package driverauth

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
)

// Gas costs and limits
const (
	defaultGasLimit uint64 = 7000000 // Default gas limit for contract calls
)

// checkOnlyOwner checks if the caller is the owner of the contract
// Returns nil if the caller is the owner, otherwise returns an ABI-encoded revert reason
func checkOnlyOwner(evm *vm.EVM, caller common.Address, methodName string) ([]byte, error) {
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	ownerAddr := common.BytesToAddress(owner.Bytes())
	if caller.Cmp(ownerAddr) != 0 {
		// Return ABI-encoded revert reason: "Ownable: caller is not the owner"
		revertReason := "Ownable: caller is not the owner"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}

// checkOnlySFC checks if the caller is the SFC contract
// Returns nil if the caller is the SFC, otherwise returns an ABI-encoded revert reason
func checkOnlySFC(evm *vm.EVM, caller common.Address, methodName string) ([]byte, error) {
	sfc := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot)))
	sfcAddr := common.BytesToAddress(sfc.Bytes())
	if caller.Cmp(sfcAddr) != 0 {
		// Return ABI-encoded revert reason: "caller is not the SFC contract"
		revertReason := "caller is not the SFC contract"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}

// checkOnlyDriver checks if the caller is the Driver contract
// Returns nil if the caller is the Driver, otherwise returns an ABI-encoded revert reason
func checkOnlyDriver(evm *vm.EVM, caller common.Address, methodName string) ([]byte, error) {
	driver := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot)))
	driverAddr := common.BytesToAddress(driver.Bytes())
	if caller.Cmp(driverAddr) != 0 {
		// Return ABI-encoded revert reason: "caller is not the Driver contract"
		revertReason := "caller is not the Driver contract"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			log.Error("DriverAuth Precompiled: Failed to encode revert reason", "err", err)
			return nil, vm.ErrExecutionReverted
		}
		log.Error("DriverAuth Precompiled: Caller is not the driver contract", "caller", caller, "driver", driverAddr)
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

// checkZeroAddress checks if an address is the zero address
// Returns nil if the address is not zero, otherwise returns an ABI-encoded revert reason
func checkZeroAddress(addr common.Address, methodName string, message string) ([]byte, error) {
	emptyAddr := common.Address{}
	if addr.Cmp(emptyAddr) == 0 {
		// Return ABI-encoded revert reason with the provided message
		revertData, err := encodeRevertReason(methodName, message)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}

// checkAlreadyInitialized checks if the contract is already initialized
// Returns nil if the contract is not initialized, otherwise returns an ABI-encoded revert reason
func checkAlreadyInitialized(evm *vm.EVM, methodName string) ([]byte, error) {
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	sfc := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot)))
	driver := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot)))

	emptyHash := common.Hash{}
	if owner.Cmp(emptyHash) != 0 || sfc.Cmp(emptyHash) != 0 || driver.Cmp(emptyHash) != 0 {
		// Return ABI-encoded revert reason: "already initialized"
		revertReason := "already initialized"
		revertData, err := encodeRevertReason(methodName, revertReason)
		if err != nil {
			return nil, vm.ErrExecutionReverted
		}
		return revertData, vm.ErrExecutionReverted
	}
	return nil, nil
}
