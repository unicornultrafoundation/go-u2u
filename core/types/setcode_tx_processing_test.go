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
	"fmt"
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/crypto"
)

func TestDelegationChainResolver(t *testing.T) {
	resolver := NewDelegationChainResolver()

	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	addr3 := common.HexToAddress("0x3333333333333333333333333333333333333333")

	tests := []struct {
		name           string
		startAddr      common.Address
		delegationMap  map[common.Address]*common.Address
		expectedAddr   common.Address
		expectedError  bool
	}{
		{
			name:           "no delegation",
			startAddr:      addr1,
			delegationMap:  map[common.Address]*common.Address{},
			expectedAddr:   addr1,
			expectedError:  false,
		},
		{
			name:      "single delegation",
			startAddr: addr1,
			delegationMap: map[common.Address]*common.Address{
				addr1: &addr2,
			},
			expectedAddr:  addr2,
			expectedError: false,
		},
		{
			name:      "chain delegation",
			startAddr: addr1,
			delegationMap: map[common.Address]*common.Address{
				addr1: &addr2,
				addr2: &addr3,
			},
			expectedAddr:  addr3,
			expectedError: false,
		},
		{
			name:      "circular delegation",
			startAddr: addr1,
			delegationMap: map[common.Address]*common.Address{
				addr1: &addr2,
				addr2: &addr1,
			},
			expectedAddr:  common.Address{},
			expectedError: true,
		},
		{
			name:      "self delegation",
			startAddr: addr1,
			delegationMap: map[common.Address]*common.Address{
				addr1: &addr1,
			},
			expectedAddr:  common.Address{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getDelegation := func(addr common.Address) *common.Address {
				return tt.delegationMap[addr]
			}

			result, err := resolver.ResolveDelegationChain(tt.startAddr, getDelegation)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expectedAddr {
					t.Errorf("expected address %v, got %v", tt.expectedAddr, result)
				}
			}
		})
	}
}

func TestDelegationChainMaxDepth(t *testing.T) {
	resolver := NewDelegationChainResolver()

	// Create a long chain
	addresses := make([]common.Address, 20)
	delegationMap := make(map[common.Address]*common.Address)
	
	for i := 0; i < 20; i++ {
		addresses[i] = common.HexToAddress(fmt.Sprintf("0x%040d", i+1))
		if i < 19 {
			delegationMap[addresses[i]] = &addresses[i+1]
		}
	}

	getDelegation := func(addr common.Address) *common.Address {
		return delegationMap[addr]
	}

	// Should hit max depth limit
	_, err := resolver.ResolveDelegationChain(addresses[0], getDelegation)
	if err == nil {
		t.Errorf("expected max depth error but got none")
	}
}

func TestAuthorizationProcessor(t *testing.T) {
	chainID := big.NewInt(1)
	processor := NewAuthorizationProcessor(chainID)

	// Create test keys
	key1, _ := crypto.GenerateKey()
	key2, _ := crypto.GenerateKey()
	addr1 := crypto.PubkeyToAddress(key1.PublicKey)
	addr2 := crypto.PubkeyToAddress(key2.PublicKey)
	
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Create signed authorizations
	auth1 := &AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   42,
	}
	auth1.SignAuthorization(key1)

	auth2 := &AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   43,
	}
	auth2.SignAuthorization(key2)

	authList := AuthorizationList{*auth1, *auth2}

	getNonce := func(addr common.Address) uint64 {
		if addr == addr1 {
			return 42
		}
		if addr == addr2 {
			return 43
		}
		return 0
	}

	delegations, err := processor.ProcessAuthorizationList(authList, big.NewInt(1000), getNonce)
	if err != nil {
		t.Fatalf("ProcessAuthorizationList() error = %v", err)
	}

	// Should have 2 delegations
	if len(delegations) != 2 {
		t.Errorf("expected 2 delegations, got %d", len(delegations))
	}

	// Check delegation mappings
	if delegations[addr1] != codeAddr {
		t.Errorf("expected delegation from %v to %v, got %v", addr1, codeAddr, delegations[addr1])
	}
	if delegations[addr2] != codeAddr {
		t.Errorf("expected delegation from %v to %v, got %v", addr2, codeAddr, delegations[addr2])
	}
}

func TestAuthorizationProcessorInvalidSignature(t *testing.T) {
	chainID := big.NewInt(1)
	processor := NewAuthorizationProcessor(chainID)

	// Create authorization with invalid signature
	auth := &AuthorizationTuple{
		ChainID: big.NewInt(1),
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Nonce:   42,
		V:       big.NewInt(27),
		R:       big.NewInt(1),
		S:       big.NewInt(1),
	}

	authList := AuthorizationList{*auth}

	getNonce := func(addr common.Address) uint64 {
		return 42
	}

	_, err := processor.ProcessAuthorizationList(authList, big.NewInt(1000), getNonce)
	if err == nil {
		t.Errorf("expected error for invalid signature but got none")
	}
}

func TestAuthorizationProcessorNonceMismatch(t *testing.T) {
	chainID := big.NewInt(1)
	processor := NewAuthorizationProcessor(chainID)

	key, _ := crypto.GenerateKey()
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Create authorization with nonce 42
	auth := &AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	authList := AuthorizationList{*auth}

	// But account nonce is 43
	getNonce := func(addr common.Address) uint64 {
		return 43
	}

	_, err := processor.ProcessAuthorizationList(authList, big.NewInt(1000), getNonce)
	if err == nil {
		t.Errorf("expected error for nonce mismatch but got none")
	}
}

func TestSetCodeTxProcessor(t *testing.T) {
	chainID := big.NewInt(1)
	processor := NewSetCodeTxProcessor(chainID)

	// Create test transaction
	key, _ := crypto.GenerateKey()
	auth := &AuthorizationTuple{
		ChainID: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	setCodeTx := &SetCodeTx{
		ChainID:           chainID,
		Nonce:             1,
		To:                nil,
		Value:             big.NewInt(0),
		Gas:               21000,
		GasFeeCap:         big.NewInt(1000000000),
		GasTipCap:         big.NewInt(1000000000),
		Data:              []byte{},
		AccessList:        AccessList{},
		AuthorizationList: AuthorizationList{*auth},
	}

	err := processor.ValidateForPool(setCodeTx, big.NewInt(1000))
	if err != nil {
		t.Errorf("ValidateForPool() error = %v", err)
	}
}

func TestSetCodeTxProcessorGasValidation(t *testing.T) {
	chainID := big.NewInt(1)
	processor := NewSetCodeTxProcessor(chainID)

	// Create transaction with insufficient gas
	setCodeTx := &SetCodeTx{
		ChainID:           chainID,
		Nonce:             1,
		To:                nil,
		Value:             big.NewInt(0),
		Gas:               1000, // Too low
		GasFeeCap:         big.NewInt(1000000000),
		GasTipCap:         big.NewInt(1000000000),
		Data:              []byte{},
		AccessList:        AccessList{},
		AuthorizationList: AuthorizationList{},
	}

	err := processor.ValidateForPool(setCodeTx, big.NewInt(1000))
	if err == nil {
		t.Errorf("expected gas validation error but got none")
	}
}

func TestSetCodeTxGasCalculator(t *testing.T) {
	chainID := big.NewInt(1)
	calculator := NewSetCodeTxGasCalculator(chainID)

	// Create test authorization
	key, _ := crypto.GenerateKey()
	auth := &AuthorizationTuple{
		ChainID: chainID,
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	setCodeTx := &SetCodeTx{
		ChainID:           chainID,
		Nonce:             1,
		To:                nil,
		Value:             big.NewInt(0),
		Gas:               21000,
		GasFeeCap:         big.NewInt(1000000000),
		GasTipCap:         big.NewInt(1000000000),
		Data:              []byte("test data"),
		AccessList:        AccessList{},
		AuthorizationList: AuthorizationList{*auth},
	}

	intrinsicGas, err := calculator.IntrinsicGas(setCodeTx)
	if err != nil {
		t.Errorf("IntrinsicGas() error = %v", err)
	}

	// Should include base cost + data cost + authorization cost
	expectedMinGas := uint64(21000) + // Base transaction cost
		uint64(len(setCodeTx.Data)*16) + // Data cost
		uint64(len(setCodeTx.AuthorizationList)*6000) // Authorization cost (authorizationBaseGas + authorizationSignatureGas)

	if intrinsicGas < expectedMinGas {
		t.Errorf("intrinsic gas %d is less than expected minimum %d", intrinsicGas, expectedMinGas)
	}
}

// Benchmark tests
func BenchmarkDelegationChainResolution(b *testing.B) {
	resolver := NewDelegationChainResolver()
	
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	addr3 := common.HexToAddress("0x3333333333333333333333333333333333333333")
	
	delegationMap := map[common.Address]*common.Address{
		addr1: &addr2,
		addr2: &addr3,
	}
	
	getDelegation := func(addr common.Address) *common.Address {
		return delegationMap[addr]
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.ResolveDelegationChain(addr1, getDelegation)
	}
}

func BenchmarkAuthorizationProcessing(b *testing.B) {
	chainID := big.NewInt(1)
	processor := NewAuthorizationProcessor(chainID)

	key, _ := crypto.GenerateKey()
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	auth := &AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	authList := AuthorizationList{*auth}

	getNonce := func(addr common.Address) uint64 {
		return 42
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.ProcessAuthorizationList(authList, big.NewInt(1000), getNonce)
	}
}