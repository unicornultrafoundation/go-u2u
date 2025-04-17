package switchable

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/unicornultrafoundation/go-u2u/helios/common/bigendian"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb/memorydb"

	"github.com/unicornultrafoundation/go-u2u/utils/dbutil/dbcounter"
)

func decodePair(b []byte) (uint32, uint32) {
	v1 := bigendian.BytesToUint32(b[:4])
	v2 := bigendian.BytesToUint32(b[4:])
	return v1, v2
}

type UncallableAfterRelease struct {
	u2udb.Snapshot
	iterators []*uncallableAfterReleaseIterator
	mu        sync.Mutex
}

type uncallableAfterReleaseIterator struct {
	u2udb.Iterator
}

func (db *UncallableAfterRelease) NewIterator(prefix []byte, start []byte) u2udb.Iterator {
	it := db.Snapshot.NewIterator(prefix, start)
	wrapped := &uncallableAfterReleaseIterator{it}

	db.mu.Lock()
	defer db.mu.Unlock()
	db.iterators = append(db.iterators, wrapped)

	return wrapped
}

func (db *UncallableAfterRelease) Release() {
	db.Snapshot.Release()
	// ensure nil pointer exception on any next call
	db.Snapshot = nil

	db.mu.Lock()
	defer db.mu.Unlock()
	for _, it := range db.iterators {
		it.Iterator = nil
	}
}

func TestSnapshot_SwitchTo(t *testing.T) {
	require := require.New(t)

	const prefixes = 100
	const keys = 100
	const checkers = 5
	const switchers = 5
	const duration = time.Millisecond * 400

	// fill DB with data
	memdb := dbcounter.WrapStore(memorydb.New(), "", false)
	for i := uint32(0); i < prefixes; i++ {
		for j := uint32(0); j < keys; j++ {
			key := append(bigendian.Uint32ToBytes(i), bigendian.Uint32ToBytes(j)...)
			val := append(bigendian.Uint32ToBytes(i*i), bigendian.Uint32ToBytes(j*j)...)
			require.NoError(memdb.Put(key, val))
		}
	}

	// 4 readers, one interrupter
	snap, err := memdb.GetSnapshot()
	require.NoError(err)
	switchable := Wrap(&UncallableAfterRelease{
		Snapshot: snap,
	})

	stop := uint32(0)
	wg := sync.WaitGroup{}
	wg.Add(checkers + switchers)
	for worker := 0; worker < checkers; worker++ {
		go func() {
			defer wg.Done()
			for atomic.LoadUint32(&stop) == 0 {
				var prevPrefix uint32
				var prevKey uint32
				prefixN := rand.Uint32() % prefixes
				prefix := bigendian.Uint32ToBytes(prefixN)
				keyN := rand.Uint32() % prefixes
				start := bigendian.Uint32ToBytes(keyN)
				prevKey = keyN - 1
				if rand.Intn(10) == 0 {
					start = nil
					prevKey = 0
					prevKey--
					if rand.Intn(2) == 0 {
						prefix = nil
						prevPrefix = 0
					}
				}
				it := switchable.NewIterator(prefix, start)
				require.NoError(it.Error())
				for it.Next() {
					require.NoError(it.Error())
					require.Equal(8, len(it.Key()))
					require.Equal(8, len(it.Value()))
					p, k := decodePair(it.Key())
					sp, sk := decodePair(it.Value())
					require.Equal(p*p, sp)
					require.Equal(k*k, sk)
					if prefix != nil {
						require.Equal(prefixN, p)
					} else if p != prevPrefix {
						require.Equal(prevPrefix+1, p)
						prevPrefix = p
						prevKey = 0
						prevKey--
					}

					require.Equal(prevKey+1, k, prefix)
					prevKey = k
				}
				require.NoError(it.Error())
				it.Release()
			}
		}()
	}
	for worker := 0; worker < switchers; worker++ {
		go func() {
			defer wg.Done()

			for atomic.LoadUint32(&stop) == 0 {
				snap, err := memdb.GetSnapshot()
				require.NoError(err)
				old := switchable.SwitchTo(&UncallableAfterRelease{
					Snapshot: snap,
				})
				old.Release()
			}
		}()
	}
	time.Sleep(duration)
	atomic.StoreUint32(&stop, 1)
	wg.Wait()
	switchable.Release()
	require.NoError(memdb.Close())
}
