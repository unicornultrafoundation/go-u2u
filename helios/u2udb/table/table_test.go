package table

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/status-im/keycard-go/hexutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/go-u2u/common"

	u2udb "github.com/unicornultrafoundation/go-u2u/helios/u2udb"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb/flushable"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb/leveldb"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb/memorydb"
)

func tempLevelDB(name string) *leveldb.Database {
	dir, err := ioutil.TempDir("", "flushable-test"+name)
	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory: %v", err))
	}

	drop := func() {
		err := os.RemoveAll(dir)
		if err != nil {
			panic(err)
		}
	}

	diskdb, err := leveldb.New(dir, 16, 0, nil, drop)
	if err != nil {
		panic(fmt.Sprintf("can't create temporary database: %v", err))
	}
	return diskdb
}

func TestTable(t *testing.T) {
	prefix0 := map[string][]byte{
		"00": {0},
		"01": {0, 1},
		"02": {0, 1, 2},
		"03": {0, 1, 2, 3},
	}
	prefix1 := map[string][]byte{
		"10": {0, 1, 2, 3, 4},
	}
	testData := join(prefix0, prefix1)

	// open raw databases
	leveldb1 := tempLevelDB("1")
	defer leveldb1.Drop()
	defer leveldb1.Close()

	leveldb2 := tempLevelDB("2")
	defer leveldb2.Drop()
	defer leveldb2.Close()

	for name, db := range map[string]u2udb.Store{
		"memory":                       memorydb.New(),
		"leveldb":                      leveldb1,
		"cache-over-leveldb":           flushable.Wrap(leveldb2),
		"cache-over-cache-over-memory": flushable.Wrap(memorydb.New()),
	} {
		t.Run(name, func(t *testing.T) {
			assertar := assert.New(t)

			// tables
			t1 := New(db, []byte("t1"))
			tables := map[string]u2udb.Store{
				"/t1":      t1,
				"/x/t1/t2": New(db, []byte("x")).NewTable([]byte("t1t2")),
				"/t2":      New(db, []byte("t2")),
			}

			// write
			for name, t := range tables {
				for k, v := range testData {
					err := t.Put([]byte(k), v)
					if !assertar.NoError(err, name) {
						return
					}
				}
			}

			// read
			for name, t := range tables {

				for pref, count := range map[string]int{
					"0": len(prefix0),
					"1": len(prefix1),
					"":  len(prefix0) + len(prefix1),
				} {
					got := 0
					var prevKey []byte

					it := t.NewIterator([]byte(pref), nil)
					defer it.Release()
					for it.Next() {
						if prevKey == nil {
							prevKey = common.CopyBytes(it.Key())
						} else {
							assertar.Equal(1, bytes.Compare(it.Key(), prevKey))
						}
						got++
						assertar.Equal(
							testData[string(it.Key())],
							it.Value(),
							name+": "+string(it.Key()),
						)
					}

					if !assertar.NoError(it.Error()) {
						return
					}

					if !assertar.Equal(count, got) {
						return
					}
				}
			}
		})
	}
}

func TestPrefixInc(t *testing.T) {
	require.Nil(t, incPrefix(hexutils.HexToBytes("ff")))
	require.Equal(t, hexutils.HexToBytes("ff"), incPrefix(hexutils.HexToBytes("fe")))
	require.Equal(t, hexutils.HexToBytes("02"), incPrefix(hexutils.HexToBytes("01")))
	require.Equal(t, hexutils.HexToBytes("01"), incPrefix(hexutils.HexToBytes("00")))

	require.Equal(t, hexutils.HexToBytes("0100"), incPrefix(hexutils.HexToBytes("00ff")))
	require.Equal(t, hexutils.HexToBytes("00ff"), incPrefix(hexutils.HexToBytes("00fe")))
	require.Equal(t, hexutils.HexToBytes("0002"), incPrefix(hexutils.HexToBytes("0001")))
	require.Equal(t, hexutils.HexToBytes("0001"), incPrefix(hexutils.HexToBytes("0000")))

	require.Nil(t, incPrefix(hexutils.HexToBytes("ffff")))
	require.Equal(t, hexutils.HexToBytes("ffff"), incPrefix(hexutils.HexToBytes("fffe")))
	require.Equal(t, hexutils.HexToBytes("ff02"), incPrefix(hexutils.HexToBytes("ff01")))
	require.Equal(t, hexutils.HexToBytes("ff01"), incPrefix(hexutils.HexToBytes("ff00")))
}

func join(aa ...map[string][]byte) map[string][]byte {
	res := make(map[string][]byte)
	for _, a := range aa {
		for k, v := range a {
			res[k] = v
		}
	}

	return res
}
