package asyncflushproducer

import (
	"github.com/unicornultrafoundation/go-hashgraph/kvdb"
)

type store struct {
	kvdb.Store
	CloseFn func() error
}

func (s *store) Close() error {
	return s.CloseFn()
}
