package gentxs

import (
	"gopkg.in/urfave/cli.v1"
)

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


func getTpsLimit(ctx *cli.Context) float64 {
	return ctx.GlobalFloat64(TpsLimitFlag.Name)
}
