package constant_manager

import (
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// handleFallback handles the fallback function (when no method is specified)
func handleFallback(evm *vm.EVM, args []interface{}, input []byte) ([]byte, uint64, error) {
	// The ConstantManager contract doesn't have a fallback function that does anything
	// Just return empty result with no error
	return nil, 0, nil
}
