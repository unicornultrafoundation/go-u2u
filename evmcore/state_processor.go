// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package evmcore

import (
	"fmt"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/evmcore/txtracer"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/params"
	"github.com/unicornultrafoundation/go-u2u/utils/signers/gsignercache"
	"github.com/unicornultrafoundation/go-u2u/utils/signers/internaltx"
)

var SfcPrecompiles = []common.Address{
	common.HexToAddress("0xFC00FACE00000000000000000000000000000000"),
	common.HexToAddress("0xD100ae0000000000000000000000000000000000"),
	common.HexToAddress("0xd100A01E00000000000000000000000000000000"),
	common.HexToAddress("0x6CA548f6DF5B540E72262E935b6Fe3e72cDd68C9"),
	common.HexToAddress("0xFC01fACE00000000000000000000000000000000"), // SFCLib
}

// StateProcessor is a basic Processor, which takes care of
// the state transitioning from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *params.ChainConfig // Chain configuration options
	bc     DummyChain          // Canonical blockchain
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc DummyChain) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
	}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the state db and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas, it will return an error.
func (p *StateProcessor) Process(
	block *EvmBlock, statedb *state.StateDB, sfcStatedb *state.StateDB, cfg vm.Config,
	usedGas *uint64, onNewLog func(*types.Log, *state.StateDB),
) (
	receipts types.Receipts, allLogs []*types.Log, skipped []uint32, err error,
) {
	skipped = make([]uint32, 0, len(block.Transactions))
	var (
		gp           = new(GasPool).AddGas(block.GasLimit)
		receipt      *types.Receipt
		skip         bool
		header       = block.Header()
		blockContext = NewEVMBlockContext(header, p.bc, nil)
		vmenv        = vm.NewEVM(blockContext, vm.TxContext{}, statedb, sfcStatedb, p.config, cfg)
		blockHash    = block.Hash
		blockNumber  = block.Number
		signer       = gsignercache.Wrap(types.MakeSigner(p.config, header.Number))
	)
	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions {
		msg, err := TxAsMessage(tx, signer, header.BaseFee)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("could not apply tx %d [%v]: %w", i, tx.Hash().Hex(), err)
		}

		statedb.Prepare(tx.Hash(), i)
		if sfcStatedb != nil {
			sfcStatedb.Prepare(tx.Hash(), i)
		}
		log.Info("StateProcessor.Process before", "tx", tx.Hash().Hex())
		receipt, _, skip, err = ApplyTransaction(msg, p.config, gp, statedb, sfcStatedb, blockNumber, blockHash, tx, usedGas, vmenv, cfg, onNewLog)
		log.Info("StateProcessor.Process after", "tx", tx.Hash().Hex())
		if skip {
			skipped = append(skipped, uint32(i))
			err = nil
			continue
		}
		if err != nil {
			return nil, nil, nil, fmt.Errorf("could not apply tx %d [%v]: %w", i, tx.Hash().Hex(), err)
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)

		// extra dual-state verification and benchmark
		if sfcStatedb != nil {
			for _, addr := range SfcPrecompiles {
				original := statedb.GetStorageRoot(addr)
				sfc := sfcStatedb.GetStorageRoot(addr)
				if original.Cmp(sfc) != 0 {
					log.Error("U2UEVMProcessor.Process: SFC storage corrupted after applying tx",
						"tx", tx.Hash().Hex(), "addr", addr, "original", original.Hex(), "sfc", sfc.Hex())
					common.SendInterrupt()
				}
				originalBalance := statedb.GetBalance(addr)
				sfcBalance := sfcStatedb.GetBalance(addr)
				if originalBalance.Cmp(sfcBalance) != 0 {
					log.Error("U2UEVMProcessor.Process: SFC balance mismatched after applying tx",
						"tx", tx.Hash().Hex(), "addr", addr, "original", originalBalance, "sfc", sfcBalance)
					common.SendInterrupt()
				}
				originalNonce := statedb.GetNonce(addr)
				sfcNonce := sfcStatedb.GetNonce(addr)
				if originalNonce != sfcNonce {
					log.Error("U2UEVMProcessor.Process: SFC nonce mismatched after applying tx",
						"tx", tx.Hash().Hex(), "addr", addr, "original", originalNonce, "sfc", sfcNonce)
					common.SendInterrupt()
				}
			}
			// Benchmark execution time difference of SFC precompiled related txs
			if tx.To() != nil {
				if _, ok := vmenv.SfcPrecompile(*tx.To()); ok {
					// Calculate percentage difference: ((sfc - evm) / evm) * 100
					var percentDiff float64
					if vm.TotalEvmExecutionElapsed > 0 {
						percentDiff = (float64(vm.TotalSfcExecutionElapsed-vm.TotalEvmExecutionElapsed) / float64(vm.TotalEvmExecutionElapsed)) * 100
					}

					log.Info("SFC execution time comparison",
						"diff", fmt.Sprintf("%.2f%%", percentDiff),
						"evm", vm.TotalEvmExecutionElapsed,
						"sfc", vm.TotalSfcExecutionElapsed,
						"txHash", tx.Hash().Hex())
					vm.ResetSFCMetrics()
				}
			}
		}
	}
	return
}

func ApplyTransaction(
	msg types.Message,
	config *params.ChainConfig,
	gp *GasPool,
	statedb *state.StateDB,
	sfcStatedb *state.StateDB,
	blockNumber *big.Int,
	blockHash common.Hash,
	tx *types.Transaction,
	usedGas *uint64,
	evm *vm.EVM,
	cfg vm.Config,
	onNewLog func(*types.Log, *state.StateDB),
) (
	*types.Receipt,
	uint64,
	bool,
	error,
) {
	// Create a new context to be used in the EVM environment.
	txContext := NewEVMTxContext(msg)
	evm.Reset(txContext, statedb, sfcStatedb)

	// Test if type of tracer is transaction tracing
	// logger, in that case, set a info for it
	var traceLogger *txtracer.TraceStructLogger
	switch cfg.Tracer.(type) {
	case *txtracer.TraceStructLogger:
		traceLogger = cfg.Tracer.(*txtracer.TraceStructLogger)
		traceLogger.SetTx(tx.Hash())
		traceLogger.SetFrom(msg.From())
		traceLogger.SetTo(msg.To())
		traceLogger.SetValue(*msg.Value())
		traceLogger.SetBlockHash(blockHash)
		traceLogger.SetBlockNumber(blockNumber)
		traceLogger.SetTxIndex(uint(statedb.TxIndex()))
	}
	// Apply the transaction to the current state (included in the env).
	result, err := ApplyMessage(evm, msg, gp)
	if err != nil {
		return nil, 0, result == nil, err
	}
	// Notify about logs with potential state changes
	logs := statedb.GetLogs(tx.Hash(), blockHash)
	for _, l := range logs {
		onNewLog(l, statedb)
	}
	if sfcStatedb != nil {
		sfcLogs := sfcStatedb.GetLogs(tx.Hash(), blockHash)
		for i := range sfcLogs {
			// just process the EVM logs once for now
			// onNewLog(l, sfcStatedb)
			if !logs[i].Equal(sfcLogs[i]) {
				log.Error("SFC log mismatch", "index", i, "txHash", tx.Hash().Hex(),
					"evm", logs[i], "sfc", sfcLogs[i])
			}
		}
		if len(logs) != len(sfcLogs) {
			log.Error("SFC log mismatch", "txHash", tx.Hash().Hex())
			log.Error("EVM", "logs", logs)
			log.Error("SFC", "logs", sfcLogs)
		}
	}

	// Update the state with pending changes.
	var root []byte
	if config.IsByzantium(blockNumber) {
		log.Trace("StateProcessor.Process during ApplyTransaction", "txHash", tx.Hash().Hex())
		statedb.Finalise(true)
		if sfcStatedb != nil {
			log.Trace("Separate two commit logs when StateProcessor.Process during ApplyTransaction",
				"above", "evm", "below", "sfc")
			sfcStatedb.Finalise(true)
		}
	} else {
		root = statedb.IntermediateRoot(config.IsEIP158(blockNumber)).Bytes()
		if sfcStatedb != nil {
			sfcStatedb.IntermediateRoot(config.IsEIP158(blockNumber)).Bytes()
		}
	}
	*usedGas += result.UsedGas

	// Create a new receipt for the transaction, storing the intermediate root and gas used
	// by the tx.
	receipt := &types.Receipt{Type: tx.Type(), PostState: root, CumulativeGasUsed: *usedGas}
	if result.Failed() {
		receipt.Status = types.ReceiptStatusFailed
	} else {
		receipt.Status = types.ReceiptStatusSuccessful
	}
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = result.UsedGas

	// If the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(evm.TxContext.Origin, tx.Nonce())
	}

	// Set the receipt logs.
	receipt.Logs = logs
	// TODO(trinhdn97): include logs and root of SfcStateDB here
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})
	receipt.BlockHash = blockHash
	receipt.BlockNumber = blockNumber
	receipt.TransactionIndex = uint(statedb.TxIndex())

	// Set post-information and save trace
	if traceLogger != nil {
		traceLogger.SetGasUsed(result.UsedGas)
		traceLogger.SetNewAddress(receipt.ContractAddress)
		traceLogger.ProcessTx()
		traceLogger.SaveTrace()
	}

	return receipt, result.UsedGas, false, err
}

func TxAsMessage(tx *types.Transaction, signer types.Signer, baseFee *big.Int) (types.Message, error) {
	if !internaltx.IsInternal(tx) {
		return tx.AsMessage(signer, baseFee)
	} else {
		msg := types.NewMessage(internaltx.InternalSender(tx), tx.To(), tx.Nonce(), tx.Value(), tx.Gas(), tx.GasPrice(), tx.GasFeeCap(), tx.GasTipCap(), tx.Data(), tx.AccessList(), true)
		return msg, nil
	}
}
