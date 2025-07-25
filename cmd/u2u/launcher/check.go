package launcher

import (
	"time"

	"github.com/unicornultrafoundation/go-helios/native/idx"
	"gopkg.in/urfave/cli.v1"

	"github.com/unicornultrafoundation/go-u2u/cmd/utils"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/native"
	"github.com/unicornultrafoundation/go-u2u/utils/caution"
)

func checkEvm(ctx *cli.Context) (err error) {
	if len(ctx.Args()) != 0 {
		utils.Fatalf("This command doesn't require an argument.")
	}

	cfg := makeAllConfigs(ctx)

	rawDbs := makeDirectDBsProducer(cfg)
	gdb := makeGossipStore(rawDbs, cfg)
	defer caution.CloseAndReportError(&err, gdb, "failed to close Gossip DB")
	evms := gdb.EvmStore()
	defer caution.CloseAndReportError(&err, evms, "failed to close EVM store")

	start, reported := time.Now(), time.Now()

	var prevPoint idx.Block
	var prevIndex idx.Block
	checkBlocks := func(stateOK func(root common.Hash) (bool, error)) {
		var (
			lastIdx            = gdb.GetLatestBlockIndex()
			prevPointRootExist bool
		)
		gdb.ForEachBlock(func(index idx.Block, block *native.Block) {
			prevIndex = index
			found, err := stateOK(common.Hash(block.Root))
			if found != prevPointRootExist {
				if index > 0 && found {
					log.Warn("EVM history is pruned", "fromBlock", prevPoint, "toBlock", index-1)
				}
				prevPointRootExist = found
				prevPoint = index
			}
			if index == lastIdx && !found {
				log.Crit("State trie for the latest block is not found", "block", index)
			}
			if !found {
				return
			}
			if err != nil {
				log.Crit("State trie error", "err", err, "block", index)
			}
			if time.Since(reported) >= statsReportLimit {
				log.Info("Checking presence of every node", "last", index, "pruned", !prevPointRootExist, "elapsed", common.PrettyDuration(time.Since(start)))
				reported = time.Now()
			}
		})
	}

	if err := evms.CheckEvm(checkBlocks, true); err != nil {
		return err
	}
	log.Info("EVM storage is verified", "last", prevIndex, "elapsed", common.PrettyDuration(time.Since(start)))
	return nil
}
