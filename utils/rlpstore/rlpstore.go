package rlpstore

import (
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb"
	"github.com/unicornultrafoundation/go-u2u/logger"
	"github.com/unicornultrafoundation/go-u2u/rlp"
)

type Helper struct {
	logger.Instance
}

// Set RLP value
func (s *Helper) Set(table u2udb.Store, key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		s.Log.Crit("Failed to encode rlp", "err", err)
	}

	if err := table.Put(key, buf); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// Get RLP value
func (s *Helper) Get(table u2udb.Store, key []byte, to interface{}) interface{} {
	buf, err := table.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return nil
	}

	err = rlp.DecodeBytes(buf, to)
	if err != nil {
		s.Log.Crit("Failed to decode rlp", "err", err, "size", len(buf))
	}
	return to
}
