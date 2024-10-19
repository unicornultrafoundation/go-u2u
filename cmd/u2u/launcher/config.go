package launcher

import (
	"bufio"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/naoina/toml"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/unicornultrafoundation/go-helios/consensus"
	"github.com/unicornultrafoundation/go-helios/utils/cachescale"
	"github.com/unicornultrafoundation/go-u2u/cmd/utils"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/node"
	"github.com/unicornultrafoundation/go-u2u/p2p/enode"
	"github.com/unicornultrafoundation/go-u2u/params"
	"gopkg.in/urfave/cli.v1"

	"github.com/unicornultrafoundation/go-u2u/evmcore"
	"github.com/unicornultrafoundation/go-u2u/gossip"
	"github.com/unicornultrafoundation/go-u2u/gossip/emitter"
	"github.com/unicornultrafoundation/go-u2u/gossip/gasprice"
	"github.com/unicornultrafoundation/go-u2u/integration"
	"github.com/unicornultrafoundation/go-u2u/integration/makefakegenesis"
	"github.com/unicornultrafoundation/go-u2u/monitoring"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesis"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesisstore"
	futils "github.com/unicornultrafoundation/go-u2u/utils"
	"github.com/unicornultrafoundation/go-u2u/utils/memory"
	"github.com/unicornultrafoundation/go-u2u/vecmt"
)

var (
	dumpConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(dumpConfig),
		Name:        "dumpconfig",
		Usage:       "Show configuration values",
		ArgsUsage:   "",
		Flags:       append(nodeFlags, testFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The dumpconfig command shows configuration values.`,
	}
	checkConfigCommand = cli.Command{
		Action:      utils.MigrateFlags(checkConfig),
		Name:        "checkconfig",
		Usage:       "Checks configuration file",
		ArgsUsage:   "",
		Flags:       append(nodeFlags, testFlags...),
		Category:    "MISCELLANEOUS COMMANDS",
		Description: `The checkconfig checks configuration file.`,
	}

	configFileFlag = cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}

	// DataDirFlag defines directory to store state and user's wallets
	DataDirFlag = utils.DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
		Value: utils.DirectoryString(DefaultDataDir()),
	}

	CacheFlag = cli.IntFlag{
		Name:  "cache",
		Usage: "Megabytes of memory allocated to internal caching",
		Value: DefaultCacheSize,
	}
	// GenesisFlag specifies network genesis configuration
	GenesisFlag = cli.StringFlag{
		Name:  "genesis",
		Usage: "'path to genesis file' - sets the network genesis configuration.",
	}
	ExperimentalGenesisFlag = cli.BoolFlag{
		Name:  "genesis.allowExperimental",
		Usage: "Allow to use experimental genesis file.",
	}

	RPCGlobalGasCapFlag = cli.Uint64Flag{
		Name:  "rpc.gascap",
		Usage: "Sets a cap on gas that can be used in eth_call/estimateGas (0=infinite)",
		Value: gossip.DefaultConfig(cachescale.Identity).RPCGasCap,
	}
	RPCGlobalTxFeeCapFlag = cli.Float64Flag{
		Name:  "rpc.txfeecap",
		Usage: "Sets a cap on transaction fee (in U2U) that can be sent via the RPC APIs (0 = no cap)",
		Value: gossip.DefaultConfig(cachescale.Identity).RPCTxFeeCap,
	}
	RPCGlobalTimeoutFlag = cli.DurationFlag{
		Name:  "rpc.timeout",
		Usage: "Time limit for RPC calls execution",
		Value: gossip.DefaultConfig(cachescale.Identity).RPCTimeout,
	}

	SyncModeFlag = cli.StringFlag{
		Name:  "syncmode",
		Usage: `Blockchain sync mode ("full" or "snap")`,
		Value: "full",
	}

	GCModeFlag = cli.StringFlag{
		Name:  "gcmode",
		Usage: `Blockchain garbage collection mode ("light", "full", "archive")`,
		Value: "archive",
	}

	ExitWhenAgeFlag = cli.DurationFlag{
		Name:  "exitwhensynced.age",
		Usage: "Exits after synchronisation reaches the required age",
	}
	ExitWhenEpochFlag = cli.Uint64Flag{
		Name:  "exitwhensynced.epoch",
		Usage: "Exits after synchronisation reaches the required epoch",
	}

	// TxTracerFlag enables transaction tracing recording
	EnableTxTracerFlag = cli.BoolFlag{
		Name:  "enabletxtracer",
		Usage: "DO NOT RUN THIS OPTION AS VALIDATOR. Enable node records inner transaction traces for debugging purpose",
	}

	DBMigrationModeFlag = cli.StringFlag{
		Name:  "db.migration.mode",
		Usage: "MultiDB migration mode ('reformat' or 'rebuild')",
	}
	DBPresetFlag = cli.StringFlag{
		Name:  "db.preset",
		Usage: "DBs layout preset ('pebble' or 'legacy-pebble')",
	}

	// MonitoringFlag defines APIs endpoint to mornitor metrics
	EnableMonitorFlag = cli.BoolFlag{
		Name:  "monitor",
		Usage: "Enable the monitor servers",
	}
	PrometheusMonitoringPortFlag = cli.IntFlag{
		Name:  "monitor.prometheus.port",
		Usage: "Opens Prometheus API port to mornitor metrics",
		Value: monitoring.DefaultConfig.Port,
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
	ConstantCacheSize = 400
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
	U2U            gossip.Config
	Emitter        emitter.Config
	TxPool         evmcore.TxPoolConfig
	U2UStore       gossip.StoreConfig
	Hashgraph      consensus.Config
	HashgraphStore consensus.StoreConfig
	VectorClock    vecmt.IndexConfig
	DBs            integration.DBsConfig
	Monitoring     monitoring.Config
}

func (c *config) AppConfigs() integration.Configs {
	return integration.Configs{
		U2U:            c.U2U,
		U2UStore:       c.U2UStore,
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
	case ctx.GlobalIsSet(FakeNetFlag.Name):
		_, num, err := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
		if err != nil {
			log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
		}
		return makefakegenesis.FakeGenesisStore(num, futils.ToU2U(1000000000), futils.ToU2U(5000000))
	case ctx.GlobalIsSet(GenesisFlag.Name):
		genesisPath := ctx.GlobalString(GenesisFlag.Name)

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
			for _, allowed := range AllowedU2UGenesis {
				if allowed.Hashes.Equal(genesisHashes) && allowed.Header.Equal(gHeader) {
					log.Info("Genesis file is a known preset", "name", allowed.Name)
					goto notExperimental
				}
			}
			if ctx.GlobalBool(ExperimentalGenesisFlag.Name) {
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
	case ctx.GlobalIsSet(DataDirFlag.Name):
		cfg.DataDir = ctx.GlobalString(DataDirFlag.Name)
	case ctx.GlobalIsSet(FakeNetFlag.Name):
		_, num, err := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
		if err != nil {
			log.Crit("Invalid flag", "flag", FakeNetFlag.Name, "err", err)
		}
		cfg.DataDir = filepath.Join(defaultDataDir, fmt.Sprintf("fakenet-%d", num))
	}
}

func setGPO(ctx *cli.Context, cfg *gasprice.Config) {}

func setTxPool(ctx *cli.Context, cfg *evmcore.TxPoolConfig) {
	if ctx.GlobalIsSet(utils.TxPoolLocalsFlag.Name) {
		locals := strings.Split(ctx.GlobalString(utils.TxPoolLocalsFlag.Name), ",")
		for _, account := range locals {
			if trimmed := strings.TrimSpace(account); !common.IsHexAddress(trimmed) {
				utils.Fatalf("Invalid account in --txpool.locals: %s", trimmed)
			} else {
				cfg.Locals = append(cfg.Locals, common.HexToAddress(account))
			}
		}
	}
	if ctx.GlobalIsSet(utils.TxPoolNoLocalsFlag.Name) {
		cfg.NoLocals = ctx.GlobalBool(utils.TxPoolNoLocalsFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolJournalFlag.Name) {
		cfg.Journal = ctx.GlobalString(utils.TxPoolJournalFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolRejournalFlag.Name) {
		cfg.Rejournal = ctx.GlobalDuration(utils.TxPoolRejournalFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolPriceLimitFlag.Name) {
		cfg.PriceLimit = ctx.GlobalUint64(utils.TxPoolPriceLimitFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolPriceBumpFlag.Name) {
		cfg.PriceBump = ctx.GlobalUint64(utils.TxPoolPriceBumpFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolAccountSlotsFlag.Name) {
		cfg.AccountSlots = ctx.GlobalUint64(utils.TxPoolAccountSlotsFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolGlobalSlotsFlag.Name) {
		cfg.GlobalSlots = ctx.GlobalUint64(utils.TxPoolGlobalSlotsFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolAccountQueueFlag.Name) {
		cfg.AccountQueue = ctx.GlobalUint64(utils.TxPoolAccountQueueFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolGlobalQueueFlag.Name) {
		cfg.GlobalQueue = ctx.GlobalUint64(utils.TxPoolGlobalQueueFlag.Name)
	}
	if ctx.GlobalIsSet(utils.TxPoolLifetimeFlag.Name) {
		cfg.Lifetime = ctx.GlobalDuration(utils.TxPoolLifetimeFlag.Name)
	}
}

func gossipConfigWithFlags(ctx *cli.Context, src gossip.Config) (gossip.Config, error) {
	cfg := src

	setGPO(ctx, &cfg.GPO)

	if ctx.GlobalIsSet(RPCGlobalGasCapFlag.Name) {
		cfg.RPCGasCap = ctx.GlobalUint64(RPCGlobalGasCapFlag.Name)
	}
	if ctx.GlobalIsSet(RPCGlobalTxFeeCapFlag.Name) {
		cfg.RPCTxFeeCap = ctx.GlobalFloat64(RPCGlobalTxFeeCapFlag.Name)
	}
	if ctx.GlobalIsSet(RPCGlobalTimeoutFlag.Name) {
		cfg.RPCTimeout = ctx.GlobalDuration(RPCGlobalTimeoutFlag.Name)
	}
	if ctx.GlobalIsSet(SyncModeFlag.Name) {
		if syncmode := ctx.GlobalString(SyncModeFlag.Name); syncmode != "full" && syncmode != "snap" {
			utils.Fatalf("--%s must be either 'full' or 'snap'", SyncModeFlag.Name)
		}
		cfg.AllowSnapsync = ctx.GlobalString(SyncModeFlag.Name) == "snap"
	}
	if ctx.GlobalIsSet(utils.AllowUnprotectedTxs.Name) {
		cfg.AllowUnprotectedTxs = ctx.GlobalBool(utils.AllowUnprotectedTxs.Name)
	}

	return cfg, nil
}

func gossipStoreConfigWithFlags(ctx *cli.Context, src gossip.StoreConfig) (gossip.StoreConfig, error) {
	cfg := src
	if ctx.GlobalIsSet(utils.GCModeFlag.Name) {
		if gcmode := ctx.GlobalString(utils.GCModeFlag.Name); gcmode != "light" && gcmode != "full" && gcmode != "archive" {
			utils.Fatalf("--%s must be 'light', 'full' or 'archive'", GCModeFlag.Name)
		}
		cfg.EVM.Cache.TrieDirtyDisabled = ctx.GlobalString(utils.GCModeFlag.Name) == "archive"
		cfg.EVM.Cache.GreedyGC = ctx.GlobalString(utils.GCModeFlag.Name) == "full"
	}
	return cfg, nil
}

func setDBConfig(ctx *cli.Context, cfg integration.DBsConfig, cacheRatio cachescale.Func) integration.DBsConfig {
	if ctx.GlobalIsSet(DBPresetFlag.Name) {
		preset := ctx.GlobalString(DBPresetFlag.Name)
		cfg = setDBConfigStr(cfg, cacheRatio, preset)
	}
	if ctx.GlobalIsSet(DBMigrationModeFlag.Name) {
		cfg.MigrationMode = ctx.GlobalString(DBMigrationModeFlag.Name)
	}
	return cfg
}

func setDBConfigStr(cfg integration.DBsConfig, cacheRatio cachescale.Func, preset string) integration.DBsConfig {
	switch preset {
	case "pebble":
		cfg = integration.Pbl1DBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles()))
	case "legacy-pebble":
		cfg = integration.PblLegacyDBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles()))
	default:
		utils.Fatalf("--%s must be 'pebble' or 'legacy-pebble'", DBPresetFlag.Name)
	}
	// sanity check
	if preset != reversePresetName(cfg.Routing) {
		log.Error("Preset name cannot be reversed")
	}
	return cfg
}

func reversePresetName(cfg integration.RoutingConfig) string {
	pbl1 := integration.Pbl1RoutingConfig()
	pblLegacy := integration.PblLegacyRoutingConfig()
	if cfg.Equal(pbl1) {
		return "pebble"
	}
	if cfg.Equal(pblLegacy) {
		return "legacy-pebble"
	}
	return ""
}

func memorizeDBPreset(cfg *config) {
	preset := reversePresetName(cfg.DBs.Routing)
	pPath := path.Join(cfg.Node.DataDir, "chaindata", "preset")
	if len(preset) != 0 {
		futils.FilePut(pPath, []byte(preset), true)
	} else {
		_ = os.Remove(pPath)
	}
}

func setDBConfigDefault(cfg config, cacheRatio cachescale.Func) config {
	if len(cfg.DBs.Routing.Table) == 0 && len(cfg.DBs.GenesisCache.Table) == 0 && len(cfg.DBs.RuntimeCache.Table) == 0 {
		// Substitute memorized db preset from datadir, unless already set
		datadirPreset := futils.FileGet(path.Join(cfg.Node.DataDir, "chaindata", "preset"))
		if len(datadirPreset) != 0 {
			cfg.DBs = setDBConfigStr(cfg.DBs, cacheRatio, string(datadirPreset))
		}
	}
	// apply default for DB config if it wasn't touched by config file or flags, and there's no datadir's default value
	dbDefault := integration.DefaultDBsConfig(cacheRatio.U64, uint64(utils.MakeDatabaseHandles()))
	if len(cfg.DBs.Routing.Table) == 0 {
		cfg.DBs.Routing = dbDefault.Routing
	}
	if len(cfg.DBs.GenesisCache.Table) == 0 {
		cfg.DBs.GenesisCache = dbDefault.GenesisCache
	}
	if len(cfg.DBs.RuntimeCache.Table) == 0 {
		cfg.DBs.RuntimeCache = dbDefault.RuntimeCache
	}
	return cfg
}

func setMonitoringConfig(ctx *cli.Context, cfg monitoring.Config) monitoring.Config {
	// apply config for monitoring
	cfg.Port = ctx.GlobalInt(PrometheusMonitoringPortFlag.Name)

	return cfg
}

func nodeConfigWithFlags(ctx *cli.Context, cfg node.Config) node.Config {
	utils.SetNodeConfig(ctx, &cfg)

	setDataDir(ctx, &cfg)
	return cfg
}

func cacheScaler(ctx *cli.Context) cachescale.Func {
	targetCache := ctx.GlobalInt(CacheFlag.Name)
	baseSize := DefaultCacheSize
	totalMemory := int(memory.TotalMemory() / opt.MiB)
	maxCache := totalMemory * 3 / 5
	if maxCache < baseSize {
		maxCache = baseSize
	}
	if !ctx.GlobalIsSet(CacheFlag.Name) {
		recommendedCache := totalMemory / 2
		if recommendedCache > baseSize {
			log.Warn(fmt.Sprintf("Please add '--%s %d' flag to allocate more cache for U2U Client. Total memory is %d MB.", CacheFlag.Name, recommendedCache, totalMemory))
		}
		return cachescale.Identity
	}
	if targetCache < baseSize {
		log.Crit("Invalid flag", "flag", CacheFlag.Name, "err", fmt.Sprintf("minimum cache size is %d MB", baseSize))
	}
	if totalMemory != 0 && targetCache > maxCache {
		log.Warn(fmt.Sprintf("Requested cache size exceeds 60%% of available memory. Reducing cache size to %d MB.", maxCache))
		targetCache = maxCache
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
		U2U:            gossip.DefaultConfig(cacheRatio),
		Emitter:        emitter.DefaultConfig(),
		TxPool:         evmcore.DefaultTxPoolConfig,
		U2UStore:       gossip.DefaultStoreConfig(cacheRatio),
		Hashgraph:      consensus.DefaultConfig(),
		HashgraphStore: consensus.DefaultStoreConfig(cacheRatio),
		VectorClock:    vecmt.DefaultConfig(cacheRatio),
	}

	cfg.U2UStore.SFC.Enable = ctx.GlobalBool(utils.ConsensusFlag.Name)

	if ctx.GlobalIsSet(FakeNetFlag.Name) {
		_, num, _ := parseFakeGen(ctx.GlobalString(FakeNetFlag.Name))
		cfg.Emitter = emitter.FakeConfig(num)
		setBootnodes(ctx, []string{}, &cfg.Node)
	} else {
		// "asDefault" means set network defaults
		cfg.Node.P2P.BootstrapNodes = asDefault
		cfg.Node.P2P.BootstrapNodesV5 = asDefault
	}

	// Load config file (medium priority)
	if file := ctx.GlobalString(configFileFlag.Name); file != "" {
		if err := loadAllConfigs(file, &cfg); err != nil {
			return &cfg, err
		}
	}

	// Apply flags (high priority)
	var err error
	cfg.U2U, err = gossipConfigWithFlags(ctx, cfg.U2U)
	if err != nil {
		return nil, err
	}
	cfg.U2UStore, err = gossipStoreConfigWithFlags(ctx, cfg.U2UStore)
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

	// Process DBs defaults in the end because they are applied only in absence of config or flags
	cfg = setDBConfigDefault(cfg, cacheRatio)
	// Sanitize GPO config
	if cfg.U2U.GPO.MinGasTip == nil || cfg.U2U.GPO.MinGasTip.Sign() == 0 {
		cfg.U2U.GPO.MinGasTip = new(big.Int).SetUint64(cfg.TxPool.PriceLimit)
	}
	if cfg.U2U.GPO.MinGasTip.Cmp(new(big.Int).SetUint64(cfg.TxPool.PriceLimit)) < 0 {
		log.Warn(fmt.Sprintf("GPO minimum gas tip (U2U.GPO.MinGasTip=%s) is lower than txpool minimum gas tip (TxPool.PriceLimit=%d)", cfg.U2U.GPO.MinGasTip.String(), cfg.TxPool.PriceLimit))
	}

	if err := cfg.U2U.Validate(); err != nil {
		return nil, err
	}

	if ctx.GlobalIsSet(EnableTxTracerFlag.Name) {
		cfg.U2UStore.TraceTransactions = true
	}

	if ctx.GlobalIsSet(EnableMonitorFlag.Name) {
		cfg.Monitoring = setMonitoringConfig(ctx, cfg.Monitoring)
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
	cfg.HTTPModules = append(cfg.HTTPModules, "eth", "dag", "abft", "web3")
	cfg.WSModules = append(cfg.WSModules, "eth", "dag", "abft", "web3")
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
