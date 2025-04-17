package gossip

/*
	In LRU cache data stored like pointer
*/

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/unicornultrafoundation/go-u2u/helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb/skiperrors"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb/table"

	"github.com/unicornultrafoundation/go-u2u/logger"
)

var (
	errDBClosed = errors.New("database closed")
)

type (
	epochStore struct {
		epoch idx.Epoch
		db    u2udb.Store
		table struct {
			LastEvents u2udb.Store `table:"t"`
			Heads      u2udb.Store `table:"H"`
			DagIndex   u2udb.Store `table:"v"`
		}
		cache struct {
			Heads      atomic.Value
			LastEvents atomic.Value
		}

		logger.Instance
	}
)

func newEpochStore(epoch idx.Epoch, db u2udb.Store) *epochStore {
	es := &epochStore{
		epoch:    epoch,
		db:       db,
		Instance: logger.New("epoch-store"),
	}
	table.MigrateTables(&es.table, db)

	// wrap with skiperrors to skip errors on reading from a dropped DB
	es.table.LastEvents = skiperrors.Wrap(es.table.LastEvents, errDBClosed)
	es.table.Heads = skiperrors.Wrap(es.table.Heads, errDBClosed)

	// load the cache to avoid a race condition
	es.GetHeads()
	es.GetLastEvents()

	return es
}

func (s *Store) getAnyEpochStore() *epochStore {
	_es := s.epochStore.Load()
	if _es == nil {
		return nil
	}
	return _es.(*epochStore)
}

// getEpochStore is safe for concurrent use.
func (s *Store) getEpochStore(epoch idx.Epoch) *epochStore {
	es := s.getAnyEpochStore()
	if es.epoch != epoch {
		return nil
	}
	return es
}

func (s *Store) resetEpochStore(newEpoch idx.Epoch) {
	oldEs := s.epochStore.Load()
	// create new DB
	s.createEpochStore(newEpoch)
	// drop previous DB
	// there may be race condition with threads which hold this DB, so wrap tables with skiperrors
	if oldEs != nil {
		err := oldEs.(*epochStore).db.Close()
		if err != nil {
			s.Log.Error("Failed to close epoch DB", "err", err)
			return
		}
		oldEs.(*epochStore).db.Drop()
	}
}

func (s *Store) loadEpochStore(epoch idx.Epoch) {
	if s.epochStore.Load() != nil {
		return
	}
	s.createEpochStore(epoch)
}

func (s *Store) closeEpochStore() error {
	es := s.getAnyEpochStore()
	if es == nil {
		return nil
	}
	return es.db.Close()
}

func (s *Store) createEpochStore(epoch idx.Epoch) {
	// create new DB
	name := fmt.Sprintf("gossip-%d", epoch)
	db, err := s.dbs.OpenDB(name)
	if err != nil {
		s.Log.Crit("Filed to open DB", "name", name, "err", err)
	}
	s.epochStore.Store(newEpochStore(epoch, db))
}
