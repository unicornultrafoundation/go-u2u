package verwatcher

import (
	"sync/atomic"

	"github.com/unicornultrafoundation/go-helios/u2udb"

	"github.com/unicornultrafoundation/go-u2u/logger"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	mainDB u2udb.Store

	cache struct {
		networkVersion atomic.Value
		missedVersion  atomic.Value
	}

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(mainDB u2udb.Store) *Store {
	s := &Store{
		mainDB:   mainDB,
		Instance: logger.New("verwatcher-store"),
	}

	return s
}
