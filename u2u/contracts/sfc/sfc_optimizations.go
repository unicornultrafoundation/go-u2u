package sfc

import (
	"math/big"
	"sync"

	"github.com/unicornultrafoundation/go-u2u/common"
)

// ConstantsCache caches frequently accessed constants
type ConstantsCache struct {
	baseRewardPerSecond *big.Int
	validatorCommission *big.Int
	burntFeeShare       *big.Int
	treasuryFeeShare    *big.Int
	decimalUnit         *big.Int
	mu                  sync.RWMutex
}

// GetBaseRewardPerSecond returns the cached base reward per second
func (c *ConstantsCache) GetBaseRewardPerSecond() *big.Int {
	c.mu.RLock()
	if c.baseRewardPerSecond != nil {
		c.mu.RUnlock()
		return c.baseRewardPerSecond
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.baseRewardPerSecond == nil {
		c.baseRewardPerSecond = getConstantsManagerVariable("baseRewardPerSecond")
	}
	return c.baseRewardPerSecond
}

// GetValidatorCommission returns the cached validator commission
func (c *ConstantsCache) GetValidatorCommission() *big.Int {
	c.mu.RLock()
	if c.validatorCommission != nil {
		c.mu.RUnlock()
		return c.validatorCommission
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.validatorCommission == nil {
		c.validatorCommission = getConstantsManagerVariable("validatorCommission")
	}
	return c.validatorCommission
}

// GetBurntFeeShare returns the cached burnt fee share
func (c *ConstantsCache) GetBurntFeeShare() *big.Int {
	c.mu.RLock()
	if c.burntFeeShare != nil {
		c.mu.RUnlock()
		return c.burntFeeShare
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.burntFeeShare == nil {
		c.burntFeeShare = getConstantsManagerVariable("burntFeeShare")
	}
	return c.burntFeeShare
}

// GetTreasuryFeeShare returns the cached treasury fee share
func (c *ConstantsCache) GetTreasuryFeeShare() *big.Int {
	c.mu.RLock()
	if c.treasuryFeeShare != nil {
		c.mu.RUnlock()
		return c.treasuryFeeShare
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.treasuryFeeShare == nil {
		c.treasuryFeeShare = getConstantsManagerVariable("treasuryFeeShare")
	}
	return c.treasuryFeeShare
}

// GetDecimalUnit returns the cached decimal unit
func (c *ConstantsCache) GetDecimalUnit() *big.Int {
	c.mu.RLock()
	if c.decimalUnit != nil {
		c.mu.RUnlock()
		return c.decimalUnit
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.decimalUnit == nil {
		c.decimalUnit = getDecimalUnit()
	}
	return c.decimalUnit
}

// GasTracker tracks gas usage
type GasTracker struct {
	used uint64
	mu   sync.Mutex
}

// Add adds gas to the tracker
func (t *GasTracker) Add(gas uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.used += gas
}

// Get returns the total gas used
func (t *GasTracker) Get() uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.used
}

// BatchStateReader reads multiple state values in a single batch
type BatchStateReader struct {
	epochCache *EpochStateCache
	reads      []struct {
		address common.Address
		key     common.Hash
		value   *common.Hash
	}
}

// NewBatchStateReader creates a new batch state reader
func NewBatchStateReader(epochCache *EpochStateCache) *BatchStateReader {
	return &BatchStateReader{
		epochCache: epochCache,
		reads: make([]struct {
			address common.Address
			key     common.Hash
			value   *common.Hash
		}, 0),
	}
}

// AddRead adds a state read to the batch
func (r *BatchStateReader) AddRead(address common.Address, key common.Hash, value *common.Hash) {
	r.reads = append(r.reads, struct {
		address common.Address
		key     common.Hash
		value   *common.Hash
	}{address, key, value})
}

// Execute executes all batched reads
func (r *BatchStateReader) Execute() {
	for _, read := range r.reads {
		*read.value = r.epochCache.GetState(read.address, read.key)
	}
}
