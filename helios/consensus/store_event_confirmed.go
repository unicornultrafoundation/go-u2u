package consensus

import (
	"github.com/unicornultrafoundation/go-u2u/helios/hash"
	"github.com/unicornultrafoundation/go-u2u/helios/native/idx"
)

// SetEventConfirmedOn stores confirmed event hash.
func (s *Store) SetEventConfirmedOn(e hash.Event, on idx.Frame) {
	key := e.Bytes()

	if err := s.epochTable.ConfirmedEvent.Put(key, on.Bytes()); err != nil {
		s.crit(err)
	}
}

// GetEventConfirmedOn returns confirmed event hash.
func (s *Store) GetEventConfirmedOn(e hash.Event) idx.Frame {
	key := e.Bytes()

	buf, err := s.epochTable.ConfirmedEvent.Get(key)
	if err != nil {
		s.crit(err)
	}
	if buf == nil {
		return 0
	}

	return idx.BytesToFrame(buf)
}
