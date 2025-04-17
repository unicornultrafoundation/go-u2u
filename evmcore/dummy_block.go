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
	"math"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/helios/hash"
	"github.com/unicornultrafoundation/go-u2u/helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/trie"

	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/u2u"
)

type (
	EvmHeader struct {
		Number       *big.Int
		Hash         common.Hash
		ParentHash   common.Hash
		SfcStateRoot common.Hash
		Root         common.Hash
		TxHash       common.Hash
		Time         native.Timestamp
		Coinbase     common.Address

		GasLimit uint64
		GasUsed  uint64

		BaseFee *big.Int
	}

	EvmBlock struct {
		EvmHeader

		Transactions types.Transactions
	}
)

// NewEvmBlock constructor.
func NewEvmBlock(h *EvmHeader, txs types.Transactions) *EvmBlock {
	b := &EvmBlock{
		EvmHeader:    *h,
		Transactions: txs,
	}

	if len(txs) == 0 {
		b.EvmHeader.TxHash = types.EmptyRootHash
	} else {
		b.EvmHeader.TxHash = types.DeriveSha(txs, trie.NewStackTrie(nil))
	}

	return b
}

// ToEvmHeader converts native.Block to EvmHeader.
func ToEvmHeader(block *native.Block, index idx.Block, prevHash hash.Event, rules u2u.Rules) *EvmHeader {
	baseFee := rules.Economy.MinGasPrice
	if !rules.Upgrades.London {
		baseFee = nil
	}
	return &EvmHeader{
		Hash:         common.Hash(block.Atropos),
		ParentHash:   common.Hash(prevHash),
		Root:         common.Hash(block.Root),
		SfcStateRoot: common.Hash(block.SfcStateRoot),
		Number:       big.NewInt(int64(index)),
		Time:         block.Time,
		GasLimit:     math.MaxUint64,
		GasUsed:      block.GasUsed,
		BaseFee:      baseFee,
	}
}

// ConvertFromEthHeader converts ETH-formatted header to Helios EVM header
func ConvertFromEthHeader(h *types.Header) *EvmHeader {
	// NOTE: incomplete conversion
	return &EvmHeader{
		Number:       h.Number,
		Coinbase:     h.Coinbase,
		GasLimit:     math.MaxUint64,
		GasUsed:      h.GasUsed,
		Root:         h.Root,
		SfcStateRoot: h.SfcStateRoot,
		TxHash:       h.TxHash,
		ParentHash:   h.ParentHash,
		Time:         native.FromUnix(int64(h.Time)),
		Hash:         common.BytesToHash(h.Extra),
		BaseFee:      h.BaseFee,
	}
}

// EthHeader returns header in ETH format
func (h *EvmHeader) EthHeader() *types.Header {
	if h == nil {
		return nil
	}
	// NOTE: incomplete conversion
	ethHeader := &types.Header{
		Number:       h.Number,
		Coinbase:     h.Coinbase,
		GasLimit:     0xffffffffffff, // don't use h.GasLimit (too much bits) here to avoid parsing issues
		GasUsed:      h.GasUsed,
		Root:         h.Root,
		SfcStateRoot: h.SfcStateRoot,
		TxHash:       h.TxHash,
		ParentHash:   h.ParentHash,
		Time:         uint64(h.Time.Unix()),
		Extra:        h.Hash.Bytes(),
		BaseFee:      h.BaseFee,

		Difficulty: new(big.Int),
	}
	ethHeader.SetExternalHash(h.Hash)
	return ethHeader
}

// Header is a copy of EvmBlock.EvmHeader.
func (b *EvmBlock) Header() *EvmHeader {
	if b == nil {
		return nil
	}
	// copy values
	h := b.EvmHeader
	// copy refs
	h.Number = new(big.Int).Set(b.Number)
	if b.BaseFee != nil {
		h.BaseFee = new(big.Int).Set(b.BaseFee)
	}

	return &h
}

func (b *EvmBlock) NumberU64() uint64 {
	return b.Number.Uint64()
}

func (b *EvmBlock) EthBlock() *types.Block {
	if b == nil {
		return nil
	}
	return types.NewBlock(b.EvmHeader.EthHeader(), b.Transactions, nil, nil, trie.NewStackTrie(nil))
}

func (b *EvmBlock) EstimateSize() int {
	est := 0
	for _, tx := range b.Transactions {
		est += len(tx.Data())
	}
	return est + b.Transactions.Len()*256
}
