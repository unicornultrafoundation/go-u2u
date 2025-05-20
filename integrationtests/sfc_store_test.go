//go:build !windows
// +build !windows

package integrationtests

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/nettest"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/sfc"
)

var (
	dataDir      string
	testnet      *IntegrationTestNet
	err          error
	testAccounts []*Account
)

func setup() error {
	if err := initTestNetwork(); err != nil {
		return fmt.Errorf("failed to init test network: %v", err)
	}
	if err := testnet.DumpSFCStorage(dataDir); err != nil {
		return fmt.Errorf("failed to dump the SFC storage: %v", err)
	}
	return nil
}

func initTestNetwork() error {
	if dataDir == "" {
		dataDir, err = nettest.LocalPath()
		if err != nil {
			return err
		}
	}
	if testnet == nil {
		// start test network at temp datadir
		testnet, err = StartIntegrationTestNet(dataDir, false)
		if err != nil {
			return err
		}

		// init 10 test accounts
		time.Sleep(5 * time.Second)
		client, err := testnet.GetClient()
		if err != nil {
			return fmt.Errorf("failed to connect to the integration test network: %v", err)
		}
		defer client.Close()
		increment := new(big.Int).Mul(big.NewInt(100000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
		for i := 0; i < 10; i++ {
			testAccounts = append(testAccounts, NewAccount())
			address := testAccounts[i].Address()
			balance, err := client.BalanceAt(context.Background(), address, nil)
			if err != nil {
				return fmt.Errorf("failed to get balance for account: %v", err)
			}
			if err := testnet.EndowAccount(address, increment); err != nil {
				return fmt.Errorf("failed to endow account 1: %v", err)
			}
			want := balance.Add(balance, increment)
			balance, err = client.BalanceAt(context.Background(), address, nil)
			if err != nil {
				return fmt.Errorf("failed to get balance for account: %v", err)
			}
			if want, got := want, balance; want.Cmp(got) != 0 {
				return fmt.Errorf("unexpected balance for account, got %v, wanted %v", got, want)
			}
		}

		// make sure the fake network passed the first epoch, to ensure the
		// network is able to sync again after running db heal
		time.Sleep(20 * time.Second)
		testnet.Stop() // then shut down the network and dump SFC contract storage
	}
	return nil
}

func TestSFCStore_CanDumpSFCStorageAndThenSyncAgain(t *testing.T) {
	if err := setup(); err != nil {
		t.Fatal(err)
	}
	testnet, err = StartIntegrationTestNet(dataDir, true)
	if err != nil {
		t.Fatalf("failed to start the fake network: %v", err)
	}
	defer testnet.Stop()

	// try to read storage from SFC state db
	owner, err := testnet.SfcGetStorageAt(
		sfc.ContractAddress,
		common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000033"),
		nil)
	if err != nil {
		t.Fatalf("Failed to get owner of SFC contract: %v", err)
	}
	assert.Equal(t, common.BytesToAddress(owner).Hex(), testnet.validator.Address().Hex())

	// try to get balance from SFC state db
	client, err := testnet.GetClient()
	if err != nil {
		t.Fatalf("failed to connect to the integration test network: %v", err)
	}
	defer client.Close()
	balance, err := client.BalanceAt(context.Background(), testAccounts[0].Address(), nil)
	if err != nil {
		t.Fatalf("failed to get balance for account: %v", err)
	}
	want := new(big.Int).Mul(big.NewInt(100000), new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	if want, got := want, balance; want.Cmp(got) != 0 {
		t.Fatalf("Unexpected balance for account, got %v, wanted %v", got, want)
	}
}

func TestSFCStore_CanDelegateToValidator(t *testing.T) {

}
