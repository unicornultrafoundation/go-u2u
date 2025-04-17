package launcher

import (
	"os"

	"github.com/unicornultrafoundation/go-u2u/helios/hash"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb"
	"gopkg.in/urfave/cli.v1"

	"github.com/unicornultrafoundation/go-u2u/cmd/utils"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/u2u"
)

// dumpSfcStorage is the 'db dump-sfc' command.
func dumpSfcStorage(ctx *cli.Context) error {
	if !ctx.Bool(experimentalFlag.Name) {
		utils.Fatalf("Add --experimental flag")
	}
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.Lvl(ctx.GlobalInt(verbosityFlag.Name)))
	log.Root().SetHandler(glogger)

	cfg := makeAllConfigs(ctx)
	cfg.U2UStore.EVM.SfcEnabled = true

	rawDbs := makeDirectDBsProducer(cfg)
	gdb := makeGossipStore(rawDbs, cfg)

	evms := gdb.EvmStore()

	defer evms.Close()
	defer gdb.Close()
	defer func(rawDbs u2udb.FullDBProducer) {
		err := rawDbs.Close()
		if err != nil {
			log.Error("Error closing raw dbs", "err", err)
		}
	}(rawDbs)

	latestBlockIndex := gdb.GetLatestBlockIndex()
	latestBlockHash := gdb.GetBlock(latestBlockIndex).Atropos
	stateRoot := evms.GetSfcStateRoot(latestBlockIndex, latestBlockHash.Bytes())
	if stateRoot != nil && evms.HasSfcStateDB(hash.Hash(*stateRoot)) {
		log.Info("Already dump SFC contract's storage at this block", "block", latestBlockIndex, "stateRoot", stateRoot.Hex())
		return nil
	}

	currentBlock := gdb.GetBlock(latestBlockIndex)
	stateDb, err := evms.StateDB(currentBlock.Root)
	if err != nil {
		return err
	}
	sfcStateDb, err := evms.SfcStateDB(hash.Zero)
	if err != nil {
		return err
	}

	sfcStateTrieHash, err := dumpSfcStorageByStateDb(stateDb, sfcStateDb)

	if err != nil {
		log.Error("Dump SFC contract's storage unsuccessfully", "err", err)
		return err
	}

	if isDumpStorageHashValid(stateDb, sfcStateDb) {
		// Save dump SFC contract's storage to disk
		evms.CommitSfcState(latestBlockIndex, hash.Hash(sfcStateTrieHash), false)
		evms.CapSfcState()

		// Save the stateTrieHash for future use (include into the block header)
		evms.SetSfcStateRoot(latestBlockIndex, latestBlockHash.Bytes(), sfcStateTrieHash)

		log.Info("Dump SFC contract's storage successfully at", "block", latestBlockIndex, "stateTrieHash", sfcStateTrieHash.Hex())
	}
	// TODO(trinhdn97): should be done more gracefully for not corrupting the DB after dumping.
	return nil
}

func dumpSfcStorageByStateDb(stateDb *state.StateDB, sfcStateDb *state.StateDB) (common.Hash, error) {
	for sfcAddress := range u2u.DefaultVMConfig.SfcPrecompiles {
		stateDb.ForEachStorage(sfcAddress, func(key, value common.Hash) bool {
			log.Debug("Looping on storage trie", "Contract", sfcAddress, "Key", key.Hex(), "Value", value.Hex())
			sfcStateDb.SetState(sfcAddress, key, value)
			return true
		})
	}

	hash, err := sfcStateDb.Commit(false)

	return hash, err
}

// Verify if storageHash of each SFC contract address between originDb & dumpDb are the same
func isDumpStorageHashValid(stateDb *state.StateDB, sfcStateDb *state.StateDB) bool {
	isValid := true
	for sfcAddress := range u2u.DefaultVMConfig.SfcPrecompiles {
		originTrie := stateDb.StorageTrie(sfcAddress)
		dumpTrie := sfcStateDb.StorageTrie(sfcAddress)
		if originTrie.Hash() != dumpTrie.Hash() {
			isValid = false
			log.Error("Storage hashes are NOT the same", "ContractAddr", sfcAddress)
		}
		log.Info("Checking contracts storage root",
			"addr", sfcAddress.Hex(),
			"origin", originTrie.Hash().Hex(),
			"dumped", dumpTrie.Hash().Hex())
	}
	return isValid
}
