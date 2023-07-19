package emitter

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/go-hashgraph/hash"
	"github.com/unicornultrafoundation/go-hashgraph/inter/idx"
	"github.com/unicornultrafoundation/go-hashgraph/inter/pos"

	"github.com/unicornultrafoundation/go-u2u/gossip/emitter/mock"
	"github.com/unicornultrafoundation/go-u2u/integration/makefakegenesis"
	"github.com/unicornultrafoundation/go-u2u/inter"
	"github.com/unicornultrafoundation/go-u2u/u2u"
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
			Return(inter.Timestamp(uint64(time.Now().UnixNano()))).
			AnyTimes()

		em.init()
	})

	t.Run("memorizeTxTimes", func(t *testing.T) {
		require := require.New(t)
		tx := types.NewTransaction(1, common.Address{}, big.NewInt(1), 1, big.NewInt(1), nil)

		external.EXPECT().IsBusy().
			Return(true).
			AnyTimes()

		_, ok := em.txTime.Get(tx.Hash())
		require.False(ok)

		before := time.Now()
		em.memorizeTxTimes(types.Transactions{tx})
		after := time.Now()

		cached, ok := em.txTime.Get(tx.Hash())
		got := cached.(time.Time)
		require.True(ok)
		require.True(got.After(before))
		require.True(got.Before(after))
	})

	t.Run("tick", func(t *testing.T) {
		em.tick()
	})
}
