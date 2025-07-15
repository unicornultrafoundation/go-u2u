package integration

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/unicornultrafoundation/go-helios/consensus"
	"github.com/unicornultrafoundation/go-helios/hash"
	"github.com/unicornultrafoundation/go-helios/native/idx"
	"github.com/unicornultrafoundation/go-helios/utils/cachescale"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/gossip"
	"github.com/unicornultrafoundation/go-u2u/integration/makefakegenesis"
	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/u2u"
	"github.com/unicornultrafoundation/go-u2u/utils"
	"github.com/unicornultrafoundation/go-u2u/vecmt"
)

func BenchmarkFlushDBs(b *testing.B) {
	dir := tmpDir("flush_bench")
	defer os.RemoveAll(dir)
	genStore := makefakegenesis.FakeGenesisStore(1, utils.ToU2U(1), utils.ToU2U(1), u2u.GetClymeneUpgrades())
	g := genStore.Genesis()
	_, _, store, s2, _, closeDBs := MakeEngine(dir, &g, Configs{
		U2U:         gossip.DefaultConfig(cachescale.Identity),
		U2UStore:    gossip.DefaultStoreConfig(cachescale.Identity),
		Helios:      consensus.DefaultConfig(),
		HeliosStore: consensus.DefaultStoreConfig(cachescale.Identity),
		VectorClock: vecmt.DefaultConfig(cachescale.Identity),
		DBs:         DefaultDBsConfig(cachescale.Identity.U64, 512),
	})
	defer closeDBs()
	defer store.Close()
	defer s2.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		n := idx.Block(0)
		randUint32s := func() []uint32 {
			arr := make([]uint32, 128)
			for i := 0; i < len(arr); i++ {
				arr[i] = uint32(i) ^ (uint32(n) << 16) ^ 0xd0ad884e
			}
			return []uint32{uint32(n), uint32(n) + 1, uint32(n) + 2}
		}
		for !store.IsCommitNeeded() {
			store.SetBlock(n, &native.Block{
				Time:        native.Timestamp(n << 32),
				Atropos:     hash.Event{},
				Events:      hash.Events{},
				Txs:         []common.Hash{},
				InternalTxs: []common.Hash{},
				SkippedTxs:  randUint32s(),
				GasUsed:     uint64(n) << 24,
				Root:        hash.Hash{},
			})
			n++
		}
		b.StartTimer()
		err := store.Commit()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func tmpDir(name string) string {
	dir, err := ioutil.TempDir("", name)
	if err != nil {
		panic(err)
	}
	return dir
}
