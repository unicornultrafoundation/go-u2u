package evmmodule

import (
	"math"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/evmcore"
	"github.com/unicornultrafoundation/go-u2u/gossip/blockproc"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/native/iblockproc"
	"github.com/unicornultrafoundation/go-u2u/params"
	"github.com/unicornultrafoundation/go-u2u/u2u"
	"github.com/unicornultrafoundation/go-u2u/utils"
)

var SfcPrecompiles = []common.Address{
	common.HexToAddress("0xFC00FACE00000000000000000000000000000000"),
	common.HexToAddress("0xD100ae0000000000000000000000000000000000"),
	common.HexToAddress("0xd100A01E00000000000000000000000000000000"),
	common.HexToAddress("0x6CA548f6DF5B540E72262E935b6Fe3e72cDd68C9"),
	common.HexToAddress("0xFC01fACE00000000000000000000000000000000"), // SFCLib
}

type EVMModule struct{}

func New() *EVMModule {
	return &EVMModule{}
}

func (p *EVMModule) Start(block iblockproc.BlockCtx, statedb *state.StateDB, sfcStatedb *state.StateDB, reader evmcore.DummyChain,
	onNewLog func(*types.Log), net u2u.Rules, evmCfg *params.ChainConfig) blockproc.EVMProcessor {
	var prevBlockHash common.Hash
	if block.Idx != 0 {
		prevBlockHash = reader.GetHeader(common.Hash{}, uint64(block.Idx-1)).Hash
	}
	processor := &U2UEVMProcessor{
		block:         block,
		reader:        reader,
		statedb:       statedb,
		onNewLog:      onNewLog,
		net:           net,
		evmCfg:        evmCfg,
		blockIdx:      utils.U64toBig(uint64(block.Idx)),
		prevBlockHash: prevBlockHash,
	}
	if !common.IsNilInterface(sfcStatedb) {
		processor.sfcStateDb = sfcStatedb
	}
	return processor
}

type U2UEVMProcessor struct {
	block      iblockproc.BlockCtx
	reader     evmcore.DummyChain
	statedb    *state.StateDB
	sfcStateDb *state.StateDB
	onNewLog   func(*types.Log)
	net        u2u.Rules
	evmCfg     *params.ChainConfig

	blockIdx      *big.Int
	prevBlockHash common.Hash

	gasUsed uint64

	incomingTxs types.Transactions
	skippedTxs  []uint32
	receipts    types.Receipts
}

func (p *U2UEVMProcessor) evmBlockWith(txs types.Transactions) *evmcore.EvmBlock {
	baseFee := p.net.Economy.MinGasPrice
	if !p.net.Upgrades.London {
		baseFee = nil
	}
	h := &evmcore.EvmHeader{
		Number:       p.blockIdx,
		Hash:         common.Hash(p.block.Atropos),
		ParentHash:   p.prevBlockHash,
		Root:         common.Hash{},
		SfcStateRoot: common.Hash{},
		Time:         p.block.Time,
		Coinbase:     common.Address{},
		GasLimit:     math.MaxUint64,
		GasUsed:      p.gasUsed,
		BaseFee:      baseFee,
	}

	return evmcore.NewEvmBlock(h, txs)
}

func (p *U2UEVMProcessor) Execute(txs types.Transactions) types.Receipts {
	evmProcessor := evmcore.NewStateProcessor(p.evmCfg, p.reader)
	txsOffset := uint(len(p.incomingTxs))

	// Process txs
	evmBlock := p.evmBlockWith(txs)
	receipts, _, skipped, err := evmProcessor.Process(evmBlock, p.statedb, p.sfcStateDb, u2u.DefaultVMConfig, &p.gasUsed, func(l *types.Log, _ *state.StateDB) {
		// Note: l.Index is properly set before
		l.TxIndex += txsOffset
		p.onNewLog(l)
	})
	if err != nil {
		log.Crit("EVM internal error", "err", err)
	}

	if txsOffset > 0 {
		for i, n := range skipped {
			skipped[i] = n + uint32(txsOffset)
		}
		for _, r := range receipts {
			r.TransactionIndex += txsOffset
		}
	}

	p.incomingTxs = append(p.incomingTxs, txs...)
	p.skippedTxs = append(p.skippedTxs, skipped...)
	p.receipts = append(p.receipts, receipts...)

	return receipts
}

func (p *U2UEVMProcessor) Finalize() (evmBlock *evmcore.EvmBlock, skippedTxs []uint32, receipts types.Receipts) {
	evmBlock = p.evmBlockWith(
		// Filter skipped transactions. Receipts are filtered already
		native.FilterSkippedTxs(p.incomingTxs, p.skippedTxs),
	)
	skippedTxs = p.skippedTxs
	receipts = p.receipts

	// Get state root
	log.Trace("U2UEVMProcessor.Finalize after block", "block", p.block.Idx)
	newStateHash, err := p.statedb.Commit(true)
	if err != nil {
		log.Crit("Failed to commit state", "err", err)
	}
	evmBlock.Root = newStateHash
	if p.sfcStateDb != nil {
		log.Trace("Separate two commit logs when U2UEVMProcessor.Finalize after block")
		newSfcStateHash, err := p.sfcStateDb.Commit(true)
		if err != nil {
			log.Crit("Failed to commit sfc state", "err", err)
		}
		evmBlock.SfcStateRoot = newSfcStateHash

		// extra dual-state verification
		if newSfcStateHash.Cmp(types.EmptyRootHash) == 0 {
			log.Error("SFC state is empty now", "block", p.block.Idx)
		}
		for _, addr := range SfcPrecompiles {
			original := p.statedb.GetStorageRoot(addr)
			sfc := p.sfcStateDb.GetStorageRoot(addr)
			if original.Cmp(sfc) != 0 {
				log.Error("U2UEVMProcessor.Finalize: SFC corrupted after applying block", "height", p.block.Idx,
					"addr", addr, "original", original.Hex(), "sfc", sfc.Hex())
				common.SendInterrupt()
			}
		}
	}
	return
}
