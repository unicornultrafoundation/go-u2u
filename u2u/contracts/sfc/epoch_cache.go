package sfc

import (
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

// ObjectPool provides a pool of reusable objects
type ObjectPool struct {
	pool sync.Pool
}

// NewObjectPool creates a new object pool
func NewObjectPool(new func() interface{}) *ObjectPool {
	return &ObjectPool{
		pool: sync.Pool{
			New: new,
		},
	}
}

// Get retrieves an object from the pool
func (p *ObjectPool) Get() interface{} {
	return p.pool.Get()
}

// Put returns an object to the pool
func (p *ObjectPool) Put(x interface{}) {
	p.pool.Put(x)
}

// EpochStateCache provides an in-memory cache for state values during an epoch
type EpochStateCache struct {
	// Use sync.Map for better concurrent performance
	stateCache sync.Map

	// Reference to the actual state DB
	stateDB vm.StateDB

	// Current epoch number
	currentEpoch *big.Int

	// Flag to track if cache is initialized
	isInitialized bool

	// Cache statistics
	hitCount  atomic.Uint64
	missCount atomic.Uint64

	// Mutex to ensure state DB consistency
	stateMutex sync.Mutex

	// Object pools for frequently used objects
	bigIntPool *ObjectPool
	hashPool   *ObjectPool
}

// NewEpochStateCache creates a new epoch state cache instance
func NewEpochStateCache(stateDB vm.StateDB) *EpochStateCache {
	return &EpochStateCache{
		stateDB:       stateDB,
		isInitialized: false,
		bigIntPool: NewObjectPool(func() interface{} {
			return new(big.Int)
		}),
		hashPool: NewObjectPool(func() interface{} {
			return common.Hash{}
		}),
	}
}

// Initialize initializes the cache for a new epoch
func (c *EpochStateCache) Initialize(epoch *big.Int) {
	c.currentEpoch = epoch
	c.stateCache = sync.Map{}
	c.isInitialized = true
	c.hitCount.Store(0)
	c.missCount.Store(0)
}

// getBigInt retrieves a big.Int from the pool
func (c *EpochStateCache) getBigInt() *big.Int {
	return c.bigIntPool.Get().(*big.Int)
}

// GetState first checks the epoch cache, then falls back to the state DB
func (c *EpochStateCache) GetState(addr common.Address, slot common.Hash) common.Hash {
	if !c.isInitialized {
		c.missCount.Add(1)
		return c.stateDB.GetState(addr, slot)
	}

	// Check epoch cache first
	if value, exists := c.stateCache.Load(slot); exists {
		c.hitCount.Add(1)
		return value.(common.Hash)
	}

	// If not in cache, read from state DB and cache the result
	c.missCount.Add(1)
	value := c.stateDB.GetState(addr, slot)
	c.stateCache.Store(slot, value)
	return value
}

// SetState updates both the cache and the state DB
func (c *EpochStateCache) SetState(addr common.Address, slot, value common.Hash) {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()

	c.stateCache.Store(slot, value)
	c.stateDB.SetState(addr, slot, value)
}

// GetBalance gets the balance from cache or state DB
func (c *EpochStateCache) GetBalance(addr common.Address) *big.Int {
	if !c.isInitialized {
		return c.stateDB.GetBalance(addr)
	}

	// For balance, we use a special slot hash
	balanceSlot := common.BytesToHash(addr.Bytes())
	if value, exists := c.stateCache.Load(balanceSlot); exists {
		return new(big.Int).SetBytes(value.(common.Hash).Bytes())
	}

	balance := c.stateDB.GetBalance(addr)
	c.stateCache.Store(balanceSlot, common.BigToHash(balance))
	return balance
}

// SetBalance updates both cache and state DB
func (c *EpochStateCache) SetBalance(addr common.Address, balance *big.Int) {
	balanceSlot := common.BytesToHash(addr.Bytes())
	c.stateCache.Store(balanceSlot, common.BigToHash(balance))
	// Use AddBalance with the difference between current and new balance
	currentBalance := c.stateDB.GetBalance(addr)
	diff := new(big.Int).Sub(balance, currentBalance)
	c.stateDB.AddBalance(addr, diff)
}

// GetNonce gets the nonce from cache or state DB
func (c *EpochStateCache) GetNonce(addr common.Address) uint64 {
	if !c.isInitialized {
		return c.stateDB.GetNonce(addr)
	}

	nonceSlot := common.BytesToHash(append(addr.Bytes(), []byte("nonce")...))
	if value, exists := c.stateCache.Load(nonceSlot); exists {
		return new(big.Int).SetBytes(value.(common.Hash).Bytes()).Uint64()
	}

	nonce := c.stateDB.GetNonce(addr)
	c.stateCache.Store(nonceSlot, common.BigToHash(big.NewInt(int64(nonce))))
	return nonce
}

// SetNonce updates both cache and state DB
func (c *EpochStateCache) SetNonce(addr common.Address, nonce uint64) {
	nonceSlot := common.BytesToHash(append(addr.Bytes(), []byte("nonce")...))
	c.stateCache.Store(nonceSlot, common.BigToHash(big.NewInt(int64(nonce))))
	c.stateDB.SetNonce(addr, nonce)
}

// GetCacheStats returns cache hit/miss statistics
func (c *EpochStateCache) GetCacheStats() (uint64, uint64) {
	return c.hitCount.Load(), c.missCount.Load()
}

// Clear clears the cache
func (c *EpochStateCache) Clear() {
	c.stateCache = sync.Map{}
	c.isInitialized = false
	c.hitCount.Store(0)
	c.missCount.Store(0)
}

// BatchSetState updates multiple state values in a single operation
// while maintaining state DB consistency
func (c *EpochStateCache) BatchSetState(addr common.Address, updates []struct {
	slot  common.Hash
	value common.Hash
}) {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()

	// Process updates in order to maintain state DB consistency
	for _, update := range updates {
		c.stateCache.Store(update.slot, update.value)
		c.stateDB.SetState(addr, update.slot, update.value)
	}
}

// PrefetchState prefetches multiple state values into the cache
func (c *EpochStateCache) PrefetchState(addr common.Address, slots []common.Hash) {
	if !c.isInitialized {
		return
	}

	// Prefetch values in parallel using goroutines
	type result struct {
		slot  common.Hash
		value common.Hash
	}
	results := make(chan result, len(slots))

	for _, slot := range slots {
		go func(s common.Hash) {
			value := c.stateDB.GetState(addr, s)
			results <- result{slot: s, value: value}
		}(slot)
	}

	// Collect results and store in cache
	for i := 0; i < len(slots); i++ {
		r := <-results
		c.stateCache.Store(r.slot, r.value)
	}
}

// BatchGetState retrieves multiple state values in a single operation
func (c *EpochStateCache) BatchGetState(addr common.Address, slots []common.Hash) ([]common.Hash, uint64) {
	var gasUsed uint64 = 0

	if !c.isInitialized {
		values := make([]common.Hash, len(slots))
		for i, slot := range slots {
			values[i] = c.stateDB.GetState(addr, slot)
			gasUsed += SloadGasCost
		}
		return values, gasUsed
	}

	values := make([]common.Hash, len(slots))
	missingSlots := make([]common.Hash, 0, len(slots))
	missingIndices := make([]int, 0, len(slots))

	// First try to get values from cache
	for i, slot := range slots {
		if value, exists := c.stateCache.Load(slot); exists {
			c.hitCount.Add(1)
			values[i] = value.(common.Hash)
		} else {
			c.missCount.Add(1)
			missingSlots = append(missingSlots, slot)
			missingIndices = append(missingIndices, i)
		}
	}

	// If there are missing values, fetch them from state DB
	if len(missingSlots) > 0 {
		// Use a worker pool for parallel fetching
		workerCount := runtime.NumCPU()
		if workerCount > len(missingSlots) {
			workerCount = len(missingSlots)
		}

		type result struct {
			index int
			value common.Hash
		}
		results := make(chan result, len(missingSlots))
		jobs := make(chan int, len(missingSlots))

		// Start workers
		var wg sync.WaitGroup
		for w := 0; w < workerCount; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range jobs {
					value := c.stateDB.GetState(addr, missingSlots[i])
					results <- result{
						index: missingIndices[i],
						value: value,
					}
				}
			}()
		}

		// Send jobs
		for i := range missingSlots {
			jobs <- i
		}
		close(jobs)

		// Wait for all workers to finish
		go func() {
			wg.Wait()
			close(results)
		}()

		// Collect results
		for r := range results {
			values[r.index] = r.value
			c.stateCache.Store(missingSlots[r.index], r.value)
			gasUsed += SloadGasCost
		}
	}

	return values, gasUsed
}
