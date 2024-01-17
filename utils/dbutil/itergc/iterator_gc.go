package itergc

import (
	"sync"

	"github.com/unicornultrafoundation/go-helios/u2udb"
)

type Snapshot struct {
	u2udb.Snapshot
	nextID uint64
	iters  map[uint64]u2udb.Iterator
	mu     sync.Locker
}

type Iterator struct {
	u2udb.Iterator
	mu    sync.Locker
	id    uint64
	iters map[uint64]u2udb.Iterator
}

// Wrap snapshot to automatically close all pending iterators upon snapshot release
func Wrap(snapshot u2udb.Snapshot, mu sync.Locker) *Snapshot {
	return &Snapshot{
		Snapshot: snapshot,
		iters:    make(map[uint64]u2udb.Iterator),
		mu:       mu,
	}
}

func (s *Snapshot) NewIterator(prefix []byte, start []byte) u2udb.Iterator {
	s.mu.Lock()
	defer s.mu.Unlock()
	it := s.Snapshot.NewIterator(prefix, start)
	id := s.nextID
	s.iters[id] = it
	s.nextID++

	return &Iterator{
		Iterator: it,
		mu:       s.mu,
		id:       id,
		iters:    s.iters,
	}
}

func (s *Iterator) Release() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Iterator.Release()
	delete(s.iters, s.id)
}

func (s *Snapshot) Release() {
	s.mu.Lock()
	defer s.mu.Unlock()
	// release all pending iterators
	for _, it := range s.iters {
		it.Release()
	}
	s.iters = nil
	s.Snapshot.Release()
}
