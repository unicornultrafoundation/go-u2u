// Package memorydb implements the key-value database layer based on memory maps.
package memorydb

import (
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb/devnulldb"
	"github.com/unicornultrafoundation/go-u2u/helios/u2udb/flushable"
)

// Database is an ephemeral key-value store. Apart from basic data storage
// functionality it also supports batch writes and iterating over the keyspace in
// binary-alphabetical order.
type Database struct {
	u2udb.Store
}

// New returns a wrapped map with all the required database interface methods
// implemented.
func New() *Database {
	return &Database{
		Store: flushable.Wrap(devnulldb.New()),
	}
}

// NewWithDrop is the same as New, but defines onDrop callback.
func NewWithDrop(drop func()) *Database {
	return &Database{
		Store: flushable.WrapWithDrop(devnulldb.New(), drop),
	}
}
