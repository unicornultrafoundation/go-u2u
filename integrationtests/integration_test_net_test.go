//go:build !windows
// +build !windows

package integrationtests

import (
	"context"
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/integrationtests/contracts/counter"
)

func TestIntegrationTestNet_CanStartAndStopIntegrationTestNet(t *testing.T) {
	StartIntegrationTestNetWithFakeGenesis(t)
}

func TestIntegrationTestNet_CanStartMultipleConsecutiveInstances(t *testing.T) {
	for i := 0; i < 2; i++ {
		StartIntegrationTestNetWithFakeGenesis(t)
	}
}

func TestIntegrationTestNet_CanFetchInformationFromTheNetwork(t *testing.T) {
	net := StartIntegrationTestNetWithFakeGenesis(t)
	client, err := net.GetClient()
	if err != nil {
		t.Fatalf("Failed to connect to the integration test network: %v", err)
	}
	defer client.Close()
	block, err := client.BlockNumber(context.Background())
	if err != nil {
		t.Fatalf("Failed to get block number: %v", err)
	}
	if block == 0 || block > 1000 {
		t.Errorf("Unexpected block number: %v", block)
	}
}

func TestIntegrationTestNet_CanEndowAccountsWithTokens(t *testing.T) {
	net := StartIntegrationTestNetWithFakeGenesis(t)
	client, err := net.GetClient()
	if err != nil {
		t.Fatalf("Failed to connect to the integration test network: %v", err)
	}
	defer client.Close()
	address := common.Address{0x01}
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		t.Fatalf("Failed to get balance for account: %v", err)
	}
	for i := 0; i < 10; i++ {
		increment := int64(1000)
		if err := net.EndowAccount(address, increment); err != nil {
			t.Fatalf("Failed to endow account 1: %v", err)
		}
		want := balance.Add(balance, big.NewInt(int64(increment)))
		balance, err = client.BalanceAt(context.Background(), address, nil)
		if err != nil {
			t.Fatalf("Failed to get balance for account: %v", err)
		}
		if want, got := want, balance; want.Cmp(got) != 0 {
			t.Fatalf("Unexpected balance for account, got %v, wanted %v", got, want)
		}
		balance = want
	}
}

func TestIntegrationTestNet_CanDeployContracts(t *testing.T) {
	net := StartIntegrationTestNetWithFakeGenesis(t)
	_, receipt, err := DeployContract(net, counter.DeployCounter)
	if err != nil {
		t.Fatalf("Failed to deploy contract: %v", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		t.Errorf("Contract deployment failed: %+v", receipt)
	}
}

func TestIntegrationTestNet_CanInteractWithContract(t *testing.T) {
	net := StartIntegrationTestNetWithFakeGenesis(t)
	contract, _, err := DeployContract(net, counter.DeployCounter)
	if err != nil {
		t.Fatalf("Failed to deploy contract: %v", err)
	}
	receipt, err := net.Apply(contract.IncrementCounter)
	if err != nil {
		t.Fatalf("Failed to send transaction: %v", err)
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		t.Errorf("Contract deployment failed: %v", receipt)
	}
}
