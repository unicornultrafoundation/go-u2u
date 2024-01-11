package gossip

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/unicornultrafoundation/go-hashgraph/hash"
	"github.com/unicornultrafoundation/go-hashgraph/native/dag"
	"github.com/unicornultrafoundation/go-hashgraph/native/idx"
	utypes "github.com/unicornultrafoundation/go-hashgraph/types"
	"github.com/unicornultrafoundation/go-hashgraph/utils/workers"
	"github.com/unicornultrafoundation/go-u2u/accounts"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/eth/protocols/snap"
	"github.com/unicornultrafoundation/go-u2u/event"
	notify "github.com/unicornultrafoundation/go-u2u/event"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/node"
	"github.com/unicornultrafoundation/go-u2u/p2p"
	"github.com/unicornultrafoundation/go-u2u/p2p/dnsdisc"
	"github.com/unicornultrafoundation/go-u2u/p2p/enode"
	"github.com/unicornultrafoundation/go-u2u/p2p/enr"
	"github.com/unicornultrafoundation/go-u2u/rpc"

	"github.com/unicornultrafoundation/go-u2u/ethapi"
	"github.com/unicornultrafoundation/go-u2u/eventcheck"
	"github.com/unicornultrafoundation/go-u2u/eventcheck/basiccheck"
	"github.com/unicornultrafoundation/go-u2u/eventcheck/epochcheck"
	"github.com/unicornultrafoundation/go-u2u/eventcheck/gaspowercheck"
	"github.com/unicornultrafoundation/go-u2u/eventcheck/heavycheck"
	"github.com/unicornultrafoundation/go-u2u/eventcheck/parentscheck"
	"github.com/unicornultrafoundation/go-u2u/evmcore"
	"github.com/unicornultrafoundation/go-u2u/gossip/blockproc"
	"github.com/unicornultrafoundation/go-u2u/gossip/blockproc/drivermodule"
	"github.com/unicornultrafoundation/go-u2u/gossip/blockproc/eventmodule"
	"github.com/unicornultrafoundation/go-u2u/gossip/blockproc/evmmodule"
	"github.com/unicornultrafoundation/go-u2u/gossip/blockproc/sealmodule"
	"github.com/unicornultrafoundation/go-u2u/gossip/blockproc/verwatcher"
	"github.com/unicornultrafoundation/go-u2u/gossip/emitter"
	"github.com/unicornultrafoundation/go-u2u/gossip/filters"
	"github.com/unicornultrafoundation/go-u2u/gossip/gasprice"
	"github.com/unicornultrafoundation/go-u2u/gossip/proclogger"
	snapsync "github.com/unicornultrafoundation/go-u2u/gossip/protocols/snap"
	"github.com/unicornultrafoundation/go-u2u/logger"
	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/utils/signers/gsignercache"
	"github.com/unicornultrafoundation/go-u2u/utils/txtime"
	"github.com/unicornultrafoundation/go-u2u/utils/wgmutex"
	"github.com/unicornultrafoundation/go-u2u/valkeystore"
	"github.com/unicornultrafoundation/go-u2u/vecmt"
)

type ServiceFeed struct {
	scope notify.SubscriptionScope

	newEpoch        notify.Feed
	newEmittedEvent notify.Feed
	newBlock        notify.Feed
	newLogs         notify.Feed
}

func (f *ServiceFeed) SubscribeNewEpoch(ch chan<- idx.Epoch) notify.Subscription {
	return f.scope.Track(f.newEpoch.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewEmitted(ch chan<- *native.EventPayload) notify.Subscription {
	return f.scope.Track(f.newEmittedEvent.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewBlock(ch chan<- evmcore.ChainHeadNotify) notify.Subscription {
	return f.scope.Track(f.newBlock.Subscribe(ch))
}

func (f *ServiceFeed) SubscribeNewLogs(ch chan<- []*types.Log) notify.Subscription {
	return f.scope.Track(f.newLogs.Subscribe(ch))
}

type BlockProc struct {
	SealerModule     blockproc.SealerModule
	TxListenerModule blockproc.TxListenerModule
	PreTxTransactor  blockproc.TxTransactor
	PostTxTransactor blockproc.TxTransactor
	EventsModule     blockproc.ConfirmedEventsModule
	EVMModule        blockproc.EVM
}

func DefaultBlockProc() BlockProc {
	return BlockProc{
		SealerModule:     sealmodule.New(),
		TxListenerModule: drivermodule.NewDriverTxListenerModule(),
		PreTxTransactor:  drivermodule.NewDriverTxPreTransactor(),
		PostTxTransactor: drivermodule.NewDriverTxTransactor(),
		EventsModule:     eventmodule.New(),
		EVMModule:        evmmodule.New(),
	}
}

// Service implements go-ethereum/node.Service interface.
type Service struct {
	config Config

	// server
	p2pServer *p2p.Server
	Name      string

	accountManager *accounts.Manager

	// application
	store               *Store
	engine              utypes.Consensus
	dagIndexer          *vecmt.Index
	engineMu            *sync.RWMutex
	emitters            []*emitter.Emitter
	txpool              TxPool
	heavyCheckReader    HeavyCheckReader
	gasPowerCheckReader GasPowerCheckReader
	checkers            *eventcheck.Checkers
	uniqueEventIDs      uniqueID

	// version watcher
	verWatcher *verwatcher.VerWarcher

	blockProcWg        sync.WaitGroup
	blockProcTasks     *workers.Workers
	blockProcTasksDone chan struct{}
	blockProcModules   BlockProc

	blockBusyFlag uint32
	eventBusyFlag uint32

	feed     ServiceFeed
	eventMux *event.TypeMux

	gpo *gasprice.Oracle

	// application protocol
	handler *handler

	u2uDialCandidates  enode.Iterator
	snapDialCandidates enode.Iterator

	EthAPI        *EthAPIBackend
	netRPCService *ethapi.PublicNetAPI

	procLogger *proclogger.Logger

	stopped   bool
	haltCheck func(oldEpoch, newEpoch idx.Epoch, time time.Time) bool

	tflusher PeriodicFlusher

	bootstrapping bool

	logger.Instance
}

func NewService(stack *node.Node, config Config, store *Store, blockProc BlockProc,
	engine utypes.Consensus, dagIndexer *vecmt.Index, newTxPool func(evmcore.StateReader) TxPool,
	haltCheck func(oldEpoch, newEpoch idx.Epoch, age time.Time) bool) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	svc, err := newService(config, store, blockProc, engine, dagIndexer, newTxPool)
	if err != nil {
		return nil, err
	}

	svc.p2pServer = stack.Server()
	svc.accountManager = stack.AccountManager()
	svc.eventMux = stack.EventMux()
	svc.EthAPI.SetExtRPCEnabled(stack.Config().ExtRPCEnabled())
	// Create the net API service
	svc.netRPCService = ethapi.NewPublicNetAPI(svc.p2pServer, store.GetRules().NetworkID)
	svc.haltCheck = haltCheck

	return svc, nil
}

func newService(config Config, store *Store, blockProc BlockProc, engine utypes.Consensus, dagIndexer *vecmt.Index, newTxPool func(evmcore.StateReader) TxPool) (*Service, error) {
	svc := &Service{
		config:             config,
		blockProcTasksDone: make(chan struct{}),
		Name:               fmt.Sprintf("Node-%d", rand.Int()),
		store:              store,
		engine:             engine,
		blockProcModules:   blockProc,
		dagIndexer:         dagIndexer,
		engineMu:           new(sync.RWMutex),
		uniqueEventIDs:     uniqueID{new(big.Int)},
		procLogger:         proclogger.NewLogger(),
		Instance:           logger.New("gossip-service"),
	}

	svc.blockProcTasks = workers.New(new(sync.WaitGroup), svc.blockProcTasksDone, 1)

	// load epoch DB
	svc.store.loadEpochStore(svc.store.GetEpoch())
	es := svc.store.getEpochStore(svc.store.GetEpoch())
	svc.dagIndexer.Reset(svc.store.GetValidators(), es.table.DagIndex, func(id hash.Event) dag.Event {
		return svc.store.GetEvent(id)
	})

	// load caches for mutable values to avoid race condition
	svc.store.GetBlockEpochState()
	svc.store.GetHighestLamport()
	svc.store.GetLastBVs()
	svc.store.GetLastEVs()
	svc.store.GetLlrState()
	svc.store.GetUpgradeHeights()
	svc.store.GetGenesisID()
	netVerStore := verwatcher.NewStore(store.table.NetworkVersion)
	netVerStore.GetNetworkVersion()
	netVerStore.GetMissedVersion()

	// create GPO
	svc.gpo = gasprice.NewOracle(svc.config.GPO)

	// create checkers
	net := store.GetRules()
	txSigner := gsignercache.Wrap(types.LatestSignerForChainID(new(big.Int).SetUint64(net.NetworkID)))
	svc.heavyCheckReader.Store = store
	svc.heavyCheckReader.Pubkeys.Store(readEpochPubKeys(svc.store, svc.store.GetEpoch()))                                          // read pub keys of current epoch from DB
	svc.gasPowerCheckReader.Ctx.Store(NewGasPowerContext(svc.store, svc.store.GetValidators(), svc.store.GetEpoch(), net.Economy)) // read gaspower check data from DB
	svc.checkers = makeCheckers(config.HeavyCheck, txSigner, &svc.heavyCheckReader, &svc.gasPowerCheckReader, svc.store)

	// create tx pool
	stateReader := svc.GetEvmStateReader()
	svc.txpool = newTxPool(stateReader)

	// init dialCandidates
	dnsclient := dnsdisc.NewClient(dnsdisc.Config{})
	var err error
	svc.u2uDialCandidates, err = dnsclient.NewIterator(config.U2UDiscoveryURLs...)
	if err != nil {
		return nil, err
	}
	svc.snapDialCandidates, err = dnsclient.NewIterator(config.SnapDiscoveryURLs...)
	if err != nil {
		return nil, err
	}

	// create protocol manager
	svc.handler, err = newHandler(handlerConfig{
		config:   config,
		notifier: &svc.feed,
		txpool:   svc.txpool,
		engineMu: svc.engineMu,
		checkers: svc.checkers,
		s:        store,
		process: processCallback{
			Event: func(event *native.EventPayload) error {
				done := svc.procLogger.EventConnectionStarted(event, false)
				defer done()
				return svc.processEvent(event)
			},
			SwitchEpochTo:    svc.SwitchEpochTo,
			PauseEvmSnapshot: svc.PauseEvmSnapshot,
			BVs:              svc.ProcessBlockVotes,
			BR:               svc.ProcessFullBlockRecord,
			EV:               svc.ProcessEpochVote,
			ER:               svc.ProcessFullEpochRecord,
		},
	})
	if err != nil {
		return nil, err
	}

	rpc.SetExecutionTimeLimit(config.RPCTimeout)

	// create API backend
	svc.EthAPI = &EthAPIBackend{false, svc, stateReader, txSigner, config.AllowUnprotectedTxs}
	if svc.EthAPI.allowUnprotectedTxs {
		log.Info("Unprotected transactions allowed")
	}

	svc.verWatcher = verwatcher.New(netVerStore)
	svc.tflusher = svc.makePeriodicFlusher()

	return svc, nil
}

// makeCheckers builds event checkers
func makeCheckers(heavyCheckCfg heavycheck.Config, txSigner types.Signer, heavyCheckReader *HeavyCheckReader, gasPowerCheckReader *GasPowerCheckReader, store *Store) *eventcheck.Checkers {
	// create signatures checker
	heavyCheck := heavycheck.New(heavyCheckCfg, heavyCheckReader, txSigner)

	// create gaspower checker
	gaspowerCheck := gaspowercheck.New(gasPowerCheckReader)

	return &eventcheck.Checkers{
		Basiccheck:    basiccheck.New(),
		Epochcheck:    epochcheck.New(store),
		Parentscheck:  parentscheck.New(),
		Heavycheck:    heavyCheck,
		Gaspowercheck: gaspowerCheck,
	}
}

// makePeriodicFlusher makes PeriodicFlusher
func (s *Service) makePeriodicFlusher() PeriodicFlusher {
	// Normally the diffs are committed by message processing. Yet, since we have async EVM snapshots generation,
	// we need to periodically commit its data
	return PeriodicFlusher{
		period: 10 * time.Millisecond,
		callback: PeriodicFlusherCallaback{
			busy: func() bool {
				// try to lock engineMu/blockProcWg pair as rarely as possible to not hurt
				// events/blocks pipeline concurrency
				return atomic.LoadUint32(&s.eventBusyFlag) != 0 || atomic.LoadUint32(&s.blockBusyFlag) != 0
			},
			commitNeeded: func() bool {
				// use slightly higher size threshold to avoid locking the mutex/wg pair and hurting events/blocks concurrency
				// PeriodicFlusher should mostly commit only data generated by async EVM snapshots generation
				return s.store.isCommitNeeded(1200, 1000)
			},
			commit: func() {
				s.engineMu.Lock()
				defer s.engineMu.Unlock()
				// Note: blockProcWg.Wait() is already called by s.commit
				if s.store.isCommitNeeded(1200, 1000) {
					s.commit(false)
				}
			},
		},
		wg:   sync.WaitGroup{},
		quit: make(chan struct{}),
	}
}

func (s *Service) EmitterWorld(signer valkeystore.SignerI) emitter.World {
	return emitter.World{
		External: &emitterWorld{
			emitterWorldProc: emitterWorldProc{s},
			emitterWorldRead: emitterWorldRead{s.store},
			WgMutex:          wgmutex.New(s.engineMu, &s.blockProcWg),
		},
		TxPool:   s.txpool,
		Signer:   signer,
		TxSigner: s.EthAPI.signer,
	}
}

// RegisterEmitter must be called before service is started
func (s *Service) RegisterEmitter(em *emitter.Emitter) {
	txtime.Enabled = true // enable tracking of tx times
	s.emitters = append(s.emitters, em)
}

// MakeProtocols constructs the P2P protocol definitions for `u2u`.
func MakeProtocols(svc *Service, backend *handler, disc enode.Iterator) []p2p.Protocol {
	protocols := make([]p2p.Protocol, len(ProtocolVersions))
	for i, version := range ProtocolVersions {
		version := version // Closure

		protocols[i] = p2p.Protocol{
			Name:    ProtocolName,
			Version: version,
			Length:  protocolLengths[version],
			Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
				// wait until handler has started
				backend.started.Wait()
				peer := newPeer(version, p, rw, backend.config.Protocol.PeerCache)
				defer peer.Close()

				select {
				case <-backend.quitSync:
					return p2p.DiscQuitting
				default:
					backend.wg.Add(1)
					defer backend.wg.Done()
					return backend.handle(peer)
				}
			},
			NodeInfo: func() interface{} {
				return backend.NodeInfo()
			},
			PeerInfo: func(id enode.ID) interface{} {
				if p := backend.peers.Peer(id.String()); p != nil {
					return p.Info()
				}
				return nil
			},
			Attributes:     []enr.Entry{currentENREntry(svc)},
			DialCandidates: disc,
		}
	}
	return protocols
}

// Protocols returns protocols the service can communicate on.
func (s *Service) Protocols() []p2p.Protocol {
	protos := append(
		MakeProtocols(s, s.handler, s.u2uDialCandidates),
		snap.MakeProtocols((*snapHandler)(s.handler), s.snapDialCandidates)...)
	return protos
}

// APIs returns api methods the service wants to expose on rpc channels.
func (s *Service) APIs() []rpc.API {
	apis := ethapi.GetAPIs(s.EthAPI)

	apis = append(apis, []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.EthAPI, s.config.FilterAPI),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   snapsync.NewPublicDownloaderAPI(s.handler.snapLeecher, s.eventMux),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)

	return apis
}

// Start method invoked when the node is ready to start the service.
func (s *Service) Start() error {
	s.gpo.Start(&GPOBackend{s.store, s.txpool})
	// start tflusher before starting snapshots generation
	s.tflusher.Start()
	// start snapshots generation
	if s.store.evm.IsEvmSnapshotPaused() && !s.config.AllowSnapsync {
		return errors.New("cannot halt snapsync and start fullsync")
	}
	s.RecoverEVM()
	root := s.store.GetBlockState().FinalizedStateRoot
	if !s.store.evm.HasStateDB(root) {
		if !s.config.AllowSnapsync {
			return errors.New("fullsync isn't possible because state root is missing")
		}
		root = hash.Zero
	}
	_ = s.store.GenerateSnapshotAt(common.Hash(root), true)

	// start blocks processor
	s.blockProcTasks.Start(1)

	// start p2p
	StartENRUpdater(s, s.p2pServer.LocalNode())
	s.handler.Start(s.p2pServer.MaxPeers)

	// start emitters
	for _, em := range s.emitters {
		em.Start()
	}

	s.verWatcher.Start()

	if s.haltCheck != nil && s.haltCheck(s.store.GetEpoch(), s.store.GetEpoch(), s.store.GetBlockState().LastBlock.Time.Time()) {
		// halt syncing
		s.stopped = true
	}

	return nil
}

// WaitBlockEnd waits until parallel block processing is complete (if any)
func (s *Service) WaitBlockEnd() {
	s.blockProcWg.Wait()
}

// Stop method invoked when the node terminates the service.
func (s *Service) Stop() error {
	defer log.Info("U2U service stopped")
	s.verWatcher.Stop()
	for _, em := range s.emitters {
		em.Stop()
	}

	// Stop all the peer-related stuff first.
	s.u2uDialCandidates.Close()
	s.snapDialCandidates.Close()

	s.handler.Stop()
	s.feed.scope.Close()
	s.eventMux.Stop()
	s.gpo.Stop()
	// it's safe to stop tflusher only before locking engineMu
	s.tflusher.Stop()

	// flush the state at exit, after all the routines stopped
	s.engineMu.Lock()
	defer s.engineMu.Unlock()
	s.stopped = true

	s.blockProcWg.Wait()
	close(s.blockProcTasksDone)
	s.store.evm.Flush(s.store.GetBlockState())
	return s.store.Commit()
}

// AccountManager return node's account manager
func (s *Service) AccountManager() *accounts.Manager {
	return s.accountManager
}
