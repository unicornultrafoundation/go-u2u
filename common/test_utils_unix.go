// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

//go:build !windows
// +build !windows

package common

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// SendInterrupt sends an interrupt signal to the current process
// Unix implementation using syscall.SIGTERM
func SendInterrupt() {
	// Create a channel to receive signals.
	sigChan := make(chan os.Signal, 1)

	// Notify the channel of incoming interrupt signals (Ctrl+C or SIGINT).
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Send interrupt signal to itself after 1 second.
	time.AfterFunc(1*time.Second, func() {
		process, err := os.FindProcess(os.Getpid())
		if err != nil {
			fmt.Println("Error finding process:", err)
			return
		}
		err = process.Signal(os.Interrupt)
		if err != nil {
			fmt.Println("Error sending signal:", err)
			return
		}
	})

	// Block until a signal is received.
	sig := <-sigChan
	fmt.Println("Received signal:", sig)

	// Perform cleanup or other actions before exiting.
	fmt.Println("Exiting...")
}
