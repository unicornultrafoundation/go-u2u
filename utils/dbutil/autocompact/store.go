package autocompact

import (
	"sync"

	"github.com/status-im/keycard-go/hexutils"

	"github.com/unicornultrafoundation/go-u2u/log"

	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb"
)

// Store implements automatic compacting of recently inserted/erased data according to provided strategy
type Store struct {
	u2udb.Store
	limit   uint64
	cont    ContainerI
	newCont func() ContainerI
	compMu  sync.Mutex
	name    string
}

type Batch struct {
	u2udb.Batch
	store *Store
	cont  ContainerI
}

func Wrap(s u2udb.Store, limit uint64, strategy func() ContainerI, name string) *Store {
	return &Store{
		Store:   s,
		limit:   limit,
		newCont: strategy,
		cont:    strategy(),
		name:    name,
	}
}

func Wrap2(s u2udb.Store, limit1 uint64, limit2 uint64, strategy func() ContainerI, name string) *Store {
	return Wrap(Wrap(s, limit1, strategy, name), limit2, strategy, name)
}

func Wrap2M(s u2udb.Store, limit1 uint64, limit2 uint64, forward bool, name string) *Store {
	strategy := NewBackwardsCont
	if forward {
		strategy = NewForwardCont
	}
	return Wrap2(s, limit1, limit2, strategy, name)
}

func estSize(keyLen int, valLen int) uint64 {
	// Storage overheads, related to adding/deleting a record,
	//wouldn't be only proportional to length of key and value.
	//E.g. if one adds 10 records with length of 2, it will be more expensive than 1 record with length 20
	// Now, 64 wasn't really calculated but is rather a guesstimation
	return uint64(keyLen + valLen + 64)
}

func (s *Store) onWrite(key []byte, size uint64) {
	s.compMu.Lock()
	defer s.compMu.Unlock()
	if key != nil {
		s.cont.Add(key, size)
	}
	s.mayCompact(false)
}

func (s *Store) onBatchWrite(batchCont ContainerI) {
	s.compMu.Lock()
	defer s.compMu.Unlock()
	s.cont.Merge(batchCont)
	s.mayCompact(false)
}

func (s *Store) compact() {
	s.compMu.Lock()
	defer s.compMu.Unlock()
	s.mayCompact(true)
}

func (s *Store) mayCompact(force bool) {
	// error handling
	err := s.cont.Error()
	if err != nil {
		s.cont.Reset()
		s.newCont = NewDevnullCont
		s.cont = s.newCont()
		log.Warn("Autocompaction failed, which may lead to performance issues", "name", s.name, "err", err)
	}

	if force || s.cont.Size() > s.limit {
		for _, r := range s.cont.Ranges() {
			log.Debug("Autocompact", "name", s.name, "from", hexutils.BytesToHex(r.minKey), "to", hexutils.BytesToHex(r.maxKey))
			_ = s.Store.Compact(r.minKey, r.maxKey)
		}
		s.cont.Reset()
	}
}

func (s *Store) Put(key []byte, value []byte) error {
	defer s.onWrite(key, estSize(len(key), len(value)))
	return s.Store.Put(key, value)
}

func (s *Store) Delete(key []byte) error {
	defer s.onWrite(key, estSize(len(key), 0))
	return s.Store.Delete(key)
}

func (s *Store) Close() error {
	s.compact()
	return s.Store.Close()
}

func (s *Store) NewBatch() u2udb.Batch {
	batch := s.Store.NewBatch()
	if batch == nil {
		return nil
	}
	return &Batch{
		Batch: batch,
		store: s,
		cont:  s.newCont(),
	}
}

func (s *Batch) Put(key []byte, value []byte) error {
	s.cont.Add(key, estSize(len(key), len(value)))
	return s.Batch.Put(key, value)
}

func (s *Batch) Delete(key []byte) error {
	s.cont.Add(key, estSize(len(key), 0))
	return s.Batch.Delete(key)
}

func (s *Batch) Reset() {
	s.cont.Reset()
	s.Batch.Reset()
}

func (s *Batch) Write() error {
	defer s.store.onBatchWrite(s.cont)
	return s.Batch.Write()
}
