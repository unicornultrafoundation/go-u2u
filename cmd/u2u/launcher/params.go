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
		"main": {
			"enode://104a461922f696758c73f69819a9c7990acb8ed692790d228602e89f45375f04d57fbaeb00f50b7ccb677551f4d91fc04ea67a4a1ac49099996f7c1d38b502f6@18.139.172.157:5050",
			"enode://6fa78c636f53ae45b0b3bf44b41f39ebcb62f0bbf52c314c496443fe80ba4ae2a1e2877dcb16a0aca3209fe20d55702909e88174f47bfd82b79ea65ae9d7e076@18.138.204.206:5050",
		},
		"test": {},
	}

	mainnetHeader = genesis.Header{
		GenesisID:   hash.HexToHash("0x44e1da45bd5435ce8108c9fad8fee0f59a14513ec00693620eeb606fc9625005"),
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
				genesisstore.EpochsSection(0): hash.HexToHash("0x0b66da086607d273e90d3a0bde28737f20a3301d6c73b60d039f551c5734d267"),
				genesisstore.BlocksSection(0): hash.HexToHash("0xac64d5db3991eeaebe605987f466662658de984e897a01d533d20f468d5f93a1"),
				genesisstore.EvmSection(0):    hash.HexToHash("0x0c2943da59dffcf3b00a1e3a76a603f475087c140b72353cb9a398c48350dcd2"),
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
	params.MainnetBootnodes = []string{}
	params.RopstenBootnodes = []string{}
	params.RinkebyBootnodes = []string{}
	params.GoerliBootnodes = []string{}
}
