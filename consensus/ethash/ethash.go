// Copyright 2017 The go-ethereum Authors
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

// Package ethash implements the ethash proof-of-work consensus engine.
package ethash

import (
	"github.com/unicornultrafoundation/go-u2u/log"
)

// Mode defines the type and amount of PoW verification an ethash engine makes.
type Mode uint

const (
	ModeNormal Mode = iota
	ModeShared
	ModeTest
	ModeFake
	ModeFullFake
)

// Config are the configuration parameters of the ethash.
type Config struct {
	CacheDir         string
	CachesInMem      int
	CachesOnDisk     int
	CachesLockMmap   bool
	DatasetDir       string
	DatasetsInMem    int
	DatasetsOnDisk   int
	DatasetsLockMmap bool
	PowMode          Mode

	// When set, notifications sent by the remote sealer will
	// be block header JSON objects instead of work package arrays.
	NotifyFull bool

	Log log.Logger `toml:"-"`
}