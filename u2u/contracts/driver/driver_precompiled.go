package driver

import (
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// DriverPrecompile implements PrecompiledSfcContract interface
type DriverPrecompile struct{}

// Run runs the precompiled contract
func (p *DriverPrecompile) Run(evm *vm.EVM, caller common.Address, input []byte, suppliedGas uint64) ([]byte, uint64, error) {
	return nil, 0, nil
}
