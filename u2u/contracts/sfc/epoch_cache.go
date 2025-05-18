package sfc

import (
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// EpochStateCache provides an in-memory cache for state values during an epoch
type EpochStateCache struct {
	// Map to store state values for the current epoch
	// Key: slot hash, Value: state value
	stateCache map[common.Hash]common.Hash

	// Reference to the actual state DB
	stateDB vm.StateDB

	// Current epoch number
	currentEpoch *big.Int

	// Flag to track if cache is initialized
	isInitialized bool
}

// NewEpochStateCache creates a new epoch state cache instance
func NewEpochStateCache(stateDB vm.StateDB) *EpochStateCache {
	return &EpochStateCache{
		stateCache:    make(map[common.Hash]common.Hash),
		stateDB:       stateDB,
		isInitialized: false,
	}
}

// Initialize initializes the cache for a new epoch
func (c *EpochStateCache) Initialize(epoch *big.Int) {
	c.currentEpoch = epoch
	c.stateCache = make(map[common.Hash]common.Hash)
	c.isInitialized = true
}

// GetState first checks the epoch cache, then falls back to the state DB
func (c *EpochStateCache) GetState(addr common.Address, slot common.Hash) common.Hash {
	if !c.isInitialized {
		return c.stateDB.GetState(addr, slot)
	}

	// Check epoch cache first
	if value, exists := c.stateCache[slot]; exists {
		return value
	}

	// If not in cache, read from state DB and cache the result
	value := c.stateDB.GetState(addr, slot)
	c.stateCache[slot] = value
	return value
}

// SetState updates both the cache and the state DB
func (c *EpochStateCache) SetState(addr common.Address, slot, value common.Hash) {
	c.stateCache[slot] = value
	c.stateDB.SetState(addr, slot, value)
}

// Clear clears the cache
func (c *EpochStateCache) Clear() {
	c.stateCache = make(map[common.Hash]common.Hash)
	c.isInitialized = false
}
