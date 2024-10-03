package launcher

import (
	"gopkg.in/urfave/cli.v1"

	"github.com/unicornultrafoundation/go-u2u/cmd/utils"
)

// dumpSfcStorage is the 'db dump-sfc' command.
func dumpSfcStorage(ctx *cli.Context) error {
	if !ctx.Bool(experimentalFlag.Name) {
		utils.Fatalf("Add --experimental flag")
	}
	cfg := makeAllConfigs(ctx)

	rawDbs := makeDirectDBsProducer(cfg)
	gdb := makeGossipStore(rawDbs, cfg)
	defer gdb.Close()

	evms := gdb.EvmStore()
	currentBlock := gdb.GetBlock(gdb.GetLatestBlockIndex())
	stateDb, err := evms.StateDB(currentBlock.Root)
	stateDb.ForEachStorage()

	return err
}
