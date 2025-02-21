package sfcstore

import (
	"encoding/binary"

	"github.com/unicornultrafoundation/go-helios/native/idx"

	"github.com/unicornultrafoundation/go-u2u/common"
)

var (
	sfcStateRootPrefix = []byte("sfc") // sfcStateRootPrefix + num (uint64 big endian) + hash -> sfc state root
)

// encodeBlockNumber encodes a block number as big endian uint64
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

// SetStateRoot stores state root hash
func (s *Store) SetStateRoot(n idx.Block, hash []byte, root common.Hash) {
	s.rlp.Set(s.table.StateRoots, SfcStateRootKey(n, hash), root)
}

// GetStateRoot returns stored state root hash
func (s *Store) GetStateRoot(n idx.Block, hash []byte) *common.Hash {
	root, _ := s.rlp.Get(s.table.StateRoots, SfcStateRootKey(n, hash), &common.Hash{}).(*common.Hash)
	return root
}

func SfcStateRootKey(n idx.Block, hash []byte) []byte {
	return append(append(sfcStateRootPrefix, encodeBlockNumber(uint64(n))...), hash...)
}
