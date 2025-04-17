package ibr

import (
	"github.com/unicornultrafoundation/go-u2u/helios/common/bigendian"
	"github.com/unicornultrafoundation/go-u2u/helios/hash"
	"github.com/unicornultrafoundation/go-u2u/helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/core/types"

	"github.com/unicornultrafoundation/go-u2u/native"
)

type LlrBlockVote struct {
	Atropos      hash.Event
	Root         hash.Hash
	TxHash       hash.Hash
	ReceiptsHash hash.Hash
	Time         native.Timestamp
	GasUsed      uint64
	SfcStateRoot hash.Hash `rlp:"optional"`
}

type LlrFullBlockRecord struct {
	Atropos      hash.Event
	Root         hash.Hash
	Txs          types.Transactions
	Receipts     []*types.ReceiptForStorage
	Time         native.Timestamp
	GasUsed      uint64
	SfcStateRoot hash.Hash `rlp:"optional"`
}

type LlrIdxFullBlockRecord struct {
	LlrFullBlockRecord
	Idx idx.Block
}

func (bv LlrBlockVote) Hash() hash.Hash {
	return hash.Of(bv.Atropos.Bytes(), bv.Root.Bytes(), bv.TxHash.Bytes(), bv.ReceiptsHash.Bytes(), bv.Time.Bytes(), bigendian.Uint64ToBytes(bv.GasUsed))
}

func (br LlrFullBlockRecord) Hash() hash.Hash {
	return LlrBlockVote{
		Atropos:      br.Atropos,
		Root:         br.Root,
		TxHash:       native.CalcTxHash(br.Txs),
		ReceiptsHash: native.CalcReceiptsHash(br.Receipts),
		Time:         br.Time,
		GasUsed:      br.GasUsed,
		SfcStateRoot: br.SfcStateRoot,
	}.Hash()
}
