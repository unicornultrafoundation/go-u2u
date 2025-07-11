package driverauth

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// handleOwner returns the owner address
func handleOwner(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	result, err := DriverAuthAbi.Methods["owner"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
	return result, 0, err
}

// handleIsOwner checks if the caller is the owner
func handleIsOwner(evm *vm.EVM, caller common.Address) ([]byte, uint64, error) {
	owner := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	isOwner := caller == common.BytesToAddress(owner.Bytes())
	result, err := DriverAuthAbi.Methods["isOwner"].Outputs.Pack(isOwner)
	return result, 0, err
}

// handleSfc returns the SFC contract address
func handleSfc(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(sfcSlot)))
	result, err := DriverAuthAbi.Methods["sfc"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
	return result, 0, err
}

// handleDriver returns the NodeDriver contract address
func handleDriver(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(driverSlot)))
	result, err := DriverAuthAbi.Methods["driver"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
	return result, 0, err
}
