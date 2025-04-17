package vecfc

import (
	"github.com/unicornultrafoundation/go-u2u/consensus/hash"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/dag"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/idx"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/pos"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb/table"
	"github.com/unicornultrafoundation/go-u2u/consensus/utils/cachescale"
	"github.com/unicornultrafoundation/go-u2u/consensus/utils/simplewlru"
	"github.com/unicornultrafoundation/go-u2u/consensus/vecengine"
)

// IndexCacheConfig - config for cache sizes of Engine
type IndexCacheConfig struct {
	ForklessCausePairs   int
	HighestBeforeSeqSize uint
	LowestAfterSeqSize   uint
}

// IndexConfig - Engine config (cache sizes)
type IndexConfig struct {
	Caches IndexCacheConfig
}

// Index is a data to detect forkless-cause condition, calculate median timestamp, detect forks.
type Index struct {
	*vecengine.Engine

	crit          func(error)
	validators    *pos.Validators
	validatorIdxs map[idx.ValidatorID]idx.Validator

	getEvent func(hash.Event) dag.Event

	vecDb u2udb.Store
	table struct {
		HighestBeforeSeq u2udb.Store `table:"S"`
		LowestAfterSeq   u2udb.Store `table:"s"`
	}

	cache struct {
		HighestBeforeSeq *simplewlru.Cache
		LowestAfterSeq   *simplewlru.Cache
		ForklessCause    *simplewlru.Cache
	}

	cfg IndexConfig
}

// DefaultConfig returns default index config
func DefaultConfig(scale cachescale.Func) IndexConfig {
	return IndexConfig{
		Caches: IndexCacheConfig{
			ForklessCausePairs:   scale.I(20000),
			HighestBeforeSeqSize: scale.U(160 * 1024),
			LowestAfterSeqSize:   scale.U(160 * 1024),
		},
	}
}

// LiteConfig returns default index config for tests
func LiteConfig() IndexConfig {
	return DefaultConfig(cachescale.Ratio{Base: 100, Target: 1})
}

// NewIndex creates Index instance.
func NewIndex(crit func(error), config IndexConfig) *Index {
	vi := &Index{
		cfg:  config,
		crit: crit,
	}
	vi.Engine = vecengine.NewIndex(crit, vi.GetEngineCallbacks())
	vi.initCaches()

	return vi
}

func NewIndexWithEngine(crit func(error), config IndexConfig, engine *vecengine.Engine) *Index {
	vi := &Index{
		Engine: engine,
		cfg:    config,
		crit:   crit,
	}
	vi.initCaches()

	return vi
}

func (vi *Index) initCaches() {
	vi.cache.ForklessCause, _ = simplewlru.New(uint(vi.cfg.Caches.ForklessCausePairs), vi.cfg.Caches.ForklessCausePairs)
	vi.cache.HighestBeforeSeq, _ = simplewlru.New(vi.cfg.Caches.HighestBeforeSeqSize, int(vi.cfg.Caches.HighestBeforeSeqSize))
	vi.cache.LowestAfterSeq, _ = simplewlru.New(vi.cfg.Caches.LowestAfterSeqSize, int(vi.cfg.Caches.HighestBeforeSeqSize))
}

// Reset resets buffers.
func (vi *Index) Reset(validators *pos.Validators, db u2udb.FlushableKVStore, getEvent func(hash.Event) dag.Event) error {
	vi.Engine.Reset(validators, db, getEvent)
	vi.vecDb = db
	err := table.MigrateTables(&vi.table, vi.vecDb)
	vi.getEvent = getEvent
	vi.validators = validators
	vi.validatorIdxs = validators.Idxs()
	vi.cache.ForklessCause.Purge()
	vi.onDropNotFlushed()
	return err
}

func (vi *Index) GetEngineCallbacks() vecengine.Callbacks {
	return vecengine.Callbacks{
		GetHighestBefore: func(event hash.Event) vecengine.HighestBeforeI {
			return vi.GetHighestBefore(event)
		},
		GetLowestAfter: func(event hash.Event) vecengine.LowestAfterI {
			return vi.GetLowestAfter(event)
		},
		SetHighestBefore: func(event hash.Event, b vecengine.HighestBeforeI) {
			vi.SetHighestBefore(event, b.(*HighestBeforeSeq))
		},
		SetLowestAfter: func(event hash.Event, b vecengine.LowestAfterI) {
			vi.SetLowestAfter(event, b.(*LowestAfterSeq))
		},
		NewHighestBefore: func(size idx.Validator) vecengine.HighestBeforeI {
			return NewHighestBeforeSeq(size)
		},
		NewLowestAfter: func(size idx.Validator) vecengine.LowestAfterI {
			return NewLowestAfterSeq(size)
		},
		OnDropNotFlushed: vi.onDropNotFlushed,
	}
}

func (vi *Index) onDropNotFlushed() {
	vi.cache.HighestBeforeSeq.Purge()
	vi.cache.LowestAfterSeq.Purge()
}

// GetMergedHighestBefore returns HighestBefore vector clock without branches, where branches are merged into one
func (vi *Index) GetMergedHighestBefore(id hash.Event) *HighestBeforeSeq {
	return vi.Engine.GetMergedHighestBefore(id).(*HighestBeforeSeq)
}
