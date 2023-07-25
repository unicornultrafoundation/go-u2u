package launcher

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/naoina/toml"
	"github.com/unicornultrafoundation/go-hashgraph/consensus"
	"github.com/unicornultrafoundation/go-hashgraph/utils/cachescale"
	"github.com/urfave/cli/v2"

	"github.com/unicornultrafoundation/go-u2u/evmcore"
	"github.com/unicornultrafoundation/go-u2u/gossip"
	"github.com/unicornultrafoundation/go-u2u/gossip/emitter"
	"github.com/unicornultrafoundation/go-u2u/gossip/gasprice"
	"github.com/unicornultrafoundation/go-u2u/integration"
	"github.com/unicornultrafoundation/go-u2u/integration/makefakegenesis"
	"github.com/unicornultrafoundation/go-u2u/internal/flags"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesis"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesisstore"
	futils "github.com/unicornultrafoundation/go-u2u/utils"
	"github.com/unicornultrafoundation/go-u2u/vecmt"
)

var (
	dumpConfigCommand = &cli.Command{
		Action:      dumpConfig,
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(nodeFlags, testFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}
	checkConfigCommand = &cli.Command{
		Action:      checkConfig,
		Name:        "checkconfig",
		Usage:       "Checks configuration file",
		ArgsUsage:   "",
		Flags:       append(nodeFlags, testFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The checkconfig checks configuration file.`,
	}

	configFileFlag = &cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}

	// DataDirFlag defines directory to store Hashgraph state and user's wallets
	DataDirFlag = &flags.DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
		Value: flags.DirectoryString(DefaultDataDir()),
	}

	CacheFlag = &cli.IntFlag{
		Name:  "cache",
		Usage: "Megabytes of memory allocated to internal caching",
		Value: DefaultCacheSize,
	}
	// GenesisFlag specifies network genesis configuration
	GenesisFlag = &cli.StringFlag{
		Name:  "genesis",
		Usage: "'path to genesis file' - sets the network genesis configuration.",
	}
	ExperimentalGenesisFlag = &cli.BoolFlag{
		Name:  "genesis.allowExperimental",
		Usage: "Allow to use experimental genesis file.",
	}

	RPCGlobalGasCapFlag = &cli.Uint64Flag{
		Name:  "rpc.gascap",
		Usage: "Sets a cap on gas that can be used in ftm_call/estimateGas (0=infinite)",
		Value: gossip.DefaultConfig(cachescale.Identity).RPCGasCap,
	}
	RPCGlobalTxFeeCapFlag = &cli.Float64Flag{
		Name:  "rpc.txfeecap",
		Usage: "Sets a cap on transaction fee (in FTM) that can be sent via the RPC APIs (0 = no cap)",
		Value: gossip.DefaultConfig(cachescale.Identity).RPCTxFeeCap,
	}
	RPCGlobalTimeoutFlag = &cli.DurationFlag{
		Name:  "rpc.timeout",
		Usage: "Time limit for RPC calls execution",
		Value: gossip.DefaultConfig(cachescale.Identity).RPCTimeout,
	}

	SyncModeFlag = &cli.StringFlag{
		Name:  "syncmode",
		Usage: `Blockchain sync mode ("full" or "snap")`,
		Value: "full",
	}

	GCModeFlag = &cli.StringFlag{
		Name:  "gcmode",
		Usage: `Blockchain garbage collection mode ("light", "full", "archive")`,
		Value: "archive",
	}

	ExitWhenAgeFlag = &cli.DurationFlag{
		Name:  "exitwhensynced.age",
		Usage: "Exits after synchronisation reaches the required age",
	}
	ExitWhenEpochFlag = &cli.Uint64Flag{
		Name:  "exitwhensynced.epoch",
		Usage: "Exits after synchronisation reaches the required epoch",
	}

	DBMigrationModeFlag = &cli.StringFlag{
		Name:  "db.migration.mode",
		Usage: "MultiDB migration mode ('reformat' or 'rebuild')",
	}
	DBPresetFlag = &cli.StringFlag{
		Name:  "db.preset",
		Usage: "DBs layout preset ('pbl-1' or 'ldb-1' or 'legacy-ldb' or 'legacy-pbl')",
	}
)

type GenesisTemplate struct {
	Name   string
	Header genesis.Header
	Hashes genesis.Hashes
}

const (
	// DefaultCacheSize is calculated as memory consumption in a worst case scenario with default configuration
	// Average memory consumption might be 3-5 times lower than the maximum
	DefaultCacheSize  = 3600
	ConstantCacheSize = 600
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		return fmt.Errorf("field '%s' is not defined in %s", field, rt.String())
	},
}

type config struct {
	Node           node.Config
	U2u            gossip.Config
	Emitter        emitter.Config
	TxPool         evmcore.TxPoolConfig
	U2uStore       gossip.StoreConfig
	Hashgraph      consensus.Config
	HashgraphStore consensus.StoreConfig
	VectorClock    vecmt.IndexConfig
	DBs            integration.DBsConfig
}

func (c *config) AppConfigs() integration.Configs {
	return integration.Configs{
		U2u:            c.U2u,
		U2uStore:       c.U2uStore,
		Hashgraph:      c.Hashgraph,
		HashgraphStore: c.HashgraphStore,
		VectorClock:    c.VectorClock,
		DBs:            c.DBs,
	}
}

func loadAllConfigs(file string, cfg *config) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	if err != nil {
		return errors.New(fmt.Sprintf("TOML config file error: %v.\n"+
			"Use 'dumpconfig' command to get an example config file.\n"+
			"If node was recently upgraded and a previous network config file is used, then check updates for the config file.", err))
	}
	return err
}

func mayGetGenesisStore(ctx *cli.Context) *genesisstore.Store {
	switch {
	case ctx.IsSet(FakeNetFlag.Name):
		_, num, err := parseFakeGen(ctx.String(FakeNetFlag.Name))
		if err != nil {
			log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
		}
		return makefakegenesis.FakeGenesisStore(num, futils.ToFtm(1000000000), futils.ToFtm(5000000))
	case ctx.IsSet(GenesisFlag.Name):
		genesisPath := ctx.String(GenesisFlag.Name)

		f, err := os.Open(genesisPath)
		if err != nil {
			utils.Fatalf("Failed to open genesis file: %v", err)
		}
		genesisStore, genesisHashes, err := genesisstore.OpenGenesisStore(f)
		if err != nil {
			utils.Fatalf("Failed to read genesis file: %v", err)
		}

		// check if it's a trusted preset
		{
			g := genesisStore.Genesis()
			gHeader := genesis.Header{
				GenesisID:   g.GenesisID,
				NetworkID:   g.NetworkID,
				NetworkName: g.NetworkName,
			}
			for _, allowed := range AllowedU2uGenesis {
				if allowed.Hashes.Equal(genesisHashes) && allowed.Header.Equal(gHeader) {
					log.Info("Genesis file is a known preset", "name", allowed.Name)
					goto notExperimental
				}
			}
			if ctx.Bool(ExperimentalGenesisFlag.Name) {
				log.Warn("Genesis file doesn't refer to any trusted preset")
			} else {
				utils.Fatalf("Genesis file doesn't refer to any trusted preset. Enable experimental genesis with --genesis.allowExperimental")
			}
		notExperimental:
		}
		return genesisStore
	}
	return nil
}

func setBootnodes(ctx *cli.Context, urls []string, cfg *node.Config) {
	cfg.P2P.BootstrapNodesV5 = []*enode.Node{}
	for _, url := range urls {
		if url != "" {
			node, err := enode.Parse(enode.ValidSchemes, url)
			if err != nil {
				log.Error("Bootstrap URL invalid", "enode", url, "err", err)
				continue
			}
			cfg.P2P.BootstrapNodesV5 = append(cfg.P2P.BootstrapNodesV5, node)
		}
	}
	cfg.P2P.BootstrapNodes = cfg.P2P.BootstrapNodesV5
}

func setDataDir(ctx *cli.Context, cfg *node.Config) {
	defaultDataDir := DefaultDataDir()

	switch {
	case ctx.IsSet(DataDirFlag.Name):
		cfg.DataDir = ctx.String(DataDirFlag.Name)
	case ctx.IsSet(FakeNetFlag.Name):
		_, num, err := parseFakeGen(ctx.String(FakeNetFlag.Name))
		if err != nil {
			log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
		}
		cfg.DataDir = filepath.Join(defaultDataDir, fmt.Sprintf("fakenet-%d", num))
	}
}

func setGPO(ctx *cli.Context, cfg *gasprice.Config) {}

func setTxPool(ctx *cli.Context, cfg *evmcore.TxPoolConfig) {
	if ctx.IsSet(utils.TxPoolLocalsFlag.Name) {
		locals := strings.Split(ctx.String(utils.TxPoolLocalsFlag.Name), ",")
		for _, account := range locals {
			if trimmed := strings.TrimSpace(account); !common.IsHexAddress(trimmed) {
				utils.Fatalf("Invalid account in --txpool.locals: %s", trimmed)
			} else {
				cfg.Locals = append(cfg.Locals, common.HexToAddress(account))
			}
		}
	}
	if ctx.IsSet(utils.TxPoolNoLocalsFlag.Name) {
		cfg.NoLocals = ctx.Bool(utils.TxPoolNoLocalsFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolJournalFlag.Name) {
		cfg.Journal = ctx.String(utils.TxPoolJournalFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolRejournalFlag.Name) {
		cfg.Rejournal = ctx.Duration(utils.TxPoolRejournalFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolPriceLimitFlag.Name) {
		cfg.PriceLimit = ctx.Uint64(utils.TxPoolPriceLimitFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolPriceBumpFlag.Name) {
		cfg.PriceBump = ctx.Uint64(utils.TxPoolPriceBumpFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolAccountSlotsFlag.Name) {
		cfg.AccountSlots = ctx.Uint64(utils.TxPoolAccountSlotsFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolGlobalSlotsFlag.Name) {
		cfg.GlobalSlots = ctx.Uint64(utils.TxPoolGlobalSlotsFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolAccountQueueFlag.Name) {
		cfg.AccountQueue = ctx.Uint64(utils.TxPoolAccountQueueFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolGlobalQueueFlag.Name) {
		cfg.GlobalQueue = ctx.Uint64(utils.TxPoolGlobalQueueFlag.Name)
	}
	if ctx.IsSet(utils.TxPoolLifetimeFlag.Name) {
		cfg.Lifetime = ctx.Duration(utils.TxPoolLifetimeFlag.Name)
	}
}

func gossipConfigWithFlags(ctx *cli.Context, src gossip.Config) (gossip.Config, error) {
	cfg := src

	setGPO(ctx, &cfg.GPO)

	if ctx.IsSet(RPCGlobalGasCapFlag.Name) {
		cfg.RPCGasCap = ctx.Uint64(RPCGlobalGasCapFlag.Name)
	}
	if ctx.IsSet(RPCGlobalTxFeeCapFlag.Name) {
		cfg.RPCTxFeeCap = ctx.Float64(RPCGlobalTxFeeCapFlag.Name)
	}
	if ctx.IsSet(RPCGlobalTimeoutFlag.Name) {
		cfg.RPCTimeout = ctx.Duration(RPCGlobalTimeoutFlag.Name)
	}
	if ctx.IsSet(SyncModeFlag.Name) {
		if syncmode := ctx.String(SyncModeFlag.Name); syncmode != "full" && syncmode != "snap" {
			utils.Fatalf("--%s must be either 'full' or 'snap'", SyncModeFlag.Name)
		}
		cfg.AllowSnapsync = ctx.String(SyncModeFlag.Name) == "snap"
	}

	return cfg, nil
}

func gossipStoreConfigWithFlags(ctx *cli.Context, src gossip.StoreConfig) (gossip.StoreConfig, error) {
	cfg := src
	if ctx.IsSet(utils.GCModeFlag.Name) {
		if gcmode := ctx.String(utils.GCModeFlag.Name); gcmode != "light" && gcmode != "full" && gcmode != "archive" {
			utils.Fatalf("--%s must be 'light', 'full' or 'archive'", GCModeFlag.Name)
		}
		cfg.EVM.Cache.TrieDirtyDisabled = ctx.String(utils.GCModeFlag.Name) == "archive"
		cfg.EVM.Cache.GreedyGC = ctx.String(utils.GCModeFlag.Name) == "full"
	}
	return cfg, nil
}

func setDBConfig(ctx *cli.Context, cfg integration.DBsConfig, cacheRatio cachescale.Func) integration.DBsConfig {
	if ctx.IsSet(DBPresetFlag.Name) {
		preset := ctx.String(DBPresetFlag.Name)
		switch preset {
		case "pbl-1":
			cfg = integration.Pbl1DBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles(ctx.Int(utils.FDLimitFlag.Name))))
		case "ldb-1":
			cfg = integration.Ldb1DBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles(ctx.Int(utils.FDLimitFlag.Name))))
		case "legacy-ldb":
			cfg = integration.LdbLegacyDBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles(ctx.Int(utils.FDLimitFlag.Name))))
		case "legacy-pbl":
			cfg = integration.PblLegacyDBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles(ctx.Int(utils.FDLimitFlag.Name))))
		default:
			utils.Fatalf("--%s must be 'pbl-1', 'ldb-1', 'legacy-pbl' or 'legacy-ldb'", DBPresetFlag.Name)
		}
	}
	if ctx.IsSet(DBMigrationModeFlag.Name) {
		cfg.MigrationMode = ctx.String(DBMigrationModeFlag.Name)
	}
	return cfg
}

func nodeConfigWithFlags(ctx *cli.Context, cfg node.Config) node.Config {
	utils.SetNodeConfig(ctx, &cfg)

	setDataDir(ctx, &cfg)
	return cfg
}

func cacheScaler(ctx *cli.Context) cachescale.Func {
	if !ctx.IsSet(CacheFlag.Name) {
		return cachescale.Identity
	}
	targetCache := ctx.Int(CacheFlag.Name)
	baseSize := DefaultCacheSize
	if targetCache < baseSize {
		log.Crit("Invalid flag", "flag", CacheFlag.Name, "err", fmt.Sprintf("minimum cache size is %d MB", baseSize))
	}
	return cachescale.Ratio{
		Base:   uint64(baseSize - ConstantCacheSize),
		Target: uint64(targetCache - ConstantCacheSize),
	}
}

func mayMakeAllConfigs(ctx *cli.Context) (*config, error) {
	// Defaults (low priority)
	cacheRatio := cacheScaler(ctx)
	cfg := config{
		Node:           defaultNodeConfig(),
		U2u:            gossip.DefaultConfig(cacheRatio),
		Emitter:        emitter.DefaultConfig(),
		TxPool:         evmcore.DefaultTxPoolConfig,
		U2uStore:       gossip.DefaultStoreConfig(cacheRatio),
		Hashgraph:      consensus.DefaultConfig(),
		HashgraphStore: consensus.DefaultStoreConfig(cacheRatio),
		VectorClock:    vecmt.DefaultConfig(cacheRatio),
	}

	if ctx.IsSet(FakeNetFlag.Name) {
		_, num, _ := parseFakeGen(ctx.String(FakeNetFlag.Name))
		cfg.Emitter = emitter.FakeConfig(num)
		setBootnodes(ctx, []string{}, &cfg.Node)
	} else {
		// "asDefault" means set network defaults
		cfg.Node.P2P.BootstrapNodes = asDefault
		cfg.Node.P2P.BootstrapNodesV5 = asDefault
	}

	// Load config file (medium priority)
	if file := ctx.String(configFileFlag.Name); file != "" {
		if err := loadAllConfigs(file, &cfg); err != nil {
			return &cfg, err
		}
	}
	// apply default for DB config if it wasn't touched by config file
	dbDefault := integration.DefaultDBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles(ctx.Int(utils.FDLimitFlag.Name))))
	if len(cfg.DBs.Routing.Table) == 0 {
		cfg.DBs.Routing = dbDefault.Routing
	}
	if len(cfg.DBs.GenesisCache.Table) == 0 {
		cfg.DBs.GenesisCache = dbDefault.GenesisCache
	}
	if len(cfg.DBs.RuntimeCache.Table) == 0 {
		cfg.DBs.RuntimeCache = dbDefault.RuntimeCache
	}

	// Apply flags (high priority)
	var err error
	cfg.U2u, err = gossipConfigWithFlags(ctx, cfg.U2u)
	if err != nil {
		return nil, err
	}
	cfg.U2uStore, err = gossipStoreConfigWithFlags(ctx, cfg.U2uStore)
	if err != nil {
		return nil, err
	}
	cfg.Node = nodeConfigWithFlags(ctx, cfg.Node)
	cfg.DBs = setDBConfig(ctx, cfg.DBs, cacheRatio)

	err = setValidator(ctx, &cfg.Emitter)
	if err != nil {
		return nil, err
	}
	if cfg.Emitter.Validator.ID != 0 && len(cfg.Emitter.PrevEmittedEventFile.Path) == 0 {
		cfg.Emitter.PrevEmittedEventFile.Path = cfg.Node.ResolvePath(path.Join("emitter", fmt.Sprintf("last-%d", cfg.Emitter.Validator.ID)))
	}
	setTxPool(ctx, &cfg.TxPool)

	if err := cfg.U2u.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func makeAllConfigs(ctx *cli.Context) *config {
	cfg, err := mayMakeAllConfigs(ctx)
	if err != nil {
		utils.Fatalf("%v", err)
	}
	return cfg
}

func defaultNodeConfig() node.Config {
	cfg := NodeDefaultConfig
	cfg.Name = clientIdentifier
	cfg.Version = params.VersionWithCommit(gitCommit, gitDate)
	cfg.HTTPModules = append(cfg.HTTPModules, "eth", "ftm", "dag", "abft", "web3")
	cfg.WSModules = append(cfg.WSModules, "eth", "ftm", "dag", "abft", "web3")
	cfg.IPCPath = "u2u.ipc"
	cfg.DataDir = DefaultDataDir()
	return cfg
}

// dumpConfig is the dumpconfig command.
func dumpConfig(ctx *cli.Context) error {
	cfg := makeAllConfigs(ctx)
	comment := ""

	out, err := tomlSettings.Marshal(&cfg)
	if err != nil {
		return err
	}

	dump := os.Stdout
	if ctx.NArg() > 0 {
		dump, err = os.OpenFile(ctx.Args().Get(0), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer dump.Close()
	}
	dump.WriteString(comment)
	dump.Write(out)

	return nil
}

func checkConfig(ctx *cli.Context) error {
	_, err := mayMakeAllConfigs(ctx)
	return err
}
