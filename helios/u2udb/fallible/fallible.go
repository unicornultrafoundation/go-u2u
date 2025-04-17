package fallible

import (
	"errors"
	"sync/atomic"

	"github.com/unicornultrafoundation/go-u2u/helios/u2udb"
)

var (
	errWriteLimit = errors.New("write limit is over")
)

// Fallible is a u2udb.Store wrapper around any u2udb.Store.
// It falls when write counter is full for test purpose.
type Fallible struct {
	Underlying u2udb.Store

	writes int32
}

// Wrap returns a wrapped u2udb.Store with counter 0. Set it manually.
func Wrap(db u2udb.Store) *Fallible {
	return &Fallible{
		Underlying: db,
	}
}

// SetWriteCount to n.
func (f *Fallible) SetWriteCount(n int) {
	if n <= 0 {
		return
	}
	count := int32(n)
	atomic.StoreInt32(&f.writes, count)
}

// GetWriteCount to left.
func (f *Fallible) GetWriteCount() int {
	count := atomic.LoadInt32(&f.writes)
	return int(count)
}

func (f *Fallible) count() bool {
	count := atomic.AddInt32(&f.writes, -1)
	return count >= 0
}

/*
 * implementation:
 */

// Has retrieves if a key is present in the key-value data store.
func (f *Fallible) Has(key []byte) (bool, error) {
	return f.Underlying.Has(key)
}

// Get retrieves the given key if it's present in the key-value data store.
func (f *Fallible) Get(key []byte) ([]byte, error) {
	return f.Underlying.Get(key)
}

// Put inserts the given value into the key-value data store.
func (f *Fallible) Put(key []byte, value []byte) error {
	if !f.count() {
		panic(errWriteLimit)
	}
	return f.Underlying.Put(key, value)
}

// Delete removes the key from the key-value data store.
func (f *Fallible) Delete(key []byte) error {
	return f.Underlying.Delete(key)
}

// NewBatch creates a write-only database that buffers changes to its host db
// until a final write is called.
func (f *Fallible) NewBatch() u2udb.Batch {
	return f.Underlying.NewBatch()
}

// NewIterator creates a binary-alphabetical iterator over a subset
// of database content with a particular key prefix, starting at a particular
// initial key (or after, if it does not exist).
func (f *Fallible) NewIterator(prefix []byte, start []byte) u2udb.Iterator {
	return f.Underlying.NewIterator(prefix, start)
}

// GetSnapshot returns a latest snapshot of the underlying DB. A snapshot
// is a frozen snapshot of a DB state at a particular point in time. The
// content of snapshot are guaranteed to be consistent.
//
// The snapshot must be released after use, by calling Release method.
func (f *Fallible) GetSnapshot() (u2udb.Snapshot, error) {
	return f.Underlying.GetSnapshot()
}

// Stat returns a particular internal stat of the database.
func (f *Fallible) Stat(property string) (string, error) {
	return f.Underlying.Stat(property)
}

// Compact flattens the underlying data store for the given key range. In essence,
// deleted and overwritten versions are discarded, and the data is rearranged to
// reduce the cost of operations needed to access them.
//
// A nil start is treated as a key before all keys in the data store; a nil limit
// is treated as a key after all keys in the data store. If both is nil then it
// will compact entire data store.
func (f *Fallible) Compact(start []byte, limit []byte) error {
	return f.Underlying.Compact(start, limit)
}

// Close closes database.
func (f *Fallible) Close() error {
	if !f.count() {
		panic(errWriteLimit)
	}

	return f.Underlying.Close()
}

// Drop drops database.
func (f *Fallible) Drop() {
	if !f.count() {
		panic(errWriteLimit)
	}

	f.Underlying.Drop()
}
