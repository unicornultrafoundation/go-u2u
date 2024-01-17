package evmstore

/*
	In LRU cache data stored like pointer
*/

import (
	"github.com/unicornultrafoundation/go-helios/hash"
	"github.com/unicornultrafoundation/go-helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/common"
)

type TxPosition struct {
	Block       idx.Block
	Event       hash.Event
	EventOffset uint32
	BlockOffset uint32
}

// SetTxPosition stores transaction block and position.
func (s *Store) SetTxPosition(txid common.Hash, position TxPosition) {
	s.rlp.Set(s.table.TxPositions, txid.Bytes(), &position)

	// Add to LRU cache.
	s.cache.TxPositions.Add(txid.String(), &position, nominalSize)
}

// GetTxPosition returns stored transaction block and position.
func (s *Store) GetTxPosition(txid common.Hash) *TxPosition {
	// Get data from LRU cache first.
	if c, ok := s.cache.TxPositions.Get(txid.String()); ok {
		if b, ok := c.(*TxPosition); ok {
			return b
		}
	}

	txPosition, _ := s.rlp.Get(s.table.TxPositions, txid.Bytes(), &TxPosition{}).(*TxPosition)

	// Add to LRU cache.
	if txPosition != nil {
		s.cache.TxPositions.Add(txid.String(), txPosition, nominalSize)
	}

	return txPosition
}
