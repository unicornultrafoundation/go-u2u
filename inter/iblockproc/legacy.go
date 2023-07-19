package iblockproc

import (
	"github.com/unicornultrafoundation/go-hashgraph/hash"
	"github.com/unicornultrafoundation/go-hashgraph/inter/idx"
	"github.com/unicornultrafoundation/go-hashgraph/inter/pos"

	"github.com/unicornultrafoundation/go-u2u/inter"
	"github.com/unicornultrafoundation/go-u2u/u2u"
)

type ValidatorEpochStateV0 struct {
	GasRefund      uint64
	PrevEpochEvent hash.Event
}

type EpochStateV0 struct {
	Epoch          idx.Epoch
	EpochStart     inter.Timestamp
	PrevEpochStart inter.Timestamp

	EpochStateRoot hash.Hash

	Validators        *pos.Validators
	ValidatorStates   []ValidatorEpochStateV0
	ValidatorProfiles ValidatorProfiles

	Rules u2u.Rules
}
