package emitter

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/unicornultrafoundation/go-helios/emitter/ancestor"
	"github.com/unicornultrafoundation/go-helios/hash"
	"github.com/unicornultrafoundation/go-helios/native/idx"
	"github.com/unicornultrafoundation/go-helios/native/pos"
	"github.com/unicornultrafoundation/go-helios/u2udb/memorydb"

	"github.com/unicornultrafoundation/go-u2u/gossip/emitter/mock"
	"github.com/unicornultrafoundation/go-u2u/integration/makefakegenesis"
	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/u2u"
	"github.com/unicornultrafoundation/go-u2u/vecmt"
)

func TestParents(t *testing.T) {
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
		Return(func() *vecmt.Index {
			vi := vecmt.NewIndex(func(err error) { panic(err) }, vecmt.LiteConfig())
			vi.Reset(validators, memorydb.New(), nil)
			return vi
		}()).
		AnyTimes()
	external.EXPECT().IsSynced().
		Return(true).
		AnyTimes()
	external.EXPECT().PeersNum().
		Return(int(3)).
		AnyTimes()
	external.EXPECT().StateDB().
		Return(nil).
		AnyTimes()

	em := NewEmitter(cfg, World{
		External: external,
		TxPool:   txPool,
		Signer:   signer,
		TxSigner: txSigner,
	})

	t.Run("init", func(t *testing.T) {
		external.EXPECT().GetRules().
			Return(u2u.FakeNetRules(u2u.GetVitriolUpgrades())).
			AnyTimes()

		external.EXPECT().GetEpochValidators().
			Return(validators, idx.Epoch(1)).
			AnyTimes()

		external.EXPECT().GetLastEvent(idx.Epoch(1), cfg.Validator.ID).
			Return((*hash.Event)(nil)).
			AnyTimes()

		external.EXPECT().GetLastEvent(idx.Epoch(2), cfg.Validator.ID).
			Return(new(hash.Event)).
			AnyTimes()

		external.EXPECT().GetLastEvent(idx.Epoch(2), gValidators[1].ID).
			Return((*hash.Event)(nil)).
			AnyTimes()

		external.EXPECT().GetHeads(idx.Epoch(1)).
			Return(hash.Events{}).
			AnyTimes()

		external.EXPECT().GetHeads(idx.Epoch(2)).
			Return(hash.Events{}).
			AnyTimes()

		external.EXPECT().GetGenesisTime().
			Return(native.Timestamp(uint64(time.Now().UnixNano()))).
			AnyTimes()

		em.init()
	})

	t.Run("build strategies 0 events", func(t *testing.T) {
		strategies := em.buildSearchStrategies(idx.Event(0))
		require.Equal(t, 0, len(strategies))
	})

	t.Run("build strategies 1 event", func(t *testing.T) {
		strategies := em.buildSearchStrategies(idx.Event(1))
		require.Equal(t, 1, len(strategies))
	})

	t.Run("build strategies 4 event", func(t *testing.T) {
		strategies := em.buildSearchStrategies(idx.Event(4))
		require.Equal(t, 4, len(strategies))
	})

	t.Run("build strategies with fcIndexer", func(t *testing.T) {
		gValidator := makefakegenesis.GetFakeValidators(1)
		vvNew := pos.NewBuilder()
		vvNew.Set(gValidator[0].ID, pos.Weight(1))
		newValidators := vvNew.Build()

		em.quorumIndexer = nil
		em.fcIndexer = ancestor.NewFCIndexer(newValidators, em.world.DagIndex(), em.config.Validator.ID)

		strategies := em.buildSearchStrategies(idx.Event(4))
		require.Equal(t, 4, len(strategies))
	})

	t.Run("choose parent not selfParent", func(t *testing.T) {
		event, events, ok := em.chooseParents(idx.Epoch(1), em.config.Validator.ID)
		require.Equal(t, true, ok)
		var eventExp *hash.Event
		require.Equal(t, eventExp, event)
		require.Equal(t, hash.Events{}, events)
	})

	t.Run("choose parent selfParent", func(t *testing.T) {
		event, events, ok := em.chooseParents(idx.Epoch(2), em.config.Validator.ID)
		require.Equal(t, true, ok)
		eventExp := new(hash.Event)
		require.Equal(t, eventExp, event)
		require.Equal(t, hash.Events{hash.Event{}}, events)
	})
}
