package iblockproc

import (
	"github.com/unicornultrafoundation/go-u2u/helios/hash"
	"github.com/unicornultrafoundation/go-u2u/helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/helios/native/pos"

	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/u2u"
)

type ValidatorEpochStateV0 struct {
	GasRefund      uint64
	PrevEpochEvent hash.Event
}

type EpochStateV0 struct {
	Epoch          idx.Epoch
	EpochStart     native.Timestamp
	PrevEpochStart native.Timestamp

	EpochStateRoot hash.Hash

	Validators        *pos.Validators
	ValidatorStates   []ValidatorEpochStateV0
	ValidatorProfiles ValidatorProfiles

	Rules u2u.Rules
}
