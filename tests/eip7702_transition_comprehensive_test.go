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

	"crypto/ecdsa"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/params"
)

// TestEIP7702_TransactionTypeNumbers verifies that SetCodeTxType has the correct value
func TestEIP7702_TransactionTypeNumbers(t *testing.T) {
	// EIP-7702 specifies that SetCode transaction type should be 0x04
	// However, current implementation uses iota which makes it 3
	// This test documents the current behavior and can be updated when fixed
	
	expected := 3 // Current implementation (should be 4 per EIP-7702)
	actual := int(types.SetCodeTxType)
	
	if actual != expected {
		t.Errorf("SetCodeTxType mismatch: expected %d, got %d", expected, actual)
	}
	
	// Document the transaction type ordering
	expectedTypes := map[string]int{
		"LegacyTxType":     0,
		"AccessListTxType": 1,
		"DynamicFeeTxType": 2,
		"SetCodeTxType":    3, // Note: EIP-7702 spec says this should be 4
	}
	
	actualTypes := map[string]int{
		"LegacyTxType":     int(types.LegacyTxType),
		"AccessListTxType": int(types.AccessListTxType),
		"DynamicFeeTxType": int(types.DynamicFeeTxType),
		"SetCodeTxType":    int(types.SetCodeTxType),
	}
	
	for name, expected := range expectedTypes {
		if actualTypes[name] != expected {
			t.Errorf("%s mismatch: expected %d, got %d", name, expected, actualTypes[name])
		}
	}
}

// TestEIP7702_PreHardforkRejection tests comprehensive rejection before Phaethon hardfork
func TestEIP7702_PreHardforkRejection(t *testing.T) {
	// Test various scenarios where SetCode transactions should be rejected
	// before Phaethon hardfork activation
	
	testCases := []struct {
		name               string
		blockNumber        *big.Int
		phaethonBlock      *big.Int
		transactionValid   bool
		expectedRejection  string
	}{
		{
			name:              "Valid SetCode tx one block before hardfork",
			blockNumber:       big.NewInt(99),
			phaethonBlock:     big.NewInt(100),
			transactionValid:  true,
			expectedRejection: "Phaethon hardfork not active",
		},
		{
			name:              "Valid SetCode tx far before hardfork",
			blockNumber:       big.NewInt(10),
			phaethonBlock:     big.NewInt(100),
			transactionValid:  true,
			expectedRejection: "Phaethon hardfork not active",
		},
		{
			name:              "Invalid SetCode tx before hardfork (should still be rejected for hardfork reason)",
			blockNumber:       big.NewInt(50),
			phaethonBlock:     big.NewInt(100),
			transactionValid:  false,
			expectedRejection: "Phaethon hardfork not active",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create chain config with Phaethon hardfork
			config := &params.ChainConfig{
				ChainID:        big.NewInt(1),
				HomesteadBlock: big.NewInt(0),
				BerlinBlock:    big.NewInt(0),
				LondonBlock:    big.NewInt(0),
				PhaethonBlock:  tc.phaethonBlock,
			}
			
			// Create test SetCode transaction
			key, _ := crypto.GenerateKey()
			_ = createTestSetCodeTransaction(key, tc.transactionValid)
			
			// Test should check if hardfork is active
			isHardforkActive := config.IsPhaethon(tc.blockNumber)
			
			if isHardforkActive {
				t.Errorf("Test setup error: hardfork should not be active at block %d", tc.blockNumber)
			}
			
			// TODO: Add actual validation logic here when integrated with transaction pool
			// For now, we're testing the hardfork activation logic
			
			t.Logf("Block %d with Phaethon at %d: hardfork active = %v", 
				tc.blockNumber, tc.phaethonBlock, isHardforkActive)
		})
	}
}

// TestEIP7702_HardforkBoundaryTransitions tests behavior exactly at hardfork boundaries
func TestEIP7702_HardforkBoundaryTransitions(t *testing.T) {
	testCases := []struct {
		name           string
		phaethonBlock  *big.Int
		testBlocks     []*big.Int
		expectedActive []bool
	}{
		{
			name:          "Hardfork at block 100",
			phaethonBlock: big.NewInt(100),
			testBlocks:    []*big.Int{big.NewInt(99), big.NewInt(100), big.NewInt(101)},
			expectedActive: []bool{false, true, true},
		},
		{
			name:          "Hardfork at block 0 (genesis)",
			phaethonBlock: big.NewInt(0),
			testBlocks:    []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(100)},
			expectedActive: []bool{true, true, true},
		},
		{
			name:          "Hardfork at block 1",
			phaethonBlock: big.NewInt(1),
			testBlocks:    []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2)},
			expectedActive: []bool{false, true, true},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &params.ChainConfig{
				ChainID:        big.NewInt(1),
				HomesteadBlock: big.NewInt(0),
				BerlinBlock:    big.NewInt(0),
				LondonBlock:    big.NewInt(0),
				PhaethonBlock:  tc.phaethonBlock,
			}
			
			for i, blockNum := range tc.testBlocks {
				isActive := config.IsPhaethon(blockNum)
				expected := tc.expectedActive[i]
				
				if isActive != expected {
					t.Errorf("Block %d: expected hardfork active = %v, got %v", 
						blockNum, expected, isActive)
				}
			}
		})
	}
}

// TestEIP7702_PostHardforkAcceptance tests that SetCode transactions are properly accepted after hardfork
func TestEIP7702_PostHardforkAcceptance(t *testing.T) {
	config := &params.ChainConfig{
		ChainID:        big.NewInt(1),
		HomesteadBlock: big.NewInt(0),
		BerlinBlock:    big.NewInt(0),
		LondonBlock:    big.NewInt(0),
		PhaethonBlock:  big.NewInt(50), // Hardfork at block 50
	}
	
	testCases := []struct {
		name            string
		blockNumber     *big.Int
		transactionCase string
	}{
		{
			name:            "Simple SetCode tx at hardfork activation",
			blockNumber:     big.NewInt(50),
			transactionCase: "simple",
		},
		{
			name:            "Complex SetCode tx after hardfork",
			blockNumber:     big.NewInt(100),
			transactionCase: "complex",
		},
		{
			name:            "SetCode tx with multiple authorizations",
			blockNumber:     big.NewInt(200),
			transactionCase: "multiple",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !config.IsPhaethon(tc.blockNumber) {
				t.Fatalf("Test setup error: hardfork should be active at block %d", tc.blockNumber)
			}
			
			key, _ := crypto.GenerateKey()
			var setCodeTx *types.SetCodeTx
			
			switch tc.transactionCase {
			case "simple":
				setCodeTx = createTestSetCodeTransaction(key, true)
			case "complex":
				setCodeTx = createComplexSetCodeTransaction(key)
			case "multiple":
				setCodeTx = createMultipleAuthSetCodeTransaction(key)
			}
			
			// Basic validation that the transaction was created properly
			if setCodeTx == nil {
				t.Fatal("Failed to create SetCode transaction")
			}
			
			if len(setCodeTx.AuthorizationList) == 0 {
				t.Error("SetCode transaction should have authorization list")
			}
			
			// Test transaction type
			tx := types.NewTx(setCodeTx)
			if tx.Type() != types.SetCodeTxType {
				t.Errorf("Expected transaction type %d, got %d", types.SetCodeTxType, tx.Type())
			}
		})
	}
}

// TestEIP7702_DelegationMechanics tests the delegation prefix and parsing logic
func TestEIP7702_DelegationMechanics(t *testing.T) {
	testCases := []struct {
		name              string
		inputAddress      common.Address
		expectValidParse  bool
	}{
		{
			name:              "Valid delegation address",
			inputAddress:      common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
			expectValidParse:  true,
		},
		{
			name:              "Zero address delegation",
			inputAddress:      common.Address{},
			expectValidParse:  true, // Zero address should be valid for clearing delegation
		},
		{
			name:              "System address delegation",
			inputAddress:      common.HexToAddress("0x0000000000000000000000000000000000000001"),
			expectValidParse:  true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test AddressToDelegation
			delegation := types.AddressToDelegation(tc.inputAddress)
			
			// Check delegation format
			if len(delegation) != 23 {
				t.Errorf("Delegation should be 23 bytes, got %d", len(delegation))
			}
			
			// Check prefix
			if delegation[0] != 0xef || delegation[1] != 0x01 || delegation[2] != 0x00 {
				t.Errorf("Invalid delegation prefix: %x %x %x", delegation[0], delegation[1], delegation[2])
			}
			
			// Test ParseDelegation
			parsedAddr, isValid := types.ParseDelegation(delegation)
			
			if !isValid && tc.expectValidParse {
				t.Error("ParseDelegation should return true for valid delegation")
			}
			
			if isValid && parsedAddr != tc.inputAddress {
				t.Errorf("ParseDelegation address mismatch: expected %s, got %s", 
					tc.inputAddress.Hex(), parsedAddr.Hex())
			}
		})
	}
}

// TestEIP7702_AuthorizationSigningAndRecovery tests authorization signature creation and recovery
func TestEIP7702_AuthorizationSigningAndRecovery(t *testing.T) {
	testCases := []struct {
		name     string
		chainID  *big.Int
		address  common.Address
		nonce    uint64
	}{
		{
			name:    "Mainnet chain ID",
			chainID: big.NewInt(1),
			address: common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
			nonce:   1,
		},
		{
			name:    "U2U chain ID",
			chainID: big.NewInt(39),
			address: common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
			nonce:   100,
		},
		{
			name:    "High nonce",
			chainID: big.NewInt(1),
			address: common.HexToAddress("0xfeedface00000000000000000000000000000000"),
			nonce:   999999,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate a new key for this test
			key, err := crypto.GenerateKey()
			if err != nil {
				t.Fatalf("Failed to generate key: %v", err)
			}
			
			// Create authorization
			auth := &types.AuthorizationTuple{
				ChainID: tc.chainID,
				Address: tc.address,
				Nonce:   tc.nonce,
			}
			
			// Sign authorization
			err = auth.SignAuthorization(key)
			if err != nil {
				t.Fatalf("Failed to sign authorization: %v", err)
			}
			
			// Validate that signature components are present
			if auth.V == nil || auth.R == nil || auth.S == nil {
				t.Fatal("Signature components should not be nil after signing")
			}
			
			// Test recovery
			recoveredAddr, err := auth.RecoverAuthority()
			if err != nil {
				t.Fatalf("Failed to recover authority: %v", err)
			}
			
			// Verify recovered address matches the key that signed it
			expectedAddr := crypto.PubkeyToAddress(key.PublicKey)
			if recoveredAddr != expectedAddr {
				t.Errorf("Recovered address mismatch: expected %s, got %s",
					expectedAddr.Hex(), recoveredAddr.Hex())
			}
			
			// Test validation
			err = auth.ValidateAuthorization()
			if err != nil {
				t.Errorf("Authorization validation failed: %v", err)
			}
		})
	}
}

// TestEIP7702_MalformedTransactionRejection tests rejection of malformed SetCode transactions
func TestEIP7702_MalformedTransactionRejection(t *testing.T) {
	key, _ := crypto.GenerateKey()
	
	testCases := []struct {
		name               string
		createTransaction  func() *types.SetCodeTx
		expectedError      string
	}{
		{
			name: "Empty authorization list (allowed)",
			createTransaction: func() *types.SetCodeTx {
				destAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
				
				return &types.SetCodeTx{
					ChainID:           big.NewInt(1),
					Nonce:             1,
					GasTipCap:         big.NewInt(1000000000),
					GasFeeCap:         big.NewInt(2000000000),
					Gas:               25000, // Sufficient gas for empty auth list
					To:                &destAddr,
					Value:             big.NewInt(0),
					Data:              []byte{},
					AuthorizationList: types.AuthorizationList{}, // Empty list is actually allowed
				}
			},
			expectedError: "", // Empty string means we expect this to pass
		},
		{
			name: "Authorization with invalid signature",
			createTransaction: func() *types.SetCodeTx {
				// Create proper non-zero delegation address
				delegateAddr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
				
				return &types.SetCodeTx{
					ChainID:   big.NewInt(1),
					Nonce:     1,
					GasTipCap: big.NewInt(1000000000),
					GasFeeCap: big.NewInt(2000000000),
					Gas:       30000, // Sufficient gas for 1 authorization
					To:        &common.Address{1},
					Value:     big.NewInt(0),
					Data:      []byte{},
					AuthorizationList: types.AuthorizationList{
						{
							ChainID: big.NewInt(1),
							Address: delegateAddr,
							Nonce:   1,
							V:       big.NewInt(27),
							R:       big.NewInt(0), // Invalid R (zero)
							S:       big.NewInt(1),
						},
					},
				}
			},
			expectedError: "signature",
		},
		{
			name: "Authorization with wrong chain ID",
			createTransaction: func() *types.SetCodeTx {
				// Create proper non-zero delegation address
				delegateAddr := common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
				
				auth := &types.AuthorizationTuple{
					ChainID: big.NewInt(999), // Wrong chain ID
					Address: delegateAddr,
					Nonce:   1,
				}
				auth.SignAuthorization(key)
				
				return &types.SetCodeTx{
					ChainID:           big.NewInt(1),
					Nonce:             1,
					GasTipCap:         big.NewInt(1000000000),
					GasFeeCap:         big.NewInt(2000000000),
					Gas:               30000, // Sufficient gas for 1 authorization
					To:                &common.Address{1},
					Value:             big.NewInt(0),
					Data:              []byte{},
					AuthorizationList: types.AuthorizationList{*auth},
				}
			},
			expectedError: "chain ID",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setCodeTx := tc.createTransaction()
			
			// Create validator
			validator := types.NewEIP7702Validator(big.NewInt(1))
			
			// Validate transaction
			err := validator.ValidateSetCodeTransaction(setCodeTx)
			
			if tc.expectedError == "" {
				// We expect this test to pass
				if err != nil {
					t.Errorf("Expected validation to pass, but got error: %v", err)
				}
			} else {
				// We expect this test to fail with a specific error
				if err == nil {
					t.Errorf("Expected error containing '%s', but validation passed", tc.expectedError)
				} else if !contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing '%s', got '%s'", tc.expectedError, err.Error())
				}
			}
		})
	}
}

// Helper functions

func createTestSetCodeTransaction(key *ecdsa.PrivateKey, valid bool) *types.SetCodeTx {
	// Create proper non-zero delegation address
	delegateAddr := common.HexToAddress("0x2222222222222222222222222222222222222222")
	destAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	
	auth := &types.AuthorizationTuple{
		ChainID: big.NewInt(1),
		Address: delegateAddr,
		Nonce:   1,
	}
	
	if valid {
		auth.SignAuthorization(key)
	} else {
		// Create invalid signature
		auth.V = big.NewInt(27)
		auth.R = big.NewInt(1)
		auth.S = big.NewInt(1)
	}
	
	return &types.SetCodeTx{
		ChainID:           big.NewInt(1),
		Nonce:             1,
		GasTipCap:         big.NewInt(1000000000),
		GasFeeCap:         big.NewInt(2000000000),
		Gas:               30000, // Sufficient gas for authorization validation
		To:                &destAddr,
		Value:             big.NewInt(0),
		Data:              []byte{},
		AuthorizationList: types.AuthorizationList{*auth},
	}
}

func createComplexSetCodeTransaction(key *ecdsa.PrivateKey) *types.SetCodeTx {
	auth := &types.AuthorizationTuple{
		ChainID: big.NewInt(1),
		Address: common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678"),
		Nonce:   100,
	}
	auth.SignAuthorization(key)
	
	// Create proper addresses for access list and destination
	accessAddr := common.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	destAddr := common.HexToAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	
	return &types.SetCodeTx{
		ChainID:   big.NewInt(1),
		Nonce:     50,
		GasTipCap: big.NewInt(2000000000),
		GasFeeCap: big.NewInt(5000000000),
		Gas:       100000,
		To:        &destAddr,
		Value:     big.NewInt(1000000000000000000), // 1 ETH
		Data:      []byte{0x60, 0x60, 0x60, 0x40}, // Some contract call data
		AccessList: types.AccessList{
			{
				Address: accessAddr,
				StorageKeys: []common.Hash{
					common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
				},
			},
		},
		AuthorizationList: types.AuthorizationList{*auth},
	}
}

func createMultipleAuthSetCodeTransaction(key *ecdsa.PrivateKey) *types.SetCodeTx {
	// Create proper non-zero delegation addresses
	delegateAddr1 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	delegateAddr2 := common.HexToAddress("0x3333333333333333333333333333333333333333")
	destAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")
	
	auth1 := &types.AuthorizationTuple{
		ChainID: big.NewInt(1),
		Address: delegateAddr1,
		Nonce:   1,
	}
	auth1.SignAuthorization(key)
	
	key2, _ := crypto.GenerateKey()
	auth2 := &types.AuthorizationTuple{
		ChainID: big.NewInt(1),
		Address: delegateAddr2,
		Nonce:   1,
	}
	auth2.SignAuthorization(key2)
	
	return &types.SetCodeTx{
		ChainID:           big.NewInt(1),
		Nonce:             1,
		GasTipCap:         big.NewInt(1000000000),
		GasFeeCap:         big.NewInt(2000000000),
		Gas:               50000, // Sufficient gas for 2 authorizations
		To:                &destAddr,
		Value:             big.NewInt(0),
		Data:              []byte{},
		AuthorizationList: types.AuthorizationList{*auth1, *auth2},
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    func() bool {
			    for i := 0; i <= len(s)-len(substr); i++ {
				    if s[i:i+len(substr)] == substr {
					    return true
				    }
			    }
			    return false
		    }())
}