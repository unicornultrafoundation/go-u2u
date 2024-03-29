package asyncflushproducer

import "github.com/unicornultrafoundation/go-helios/u2udb"

type store struct {
	u2udb.Store
	CloseFn func() error
}

func (s *store) Close() error {
	return s.CloseFn()
}
