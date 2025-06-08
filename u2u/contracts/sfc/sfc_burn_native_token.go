package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
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
			log.Error("handleBurnU2U: pack BurntU2U failed", "err", err, "amount", amount)
			return nil, 0, err
		}
		burntU2ULog := &types.Log{
			Address: ContractAddress,
			Topics:  topics,
			Data:    data,
		}
		evm.SfcStateDB.AddLog(burntU2ULog)
		log.Info("handleBurnU2U: BurntU2U event", "Address", burntU2ULog.Address.Hex(),
			"Topics", burntU2ULog.Topics, "Data", common.Bytes2Hex(burntU2ULog.Data))
	}

	return nil, gasUsed, nil
}
