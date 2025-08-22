// Copyright 2024 The go-u2u Authors
// This file is part of the go-u2u library.
//
// The go-u2u library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-u2u library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-u2u library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/crypto"
)

func TestEIP7702Validator_ValidateSetCodeTransaction(t *testing.T) {
	chainID := big.NewInt(1)
	validator := NewEIP7702Validator(chainID)

	// Create valid authorization
	key, _ := crypto.GenerateKey()
	auth := &AuthorizationTuple{
		ChainID: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	tests := []struct {
		name    string
		tx      *SetCodeTx
		wantErr bool
	}{
		{
			name: "valid transaction",
			tx: &SetCodeTx{
				ChainID:           chainID,
				Nonce:             1,
				To:                nil,
				Value:             big.NewInt(0),
				Gas:               50000,
				GasFeeCap:         big.NewInt(1000000000),
				GasTipCap:         big.NewInt(1000000000),
				Data:              []byte{},
				AccessList:        AccessList{},
				AuthorizationList: AuthorizationList{*auth},
			},
			wantErr: false,
		},
		{
			name: "invalid chain ID",
			tx: &SetCodeTx{
				ChainID:           big.NewInt(2), // Wrong chain ID
				Nonce:             1,
				To:                nil,
				Value:             big.NewInt(0),
				Gas:               50000,
				GasFeeCap:         big.NewInt(1000000000),
				GasTipCap:         big.NewInt(1000000000),
				Data:              []byte{},
				AccessList:        AccessList{},
				AuthorizationList: AuthorizationList{*auth},
			},
			wantErr: true,
		},
		{
			name: "zero gas limit",
			tx: &SetCodeTx{
				ChainID:           chainID,
				Nonce:             1,
				To:                nil,
				Value:             big.NewInt(0),
				Gas:               0, // Invalid gas
				GasFeeCap:         big.NewInt(1000000000),
				GasTipCap:         big.NewInt(1000000000),
				Data:              []byte{},
				AccessList:        AccessList{},
				AuthorizationList: AuthorizationList{*auth},
			},
			wantErr: true,
		},
		{
			name: "gas fee cap less than tip cap",
			tx: &SetCodeTx{
				ChainID:           chainID,
				Nonce:             1,
				To:                nil,
				Value:             big.NewInt(0),
				Gas:               50000,
				GasFeeCap:         big.NewInt(500000000), // Less than tip cap
				GasTipCap:         big.NewInt(1000000000),
				Data:              []byte{},
				AccessList:        AccessList{},
				AuthorizationList: AuthorizationList{*auth},
			},
			wantErr: true,
		},
		{
			name: "insufficient gas for intrinsic cost",
			tx: &SetCodeTx{
				ChainID:           chainID,
				Nonce:             1,
				To:                nil,
				Value:             big.NewInt(0),
				Gas:               1000, // Too low
				GasFeeCap:         big.NewInt(1000000000),
				GasTipCap:         big.NewInt(1000000000),
				Data:              []byte{},
				AccessList:        AccessList{},
				AuthorizationList: AuthorizationList{*auth},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSetCodeTransaction(tt.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSetCodeTransaction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEIP7702Validator_ValidateAuthorizationList(t *testing.T) {
	chainID := big.NewInt(1)
	validator := NewEIP7702Validator(chainID)

	// Create valid authorizations
	key1, _ := crypto.GenerateKey()
	key2, _ := crypto.GenerateKey()

	auth1 := &AuthorizationTuple{
		ChainID: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Nonce:   42,
	}
	auth1.SignAuthorization(key1)

	auth2 := &AuthorizationTuple{
		ChainID: chainID,
		Address: common.HexToAddress("0x9876543210987654321098765432109876543210"),
		Nonce:   43,
	}
	auth2.SignAuthorization(key2)

	// Create duplicate authorization (same authority)
	authDuplicate := &AuthorizationTuple{
		ChainID: chainID,
		Address: common.HexToAddress("0x5555555555555555555555555555555555555555"),
		Nonce:   44,
	}
	authDuplicate.SignAuthorization(key1) // Same key as auth1

	tests := []struct {
		name     string
		authList AuthorizationList
		wantErr  bool
	}{
		{
			name:     "empty list",
			authList: AuthorizationList{},
			wantErr:  false,
		},
		{
			name:     "single authorization",
			authList: AuthorizationList{*auth1},
			wantErr:  false,
		},
		{
			name:     "multiple unique authorizations",
			authList: AuthorizationList{*auth1, *auth2},
			wantErr:  false,
		},
		{
			name:     "duplicate authorization",
			authList: AuthorizationList{*auth1, *authDuplicate},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateAuthorizationList(tt.authList)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAuthorizationList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEIP7702Validator_ValidateAuthorization(t *testing.T) {
	chainID := big.NewInt(1)
	validator := NewEIP7702Validator(chainID)

	// Create valid authorization
	key, _ := crypto.GenerateKey()
	validAuth := &AuthorizationTuple{
		ChainID: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Nonce:   42,
	}
	validAuth.SignAuthorization(key)

	tests := []struct {
		name    string
		auth    *AuthorizationTuple
		wantErr bool
	}{
		{
			name:    "valid authorization",
			auth:    validAuth,
			wantErr: false,
		},
		{
			name: "wrong chain ID",
			auth: &AuthorizationTuple{
				ChainID: big.NewInt(2), // Wrong chain ID
				Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
				Nonce:   42,
				V:       validAuth.V,
				R:       validAuth.R,
				S:       validAuth.S,
			},
			wantErr: true,
		},
		{
			name: "zero address",
			auth: &AuthorizationTuple{
				ChainID: chainID,
				Address: common.Address{}, // Zero address
				Nonce:   42,
				V:       validAuth.V,
				R:       validAuth.R,
				S:       validAuth.S,
			},
			wantErr: true,
		},
		{
			name: "zero R value",
			auth: &AuthorizationTuple{
				ChainID: chainID,
				Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
				Nonce:   42,
				V:       validAuth.V,
				R:       big.NewInt(0), // Zero R
				S:       validAuth.S,
			},
			wantErr: true,
		},
		{
			name: "zero S value",
			auth: &AuthorizationTuple{
				ChainID: chainID,
				Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
				Nonce:   42,
				V:       validAuth.V,
				R:       validAuth.R,
				S:       big.NewInt(0), // Zero S
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateAuthorization(tt.auth, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAuthorization() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEIP7702Validator_CalculateIntrinsicGas(t *testing.T) {
	chainID := big.NewInt(1)
	validator := NewEIP7702Validator(chainID)

	// Create authorization
	key, _ := crypto.GenerateKey()
	auth := &AuthorizationTuple{
		ChainID: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	tests := []struct {
		name        string
		tx          *SetCodeTx
		expectedGas uint64
	}{
		{
			name: "minimal transaction",
			tx: &SetCodeTx{
				ChainID:           chainID,
				Nonce:             1,
				To:                nil,
				Value:             big.NewInt(0),
				Gas:               50000,
				GasFeeCap:         big.NewInt(1000000000),
				GasTipCap:         big.NewInt(1000000000),
				Data:              []byte{},
				AccessList:        AccessList{},
				AuthorizationList: AuthorizationList{},
			},
			expectedGas: 21000, // Base cost only
		},
		{
			name: "transaction with authorization",
			tx: &SetCodeTx{
				ChainID:           chainID,
				Nonce:             1,
				To:                nil,
				Value:             big.NewInt(0),
				Gas:               50000,
				GasFeeCap:         big.NewInt(1000000000),
				GasTipCap:         big.NewInt(1000000000),
				Data:              []byte{},
				AccessList:        AccessList{},
				AuthorizationList: AuthorizationList{*auth},
			},
			expectedGas: 21000 + 6000, // Base + authorization cost
		},
		{
			name: "transaction with data",
			tx: &SetCodeTx{
				ChainID:           chainID,
				Nonce:             1,
				To:                nil,
				Value:             big.NewInt(0),
				Gas:               50000,
				GasFeeCap:         big.NewInt(1000000000),
				GasTipCap:         big.NewInt(1000000000),
				Data:              []byte{1, 2, 3, 4}, // 4 non-zero bytes
				AccessList:        AccessList{},
				AuthorizationList: AuthorizationList{},
			},
			expectedGas: 21000 + 4*16, // Base + data cost (non-zero bytes)
		},
		{
			name: "transaction with access list",
			tx: &SetCodeTx{
				ChainID:   chainID,
				Nonce:     1,
				To:        nil,
				Value:     big.NewInt(0),
				Gas:       50000,
				GasFeeCap: big.NewInt(1000000000),
				GasTipCap: big.NewInt(1000000000),
				Data:      []byte{},
				AccessList: AccessList{
					{Address: common.HexToAddress("0x1234"), StorageKeys: []common.Hash{common.HexToHash("0x5678")}},
				},
				AuthorizationList: AuthorizationList{},
			},
			expectedGas: 21000 + 2400 + 1900, // Base + access list address + storage key
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gas := validator.calculateIntrinsicGas(tt.tx)
			if gas != tt.expectedGas {
				t.Errorf("calculateIntrinsicGas() = %d, want %d", gas, tt.expectedGas)
			}
		})
	}
}

func TestValidateAuthorizationListSize(t *testing.T) {
	// Create large authorization list
	largeList := make(AuthorizationList, MaxAuthorizationListSize+1)
	for i := range largeList {
		largeList[i] = AuthorizationTuple{
			ChainID: big.NewInt(1),
			Address: common.BigToAddress(big.NewInt(int64(i + 1))),
			Nonce:   uint64(i),
		}
	}

	tests := []struct {
		name     string
		authList AuthorizationList
		wantErr  bool
	}{
		{
			name:     "empty list",
			authList: AuthorizationList{},
			wantErr:  false,
		},
		{
			name:     "maximum size list",
			authList: largeList[:MaxAuthorizationListSize],
			wantErr:  false,
		},
		{
			name:     "oversized list",
			authList: largeList,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAuthorizationListSize(tt.authList)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAuthorizationListSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDelegationDepth(t *testing.T) {
	tests := []struct {
		name    string
		depth   int
		wantErr bool
	}{
		{
			name:    "zero depth",
			depth:   0,
			wantErr: false,
		},
		{
			name:    "maximum depth",
			depth:   MaxDelegationDepth,
			wantErr: false,
		},
		{
			name:    "excessive depth",
			depth:   MaxDelegationDepth + 1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDelegationDepth(tt.depth)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDelegationDepth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidDelegationAddress(t *testing.T) {
	tests := []struct {
		name  string
		addr  common.Address
		valid bool
	}{
		{
			name:  "zero address",
			addr:  common.Address{},
			valid: false,
		},
		{
			name:  "system address 0x01",
			addr:  common.HexToAddress("0x0000000000000000000000000000000000000001"),
			valid: false,
		},
		{
			name:  "system address 0xff",
			addr:  common.HexToAddress("0x00000000000000000000000000000000000000ff"),
			valid: false,
		},
		{
			name:  "valid address",
			addr:  common.HexToAddress("0x1234567890123456789012345678901234567890"),
			valid: true,
		},
		{
			name:  "address 0x100",
			addr:  common.HexToAddress("0x0000000000000000000000000000000000000100"),
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := IsValidDelegationAddress(tt.addr)
			if valid != tt.valid {
				t.Errorf("IsValidDelegationAddress() = %v, want %v", valid, tt.valid)
			}
		})
	}
}

func TestGetDelegationStorageKey(t *testing.T) {
	authority := common.HexToAddress("0x1234567890123456789012345678901234567890")
	key := GetDelegationStorageKey(authority)

	// Check that the key is deterministic
	key2 := GetDelegationStorageKey(authority)
	if key != key2 {
		t.Errorf("GetDelegationStorageKey() not deterministic")
	}

	// Check that different addresses produce different keys
	authority2 := common.HexToAddress("0x9876543210987654321098765432109876543210")
	key3 := GetDelegationStorageKey(authority2)
	if key == key3 {
		t.Errorf("GetDelegationStorageKey() same key for different addresses")
	}
}

func TestGetDelegationStorageAddress(t *testing.T) {
	addr := GetDelegationStorageAddress()
	expected := common.HexToAddress(DelegationStorageAddress)
	if addr != expected {
		t.Errorf("GetDelegationStorageAddress() = %v, want %v", addr, expected)
	}
}

func TestEIP7702Compliance_CheckTransactionCompliance(t *testing.T) {
	chainID := big.NewInt(1)
	compliance := &EIP7702Compliance{}

	// Create valid authorization
	key, _ := crypto.GenerateKey()
	auth := &AuthorizationTuple{
		ChainID: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	validTx := &SetCodeTx{
		ChainID:           chainID,
		Nonce:             1,
		To:                nil,
		Value:             big.NewInt(0),
		Gas:               50000,
		GasFeeCap:         big.NewInt(1000000000),
		GasTipCap:         big.NewInt(1000000000),
		Data:              []byte{},
		AccessList:        AccessList{},
		AuthorizationList: AuthorizationList{*auth},
	}

	// Valid transaction should have no compliance errors
	errors := compliance.CheckTransactionCompliance(validTx, chainID)
	if len(errors) != 0 {
		t.Errorf("CheckTransactionCompliance() found %d errors for valid transaction: %v", len(errors), errors)
	}

	// Invalid transaction with system address delegation
	invalidAuth := &AuthorizationTuple{
		ChainID: chainID,
		Address: common.HexToAddress("0x0000000000000000000000000000000000000001"), // System address
		Nonce:   42,
	}
	invalidAuth.SignAuthorization(key)

	invalidTx := &SetCodeTx{
		ChainID:           chainID,
		Nonce:             1,
		To:                nil,
		Value:             big.NewInt(0),
		Gas:               50000,
		GasFeeCap:         big.NewInt(1000000000),
		GasTipCap:         big.NewInt(1000000000),
		Data:              []byte{},
		AccessList:        AccessList{},
		AuthorizationList: AuthorizationList{*invalidAuth},
	}

	errors = compliance.CheckTransactionCompliance(invalidTx, chainID)
	if len(errors) == 0 {
		t.Errorf("CheckTransactionCompliance() should find errors for invalid delegation address")
	}
}

func TestEIP7702Compliance_EstimateGas(t *testing.T) {
	chainID := big.NewInt(1)
	compliance := &EIP7702Compliance{}

	// Create authorization
	key, _ := crypto.GenerateKey()
	auth := &AuthorizationTuple{
		ChainID: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	tx := &SetCodeTx{
		ChainID:           chainID,
		Nonce:             1,
		To:                nil,
		Value:             big.NewInt(0),
		Gas:               50000,
		GasFeeCap:         big.NewInt(1000000000),
		GasTipCap:         big.NewInt(1000000000),
		Data:              []byte{},
		AccessList:        AccessList{},
		AuthorizationList: AuthorizationList{*auth},
	}

	estimatedGas := compliance.EstimateGas(tx)
	expectedGas := uint64(21000 + 6000) // Base + authorization cost

	if estimatedGas != expectedGas {
		t.Errorf("EstimateGas() = %d, want %d", estimatedGas, expectedGas)
	}
}

// Benchmark tests
func BenchmarkEIP7702Validator_ValidateSetCodeTransaction(b *testing.B) {
	chainID := big.NewInt(1)
	validator := NewEIP7702Validator(chainID)

	key, _ := crypto.GenerateKey()
	auth := &AuthorizationTuple{
		ChainID: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	tx := &SetCodeTx{
		ChainID:           chainID,
		Nonce:             1,
		To:                nil,
		Value:             big.NewInt(0),
		Gas:               50000,
		GasFeeCap:         big.NewInt(1000000000),
		GasTipCap:         big.NewInt(1000000000),
		Data:              []byte{},
		AccessList:        AccessList{},
		AuthorizationList: AuthorizationList{*auth},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateSetCodeTransaction(tx)
	}
}

func BenchmarkCalculateIntrinsicGas(b *testing.B) {
	chainID := big.NewInt(1)
	validator := NewEIP7702Validator(chainID)

	key, _ := crypto.GenerateKey()
	auth := &AuthorizationTuple{
		ChainID: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	tx := &SetCodeTx{
		ChainID:           chainID,
		Nonce:             1,
		To:                nil,
		Value:             big.NewInt(0),
		Gas:               50000,
		GasFeeCap:         big.NewInt(1000000000),
		GasTipCap:         big.NewInt(1000000000),
		Data:              make([]byte, 1000), // Large data
		AccessList:        AccessList{},
		AuthorizationList: AuthorizationList{*auth},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.calculateIntrinsicGas(tx)
	}
}