//go:build !windows
// +build !windows

package integrationtests

import (
	"golang.org/x/net/nettest"
	"testing"
)

var (
	sfcDataDir string
	sfcTestOpt IntegrationTestNetOptions
)

func TestSFCStore000_Setup(t *testing.T) {
	var err error
	sfcDataDir, err = nettest.LocalPath()
	if err != nil {
		t.Fatalf("Failed to init the test network: %v", err)
	}
	sfcTestOpt = IntegrationTestNetOptions{Directory: sfcDataDir}
	StartIntegrationTestNetWithFakeGenesis(t, sfcTestOpt)
}

func TestSFCStore001_CanDumpSFCStorageAndThenSyncAgain(t *testing.T) {
	if err := DumpSFCStorage(sfcDataDir); err != nil {
		t.Fatalf("Failed to dump the SFC storage: %v", err)
	}
	sfcTestOpt.Sfc = true
}
