package flushable

import (
	"bytes"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/unicornultrafoundation/go-u2u/common"

	"github.com/unicornultrafoundation/go-u2u/common/bigendian"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb/leveldb"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb/table"
)

func TestFlushable(t *testing.T) {
	assertar := assert.New(t)

	tries := 60            // number of test iterations
	opsPerIter := 0x140    // max number of put/delete ops per iteration
	dictSize := opsPerIter // number of different words

	disk := dbProducer("TestFlushable")

	// open raw databases
	leveldb1, _ := disk.OpenDB("1")
	defer leveldb1.Drop()
	defer leveldb1.Close()

	leveldb2, _ := disk.OpenDB("2")
	defer leveldb2.Drop()
	defer leveldb2.Close()

	// create wrappers
	dbs := map[string]u2udb.Store{
		"leveldb": leveldb1,
		"memory":  Wrap(devnull),
	}

	flushableDbs := map[string]*Flushable{
		"cache-over-leveldb": Wrap(leveldb2),
		"cache-over-memory":  Wrap(Wrap(devnull)),
	}

	baseLdb := table.New(dbs["leveldb"], []byte{})
	baseMem := table.New(dbs["memory"], []byte{})

	dbsTables := [][]u2udb.Store{
		{
			dbs["leveldb"],
			baseLdb.NewTable([]byte{0, 1}),
			baseLdb.NewTable([]byte{0}).NewTable(common.Hex2Bytes("ffffffffffffffffffffffffffffffffffff")),
		},
		{
			dbs["memory"],
			baseMem.NewTable([]byte{0, 1}),
			baseMem.NewTable([]byte{0}).NewTable(common.Hex2Bytes("ffffffffffffffffffffffffffffffffffff")),
		},
	}

	baseLdb = table.New(flushableDbs["cache-over-leveldb"], []byte{})
	baseMem = table.New(flushableDbs["cache-over-memory"], []byte{})
	flushableDbsTables := [][]u2udb.Store{
		{
			flushableDbs["cache-over-leveldb"],
			baseLdb.NewTable([]byte{0, 1}),
			baseLdb.NewTable([]byte{0}).NewTable(common.Hex2Bytes("ffffffffffffffffffffffffffffffffffff")),
		},
		{
			flushableDbs["cache-over-memory"],
			baseMem.NewTable([]byte{0, 1}),
			baseMem.NewTable([]byte{0}).NewTable(common.Hex2Bytes("ffffffffffffffffffffffffffffffffffff")),
		},
	}

	assertar.Equal(len(dbsTables), len(flushableDbsTables))
	assertar.Equal(len(dbsTables[0]), len(flushableDbsTables[0]))

	groupsNum := len(dbsTables)
	tablesNum := len(dbsTables[0])

	// use the same seed for determinism
	rand := rand.New(rand.NewSource(0)) // nolint:gosec

	// words dictionary
	prefixes := [][]byte{
		{},
		{0},
		{0x1},
		{0x22},
		{0x33},
		{0x11},
		{0x11, 0x22},
		{0x11, 0x23},
		{0x11, 0x22, 0x33},
		{0x11, 0x22, 0x34},
	}
	dict := [][]byte{}
	for i := 0; i < dictSize; i++ {
		b := append(prefixes[i%len(prefixes)], big.NewInt(rand.Int63()).Bytes()...)
		dict = append(dict, b)
	}

	for try := 0; try < tries; try++ {

		// random put/delete operations
		putDeleteRandom := func() {
			for j := 0; j < tablesNum; j++ {
				var batches []u2udb.Batch
				for i := 0; i < groupsNum; i++ {
					batches = append(batches, dbsTables[i][j].NewBatch())
					batches = append(batches, flushableDbsTables[i][j].NewBatch())
				}

				ops := 1 + rand.Intn(opsPerIter)
				for p := 0; p < ops; p++ {
					var pair kv
					if rand.Intn(2) == 0 { // put
						pair = kv{
							k: dict[rand.Intn(len(dict))],
							v: dict[rand.Intn(len(dict))],
						}
					} else { // delete
						pair = kv{
							k: dict[rand.Intn(len(dict))],
							v: nil,
						}
					}

					for _, batch := range batches {
						if pair.v != nil {
							assertar.NoError(batch.Put(pair.k, pair.v))
						} else {
							assertar.NoError(batch.Delete(pair.k))
						}
					}
				}

				for _, batch := range batches {
					size := batch.ValueSize()
					assertar.NotEqual(0, size)
					assertar.NoError(batch.Write())
					assertar.Equal(size, batch.ValueSize())
					batch.Reset()
					assertar.Equal(0, batch.ValueSize())
				}
			}
		}
		// put/delete values
		putDeleteRandom()

		// flush
		for _, db := range flushableDbs {
			if try == 0 && !assertar.NotEqual(0, db.NotFlushedPairs()) {
				return
			}
			assertar.NoError(db.Flush())
			assertar.Equal(0, db.NotFlushedPairs())
		}

		// put/delete values (not flushed)
		putDeleteRandom()

		// try to ForEach random prefix
		prefix := prefixes[try%len(prefixes)]
		if try == 1 {
			prefix = []byte{0, 0, 0, 0, 0, 0} // not existing prefix
		}

		for j := 0; j < tablesNum; j++ {
			expectPairs := []kv{}

			testForEach := func(db u2udb.Store, first bool) {

				var it u2udb.Iterator
				if try%4 == 0 {
					it = db.NewIterator(nil, nil)
				} else if try%4 == 1 {
					it = db.NewIterator(prefix, nil)
				} else if try%4 == 2 {
					it = db.NewIterator(nil, prefix)
				} else {
					it = db.NewIterator(prefix[:len(prefix)/2], prefix[len(prefix)/2:])
				}
				defer it.Release()

				var got int

				for got = 0; it.Next(); got++ {
					if first {
						expectPairs = append(expectPairs, kv{
							k: common.CopyBytes(it.Key()),
							v: common.CopyBytes(it.Value()),
						})
					} else {
						assertar.NotEqual(len(expectPairs), got, try) // check that we've for the same num of values
						if t.Failed() {
							return
						}
						assertar.Equal(expectPairs[got].k, it.Key(), try)
						assertar.Equal(expectPairs[got].v, it.Value(), try)
					}
				}

				if !assertar.NoError(it.Error()) {
					return
				}

				assertar.Equal(len(expectPairs), got) // check that we've got the same num of pairs
			}

			// check that all groups return the same result
			for i := 0; i < groupsNum; i++ {
				testForEach(dbsTables[i][j], i == 0)
				if t.Failed() {
					return
				}
				testForEach(flushableDbsTables[i][j], false)
				if t.Failed() {
					return
				}
			}
		}

		// try to get random values
		ops := rand.Intn(opsPerIter)
		for p := 0; p < ops; p++ {
			key := dict[rand.Intn(len(dict))]

			for j := 0; j < tablesNum; j++ {
				// get values for first group, so we could check that all groups return the same result
				ok, _ := dbsTables[0][j].Has(key)
				vl, _ := dbsTables[0][j].Get(key)

				// check that all groups return the same result
				for i := 0; i < groupsNum; i++ {
					ok1, err := dbsTables[i][j].Has(key)
					assertar.NoError(err)
					vl1, err := dbsTables[i][j].Get(key)
					assertar.NoError(err)

					ok2, err := flushableDbsTables[i][j].Has(key)
					assertar.NoError(err)
					vl2, err := flushableDbsTables[i][j].Get(key)
					assertar.NoError(err)

					assertar.Equal(ok1, ok2)
					assertar.Equal(vl1, vl2)
					assertar.Equal(ok1, ok)
					assertar.Equal(vl1, vl)
				}
			}
		}

		if t.Failed() {
			return
		}
	}
}

func TestFlushableIterator(t *testing.T) {
	require := require.New(t)

	disk := dbProducer("TestFlushableIterator")

	leveldb, err := disk.OpenDB("1")
	require.NoError(err)
	defer leveldb.Drop()
	defer leveldb.Close()

	flushable1 := Wrap(leveldb)
	flushable2 := Wrap(leveldb)

	allkeys := [][]byte{
		{0x11, 0x00},
		{0x12, 0x00},
		{0x13, 0x00},
		{0x14, 0x00},
		{0x15, 0x00},
		{0x16, 0x00},
		{0x17, 0x00},
		{0x18, 0x00},
		{0x19, 0x00},
		{0x1a, 0x00},
		{0x1b, 0x00},
		{0x1c, 0x00},
		{0x1d, 0x00},
		{0x1e, 0x00},
		{0x1f, 0x00},
	}

	veryFirstKey := allkeys[0]
	veryLastKey := allkeys[len(allkeys)-1]
	expected := allkeys[1 : len(allkeys)-1]

	for _, key := range expected {
		require.NoError(leveldb.Put(key, []byte("in-order")))
	}

	require.NoError(flushable2.Put(veryFirstKey, []byte("first")))
	require.NoError(flushable2.Put(veryLastKey, []byte("last")))

	it := flushable1.NewIterator(nil, nil)
	defer it.Release()

	require.NoError(flushable2.Flush())

	for i := 0; it.Next(); i++ {
		require.Equal(expected[i], it.Key())
		require.Equal([]byte("in-order"), it.Value())
	}
}

func BenchmarkFlushable(b *testing.B) {
	disk := dbProducer("BenchmarkFlushable")

	leveldb, _ := disk.OpenDB("1")
	defer leveldb.Drop()
	defer leveldb.Close()

	flushable := Wrap(leveldb)

	const recs = 64

	for _, flushPeriod := range []int{0, 1, 10, 100, 1000} {
		for goroutines := 1; goroutines <= recs; goroutines *= 2 {
			for reading := 0; reading <= 1; reading++ {
				name := fmt.Sprintf(
					"%d goroutines with flush every %d ops, readingExtensive=%d",
					goroutines, flushPeriod, reading)
				b.Run(name, func(b *testing.B) {
					benchmarkFlushable(flushable, goroutines, b.N, flushPeriod, reading != 0)
				})
			}
		}
	}
}

func benchmarkFlushable(db *Flushable, goroutines, recs, flushPeriod int, readingExtensive bool) {
	leftRecs := recs
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		var ops = recs/goroutines + 1
		if ops > leftRecs {
			ops = leftRecs
		}
		leftRecs -= ops
		go func(i int) {
			defer wg.Done()

			flushOffset := flushPeriod * i / goroutines

			for op := 0; op < ops; op++ {
				step := op & 0xff
				key := bigendian.Uint64ToBytes(uint64(step << 48))
				val := bigendian.Uint64ToBytes(uint64(step))

				rare := time.Now().Unix()%100 == 0
				if readingExtensive == rare {
					err := db.Put(key, val)
					if err != nil {
						panic(err)
					}
				} else {
					got, err := db.Get(key)
					if err != nil {
						panic(err)
					}
					if got != nil && !bytes.Equal(val, got) {
						panic("invalid value")
					}
				}

				if flushPeriod != 0 && (op+flushOffset)%flushPeriod == 0 {
					err := db.Flush()
					if err != nil {
						panic(err)
					}
				}
			}
		}(i)
	}
	wg.Wait()
}

func cache16mb(string) (int, int) {
	return 16 * opt.MiB, 64
}

func dbProducer(name string) u2udb.DBProducer {
	dir, err := os.MkdirTemp("", name)
	if err != nil {
		panic(err)
	}
	return leveldb.NewProducer(dir, cache16mb)
}
