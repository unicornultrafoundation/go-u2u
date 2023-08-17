package gentxs

import (
	"gopkg.in/urfave/cli.v1"
)

var ConfigFileFlag = cli.StringFlag{
	Name:  "config",
	Usage: "TOML configuration file",
	Value: "txsgen.toml",
}

var TpsLimitFlag = cli.Float64Flag{
	Name:  "tpslimit",
	Usage: "transactions per second limit",
	Value: -1.0,
}

var KeyStoreDirFlag = cli.StringFlag{
	Name:  "keystore",
	Usage: "Directory for the keystore",
	Value: "keys_txsgen",
}

var VerbosityFlag = cli.IntFlag{
	Name:  "verbosity",
	Usage: "sets the verbosity level",
	Value: 3,
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


func getTpsLimit(ctx *cli.Context) float64 {
	return ctx.GlobalFloat64(TpsLimitFlag.Name)
}
