package integrationtests

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand/v2"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/nettest"

	go_u2u "github.com/unicornultrafoundation/go-u2u"
	"github.com/unicornultrafoundation/go-u2u/accounts/abi"
	"github.com/unicornultrafoundation/go-u2u/accounts/abi/bind"
	"github.com/unicornultrafoundation/go-u2u/cmd/u2u/launcher"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/ethclient"
	"github.com/unicornultrafoundation/go-u2u/evmcore"
	"github.com/unicornultrafoundation/go-u2u/gossip/contract/driverauth100"
	"github.com/unicornultrafoundation/go-u2u/u2u"
)

var (
	NodeDriverAuthAddr = common.HexToAddress("0xd100ae0000000000000000000000000000000000")

	NodeDriverAuthAbi abi.ABI
)

func init() {
	NodeDriverAuthAbi, _ = abi.JSON(strings.NewReader(driverauth100.ContractMetaData.ABI))
}

// IntegrationTestNetOptions are configuration options for the integration test network.
type IntegrationTestNetOptions struct {
	// Upgrades specifies the upgrades to be used for the integration test network.
	// nil value will initialize network using SolarisUpgrades.
	Upgrades *u2u.Upgrades
	// NumNodes specifies the number of nodes to be started on the integration
	// test network.
	// A value of 0 is interpreted as 1.
	NumNodes  int
	Sfc       bool
	Directory string
}

// AsPointer is a utility function that returns a pointer to the given value.
// Useful to initialize values which nil value is semantically significant.
// E.g., to initialize the `Upgrades` field in `IntegrationTestNetOptions` to a non-nil value.
func AsPointer[T any](v T) *T {
	return &v
}

// IntegrationTestNet is a in-process test network for integration tests. When
// started, it runs a full U2U node maintaining a chain within the process
// containing this object. The network can be used to run transactions on and
// to perform queries against.
//
// The main purpose of this network is to facilitate end-to-end debugging of
// client code in the controlled scope of individual unit tests. When running
// tests against an integration test network instance, break-points can be set
// in the client code, thereby facilitating debugging.
//
// Additionally, by providing support for scripting test traffic on a network,
// integration test networks can also be used for automated integration and
// regression tests for client code.
type IntegrationTestNet struct {
	options        IntegrationTestNetOptions
	done           <-chan struct{}
	validator      Account
	httpClientPort int
}

func isPortFree(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

func getFreePort() (int, error) {
	var port int
	retries := 10
	for i := 0; i < retries; i++ {
		port = 1023 + (rand.Int()%math.MaxUint16 - 1023)
		if isPortFree("127.0.0.1", port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("failed to find a free port after %d retries (last %d)", retries, port)
}

// StartIntegrationTestNetWithFakeGenesis starts a single-node test network for
// integration tests using the Fake-Genesis procedure.
// The fake genesis procedure is mainly intended for demo and small scale test networks
// used for development and integration testing in Norma.
func StartIntegrationTestNetWithFakeGenesis(
	t *testing.T,
	options ...IntegrationTestNetOptions,
) *IntegrationTestNet {
	t.Helper()

	effectiveOptions, err := validateAndSanitizeOptions(t, options...)
	if err != nil {
		t.Fatal("failed to validate and sanitize options: ", err)
	}

	var upgrades string
	if *effectiveOptions.Upgrades == u2u.GetSolarisUpgrades() {
		upgrades = "solaris"
	} else if *effectiveOptions.Upgrades == u2u.GetVitriolUpgrades() {
		upgrades = "vitriol"
	} else {
		t.Fatal("fake genesis only supports sonic and allegro feature sets")
	}

	extraArgs := []string{"--fakenet", "1/1", "--upgrades", upgrades}
	if effectiveOptions.Sfc {
		extraArgs = append(extraArgs, "--sfc")
	}
	net, err := startIntegrationTestNet(
		t,
		extraArgs,
		effectiveOptions,
	)
	if err != nil {
		t.Fatal("failed to start integration test network: ", err)
	}
	return net
}

// startIntegrationTestNet starts a single-node test network for integration tests.
// The node serving the network is started in the same process as the caller. This
// is intended to facilitate debugging of client code in the context of a running
// node.
func startIntegrationTestNet(
	t *testing.T,
	fakenetArgs []string,
	options IntegrationTestNetOptions,
) (*IntegrationTestNet, error) {
	// find free ports for the http-client, ws-client, and network interfaces
	var err error
	httpClientPort, err := getFreePort()
	if err != nil {
		return nil, err
	}
	wsPort, err := getFreePort()
	if err != nil {
		return nil, err
	}
	netPort, err := getFreePort()
	if err != nil {
		return nil, err
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()
		// start the fakenet u2u node
		// equivalent to running `u2u ...` but in this local process
		os.Args = append([]string{
			"u2u",

			// data storage options
			"--datadir", options.Directory,
			"--datadir.minfreedisk", "0",

			// http-client option
			"--http", "--http.addr", "0.0.0.0", "--http.port", fmt.Sprint(httpClientPort),
			"--http.api", "admin,eth,web3,net,txpool,trace,debug,sfc",

			// websocket-client options
			"--ws", "--ws.addr", "0.0.0.0", "--ws.port", fmt.Sprint(wsPort),
			"--ws.api", "admin,eth",

			//  net options
			"--port", fmt.Sprint(netPort),
			"--nat", "none",
			"--nodiscover",
			"--ipcpath", getIPCPath(),
			"--cache", "8192",
		}, fakenetArgs...)
		err := launcher.Run()
		if err != nil {
			panic(fmt.Sprint("Failed to start the fake network:", err))
		}
	}()
	net := &IntegrationTestNet{
		options:        options,
		done:           done,
		validator:      Account{evmcore.FakeKey(1)},
		httpClientPort: httpClientPort,
	}
	// connect to blockchain network
	client, err := net.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the U2U client: %w", err)
	}
	defer client.Close()
	const timeout = 300 * time.Second
	start := time.Now()
	// wait for the node to be ready to serve requests
	const maxDelay = 100 * time.Millisecond
	delay := time.Millisecond
	for time.Since(start) < timeout {
		_, err := client.ChainID(context.Background())
		if err != nil {
			time.Sleep(delay)
			delay = 2 * delay
			if delay > maxDelay {
				delay = maxDelay
			}
			continue
		}
		t.Cleanup(net.Stop)
		return net, nil
	}
	return nil, fmt.Errorf("failed to successfully start up a test network within %d", timeout)
}

// EndowAccount sends a requested amount of tokens to the given account. This is
// mainly intended to provide funds to accounts for testing purposes.
func (n *IntegrationTestNet) EndowAccount(
	address common.Address,
	value int64,
) error {
	client, err := n.GetClient()
	if err != nil {
		return fmt.Errorf("failed to connect to the network: %w", err)
	}
	defer client.Close()
	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}
	// The requested funds are moved from the validator account to the target account.
	nonce, err := client.NonceAt(context.Background(), n.validator.Address(), nil)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}
	price, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}
	transaction, err := types.SignTx(types.NewTx(&types.AccessListTx{
		ChainID:  chainId,
		Gas:      21000,
		GasPrice: price,
		To:       &address,
		Value:    big.NewInt(value),
		Nonce:    nonce,
	}), types.NewLondonSigner(chainId), n.validator.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}
	_, err = n.Run(transaction)
	return err
}

// Run sends the given transaction to the network and waits for it to be processed.
// The resulting receipt is returned. This function times out after 10 seconds.
func (n *IntegrationTestNet) Run(tx *types.Transaction) (*types.Receipt, error) {
	client, err := n.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the U2U client: %w", err)
	}
	defer client.Close()
	err = client.SendTransaction(context.Background(), tx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}
	return n.GetReceipt(tx.Hash())
}

// GetReceipt waits for the receipt of the given transaction hash to be available.
// The function times out after 10 seconds.
func (n *IntegrationTestNet) GetReceipt(txHash common.Hash) (*types.Receipt, error) {
	client, err := n.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the U2U client: %w", err)
	}
	defer client.Close()
	// Wait for the response with some exponential backoff.
	const maxDelay = 100 * time.Millisecond
	now := time.Now()
	delay := time.Millisecond
	for time.Since(now) < 100*time.Second {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if errors.Is(err, go_u2u.NotFound) {
			time.Sleep(delay)
			delay = 2 * delay
			if delay > maxDelay {
				delay = maxDelay
			}
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
		}
		return receipt, nil
	}
	return nil, fmt.Errorf("failed to get transaction receipt: timeout")
}

// Apply sends a transaction to the network using the network's validator account
// and waits for the transaction to be processed. The resulting receipt is returned.
func (n *IntegrationTestNet) Apply(
	issue func(*bind.TransactOpts) (*types.Transaction, error),
) (*types.Receipt, error) {
	txOpts, err := n.GetTransactOptions(&n.validator)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	transaction, err := issue(txOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}
	return n.GetReceipt(transaction.Hash())
}

// GetTransactOptions provides transaction options to be used to send a transaction
// with the given account. The options include the chain ID, a suggested gas price,
// the next free nonce of the given account, and a hard-coded gas limit of 1e6.
// The main purpose of this function is to provide a convenient way to collect all
// the necessary information required to create a transaction in one place.
func (n *IntegrationTestNet) GetTransactOptions(account *Account) (*bind.TransactOpts, error) {
	client, err := n.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the U2U client: %w", err)
	}
	defer client.Close()
	ctxt := context.Background()
	chainId, err := client.ChainID(ctxt)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}
	gasPrice, err := client.SuggestGasPrice(ctxt)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price suggestion: %w", err)
	}
	nonce, err := client.NonceAt(ctxt, account.Address(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}
	txOpts, err := bind.NewKeyedTransactorWithChainID(account.PrivateKey, chainId)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction options: %w", err)
	}
	txOpts.GasPrice = new(big.Int).Mul(gasPrice, big.NewInt(2))
	txOpts.Nonce = big.NewInt(int64(nonce))
	txOpts.GasLimit = 1e6
	return txOpts, nil
}

// GetClient provides raw access to a fresh connection to the network.
// The resulting client must be closed after use.
func (n *IntegrationTestNet) GetClient() (*ethclient.Client, error) {
	return ethclient.Dial(fmt.Sprintf("http://localhost:%d", n.httpClientPort))
}

// DeployContract is a utility function handling the deployment of a contract on the network.
// The contract is deployed with by the network's validator account. The function returns the
// deployed contract instance and the transaction receipt.
func DeployContract[T any](n *IntegrationTestNet, deploy contractDeployer[T]) (*T, *types.Receipt, error) {
	client, err := n.GetClient()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to the U2U client: %w", err)
	}
	defer client.Close()
	transactOptions, err := n.GetTransactOptions(&n.validator)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get transaction options: %w", err)
	}
	_, transaction, contract, err := deploy(transactOptions, client)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to deploy contract: %w", err)
	}
	receipt, err := n.GetReceipt(transaction.Hash())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get receipt: %w", err)
	}
	return contract, receipt, nil
}

// contractDeployer is the type of the deployment functions generated by abigen.
type contractDeployer[T any] func(*bind.TransactOpts, bind.ContractBackend) (common.Address, *types.Transaction, *T, error)

// GetStorageAt returns the storage from the state at the given address, key and
// block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta block
// numbers are also allowed.
func (n *IntegrationTestNet) GetStorageAt(account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	client, err := n.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the U2U client: %w", err)
	}
	defer client.Close()
	return client.StorageAt(context.Background(), account, key, blockNumber)
}

// validateAndSanitizeOptions ensures that the options are valid and sets the default values.
func validateAndSanitizeOptions(t *testing.T, options ...IntegrationTestNetOptions) (IntegrationTestNetOptions, error) {
	if len(options) > 1 {
		return IntegrationTestNetOptions{}, fmt.Errorf("expected at most one option, got %d", len(options))
	}

	if len(options) == 0 {
		dir, err := nettest.LocalPath()
		if err != nil {
			return IntegrationTestNetOptions{}, fmt.Errorf("failed to init temp dir, error %s", err)
		}
		return IntegrationTestNetOptions{
			Upgrades:  AsPointer(u2u.GetSolarisUpgrades()),
			NumNodes:  1,
			Directory: dir,
		}, nil
	}
	if options[0].NumNodes <= 0 {
		options[0].NumNodes = 1
	}
	if options[0].Upgrades == nil {
		options[0].Upgrades = AsPointer(u2u.GetVitriolUpgrades())
	}
	if options[0].Directory == "" {
		dir, err := nettest.LocalPath()
		if err != nil {
			return IntegrationTestNetOptions{}, fmt.Errorf("failed to init temp dir, error %s", err)
		}
		options[0].Directory = dir
	}

	return options[0], nil
}
