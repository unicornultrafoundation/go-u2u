// Copyright 2021 The go-ethereum Authors
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

package gossip

import (
	"errors"
	"fmt"
	"time"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/evmcore"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/trie"
)

// stateAtBlock retrieves the state database associated with a certain block.
// If no state is locally available for the given block, a number of blocks
// are attempted to be reexecuted to generate the desired state. The optional
// base layer statedb can be passed then it's regarded as the statedb of the
// parent block.
func (eth *Service) stateAtBlock(evmblock *evmcore.EvmBlock, reexec uint64, base *state.StateDB,
	checkLive bool) (statedb *state.StateDB, sfcStatedb *state.StateDB, err error) {
	block := evmblock.EthBlock()
	var (
		current  *evmcore.EvmBlock
		database state.Database
		report   = true
		origin   = block.NumberU64()
	)
	// Check the live database first if we have the state fully available, use that.
	if checkLive {
		statedb, err = eth.EthAPI.state.StateAt(block.Root())
		if err == nil {
			if sfcStatedb, err = eth.EthAPI.state.SfcStateAt(block.SfcStateRoot()); err != nil {
				log.Warn("Failed to get SFC state", "sfcStateRoot", block.SfcStateRoot())
			}
			return statedb, sfcStatedb, nil
		}
		// TODO(trinhdn97): implement diff layer for SFC state and re-execute SFC state at block as well
	}
	if base != nil {
		// The optional base statedb is given, mark the start point as parent block
		statedb, database, report = base, base.Database(), false
		current = eth.EthAPI.state.GetBlock(block.ParentHash(), block.NumberU64()-1)
	} else {
		// Otherwise try to reexec blocks until we find a state or reach our limit
		current = evmblock

		// Create an ephemeral trie.Database for isolating the live one. Otherwise
		// the internal junks created by tracing will be persisted into the disk.
		database = state.NewDatabaseWithConfig(eth.EthAPI.ChainDb(), &trie.Config{Cache: 16})

		// If we didn't check the dirty database, do check the clean one, otherwise
		// we would rewind past a persisted block (specific corner case is chain
		// tracing from the genesis).
		if !checkLive {
			statedb, err = state.New(current.Root, database, nil)
			if err == nil {
				return statedb, nil, nil
			}
		}
		// Database does not have the state for the given block, try to regenerate
		for i := uint64(0); i < reexec; i++ {
			if current.NumberU64() == 0 {
				return nil, nil, errors.New("genesis state is missing")
			}
			parent := eth.EthAPI.state.GetBlock(current.ParentHash, current.NumberU64()-1)
			if parent == nil {
				return nil, nil, fmt.Errorf("missing block %v %d", current.ParentHash, current.NumberU64()-1)
			}
			current = parent

			statedb, err = state.New(current.Root, database, nil)
			if err == nil {
				break
			}
		}
		if err != nil {
			switch err.(type) {
			case *trie.MissingNodeError:
				return nil, nil, fmt.Errorf("required historical state unavailable (reexec=%d)", reexec)
			default:
				return nil, nil, err
			}
		}
	}
	// State was available at historical point, regenerate
	var (
		start  = time.Now()
		logged time.Time
		parent common.Hash
	)
	for current.NumberU64() < origin {
		// Print progress logs if long enough time elapsed
		if time.Since(logged) > 8*time.Second && report {
			log.Info("Regenerating historical state", "block", current.NumberU64()+1, "target", origin, "remaining", origin-current.NumberU64()-1, "elapsed", time.Since(start))
			logged = time.Now()
		}
		// Retrieve the next block to regenerate and process it
		next := current.NumberU64() + 1
		if current = eth.EthAPI.state.GetBlock(common.Hash{}, next); current == nil {
			return nil, nil, fmt.Errorf("block #%d not found", next)
		}
		evmProcessor := evmcore.NewStateProcessor(eth.EthAPI.ChainConfig(), eth.EthAPI.state)
		var gasUsed uint64 = 0
		_, _, _, err := evmProcessor.Process(current, statedb, sfcStatedb, vm.Config{}, &gasUsed, func(l *types.Log, _ *state.StateDB) {})
		if err != nil {
			return nil, nil, fmt.Errorf("processing block %d failed: %v", current.NumberU64(), err)
		}
		// Finalize the state so any modifications are written to the trie
		root, err := statedb.Commit(eth.EthAPI.ChainConfig().IsEIP158(current.Number))
		if err != nil {
			return nil, nil, err
		}
		statedb, err = state.New(root, database, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("state reset after block %d failed: %v", current.NumberU64(), err)
		}
		database.TrieDB().Reference(root, common.Hash{})
		if parent != (common.Hash{}) {
			database.TrieDB().Dereference(parent)
		}
		parent = root
	}
	if report {
		nodes, imgs := database.TrieDB().Size()
		log.Info("Historical state regenerated", "block", current.NumberU64(), "elapsed", time.Since(start), "nodes", nodes, "preimages", imgs)
	}
	// TODO(trinhdn97): return the SFC state at this block as well
	return statedb, sfcStatedb, nil
}

// stateAtTransaction returns the execution environment of a certain transaction.
func (eth *Service) stateAtTransaction(evmblock *evmcore.EvmBlock, txIndex int, reexec uint64) (evmcore.Message, vm.BlockContext, *state.StateDB, error) {
	block := evmblock.EthBlock()
	// Short circuit if it's the genesis block.
	if block.NumberU64() == 0 {
		return nil, vm.BlockContext{}, nil, errors.New("no transaction in genesis")
	}
	// Create the parent state database
	parent := eth.EthAPI.state.GetBlock(block.ParentHash(), block.NumberU64()-1)
	if parent == nil {
		return nil, vm.BlockContext{}, nil, fmt.Errorf("parent %#x not found", block.ParentHash())
	}
	// Lookup the statedb of parent block from the live database,
	// otherwise regenerate it on the flight.
	statedb, sfcStatedb, err := eth.stateAtBlock(parent, reexec, nil, true)
	if err != nil {
		return nil, vm.BlockContext{}, nil, err
	}
	if txIndex == 0 && len(block.Transactions()) == 0 {
		return nil, vm.BlockContext{}, statedb, nil
	}
	// Recompute transactions up to the target index.
	signer := types.MakeSigner(eth.EthAPI.ChainConfig(), block.Number())
	for idx, tx := range block.Transactions() {
		// Assemble the transaction call message and return if the requested offset
		msg, _ := tx.AsMessage(signer, block.BaseFee())
		txContext := evmcore.NewEVMTxContext(msg)
		context := evmcore.NewEVMBlockContext(evmcore.ConvertFromEthHeader(block.Header()), eth.EthAPI.state, nil)
		if idx == txIndex {
			return msg, context, statedb, nil
		}
		// Not yet the searched for transaction, execute on top of the current state
		vmenv := vm.NewEVM(context, txContext, statedb, sfcStatedb, eth.EthAPI.ChainConfig(), vm.Config{})
		statedb.Prepare(tx.Hash(), idx)
		if _, err := evmcore.ApplyMessage(vmenv, msg, new(evmcore.GasPool).AddGas(tx.Gas())); err != nil {
			return nil, vm.BlockContext{}, nil, fmt.Errorf("transaction %#x failed: %v", tx.Hash(), err)
		}
		// Ensure any modifications are committed to the state
		// Only delete empty objects if EIP158/161 (a.k.a Spurious Dragon) is in effect
		statedb.Finalise(vmenv.ChainConfig().IsEIP158(block.Number()))
	}
	return nil, vm.BlockContext{}, nil, fmt.Errorf("transaction index %d out of range for block %#x", txIndex, block.Hash())
}
