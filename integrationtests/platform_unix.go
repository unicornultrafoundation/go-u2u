//go:build !windows
// +build !windows

package integrationtests

import (
	"syscall"
)

// getIPCPath returns a platform-specific IPC path
func getIPCPath() string {
	// On Unix systems, we use a standard IPC path
	return "u2u.ipc"
}

// Stop shuts the underlying network down.
func (n *IntegrationTestNet) Stop() {
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	<-n.done
	n.done = nil
}
