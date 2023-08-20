package launcher

import (
	"fmt"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/unicornultrafoundation/go-hashgraph/common/bigendian"
	"github.com/unicornultrafoundation/go-hashgraph/consensus"
	"github.com/unicornultrafoundation/go-hashgraph/hash"
	"github.com/unicornultrafoundation/go-hashgraph/native/idx"
	"github.com/unicornultrafoundation/go-hashgraph/u2udb"
	"github.com/unicornultrafoundation/go-hashgraph/u2udb/batched"
	"github.com/unicornultrafoundation/go-hashgraph/u2udb/flushable"

	"github.com/unicornultrafoundation/go-u2u/gossip"
	"github.com/unicornultrafoundation/go-u2u/integration"
	"github.com/unicornultrafoundation/go-u2u/native/iblockproc"
)

// maxEpochsToTry represents amount of last closed epochs to try (in case that the last one has the state unavailable)
const maxEpochsToTry = 10000

// healDirty is the 'db heal' command.
func healDirty(ctx *cli.Context) error {
	if !ctx.Bool(experimentalFlag.Name) {
		utils.Fatalf("Add --experimental flag")
	}
	cfg := makeAllConfigs(ctx)

	log.Info("Opening databases")
	dbTypes := makeUncheckedDBsProducers(cfg)
	multiProducer := makeDirectDBsProducerFrom(dbTypes, cfg)

	// reverts the gossip database state
	epochState, topEpoch, err := fixDirtyGossipDb(multiProducer, cfg)
	if err != nil {
		return err
	}

	// drop epoch-related databases and consensus database
	log.Info("Removing epoch DBs - will be recreated on next start")
	for _, name := range []string{
		fmt.Sprintf("gossip-%d", topEpoch),
		fmt.Sprintf("hashgraph-%d", topEpoch),
		"hashgraph",
	} {
		err = eraseTable(name, multiProducer)
		if err != nil {
			return err
		}
	}

	// prepare consensus database from epochState
	log.Info("Recreating hashgraph DB")
	cMainDb := mustOpenDB(multiProducer, "hashgraph")
	cGetEpochDB := func(epoch idx.Epoch) u2udb.Store {
		return mustOpenDB(multiProducer, fmt.Sprintf("hashgraph-%d", epoch))
	}
	cdb := consensus.NewStore(cMainDb, cGetEpochDB, panics("hashgraph store"), cfg.HashgraphStore)
	err = cdb.ApplyGenesis(&consensus.Genesis{
		Epoch:      epochState.Epoch,
		Validators: epochState.Validators,
	})
	if err != nil {
		return fmt.Errorf("failed to init consensus database: %v", err)
	}
	_ = cdb.Close()

	log.Info("Clearing DBs dirty flags")
	id := bigendian.Uint64ToBytes(uint64(time.Now().UnixNano()))
	for typ, producer := range dbTypes {
		err := clearDirtyFlags(id, producer)
		if err != nil {
			log.Crit("Failed to write clean FlushID", "type", typ, "err", err)
		}
	}

	log.Info("Recovery is complete")
	return nil
}

// fixDirtyGossipDb reverts the gossip database into state, when was one of last epochs sealed
func fixDirtyGossipDb(producer u2udb.FlushableDBProducer, cfg *config) (
	epochState *iblockproc.EpochState, topEpoch idx.Epoch, err error) {
	gdb := makeGossipStore(producer, cfg) // requires FlushIDKey present (not clean) in all dbs
	defer gdb.Close()
	topEpoch = gdb.GetEpoch()

	// find the last closed epoch with the state available
	epochIdx, blockState, epochState := getLastEpochWithState(gdb, maxEpochsToTry)
	if blockState == nil || epochState == nil {
		return nil, 0, fmt.Errorf("state for last %d closed epochs is pruned, recovery isn't possible", maxEpochsToTry)
	}

	// set the historic state to be the current
	log.Info("Reverting to epoch state", "epoch", epochIdx)
	gdb.SetBlockEpochState(*blockState, *epochState)
	gdb.FlushBlockEpochState()

	// Service.switchEpochTo
	gdb.SetHighestLamport(0)
	gdb.FlushHighestLamport()

	// removing excessive events (event epoch >= closed epoch)
	log.Info("Removing excessive events")
	gdb.ForEachEventRLP(epochIdx.Bytes(), func(id hash.Event, _ rlp.RawValue) bool {
		gdb.DelEvent(id)
		return true
	})

	return epochState, topEpoch, nil
}

// getLastEpochWithState finds the last closed epoch with the state available
func getLastEpochWithState(gdb *gossip.Store, epochsToTry idx.Epoch) (epochIdx idx.Epoch, blockState *iblockproc.BlockState, epochState *iblockproc.EpochState) {
	currentEpoch := gdb.GetEpoch()
	endEpoch := idx.Epoch(1)
	if currentEpoch > epochsToTry {
		endEpoch = currentEpoch - epochsToTry
	}

	for epochIdx = currentEpoch; epochIdx > endEpoch; epochIdx-- {
		blockState, epochState = gdb.GetHistoryBlockEpochState(epochIdx)
		if blockState == nil || epochState == nil {
			log.Debug("Epoch is not available", "epoch", epochIdx)
			continue
		}
		if !gdb.EvmStore().HasStateDB(blockState.FinalizedStateRoot) {
			log.Debug("EVM state for the epoch is not available", "epoch", epochIdx)
			continue
		}
		log.Debug("Latest epoch with available state found", "epoch", epochIdx)
		return epochIdx, blockState, epochState
	}

	return 0, nil, nil
}

func eraseTable(name string, producer u2udb.IterableDBProducer) error {
	log.Info("Cleaning table", "name", name)
	db, err := producer.OpenDB(name)
	if err != nil {
		return fmt.Errorf("unable to open DB %s; %s", name, err)
	}
	db = batched.Wrap(db)
	defer db.Close()
	it := db.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		err := db.Delete(it.Key())
		if err != nil {
			return err
		}
	}
	return nil
}

// clearDirtyFlags - writes the CleanPrefix into all databases
func clearDirtyFlags(id []byte, rawProducer u2udb.IterableDBProducer) error {
	names := rawProducer.Names()
	for _, name := range names {
		db, err := rawProducer.OpenDB(name)
		if err != nil {
			return err
		}

		err = db.Put(integration.FlushIDKey, append([]byte{flushable.CleanPrefix}, id...))
		if err != nil {
			log.Crit("Failed to write CleanPrefix", "name", name)
			return err
		}
		log.Info("Database set clean", "name", name)
		_ = db.Close()
	}
	return nil
}

func mustOpenDB(producer u2udb.DBProducer, name string) u2udb.Store {
	db, err := producer.OpenDB(name)
	if err != nil {
		utils.Fatalf("Failed to open '%s' database: %v", name, err)
	}
	return db
}

func panics(name string) func(error) {
	return func(err error) {
		log.Crit(fmt.Sprintf("%s error", name), "err", err)
	}
}
