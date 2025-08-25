package constant_manager

import (
	"math/big"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/log"
)

// CachedValue represents a cached constant value with validation info
type CachedValue struct {
	Value       *big.Int
	StateRoot   common.Hash
	BlockNumber *big.Int
	Timestamp   time.Time
}

// CMCache stores cached values from the ConstantManager contract
type CMCache struct {
	// LRU cache for values with automatic eviction
	cache *lru.Cache
	
	// Owner address stored separately since it's not a big.Int
	Owner common.Address

	// State validation for cache consistency
	StateRoot   common.Hash
	BlockNumber *big.Int
	
	// Batched state root checking optimization
	cachedStorageRoot common.Hash
	storageRootValid  bool
	
	// Mutex for thread safety
	mutex sync.RWMutex

	// Cache metadata
	LastInvalidated time.Time

	// Invalidate cache signal
	NeedInvalidating bool
}

// Cache configuration constants
const (
	DefaultCacheSize = 50  // Small cache since CM constants rarely change
	MaxCacheAge      = 30 * time.Second // Auto-invalidate after 30 seconds
)

// Package-level cache instance
var cmCache *CMCache

// init initializes the global CM cache
func init() {
	cache, err := lru.New(DefaultCacheSize)
	if err != nil {
		log.Error("Failed to create ConstantManager LRU cache", "err", err)
		panic(err)
	}
	
	cmCache = &CMCache{
		cache:            cache,
		NeedInvalidating: true,
		LastInvalidated:  time.Now(),
	}
}

// GetCMCache returns the ConstantManager cache instance
func GetCMCache() *CMCache {
	return cmCache
}

// getStorageRootCached returns the storage root, using cached value if available
func (c *CMCache) getStorageRootCached(evm *vm.EVM) common.Hash {
	if c.storageRootValid {
		return c.cachedStorageRoot
	}
	
	if evm != nil && evm.SfcStateDB != nil {
		c.cachedStorageRoot = evm.SfcStateDB.GetStorageRoot(ContractAddress)
		c.storageRootValid = true
		return c.cachedStorageRoot
	}
	
	return common.Hash{}
}

// IsStale checks if the cache is stale based on state root or time
func (c *CMCache) IsStale(evm *vm.EVM) bool {
	// Quick check without lock first
	if c.NeedInvalidating {
		return true
	}
	
	// Check cache age without lock
	if time.Since(c.LastInvalidated) > MaxCacheAge {
		return true
	}
	
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	// Check state root mismatch using batched storage root
	if evm != nil && evm.SfcStateDB != nil {
		currentStateRoot := c.getStorageRootCached(evm)
		if c.StateRoot != currentStateRoot {
			return true
		}
		
		// Check block number mismatch
		currentBlockNumber := evm.Context.BlockNumber
		if c.BlockNumber == nil || c.BlockNumber.Cmp(currentBlockNumber) != 0 {
			return true
		}
	}
	
	return false
}

// GetValueSafe returns a cached value with lazy loading and validation
func (c *CMCache) GetValueSafe(evm *vm.EVM, key string) *big.Int {
	// Check if cache needs refresh
	if c.IsStale(evm) {
		log.Debug("ConstantManager cache is stale, refreshing", "key", key)
		InvalidateCmCache(evm)
	}
	
	// No need for per-value validation since IsStale() already validated
	return c.GetValue(key)
}

// InvalidateCmCache invalidates the cache with the correct values from CM contract.
func InvalidateCmCache(evm *vm.EVM) {
	log.Info("Invalidating ConstantsManager cache...")
	cmCache.mutex.Lock()
	defer cmCache.mutex.Unlock()

	var currentStateRoot common.Hash
	var currentBlockNumber *big.Int
	
	if evm != nil && evm.SfcStateDB != nil {
		currentStateRoot = cmCache.getStorageRootCached(evm)
		currentBlockNumber = new(big.Int).Set(evm.Context.BlockNumber)
	}
	
	now := time.Now()

	// Helper function to load and cache a value
	loadValue := func(key string, slot int64) {
		val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(slot)))
		cachedVal := &CachedValue{
			Value:       val.Big(),
			StateRoot:   currentStateRoot,
			BlockNumber: currentBlockNumber,
			Timestamp:   now,
		}
		cmCache.cache.Add(key, cachedVal)
	}

	// Load all constant values into LRU cache
	loadValue(MinSelfStakeKey, minSelfStakeSlot)
	loadValue(MaxDelegatedRatioKey, maxDelegatedRatioSlot)
	loadValue(ValidatorCommissionKey, validatorCommissionSlot)
	loadValue(BurntFeeShareKey, burntFeeShareSlot)
	loadValue(TreasuryFeeShareKey, treasuryFeeShareSlot)
	loadValue(UnlockedRewardRatioKey, unlockedRewardRatioSlot)
	loadValue(MinLockupDurationKey, minLockupDurationSlot)
	loadValue(MaxLockupDurationKey, maxLockupDurationSlot)
	loadValue(WithdrawalPeriodEpochsKey, withdrawalPeriodEpochsSlot)
	loadValue(WithdrawalPeriodTimeKey, withdrawalPeriodTimeSlot)
	loadValue(BaseRewardPerSecondKey, baseRewardPerSecondSlot)
	loadValue(OfflinePenaltyThresholdBlocksNumKey, offlinePenaltyThresholdBlocksNumSlot)
	loadValue(OfflinePenaltyThresholdTimeKey, offlinePenaltyThresholdTimeSlot)
	loadValue(TargetGasPowerPerSecondKey, targetGasPowerPerSecondSlot)
	loadValue(GasPriceBalancingCounterweightKey, gasPriceBalancingCounterweightSlot)

	// Load owner separately
	val := evm.SfcStateDB.GetState(ContractAddress, common.BigToHash(big.NewInt(ownerSlot)))
	cmCache.Owner = common.BytesToAddress(val.Bytes())

	// Update cache metadata
	cmCache.StateRoot = currentStateRoot
	cmCache.BlockNumber = currentBlockNumber
	cmCache.LastInvalidated = now
	cmCache.NeedInvalidating = false
	
	// Reset batched storage root cache
	cmCache.storageRootValid = false
	
	log.Info("ConstantsManager cache invalidated successfully")
}

// UpdateCacheValue updates a specific value in the cache and marks it for invalidation
func UpdateCacheValue(slot int64, value *big.Int) {
	cmCache.mutex.Lock()
	defer cmCache.mutex.Unlock()

	var key string
	switch slot {
	case minSelfStakeSlot:
		key = MinSelfStakeKey
	case maxDelegatedRatioSlot:
		key = MaxDelegatedRatioKey
	case validatorCommissionSlot:
		key = ValidatorCommissionKey
	case burntFeeShareSlot:
		key = BurntFeeShareKey
	case treasuryFeeShareSlot:
		key = TreasuryFeeShareKey
	case unlockedRewardRatioSlot:
		key = UnlockedRewardRatioKey
	case minLockupDurationSlot:
		key = MinLockupDurationKey
	case maxLockupDurationSlot:
		key = MaxLockupDurationKey
	case withdrawalPeriodEpochsSlot:
		key = WithdrawalPeriodEpochsKey
	case withdrawalPeriodTimeSlot:
		key = WithdrawalPeriodTimeKey
	case baseRewardPerSecondSlot:
		key = BaseRewardPerSecondKey
	case offlinePenaltyThresholdBlocksNumSlot:
		key = OfflinePenaltyThresholdBlocksNumKey
	case offlinePenaltyThresholdTimeSlot:
		key = OfflinePenaltyThresholdTimeKey
	case targetGasPowerPerSecondSlot:
		key = TargetGasPowerPerSecondKey
	case gasPriceBalancingCounterweightSlot:
		key = GasPriceBalancingCounterweightKey
	default:
		return
	}

	// Remove the specific key from cache since it's now stale
	cmCache.cache.Remove(key)
	
	// Mark cache as needing full invalidation
	cmCache.NeedInvalidating = true
}

// UpdateOwner updates the owner address in the cache
func UpdateOwner(owner common.Address) {
	cmCache.mutex.Lock()
	defer cmCache.mutex.Unlock()

	cmCache.Owner = owner
}

// InvalidateCMCacheFlag marks the ConstantManager cache as needing invalidation
func InvalidateCMCacheFlag() {
	cmCache.mutex.Lock()
	defer cmCache.mutex.Unlock()
	
	cmCache.NeedInvalidating = true
}

// Getter functions for cached values

// GetValue returns a cached value by key (legacy method, prefer GetValueSafe)
func (c *CMCache) GetValue(key string) *big.Int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	// Try to get from LRU cache
	if val, found := c.cache.Get(key); found {
		if cachedVal, ok := val.(*CachedValue); ok {
			return new(big.Int).Set(cachedVal.Value)
		}
	}

	log.Warn("ConstantsManager cache miss (legacy method)", "key", key)
	return big.NewInt(0)
}


// GetOwner returns the cached owner of the contract
func (c *CMCache) GetOwner() common.Address {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.Owner
}
