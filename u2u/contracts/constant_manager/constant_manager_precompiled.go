package constant_manager

import (
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// ConstantManagerPrecompile implements PrecompiledContract interface
type ConstantManagerPrecompile struct{}

// Run runs the precompiled contract
func (c *ConstantManagerPrecompile) Run(stateDB vm.StateDB, blockCtx vm.BlockContext, txCtx vm.TxContext, caller common.Address, input []byte, suppliedGas uint64) ([]byte, uint64, error) {
	return nil, 0, nil
}
