package integration

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/unicornultrafoundation/go-hashgraph/u2udb/multidb"
)

var DefaultDBsConfig = PblLegacyDBsConfig
var DefaultDBsConfigName = "pbl-1"

/*
 * pbl-1 config
 */

func Pbl1DBsConfig(scale func(uint64) uint64, fdlimit uint64) DBsConfig {
	return DBsConfig{
		Routing:      Pbl1RoutingConfig(),
		RuntimeCache: Pbl1RuntimeDBsCacheConfig(scale, fdlimit),
		GenesisCache: Pbl1GenesisDBsCacheConfig(scale, fdlimit),
	}
}

func Pbl1RoutingConfig() RoutingConfig {
	return RoutingConfig{
		Table: map[string]multidb.Route{
			"": {
				Type: "pebble-fsh",
			},
			"hashgraph": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "C",
			},
			"gossip": {
				Type: "pebble-fsh",
				Name: "main",
			},
			"evm": {
				Type: "pebble-fsh",
				Name: "main",
			},
			"gossip/e": {
				Type: "pebble-fsh",
				Name: "events",
			},
			"evm/M": {
				Type: "pebble-drc",
				Name: "evm-data",
			},
			"evm-logs": {
				Type: "pebble-fsh",
				Name: "evm-logs",
			},
			"gossip-%d": {
				Type:  "leveldb-fsh",
				Name:  "epoch-%d",
				Table: "G",
			},
			"hashgraph-%d": {
				Type:   "leveldb-fsh",
				Name:   "epoch-%d",
				Table:  "L",
				NoDrop: true,
			},
		},
	}
}

func Pbl1RuntimeDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"evm-data": {
				Cache:   scale(480 * opt.MiB),
				Fdlimit: fdlimit*480/1400 + 1,
			},
			"evm-logs": {
				Cache:   scale(260 * opt.MiB),
				Fdlimit: fdlimit*260/1400 + 1,
			},
			"main": {
				Cache:   scale(320 * opt.MiB),
				Fdlimit: fdlimit*320/1400 + 1,
			},
			"events": {
				Cache:   scale(240 * opt.MiB),
				Fdlimit: fdlimit*240/1400 + 1,
			},
			"epoch-%d": {
				Cache:   scale(100 * opt.MiB),
				Fdlimit: fdlimit*100/1400 + 1,
			},
			"": {
				Cache:   64 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}

func Pbl1GenesisDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"main": {
				Cache:   scale(1000 * opt.MiB),
				Fdlimit: fdlimit*1000/3000 + 1,
			},
			"evm-data": {
				Cache:   scale(1000 * opt.MiB),
				Fdlimit: fdlimit*1000/3000 + 1,
			},
			"evm-logs": {
				Cache:   scale(1000 * opt.MiB),
				Fdlimit: fdlimit*1000/3000 + 1,
			},
			"events": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3000 + 1,
			},
			"epoch-%d": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3000 + 1,
			},
			"": {
				Cache:   16 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}

/*
 * legacy-pbl config
 */

func PblLegacyDBsConfig(scale func(uint64) uint64, fdlimit uint64) DBsConfig {
	return DBsConfig{
		Routing:      PblLegacyRoutingConfig(),
		RuntimeCache: PblLegacyRuntimeDBsCacheConfig(scale, fdlimit),
		GenesisCache: PblLegacyGenesisDBsCacheConfig(scale, fdlimit),
	}
}

func PblLegacyRoutingConfig() RoutingConfig {
	return RoutingConfig{
		Table: map[string]multidb.Route{
			"": {
				Type: "pebble-fsh",
			},
			"hashgraph": {
				Type: "pebble-fsh",
				Name: "hashgraph",
			},
			"gossip": {
				Type: "pebble-fsh",
				Name: "main",
			},

			"gossip/S": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "!",
			},
			"gossip/R": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "@",
			},
			"gossip/Q": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "#",
			},

			"gossip/T": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "$",
			},
			"gossip/J": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "%",
			},
			"gossip/E": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "^",
			},

			"gossip/I": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "&",
			},
			"gossip/G": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "*",
			},
			"gossip/F": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "(",
			},

			"evm": {
				Type: "pebble-fsh",
				Name: "main",
			},
			"evm-logs": {
				Type:  "pebble-fsh",
				Name:  "main",
				Table: "L",
			},
			"gossip-%d": {
				Type: "pebble-fsh",
				Name: "gossip-%d",
			},
			"hashgraph-%d": {
				Type: "pebble-fsh",
				Name: "hashgraph-%d",
			},
		},
	}
}

func PblLegacyRuntimeDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"main": {
				Cache:   scale(950 * opt.MiB),
				Fdlimit: fdlimit*950/1400 + 1,
			},
			"hashgraph": {
				Cache:   scale(150 * opt.MiB),
				Fdlimit: fdlimit*150/1400 + 1,
			},
			"gossip-%d": {
				Cache:   scale(150 * opt.MiB),
				Fdlimit: fdlimit*150/1400 + 1,
			},
			"hashgraph-%d": {
				Cache:   scale(150 * opt.MiB),
				Fdlimit: fdlimit*150/1400 + 1,
			},
			"": {
				Cache:   64 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}

func PblLegacyGenesisDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"main": {
				Cache:   scale(3000 * opt.MiB),
				Fdlimit: fdlimit,
			},
			"hashgraph": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3000 + 1,
			},
			"gossip-%d": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3000 + 1,
			},
			"hashgraph-%d": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3000 + 1,
			},
			"": {
				Cache:   16 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}
