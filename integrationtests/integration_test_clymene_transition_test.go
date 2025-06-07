package integrationtests

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unicornultrafoundation/go-u2u/accounts/abi/bind"
	"github.com/unicornultrafoundation/go-u2u/ethclient"
	"github.com/unicornultrafoundation/go-u2u/integrationtests/contracts/prevrandao"
	"github.com/unicornultrafoundation/go-u2u/u2u"
)

func TestClymeneTransition_TestComputePrevRandao(t *testing.T) {
	net := StartIntegrationTestNetWithFakeGenesis(t,
		IntegrationTestNetOptions{
			Upgrades: AsPointer(u2u.GetClymeneUpgrades()),
		})
	// Deploy the contract.
	contract, _, err := DeployContract(net, prevrandao.DeployPrevrandao)
	if err != nil {
		t.Fatalf("failed to deploy contract; %v", err)
	}
	// Collect the current PrevRandao fee from the head state.
	receipt, err := net.Apply(contract.LogCurrentPrevRandao)
	if err != nil {
		t.Fatalf("failed to log current prevrandao; %v", err)
	}
	if len(receipt.Logs) != 1 {
		t.Fatalf("unexpected number of logs; expected 1, got %d", len(receipt.Logs))
	}
	entry, err := contract.ParseCurrentPrevRandao(*receipt.Logs[0])
	if err != nil {
		t.Fatalf("failed to parse log; %v", err)
	}
	fromLog := entry.Prevrandao

	client, err := net.GetClient()
	if err != nil {
		t.Fatalf("failed to get client; %v", err)
	}
	defer client.Close()
	block, err := client.BlockByNumber(context.Background(), receipt.BlockNumber)
	if err != nil {
		t.Fatalf("failed to get block header; %v", err)
	}
	fromLatestBlock := block.MixDigest().Big() // MixDigest == MixHash == PrevRandao
	if block.Difficulty().Uint64() != 0 {
		t.Errorf("incorrect header difficulty got: %d, want: %d", block.Difficulty().Uint64(), 0)
	}
	// Collect the prevrandao from the archive.
	fromArchive, err := contract.GetPrevRandao(&bind.CallOpts{BlockNumber: receipt.BlockNumber})
	if err != nil {
		t.Fatalf("failed to get prevrandao from archive; %v", err)
	}
	if fromLog.Sign() < 1 {
		t.Fatalf("invalid prevrandao from log; %v", fromLog)
	}

	if fromLog.Cmp(fromLatestBlock) != 0 {
		t.Errorf("prevrandao of block %v mismatch; from log %v, from block %v", receipt.BlockNumber, fromLog, fromLatestBlock)
	}
	if fromLog.Cmp(fromArchive) != 0 {
		t.Errorf("prevrandao mismatch; from log %v, from archive %v", fromLog, fromArchive)
	}

	fromSecondLastBlock, err := contract.GetPrevRandao(&bind.CallOpts{BlockNumber: big.NewInt(receipt.BlockNumber.Int64() - 1)})
	if err != nil {
		t.Fatalf("failed to get prevrandao from archive; %v", err)
	}

	if fromSecondLastBlock.Cmp(fromLatestBlock) == 0 {
		t.Errorf("prevrandao must be different for each block, found same: %s, %s", fromSecondLastBlock, fromLatestBlock)
	}
}

func TestClymeneTransition_CanUpgradeNetworkRulesToClymene(t *testing.T) {
	// start test network with Solaris rules
	net := StartIntegrationTestNetWithFakeGenesis(t,
		IntegrationTestNetOptions{
			Upgrades: AsPointer(u2u.GetSolarisUpgrades()),
		})
	assert := require.New(t)
	client, err := net.GetClient()
	if err != nil {
		t.Fatalf("failed to get client; %v", err)
	}
	defer client.Close()

	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		t.Fatalf("failed to get block header; %v", err)
	}
	assert.Equal(block.MixDigest().Big().Cmp(big.NewInt(0)), 0, "block prevrandao must be 0 before Clymene upgrade")

	// start upgrading to Clymene
	receipt, err := net.CraftSFCTx(&net.validator, NodeDriverAuthAbi, &NodeDriverAuthAddr, big.NewInt(0),
		"updateNetworkRules", []byte(`{"Upgrades":{"Clymene":true}}`))
	if err != nil {
		t.Fatalf("failed to send tx to upgrade network to Clymene: %v", err)
	}
	assert.Equal(receipt.Status, uint64(1), "transaction to upgrade network to Clymene must succeed")
	receipt, err = net.CraftSFCTx(&net.validator, NodeDriverAuthAbi, &NodeDriverAuthAddr, big.NewInt(0),
		"advanceEpochs", big.NewInt(10))
	if err != nil {
		t.Fatalf("failed to send tx to advance epoch: %v", err)
	}
	assert.Equal(receipt.Status, uint64(1), "transaction to advance epoch must succeed")
	// trigger new block to persist previous network changes
	if err := net.EndowAccount(net.validator.Address(), 1); err != nil {
		t.Fatalf("Failed to endow account: %v", err)
	}
	// done upgrading to Clymene

	testPrevRandaoMustBeSet(t, net, client)
}

func testPrevRandaoMustBeSet(t *testing.T, net *IntegrationTestNet, client *ethclient.Client) {
	assert := require.New(t)
	// trigger new block for the latest block to have prevrandao value
	if err := net.EndowAccount(net.validator.Address(), 1); err != nil {
		t.Fatalf("Failed to endow account: %v", err)
	}
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		t.Fatalf("failed to get block header; %v", err)
	}
	assert.NotEqual(block.MixDigest().Big().Cmp(big.NewInt(0)), 0, "block prevrandao must be set after Clymene upgrade")
}
