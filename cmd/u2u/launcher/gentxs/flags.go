package gentxs

import (
	"gopkg.in/urfave/cli.v1"
)

var NetworkConfigFileFlag = cli.StringFlag{
	Name:  "networkconfig",
	Usage: "TOML configuration file",
	Value: "txsgen.toml",
}

var TpsLimitFlag = cli.Float64Flag{
	Name:  "tpslimit",
	Usage: "transactions per second limit",
	Value: -1.0,
}

var AccKeyStoreDirFlag = cli.StringFlag{
	Name:  "acckeystore",
	Usage: "Directory for the keystore",
	Value: "keys_txsgen",
}

var GenerateAccountFlag = cli.IntFlag{
	Name:  "fakeaccs",
	Usage: "Generates fakenet accounts and saves them in the keystore dir.",
	Value: 1000,
}

var GenerateAccountBalanceFlag = cli.IntFlag{
	Name:  "fakebalance",
	Usage: "Pays from config.Payer to each other account in the keystore dir.",
	Value: 1,
}

var GenerateTxTransferFlag = cli.StringFlag{
	Name:  "faketransfers",
	Usage: "Generates a lot of transfer txs between accounts in the keystore dir (except config.Payer).",
}

func getTpsLimit(ctx *cli.Context) float64 {
	return ctx.GlobalFloat64(TpsLimitFlag.Name)
}
