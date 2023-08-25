package benchtest

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/urfave/cli.v1"

	"github.com/unicornultrafoundation/go-u2u/u2u"
	"github.com/unicornultrafoundation/go-u2u/utils/toml"
)

type Config struct {
	ChainId int64 // chain id for sign transactions
	Payer   common.Address
	URLs    []string // WS nodes API URL
}

func DefaultConfig() *Config {
	return &Config{
		ChainId: int64(u2u.FakeNetworkID),
		URLs: []string{
			"ws://127.0.0.1:4501",
			"ws://127.0.0.1:4502",
		},
	}
}

func OpenConfig(ctx *cli.Context) *Config {
	cfg := DefaultConfig()
	f := ctx.GlobalString(NetworkConfigFileFlag.Name)
	err := cfg.Load(f)
	if err != nil {
		panic(err)
	}
	return cfg
}

func (cfg *Config) Load(file string) error {
	data, err := toml.ParseFile(file)
	if err != nil {
		return err
	}

	err = toml.Settings.UnmarshalTable(data, cfg)
	if err != nil {
		err = errors.New(file + ", " + err.Error())
		return err
	}

	return nil
}
