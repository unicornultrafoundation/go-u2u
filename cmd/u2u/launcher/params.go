package launcher

import (
	"github.com/ethereum/go-ethereum/params"
	"github.com/unicornultrafoundation/go-hashgraph/hash"

	"github.com/unicornultrafoundation/go-u2u/u2u"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesis"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesisstore"
)

var (
	Bootnodes = map[string][]string{
		"main": {},
		"test": {},
	}

	mainnetHeader = genesis.Header{
		GenesisID:   hash.HexToHash("0x44e1da45bd5435ce8108c9fad8fee0f59a14513ec00693620eeb606fc9625005"),
		NetworkID:   u2u.MainNetworkID,
		NetworkName: "main",
	}

	// testnetHeader = genesis.Header{
	// 	GenesisID:   hash.HexToHash("0xe633041cd774e07fce1910e99d16372af38875b16f8ce4d7131180c414ecd9a1"),
	// 	NetworkID:   u2u.TestNetworkID,
	// 	NetworkName: "testnet",
	// }

	AllowedU2UGenesis = []GenesisTemplate{
		{
			Name:   "Mainnet",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x9b90f8cca338c3470a5a07b1f79d556c6ed13069dd1058ed2685aa33cfb16064"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x6174b693a6abada6acafa380000ed1b3c1d33220b83b39b60956efba883e88e0"),
				genesisstore.EvmSection(0):    hash.HexToHash("0xa13019f730ccd268d4c527cf3028f33a5120f16c72fe7ac10bc6f23529449188"),
			},
		},
	}
)

func overrideParams() {
	params.MainnetBootnodes = []string{}
	params.RopstenBootnodes = []string{}
	params.RinkebyBootnodes = []string{}
	params.GoerliBootnodes = []string{}
}
