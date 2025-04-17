package multidb

import "github.com/unicornultrafoundation/go-u2u/helios/u2udb"

type closableTable struct {
	u2udb.Store
	underlying u2udb.Store
	noDrop     bool
}

// Close leaves underlying database.
func (s *closableTable) Close() error {
	return s.underlying.Close()
}

// Drop whole database.
func (s *closableTable) Drop() {
	if s.noDrop {
		return
	}
	s.underlying.Drop()
}
