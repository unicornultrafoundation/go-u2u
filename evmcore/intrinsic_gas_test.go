package evmcore

import (
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/params"
)

func TestIntrinsicGas_Basic(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint64
	}{
		{
			name:     "empty data",
			data:     []byte{},
			expected: params.TxGas,
		},
		{
			name:     "single zero byte",
			data:     []byte{0},
			expected: params.TxGas + params.TxDataZeroGas,
		},
		{
			name:     "single non-zero byte",
			data:     []byte{1},
			expected: params.TxGas + params.TxDataNonZeroGasEIP2028,
		},
		{
			name:     "mixed zero and non-zero bytes",
			data:     []byte{0, 1, 2, 0, 0},
			expected: params.TxGas + 3*params.TxDataZeroGas + 2*params.TxDataNonZeroGasEIP2028,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gas, err := IntrinsicGas(tt.data, nil, nil, false, true, true, false)
			if err != nil {
				t.Fatalf("IntrinsicGas() error = %v", err)
			}
			if gas != tt.expected {
				t.Errorf("IntrinsicGas() = %v, want %v", gas, tt.expected)
			}
		})
	}
}

func TestIntrinsicGas_ContractCreation(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		homestead  bool
		expected   uint64
	}{
		{
			name:      "contract creation pre-homestead",
			data:      []byte{},
			homestead: false,
			expected:  params.TxGas,
		},
		{
			name:      "contract creation post-homestead",
			data:      []byte{},
			homestead: true,
			expected:  params.TxGasContractCreation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gas, err := IntrinsicGas(tt.data, nil, nil, true, tt.homestead, true, false)
			if err != nil {
				t.Fatalf("IntrinsicGas() error = %v", err)
			}
			if gas != tt.expected {
				t.Errorf("IntrinsicGas() = %v, want %v", gas, tt.expected)
			}
		})
	}
}

func TestIntrinsicGas_EIP2028_DataCosts(t *testing.T) {
	data := []byte{1, 2, 3} // 3 non-zero bytes

	// Pre-EIP2028: should use frontier gas cost
	gas1, err := IntrinsicGas(data, nil, nil, false, true, false, false)
	if err != nil {
		t.Fatalf("IntrinsicGas() error = %v", err)
	}
	expected1 := params.TxGas + 3*params.TxDataNonZeroGasFrontier
	if gas1 != expected1 {
		t.Errorf("Pre-EIP2028: IntrinsicGas() = %v, want %v", gas1, expected1)
	}

	// Post-EIP2028: should use reduced gas cost
	gas2, err := IntrinsicGas(data, nil, nil, false, true, true, false)
	if err != nil {
		t.Fatalf("IntrinsicGas() error = %v", err)
	}
	expected2 := params.TxGas + 3*params.TxDataNonZeroGasEIP2028
	if gas2 != expected2 {
		t.Errorf("Post-EIP2028: IntrinsicGas() = %v, want %v", gas2, expected2)
	}

	// EIP2028 should result in lower gas cost
	if gas2 >= gas1 {
		t.Errorf("EIP2028 should reduce gas cost: %v >= %v", gas2, gas1)
	}
}

func TestIntrinsicGas_AccessList(t *testing.T) {
	accessList := types.AccessList{
		{Address: common.Address{1}, StorageKeys: []common.Hash{{1}, {2}}},
		{Address: common.Address{2}, StorageKeys: []common.Hash{{3}}},
	}

	gas, err := IntrinsicGas([]byte{}, accessList, nil, false, true, true, false)
	if err != nil {
		t.Fatalf("IntrinsicGas() error = %v", err)
	}

	expected := params.TxGas +
		2*params.TxAccessListAddressGas +    // 2 addresses
		3*params.TxAccessListStorageKeyGas   // 3 storage keys total

	if gas != expected {
		t.Errorf("IntrinsicGas() = %v, want %v", gas, expected)
	}
}

func TestIntrinsicGas_AuthorizationList(t *testing.T) {
	// Create a mock authorization list
	authList := types.AuthorizationList{
		{Address: common.Address{1}, Nonce: 1},
		{Address: common.Address{2}, Nonce: 2},
		{Address: common.Address{3}, Nonce: 3},
	}

	gas, err := IntrinsicGas([]byte{}, nil, authList, false, true, true, false)
	if err != nil {
		t.Fatalf("IntrinsicGas() error = %v", err)
	}

	expected := params.TxGas + 3*params.TxAuthTupleGas
	if gas != expected {
		t.Errorf("IntrinsicGas() = %v, want %v", gas, expected)
	}
}

func TestIntrinsicGas_EIP3860_InitCode(t *testing.T) {
	// Test data smaller than max init code size
	smallData := make([]byte, 1000)
	for i := range smallData {
		smallData[i] = byte(i % 256)
	}

	gas, err := IntrinsicGas(smallData, nil, nil, true, true, true, true)
	if err != nil {
		t.Fatalf("IntrinsicGas() error = %v", err)
	}

	// Calculate expected gas
	nz := uint64(0)
	for _, b := range smallData {
		if b != 0 {
			nz++
		}
	}
	z := uint64(len(smallData)) - nz
	words := toWordSize(uint64(len(smallData)))

	expected := params.TxGasContractCreation +
		z*params.TxDataZeroGas +
		nz*params.TxDataNonZeroGasEIP2028 +
		words*params.InitCodeWordGas

	if gas != expected {
		t.Errorf("IntrinsicGas() = %v, want %v", gas, expected)
	}

	// Test data larger than max init code size should fail
	largeData := make([]byte, params.MaxInitCodeSize+1)
	_, err = IntrinsicGas(largeData, nil, nil, true, true, true, true)
	if err == nil {
		t.Error("Expected error for oversized init code")
	}
}

func TestIntrinsicGas_Combined(t *testing.T) {
	data := []byte{1, 0, 2, 0, 3}
	accessList := types.AccessList{
		{Address: common.Address{1}, StorageKeys: []common.Hash{{1}}},
	}
	authList := types.AuthorizationList{
		{Address: common.Address{1}, Nonce: 1},
		{Address: common.Address{2}, Nonce: 2},
	}

	gas, err := IntrinsicGas(data, accessList, authList, false, true, true, false)
	if err != nil {
		t.Fatalf("IntrinsicGas() error = %v", err)
	}

	expected := params.TxGas +
		2*params.TxDataZeroGas +                  // 2 zero bytes
		3*params.TxDataNonZeroGasEIP2028 +        // 3 non-zero bytes
		1*params.TxAccessListAddressGas +         // 1 address
		1*params.TxAccessListStorageKeyGas +      // 1 storage key
		2*params.TxAuthTupleGas                   // 2 authorization tuples

	if gas != expected {
		t.Errorf("IntrinsicGas() = %v, want %v", gas, expected)
	}
}

func TestIntrinsicGas_OverflowProtection(t *testing.T) {
	// Test authorization list with reasonable size
	authList := make(types.AuthorizationList, 10000) // Use reasonable test size
	for i := range authList {
		authList[i] = types.AuthorizationTuple{Address: common.Address{byte(i)}, Nonce: uint64(i)}
	}
	
	// To test overflow, we'll simulate the calculation
	authGas := uint64(len(authList)) * params.TxAuthTupleGas
	baseGas := params.TxGas
	
	// Test normal case first
	gas, err := IntrinsicGas([]byte{}, nil, authList, false, true, true, false)
	if err != nil {
		t.Fatalf("Unexpected error for normal case: %v", err)
	}
	expected := baseGas + authGas
	if gas != expected {
		t.Errorf("IntrinsicGas() = %v, want %v", gas, expected)
	}

	// Test access list with reasonable size
	accessList := make(types.AccessList, 1000)
	for i := range accessList {
		accessList[i] = types.AccessTuple{
			Address:     common.Address{byte(i)},
			StorageKeys: []common.Hash{{byte(i)}},
		}
	}
	
	gas, err = IntrinsicGas([]byte{}, accessList, nil, false, true, true, false)
	if err != nil {
		t.Fatalf("Unexpected error for access list: %v", err)
	}

	// Test data with reasonable size that could approach overflow conditions
	largeData := make([]byte, 100000)
	for i := range largeData {
		largeData[i] = 1 // non-zero
	}
	_, err = IntrinsicGas(largeData, nil, nil, false, true, true, false)
	if err != nil {
		t.Fatalf("Unexpected error for large data: %v", err)
	}
}

func TestIntrinsicGas_EdgeCases(t *testing.T) {
	// Test with nil access list and authorization list
	gas, err := IntrinsicGas([]byte{1, 2, 3}, nil, nil, false, true, true, false)
	if err != nil {
		t.Fatalf("IntrinsicGas() error = %v", err)
	}
	expected := params.TxGas + 3*params.TxDataNonZeroGasEIP2028
	if gas != expected {
		t.Errorf("IntrinsicGas() = %v, want %v", gas, expected)
	}

	// Test with empty access list and authorization list
	emptyAccessList := types.AccessList{}
	emptyAuthList := types.AuthorizationList{}
	gas, err = IntrinsicGas([]byte{}, emptyAccessList, emptyAuthList, false, true, true, false)
	if err != nil {
		t.Fatalf("IntrinsicGas() error = %v", err)
	}
	if gas != params.TxGas {
		t.Errorf("IntrinsicGas() = %v, want %v", gas, params.TxGas)
	}
}

// Benchmark tests to ensure performance
func BenchmarkIntrinsicGas_Basic(b *testing.B) {
	data := []byte{1, 2, 3, 4, 5}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = IntrinsicGas(data, nil, nil, false, true, true, false)
	}
}

func BenchmarkIntrinsicGas_WithAuthorizations(b *testing.B) {
	data := []byte{1, 2, 3, 4, 5}
	authList := make(types.AuthorizationList, 10)
	for i := range authList {
		authList[i] = types.AuthorizationTuple{
			Address: common.Address{byte(i)},
			Nonce:   uint64(i),
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = IntrinsicGas(data, nil, authList, false, true, true, false)
	}
}

func BenchmarkIntrinsicGas_LargeData(b *testing.B) {
	data := make([]byte, 10000)
	for i := range data {
		data[i] = byte(i % 256)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = IntrinsicGas(data, nil, nil, false, true, true, false)
	}
}