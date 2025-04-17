package snap2udb

import (
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb/devnulldb"
	"github.com/unicornultrafoundation/go-u2u/log"
)

type Adapter struct {
	u2udb.Snapshot
}

var _ u2udb.Store = (*Adapter)(nil)

func Wrap(v u2udb.Snapshot) *Adapter {
	return &Adapter{v}
}

func (db *Adapter) Put(key []byte, value []byte) error {
	log.Warn("called Put on snapshot")
	return nil
}

func (db *Adapter) Delete(key []byte) error {
	log.Warn("called Delete on snapshot")
	return nil
}

func (db *Adapter) GetSnapshot() (u2udb.Snapshot, error) {
	return db.Snapshot, nil
}

func (db *Adapter) NewBatch() u2udb.Batch {
	log.Warn("called NewBatch on snapshot")
	return devnulldb.New().NewBatch()
}

func (db *Adapter) Compact(start []byte, limit []byte) error {
	return nil
}

func (db *Adapter) Close() error {
	return nil
}

func (db *Adapter) Drop() {}

func (db *Adapter) Stat(property string) (string, error) {
	return "", nil
}
