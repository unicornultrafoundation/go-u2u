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

// Stop shuts down the underlying network gracefully using pure channel communication.
// This implementation follows the Sonic network pattern - no OS signals needed!
// The launcher listens to the shutdown channel via AppControl and closes gracefully.
// It is safe to call this method multiple times (idempotent).
func (n *IntegrationTestNet) Stop() {
	// Step 1: Idempotency check - if already stopped, return early
	if n.done == nil {
		return
	}

	// Step 2: Signal shutdown via channel (no OS signals needed!)
	// Closing the shutdown channel signals the launcher to stop gracefully
	if n.shutdown != nil {
		close(n.shutdown)
	}

	// Step 3: Wait for the node goroutine to complete shutdown
	// The done channel will be closed when the launcher exits cleanly
	<-n.done

	// Step 4: Resource cleanup
	// Mark channels as nil to prevent double-close and indicate stopped state
	n.done = nil
	n.shutdown = nil

	// Step 5: Debug logging
	if n.tempDir != "" {
		fmt.Printf("Integration test network stopped. Data directory: %s\n", n.tempDir)
	}
}

// trulyRandInt generates a crypto random integer used by the console tests to
// not clash network ports with other tests running cocurrently.
func trulyRandInt(lo, hi int) int {
	num, _ := rand.Int(rand.Reader, big.NewInt(int64(hi-lo)))
	return int(num.Int64()) + lo
}
