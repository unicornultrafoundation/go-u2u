package integrationtests

import (
	"testing"
	"time"

	"golang.org/x/net/nettest"
)

func TestSFCStore_CanDumpSFCStorageAndThenSyncAgain(t *testing.T) {
	dataDir, err := nettest.LocalPath()
	if err != nil {
		t.Fatalf(err.Error())
	}
	net, err := StartIntegrationTestNet(dataDir)
	if err != nil {
		t.Fatalf("Failed to start the fake network: %v", err)
	}
	// make sure the fake network passed the first epoch, to ensure the
	// network is able to sync again after running db heal
	time.Sleep(20 * time.Second)
	net.Stop() // then shut down the network and dump SFC contract storage
	if err := net.DumpSFCStorage(dataDir); err != nil {
		t.Fatalf("Failed to dump the SFC storage: %v", err)
	}
	if err := net.HealDB(dataDir); err != nil {
		t.Fatalf("Failed to heal the DB after dumping: %v", err)
	}
	// restart the network on that healed DB after dumping
	net, err = StartIntegrationTestNet(dataDir)
	if err != nil {
		t.Fatalf("Failed to start the fake network: %v", err)
	}
	defer net.Stop()
}
