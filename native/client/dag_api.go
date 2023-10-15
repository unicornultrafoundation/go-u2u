package native

import (
	"context"
	"math/big"

	"github.com/unicornultrafoundation/go-hashgraph/hash"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/common/hexutil"
	ethereum "github.com/unicornultrafoundation/go-u2u/libs"

	"github.com/unicornultrafoundation/go-u2u/native"
)

// GetEvent returns Hashgraph event by hash or short ID.
func (ec *Client) GetEvent(ctx context.Context, h hash.Event) (e native.EventI, err error) {
	var raw map[string]interface{}
	err = ec.c.CallContext(ctx, &raw, "dag_getEvent", h.Hex())
	if err != nil {
		return
	} else if len(raw) == 0 {
		err = ethereum.NotFound
		return
	}

	e = native.RPCUnmarshalEvent(raw)
	return
}

// GetEvent returns Hashgraph event by hash or short ID.
func (ec *Client) GetEventPayload(ctx context.Context, h hash.Event, inclTx bool) (e native.EventI, txs []common.Hash, err error) {
	var raw map[string]interface{}
	err = ec.c.CallContext(ctx, &raw, "dag_getEventPayload", h.Hex(), inclTx)
	if err != nil {
		return
	} else if len(raw) == 0 {
		err = ethereum.NotFound
		return
	}

	e = native.RPCUnmarshalEvent(raw)

	if inclTx {
		vv := raw["transactions"].([]interface{})
		txs = make([]common.Hash, len(vv))
		for i, v := range vv {
			txs[i] = common.HexToHash(v.(string))
		}
	}

	return
}

// GetHeads returns IDs of all the epoch events with no descendants.
// * When epoch is -2 the heads for latest epoch are returned.
// * When epoch is -1 the heads for latest sealed epoch are returned.
func (ec *Client) GetHeads(ctx context.Context, epoch *big.Int) (hash.Events, error) {
	var raw []interface{}
	err := ec.c.CallContext(ctx, &raw, "dag_getHeads", toBlockNumArg(epoch))
	if err != nil {
		return nil, err
	}

	return native.HexToEventIDs(raw), nil
}

// GetEpochStats returns epoch statistics.
// * When epoch is -2 the statistics for latest epoch is returned.
// * When epoch is -1 the statistics for latest sealed epoch is returned.
func (ec *Client) GetEpochStats(ctx context.Context, epoch *big.Int) (map[string]interface{}, error) {
	var raw map[string]interface{}
	err := ec.c.CallContext(ctx, &raw, "dag_getEpochStats", toBlockNumArg(epoch))
	if err != nil {
		return nil, err
	} else if len(raw) == 0 {
		return nil, ethereum.NotFound
	}

	return raw, nil
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	pending := big.NewInt(-1)
	if number.Cmp(pending) == 0 {
		return "pending"
	}
	return hexutil.EncodeBig(number)
}
