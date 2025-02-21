package sfcstore

import (
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/driverauth"
)

// GetNdaSfc get the sfc contract address from NodeDriverAuth contract
func (s *Store) GetNdaSfc(state *state.StateDB) common.Hash {
	return state.GetState(driverauth.ContractAddress, NdaSfc)
}
