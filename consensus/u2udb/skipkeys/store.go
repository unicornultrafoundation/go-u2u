package skipkeys

import (
	"bytes"

	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb"
)

type Store struct {
	u2udb.Store
	skipPrefix []byte
}

func Wrap(store u2udb.Store, skipPrefix []byte) *Store {
	return &Store{store, skipPrefix}
}

// Has retrieves if a key is present in the key-value data store.
func (s *Store) Has(key []byte) (bool, error) {
	if bytes.HasPrefix(key, s.skipPrefix) {
		return false, nil
	}
	return s.Store.Has(key)
}

// Get retrieves the given key if it's present in the key-value data store.
func (s *Store) Get(key []byte) ([]byte, error) {
	if bytes.HasPrefix(key, s.skipPrefix) {
		return nil, nil
	}
	return s.Store.Get(key)
}

// NewIterator creates a binary-alphabetical iterator over a subset
// of database content with a particular key prefix, starting at a particular
// initial key (or after, if it does not exist).
func (s *Store) NewIterator(prefix []byte, start []byte) u2udb.Iterator {
	return iterator{s.Store.NewIterator(prefix, start), s.skipPrefix}
}

type iterator struct {
	u2udb.Iterator
	skipPrefix []byte
}

func (it iterator) Next() bool {
	first := true
	for first || bytes.HasPrefix(it.Key(), it.skipPrefix) {
		if !it.Iterator.Next() {
			return false
		}
		first = false
	}
	return true
}
