package ethapi

import (
	"context"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/common/hexutil"
	"github.com/unicornultrafoundation/go-u2u/rpc"
)

var SfcPrecompiles = []common.Address{
	common.HexToAddress("0xFC00FACE00000000000000000000000000000000"),
	common.HexToAddress("0xD100ae0000000000000000000000000000000000"),
	common.HexToAddress("0xd100A01E00000000000000000000000000000000"),
	common.HexToAddress("0x6CA548f6DF5B540E72262E935b6Fe3e72cDd68C9"),
	common.HexToAddress("0xFC01fACE00000000000000000000000000000000"), // SFCLib
}

// PublicSfcAPI provides an API to access SFC related information.
type PublicSfcAPI struct {
	b Backend
}

func NewPublicSfcAPI(b Backend) *PublicSfcAPI {
	return &PublicSfcAPI{b}
}

// GetStorageAt returns the storage from the SFC state at the given address, key and
// block number.
// The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta-block numbers are also allowed.
func (s *PublicSfcAPI) GetStorageAt(ctx context.Context, address common.Address, key string,
	blockNr rpc.BlockNumberOrHash) (hexutil.Bytes, error) {
	state, _, err := s.b.SfcStateAndHeaderByNumberOrHash(ctx, blockNr)
	if state == nil || err != nil {
		return nil, err
	}
	res := state.GetState(address, common.HexToHash(key))
	return res[:], state.Error()
}

func (s *PublicSfcAPI) CheckIntegrity(ctx context.Context, blockNr rpc.BlockNumberOrHash) (bool, error) {
	state, _, err := s.b.StateAndHeaderByNumberOrHash(ctx, blockNr)
	if state == nil || err != nil {
		return false, err
	}
	sfcState, _, err := s.b.SfcStateAndHeaderByNumberOrHash(ctx, blockNr)
	if sfcState == nil || err != nil {
		return false, err
	}
	for _, addr := range SfcPrecompiles {
		original := state.GetStorageRoot(addr)
		sfc := sfcState.GetStorageRoot(addr)
		if original.Cmp(sfc) != 0 {
			return false, nil
		}
		originalBalance := state.GetBalance(addr)
		sfcBalance := sfcState.GetBalance(addr)
		if originalBalance.Cmp(sfcBalance) != 0 {
			return false, nil
		}
		originalNonce := state.GetNonce(addr)
		sfcNonce := sfcState.GetNonce(addr)
		if originalNonce != sfcNonce {
			return false, nil
		}
	}
	return true, nil
}
