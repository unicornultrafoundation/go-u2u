package launcher

import (
	"fmt"
	"path"

	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/unicornultrafoundation/go-helios/u2udb"
	"github.com/unicornultrafoundation/go-helios/u2udb/cachedproducer"
	"github.com/unicornultrafoundation/go-helios/u2udb/multidb"
	"github.com/unicornultrafoundation/go-u2u/cmd/utils"
	"github.com/unicornultrafoundation/go-u2u/ethdb"
	"github.com/unicornultrafoundation/go-u2u/log"
	"gopkg.in/urfave/cli.v1"

	"github.com/unicornultrafoundation/go-u2u/gossip"
	"github.com/unicornultrafoundation/go-u2u/integration"
	"github.com/unicornultrafoundation/go-u2u/utils/dbutil/compactdb"
)

var (
	experimentalFlag = cli.BoolFlag{
		Name:  "experimental",
		Usage: "Allow experimental DB fixing",
	}
	dbCommand = cli.Command{
		Name:        "db",
		Usage:       "A set of commands related to leveldb database",
		Category:    "DB COMMANDS",
		Description: "",
		Subcommands: []cli.Command{
			{
				Name:      "compact",
				Usage:     "Compact all databases",
				ArgsUsage: "",
				Action:    utils.MigrateFlags(compact),
				Category:  "DB COMMANDS",
				Flags: []cli.Flag{
					utils.DataDirFlag,
				},
				Description: `
u2u db compact
will compact all databases under datadir's chaindata.
`,
			},
			{
				Name:      "transform",
				Usage:     "Transform DBs layout",
				ArgsUsage: "",
				Action:    utils.MigrateFlags(dbTransform),
				Category:  "DB COMMANDS",
				Flags: []cli.Flag{
					utils.DataDirFlag,
				},
				Description: `
u2u db transform
will migrate tables layout according to the configuration.
`,
			},
			{
				Name:      "heal",
				Usage:     "Experimental - try to heal dirty DB",
				ArgsUsage: "",
				Action:    utils.MigrateFlags(healDirty),
				Category:  "DB COMMANDS",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					experimentalFlag,
				},
				Description: `
u2u db heal --experimental
Experimental - try to heal dirty DB.
`,
			},
			{
				Name:      "dump-sfc",
				Usage:     "Experimental - try to dump the storage of SFC contract to KVDB",
				ArgsUsage: "",
				Action:    utils.MigrateFlags(dumpSfcStorage),
				Category:  "DB COMMANDS",
				Flags: []cli.Flag{
					utils.DataDirFlag,
					experimentalFlag,
				},
				Description: `
u2u db dump-sfc --experimental
Experimental - try to dump the storage of SFC contract to KVDB.
`,
			},
		},
	}
)

func makeUncheckedDBsProducers(cfg *config) map[multidb.TypeName]u2udb.IterableDBProducer {
	dbsList, _ := integration.SupportedDBs(path.Join(cfg.Node.DataDir, "chaindata"), cfg.DBs.RuntimeCache)
	return dbsList
}

func makeUncheckedCachedDBsProducers(chaindataDir string) map[multidb.TypeName]u2udb.FullDBProducer {
	dbTypes, _ := integration.SupportedDBs(chaindataDir, integration.DBsCacheConfig{
		Table: map[string]integration.DBCacheConfig{
			"": {
				Cache:   1024 * opt.MiB,
				Fdlimit: uint64(utils.MakeDatabaseHandles() / 2),
			},
		},
	})
	wrappedDbTypes := make(map[multidb.TypeName]u2udb.FullDBProducer)
	for typ, producer := range dbTypes {
		wrappedDbTypes[typ] = cachedproducer.WrapAll(&integration.DummyScopedProducer{IterableDBProducer: producer})
	}
	return wrappedDbTypes
}

func makeCheckedDBsProducers(cfg *config) map[multidb.TypeName]u2udb.IterableDBProducer {
	if err := integration.CheckStateInitialized(path.Join(cfg.Node.DataDir, "chaindata"), cfg.DBs); err != nil {
		utils.Fatalf(err.Error())
	}
	return makeUncheckedDBsProducers(cfg)
}

func makeDirectDBsProducerFrom(dbsList map[multidb.TypeName]u2udb.IterableDBProducer, cfg *config) u2udb.FullDBProducer {
	multiRawDbs, err := integration.MakeDirectMultiProducer(dbsList, cfg.DBs.Routing)
	if err != nil {
		utils.Fatalf("Failed to initialize multi DB producer: %v", err)
	}
	return multiRawDbs
}

func makeDirectDBsProducer(cfg *config) u2udb.FullDBProducer {
	dbsList := makeCheckedDBsProducers(cfg)
	return makeDirectDBsProducerFrom(dbsList, cfg)
}

func makeGossipStore(producer u2udb.FlushableDBProducer, cfg *config) *gossip.Store {
	return gossip.NewStore(producer, cfg.U2UStore)
}

func compact(ctx *cli.Context) error {

	cfg := makeAllConfigs(ctx)

	producers := makeCheckedDBsProducers(cfg)
	for typ, p := range producers {
		for _, name := range p.Names() {
			if err := compactDB(typ, name, p); err != nil {
				return err
			}
		}
	}

	return nil
}

func compactDB(typ multidb.TypeName, name string, producer u2udb.DBProducer) error {
	humanName := path.Join(string(typ), name)
	db, err := producer.OpenDB(name)
	defer db.Close()
	if err != nil {
		log.Error("Cannot open db or db does not exists", "db", humanName)
		return err
	}

	log.Info("Stats before compaction", "db", humanName)
	showDbStats(db)

	err = compactdb.Compact(db, humanName, 64*opt.GiB)
	if err != nil {
		log.Error("Database compaction failed", "err", err)
		return err
	}

	log.Info("Stats after compaction", "db", humanName)
	showDbStats(db)

	return nil
}

func showDbStats(db ethdb.KeyValueStater) {
	if stats, err := db.Stat("stats"); err != nil {
		log.Warn("Failed to read database stats", "error", err)
	} else {
		fmt.Println(stats)
	}
	if ioStats, err := db.Stat("iostats"); err != nil {
		log.Warn("Failed to read database iostats", "error", err)
	} else {
		fmt.Println(ioStats)
	}
}
