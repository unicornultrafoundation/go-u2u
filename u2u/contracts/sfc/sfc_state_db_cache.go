package sfc

import (
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
)

var (
	cache        map[common.Address]map[common.Hash]common.Hash
	needFlushMap map[common.Address]map[common.Hash]struct{}
)

type SfcStateDBCache struct {
	vm.StateDB
}

func NewSfcStateDBCache(originalStateDB vm.StateDB) *SfcStateDBCache {
	if cache == nil {
		cache = make(map[common.Address]map[common.Hash]common.Hash)
	}
	if needFlushMap == nil {
		needFlushMap = make(map[common.Address]map[common.Hash]struct{})
	}

	return &SfcStateDBCache{
		StateDB: originalStateDB,
	}
}

func (s *SfcStateDBCache) GetState(addr common.Address, hash common.Hash) common.Hash {
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
}
