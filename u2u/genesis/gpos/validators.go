package gpos

import (
	"github.com/unicornultrafoundation/go-u2u/libs/common"
	"github.com/unicornultrafoundation/go-u2u/native"

	"github.com/unicornultrafoundation/go-hashgraph/native/idx"
	"github.com/unicornultrafoundation/go-u2u/native/validatorpk"
)

type (
	// Validator is a helper structure to define genesis validators
	Validator struct {
		ID               idx.ValidatorID
		Address          common.Address
		PubKey           validatorpk.PubKey
		CreationTime     native.Timestamp
		CreationEpoch    idx.Epoch
		DeactivatedTime  native.Timestamp
		DeactivatedEpoch idx.Epoch
		Status           uint64
	}

	Validators []Validator
)

// Map converts Validators to map
func (gv Validators) Map() map[idx.ValidatorID]Validator {
	validators := map[idx.ValidatorID]Validator{}
	for _, val := range gv {
		validators[val.ID] = val
	}
	return validators
}

// PubKeys returns not sorted genesis pub keys
func (gv Validators) PubKeys() []validatorpk.PubKey {
	res := make([]validatorpk.PubKey, 0, len(gv))
	for _, v := range gv {
		res = append(res, v.PubKey)
	}
	return res
}

// Addresses returns not sorted genesis addresses
func (gv Validators) Addresses() []common.Address {
	res := make([]common.Address, 0, len(gv))
	for _, v := range gv {
		res = append(res, v.Address)
	}
	return res
}
