package launcher

import (
	"compress/gzip"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/status-im/keycard-go/hexutils"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/unicornultrafoundation/go-u2u/cmd/utils"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/rlp"

	"github.com/unicornultrafoundation/go-u2u/consensus/hash"
	"github.com/unicornultrafoundation/go-u2u/consensus/native/idx"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb/batched"
	"github.com/unicornultrafoundation/go-u2u/consensus/u2udb/pebble"

	"github.com/unicornultrafoundation/go-u2u/gossip"
	"github.com/unicornultrafoundation/go-u2u/utils/dbutil/autocompact"
)

var (
	eventsFileHeader  = hexutils.HexToBytes("7e995678")
	eventsFileVersion = hexutils.HexToBytes("00010001")
)

// statsReportLimit is the time limit during import and export after which we
// always print out progress. This avoids the user wondering what's going on.
const statsReportLimit = 8 * time.Second

func exportEvents(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		utils.Fatalf("This command requires an argument.")
	}

	cfg := makeAllConfigs(ctx)

	rawDbs := makeDirectDBsProducer(cfg)
	gdb := makeGossipStore(rawDbs, cfg)
	defer gdb.Close()

	fn := ctx.Args().First()

	// Open the file handle and potentially wrap with a gzip stream
	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer fh.Close()

	var writer io.Writer = fh
	if strings.HasSuffix(fn, ".gz") {
		writer = gzip.NewWriter(writer)
		defer writer.(*gzip.Writer).Close()
	}

	from := idx.Epoch(1)
	if len(ctx.Args()) > 1 {
		n, err := strconv.ParseUint(ctx.Args().Get(1), 10, 32)
		if err != nil {
			return err
		}
		from = idx.Epoch(n)
	}
	to := idx.Epoch(0)
	if len(ctx.Args()) > 2 {
		n, err := strconv.ParseUint(ctx.Args().Get(2), 10, 32)
		if err != nil {
			return err
		}
		to = idx.Epoch(n)
	}

	log.Info("Exporting events to file", "file", fn)
	// Write header and version
	_, err = writer.Write(append(eventsFileHeader, eventsFileVersion...))
	if err != nil {
		return err
	}
	err = exportTo(writer, gdb, from, to)
	if err != nil {
		utils.Fatalf("Export error: %v\n", err)
	}

	return nil
}

// exportTo writer the active chain.
func exportTo(w io.Writer, gdb *gossip.Store, from, to idx.Epoch) (err error) {
	start, reported := time.Now(), time.Time{}

	var (
		counter int
		last    hash.Event
	)
	gdb.ForEachEventRLP(from.Bytes(), func(id hash.Event, event rlp.RawValue) bool {
		if to >= from && id.Epoch() > to {
			return false
		}
		counter++
		_, err = w.Write(event)
		if err != nil {
			return false
		}
		last = id
		if counter%100 == 1 && time.Since(reported) >= statsReportLimit {
			log.Info("Exporting events", "last", last.String(), "exported", counter, "elapsed", common.PrettyDuration(time.Since(start)))
			reported = time.Now()
		}
		return true
	})
	log.Info("Exported events", "last", last.String(), "exported", counter, "elapsed", common.PrettyDuration(time.Since(start)))

	return
}

func exportEvmKeys(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		utils.Fatalf("This command requires an argument.")
	}

	cfg := makeAllConfigs(ctx)

	rawDbs := makeDirectDBsProducer(cfg)
	gdb := makeGossipStore(rawDbs, cfg)
	defer gdb.Close()

	fn := ctx.Args().First()

	keysDB_, err := pebble.New(fn, 1024*opt.MiB, utils.MakeDatabaseHandles()/2, nil, nil)
	if err != nil {
		return err
	}
	keysDB := batched.Wrap(autocompact.Wrap2M(keysDB_, opt.GiB, 16*opt.GiB, true, "evm-keys"))
	defer keysDB.Close()

	it := gdb.EvmStore().EvmDb.NewIterator(nil, nil)
	// iterate only over MPT data
	it = mptAndPreimageIterator{it}
	defer it.Release()

	log.Info("Exporting EVM keys", "dir", fn)
	for it.Next() {
		if err := keysDB.Put(it.Key(), []byte{0}); err != nil {
			return err
		}
	}
	log.Info("Exported EVM keys", "dir", fn)
	return nil
}
