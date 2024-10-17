package snapleecher

import (
	"errors"
	"fmt"
	"sync"

	u2u "github.com/unicornultrafoundation/go-u2u"
	"github.com/unicornultrafoundation/go-u2u/core/rawdb"
	"github.com/unicornultrafoundation/go-u2u/eth/protocols/snap"
	"github.com/unicornultrafoundation/go-u2u/ethdb"
	"github.com/unicornultrafoundation/go-u2u/trie"
)

var (
	MaxStateFetch = 384 // Amount of node state values to allow fetching per request
)

var (
	errCancelStateFetch = errors.New("state data download canceled (requested)")
)

type Leecher struct {
	peers *peerSet // Set of active peers from which download can proceed
	// Callbacks
	dropPeer peerDropFn // Drops a peer for misbehaving

	stateDB    ethdb.Database  // Database to state sync into (and deduplicate via)
	stateBloom *trie.SyncBloom // Bloom filter for fast trie node and contract code existence checks
	SnapSyncer *snap.Syncer    // TODO(karalabe): make private! hack for now

	stateSyncStart chan *stateSync
	trackStateReq  chan *stateReq
	stateCh        chan dataPack // Channel receiving inbound node state data

	// Statistics
	syncStatsState stateSyncStats
	syncStatsLock  sync.RWMutex // Lock protecting the sync stats fields

	quitCh chan struct{} // Quit channel to signal termination
}

// New creates a new downloader to fetch hashes and blocks from remote peers.
func New(stateDb ethdb.Database, stateBloom *trie.SyncBloom, dropPeer peerDropFn) *Leecher {
	d := &Leecher{
		stateDB:        stateDb,
		stateBloom:     stateBloom,
		dropPeer:       dropPeer,
		peers:          newPeerSet(),
		quitCh:         make(chan struct{}),
		stateCh:        make(chan dataPack),
		SnapSyncer:     snap.NewSyncer(stateDb),
		stateSyncStart: make(chan *stateSync),
		syncStatsState: stateSyncStats{
			processed: rawdb.ReadFastTrieProgress(stateDb),
		},
		trackStateReq: make(chan *stateReq),
	}
	go d.stateFetcher()
	return d
}

// DeliverSnapPacket is invoked from a peer's message handler when it transmits a
// data packet for the local node to consume.
func (d *Leecher) DeliverSnapPacket(peer *snap.Peer, packet snap.Packet) error {
	switch packet := packet.(type) {
	case *snap.AccountRangePacket:
		hashes, accounts, err := packet.Unpack()
		if err != nil {
			return err
		}
		return d.SnapSyncer.OnAccounts(peer, packet.ID, hashes, accounts, packet.Proof)

	case *snap.StorageRangesPacket:
		hashset, slotset := packet.Unpack()
		return d.SnapSyncer.OnStorage(peer, packet.ID, hashset, slotset, packet.Proof)

	case *snap.ByteCodesPacket:
		return d.SnapSyncer.OnByteCodes(peer, packet.ID, packet.Codes)

	case *snap.TrieNodesPacket:
		return d.SnapSyncer.OnTrieNodes(peer, packet.ID, packet.Nodes)

	default:
		return fmt.Errorf("unexpected snap packet type: %T", packet)
	}
}

// Progress retrieves the synchronisation boundaries, specifically the origin
// block where synchronisation started at (may have failed/suspended); the block
// or header sync is currently at; and the latest known block which the sync targets.
//
// In addition, during the state download phase of fast synchronisation the number
// of processed and the total number of known states are also returned. Otherwise
// these are zero.
func (d *Leecher) Progress() u2u.SyncProgress {
	// Lock the current stats and return the progress
	d.syncStatsLock.RLock()
	defer d.syncStatsLock.RUnlock()

	return u2u.SyncProgress{
		PulledStates: d.syncStatsState.processed,
		KnownStates:  d.syncStatsState.processed + d.syncStatsState.pending,
	}
}
