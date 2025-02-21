package integrationtests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/nettest"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/gossip/sfcstore"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/driverauth"
	"github.com/unicornultrafoundation/go-u2u/u2u/contracts/sfc"
)

var (
	dataDir string
	testnet *IntegrationTestNet
	err     error
)

func setup() error {
	if dataDir == "" {
		dataDir, err = nettest.LocalPath()
		if err != nil {
			return err
		}
	}
	if testnet == nil {
		testnet, err = StartIntegrationTestNet(dataDir)
		if err != nil {
			return err
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
	if err := testnet.DumpSFCStorage(dataDir); err != nil {
		t.Fatalf("Failed to dump the SFC storage: %v", err)
	}
	if err := testnet.HealDB(dataDir); err != nil {
		t.Fatalf("Failed to heal the DB after dumping: %v", err)
	}
	// restart the network on that healed DB after dumping
	testnet, err = StartIntegrationTestNet(dataDir)
	if err != nil {
		t.Fatalf("Failed to start the fake network: %v", err)
	}
	defer testnet.Stop()
}

func TestSFCStore_ReturnCorrectNdaSfc(t *testing.T) {
	if err := setup(); err != nil {
		t.Fatal(err)
	}
	testnet, err = StartIntegrationTestNet(dataDir)
	if err != nil {
		t.Fatalf("Failed to start the fake network: %v", err)
	}
	defer testnet.Stop()
	ndaSfc, err := testnet.GetStorageAt(driverauth.ContractAddress, sfcstore.NdaSfc, nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, sfc.ContractAddress, common.BytesToAddress(ndaSfc))
}
