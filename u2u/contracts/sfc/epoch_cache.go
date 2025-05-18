package sfc

import (
	"math/big"
	"sync"

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

	// Mutex for thread safety
	mu sync.RWMutex

	// Cache statistics
	hitCount  uint64
	missCount uint64
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
	c.mu.Lock()
	defer c.mu.Unlock()

	c.currentEpoch = epoch
	c.stateCache = make(map[common.Hash]common.Hash)
	c.isInitialized = true
	c.hitCount = 0
	c.missCount = 0
}

// GetState first checks the epoch cache, then falls back to the state DB
func (c *EpochStateCache) GetState(addr common.Address, slot common.Hash) common.Hash {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isInitialized {
		c.missCount++
		return c.stateDB.GetState(addr, slot)
	}

	// Check epoch cache first
	if value, exists := c.stateCache[slot]; exists {
		c.hitCount++
		return value
	}

	// If not in cache, read from state DB and cache the result
	c.missCount++
	value := c.stateDB.GetState(addr, slot)
	c.stateCache[slot] = value
	return value
}

// SetState updates both the cache and the state DB
func (c *EpochStateCache) SetState(addr common.Address, slot, value common.Hash) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stateCache[slot] = value
	c.stateDB.SetState(addr, slot, value)
}

// GetBalance gets the balance from cache or state DB
func (c *EpochStateCache) GetBalance(addr common.Address) *big.Int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isInitialized {
		return c.stateDB.GetBalance(addr)
	}

	// For balance, we use a special slot hash
	balanceSlot := common.BytesToHash(addr.Bytes())
	if value, exists := c.stateCache[balanceSlot]; exists {
		c.hitCount++
		return new(big.Int).SetBytes(value.Bytes())
	}

	c.missCount++
	balance := c.stateDB.GetBalance(addr)
	c.stateCache[balanceSlot] = common.BigToHash(balance)
	return balance
}

// SetBalance updates both cache and state DB
func (c *EpochStateCache) SetBalance(addr common.Address, balance *big.Int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	balanceSlot := common.BytesToHash(addr.Bytes())
	c.stateCache[balanceSlot] = common.BigToHash(balance)
	// Use AddBalance with the difference between current and new balance
	currentBalance := c.stateDB.GetBalance(addr)
	diff := new(big.Int).Sub(balance, currentBalance)
	c.stateDB.AddBalance(addr, diff)
}

// GetNonce gets the nonce from cache or state DB
func (c *EpochStateCache) GetNonce(addr common.Address) uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.isInitialized {
		return c.stateDB.GetNonce(addr)
	}

	nonceSlot := common.BytesToHash(append(addr.Bytes(), []byte("nonce")...))
	if value, exists := c.stateCache[nonceSlot]; exists {
		c.hitCount++
		return new(big.Int).SetBytes(value.Bytes()).Uint64()
	}

	c.missCount++
	nonce := c.stateDB.GetNonce(addr)
	c.stateCache[nonceSlot] = common.BigToHash(big.NewInt(int64(nonce)))
	return nonce
}

// SetNonce updates both cache and state DB
func (c *EpochStateCache) SetNonce(addr common.Address, nonce uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	nonceSlot := common.BytesToHash(append(addr.Bytes(), []byte("nonce")...))
	c.stateCache[nonceSlot] = common.BigToHash(big.NewInt(int64(nonce)))
	c.stateDB.SetNonce(addr, nonce)
}

// GetCacheStats returns cache hit/miss statistics
func (c *EpochStateCache) GetCacheStats() (uint64, uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hitCount, c.missCount
}

// Clear clears the cache
func (c *EpochStateCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.stateCache = make(map[common.Hash]common.Hash)
	c.isInitialized = false
	c.hitCount = 0
	c.missCount = 0
}

// BatchSetState updates multiple state values in a single operation
func (c *EpochStateCache) BatchSetState(addr common.Address, updates []struct {
	slot  common.Hash
	value common.Hash
}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, update := range updates {
		c.stateCache[update.slot] = update.value
		c.stateDB.SetState(addr, update.slot, update.value)
	}
}
