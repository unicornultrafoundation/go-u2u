package u2u

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/unicornultrafoundation/go-helios/native/idx"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/native"
	ethparams "github.com/unicornultrafoundation/go-u2u/params"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/constant_manager"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/driver"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/driverauth"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/evmwriter"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/sfc"
)

const (
	MainNetworkID   uint64 = 39
	TestNetworkID   uint64 = 2484
	FakeNetworkID   uint64 = 4439
	DefaultEventGas uint64 = 28000
	berlinBit              = 1 << 0
	londonBit              = 1 << 1
	llrBit                 = 1 << 2
)

var DefaultVMConfig = vm.Config{
	StatePrecompiles: map[common.Address]vm.PrecompiledStateContract{
		evmwriter.ContractAddress: &evmwriter.PreCompiledContract{},
	},
	SfcPrecompiles: map[common.Address]vm.PrecompiledStateContract{
		common.HexToAddress("0xFC00FACE00000000000000000000000000000000"): &sfc.SfcPrecompile{},
		common.HexToAddress("0xD100ae0000000000000000000000000000000000"): &driverauth.DriverAuthPrecompile{},
		common.HexToAddress("0xd100A01E00000000000000000000000000000000"): &driver.DriverPrecompile{},
		common.HexToAddress("0x6CA548f6DF5B540E72262E935b6Fe3e72cDd68C9"): &constant_manager.ConstantManagerPrecompile{},
	},
}

type RulesRLP struct {
	Name      string
	NetworkID uint64

	// Graph options
	Dag DagRules

	// Epochs options
	Epochs EpochsRules

	// Blockchain options
	Blocks BlocksRules

	// Economy options
	Economy EconomyRules

	Upgrades Upgrades `rlp:"-"`
}

// Rules describes u2u net.
// Note keep track of all the non-copiable variables in Copy()
type Rules RulesRLP

// GasPowerRules defines gas power rules in the consensus.
type GasPowerRules struct {
	AllocPerSec        uint64
	MaxAllocPeriod     native.Timestamp
	StartupAllocPeriod native.Timestamp
	MinStartupGas      uint64
}

type GasRulesRLPV1 struct {
	MaxEventGas  uint64
	EventGas     uint64
	ParentGas    uint64
	ExtraDataGas uint64
	// Post-LLR fields
	BlockVotesBaseGas    uint64
	BlockVoteGas         uint64
	EpochVoteGas         uint64
	MisbehaviourProofGas uint64
}

type GasRules GasRulesRLPV1

type EpochsRules struct {
	MaxEpochGas      uint64
	MaxEpochDuration native.Timestamp
}

// DagRules of Helios DAG (directed acyclic graph).
type DagRules struct {
	MaxParents     idx.Event
	MaxFreeParents idx.Event // maximum number of parents with no gas cost
	MaxExtraData   uint32
}

// BlocksMissed is information about missed blocks from a staker
type BlocksMissed struct {
	BlocksNum idx.Block
	Period    native.Timestamp
}

// EconomyRules contains economy constants
type EconomyRules struct {
	BlockMissedSlack idx.Block

	Gas GasRules

	MinGasPrice *big.Int

	ShortGasPower GasPowerRules
	LongGasPower  GasPowerRules
}

// BlocksRules contains blocks constants
type BlocksRules struct {
	MaxBlockGas             uint64 // technical hard limit, gas is mostly governed by gas power allocation
	MaxEmptyBlockSkipPeriod native.Timestamp
}

type Upgrades struct {
	Berlin bool
	London bool
	Llr    bool
}

type UpgradeHeight struct {
	Upgrades Upgrades
	Height   idx.Block
}

// EvmChainConfig returns ChainConfig for transactions signing and execution
func (r Rules) EvmChainConfig(hh []UpgradeHeight) *ethparams.ChainConfig {
	cfg := *ethparams.AllProtocolChanges
	cfg.ChainID = new(big.Int).SetUint64(r.NetworkID)
	cfg.BerlinBlock = nil
	cfg.LondonBlock = nil
	for i, h := range hh {
		height := new(big.Int)
		if i > 0 {
			height.SetUint64(uint64(h.Height))
		}
		if cfg.BerlinBlock == nil && h.Upgrades.Berlin {
			cfg.BerlinBlock = height
		}
		if !h.Upgrades.Berlin {
			cfg.BerlinBlock = nil
		}

		if cfg.LondonBlock == nil && h.Upgrades.London {
			cfg.LondonBlock = height
		}
		if !h.Upgrades.London {
			cfg.LondonBlock = nil
		}
	}
	return &cfg
}

func MainNetRules() Rules {
	return Rules{
		Name:      "main",
		NetworkID: MainNetworkID,
		Dag:       DefaultDagRules(),
		Epochs:    DefaultEpochsRules(),
		Economy:   DefaultEconomyRules(),
		Blocks: BlocksRules{
			MaxBlockGas:             20500000,
			MaxEmptyBlockSkipPeriod: native.Timestamp(1 * time.Second),
		},
		Upgrades: Upgrades{
			Berlin: true,
			London: true,
			Llr:    true,
		},
	}
}

func TestNetRules() Rules {
	return Rules{
		Name:      "test",
		NetworkID: TestNetworkID,
		Dag:       DefaultDagRules(),
		Epochs:    DefaultEpochsRules(),
		Economy:   DefaultEconomyRules(),
		Blocks: BlocksRules{
			MaxBlockGas:             20500000,
			MaxEmptyBlockSkipPeriod: native.Timestamp(1 * time.Second),
		},
		Upgrades: Upgrades{
			Berlin: true,
			London: true,
			Llr:    true,
		},
	}
}

func FakeNetRules() Rules {
	return Rules{
		Name:      "fake",
		NetworkID: FakeNetworkID,
		Dag:       DefaultDagRules(),
		Epochs:    FakeNetEpochsRules(),
		Economy:   FakeEconomyRules(),
		Blocks: BlocksRules{
			MaxBlockGas:             20500000,
			MaxEmptyBlockSkipPeriod: native.Timestamp(3 * time.Second),
		},
		Upgrades: Upgrades{
			Berlin: true,
			London: true,
			Llr:    true,
		},
	}
}

// DefaultEconomyRules returns mainnet economy
func DefaultEconomyRules() EconomyRules {
	return EconomyRules{
		BlockMissedSlack: 50,
		Gas:              DefaultGasRules(),
		MinGasPrice:      big.NewInt(1e9),
		ShortGasPower:    DefaultShortGasPowerRules(),
		LongGasPower:     DefaulLongGasPowerRules(),
	}
}

// FakeEconomyRules returns fakenet economy
func FakeEconomyRules() EconomyRules {
	cfg := DefaultEconomyRules()
	cfg.ShortGasPower = FakeShortGasPowerRules()
	cfg.LongGasPower = FakeLongGasPowerRules()
	return cfg
}

func DefaultDagRules() DagRules {
	return DagRules{
		MaxParents:     10,
		MaxFreeParents: 3,
		MaxExtraData:   128,
	}
}

func DefaultEpochsRules() EpochsRules {
	return EpochsRules{
		MaxEpochGas:      300000000,
		MaxEpochDuration: native.Timestamp(7 * time.Minute),
	}
}

func DefaultGasRules() GasRules {
	return GasRules{
		MaxEventGas:          10000000 + DefaultEventGas,
		EventGas:             DefaultEventGas,
		ParentGas:            2400,
		ExtraDataGas:         25,
		BlockVotesBaseGas:    1024,
		BlockVoteGas:         512,
		EpochVoteGas:         1536,
		MisbehaviourProofGas: 71536,
	}
}

func FakeNetEpochsRules() EpochsRules {
	cfg := DefaultEpochsRules()
	cfg.MaxEpochGas /= 5
	cfg.MaxEpochDuration = native.Timestamp(5 * time.Second)
	return cfg
}

// DefaulLongGasPowerRules is long-window config
func DefaulLongGasPowerRules() GasPowerRules {
	return GasPowerRules{
		AllocPerSec:        100 * DefaultEventGas,
		MaxAllocPeriod:     native.Timestamp(60 * time.Minute),
		StartupAllocPeriod: native.Timestamp(5 * time.Second),
		MinStartupGas:      DefaultEventGas * 20,
	}
}

// DefaultShortGasPowerRules is short-window config
func DefaultShortGasPowerRules() GasPowerRules {
	// 2x faster allocation rate, 6x lower max accumulated gas power
	cfg := DefaulLongGasPowerRules()
	cfg.AllocPerSec *= 2
	cfg.StartupAllocPeriod /= 2
	cfg.MaxAllocPeriod /= 2 * 6
	return cfg
}

// FakeLongGasPowerRules is fake long-window config
func FakeLongGasPowerRules() GasPowerRules {
	config := DefaulLongGasPowerRules()
	config.AllocPerSec *= 1000
	return config
}

// FakeShortGasPowerRules is fake short-window config
func FakeShortGasPowerRules() GasPowerRules {
	config := DefaultShortGasPowerRules()
	config.AllocPerSec *= 1000
	return config
}

func (r Rules) Copy() Rules {
	cp := r
	cp.Economy.MinGasPrice = new(big.Int).Set(r.Economy.MinGasPrice)
	return cp
}

func (r Rules) String() string {
	b, _ := json.Marshal(&r)
	return string(b)
}
