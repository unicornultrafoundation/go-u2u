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

// Package ethapi implements the general Ethereum API functions.
package ethapi

import (
	"context"
	"math/big"
	"time"

	"github.com/unicornultrafoundation/go-helios/hash"
	"github.com/unicornultrafoundation/go-helios/native/idx"

	"github.com/unicornultrafoundation/go-u2u/accounts"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/ethdb"
	notify "github.com/unicornultrafoundation/go-u2u/event"
	"github.com/unicornultrafoundation/go-u2u/evmcore"
	"github.com/unicornultrafoundation/go-u2u/evmcore/txtracer"
	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/native/iblockproc"
	"github.com/unicornultrafoundation/go-u2u/params"
	"github.com/unicornultrafoundation/go-u2u/rpc"
)

// PeerProgress is synchronization status of a peer
type PeerProgress struct {
	CurrentEpoch     idx.Epoch
	CurrentBlock     idx.Block
	CurrentBlockHash hash.Event
	CurrentBlockTime native.Timestamp
	HighestBlock     idx.Block
	HighestEpoch     idx.Epoch
}

// Backend interface provides the common API services (that are provided by
// both full and light clients) with access to necessary functions.
type Backend interface {
	// General Ethereum API
	Progress() PeerProgress
	SuggestGasTipCap(ctx context.Context, certainty uint64) *big.Int
	EffectiveMinGasPrice(ctx context.Context) *big.Int
	ChainDb() ethdb.Database
	AccountManager() *accounts.Manager
	ExtRPCEnabled() bool
	RPCGasCap() uint64    // global gas cap for eth_call over rpc: DoS protection
	RPCTxFeeCap() float64 // global tx fee cap for all transaction related APIs
	RPCTimeout() time.Duration
	UnprotectedAllowed() bool // allows only for EIP155 transactions.
	CalcBlockExtApi() bool
	StateAtBlock(ctx context.Context, block *evmcore.EvmBlock, reexec uint64, base *state.StateDB, checkLive bool) (*state.StateDB, error)
	StateAtTransaction(ctx context.Context, block *evmcore.EvmBlock, txIndex int, reexec uint64) (evmcore.Message, vm.BlockContext, *state.StateDB, error)
	// Blockchain API
	HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*evmcore.EvmHeader, error)
	HeaderByHash(ctx context.Context, hash common.Hash) (*evmcore.EvmHeader, error)
	BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*evmcore.EvmBlock, error)
	StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *evmcore.EvmHeader, error)
	ResolveRpcBlockNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (idx.Block, error)
	BlockByHash(ctx context.Context, hash common.Hash) (*evmcore.EvmBlock, error)
	GetReceiptsByNumber(ctx context.Context, number rpc.BlockNumber) (types.Receipts, error)
	GetTd(hash common.Hash) *big.Int
	GetEVM(ctx context.Context, msg evmcore.Message, state *state.StateDB, sfcState *state.StateDB, header *evmcore.EvmHeader, vmConfig *vm.Config) (*vm.EVM, func() error, error)
	GetBlockContext(header *evmcore.EvmHeader) vm.BlockContext
	MinGasPrice() *big.Int
	MaxGasLimit() uint64

	// Transaction trace API
	TxTraceByHash(ctx context.Context, h common.Hash) (*[]txtracer.ActionTrace, error)
	TxTraceSave(ctx context.Context, h common.Hash, traces []byte) error

	// Transaction pool API
	SendTx(ctx context.Context, signedTx *types.Transaction) error
	GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, uint64, uint64, error)
	GetPoolTransactions() (types.Transactions, error)
	GetPoolTransaction(txHash common.Hash) *types.Transaction
	GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error)
	Stats() (pending int, queued int)
	TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions)
	TxPoolContentFrom(addr common.Address) (types.Transactions, types.Transactions)
	SubscribeNewTxsNotify(chan<- evmcore.NewTxsNotify) notify.Subscription

	ChainConfig() *params.ChainConfig
	CurrentBlock() *evmcore.EvmBlock

	// Helios DAG API
	GetEventPayload(ctx context.Context, shortEventID string) (*native.EventPayload, error)
	GetEvent(ctx context.Context, shortEventID string) (*native.Event, error)
	GetHeads(ctx context.Context, epoch rpc.BlockNumber) (hash.Events, error)
	CurrentEpoch(ctx context.Context) idx.Epoch
	SealedEpochTiming(ctx context.Context) (start native.Timestamp, end native.Timestamp)

	// Helios aBFT API
	GetEpochBlockState(ctx context.Context, epoch rpc.BlockNumber) (*iblockproc.BlockState, *iblockproc.EpochState, error)
	GetDowntime(ctx context.Context, vid idx.ValidatorID) (idx.Block, native.Timestamp, error)
	GetUptime(ctx context.Context, vid idx.ValidatorID) (*big.Int, error)
	GetOriginatedFee(ctx context.Context, vid idx.ValidatorID) (*big.Int, error)

	// SFC state API
	SfcStateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *evmcore.EvmHeader, error)
}

func GetAPIs(apiBackend Backend) []rpc.API {
	nonceLock := new(AddrLocker)
	orig := []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicBlockChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "dag",
			Version:   "1.0",
			Service:   NewPublicDAGChainAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicTransactionPoolAPI(apiBackend, nonceLock),
			Public:    true,
		}, {
			Namespace: "txpool",
			Version:   "1.0",
			Service:   NewPublicTxPoolAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(apiBackend),
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicAccountAPI(apiBackend.AccountManager()),
			Public:    true,
		}, {
			Namespace: "personal",
			Version:   "1.0",
			Service:   NewPrivateAccountAPI(apiBackend, nonceLock),
			Public:    false,
		}, {
			Namespace: "abft",
			Version:   "1.0",
			Service:   NewPublicAbftAPI(apiBackend),
			Public:    true,
		}, {
			Namespace: "trace",
			Version:   "1.0",
			Service:   NewPublicTxTraceAPI(apiBackend),
			Public:    true,
		},
	}

	return append(orig)
}
