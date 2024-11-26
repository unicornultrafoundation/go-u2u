package ethdb2udb

import (
	"github.com/unicornultrafoundation/go-helios/u2udb"
	"github.com/unicornultrafoundation/go-u2u/ethdb"
)

type Adapter struct {
	ethdb.KeyValueStore
}

var _ u2udb.Store = (*Adapter)(nil)

func Wrap(v ethdb.KeyValueStore) *Adapter {
	return &Adapter{v}
}

func (db *Adapter) Drop() {
	panic("called Drop on ethdb")
}

// batch is a write-only memory batch that commits changes to its host
// database when Write is called. A batch cannot be used concurrently.
type batch struct {
	ethdb.Batch
}

// Replay replays the batch contents.
func (b *batch) Replay(w u2udb.Writer) error {
	return b.Batch.Replay(w)
}

// NewBatch creates a write-only key-value store that buffers changes to its host
// database until a final write is called.
func (db *Adapter) NewBatch() u2udb.Batch {
	return &batch{db.KeyValueStore.NewBatch()}
}

func (db *Adapter) GetSnapshot() (u2udb.Snapshot, error) {
	panic("called GetSnapshot on ethdb")
}

// NewIterator creates a binary-alphabetical iterator over a subset
// of database content with a particular key prefix, starting at a particular
// initial key (or after, if it does not exist).
func (db *Adapter) NewIterator(prefix []byte, start []byte) u2udb.Iterator {
	return db.KeyValueStore.NewIterator(prefix, start)
}
