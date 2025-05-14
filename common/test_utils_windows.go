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

//go:build windows
// +build windows

package common

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sys/windows"
)

// SendInterrupt sends an interrupt signal to the current process
// Windows implementation using Windows-specific APIs
func SendInterrupt() {
	// Create a channel to receive signals
	sigChan := make(chan os.Signal, 1)

	// Notify the channel of incoming interrupt signals
	signal.Notify(sigChan, os.Interrupt)

	// Send Ctrl+C event to the current console after 1 second
	time.AfterFunc(1*time.Second, func() {
		// Get the current console handle
		console := windows.Handle(windows.STD_INPUT_HANDLE)
		var mode uint32
		err := windows.GetConsoleMode(console, &mode)
		if err != nil {
			fmt.Println("Error getting console mode:", err)
			return
		}

		// Generate a Ctrl+C event
		kernel32 := windows.NewLazySystemDLL("kernel32.dll")
		generateConsoleCtrlEvent := kernel32.NewProc("GenerateConsoleCtrlEvent")
		r1, _, err := generateConsoleCtrlEvent.Call(
			uintptr(windows.CTRL_C_EVENT), // CTRL_C_EVENT = 0
			0) // 0 means to all processes attached to this console

		if r1 == 0 { // 0 means failure
			fmt.Println("Error generating Ctrl+C event:", err)
			return
		}
		fmt.Println("Ctrl+C event generated successfully")
	})

	// Block until a signal is received
	sig := <-sigChan
	fmt.Println("Received signal:", sig)

	// Perform cleanup or other actions before exiting
	fmt.Println("Exiting...")
}
