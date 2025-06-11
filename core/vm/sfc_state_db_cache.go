package vm

import (
	"log"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
)

var (
	cache        map[common.Address]map[common.Hash]common.Hash
	needFlushMap map[common.Address]map[common.Hash]struct{}
	currentEpoch *big.Int
)

type SfcStateDBCache struct {
	StateDB
}

func NewSfcStateDBCache(originalStateDB StateDB) *SfcStateDBCache {
	if cache == nil {
		cache = make(map[common.Address]map[common.Hash]common.Hash)
	}
	if needFlushMap == nil {
		needFlushMap = make(map[common.Address]map[common.Hash]struct{})
	}
	if currentEpoch == nil {
		currentEpoch = big.NewInt(0)
	}

	return &SfcStateDBCache{
		StateDB: originalStateDB,
	}
}

func (s *SfcStateDBCache) GetState(addr common.Address, hash common.Hash) common.Hash {
	// Check if we need to flush due to epoch change
	s.CheckAndFlushEpoch()

	if value, ok := cache[addr][hash]; ok {
		// Cache hit
		return value
	}

	// Get state from original state DB
	value := s.StateDB.GetState(addr, hash)
	s.SetState(addr, hash, value)
	return value
}

func (s *SfcStateDBCache) SetState(addr common.Address, hash common.Hash, value common.Hash) {
	if cache[addr] == nil {
		cache[addr] = make(map[common.Hash]common.Hash)
	}

	if needFlushMap[addr] == nil {
		needFlushMap[addr] = make(map[common.Hash]struct{})
	}

	// Only mark for flush if the value actually changed
	if existing, exists := cache[addr][hash]; !exists || existing.Cmp(value) != 0 {
		cache[addr][hash] = value
		needFlushMap[addr][hash] = struct{}{}
	}
}

func (s *SfcStateDBCache) Flush() {
	for addr, hashMap := range needFlushMap {
		for hash := range hashMap {
			if value, ok := cache[addr][hash]; ok {
				s.StateDB.SetState(addr, hash, value)
			}
		}
		delete(needFlushMap, addr)
	}

	// Log the size of the cache (in bytes) in other goroutine
	go func() {
		cacheSize := 0
		for _, hashMap := range cache {
			for _, value := range hashMap {
				cacheSize += len(value.Bytes())
			}
		}
		log.Printf("Cache size: %d bytes, current epoch: %d\n", cacheSize, currentEpoch)
	}()
}

func (s *SfcStateDBCache) CheckAndFlushEpoch() {
	currentSealedEpochSlot := int64(1 + 0x66)
	contractAddress := common.HexToAddress("0xfc00face00000000000000000000000000000000")
	// Get current epoch from cache
	newEpoch, ok := cache[contractAddress][common.BigToHash(big.NewInt(currentSealedEpochSlot))]
	if !ok {
		return
	}

	// If epoch changed, flush the cache
	if currentEpoch == nil || currentEpoch.Cmp(newEpoch.Big()) != 0 {
		s.Flush()
		currentEpoch = newEpoch.Big()
	}
}
