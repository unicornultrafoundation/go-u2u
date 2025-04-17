package flushable

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/go-u2u/consensus/common/bigendian"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb/leveldb"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb/table"
)

func TestFlushableParallel(t *testing.T) {
	testDuration := 2 * time.Second
	testPairsNum := uint64(1000)

	t.Run("Exactly flush", func(t *testing.T) {
		disk := tmpDir()
		// open raw databases
		ldb, _ := disk.OpenDB("1")
		defer ldb.Drop()
		defer ldb.Close()

		flushableDb := Wrap(ldb)

		tableImmutable := table.New(flushableDb, []byte("2"))

		// fill data
		for i := uint64(0); i < testPairsNum; i++ {
			_ = tableImmutable.Put(bigendian.Uint64ToBytes(i), bigendian.Uint64ToBytes(i))
			if i == testPairsNum/2 { // a half of data is flushed, other half isn't
				_ = flushableDb.Flush()
			}
		}

		require := require.New(t)

		// iterate over tableImmutable and check its content
		it := tableImmutable.NewIterator(nil, nil)
		defer it.Release()

		_ = flushableDb.Flush() // !breaking flush

		i := uint64(0)
		for ; it.Next(); i++ {
			require.NoError(it.Error(), i)
			require.Equal(bigendian.Uint64ToBytes(i), it.Key(), i)
			require.Equal(bigendian.Uint64ToBytes(i), it.Value(), i)
		}
		require.Equal(testPairsNum, i) // !here

	})

	t.Run("Random flush", func(t *testing.T) {
		disk := tmpDir()
		// open raw databases
		ldb, _ := disk.OpenDB("1")
		defer ldb.Drop()
		defer ldb.Close()

		flushableDb := Wrap(ldb)

		tableMutable1 := table.New(flushableDb, []byte("1"))
		tableImmutable := table.New(flushableDb, []byte("2"))
		tableMutable2 := table.New(flushableDb, []byte("3"))

		// fill data
		for i := uint64(0); i < testPairsNum; i++ {
			_ = tableImmutable.Put(bigendian.Uint64ToBytes(i), bigendian.Uint64ToBytes(i))
			if i == testPairsNum/2 { // a half of data is flushed, other half isn't
				_ = flushableDb.Flush()
			}
		}

		stop := make(chan struct{})
		stopped := func() bool {
			select {
			case <-stop:
				return true
			default:
				return false
			}
		}

		work := sync.WaitGroup{}
		work.Add(4)
		for g := 0; g < 2; g++ {
			go func() {
				defer work.Done()
				require := require.New(t)
				for !stopped() {
					// iterate over tableImmutable and check its content
					it := tableImmutable.NewIterator(nil, nil)
					defer it.Release()

					i := uint64(0)
					for ; it.Next(); i++ {
						require.NoError(it.Error(), i)
						require.Equal(bigendian.Uint64ToBytes(i), it.Key(), i)
						require.Equal(bigendian.Uint64ToBytes(i), it.Value(), i)
					}
					require.Equal(testPairsNum, i)
				}
			}()
		}

		for g := 0; g < 2; g++ {
			go func() {
				defer work.Done()
				r := rand.New(rand.NewSource(0)) // nolint:gosec
				for !stopped() {
					// try to spoil data in tableImmutable by updating other tables
					_ = tableMutable1.Put(bigendian.Uint64ToBytes(r.Uint64()%testPairsNum), bigendian.Uint64ToBytes(r.Uint64()))
					_ = tableMutable2.Put(bigendian.Uint64ToBytes(r.Uint64() % testPairsNum)[:7], bigendian.Uint64ToBytes(r.Uint64()))
					if r.Int63n(100) == 0 {
						_ = flushableDb.Flush() // flush with 1% chance
					}
				}
			}()
		}

		time.Sleep(testDuration)
		close(stop)
		work.Wait()
	})
}

func tmpDir() u2udb.DBProducer {
	dir, err := ioutil.TempDir("", "test-flushable")
	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory %s: %v", dir, err))
	}
	disk := leveldb.NewProducer(dir, cache16mb)
	return disk
}
