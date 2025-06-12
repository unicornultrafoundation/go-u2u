package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

func handleEpochEndTime(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}

func handleRecountVotes(evm *vm.EVM, delegator common.Address, validatorAuth common.Address, strict bool, gas *big.Int) ([]byte, uint64, error) {
	return handleInternalRecountVotes(evm, delegator, validatorAuth, strict)
}
