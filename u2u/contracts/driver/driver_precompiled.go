package driver

import (
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// DriverPrecompile implements PrecompiledStateContract interface
type DriverPrecompile struct{}

// Run runs the precompiled contract
func (p *DriverPrecompile) Run(stateDB vm.StateDB, blockCtx vm.BlockContext, txCtx vm.TxContext, caller common.Address, input []byte, suppliedGas uint64) ([]byte, uint64, error) {
	return nil, 0, nil
}
