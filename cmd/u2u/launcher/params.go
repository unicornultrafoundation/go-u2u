package launcher

import (
	"github.com/unicornultrafoundation/go-helios/hash"

	"github.com/unicornultrafoundation/go-u2u/u2u"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesis"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesisstore"
)

var (
	Bootnodes = map[string][]string{
		"main": {
			"enode://21dfee41ddd127ebbd68fb14b39945f6e993ad9eb35c57e5e2e17ec1740960400d6d174f6c119fb9940072eec2d468ee5d767752bf9a44900ac8ac6d6de61330@18.143.208.170:5050",
			"enode://a1e1999ab32c7ea71b3fb4fd4e2143beadc3f71365e2a5a0e54e15780d28e5a80576a387406d9b60eee7c31289618c6a5ef93bfe295215518cecbf23bc50211e@3.1.11.147:5050",
		},
		"test": {},
	}

	mainnetHeader = genesis.Header{
		GenesisID:   hash.HexToHash("0x54e033c612a9b1a8ac8c6cb131f513202648f19b3a2756f8e2e40877d280606c"),
		NetworkID:   u2u.MainNetworkID,
		NetworkName: "main",
	}

	testnetHeader = genesis.Header{
		GenesisID:   hash.HexToHash("0xe633041cd774e07fce1910e99d16372af38875b16f8ce4d7131180c414ecd9a1"),
		NetworkID:   u2u.TestNetworkID,
		NetworkName: "testnet",
	}

	AllowedU2UGenesis = []GenesisTemplate{
		{
			Name:   "Mainnet",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x85307dd741356839d24176a1e015e90ceb9da53d9223d396a18524b9f8b24cb3"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x8847aff8a1934306902448a92c8a56e91ef843a550c61fa043a8e3881ef8a0ea"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x321932aa0bf71bc8ac9b26bfbdef111897c38120b7a1329f232d29ea9b26f6d3"),
			},
		},

		{
			Name:   "Mainnet-6321132-Full",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x7e0a3f77a734819b9d9c48b9c8c8756534b1f081e7eaabded85b4d2f4bf42bfa"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x580d1417f8be5e605c86354e490cc635b9a2e3b3d2fab37e9ae5cdba6401be41"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x8df818beac276736e2bebbc650514414da62e222ac23d629fbeb748a5dabcbc8"),
			},
		},

		{
			Name:   "Mainnet-6321132-MPT",
			Header: mainnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0x7e0a3f77a734819b9d9c48b9c8c8756534b1f081e7eaabded85b4d2f4bf42bfa"),
				genesisstore.BlocksSection(0): hash.HexToHash("0x580d1417f8be5e605c86354e490cc635b9a2e3b3d2fab37e9ae5cdba6401be41"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x33799f09da9aedd5afcadb630a76aaa729054bbd829efa4b1fd04dcff11f1cab"),
			},
		},

		{
			Name:   "Testnet",
			Header: testnetHeader,
			Hashes: genesis.Hashes{
				genesisstore.EpochsSection(0): hash.HexToHash("0xbe8c8541f429c14621766a2289a1a370db247f955b6c29e6896e80fddeedf26f"),
				genesisstore.BlocksSection(0): hash.HexToHash("0xd1cbc5a1ad98fbec03cb808ae69b707409e09d913c05fca4ee62a12bcd15e9d9"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x176dc5c014089ff165fb815ce57aeb652ad15e4d7b8a17c9c06ce2a48c1201ce"),
			},
		},
	}
)

func overrideParams() {
	// Below params are removed from source code.
	// Does not need to override.

	// params.MainnetBootnodes = []string{}
	// params.RopstenBootnodes = []string{}
	// params.RinkebyBootnodes = []string{}
	// params.GoerliBootnodes = []string{}
}
