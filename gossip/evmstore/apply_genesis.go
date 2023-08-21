package evmstore

import (
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/unicornultrafoundation/go-hashgraph/u2udb/batched"

	"github.com/unicornultrafoundation/go-u2u/u2u/genesis"
	"github.com/unicornultrafoundation/go-u2u/utils/adapters/ethdb2udb"
	"github.com/unicornultrafoundation/go-u2u/utils/dbutil/autocompact"
)

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(g genesis.Genesis) (err error) {
	db := batched.Wrap(autocompact.Wrap2M(ethdb2udb.Wrap(s.EvmDb), opt.GiB, 16*opt.GiB, true, "evm"))
	g.RawEvmItems.ForEach(func(key, value []byte) bool {
		err = db.Put(key, value)
		if err != nil {
			return false
		}
		return true
	})
	if err != nil {
		return err
	}
	return db.Write()
}

func (s *Store) WrapTablesAsBatched() (unwrap func()) {
	origTables := s.table

	batchedTxs := batched.Wrap(s.table.Txs)
	s.table.Txs = batchedTxs

	batchedTxPositions := batched.Wrap(s.table.TxPositions)
	s.table.TxPositions = batchedTxPositions

	unwrapLogs := s.EvmLogs.WrapTablesAsBatched()

	batchedReceipts := batched.Wrap(autocompact.Wrap2M(s.table.Receipts, opt.GiB, 16*opt.GiB, false, "receipts"))
	s.table.Receipts = batchedReceipts
	return func() {
		_ = batchedTxs.Flush()
		_ = batchedTxPositions.Flush()
		_ = batchedReceipts.Flush()
		unwrapLogs()
		s.table = origTables
	}
}
