package readonlystore

import "github.com/unicornultrafoundation/go-u2u/consensus/u2udb"

type Store struct {
	u2udb.Store
}

func Wrap(s u2udb.Store) *Store {
	return &Store{s}
}

// Put inserts the given value into the key-value data store.
func (s *Store) Put(key []byte, value []byte) error {
	return u2udb.ErrUnsupportedOp
}

// Delete removes the key from the key-value data store.
func (s *Store) Delete(key []byte) error {
	return u2udb.ErrUnsupportedOp
}

type Batch struct {
	u2udb.Batch
}

func (s *Store) NewBatch() u2udb.Batch {
	return &Batch{s.Store.NewBatch()}
}

// Put inserts the given value into the key-value data store.
func (s *Batch) Put(key []byte, value []byte) error {
	return u2udb.ErrUnsupportedOp
}

// Delete removes the key from the key-value data store.
func (s *Batch) Delete(key []byte) error {
	return u2udb.ErrUnsupportedOp
}
