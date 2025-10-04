// Package sfc implements the SFC (Special Fee Contract) precompiled contract.
package sfc

import (
	"math/big"
	"sync"

	"github.com/unicornultrafoundation/go-u2u/common"
)

// BigIntPool is a pool of reusable big.Int objects
var BigIntPool = sync.Pool{
	New: func() interface{} {
		return new(big.Int)
	},
}

// ByteSlicePool is a pool of reusable byte slices
var ByteSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 32) // Pre-allocate capacity for common use case (32 bytes)
	},
}

// GetBigInt gets a big.Int from the pool and initializes it to zero
func GetBigInt() *big.Int {
	return BigIntPool.Get().(*big.Int).SetInt64(0)
}

// PutBigInt returns a big.Int to the pool
func PutBigInt(b *big.Int) {
	if b != nil {
		BigIntPool.Put(b)
	}
}

// GetByteSlice gets a byte slice from the pool
func GetByteSlice() []byte {
	return ByteSlicePool.Get().([]byte)[:0] // Reset length but keep capacity
}

// PutByteSlice returns a byte slice to the pool
func PutByteSlice(b []byte) {
	if cap(b) >= 32 {
		ByteSlicePool.Put(b[:0]) // Reset length but keep capacity
	}
}

// GetPaddedBytes gets a byte slice from the pool and pads it
func GetPaddedBytes(data []byte, length int) []byte {
	result := GetByteSlice()
	if cap(result) < length {
		// If the slice from the pool is too small, allocate a new one
		result = make([]byte, length)
	} else {
		// Otherwise, resize the slice from the pool
		result = result[:length]
	}

	// Perform the padding (right-aligned)
	copy(result[length-len(data):], data)
	return result
}

// GetLeftPaddedBytes gets a byte slice from the pool and left-pads it
// This is a replacement for common.LeftPadBytes that uses the pool
func GetLeftPaddedBytes(data []byte, length int) []byte {
	result := GetPaddedBytes(data, length)
	return result
}

// BigIntToBytes converts a big.Int to a byte slice using the pool
func BigIntToBytes(b *big.Int) []byte {
	if b == nil {
		return GetByteSlice()
	}
	return GetLeftPaddedBytes(b.Bytes(), 32)
}

// AddressToPaddedBytes converts an address to a padded byte slice using the pool
func AddressToPaddedBytes(addr common.Address) []byte {
	return GetLeftPaddedBytes(addr.Bytes(), 32)
}


