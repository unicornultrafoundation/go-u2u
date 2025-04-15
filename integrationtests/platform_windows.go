//go:build windows
// +build windows

package integrationtests

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// getIPCPath returns a platform-specific IPC path
func getIPCPath() string {
	// On Windows, we use a named pipe
	return fmt.Sprintf("\\\\.\\pipe\\u2u-%d", trulyRandInt(100000, 999999))
}

// Stop shuts the underlying network down.
func (n *IntegrationTestNet) Stop() {
	// Wait for the done channel to be closed
	// This ensures we don't return until the network is fully shut down
	if n.done != nil {
		<-n.done
		n.done = nil
	}
}

// trulyRandInt generates a crypto random integer used by the console tests to
// not clash network ports with other tests running cocurrently.
func trulyRandInt(lo, hi int) int {
	num, _ := rand.Int(rand.Reader, big.NewInt(int64(hi-lo)))
	return int(num.Int64()) + lo
}
