package emitter

import (
	"math/big"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/go-hashgraph/hash"
	"github.com/unicornultrafoundation/go-hashgraph/native/idx"
	"github.com/unicornultrafoundation/go-hashgraph/native/pos"
	"github.com/unicornultrafoundation/go-u2u/libs/common"
	"github.com/unicornultrafoundation/go-u2u/libs/core/types"

	"github.com/unicornultrafoundation/go-u2u/gossip/emitter/mock"
	"github.com/unicornultrafoundation/go-u2u/integration/makefakegenesis"
	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/u2u"
	"github.com/unicornultrafoundation/go-u2u/utils/txtime"
	"github.com/unicornultrafoundation/go-u2u/vecmt"
)

//go:generate go run github.com/golang/mock/mockgen -package=mock -destination=mock/world.go github.com/unicornultrafoundation/go-u2u/gossip/emitter External,TxPool,TxSigner,Signer

func TestEmitter(t *testing.T) {
	cfg := DefaultConfig()
	gValidators := makefakegenesis.GetFakeValidators(3)
	vv := pos.NewBuilder()
	for _, v := range gValidators {
		vv.Set(v.ID, pos.Weight(1))
	}
	validators := vv.Build()
	cfg.Validator.ID = gValidators[0].ID

	ctrl := gomock.NewController(t)
	external := mock.NewMockExternal(ctrl)
	txPool := mock.NewMockTxPool(ctrl)
	signer := mock.NewMockSigner(ctrl)
	txSigner := mock.NewMockTxSigner(ctrl)

	external.EXPECT().Lock().
		AnyTimes()
	external.EXPECT().Unlock().
		AnyTimes()
	external.EXPECT().DagIndex().
		Return((*vecmt.Index)(nil)).
		AnyTimes()
	external.EXPECT().IsSynced().
		Return(true).
		AnyTimes()
	external.EXPECT().PeersNum().
		Return(int(3)).
		AnyTimes()

	em := NewEmitter(cfg, World{
		External: external,
		TxPool:   txPool,
		Signer:   signer,
		TxSigner: txSigner,
	})

	t.Run("init", func(t *testing.T) {
		external.EXPECT().GetRules().
			Return(u2u.FakeNetRules()).
			AnyTimes()

		external.EXPECT().GetEpochValidators().
			Return(validators, idx.Epoch(1)).
			AnyTimes()

		external.EXPECT().GetLastEvent(idx.Epoch(1), cfg.Validator.ID).
			Return((*hash.Event)(nil)).
			AnyTimes()

		external.EXPECT().GetGenesisTime().
			Return(native.Timestamp(uint64(time.Now().UnixNano()))).
			AnyTimes()

		em.init()
	})

	t.Run("memorizeTxTimes", func(t *testing.T) {
		txtime.Enabled = true
		require := require.New(t)
		tx1 := types.NewTransaction(1, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)
		tx2 := types.NewTransaction(2, common.Address{}, big.NewInt(2), 2, big.NewInt(2), nil)

		external.EXPECT().IsBusy().
			Return(true).
			AnyTimes()

		txtime.Saw(tx1.Hash(), time.Unix(1, 0))

		require.Equal(time.Unix(1, 0), txtime.Of(tx1.Hash()))
		txtime.Saw(tx1.Hash(), time.Unix(2, 0))
		require.Equal(time.Unix(1, 0), txtime.Of(tx1.Hash()))
		txtime.Validated(tx1.Hash(), time.Unix(2, 0))
		require.Equal(time.Unix(1, 0), txtime.Of(tx1.Hash()))

		// reversed order
		txtime.Validated(tx2.Hash(), time.Unix(3, 0))
		txtime.Saw(tx2.Hash(), time.Unix(2, 0))

		require.Equal(time.Unix(3, 0), txtime.Of(tx2.Hash()))
		txtime.Saw(tx2.Hash(), time.Unix(3, 0))
		require.Equal(time.Unix(3, 0), txtime.Of(tx2.Hash()))
		txtime.Validated(tx2.Hash(), time.Unix(3, 0))
		require.Equal(time.Unix(3, 0), txtime.Of(tx2.Hash()))
	})

	t.Run("tick", func(t *testing.T) {
		em.tick()
	})
}
