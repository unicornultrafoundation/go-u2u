//go:build !windows
// +build !windows

package integrationtests

import (
	"testing"
	"time"

	"golang.org/x/net/nettest"
)

var (
	dataDir    string
	testnet    *IntegrationTestNet
	err        error
	sfcTestOpt IntegrationTestNetOptions
)

func setup(t *testing.T) error {
	if dataDir == "" {
		dataDir, err = nettest.LocalPath()
		if err != nil {
			return err
		}
	}
	if testnet == nil {
		sfcTestOpt = IntegrationTestNetOptions{Directory: dataDir}
		testnet = StartIntegrationTestNetWithFakeGenesis(t, sfcTestOpt)
		// make sure the fake network passed the first epoch, to ensure the
		// network is able to sync again after running db heal
		time.Sleep(20 * time.Second)
		testnet.Stop() // then shut down the network and dump SFC contract storage
	}
	return nil
}

func TestSFCStore_CanDumpSFCStorageAndThenSyncAgain(t *testing.T) {
	if err := setup(t); err != nil {
		t.Fatal(err)
	}
	if err := testnet.DumpSFCStorage(dataDir); err != nil {
		t.Fatalf("Failed to dump the SFC storage: %v", err)
	}
	if err := testnet.HealDB(dataDir); err != nil {
		t.Fatalf("Failed to heal the DB after dumping: %v", err)
	}
	// restart the network on that healed DB after dumping
	testnet = StartIntegrationTestNetWithFakeGenesis(t, sfcTestOpt)
	defer testnet.Stop()
}
