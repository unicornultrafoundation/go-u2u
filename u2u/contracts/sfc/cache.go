// Package sfc implements the SFC (Special Fee Contract) precompiled contract.
package sfc

import (
	"fmt"
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

	// Hash input caches
	HashInputs      map[string][]byte // Cache for hash inputs (validatorID + slot)
	AddressHashInputs map[string][]byte // Cache for address hash inputs (address + slot)
	NestedHashInputs map[string][]byte // Cache for nested hash inputs

	// Unified ABI encoding cache
	AbiPackCache map[string][]byte // Cache for all ABI packed results
}

// NewSFCCache creates a new SFC cache
func NewSFCCache() *SFCCache {
	cache := &SFCCache{
		Hash:              NewHashCache(),
		Slot:              NewSlotCache(),
		ValidatorSlot:     make(map[string]*big.Int),
		EpochSlot:         make(map[string]*big.Int),
		HashInputs:         make(map[string][]byte),
		AddressHashInputs:  make(map[string][]byte),
		NestedHashInputs:   make(map[string][]byte),
		AbiPackCache:       make(map[string][]byte),
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



// CreateHashInput creates a hash input from a validator ID and slot constant
func CreateHashInput(validatorID *big.Int, slotConstant int64) []byte {
	// Create a cache key from validatorID and slotConstant
	cacheKey := validatorID.String() + "_" + strconv.FormatInt(slotConstant, 10)

	// Check if the hash input is already cached
	if hashInput, found := sfcCache.HashInputs[cacheKey]; found {
		return hashInput
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
	sfcCache.AddressHashInputs[cacheKey] = hashInput

	return hashInput
}

// CreateNestedHashInput creates a hash input from a validator ID and a hash
func CreateNestedHashInput(validatorID *big.Int, hash []byte) []byte {
	// Create a cache key from validatorID and hash
	cacheKey := validatorID.String() + "_" + common.Bytes2Hex(hash)

	// Check if the hash input is already cached
	if hashInput, found := sfcCache.NestedHashInputs[cacheKey]; found {
		return hashInput
	}

	// If not in cache, create the hash input directly
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

	// Store in cache
	sfcCache.NestedHashInputs[cacheKey] = hashInput

	return hashInput
}

// CreateNestedAddressHashInput creates a hash input from an address and a hash
func CreateNestedAddressHashInput(addr common.Address, hash []byte) []byte {
	// Create a cache key from address and hash
	cacheKey := addr.String() + "_" + common.Bytes2Hex(hash)

	// Check if the hash input is already cached
	if hashInput, found := sfcCache.NestedHashInputs[cacheKey]; found {
		return hashInput
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
	sfcCache.NestedHashInputs[cacheKey] = hashInput

	return hashInput
}

// CreateValidatorMappingHashInput creates a hash input from a validator ID and a mapping slot
func CreateValidatorMappingHashInput(validatorID *big.Int, mappingSlot *big.Int) []byte {
	// Create a cache key from validatorID and mappingSlot
	cacheKey := validatorID.String() + "_mapping_" + mappingSlot.String()

	// Check if the hash input is already cached
	if hashInput, found := sfcCache.HashInputs[cacheKey]; found {
		return hashInput
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
	sfcCache.HashInputs[cacheKey] = hashInput

	return hashInput
}

// CreateAddressMethodHashInput creates a hash input from an address and a method ID
func CreateAddressMethodHashInput(addr common.Address, methodID []byte) []byte {
	// Create a cache key from address and methodID
	cacheKey := addr.String() + "_method_" + common.Bytes2Hex(methodID)

	// Check if the hash input is already cached
	if hashInput, found := sfcCache.AddressHashInputs[cacheKey]; found {
		return hashInput
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
	sfcCache.AddressHashInputs[cacheKey] = hashInput

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
	if hashInput, found := sfcCache.AddressHashInputs[cacheKey]; found {
		return hashInput
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
	sfcCache.AddressHashInputs[cacheKey] = hashInput

	return hashInput
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
		if packed, ok := sfcCache.AbiPackCache[key]; ok {
			return packed, nil
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
		sfcCache.AbiPackCache[key] = packed

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
