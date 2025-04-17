package consensus

import (
	"errors"

	"github.com/unicornultrafoundation/go-u2u/rlp"

	"github.com/unicornultrafoundation/go-u2u/helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb/memorydb"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb/table"
	"github.com/unicornultrafoundation/go-u2u/helios/utils/simplewlru"
)

// Store is a consensus persistent storage working over parent key-value database.
type Store struct {
	getEpochDB EpochDBProducer
	cfg        StoreConfig
	crit       func(error)

	mainDB u2udb.Store
	table  struct {
		LastDecidedState u2udb.Store `table:"c"`
		EpochState       u2udb.Store `table:"e"`
	}

	cache struct {
		LastDecidedState *LastDecidedState
		EpochState       *EpochState
		FrameRoots       *simplewlru.Cache `cache:"-"` // store by pointer
	}

	epochDB    u2udb.Store
	epochTable struct {
		Roots          u2udb.Store `table:"r"`
		VectorIndex    u2udb.Store `table:"v"`
		ConfirmedEvent u2udb.Store `table:"C"`
	}
}

var (
	ErrNoGenesis = errors.New("genesis not applied")
)

type EpochDBProducer func(epoch idx.Epoch) u2udb.Store

// NewStore creates store over key-value db.
func NewStore(mainDB u2udb.Store, getDB EpochDBProducer, crit func(error), cfg StoreConfig) (*Store, error) {
	s := &Store{
		getEpochDB: getDB,
		cfg:        cfg,
		crit:       crit,
		mainDB:     mainDB,
	}

	err := table.MigrateTables(&s.table, s.mainDB)

	s.initCache()

	return s, err
}

func (s *Store) initCache() {
	s.cache.FrameRoots = s.makeCache(s.cfg.Cache.RootsNum, s.cfg.Cache.RootsFrames)
}

// NewMemStore creates store over memory map.
// Store is always blank.
func NewMemStore() (*Store, error) {
	getDb := func(epoch idx.Epoch) u2udb.Store {
		return memorydb.New()
	}
	cfg := LiteStoreConfig()
	crit := func(err error) {
		panic(err)
	}
	return NewStore(memorydb.New(), getDb, crit, cfg)
}

// Close leaves underlying database.
func (s *Store) Close() error {
	setnil := func() interface{} {
		return nil
	}

	if err := table.MigrateTables(&s.table, nil); err != nil {
		return err
	}
	table.MigrateCaches(&s.cache, setnil)
	if err := table.MigrateTables(&s.epochTable, nil); err != nil {
		return err
	}
	err := s.mainDB.Close()
	if err != nil {
		return err
	}

	if s.epochDB != nil {
		err = s.epochDB.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// dropEpochDB drops existing epoch DB
func (s *Store) dropEpochDB() error {
	prevDb := s.epochDB
	if prevDb != nil {
		err := prevDb.Close()
		if err != nil {
			return err
		}
		prevDb.Drop()
	}
	return nil
}

// openEpochDB makes new epoch DB
func (s *Store) openEpochDB(n idx.Epoch) error {
	// Clear full LRU cache.
	s.cache.FrameRoots.Purge()

	s.epochDB = s.getEpochDB(n)
	return table.MigrateTables(&s.epochTable, s.epochDB)
}

/*
 * Utils:
 */

// set RLP value
func (s *Store) set(table u2udb.Store, key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		s.crit(err)
	}

	if err := table.Put(key, buf); err != nil {
		s.crit(err)
	}
}

// get RLP value
func (s *Store) get(table u2udb.Store, key []byte, to interface{}) interface{} {
	buf, err := table.Get(key)
	if err != nil {
		s.crit(err)
	}
	if buf == nil {
		return nil
	}

	err = rlp.DecodeBytes(buf, to)
	if err != nil {
		s.crit(err)
	}
	return to
}

func (s *Store) makeCache(weight uint, size int) *simplewlru.Cache {
	cache, err := simplewlru.New(weight, size)
	if err != nil {
		s.crit(err)
	}
	return cache
}
