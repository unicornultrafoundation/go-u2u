package sfcstore

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/unicornultrafoundation/go-helios/utils/cachescale"
)

type (
	// StoreCacheConfig is a config for the db.
	StoreCacheConfig struct {
		// Cache size for EVM database.
		EvmDatabase int
		// Cache size for EVM snapshot.
		EvmSnap int
		// Cache size for EvmBlock (number of blocks).
		EvmBlocksNum int
		// Cache size for EvmBlock (size in bytes).
		EvmBlocksSize uint
		// Disk journal for saving clean cache entries.
		TrieCleanJournal string
		// Whether to disable trie write caching and GC altogether (archive node)
		TrieDirtyDisabled bool
		// Memory limit (MB) at which to start flushing dirty trie nodes to disk
		TrieDirtyLimit uint
		// Whether to enable greedy gc mode
		GreedyGC bool
	}
	// StoreConfig is a config for store db.
	StoreConfig struct {
		Enable bool
		Cache  StoreCacheConfig
		// Enables tracking of SHA3 preimages in the VM
		EnablePreimageRecording bool
	}
)

// DefaultStoreConfig for product.
func DefaultStoreConfig(scale cachescale.Func) StoreConfig {
	return StoreConfig{
		Enable: false,
		Cache: StoreCacheConfig{
			EvmDatabase:       scale.I(32 * opt.MiB),
			EvmSnap:           scale.I(32 * opt.MiB),
			EvmBlocksNum:      scale.I(5000),
			EvmBlocksSize:     scale.U(6 * opt.MiB),
			TrieDirtyDisabled: true,
			GreedyGC:          false,
			TrieDirtyLimit:    scale.U(256 * opt.MiB),
		},
		EnablePreimageRecording: true,
	}
}

// LiteStoreConfig is for tests or inmemory.
func LiteStoreConfig() StoreConfig {
	return DefaultStoreConfig(cachescale.Ratio{Base: 10, Target: 1})
}
