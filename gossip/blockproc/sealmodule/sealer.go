package sealmodule

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/consensus/native/idx"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/pos"
	"github.com/unicornultrafoundation/go-u2u/consensus/types"
	"github.com/unicornultrafoundation/go-u2u/gossip/blockproc"
	"github.com/unicornultrafoundation/go-u2u/native/iblockproc"
)

type U2UEpochsSealerModule struct{}

func New() *U2UEpochsSealerModule {
	return &U2UEpochsSealerModule{}
}

func (m *U2UEpochsSealerModule) Start(block iblockproc.BlockCtx, bs iblockproc.BlockState, es iblockproc.EpochState) blockproc.SealerProcessor {
	return &U2UEpochsSealer{
		block: block,
		es:    es,
		bs:    bs,
	}
}

type U2UEpochsSealer struct {
	block iblockproc.BlockCtx
	es    iblockproc.EpochState
	bs    iblockproc.BlockState
}

func (s *U2UEpochsSealer) EpochSealing() bool {
	sealEpoch := s.bs.EpochGas >= s.es.Rules.Epochs.MaxEpochGas
	sealEpoch = sealEpoch || (s.block.Time-s.es.EpochStart) >= s.es.Rules.Epochs.MaxEpochDuration
	sealEpoch = sealEpoch || s.bs.AdvanceEpochs > 0
	return sealEpoch || s.bs.EpochCheaters.Len() != 0
}

func (p *U2UEpochsSealer) Update(bs iblockproc.BlockState, es iblockproc.EpochState) {
	p.bs, p.es = bs, es
}

// SealEpoch is called after pre-internal transactions are executed
func (s *U2UEpochsSealer) SealEpoch() (iblockproc.BlockState, iblockproc.EpochState) {
	// Select new validators
	oldValidators := s.es.Validators
	builder := pos.NewBigBuilder()
	for v, profile := range s.bs.NextValidatorProfiles {
		builder.Set(v, profile.Weight)
	}
	newValidators := builder.Build()
	s.es.Validators = newValidators
	s.es.ValidatorProfiles = s.bs.NextValidatorProfiles.Copy()

	// Build new []ValidatorEpochState and []ValidatorBlockState
	newValidatorEpochStates := make([]iblockproc.ValidatorEpochState, newValidators.Len())
	newValidatorBlockStates := make([]iblockproc.ValidatorBlockState, newValidators.Len())
	for newValIdx := idx.Validator(0); newValIdx < newValidators.Len(); newValIdx++ {
		// default values
		newValidatorBlockStates[newValIdx] = iblockproc.ValidatorBlockState{
			Originated: new(big.Int),
		}
		// inherit validator's state from previous epoch, if any
		valID := newValidators.GetID(newValIdx)
		if !oldValidators.Exists(valID) {
			// new validator
			newValidatorBlockStates[newValIdx].LastBlock = s.block.Idx
			newValidatorBlockStates[newValIdx].LastOnlineTime = s.block.Time
			continue
		}
		oldValIdx := oldValidators.GetIdx(valID)
		newValidatorBlockStates[newValIdx] = s.bs.ValidatorStates[oldValIdx]
		newValidatorBlockStates[newValIdx].DirtyGasRefund = 0
		newValidatorBlockStates[newValIdx].Uptime = 0
		newValidatorEpochStates[newValIdx].GasRefund = s.bs.ValidatorStates[oldValIdx].DirtyGasRefund
		newValidatorEpochStates[newValIdx].PrevEpochEvent = s.bs.ValidatorStates[oldValIdx].LastEvent
	}
	s.es.ValidatorStates = newValidatorEpochStates
	s.bs.ValidatorStates = newValidatorBlockStates
	s.es.Validators = newValidators

	// dirty data becomes active
	s.es.PrevEpochStart = s.es.EpochStart
	s.es.EpochStart = s.block.Time
	if s.bs.DirtyRules != nil {
		s.es.Rules = s.bs.DirtyRules.Copy()
		s.bs.DirtyRules = nil
	}
	s.es.EpochStateRoot = s.bs.FinalizedStateRoot

	s.bs.EpochGas = 0
	s.bs.EpochCheaters = types.Cheaters{}
	s.bs.CheatersWritten = 0
	newEpoch := s.es.Epoch + 1
	s.es.Epoch = newEpoch

	if s.bs.AdvanceEpochs > 0 {
		s.bs.AdvanceEpochs--
	}

	return s.bs, s.es
}
