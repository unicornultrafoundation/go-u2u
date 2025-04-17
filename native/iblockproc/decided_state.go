package iblockproc

import (
	"crypto/sha256"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/helios/hash"
	"github.com/unicornultrafoundation/go-u2u/helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/helios/native/pos"
	"github.com/unicornultrafoundation/go-u2u/helios/types"
	"github.com/unicornultrafoundation/go-u2u/rlp"

	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/u2u"
)

type ValidatorBlockState struct {
	LastEvent        EventInfo
	Uptime           native.Timestamp
	LastOnlineTime   native.Timestamp
	LastGasPowerLeft native.GasPowerLeft
	LastBlock        idx.Block
	DirtyGasRefund   uint64
	Originated       *big.Int
}

type EventInfo struct {
	ID           hash.Event
	GasPowerLeft native.GasPowerLeft
	Time         native.Timestamp
}

type ValidatorEpochState struct {
	GasRefund      uint64
	PrevEpochEvent EventInfo
}

type BlockCtx struct {
	Idx     idx.Block
	Time    native.Timestamp
	Atropos hash.Event
}

type BlockState struct {
	LastBlock          BlockCtx
	FinalizedStateRoot hash.Hash

	EpochGas        uint64
	EpochCheaters   types.Cheaters
	CheatersWritten uint32

	ValidatorStates       []ValidatorBlockState
	NextValidatorProfiles ValidatorProfiles

	DirtyRules *u2u.Rules `rlp:"nil"` // nil means that there's no changes compared to epoch rules

	AdvanceEpochs idx.Epoch

	SfcStateRoot hash.Hash `rlp:"optional"`
}

func (bs BlockState) Copy() BlockState {
	cp := bs
	cp.EpochCheaters = make(types.Cheaters, len(bs.EpochCheaters))
	copy(cp.EpochCheaters, bs.EpochCheaters)
	cp.ValidatorStates = make([]ValidatorBlockState, len(bs.ValidatorStates))
	copy(cp.ValidatorStates, bs.ValidatorStates)
	for i := range cp.ValidatorStates {
		cp.ValidatorStates[i].Originated = new(big.Int).Set(cp.ValidatorStates[i].Originated)
	}
	cp.NextValidatorProfiles = bs.NextValidatorProfiles.Copy()
	if bs.DirtyRules != nil {
		rules := bs.DirtyRules.Copy()
		cp.DirtyRules = &rules
	}
	return cp
}

func (bs *BlockState) GetValidatorState(id idx.ValidatorID, validators *pos.Validators) *ValidatorBlockState {
	validatorIdx := validators.GetIdx(id)
	return &bs.ValidatorStates[validatorIdx]
}

func (bs BlockState) Hash() hash.Hash {
	hasher := sha256.New()
	err := rlp.Encode(hasher, &bs)
	if err != nil {
		panic("can't hash: " + err.Error())
	}
	return hash.BytesToHash(hasher.Sum(nil))
}

type EpochStateV1 struct {
	Epoch          idx.Epoch
	EpochStart     native.Timestamp
	PrevEpochStart native.Timestamp

	EpochStateRoot hash.Hash

	Validators        *pos.Validators
	ValidatorStates   []ValidatorEpochState
	ValidatorProfiles ValidatorProfiles

	Rules u2u.Rules
}

type EpochState EpochStateV1

func (es *EpochState) GetValidatorState(id idx.ValidatorID, validators *pos.Validators) *ValidatorEpochState {
	validatorIdx := validators.GetIdx(id)
	return &es.ValidatorStates[validatorIdx]
}

func (es EpochState) Duration() native.Timestamp {
	return es.EpochStart - es.PrevEpochStart
}

func (es EpochState) Hash() hash.Hash {
	var hashed interface{}
	if es.Rules.Upgrades.London {
		hashed = &es
	} else {
		es0 := EpochStateV0{
			Epoch:             es.Epoch,
			EpochStart:        es.EpochStart,
			PrevEpochStart:    es.PrevEpochStart,
			EpochStateRoot:    es.EpochStateRoot,
			Validators:        es.Validators,
			ValidatorStates:   make([]ValidatorEpochStateV0, len(es.ValidatorStates)),
			ValidatorProfiles: es.ValidatorProfiles,
			Rules:             es.Rules,
		}
		for i, v := range es.ValidatorStates {
			es0.ValidatorStates[i].GasRefund = v.GasRefund
			es0.ValidatorStates[i].PrevEpochEvent = v.PrevEpochEvent.ID
		}
		hashed = &es0
	}
	hasher := sha256.New()
	err := rlp.Encode(hasher, hashed)
	if err != nil {
		panic("can't hash: " + err.Error())
	}
	return hash.BytesToHash(hasher.Sum(nil))
}

func (es EpochState) Copy() EpochState {
	cp := es
	cp.ValidatorStates = make([]ValidatorEpochState, len(es.ValidatorStates))
	copy(cp.ValidatorStates, es.ValidatorStates)
	cp.ValidatorProfiles = es.ValidatorProfiles.Copy()
	if es.Rules != (u2u.Rules{}) {
		cp.Rules = es.Rules.Copy()
	}
	return cp
}
