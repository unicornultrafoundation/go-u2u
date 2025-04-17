package consensus

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"

	"github.com/unicornultrafoundation/go-u2u/helios/consensus/dagidx"
	"github.com/unicornultrafoundation/go-u2u/helios/hash"
	"github.com/unicornultrafoundation/go-u2u/helios/native/dag"
	"github.com/unicornultrafoundation/go-u2u/helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/helios/native/pos"
	"github.com/unicornultrafoundation/go-u2u/helios/types"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb/flushable"
)

var _ types.Consensus = (*Indexed)(nil)

// Indexed performs events ordering and detects cheaters
// It's a wrapper around Orderer, which adds features which might potentially be application-specific:
// confirmed events traversal, DAG index updates and cheaters detection.
// Use this structure if need a general-purpose consensus. Instead, use lower-level consensus.Orderer.
type Indexed struct {
	*Consensus
	dagIndexer    DagIndexer
	uniqueDirtyID uniqueID
}

type DagIndexer interface {
	dagidx.VectorClock
	dagidx.ForklessCause

	Add(dag.Event) error
	Flush()
	DropNotFlushed()

	Reset(validators *pos.Validators, db u2udb.FlushableKVStore, getEvent func(hash.Event) dag.Event) error
}

// New creates Indexed instance.
func NewIndexed(store *Store, input EventSource, dagIndexer DagIndexer, crit func(error), config Config) *Indexed {
	p := &Indexed{
		Consensus:     NewConsensus(store, input, dagIndexer, crit, config),
		dagIndexer:    dagIndexer,
		uniqueDirtyID: uniqueID{new(big.Int)},
	}

	return p
}

// Build fills consensus-related fields: Frame, IsRoot
// returns error if event should be dropped
func (p *Indexed) Build(e dag.MutableEvent) error {
	e.SetID(p.uniqueDirtyID.sample())

	defer p.dagIndexer.DropNotFlushed()
	err := p.dagIndexer.Add(e)
	if err != nil {
		return err
	}

	return p.Consensus.Build(e)
}

// Process takes event into processing.
// Event order matter: parents first.
// All the event checkers must be launched.
// Process is not safe for concurrent use.
func (p *Indexed) Process(e dag.Event) (err error) {
	defer p.dagIndexer.DropNotFlushed()
	err = p.dagIndexer.Add(e)
	if err != nil {
		return err
	}

	err = p.Consensus.Process(e)
	if err != nil {
		return err
	}
	p.dagIndexer.Flush()
	return nil
}

func (p *Indexed) Bootstrap(callback types.ConsensusCallbacks) error {
	base := p.Consensus.OrdererCallbacks()
	ordererCallbacks := OrdererCallbacks{
		ApplyEvent: base.ApplyEvent,
		EpochDBLoaded: func(epoch idx.Epoch) {
			if base.EpochDBLoaded != nil {
				base.EpochDBLoaded(epoch)
			}
			p.dagIndexer.Reset(p.store.GetValidators(), flushable.Wrap(p.store.epochTable.VectorIndex), p.input.GetEvent)
		},
	}
	return p.Consensus.BootstrapWithOrderer(callback, ordererCallbacks)
}

type uniqueID struct {
	counter *big.Int
}

func (u *uniqueID) sample() [24]byte {
	u.counter = u.counter.Add(u.counter, common.Big1)
	var id [24]byte
	copy(id[:], u.counter.Bytes())
	return id
}
