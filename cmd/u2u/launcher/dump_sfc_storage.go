package launcher

import (
	"gopkg.in/urfave/cli.v1"

	"github.com/unicornultrafoundation/go-helios/hash"
	"github.com/unicornultrafoundation/go-u2u/cmd/utils"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/driver"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/driverauth"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/sfc"
)

var (
	sfcContractAddresses = []common.Address{sfc.ContractAddress, driverauth.ContractAddress, driver.ContractAddress}
)

// dumpSfcStorage is the 'db dump-sfc' command.
func dumpSfcStorage(ctx *cli.Context) error {
	if !ctx.Bool(experimentalFlag.Name) {
		utils.Fatalf("Add --experimental flag")
	}
	cfg := makeAllConfigs(ctx)
	cfg.U2UStore.SFC.Enable = true

	rawDbs := makeDirectDBsProducer(cfg)
	gdb := makeGossipStore(rawDbs, cfg)
	defer gdb.Close()

	evms := gdb.EvmStore()
	sfcs := gdb.SfcStore()

	latestBlockIndex := gdb.GetLatestBlockIndex()
	stateRoot := sfcs.GetStateRoot(latestBlockIndex)
	if stateRoot != nil && sfcs.HasStateDB(hash.Hash(*stateRoot)) {
		log.Info("Already dump SFC contract's storage with this block", "block", latestBlockIndex, "stateRoot", stateRoot)
		return nil
	}

	currentBlock := gdb.GetBlock(latestBlockIndex)
	stateDb, err := evms.StateDB(currentBlock.Root)
	if err != nil {
		return err
	}
	sfcStateDb, err := sfcs.StateDB(hash.Zero)
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
		sfcs.Commit(latestBlockIndex, hash.Hash(sfcStateTrieHash), false)
		sfcs.Cap()

		// Save the stateTrieHash for future use (include into the block header)
		sfcs.SetStateRoot(latestBlockIndex, sfcStateTrieHash)

		log.Info("Dump SFC contract's storage successfully", "stateTrieHash", sfcStateTrieHash)
	}

	return nil
}

func dumpSfcStorageByStateDb(stateDb *state.StateDB, sfcStateDb *state.StateDB) (common.Hash, error) {
	for _, sfcAddress := range sfcContractAddresses {
		stateDb.ForEachStorage(sfcAddress, func(key, value common.Hash) bool {
			log.Debug("Looping on storage trie", "Contract", sfcAddress, "Key", key, "Value", value)
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

	for _, sfcAddress := range sfcContractAddresses {
		originTrie := stateDb.StorageTrie(sfcAddress)
		dumpTrie := sfcStateDb.StorageTrie(sfcAddress)

		if originTrie.Hash() != dumpTrie.Hash() {
			isValid = false
			log.Error("Storage hashes are NOT the same", "ContractAddr", sfcAddress)
		}
	}

	return isValid
}
