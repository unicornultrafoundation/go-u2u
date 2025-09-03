package driver

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// handleBackend returns the backend address
func handleBackend(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(backendSlot)))
	return val.Bytes(), 0, nil
}

// handleEvmWriter returns the EVMWriter address
func handleEvmWriter(evm *vm.EVM) ([]byte, uint64, error) {
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(evmWriterSlot)))
	result, err := DriverAbi.Methods["evmWriter"].Outputs.Pack(common.BytesToAddress(val.Bytes()))
	return result, 0, err
}
