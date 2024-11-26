package launcher

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/unicornultrafoundation/go-helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/accounts"
	"github.com/unicornultrafoundation/go-u2u/accounts/keystore"
	"github.com/unicornultrafoundation/go-u2u/cmd/u2u/launcher/monitoring"
	"github.com/unicornultrafoundation/go-u2u/cmd/u2u/launcher/tracing"
	"github.com/unicornultrafoundation/go-u2u/cmd/utils"
	"github.com/unicornultrafoundation/go-u2u/console/prompt"
	"github.com/unicornultrafoundation/go-u2u/debug"
	"github.com/unicornultrafoundation/go-u2u/ethclient"
	"github.com/unicornultrafoundation/go-u2u/evmcore"
	"github.com/unicornultrafoundation/go-u2u/flags"
	"github.com/unicornultrafoundation/go-u2u/gossip"
	"github.com/unicornultrafoundation/go-u2u/gossip/emitter"
	"github.com/unicornultrafoundation/go-u2u/integration"
	"github.com/unicornultrafoundation/go-u2u/log"
	evmetrics "github.com/unicornultrafoundation/go-u2u/metrics"
	"github.com/unicornultrafoundation/go-u2u/node"
	"github.com/unicornultrafoundation/go-u2u/p2p/discover/discfilter"
	"github.com/unicornultrafoundation/go-u2u/params"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesis"
	"github.com/unicornultrafoundation/go-u2u/u2u/genesisstore"
	"github.com/unicornultrafoundation/go-u2u/utils/errlock"
	"github.com/unicornultrafoundation/go-u2u/valkeystore"
	_ "github.com/unicornultrafoundation/go-u2u/version"
)

const (
	// clientIdentifier to advertise over the network.
	clientIdentifier = "go-u2u"
)

var (
	// Git SHA1 commit hash of the release (set via linker flags).
	gitCommit = ""
	gitDate   = ""
	// The app that holds all commands and flags.
	app = flags.NewApp(gitCommit, gitDate, "the go-u2u command line interface")

	nodeFlags        []cli.Flag
	testFlags        []cli.Flag
	gpoFlags         []cli.Flag
	accountFlags     []cli.Flag
	performanceFlags []cli.Flag
	networkingFlags  []cli.Flag
	txpoolFlags      []cli.Flag
	u2uFlags         []cli.Flag
	rpcFlags         []cli.Flag
	metricsFlags     []cli.Flag
)

func initFlags() {
	// Flags for testing purpose.
	testFlags = []cli.Flag{
		FakeNetFlag,
	}

	// Flags that configure the node.
	gpoFlags = []cli.Flag{}
	accountFlags = []cli.Flag{
		utils.UnlockedAccountFlag,
		utils.PasswordFileFlag,
		utils.ExternalSignerFlag,
		utils.InsecureUnlockAllowedFlag,
	}
	performanceFlags = []cli.Flag{
		CacheFlag,
	}
	networkingFlags = []cli.Flag{
		utils.BootnodesFlag,
		utils.ListenPortFlag,
		utils.MaxPeersFlag,
		utils.MaxPendingPeersFlag,
		utils.NATFlag,
		utils.NoDiscoverFlag,
		utils.DiscoveryV5Flag,
		utils.NetrestrictFlag,
		utils.IPrestrictFlag,
		utils.PrivateNodeFlag,
		utils.NodeKeyFileFlag,
		utils.NodeKeyHexFlag,
	}
	txpoolFlags = []cli.Flag{
		utils.TxPoolLocalsFlag,
		utils.TxPoolNoLocalsFlag,
		utils.TxPoolJournalFlag,
		utils.TxPoolRejournalFlag,
		utils.TxPoolPriceLimitFlag,
		utils.TxPoolPriceBumpFlag,
		utils.TxPoolAccountSlotsFlag,
		utils.TxPoolGlobalSlotsFlag,
		utils.TxPoolAccountQueueFlag,
		utils.TxPoolGlobalQueueFlag,
		utils.TxPoolLifetimeFlag,
	}
	u2uFlags = []cli.Flag{
		GenesisFlag,
		ExperimentalGenesisFlag,
		utils.IdentityFlag,
		DataDirFlag,
		utils.MinFreeDiskSpaceFlag,
		utils.KeyStoreDirFlag,
		utils.USBFlag,
		ExitWhenAgeFlag,
		ExitWhenEpochFlag,
		utils.LightKDFFlag,
		configFileFlag,
		validatorIDFlag,
		validatorPubkeyFlag,
		validatorPasswordFlag,
		SyncModeFlag,
		GCModeFlag,
		DBPresetFlag,
		DBMigrationModeFlag,
		EnableTxTracerFlag,
		EnableMonitorFlag,
		PrometheusMonitoringPortFlag,
	}

	rpcFlags = []cli.Flag{
		utils.HTTPEnabledFlag,
		utils.HTTPListenAddrFlag,
		utils.HTTPPortFlag,
		utils.HTTPCORSDomainFlag,
		utils.HTTPVirtualHostsFlag,
		utils.HTTPApiFlag,
		utils.HTTPPathPrefixFlag,
		utils.WSEnabledFlag,
		utils.WSListenAddrFlag,
		utils.WSPortFlag,
		utils.WSApiFlag,
		utils.WSAllowedOriginsFlag,
		utils.WSPathPrefixFlag,
		utils.IPCDisabledFlag,
		utils.IPCPathFlag,
		utils.AllowUnprotectedTxs,
		RPCGlobalGasCapFlag,
		RPCGlobalTxFeeCapFlag,
		RPCGlobalTimeoutFlag,
	}

	metricsFlags = []cli.Flag{
		utils.MetricsEnabledFlag,
		utils.MetricsEnabledExpensiveFlag,
		utils.MetricsHTTPFlag,
		utils.MetricsPortFlag,
		utils.MetricsEnableInfluxDBFlag,
		utils.MetricsInfluxDBEndpointFlag,
		utils.MetricsInfluxDBDatabaseFlag,
		utils.MetricsInfluxDBUsernameFlag,
		utils.MetricsInfluxDBPasswordFlag,
		utils.MetricsInfluxDBTagsFlag,
		utils.MetricsEnableInfluxDBV2Flag,
		utils.MetricsInfluxDBTokenFlag,
		utils.MetricsInfluxDBBucketFlag,
		utils.MetricsInfluxDBOrganizationFlag,
		tracing.EnableFlag,
	}

	nodeFlags = []cli.Flag{}
	nodeFlags = append(nodeFlags, gpoFlags...)
	nodeFlags = append(nodeFlags, accountFlags...)
	nodeFlags = append(nodeFlags, performanceFlags...)
	nodeFlags = append(nodeFlags, networkingFlags...)
	nodeFlags = append(nodeFlags, txpoolFlags...)
	nodeFlags = append(nodeFlags, u2uFlags...)
}

// init the CLI app.
func init() {
	discfilter.Enable()
	overrideFlags()
	overrideParams()

	initFlags()

	// App.

	app.Action = heliosMain
	app.Version = params.VersionWithCommit(gitCommit, gitDate)
	app.HideVersion = true // we have a command to print the version
	app.Commands = []cli.Command{
		// See accountcmd.go:
		accountCommand,
		walletCommand,
		// see validatorcmd.go:
		validatorCommand,
		// See consolecmd.go:
		consoleCommand,
		attachCommand,
		javascriptCommand,
		// See config.go:
		dumpConfigCommand,
		checkConfigCommand,
		// See misccmd.go:
		versionCommand,
		licenseCommand,
		// See chaincmd.go
		importCommand,
		exportCommand,
		checkCommand,
		// See snapshot.go
		snapshotCommand,
		// See dbcmd.go
		dbCommand,
		deleteCommand,
	}
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Flags = append(app.Flags, testFlags...)
	app.Flags = append(app.Flags, nodeFlags...)
	app.Flags = append(app.Flags, rpcFlags...)
	app.Flags = append(app.Flags, consoleFlags...)
	app.Flags = append(app.Flags, debug.Flags...)
	app.Flags = append(app.Flags, metricsFlags...)

	app.Before = func(ctx *cli.Context) error {
		if err := debug.Setup(ctx); err != nil {
			return err
		}

		// Start metrics export if enabled
		utils.SetupMetrics(ctx)
		// Start system runtime metrics collection
		go evmetrics.CollectProcessMetrics(3 * time.Second)
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		debug.Exit()
		prompt.Stdin.Close() // Resets terminal mode.

		return nil
	}
}

func Launch(args []string) error {
	return app.Run(args)
}

// u2u is the main entry point into the system if no special subcommand is ran.
// It creates a default node based on the command line arguments and runs it in
// blocking mode, waiting for it to be shut down.
func heliosMain(ctx *cli.Context) error {
	if args := ctx.Args(); len(args) > 0 {
		return fmt.Errorf("invalid command: %q", args[0])
	}

	// TODO: tracing flags
	//tracingStop, err := tracing.Start(ctx)
	//if err != nil {
	//	return err
	//}
	//defer tracingStop()

	cfg := makeAllConfigs(ctx)
	genesisStore := mayGetGenesisStore(ctx)
	node, _, nodeClose := makeNode(ctx, cfg, genesisStore)
	defer nodeClose()
	startNode(ctx, node)
	node.Wait()
	return nil
}

func makeNode(ctx *cli.Context, cfg *config, genesisStore *genesisstore.Store) (*node.Node, *gossip.Service, func()) {
	// check errlock file
	errlock.SetDefaultDatadir(cfg.Node.DataDir)
	errlock.Check()

	var g *genesis.Genesis
	if genesisStore != nil {
		gv := genesisStore.Genesis()
		g = &gv
	}

	engine, dagIndex, gdb, cdb, blockProc, closeDBs := integration.MakeEngine(path.Join(cfg.Node.DataDir, "chaindata"), g, cfg.AppConfigs())
	if genesisStore != nil {
		_ = genesisStore.Close()
	}

	monitoring.SetupPrometheus(fmt.Sprintf(":%d", cfg.Monitoring.Port))
	monitoring.SetDataDir(cfg.Node.DataDir)

	memorizeDBPreset(cfg)

	// substitute default bootnodes if requested
	networkName := ""
	if gdb.HasBlockEpochState() {
		networkName = gdb.GetRules().Name
	}
	if len(networkName) == 0 && genesisStore != nil {
		networkName = genesisStore.Header().NetworkName
	}
	if needDefaultBootnodes(cfg.Node.P2P.BootstrapNodes) {
		bootnodes := Bootnodes[networkName]
		if bootnodes == nil {
			bootnodes = []string{}
		}
		setBootnodes(ctx, bootnodes, &cfg.Node)
	}

	stack := makeConfigNode(ctx, &cfg.Node)

	valKeystore := valkeystore.NewDefaultFileKeystore(path.Join(getValKeystoreDir(cfg.Node), "validator"))
	valPubkey := cfg.Emitter.Validator.PubKey
	if key := getFakeValidatorKey(ctx); key != nil && cfg.Emitter.Validator.ID != 0 {
		addFakeValidatorKey(ctx, key, valPubkey, valKeystore)
		coinbase := integration.SetAccountKey(stack.AccountManager(), key, "fakepassword")
		log.Info("Unlocked fake validator account", "address", coinbase.Address.Hex())
	}

	// unlock validator key
	if !valPubkey.Empty() {
		err := unlockValidatorKey(ctx, valPubkey, valKeystore)
		if err != nil {
			utils.Fatalf("Failed to unlock validator key: %v", err)
		}
	}
	signer := valkeystore.NewSigner(valKeystore)

	// Create and register a gossip network service.
	newTxPool := func(reader evmcore.StateReader) gossip.TxPool {
		if cfg.TxPool.Journal != "" {
			cfg.TxPool.Journal = stack.ResolvePath(cfg.TxPool.Journal)
		}
		return evmcore.NewTxPool(cfg.TxPool, reader.Config(), reader)
	}
	haltCheck := func(oldEpoch, newEpoch idx.Epoch, age time.Time) bool {
		stop := ctx.GlobalIsSet(ExitWhenAgeFlag.Name) && ctx.GlobalDuration(ExitWhenAgeFlag.Name) >= time.Since(age)
		stop = stop || ctx.GlobalIsSet(ExitWhenEpochFlag.Name) && idx.Epoch(ctx.GlobalUint64(ExitWhenEpochFlag.Name)) <= newEpoch
		if stop {
			go func() {
				// do it in a separate thread to avoid deadlock
				_ = stack.Close()
			}()
			return true
		}
		return false
	}
	svc, err := gossip.NewService(stack, cfg.U2U, gdb, blockProc, engine, dagIndex, newTxPool, haltCheck)
	if err != nil {
		utils.Fatalf("Failed to create the service: %v", err)
	}
	err = engine.StartFrom(svc.GetConsensusCallbacks(), gdb.GetEpoch(), gdb.GetValidators())
	if err != nil {
		utils.Fatalf("Failed to bootstrap the engine: %v", err)
	}
	svc.ReprocessEpochEvents()
	if cfg.Emitter.Validator.ID != 0 {
		svc.RegisterEmitter(emitter.NewEmitter(cfg.Emitter, svc.EmitterWorld(signer)))
	}

	stack.RegisterAPIs(svc.APIs())
	stack.RegisterProtocols(svc.Protocols())
	stack.RegisterLifecycle(svc)

	return stack, svc, func() {
		_ = stack.Close()
		gdb.Close()
		_ = cdb.Close()
		if closeDBs != nil {
			_ = closeDBs()
		}
	}
}

func makeConfigNode(ctx *cli.Context, cfg *node.Config) *node.Node {
	stack, err := node.New(cfg)
	if err != nil {
		utils.Fatalf("Failed to create the protocol stack: %v", err)
	}

	return stack
}

// startNode boots up the system node and all registered protocols, after which
// it unlocks any requested accounts, and starts the RPC/IPC interfaces.
func startNode(ctx *cli.Context, stack *node.Node) {
	// Start up the node itself
	utils.StartNode(ctx, stack)

	// Unlock any account specifically requested
	unlockAccounts(ctx, stack)

	// Register wallet event handlers to open and auto-derive wallets
	events := make(chan accounts.WalletEvent, 16)
	stack.AccountManager().Subscribe(events)

	// Create a client to interact with local u2u node.
	rpcClient, err := stack.Attach()
	if err != nil {
		utils.Fatalf("Failed to attach to self: %v", err)
	}
	ethClient := ethclient.NewClient(rpcClient)
	go func() {
		// Open any wallets already attached
		for _, wallet := range stack.AccountManager().Wallets() {
			if err := wallet.Open(""); err != nil {
				log.Warn("Failed to open wallet", "url", wallet.URL(), "err", err)
			}
		}
		// Listen for wallet event till termination
		for event := range events {
			switch event.Kind {
			case accounts.WalletArrived:
				if err := event.Wallet.Open(""); err != nil {
					log.Warn("New wallet appeared, failed to open", "url", event.Wallet.URL(), "err", err)
				}
			case accounts.WalletOpened:
				status, _ := event.Wallet.Status()
				log.Info("New wallet appeared", "url", event.Wallet.URL(), "status", status)

				var derivationPaths []accounts.DerivationPath
				if event.Wallet.URL().Scheme == "ledger" {
					derivationPaths = append(derivationPaths, accounts.LegacyLedgerBaseDerivationPath)
				}
				derivationPaths = append(derivationPaths, accounts.DefaultBaseDerivationPath)

				event.Wallet.SelfDerive(derivationPaths, ethClient)

			case accounts.WalletDropped:
				log.Info("Old wallet dropped", "url", event.Wallet.URL())
				event.Wallet.Close()
			}
		}
	}()
}

// unlockAccounts unlocks any account specifically requested.
func unlockAccounts(ctx *cli.Context, stack *node.Node) {
	var unlocks []string
	inputs := strings.Split(ctx.GlobalString(utils.UnlockedAccountFlag.Name), ",")
	for _, input := range inputs {
		if trimmed := strings.TrimSpace(input); trimmed != "" {
			unlocks = append(unlocks, trimmed)
		}
	}
	// Short circuit if there is no account to unlock.
	if len(unlocks) == 0 {
		return
	}
	// If insecure account unlocking is not allowed if node's APIs are exposed to external.
	// Print warning log to user and skip unlocking.
	if !stack.Config().InsecureUnlockAllowed && stack.Config().ExtRPCEnabled() {
		utils.Fatalf("Account unlock with HTTP access is forbidden!")
	}
	ks := stack.AccountManager().Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
	passwords := utils.MakePasswordList(ctx)
	for i, account := range unlocks {
		unlockAccount(ks, account, i, passwords)
	}
}
