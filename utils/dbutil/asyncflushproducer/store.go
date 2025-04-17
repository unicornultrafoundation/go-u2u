package asyncflushproducer

import "github.com/unicornultrafoundation/go-u2u/helios/u2udb"

type store struct {
	u2udb.Store
	CloseFn func() error
}

func (s *store) Close() error {
	return s.CloseFn()
}
