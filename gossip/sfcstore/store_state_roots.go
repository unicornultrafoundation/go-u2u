package sfcstore

import (
	"github.com/unicornultrafoundation/go-helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/common"
)

// SetStateRoot stores state root hash
func (s *Store) SetStateRoot(n idx.Block, root common.Hash) {
	s.rlp.Set(s.table.StateRoots, n.Bytes(), root)
}

// GetStateRoot returns stored state root hash
func (s *Store) GetStateRoot(n idx.Block) *common.Hash {
	root, _ := s.rlp.Get(s.table.StateRoots, n.Bytes(), &common.Hash{}).(*common.Hash)

	return root
}
