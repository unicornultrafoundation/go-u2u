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

package tests

import (
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/params"
)

// TestEIP7702_ProtocolParametersValidation tests that protocol parameters are correctly defined
func TestEIP7702_ProtocolParametersValidation(t *testing.T) {
	// Test that EIP-7702 constants are defined correctly according to specification
	
	// Check maximum authorization list size
	expectedMaxSize := 256
	actualMaxSize := types.MaxAuthorizationListSize
	
	if actualMaxSize != expectedMaxSize {
		t.Errorf("MaxAuthorizationListSize mismatch: expected %d, got %d", 
			expectedMaxSize, actualMaxSize)
	}
	
	// Check delegation prefix
	expectedPrefix := []byte{0xef, 0x01, 0x00}
	actualPrefix := types.DelegationPrefix
	
	if len(actualPrefix) != 3 {
		t.Errorf("DelegationPrefix should be 3 bytes, got %d", len(actualPrefix))
	}
	
	for i, expected := range expectedPrefix {
		if actualPrefix[i] != expected {
			t.Errorf("DelegationPrefix[%d] mismatch: expected 0x%02x, got 0x%02x", 
				i, expected, actualPrefix[i])
		}
	}
	
	// Check delegation depth limit
	expectedMaxDepth := 16
	actualMaxDepth := types.MaxDelegationDepth
	
	if actualMaxDepth != expectedMaxDepth {
		t.Errorf("MaxDelegationDepth mismatch: expected %d, got %d", 
			expectedMaxDepth, actualMaxDepth)
	}
}

// TestEIP7702_ChainConfigValidation tests chain configuration for EIP-7702 support
func TestEIP7702_ChainConfigValidation(t *testing.T) {
	testCases := []struct {
		name          string
		config        *params.ChainConfig
		blockNumber   *big.Int
		expectActive  bool
	}{
		{
			name: "Phaethon at genesis",
			config: &params.ChainConfig{
				ChainID:       big.NewInt(1),
				PhaethonBlock: big.NewInt(0),
			},
			blockNumber:  big.NewInt(0),
			expectActive: true,
		},
		{
			name: "Phaethon disabled (nil)",
			config: &params.ChainConfig{
				ChainID:       big.NewInt(1),
				PhaethonBlock: nil,
			},
			blockNumber:  big.NewInt(100),
			expectActive: false,
		},
		{
			name: "Phaethon far in future",
			config: &params.ChainConfig{
				ChainID:       big.NewInt(1),
				PhaethonBlock: big.NewInt(1000000),
			},
			blockNumber:  big.NewInt(100),
			expectActive: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isActive := tc.config.IsPhaethon(tc.blockNumber)
			
			if isActive != tc.expectActive {
				t.Errorf("IsPhaethon(%d) = %v, expected %v", 
					tc.blockNumber, isActive, tc.expectActive)
			}
		})
	}
}

// TestEIP7702_ForkOrderingValidation tests that Phaethon hardfork is properly ordered
func TestEIP7702_ForkOrderingValidation(t *testing.T) {
	// Test various fork ordering scenarios
	
	testCases := []struct {
		name           string
		config         *params.ChainConfig
		expectValid    bool
		expectedError  string
	}{
		{
			name: "Proper fork ordering",
			config: &params.ChainConfig{
				ChainID:             big.NewInt(1),
				HomesteadBlock:      big.NewInt(0),
				EIP150Block:         big.NewInt(0),
				EIP155Block:         big.NewInt(0),
				EIP158Block:         big.NewInt(0),
				ByzantiumBlock:      big.NewInt(0),
				ConstantinopleBlock: big.NewInt(0),
				PetersburgBlock:     big.NewInt(0),
				IstanbulBlock:       big.NewInt(0),
				BerlinBlock:         big.NewInt(0),
				LondonBlock:         big.NewInt(0),
				ClymeneBlock:        big.NewInt(10),
				PhaethonBlock:       big.NewInt(20),
			},
			expectValid: true,
		},
		{
			name: "Phaethon before Clymene",
			config: &params.ChainConfig{
				ChainID:             big.NewInt(1),
				HomesteadBlock:      big.NewInt(0),
				EIP150Block:         big.NewInt(0),
				EIP155Block:         big.NewInt(0),
				EIP158Block:         big.NewInt(0),
				ByzantiumBlock:      big.NewInt(0),
				ConstantinopleBlock: big.NewInt(0),
				PetersburgBlock:     big.NewInt(0),
				IstanbulBlock:       big.NewInt(0),
				BerlinBlock:         big.NewInt(0),
				LondonBlock:         big.NewInt(0),
				ClymeneBlock:        big.NewInt(20),
				PhaethonBlock:       big.NewInt(10), // Before Clymene - should cause error
			},
			expectValid:   false,
			expectedError: "fork ordering",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.CheckConfigForkOrder()
			
			if tc.expectValid && err != nil {
				t.Errorf("Expected valid fork ordering, got error: %v", err)
			}
			
			if !tc.expectValid && err == nil {
				t.Errorf("Expected fork ordering error, but validation passed")
			}
			
			if !tc.expectValid && tc.expectedError != "" && err != nil {
				if !contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", 
						tc.expectedError, err.Error())
				}
			}
		})
	}
}

// TestEIP7702_GasCalculationTransition tests gas calculation for SetCode transactions
func TestEIP7702_GasCalculationTransition(t *testing.T) {
	testCases := []struct {
		name            string
		authListSize    int
		dataSize        int
		accessListSize  int
		expectedMinGas  uint64
	}{
		{
			name:           "Single authorization, no data",
			authListSize:   1,
			dataSize:       0,
			accessListSize: 0,
			expectedMinGas: 21000 + 6000, // Base tx + auth cost
		},
		{
			name:           "Multiple authorizations",
			authListSize:   5,
			dataSize:       0,
			accessListSize: 0,
			expectedMinGas: 21000 + (5 * 6000), // Base tx + 5 auth costs
		},
		{
			name:           "With transaction data",
			authListSize:   1,
			dataSize:       100,
			accessListSize: 0,
			expectedMinGas: 21000 + 6000 + (100 * 16), // Base + auth + data cost
		},
		{
			name:           "With access list",
			authListSize:   1,
			dataSize:       0,
			accessListSize: 1,
			expectedMinGas: 21000 + 6000 + 2400, // Base + auth + access list
		},
		{
			name:           "Maximum authorizations",
			authListSize:   256,
			dataSize:       0,
			accessListSize: 0,
			expectedMinGas: 21000 + (256 * 6000), // Base + max auths
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create authorization list with unique keys and proper addresses
			authList := make(types.AuthorizationList, tc.authListSize)
			for i := 0; i < tc.authListSize; i++ {
				// Generate unique key for each authorization to avoid duplicates
				key, _ := crypto.GenerateKey()
				
				// Create proper non-zero address for delegation
				var addr common.Address
				copy(addr[:], []byte{0x11, 0x22, 0x33, 0x44, byte(i + 1), 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, byte(i + 10)})
				
				auth := &types.AuthorizationTuple{
					ChainID: big.NewInt(1),
					Address: addr,
					Nonce:   uint64(i + 1),
				}
				auth.SignAuthorization(key)
				authList[i] = *auth
			}
			
			// Create access list with proper addresses
			accessList := make(types.AccessList, tc.accessListSize)
			for i := 0; i < tc.accessListSize; i++ {
				var addr common.Address
				copy(addr[:], []byte{0xaa, 0xbb, 0xcc, byte(i + 1), 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x11, 0x22, 0x33, byte(i + 20)})
				
				var storageKey common.Hash
				copy(storageKey[:], []byte{byte(i + 1), 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, byte(i + 30)})
				
				accessList[i] = types.AccessTuple{
					Address:     addr,
					StorageKeys: []common.Hash{storageKey},
				}
			}
			
			// Create data
			data := make([]byte, tc.dataSize)
			for i := range data {
				data[i] = byte(i % 256)
			}
			
			// Create proper destination address
			destAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
			
			setCodeTx := &types.SetCodeTx{
				ChainID:           big.NewInt(1),
				Nonce:             1,
				GasTipCap:         big.NewInt(1000000000),
				GasFeeCap:         big.NewInt(2000000000),
				Gas:               tc.expectedMinGas + 10000, // Add buffer
				To:                &destAddr,
				Value:             big.NewInt(0),
				Data:              data,
				AccessList:        accessList,
				AuthorizationList: authList,
			}
			
			// Calculate intrinsic gas using compliance checker
			compliance := &types.EIP7702Compliance{}
			actualGas := compliance.EstimateGas(setCodeTx)
			
			if actualGas < tc.expectedMinGas {
				t.Errorf("Intrinsic gas %d less than expected minimum %d", 
					actualGas, tc.expectedMinGas)
			}
			
			// Test that transaction validation accepts this gas amount
			setCodeTx.Gas = actualGas
			validator := types.NewEIP7702Validator(big.NewInt(1))
			err := validator.ValidateSetCodeTransaction(setCodeTx)
			if err != nil {
				t.Errorf("Transaction validation failed with calculated gas: %v", err)
			}
			
			// Test that insufficient gas is rejected
			setCodeTx.Gas = actualGas - 1
			err = validator.ValidateSetCodeTransaction(setCodeTx)
			if err == nil {
				t.Error("Transaction should be rejected with insufficient gas")
			}
		})
	}
}

// TestEIP7702_DelegationStorageTransition tests delegation storage mechanisms
func TestEIP7702_DelegationStorageTransition(t *testing.T) {
	testCases := []struct {
		name              string
		authorityAddress  common.Address
		delegateAddress   common.Address
		clearDelegation   bool
	}{
		{
			name:             "Delegate to contract",
			authorityAddress: common.HexToAddress("0x1111111111111111111111111111111111111111"),
			delegateAddress:  common.HexToAddress("0x2222222222222222222222222222222222222222"),
			clearDelegation:  false,
		},
		{
			name:             "Clear delegation",
			authorityAddress: common.HexToAddress("0x3333333333333333333333333333333333333333"),
			delegateAddress:  common.Address{}, // Zero address clears delegation
			clearDelegation:  true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var delegationCode []byte
			
			if tc.clearDelegation {
				// For clearing delegation, code should be empty
				delegationCode = nil
			} else {
				// Create delegation code
				delegationCode = types.AddressToDelegation(tc.delegateAddress)
			}
			
			// Test delegation code format
			if !tc.clearDelegation {
				if len(delegationCode) != 23 {
					t.Errorf("Delegation code should be 23 bytes, got %d", len(delegationCode))
				}
				
				// Parse delegation
				parsedAddr, isValid := types.ParseDelegation(delegationCode)
				if !isValid {
					t.Error("Failed to parse delegation code")
				}
				
				if parsedAddr != tc.delegateAddress {
					t.Errorf("Parsed address mismatch: expected %s, got %s",
						tc.delegateAddress.Hex(), parsedAddr.Hex())
				}
			}
			
			// Test delegation storage key generation
			storageKey := types.GetDelegationStorageKey(tc.authorityAddress)
			if storageKey == (common.Hash{}) {
				t.Error("Storage key should not be zero hash")
			}
			
			// Test delegation storage address
			storageAddr := types.GetDelegationStorageAddress()
			expectedStorageAddr := common.HexToAddress("0x7702000000000000000000000000000000000000")
			if storageAddr != expectedStorageAddr {
				t.Errorf("Storage address mismatch: expected %s, got %s",
					expectedStorageAddr.Hex(), storageAddr.Hex())
			}
		})
	}
}

// TestEIP7702_AuthorizationListSizeLimits tests authorization list size validation
func TestEIP7702_AuthorizationListSizeLimits(t *testing.T) {
	key, _ := crypto.GenerateKey()
	
	testCases := []struct {
		name          string
		listSize      int
		expectValid   bool
		expectedError string
	}{
		{
			name:        "Empty list",
			listSize:    0,
			expectValid: true, // Empty lists might be valid in some contexts
		},
		{
			name:        "Single authorization",
			listSize:    1,
			expectValid: true,
		},
		{
			name:        "Maximum valid size",
			listSize:    256,
			expectValid: true,
		},
		{
			name:          "Exceeds maximum size",
			listSize:      257,
			expectValid:   false,
			expectedError: "exceeds maximum size",
		},
		{
			name:          "Far exceeds maximum size",
			listSize:      1000,
			expectValid:   false,
			expectedError: "exceeds maximum size",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create authorization list
			authList := make(types.AuthorizationList, tc.listSize)
			for i := 0; i < tc.listSize; i++ {
				auth := &types.AuthorizationTuple{
					ChainID: big.NewInt(1),
					Address: common.Address{byte(i % 256)},
					Nonce:   uint64(i + 1),
				}
				auth.SignAuthorization(key)
				authList[i] = *auth
			}
			
			// Test list size validation
			err := types.ValidateAuthorizationListSize(authList)
			
			if tc.expectValid && err != nil {
				t.Errorf("Expected valid list size, got error: %v", err)
			}
			
			if !tc.expectValid && err == nil {
				t.Errorf("Expected list size error, but validation passed")
			}
			
			if !tc.expectValid && tc.expectedError != "" && err != nil {
				if !contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", 
						tc.expectedError, err.Error())
				}
			}
		})
	}
}

// TestEIP7702_CrossChainProtection tests that cross-chain replay protection works
func TestEIP7702_CrossChainProtection(t *testing.T) {
	key, _ := crypto.GenerateKey()
	
	testCases := []struct {
		name              string
		authChainID       *big.Int
		validationChainID *big.Int
		expectValid       bool
	}{
		{
			name:              "Same chain ID",
			authChainID:       big.NewInt(1),
			validationChainID: big.NewInt(1),
			expectValid:       true,
		},
		{
			name:              "Different chain ID",
			authChainID:       big.NewInt(1),
			validationChainID: big.NewInt(42),
			expectValid:       false,
		},
		{
			name:              "Mainnet vs testnet",
			authChainID:       big.NewInt(1),
			validationChainID: big.NewInt(5), // Goerli
			expectValid:       false,
		},
		{
			name:              "High chain ID",
			authChainID:       big.NewInt(1337),
			validationChainID: big.NewInt(1337),
			expectValid:       true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create authorization with specific chain ID
			auth := &types.AuthorizationTuple{
				ChainID: tc.authChainID,
				Address: common.Address{2},
				Nonce:   1,
			}
			auth.SignAuthorization(key)
			
			// Create validator for different chain ID
			validator := types.NewEIP7702Validator(tc.validationChainID)
			
			// Create SetCode transaction
			setCodeTx := &types.SetCodeTx{
				ChainID:           tc.validationChainID,
				Nonce:             1,
				GasTipCap:         big.NewInt(1000000000),
				GasFeeCap:         big.NewInt(2000000000),
				Gas:               30000,
				To:                &common.Address{1},
				Value:             big.NewInt(0),
				Data:              []byte{},
				AuthorizationList: types.AuthorizationList{*auth},
			}
			
			// Validate transaction
			err := validator.ValidateSetCodeTransaction(setCodeTx)
			
			if tc.expectValid && err != nil {
				t.Errorf("Expected valid cross-chain validation, got error: %v", err)
			}
			
			if !tc.expectValid && err == nil {
				t.Errorf("Expected cross-chain validation error, but validation passed")
			}
		})
	}
}

// TestEIP7702_ComprehensiveComplianceCheck tests the comprehensive compliance checker
func TestEIP7702_ComprehensiveComplianceCheck(t *testing.T) {
	key, _ := crypto.GenerateKey()
	compliance := &types.EIP7702Compliance{}
	
	testCases := []struct {
		name               string
		createTransaction  func() *types.SetCodeTx
		expectErrors       int
		errorSubstrings    []string
	}{
		{
			name: "Fully compliant transaction",
			createTransaction: func() *types.SetCodeTx {
				auth := &types.AuthorizationTuple{
					ChainID: big.NewInt(1),
					Address: common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
					Nonce:   1,
				}
				auth.SignAuthorization(key)
				
				return &types.SetCodeTx{
					ChainID:           big.NewInt(1),
					Nonce:             1,
					GasTipCap:         big.NewInt(1000000000),
					GasFeeCap:         big.NewInt(2000000000),
					Gas:               30000,
					To:                &common.Address{1},
					Value:             big.NewInt(0),
					Data:              []byte{},
					AuthorizationList: types.AuthorizationList{*auth},
				}
			},
			expectErrors: 0,
		},
		{
			name: "Invalid delegation address",
			createTransaction: func() *types.SetCodeTx {
				auth := &types.AuthorizationTuple{
					ChainID: big.NewInt(1),
					Address: common.Address{}, // Zero address - might be invalid for delegation
					Nonce:   1,
				}
				auth.SignAuthorization(key)
				
				return &types.SetCodeTx{
					ChainID:           big.NewInt(1),
					Nonce:             1,
					GasTipCap:         big.NewInt(1000000000),
					GasFeeCap:         big.NewInt(2000000000),
					Gas:               30000,
					To:                &common.Address{1},
					Value:             big.NewInt(0),
					Data:              []byte{},
					AuthorizationList: types.AuthorizationList{*auth},
				}
			},
			expectErrors:    1,
			errorSubstrings: []string{"invalid delegation address"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setCodeTx := tc.createTransaction()
			
			// Run compliance check
			errors := compliance.CheckTransactionCompliance(setCodeTx, big.NewInt(1))
			
			if len(errors) != tc.expectErrors {
				t.Errorf("Expected %d errors, got %d: %v", 
					tc.expectErrors, len(errors), errors)
			}
			
			// Check error messages contain expected substrings
			for i, expectedSubstring := range tc.errorSubstrings {
				if i >= len(errors) {
					t.Errorf("Expected error %d with substring '%s', but only got %d errors",
						i, expectedSubstring, len(errors))
					continue
				}
				
				if !contains(errors[i].Error(), expectedSubstring) {
					t.Errorf("Error %d should contain '%s', got '%s'",
						i, expectedSubstring, errors[i].Error())
				}
			}
		})
	}
}

// TestEIP7702_EstimateGas tests gas estimation for SetCode transactions
func TestEIP7702_EstimateGas(t *testing.T) {
	key, _ := crypto.GenerateKey()
	compliance := &types.EIP7702Compliance{}
	
	testCases := []struct {
		name           string
		authCount      int
		dataSize       int
		expectedMinGas uint64
	}{
		{
			name:           "Simple transaction",
			authCount:      1,
			dataSize:       0,
			expectedMinGas: 21000 + 6000, // Base + 1 auth
		},
		{
			name:           "Multiple authorizations",
			authCount:      3,
			dataSize:       0,
			expectedMinGas: 21000 + (3 * 6000), // Base + 3 auths
		},
		{
			name:           "With data",
			authCount:      1,
			dataSize:       32,
			expectedMinGas: 21000 + 6000 + (32 * 16), // Base + auth + data
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create authorization list
			authList := make(types.AuthorizationList, tc.authCount)
			for i := 0; i < tc.authCount; i++ {
				auth := &types.AuthorizationTuple{
					ChainID: big.NewInt(1),
					Address: common.Address{byte(i + 1)},
					Nonce:   uint64(i + 1),
				}
				auth.SignAuthorization(key)
				authList[i] = *auth
			}
			
			// Create data
			data := make([]byte, tc.dataSize)
			for i := range data {
				data[i] = 0x60 // PUSH1 opcode
			}
			
			setCodeTx := &types.SetCodeTx{
				ChainID:           big.NewInt(1),
				Nonce:             1,
				GasTipCap:         big.NewInt(1000000000),
				GasFeeCap:         big.NewInt(2000000000),
				Gas:               100000, // Will be ignored by estimation
				To:                &common.Address{1},
				Value:             big.NewInt(0),
				Data:              data,
				AuthorizationList: authList,
			}
			
			// Estimate gas
			estimatedGas := compliance.EstimateGas(setCodeTx)
			
			if estimatedGas < tc.expectedMinGas {
				t.Errorf("Estimated gas %d less than expected minimum %d",
					estimatedGas, tc.expectedMinGas)
			}
			
			// Estimated gas should be reasonable (not excessively high)
			if estimatedGas > tc.expectedMinGas*2 {
				t.Errorf("Estimated gas %d seems too high for expected minimum %d",
					estimatedGas, tc.expectedMinGas)
			}
		})
	}
}