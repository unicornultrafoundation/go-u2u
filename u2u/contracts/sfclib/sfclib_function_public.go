package sfclib

import "github.com/unicornultrafoundation/go-u2u/core/vm"

// handleFallback is the fallback function of the SFCLib contract
func handleFallback(evm *vm.EVM, args []interface{}, input []byte) ([]byte, uint64, error) {
	return nil, 0, vm.ErrSfcFunctionNotImplemented
}
