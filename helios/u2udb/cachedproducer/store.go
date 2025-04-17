package cachedproducer

import "github.com/unicornultrafoundation/go-u2u/helios/u2udb"

type StoreWithFn struct {
	u2udb.Store
	CloseFn func() error
	DropFn  func()
}

func (s *StoreWithFn) Close() error {
	return s.CloseFn()
}

func (s *StoreWithFn) Drop() {
	s.DropFn()
}
