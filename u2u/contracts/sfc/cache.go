// Package sfc implements the SFC (Special Fee Contract) precompiled contract.
package sfc

import (
	"fmt"
	"math/big"
	"strconv"

	lru "github.com/hashicorp/golang-lru"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/crypto"
)

// HashCache stores previously calculated hashes to avoid redundant calculations
type HashCache struct {
	cache *lru.Cache
}

// NewHashCache creates a new hash cache with LRU eviction
func NewHashCache(capacity int) *HashCache {
	if capacity <= 0 {
		capacity = 1000 // Default capacity
	}
	cache, _ := lru.New(capacity)
	return &HashCache{
		cache: cache,
	}
}

// GetOrCompute gets a hash from the cache or computes it if not found
func (c *HashCache) GetOrCompute(input []byte) common.Hash {
	// Convert input to string for map key
	key := string(input)

	if value, found := c.cache.Get(key); found {
		return value.(common.Hash)
	}

	// Compute hash if not in cache
	hash := crypto.Keccak256Hash(input)

	// Store in cache
	c.cache.Add(key, hash)

	return hash
}

// CachedKeccak256 computes the Keccak256 hash of the input using the cache
func (c *HashCache) CachedKeccak256(input []byte) []byte {
	return c.GetOrCompute(input).Bytes()
}

// CachedKeccak256Hash computes the Keccak256 hash of the input using the cache
func (c *HashCache) CachedKeccak256Hash(input []byte) common.Hash {
	return c.GetOrCompute(input)
}

// SlotCache stores previously calculated storage slots
type SlotCache struct {
	cache *lru.Cache
}

// NewSlotCache creates a new slot cache with LRU eviction
func NewSlotCache(capacity int) *SlotCache {
	if capacity <= 0 {
		capacity = 1000 // Default capacity
	}
	cache, _ := lru.New(capacity)
	return &SlotCache{
		cache: cache,
	}
}

// GetOrCompute gets a slot from the cache or computes it using the provided function
func (c *SlotCache) GetOrCompute(key string, computeFunc func() (*big.Int, uint64)) (*big.Int, uint64) {
	if value, found := c.cache.Get(key); found {
		return value.(*big.Int), 0 // No gas used for cache hit
	}

	// Compute if not found
	slot, gasUsed := computeFunc()

	// Store in cache
	c.cache.Add(key, slot)

	return slot, gasUsed
}

// SFCCache contains all caches used by the SFC package
type SFCCache struct {
	Hash *HashCache
	Slot *SlotCache

	// Specialized caches for common operations
	ValidatorSlot *lru.Cache
	EpochSlot     *lru.Cache

	// Hash input caches
	HashInputs        *lru.Cache // Cache for hash inputs (validatorID + slot)
	AddressHashInputs *lru.Cache // Cache for address hash inputs (address + slot)
	NestedHashInputs  *lru.Cache // Cache for nested hash inputs

	// Unified ABI encoding cache
	AbiPackCache *lru.Cache // Cache for all ABI packed results
}

// Cache capacity constants
const (
	DefaultHashCacheSize          = 2000
	DefaultSlotCacheSize          = 1000
	DefaultValidatorSlotCacheSize = 1000
	DefaultEpochSlotCacheSize     = 200
	DefaultHashInputCacheSize     = 1500
	DefaultAbiPackCacheSize       = 300
)

// NewSFCCache creates a new SFC cache with default capacities
func NewSFCCache() *SFCCache {
	return NewSFCCacheWithCapacities(
		DefaultHashCacheSize,
		DefaultSlotCacheSize,
		DefaultValidatorSlotCacheSize,
		DefaultEpochSlotCacheSize,
		DefaultHashInputCacheSize,
		DefaultAbiPackCacheSize,
	)
}

// NewSFCCacheWithCapacities creates a new SFC cache with custom capacities
func NewSFCCacheWithCapacities(hashCap, slotCap, validatorSlotCap, epochSlotCap, hashInputCap, abiPackCap int) *SFCCache {
	validatorSlotCache, _ := lru.New(validatorSlotCap)
	epochSlotCache, _ := lru.New(epochSlotCap)
	hashInputsCache, _ := lru.New(hashInputCap)
	addressHashInputsCache, _ := lru.New(hashInputCap)
	nestedHashInputsCache, _ := lru.New(hashInputCap)
	abiPackCache, _ := lru.New(abiPackCap)
	
	cache := &SFCCache{
		Hash:              NewHashCache(hashCap),
		Slot:              NewSlotCache(slotCap),
		ValidatorSlot:     validatorSlotCache,
		EpochSlot:         epochSlotCache,
		HashInputs:        hashInputsCache,
		AddressHashInputs: addressHashInputsCache,
		NestedHashInputs:  nestedHashInputsCache,
		AbiPackCache:      abiPackCache,
	}

	return cache
}

// Package-level cache instance
var sfcCache = NewSFCCache()

// GetSFCCache returns the SFC cache instance
func GetSFCCache() *SFCCache {
	return sfcCache
}

// Helper functions for common operations

// CachedKeccak256 computes the Keccak256 hash using the cache
func CachedKeccak256(input []byte) []byte {
	return sfcCache.Hash.CachedKeccak256(input)
}

// CachedKeccak256Hash computes the Keccak256 hash using the cache
func CachedKeccak256Hash(input []byte) common.Hash {
	return sfcCache.Hash.CachedKeccak256Hash(input)
}

// GetCachedValidatorSlot gets or computes the validator slot
func GetCachedValidatorSlot(validatorID *big.Int) (*big.Int, uint64) {
	key := validatorID.String()

	if value, found := sfcCache.ValidatorSlot.Get(key); found {
		return value.(*big.Int), 0 // No gas used for cache hit
	}

	// Compute if not found
	slot, gasUsed := getValidatorStatusSlot(validatorID)

	// Store in cache
	sfcCache.ValidatorSlot.Add(key, slot)

	return slot, gasUsed
}

// GetCachedEpochSnapshotSlot gets or computes the epoch snapshot slot
func GetCachedEpochSnapshotSlot(epoch *big.Int) (*big.Int, uint64) {
	key := epoch.String()

	if value, found := sfcCache.EpochSlot.Get(key); found {
		return value.(*big.Int), 0 // No gas used for cache hit
	}

	// Compute if not found
	slot, gasUsed := getEpochSnapshotSlot(epoch)

	// Store in cache
	sfcCache.EpochSlot.Add(key, slot)

	return slot, gasUsed
}

// ClearCache clears all caches
func ClearCache() {
	sfcCache.Hash.cache.Purge()
	sfcCache.Slot.cache.Purge()
	sfcCache.ValidatorSlot.Purge()
	sfcCache.EpochSlot.Purge()
	sfcCache.HashInputs.Purge()
	sfcCache.AddressHashInputs.Purge()
	sfcCache.NestedHashInputs.Purge()
	sfcCache.AbiPackCache.Purge()
}

// CreateHashInput creates a hash input from a validator ID and slot constant
func CreateHashInput(validatorID *big.Int, slotConstant int64) []byte {
	// Create a cache key from validatorID and slotConstant
	cacheKey := validatorID.String() + "_" + strconv.FormatInt(slotConstant, 10)

	// Check if the hash input is already cached
	if value, found := sfcCache.HashInputs.Get(cacheKey); found {
		return value.([]byte)
	}

	// If not in cache, create the hash input directly
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)
	slotBytes := common.LeftPadBytes(big.NewInt(slotConstant).Bytes(), 32)

	// Use the byte slice pool for the result
	hashInput := GetByteSlice()
	if cap(hashInput) < len(validatorIDBytes)+len(slotBytes) {
		// If the slice from the pool is too small, allocate a new one
		hashInput = make([]byte, 0, len(validatorIDBytes)+len(slotBytes))
	}

	// Combine the bytes
	hashInput = append(hashInput, validatorIDBytes...)
	hashInput = append(hashInput, slotBytes...)

	// Store in cache
	sfcCache.HashInputs.Add(cacheKey, hashInput)

	return hashInput
}

// CreateAddressHashInput creates a hash input from an address and slot constant
func CreateAddressHashInput(addr common.Address, slotConstant int64) []byte {
	// Create a cache key from address and slotConstant
	cacheKey := addr.String() + "_" + strconv.FormatInt(slotConstant, 10)

	// Check if the hash input is already cached
	if value, found := sfcCache.AddressHashInputs.Get(cacheKey); found {
		return value.([]byte)
	}

	// If not in cache, create the hash input directly
	addrBytes := common.LeftPadBytes(addr.Bytes(), 32)
	slotBytes := common.LeftPadBytes(big.NewInt(slotConstant).Bytes(), 32)

	// Use the byte slice pool for the result
	hashInput := GetByteSlice()
	if cap(hashInput) < len(addrBytes)+len(slotBytes) {
		// If the slice from the pool is too small, allocate a new one
		hashInput = make([]byte, 0, len(addrBytes)+len(slotBytes))
	}

	// Combine the bytes
	hashInput = append(hashInput, addrBytes...)
	hashInput = append(hashInput, slotBytes...)

	// Store in cache
	sfcCache.AddressHashInputs.Add(cacheKey, hashInput)

	return hashInput
}

// CreateNestedHashInput creates a hash input from a validator ID and a hash
func CreateNestedHashInput(validatorID *big.Int, hash []byte) []byte {
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)

	// Use the byte slice pool for the result
	hashInput := GetByteSlice()
	if cap(hashInput) < len(validatorIDBytes)+len(hash) {
		// If the slice from the pool is too small, allocate a new one
		hashInput = make([]byte, 0, len(validatorIDBytes)+len(hash))
	}

	// Combine the bytes
	hashInput = append(hashInput, validatorIDBytes...)
	hashInput = append(hashInput, hash...)

	return hashInput
}

// CreateNestedAddressHashInput creates a hash input from an address and a hash
func CreateNestedAddressHashInput(addr common.Address, hash []byte) []byte {
	// Create a cache key from address and hash
	cacheKey := addr.String() + "_" + common.Bytes2Hex(hash)

	// Check if the hash input is already cached
	if value, found := sfcCache.NestedHashInputs.Get(cacheKey); found {
		return value.([]byte)
	}

	// If not in cache, create the hash input directly
	addrBytes := common.LeftPadBytes(addr.Bytes(), 32)

	// Use the byte slice pool for the result
	hashInput := GetByteSlice()
	if cap(hashInput) < len(addrBytes)+len(hash) {
		// If the slice from the pool is too small, allocate a new one
		hashInput = make([]byte, 0, len(addrBytes)+len(hash))
	}

	// Combine the bytes
	hashInput = append(hashInput, addrBytes...)
	hashInput = append(hashInput, hash...)

	// Store in cache
	sfcCache.NestedHashInputs.Add(cacheKey, hashInput)

	return hashInput
}

// CreateValidatorMappingHashInput creates a hash input from a validator ID and a mapping slot
func CreateValidatorMappingHashInput(validatorID *big.Int, mappingSlot *big.Int) []byte {
	// Create a cache key from validatorID and mappingSlot
	cacheKey := validatorID.String() + "_mapping_" + mappingSlot.String()

	// Check if the hash input is already cached
	if value, found := sfcCache.HashInputs.Get(cacheKey); found {
		return value.([]byte)
	}

	// If not in cache, create the hash input directly
	validatorIDBytes := common.LeftPadBytes(validatorID.Bytes(), 32)
	mappingSlotBytes := common.LeftPadBytes(mappingSlot.Bytes(), 32)

	// Use the byte slice pool for the result
	hashInput := GetByteSlice()
	if cap(hashInput) < len(validatorIDBytes)+len(mappingSlotBytes) {
		// If the slice from the pool is too small, allocate a new one
		hashInput = make([]byte, 0, len(validatorIDBytes)+len(mappingSlotBytes))
	}

	// Combine the bytes
	hashInput = append(hashInput, validatorIDBytes...)
	hashInput = append(hashInput, mappingSlotBytes...)

	// Store in cache
	sfcCache.HashInputs.Add(cacheKey, hashInput)

	return hashInput
}

// CreateAddressMethodHashInput creates a hash input from an address and a method ID
func CreateAddressMethodHashInput(addr common.Address, methodID []byte) []byte {
	// Create a cache key from address and methodID
	cacheKey := addr.String() + "_method_" + common.Bytes2Hex(methodID)

	// Check if the hash input is already cached
	if value, found := sfcCache.AddressHashInputs.Get(cacheKey); found {
		return value.([]byte)
	}

	// If not in cache, create the hash input directly
	addrBytes := common.LeftPadBytes(addr.Bytes(), 32)

	// Use the byte slice pool for the result
	hashInput := GetByteSlice()
	if cap(hashInput) < len(addrBytes)+len(methodID) {
		// If the slice from the pool is too small, allocate a new one
		hashInput = make([]byte, 0, len(addrBytes)+len(methodID))
	}

	// Combine the bytes
	hashInput = append(hashInput, addrBytes...)
	hashInput = append(hashInput, methodID...)

	// Store in cache
	sfcCache.AddressHashInputs.Add(cacheKey, hashInput)

	return hashInput
}

// CreateAddressParamsHashInput creates a hash input from an address and multiple parameters
func CreateAddressParamsHashInput(addr common.Address, params ...[]byte) []byte {
	// Create a cache key from address and params
	cacheKey := addr.String()
	for _, param := range params {
		cacheKey += "_param_" + common.Bytes2Hex(param)
	}

	// Check if the hash input is already cached
	if value, found := sfcCache.AddressHashInputs.Get(cacheKey); found {
		return value.([]byte)
	}

	// If not in cache, create the hash input directly
	addrBytes := common.LeftPadBytes(addr.Bytes(), 32)

	// Calculate total length needed
	totalLength := len(addrBytes)
	for _, param := range params {
		totalLength += len(param)
	}

	// Use the byte slice pool for the result
	hashInput := GetByteSlice()
	if cap(hashInput) < totalLength {
		// If the slice from the pool is too small, allocate a new one
		hashInput = make([]byte, 0, totalLength)
	}

	// Combine the bytes
	hashInput = append(hashInput, addrBytes...)
	for _, param := range params {
		hashInput = append(hashInput, param...)
	}

	// Store in cache
	sfcCache.AddressHashInputs.Add(cacheKey, hashInput)

	return hashInput
}

// CreateOffsetSlotHashInput creates a hash input from an offset and a slot
func CreateOffsetSlotHashInput(offset int64, slot *big.Int) []byte {
	// Create a cache key from offset and slot
	cacheKey := strconv.FormatInt(offset, 10) + "_slot_" + slot.String()

	// Check if the hash input is already cached
	if value, found := sfcCache.HashInputs.Get(cacheKey); found {
		return value.([]byte)
	}

	// If not in cache, create the hash input directly
	offsetBytes := common.LeftPadBytes(big.NewInt(offset).Bytes(), 32)
	slotBytes := common.LeftPadBytes(slot.Bytes(), 32)

	// Use the byte slice pool for the result
	hashInput := GetByteSlice()
	if cap(hashInput) < len(offsetBytes)+len(slotBytes) {
		// If the slice from the pool is too small, allocate a new one
		hashInput = make([]byte, 0, len(offsetBytes)+len(slotBytes))
	}

	// Combine the bytes
	hashInput = append(hashInput, offsetBytes...)
	hashInput = append(hashInput, slotBytes...)

	// Store in cache
	sfcCache.HashInputs.Add(cacheKey, hashInput)

	return hashInput
}

// CreateAndHashOffsetSlot creates a hash input from an offset and a slot, then hashes it
func CreateAndHashOffsetSlot(offset int64, slot *big.Int) []byte {
	// Get the hash input
	hashInput := CreateOffsetSlotHashInput(offset, slot)

	// Use cached hash calculation
	return CachedKeccak256(hashInput)
}

// ABI type constants for identifying which ABI to use
const (
	SfcAbiType            = "sfc"
	CMAbiType             = "cm"
	NodeDriverAbiType     = "nodedriver"
	NodeDriverAuthAbiType = "nodedriverauth"
)

// CachedAbiPack packs arguments using the specified ABI and caches the result
// Only caches results for calls without parameters to avoid cache bloat
func CachedAbiPack(abiType, method string, args ...interface{}) ([]byte, error) {
	// Only cache if there are no arguments
	if len(args) == 0 {
		// Create a cache key from just the ABI type and method
		key := abiType + ":" + method

		// Check if the result is already cached
		if value, ok := sfcCache.AbiPackCache.Get(key); ok {
			return value.([]byte), nil
		}

		// Not in cache, pack it
		var packed []byte
		var err error

		switch abiType {
		case SfcAbiType:
			packed, err = SfcAbi.Methods[method].Outputs.Pack()
		case CMAbiType:
			packed, err = CMAbi.Methods[method].Outputs.Pack()
		case NodeDriverAbiType:
			packed, err = NodeDriverAbi.Methods[method].Outputs.Pack()
		case NodeDriverAuthAbiType:
			packed, err = NodeDriverAuthAbi.Methods[method].Outputs.Pack()
		default:
			return nil, fmt.Errorf("unknown ABI type: %s", abiType)
		}

		if err != nil {
			return nil, err
		}

		// Store in cache
		sfcCache.AbiPackCache.Add(key, packed)

		return packed, nil
	}

	// For calls with parameters, don't use cache
	var packed []byte
	var err error

	switch abiType {
	case SfcAbiType:
		packed, err = SfcAbi.Methods[method].Outputs.Pack(args...)
	case CMAbiType:
		packed, err = CMAbi.Methods[method].Outputs.Pack(args...)
	case NodeDriverAbiType:
		packed, err = NodeDriverAbi.Methods[method].Outputs.Pack(args...)
	case NodeDriverAuthAbiType:
		packed, err = NodeDriverAuthAbi.Methods[method].Outputs.Pack(args...)
	default:
		return nil, fmt.Errorf("unknown ABI type: %s", abiType)
	}

	return packed, err
}
