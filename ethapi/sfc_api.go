package ethapi

import (
	"context"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/common/hexutil"
	"github.com/unicornultrafoundation/go-u2u/rpc"
)

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
