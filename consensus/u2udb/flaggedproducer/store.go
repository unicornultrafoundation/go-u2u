package flaggedproducer

import (
	"sync/atomic"

	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb/flushable"
)

type flaggedStore struct {
	u2udb.Store
	DropFn     func()
	Dirty      uint32
	flushIDKey []byte
}

type flaggedBatch struct {
	u2udb.Batch
	db *flaggedStore
}

func (s *flaggedStore) Close() error {
	return nil
}

func (s *flaggedStore) Drop() {
	s.DropFn()
}

func (s *flaggedStore) modified() error {
	if atomic.LoadUint32(&s.Dirty) == 0 {
		atomic.StoreUint32(&s.Dirty, 1)
		err := s.Store.Put(s.flushIDKey, []byte{flushable.DirtyPrefix})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *flaggedStore) Put(key []byte, value []byte) error {
	err := s.modified()
	if err != nil {
		return err
	}
	return s.Store.Put(key, value)
}

func (s *flaggedStore) Delete(key []byte) error {
	err := s.modified()
	if err != nil {
		return err
	}
	return s.Store.Delete(key)
}

func (s *flaggedStore) NewBatch() u2udb.Batch {
	return &flaggedBatch{
		Batch: s.Store.NewBatch(),
		db:    s,
	}
}

func (s *flaggedBatch) Write() error {
	err := s.db.modified()
	if err != nil {
		return err
	}
	return s.Batch.Write()
}
