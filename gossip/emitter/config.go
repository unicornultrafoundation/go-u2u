package emitter

import (
	"math/rand"
	"time"

	"github.com/unicornultrafoundation/go-u2u/helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/params"

	"github.com/unicornultrafoundation/go-u2u/native/validatorpk"
	"github.com/unicornultrafoundation/go-u2u/u2u"
)

// EmitIntervals is the configuration of emit intervals.
type EmitIntervals struct {
	Min                        time.Duration
	Max                        time.Duration
	Confirming                 time.Duration // emit time when there's no txs to originate, but at least 1 tx to confirm
	ParallelInstanceProtection time.Duration
	DoublesignProtection       time.Duration
}

type ValidatorConfig struct {
	ID     idx.ValidatorID
	PubKey validatorpk.PubKey
}

type FileConfig struct {
	Path     string
	SyncMode bool
}

// Config is the configuration of events emitter.
type Config struct {
	VersionToPublish string

	Validator ValidatorConfig

	EmitIntervals EmitIntervals // event emission intervals

	MaxTxsPerAddress int

	MaxParents idx.Event

	// thresholds on GasLeft
	LimitedTpsThreshold uint64
	NoTxsThreshold      uint64
	EmergencyThreshold  uint64

	TxsCacheInvalidation time.Duration

	PrevEmittedEventFile FileConfig
	PrevBlockVotesFile   FileConfig
	PrevEpochVoteFile    FileConfig
}

// DefaultConfig returns the default configurations for the events emitter.
func DefaultConfig() Config {
	return Config{
		VersionToPublish: params.VersionWithMeta(),

		EmitIntervals: EmitIntervals{
			Min:                        150 * time.Millisecond,
			Max:                        1 * time.Second,
			Confirming:                 170 * time.Millisecond,
			DoublesignProtection:       27 * time.Minute, // should be greater than MaxEmitInterval
			ParallelInstanceProtection: 1 * time.Minute,
		},

		MaxTxsPerAddress: TxTurnNonces,

		MaxParents: 0,

		LimitedTpsThreshold: u2u.DefaultEventGas * 120,
		NoTxsThreshold:      u2u.DefaultEventGas * 30,
		EmergencyThreshold:  u2u.DefaultEventGas * 5,

		TxsCacheInvalidation: 200 * time.Millisecond,
	}
}

// RandomizeEmitTime and return new config
func (cfg EmitIntervals) RandomizeEmitTime(r *rand.Rand) EmitIntervals {
	config := cfg
	// value = value - 0.1 * value + 0.1 * random value
	if config.Max > 10 {
		config.Max = config.Max - config.Max/10 + time.Duration(r.Int63n(int64(config.Max/10)))
	}
	// value = value + 0.33 * random value
	if config.DoublesignProtection > 3 {
		config.DoublesignProtection = config.DoublesignProtection + time.Duration(r.Int63n(int64(config.DoublesignProtection/3)))
	}
	return config
}

// FakeConfig returns the testing configurations for the events emitter.
func FakeConfig(num idx.Validator) Config {
	cfg := DefaultConfig()
	cfg.EmitIntervals.Max = 10 * time.Second // don't wait long in fakenet
	cfg.EmitIntervals.DoublesignProtection = cfg.EmitIntervals.Max / 2
	if num <= 1 {
		// disable self-fork protection if fakenet 1/1
		cfg.EmitIntervals.DoublesignProtection = 0
	}
	return cfg
}
