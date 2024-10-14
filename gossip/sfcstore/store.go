package sfcstore

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/unicornultrafoundation/go-helios/hash"
	"github.com/unicornultrafoundation/go-helios/native/idx"
	"github.com/unicornultrafoundation/go-helios/u2udb"
	"github.com/unicornultrafoundation/go-helios/u2udb/nokeyiserr"
	"github.com/unicornultrafoundation/go-helios/u2udb/table"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/common/prque"
	"github.com/unicornultrafoundation/go-u2u/core/rawdb"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/state/snapshot"
	"github.com/unicornultrafoundation/go-u2u/ethdb"
	"github.com/unicornultrafoundation/go-u2u/trie"

	"github.com/unicornultrafoundation/go-u2u/logger"
	"github.com/unicornultrafoundation/go-u2u/topicsdb"
	"github.com/unicornultrafoundation/go-u2u/utils/adapters/udb2ethdb"
	"github.com/unicornultrafoundation/go-u2u/utils/rlpstore"
)

const nominalSize uint = 1

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	cfg StoreConfig

	table struct {
		Evm        u2udb.Store `table:"A"`
		StateRoots u2udb.Store `table:"C"`
	}

	EvmDb    ethdb.Database
	EvmState state.Database
	EvmLogs  topicsdb.Index
	Snaps    *snapshot.Tree

	rlp rlpstore.Helper

	triegc *prque.Prque // Priority queue mapping block numbers to tries to gc

	logger.Instance
}

const (
	TriesInMemory = 16
)

// NewStore creates store over key-value db.
func NewStore(dbs u2udb.DBProducer, cfg StoreConfig) *Store {
	s := &Store{
		cfg:      cfg,
		Instance: logger.New("sfc-store"),
		rlp:      rlpstore.Helper{logger.New("rlp")},
		triegc:   prque.New(nil),
	}

	err := table.OpenTables(&s.table, dbs, "sfc")
	if err != nil {
		s.Log.Crit("Failed to open tables", "err", err)
	}

	s.initEVMDB()
	s.EvmLogs = topicsdb.NewWithThreadPool(dbs)

	return s
}

// Close closes underlying database.
func (s *Store) Close() {
	_ = table.CloseTables(&s.table)
	table.MigrateTables(&s.table, nil)
	s.EvmLogs.Close()
}

func (s *Store) initEVMDB() {
	s.EvmDb = rawdb.NewDatabase(
		udb2ethdb.Wrap(
			nokeyiserr.Wrap(
				s.table.Evm)))
	s.EvmState = state.NewDatabaseWithConfig(s.EvmDb, &trie.Config{
		Cache:     s.cfg.Cache.EvmDatabase / opt.MiB,
		Journal:   s.cfg.Cache.TrieCleanJournal,
		Preimages: s.cfg.EnablePreimageRecording,
		GreedyGC:  s.cfg.Cache.GreedyGC,
	})
}

func (s *Store) ResetWithEVMDB(evmStore u2udb.Store) *Store {
	cp := *s
	cp.table.Evm = evmStore
	cp.initEVMDB()
	cp.Snaps = nil
	return &cp
}

func (s *Store) EVMDB() u2udb.Store {
	return s.table.Evm
}

// Commit changes.
func (s *Store) Commit(block idx.Block, root hash.Hash, flush bool) error {
	triedb := s.EvmState.TrieDB()
	stateRoot := common.Hash(root)
	// If we're applying genesis or running an archive node, always flush
	if flush || s.cfg.Cache.TrieDirtyDisabled {
		err := triedb.Commit(stateRoot, false, nil)
		if err != nil {
			s.Log.Error("Failed to flush trie DB into main DB", "err", err)
		}
		return err
	} else {
		// Full but not archive node, do proper garbage collection
		triedb.Reference(stateRoot, common.Hash{}) // metadata reference to keep trie alive
		s.triegc.Push(stateRoot, -int64(block))

		if current := uint64(block); current > TriesInMemory {
			// If we exceeded our memory allowance, flush matured singleton nodes to disk
			s.Cap()

			// Find the next state trie we need to commit
			chosen := current - TriesInMemory

			// Garbage collect all below the chosen block
			for !s.triegc.Empty() {
				root, number := s.triegc.Pop()
				if uint64(-number) > chosen {
					s.triegc.Push(root, number)
					break
				}
				triedb.Dereference(root.(common.Hash))
			}
		}
		return nil
	}
}

// Cap flush matured singleton nodes to disk
func (s *Store) Cap() {
	triedb := s.EvmState.TrieDB()
	var (
		nodes, imgs = triedb.Size()
		limit       = common.StorageSize(s.cfg.Cache.TrieDirtyLimit)
	)
	// If we exceeded our memory allowance, flush matured singleton nodes to disk
	if nodes > limit+ethdb.IdealBatchSize || imgs > 4*1024*1024 {
		triedb.Cap(limit)
	}
}

// StateDB returns state database.
func (s *Store) StateDB(from hash.Hash) (*state.StateDB, error) {
	return state.NewWithSnapLayers(common.Hash(from), s.EvmState, s.Snaps, 0)
}

// HasStateDB returns if state database exists
func (s *Store) HasStateDB(from hash.Hash) bool {
	_, err := s.StateDB(from)
	return err == nil
}
