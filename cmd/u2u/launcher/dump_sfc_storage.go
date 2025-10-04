package launcher

import (
	"os"

	"github.com/unicornultrafoundation/go-helios/hash"
	"gopkg.in/urfave/cli.v1"

	"github.com/unicornultrafoundation/go-u2u/cmd/utils"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/native/ibr"
	"github.com/unicornultrafoundation/go-u2u/u2u"
	"github.com/unicornultrafoundation/go-u2u/utils/caution"
)

// dumpSfcStorage is the 'db dump-sfc' command.
func dumpSfcStorage(ctx *cli.Context) (err error) {
	if !ctx.Bool(experimentalFlag.Name) {
		utils.Fatalf("Add --experimental flag")
	}
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.Lvl(ctx.GlobalInt(verbosityFlag.Name)))
	log.Root().SetHandler(glogger)

	cfg := makeAllConfigs(ctx)
	cfg.U2UStore.EVM.SfcEnabled = true

	rawDbs := makeDirectDBsProducer(cfg)
	defer caution.CloseAndReportError(&err, rawDbs, "failed to close raw DBs")
	gdb := makeGossipStore(rawDbs, cfg)
	defer caution.CloseAndReportError(&err, gdb, "failed to close Gossip DB")
	evms := gdb.EvmStore()
	defer caution.CloseAndReportError(&err, evms, "failed to close EVM store")

	latestBlockIndex := gdb.GetLatestBlockIndex()
	latestBlockHash := gdb.GetBlock(latestBlockIndex).Atropos
	stateRoot := evms.GetSfcStateRoot(latestBlockIndex, latestBlockHash.Bytes())
	if stateRoot != nil && evms.HasSfcStateDB(hash.Hash(*stateRoot)) {
		log.Info("Already dump SFC contract's storage at this block", "block", latestBlockIndex, "stateRoot", stateRoot.Hex())
		// TODO(trinhdn97): must prevent re-dumping the SFC contracts' storage after completely removed the EVM version
		// return nil
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
		// Save dumped SFC contract's storage to disk
		if err := evms.CommitSfcState(latestBlockIndex, hash.Hash(sfcStateTrieHash), false); err != nil {
			panic(err)
		}
		evms.CapSfcState()

		// Save the sfcStateTrieHash for future use (include into the native block header)
		evms.SetSfcStateRoot(latestBlockIndex, latestBlockHash.Bytes(), sfcStateTrieHash)
		evms.SetLatestSfcStateRoot(sfcStateTrieHash)

		latestBlock := gdb.GetBlock(latestBlockIndex)
		latestBlock.SfcStateRoot = hash.Hash(sfcStateTrieHash)
		gdb.SetBlock(latestBlockIndex, latestBlock)
		bs, es := gdb.GetBlockEpochState()
		bs.SfcStateRoot = hash.Hash(sfcStateTrieHash)
		gdb.SetBlockEpochState(bs, es)
		gdb.FlushBlockEpochState()
		br := gdb.GetFullBlockRecord(latestBlockIndex)
		br.SfcStateRoot = hash.Hash(sfcStateTrieHash)
		gdb.WriteFullBlockRecord(ibr.LlrIdxFullBlockRecord{
			LlrFullBlockRecord: *br,
			Idx:                latestBlockIndex,
		})

		log.Info("Dump SFC contract's storage successfully at", "block", latestBlockIndex, "stateTrieHash", sfcStateTrieHash.Hex())
	}
	// TODO(trinhdn97): should be done more gracefully for not corrupting the DB after dumping.
	return nil
}

func dumpSfcStorageByStateDb(stateDb *state.StateDB, sfcStateDb *state.StateDB) (common.Hash, error) {
	for sfcAddress := range u2u.DefaultVMConfig.SfcPrecompiles {
		// Warming up the state object to persist later
		sfcStateDb.AddBalance(sfcAddress, stateDb.GetBalance(sfcAddress))
		sfcStateDb.SetNonce(sfcAddress, stateDb.GetNonce(sfcAddress))
		sfcStateDb.SetCode(sfcAddress, stateDb.GetCode(sfcAddress))
		stateDb.ForEachStorage(sfcAddress, func(key, value common.Hash) bool {
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
