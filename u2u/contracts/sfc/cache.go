// Package sfc implements the SFC (Special Fee Contract) precompiled contract.
package sfc

import (
	"math/big"
	"strconv"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/crypto"
)

// HashCache stores previously calculated hashes to avoid redundant calculations
type HashCache struct {
	// Map from input bytes to calculated hash
	cache map[string]common.Hash
}

// NewHashCache creates a new hash cache
func NewHashCache() *HashCache {
	return &HashCache{
		cache: make(map[string]common.Hash),
	}
}

// GetOrCompute gets a hash from the cache or computes it if not found
func (c *HashCache) GetOrCompute(input []byte) common.Hash {
	// Convert input to string for map key
	key := string(input)

	if hash, found := c.cache[key]; found {
		return hash
	}

	// Compute hash if not in cache
	hash := crypto.Keccak256Hash(input)

	// Store in cache
	c.cache[key] = hash

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
	// Map from string representation of inputs to calculated slots
	cache map[string]*big.Int
}

// NewSlotCache creates a new slot cache
func NewSlotCache() *SlotCache {
	return &SlotCache{
		cache: make(map[string]*big.Int),
	}
}

// GetOrCompute gets a slot from the cache or computes it using the provided function
func (c *SlotCache) GetOrCompute(key string, computeFunc func() (*big.Int, uint64)) (*big.Int, uint64) {
	if slot, found := c.cache[key]; found {
		return slot, 0 // No gas used for cache hit
	}

	// Compute if not found
	slot, gasUsed := computeFunc()

	// Store in cache
	c.cache[key] = slot

	return slot, gasUsed
}

// SFCCache contains all caches used by the SFC package
type SFCCache struct {
	Hash *HashCache
	Slot *SlotCache

	// Specialized caches for common operations
	ValidatorSlot map[string]*big.Int
	EpochSlot     map[string]*big.Int

	// Padding caches
	PaddedValidatorIDs map[string][]byte // Cache for padded validator IDs
	PaddedAddresses    map[string][]byte // Cache for padded addresses

	// Pre-computed padded slot constants
	PaddedSlots map[int64][]byte // Cache for padded slot constants

	// Hash input caches
	HashInputs      map[string][]byte // Cache for hash inputs (validatorID + slot)
	AddressHashInputs map[string][]byte // Cache for address hash inputs (address + slot)
}

// NewSFCCache creates a new SFC cache
func NewSFCCache() *SFCCache {
	cache := &SFCCache{
		Hash:              NewHashCache(),
		Slot:              NewSlotCache(),
		ValidatorSlot:     make(map[string]*big.Int),
		EpochSlot:         make(map[string]*big.Int),
		PaddedValidatorIDs: make(map[string][]byte),
		PaddedAddresses:    make(map[string][]byte),
		PaddedSlots:        make(map[int64][]byte),
		HashInputs:         make(map[string][]byte),
		AddressHashInputs:  make(map[string][]byte),
	}

	// Initialize padded slot constants
	cache.initPaddedSlots()

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

	if slot, found := sfcCache.ValidatorSlot[key]; found {
		return slot, 0 // No gas used for cache hit
	}

	// Compute if not found
	slot, gasUsed := getValidatorStatusSlot(validatorID)

	// Store in cache
	sfcCache.ValidatorSlot[key] = slot

	return slot, gasUsed
}

// GetCachedEpochSnapshotSlot gets or computes the epoch snapshot slot
func GetCachedEpochSnapshotSlot(epoch *big.Int) (*big.Int, uint64) {
	key := epoch.String()

	if slot, found := sfcCache.EpochSlot[key]; found {
		return slot, 0 // No gas used for cache hit
	}

	// Compute if not found
	slot, gasUsed := getEpochSnapshotSlot(epoch)

	// Store in cache
	sfcCache.EpochSlot[key] = slot

	return slot, gasUsed
}

// ClearCache clears all caches
func ClearCache() {
	sfcCache = NewSFCCache()
}

// initPaddedSlots initializes the padded slot constants
func (c *SFCCache) initPaddedSlots() {
	// Pre-compute padded bytes for common slot constants
	c.PaddedSlots[validatorSlot] = common.LeftPadBytes(big.NewInt(validatorSlot).Bytes(), 32)
	c.PaddedSlots[validatorCommissionSlot] = common.LeftPadBytes(big.NewInt(validatorCommissionSlot).Bytes(), 32)
	c.PaddedSlots[validatorPubkeySlot] = common.LeftPadBytes(big.NewInt(validatorPubkeySlot).Bytes(), 32)
	c.PaddedSlots[stakeSlot] = common.LeftPadBytes(big.NewInt(stakeSlot).Bytes(), 32)
	c.PaddedSlots[lockupInfoSlot] = common.LeftPadBytes(big.NewInt(lockupInfoSlot).Bytes(), 32)
	c.PaddedSlots[rewardsStashSlot] = common.LeftPadBytes(big.NewInt(rewardsStashSlot).Bytes(), 32)
	c.PaddedSlots[stashedLockupRewardsSlot] = common.LeftPadBytes(big.NewInt(stashedLockupRewardsSlot).Bytes(), 32)
	c.PaddedSlots[stashedRewardsUntilEpochSlot] = common.LeftPadBytes(big.NewInt(stashedRewardsUntilEpochSlot).Bytes(), 32)
	c.PaddedSlots[withdrawalRequestSlot] = common.LeftPadBytes(big.NewInt(withdrawalRequestSlot).Bytes(), 32)
	c.PaddedSlots[validatorIDSlot] = common.LeftPadBytes(big.NewInt(validatorIDSlot).Bytes(), 32)
	c.PaddedSlots[epochSnapshotSlot] = common.LeftPadBytes(big.NewInt(epochSnapshotSlot).Bytes(), 32)
}

// GetPaddedValidatorID returns the padded bytes for a validator ID
func GetPaddedValidatorID(validatorID *big.Int) []byte {
	key := validatorID.String()

	if padded, found := sfcCache.PaddedValidatorIDs[key]; found {
		return padded
	}

	padded := common.LeftPadBytes(validatorID.Bytes(), 32)
	sfcCache.PaddedValidatorIDs[key] = padded

	return padded
}

// GetPaddedAddress returns the padded bytes for an address
func GetPaddedAddress(addr common.Address) []byte {
	key := addr.String()

	if padded, found := sfcCache.PaddedAddresses[key]; found {
		return padded
	}

	padded := common.LeftPadBytes(addr.Bytes(), 32)
	sfcCache.PaddedAddresses[key] = padded

	return padded
}

// GetPaddedSlot returns the padded bytes for a slot constant
func GetPaddedSlot(slot int64) []byte {
	if padded, found := sfcCache.PaddedSlots[slot]; found {
		return padded
	}

	// If not pre-computed, compute and store it
	padded := common.LeftPadBytes(big.NewInt(slot).Bytes(), 32)
	sfcCache.PaddedSlots[slot] = padded

	return padded
}

// CreateHashInput creates a hash input from a validator ID and slot constant
func CreateHashInput(validatorID *big.Int, slotConstant int64) []byte {
	// Create a cache key from validatorID and slotConstant
	cacheKey := validatorID.String() + "_" + strconv.FormatInt(slotConstant, 10)

	// Check if the hash input is already cached
	if hashInput, found := sfcCache.HashInputs[cacheKey]; found {
		return hashInput
	}

	// If not in cache, create the hash input
	validatorIDBytes := GetPaddedValidatorID(validatorID)
	slotBytes := GetPaddedSlot(slotConstant)
	hashInput := append(validatorIDBytes, slotBytes...)

	// Store in cache
	sfcCache.HashInputs[cacheKey] = hashInput

	return hashInput
}

// CreateAddressHashInput creates a hash input from an address and slot constant
func CreateAddressHashInput(addr common.Address, slotConstant int64) []byte {
	// Create a cache key from address and slotConstant
	cacheKey := addr.String() + "_" + strconv.FormatInt(slotConstant, 10)

	// Check if the hash input is already cached
	if hashInput, found := sfcCache.AddressHashInputs[cacheKey]; found {
		return hashInput
	}

	// If not in cache, create the hash input
	addrBytes := GetPaddedAddress(addr)
	slotBytes := GetPaddedSlot(slotConstant)
	hashInput := append(addrBytes, slotBytes...)

	// Store in cache
	sfcCache.AddressHashInputs[cacheKey] = hashInput

	return hashInput
}

// CreateNestedHashInput creates a hash input from a validator ID and a hash
func CreateNestedHashInput(validatorID *big.Int, hash []byte) []byte {
	// Create a cache key from validatorID and hash
	cacheKey := validatorID.String() + "_" + common.Bytes2Hex(hash)

	// Check if the hash input is already cached
	if hashInput, found := sfcCache.HashInputs[cacheKey]; found {
		return hashInput
	}

	// If not in cache, create the hash input
	validatorIDBytes := GetPaddedValidatorID(validatorID)
	hashInput := append(validatorIDBytes, hash...)

	// Store in cache
	sfcCache.HashInputs[cacheKey] = hashInput

	return hashInput
}

// CreateNestedAddressHashInput creates a hash input from an address and a hash
func CreateNestedAddressHashInput(addr common.Address, hash []byte) []byte {
	// Create a cache key from address and hash
	cacheKey := addr.String() + "_" + common.Bytes2Hex(hash)

	// Check if the hash input is already cached
	if hashInput, found := sfcCache.AddressHashInputs[cacheKey]; found {
		return hashInput
	}

	// If not in cache, create the hash input
	addrBytes := GetPaddedAddress(addr)
	hashInput := append(addrBytes, hash...)

	// Store in cache
	sfcCache.AddressHashInputs[cacheKey] = hashInput

	return hashInput
}
