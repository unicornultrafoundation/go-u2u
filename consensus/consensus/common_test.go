package consensus

import (
	"math/rand"

	"github.com/unicornultrafoundation/go-u2u/consensus/hash"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/idx"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/pos"
	"github.com/unicornultrafoundation/go-u2u/consensus/types"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb/memorydb"
	"github.com/unicornultrafoundation/go-u2u/consensus/utils/adapters"
	"github.com/unicornultrafoundation/go-u2u/consensus/vecfc"
)

type applyBlockFn func(block *types.Block) *pos.Validators

type BlockKey struct {
	Epoch idx.Epoch
	Frame idx.Frame
}

type BlockResult struct {
	Event      hash.Event
	Cheaters   types.Cheaters
	Validators *pos.Validators
}

// TestConsensus extends Consensus for tests.
type TestConsensus struct {
	*Indexed

	blocks      map[BlockKey]*BlockResult
	lastBlock   BlockKey
	epochBlocks map[idx.Epoch]idx.Frame

	applyBlock applyBlockFn
}

// FakeConsensus creates empty consensus with mem store and equal weights of nodes in genesis.
func FakeConsensus(nodes []idx.ValidatorID, weights []pos.Weight, mods ...memorydb.Mod) (*TestConsensus, *Store, *EventStore, *adapters.VectorToDagIndexer) {
	validators := make(pos.ValidatorsBuilder, len(nodes))
	for i, v := range nodes {
		if weights == nil {
			validators[v] = 1
		} else {
			validators[v] = weights[i]
		}
	}

	openEDB := func(epoch idx.Epoch) u2udb.Store {
		return memorydb.New()
	}
	crit := func(err error) {
		panic(err)
	}
	store, err := NewStore(memorydb.New(), openEDB, crit, LiteStoreConfig())
	if err != nil {
		panic(err)
	}

	err = store.ApplyGenesis(&Genesis{
		Validators: validators.Build(),
		Epoch:      FirstEpoch,
	})
	if err != nil {
		panic(err)
	}

	input := NewEventStore()

	config := LiteConfig()
	dagIndexer := &adapters.VectorToDagIndexer{Index: vecfc.NewIndex(crit, vecfc.LiteConfig())}
	lch := NewIndexed(store, input, dagIndexer, crit, config)

	extended := &TestConsensus{
		Indexed:     lch,
		blocks:      map[BlockKey]*BlockResult{},
		epochBlocks: map[idx.Epoch]idx.Frame{},
	}

	err = extended.Bootstrap(types.ConsensusCallbacks{
		BeginBlock: func(block *types.Block) types.BlockCallbacks {
			return types.BlockCallbacks{
				EndBlock: func() (sealEpoch *pos.Validators) {
					// track blocks
					key := BlockKey{
						Epoch: extended.store.GetEpoch(),
						Frame: extended.store.GetLastDecidedFrame() + 1,
					}
					extended.blocks[key] = &BlockResult{
						Event:      block.Event,
						Cheaters:   block.Cheaters,
						Validators: extended.store.GetValidators(),
					}
					// check that prev block exists
					if extended.lastBlock.Epoch != key.Epoch && key.Frame != 1 {
						panic("first frame must be 1")
					}
					extended.epochBlocks[key.Epoch]++
					extended.lastBlock = key
					if extended.applyBlock != nil {
						return extended.applyBlock(block)
					}
					return nil
				},
			}
		},
	})
	if err != nil {
		panic(err)
	}

	return extended, store, input, dagIndexer
}

func mutateValidators(validators *pos.Validators) *pos.Validators {
	r := rand.New(rand.NewSource(int64(validators.TotalWeight()))) // nolint:gosec
	builder := pos.NewBuilder()
	for _, vid := range validators.IDs() {
		stake := uint64(validators.Get(vid))*uint64(500+r.Intn(500))/1000 + 1
		builder.Set(vid, pos.Weight(stake))
	}
	return builder.Build()
}
