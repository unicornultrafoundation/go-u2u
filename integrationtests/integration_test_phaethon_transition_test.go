package integrationtests

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/unicornultrafoundation/go-u2u/ethclient"
	"github.com/unicornultrafoundation/go-u2u/u2u"
)

func TestPhaethonTransition_CanUpgradeNetworkRulesToPhaethon(t *testing.T) {
	// start test network with Clymene rules
	net := StartIntegrationTestNetWithFakeGenesis(t,
		IntegrationTestNetOptions{
			Upgrades: AsPointer(u2u.GetClymeneUpgrades()),
		})
	assert := require.New(t)
	client, err := net.GetClient()
	if err != nil {
		t.Fatalf("failed to get client; %v", err)
	}
	defer client.Close()

	// Verify Clymene is active (prevrandao should be set)
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		t.Fatalf("failed to get block header; %v", err)
	}
	assert.NotEqual(block.MixDigest().Big().Cmp(big.NewInt(0)), 0, "block prevrandao must be set in Clymene")

	// start upgrading to Phaethon
	receipt, err := net.CraftSFCTx(&net.validator, NodeDriverAuthAbi, &NodeDriverAuthAddr, big.NewInt(0),
		"updateNetworkRules", []byte(`{"Upgrades":{"Phaethon":true}}`))
	if err != nil {
		t.Fatalf("failed to send tx to upgrade network to Phaethon: %v", err)
	}
	assert.Equal(receipt.Status, uint64(1), "transaction to upgrade network to Phaethon must succeed")
	receipt, err = net.CraftSFCTx(&net.validator, NodeDriverAuthAbi, &NodeDriverAuthAddr, big.NewInt(0),
		"advanceEpochs", big.NewInt(10))
	if err != nil {
		t.Fatalf("failed to send tx to advance epoch: %v", err)
	}
	assert.Equal(receipt.Status, uint64(1), "transaction to advance epoch must succeed")
	// trigger new block to persist previous network changes
	if err := net.EndowAccount(net.validator.Address(), big.NewInt(1)); err != nil {
		t.Fatalf("Failed to endow account: %v", err)
	}
	// done upgrading to Phaethon

	// Verify Phaethon is active and inherited features still work
	testPhaethonInheritedFeatures(t, net, client)
}

func testPhaethonInheritedFeatures(t *testing.T, net *IntegrationTestNet, client *ethclient.Client) {
	assert := require.New(t)
	// trigger new block
	if err := net.EndowAccount(net.validator.Address(), big.NewInt(1)); err != nil {
		t.Fatalf("Failed to endow account: %v", err)
	}
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		t.Fatalf("failed to get block header; %v", err)
	}
	// Verify inherited Clymene features still work
	assert.NotEqual(block.MixDigest().Big().Cmp(big.NewInt(0)), 0, "block prevrandao must still be set after Phaethon upgrade")
	assert.Equal(block.Difficulty().Uint64(), uint64(0), "block difficulty must be 0 in Phaethon (inherited from Clymene)")
}
