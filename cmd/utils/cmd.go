// Copyright 2014 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

// Package utils contains internal helper functions for go-ethereum commands.
package utils

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/eth/ethconfig"
	"github.com/unicornultrafoundation/go-u2u/internal/debug"
	"github.com/unicornultrafoundation/go-u2u/log"
	"github.com/unicornultrafoundation/go-u2u/node"
)

// Fatalf formats a message to standard error and exits the program.
// The message is also printed to standard output if standard error
// is redirected to a different file.
func Fatalf(format string, args ...interface{}) {
	w := io.MultiWriter(os.Stdout, os.Stderr)
	if runtime.GOOS == "windows" {
		// The SameFile check below doesn't work on Windows.
		// stdout is unlikely to get redirected though, so just print there.
		w = os.Stdout
	} else {
		outf, _ := os.Stdout.Stat()
		errf, _ := os.Stderr.Stat()
		if outf != nil && errf != nil && os.SameFile(outf, errf) {
			w = os.Stderr
		}
	}
	fmt.Fprintf(w, "Fatal: "+format+"\n", args...)
	os.Exit(1)
}

func StartNode(ctx *cli.Context, stack *node.Node) {
	if err := stack.Start(); err != nil {
		Fatalf("Error starting protocol stack: %v", err)
	}
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)

		minFreeDiskSpace := ethconfig.Defaults.TrieDirtyCache
		if ctx.GlobalIsSet(MinFreeDiskSpaceFlag.Name) {
			minFreeDiskSpace = ctx.GlobalInt(MinFreeDiskSpaceFlag.Name)
		} else {
			minFreeDiskSpace = 8192
		}
		if minFreeDiskSpace > 0 {
			go monitorFreeDiskSpace(sigc, stack.InstanceDir(), uint64(minFreeDiskSpace)*1024*1024)
		}

		<-sigc
		log.Info("Got interrupt, shutting down...")
		go stack.Close()
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				log.Warn("Already shutting down, interrupt more to panic.", "times", i-1)
			}
		}
		debug.Exit() // ensure trace and CPU profile data is flushed.
		debug.LoudPanic("boom")
	}()
}

func monitorFreeDiskSpace(sigc chan os.Signal, path string, freeDiskSpaceCritical uint64) {
	for {
		freeSpace, err := GetFreeDiskSpace(path)
		if err != nil {
			log.Warn("Failed to get free disk space", "path", path, "err", err)
			break
		}
		if freeSpace < freeDiskSpaceCritical {
			log.Error("Low disk space. Gracefully shutting down client to prevent database corruption.", "available", common.StorageSize(freeSpace))
			sigc <- syscall.SIGTERM
			break
		} else if freeSpace < 2*freeDiskSpaceCritical {
			log.Warn("Disk space is running low. client will shutdown if disk space runs below critical level.", "available", common.StorageSize(freeSpace), "critical_level", common.StorageSize(freeDiskSpaceCritical))
		}
		time.Sleep(60 * time.Second)
	}
}
