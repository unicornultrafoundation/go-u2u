package ethclient

import (
	"context"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/common/hexutil"
)

// SfcStorageAt returns the value of key in the SFC contract storage of the given account.
// The block number can be nil, in which case the value is taken from the latest known block.
func (ec *Client) SfcStorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	var result hexutil.Bytes
	err := ec.c.CallContext(ctx, &result, "sfc_getStorageAt", account, key, toBlockNumArg(blockNumber))
	return result, err
}
