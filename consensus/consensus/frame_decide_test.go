package consensus

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/dag"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/dag/tdag"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/idx"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/pos"
	"github.com/unicornultrafoundation/go-u2u/consensus/types"
)

func TestConfirmBlocks_1(t *testing.T) {
	testConfirmBlocks(t, []pos.Weight{1}, 0)
}

func TestConfirmBlocks_big1(t *testing.T) {
	testConfirmBlocks(t, []pos.Weight{math.MaxUint32 / 2}, 0)
}

func TestConfirmBlocks_big2(t *testing.T) {
	testConfirmBlocks(t, []pos.Weight{math.MaxUint32 / 4, math.MaxUint32 / 4}, 0)
}

func TestConfirmBlocks_big3(t *testing.T) {
	testConfirmBlocks(t, []pos.Weight{math.MaxUint32 / 8, math.MaxUint32 / 8, math.MaxUint32 / 4}, 0)
}

func TestConfirmBlocks_4(t *testing.T) {
	testConfirmBlocks(t, []pos.Weight{1, 2, 3, 4}, 0)
}

func TestConfirmBlocks_3_1(t *testing.T) {
	testConfirmBlocks(t, []pos.Weight{1, 1, 1, 1}, 1)
}

func TestConfirmBlocks_67_33(t *testing.T) {
	testConfirmBlocks(t, []pos.Weight{33, 67}, 1)
}

func TestConfirmBlocks_67_33_4(t *testing.T) {
	testConfirmBlocks(t, []pos.Weight{11, 11, 11, 67}, 3)
}

func TestConfirmBlocks_67_33_5(t *testing.T) {
	testConfirmBlocks(t, []pos.Weight{11, 11, 11, 33, 34}, 3)
}

func TestConfirmBlocks_2_8_10(t *testing.T) {
	testConfirmBlocks(t, []pos.Weight{1, 2, 1, 2, 1, 2, 1, 2, 1, 2}, 3)
}

func testConfirmBlocks(t *testing.T, weights []pos.Weight, cheatersCount int) {
	t.Helper()
	assertar := assert.New(t)

	nodes := tdag.GenNodes(len(weights))
	lch, _, input, _ := FakeConsensus(nodes, weights)

	var (
		frames []idx.Frame
		blocks []*types.Block
	)
	lch.applyBlock = func(block *types.Block) *pos.Validators {
		frames = append(frames, lch.store.GetLastDecidedFrame()+1)
		blocks = append(blocks, block)

		return nil
	}

	eventCount := int(TestMaxEpochEvents)
	parentCount := 5
	if parentCount > len(nodes) {
		parentCount = len(nodes)
	}
	r := rand.New(rand.NewSource(int64(len(nodes) + cheatersCount))) // nolint:gosec
	tdag.ForEachRandFork(nodes, nodes[:cheatersCount], eventCount, parentCount, 10, r, tdag.ForEachEvent{
		Process: func(e dag.Event, name string) {
			input.SetEvent(e)
			assertar.NoError(
				lch.Process(e))

		},
		Build: func(e dag.MutableEvent, name string) error {
			e.SetEpoch(FirstEpoch)
			return lch.Build(e)
		},
	})

	// unconfirm all events
	it := lch.store.epochTable.ConfirmedEvent.NewIterator(nil, nil)
	batch := lch.store.epochTable.ConfirmedEvent.NewBatch()
	for it.Next() {
		assertar.NoError(batch.Delete(it.Key()))
	}
	assertar.NoError(batch.Write())
	it.Release()

	for i, block := range blocks {
		frame := frames[i]
		event := blocks[i].Event

		// call confirmBlock again
		_, err := lch.onFrameDecided(frame, event)
		gotBlock := lch.blocks[lch.lastBlock]

		if !assertar.NoError(err) {
			break
		}
		if !assertar.LessOrEqual(gotBlock.Cheaters.Len(), cheatersCount) {
			break
		}
		if !assertar.Equal(block.Cheaters, gotBlock.Cheaters) {
			break
		}
		if !assertar.Equal(block.Event, gotBlock.Event) {
			break
		}
	}
	assertar.GreaterOrEqual(len(blocks), TestMaxEpochEvents/5)
}
