package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// handleBurnU2U implements the _burnU2U function from SFCLib.sol
func handleBurnU2U(evm *vm.EVM, args []interface{}) ([]byte, uint64, error) {
	var gasUsed uint64 = 0

	// Parse arguments
	if len(args) != 1 {
		return nil, 0, vm.ErrExecutionReverted
	}
	amount, ok := args[0].(*big.Int)
	if !ok {
		return nil, 0, vm.ErrExecutionReverted
	}

	// Only burn if amount > 0
	if amount.Cmp(big.NewInt(0)) > 0 {
		// Transfer to zero address (burn)
		evm.SfcStateDB.SubBalance(ContractAddress, amount)

		// Emit BurntU2U event
		topics := []common.Hash{
			SfcLibAbi.Events["BurntU2U"].ID,
		}
		data, err := SfcLibAbi.Events["BurntU2U"].Inputs.NonIndexed().Pack(
			amount,
		)
		if err != nil {
			return nil, 0, err
		}
		evm.SfcStateDB.AddLog(&types.Log{
			Address: ContractAddress,
			Topics:  topics,
			Data:    data,
		})
	}

	return nil, gasUsed, nil
}
