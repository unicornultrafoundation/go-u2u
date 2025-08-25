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

package evmcore

import (
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/params"
)

// TestEIP7702_HardforkActivation tests that SetCode transactions are properly validated
// based on hardfork activation
func TestEIP7702_HardforkActivation(t *testing.T) {
	tests := []struct {
		name           string
		blockNumber    *big.Int
		phaethonBlock  *big.Int
		expectAccepted bool
		errorMsg       string
	}{
		{
			name:           "SetCode transaction before Phaethon hardfork",
			blockNumber:    big.NewInt(10),
			phaethonBlock:  big.NewInt(20),
			expectAccepted: false,
			errorMsg:       "pool not yet in Phaethon hardfork",
		},
		{
			name:           "SetCode transaction at Phaethon hardfork activation",
			blockNumber:    big.NewInt(20),
			phaethonBlock:  big.NewInt(20),
			expectAccepted: true,
		},
		{
			name:           "SetCode transaction after Phaethon hardfork",
			blockNumber:    big.NewInt(25),
			phaethonBlock:  big.NewInt(20),
			expectAccepted: true,
		},
		{
			name:           "SetCode transaction with nil Phaethon block",
			blockNumber:    big.NewInt(10),
			phaethonBlock:  nil,
			expectAccepted: false,
			errorMsg:       "pool not yet in Phaethon hardfork",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test chain config
			config := &params.ChainConfig{
				ChainID:       big.NewInt(1),
				HomesteadBlock: big.NewInt(0),
				BerlinBlock:   big.NewInt(0),
				LondonBlock:   big.NewInt(0),
				PhaethonBlock: tt.phaethonBlock,
			}

			// Create private key for signing authorization
			key, _ := crypto.GenerateKey()

			// Create and sign authorization
			auth := &types.AuthorizationTuple{
				ChainID: big.NewInt(1),
				Address: common.Address{2},
				Nonce:   1,
			}
			auth.SignAuthorization(key)

			// Create a mock SetCode transaction
			setCodeTx := &types.SetCodeTx{
				ChainID:   big.NewInt(1),
				Nonce:     1,
				GasTipCap: big.NewInt(1000000000),
				GasFeeCap: big.NewInt(2000000000),
				Gas:       21000,
				To:        &common.Address{1},
				Value:     big.NewInt(0),
				Data:      []byte{},
				AuthorizationList: types.AuthorizationList{*auth},
			}

			tx := types.NewTx(setCodeTx)

			// Sign the transaction
			signer := types.NewPhaethonSigner(big.NewInt(1))
			signedTx, err := types.SignTx(tx, signer, key)
			if err != nil {
				t.Fatalf("Failed to sign transaction: %v", err)
			}
			tx = signedTx

			// Create validation options
			opts := &ValidationOptions{
				Config:       config,
				Accept:       (1 << types.SetCodeTxType),
				MaxSize:      1024 * 1024,
				MinTip:       big.NewInt(0),
				MinGasPrice:  big.NewInt(0),
				PoolGasPrice: big.NewInt(1000000000), // 1 gwei
			}

			// Create mock EVM header
			header := &EvmHeader{
				Number:   tt.blockNumber,
				GasLimit: 8000000,
			}

			// Validate transaction  
			err = ValidateTransaction(tx, header, signer, opts)

			if tt.expectAccepted {
				if err != nil {
					t.Errorf("Expected transaction to be accepted, but got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected transaction to be rejected, but it was accepted")
				}
				if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					// Check if error message contains expected substring
					if !contains(err.Error(), tt.errorMsg) {
						t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
					}
				}
			}
		})
	}
}

// TestEIP7702_TransactionPoolIntegration tests the integration of EIP-7702 with transaction pool
func TestEIP7702_TransactionPoolIntegration(t *testing.T) {
	// Test that SetCode transactions are included in the accepted transaction types
	opts := &ValidationOptions{
		Config: &params.ChainConfig{
			ChainID:       big.NewInt(1),
			BerlinBlock:   big.NewInt(0),
			LondonBlock:   big.NewInt(0),
			PhaethonBlock: big.NewInt(0),
		},
		Accept: (1 << types.LegacyTxType) |
			(1 << types.AccessListTxType) |
			(1 << types.DynamicFeeTxType) |
			(1 << types.SetCodeTxType),
		MaxSize:      1024 * 1024,
		MinTip:       big.NewInt(0),
		MinGasPrice:  big.NewInt(0),
		PoolGasPrice: big.NewInt(1000000000), // 1 gwei
	}

	// Create private key for signing authorization
	key, _ := crypto.GenerateKey()

	// Create and sign authorization
	auth := &types.AuthorizationTuple{
		ChainID: big.NewInt(1),
		Address: common.Address{2},
		Nonce:   1,
	}
	auth.SignAuthorization(key)

	// Create a valid SetCode transaction
	setCodeTx := &types.SetCodeTx{
		ChainID:   big.NewInt(1),
		Nonce:     1,
		GasTipCap: big.NewInt(1000000000),
		GasFeeCap: big.NewInt(2000000000),
		Gas:       21000,
		To:        &common.Address{1},
		Value:     big.NewInt(0),
		Data:      []byte{},
		AuthorizationList: types.AuthorizationList{*auth},
	}

	tx := types.NewTx(setCodeTx)

	// Sign the transaction
	signer := types.NewPhaethonSigner(big.NewInt(1))
	signedTx, signErr := types.SignTx(tx, signer, key)
	if signErr != nil {
		t.Fatalf("Failed to sign transaction: %v", signErr)
	}
	tx = signedTx

	header := &EvmHeader{
		Number:   big.NewInt(100),
		GasLimit: 8000000,
	}

	// Validate transaction
	err := ValidateTransaction(tx, header, signer, opts)
	if err != nil {
		t.Errorf("Valid SetCode transaction should be accepted: %v", err)
	}

	// Test that SetCode is not accepted when not in the Accept bitmask
	optsNoSetCode := &ValidationOptions{
		Config: opts.Config,
		Accept: (1 << types.LegacyTxType) |
			(1 << types.AccessListTxType) |
			(1 << types.DynamicFeeTxType),
		MaxSize: 1024 * 1024,
	}

	err = ValidateTransaction(tx, header, signer, optsNoSetCode)
	if err == nil {
		t.Errorf("SetCode transaction should be rejected when not in Accept bitmask")
	}
}

// TestEIP7702_AuthorizationListValidation tests the authorization list validation
func TestEIP7702_AuthorizationListValidation(t *testing.T) {
	config := &params.ChainConfig{
		ChainID:       big.NewInt(1),
		HomesteadBlock: big.NewInt(0),
		BerlinBlock:   big.NewInt(0),
		LondonBlock:   big.NewInt(0),
		PhaethonBlock: big.NewInt(0),
	}

	opts := &ValidationOptions{
		Config:       config,
		Accept:       (1 << types.SetCodeTxType),
		MaxSize:      1024 * 1024,
		MinTip:       big.NewInt(0),
		MinGasPrice:  big.NewInt(0),
		PoolGasPrice: big.NewInt(1000000000), // 1 gwei
	}

	header := &EvmHeader{
		Number:   big.NewInt(100),
		GasLimit: 8000000,
	}
	signer := types.NewPhaethonSigner(big.NewInt(1))

	// Create properly signed authorization for valid test case
	key, _ := crypto.GenerateKey()
	validAuth := &types.AuthorizationTuple{
		ChainID: big.NewInt(1),
		Address: common.Address{2},
		Nonce:   1,
	}
	validAuth.SignAuthorization(key)

	tests := []struct {
		name        string
		authList    types.AuthorizationList
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid authorization list",
			authList:    types.AuthorizationList{*validAuth},
			expectError: false,
		},
		{
			name: "Authorization list too large",
			authList: func() types.AuthorizationList {
				list := make(types.AuthorizationList, 257) // Over the limit of 256
				for i := range list {
					list[i] = types.AuthorizationTuple{
						ChainID: big.NewInt(1),
						Address: common.Address{byte(i % 256)},
						Nonce:   uint64(i),
						V:       big.NewInt(27),
						R:       big.NewInt(1),
						S:       big.NewInt(1),
					}
				}
				return list
			}(),
			expectError: true,
			errorMsg:    "EIP-7702 authorization list exceeds maximum size of 256",
		},
		{
			name: "Missing signature components",
			authList: types.AuthorizationList{
				{
					ChainID: big.NewInt(1),
					Address: common.Address{2},
					Nonce:   1,
					V:       nil, // Missing V
					R:       big.NewInt(1),
					S:       big.NewInt(1),
				},
			},
			expectError: true,
			errorMsg:    "authorization 0: failed to recover authority: invalid authorization signature",
		},
		{
			name: "Zero signature values",
			authList: types.AuthorizationList{
				{
					ChainID: big.NewInt(1),
					Address: common.Address{2},
					Nonce:   1,
					V:       big.NewInt(27),
					R:       big.NewInt(0), // Zero R value
					S:       big.NewInt(1),
				},
			},
			expectError: true,
			errorMsg:    "authorization 0: failed to recover authority: invalid authorization signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setCodeTx := &types.SetCodeTx{
				ChainID:           big.NewInt(1),
				Nonce:             1,
				GasTipCap:         big.NewInt(1000000000),
				GasFeeCap:         big.NewInt(2000000000),
				Gas:               21000,
				To:                &common.Address{1},
				Value:             big.NewInt(0),
				Data:              []byte{},
				AuthorizationList: tt.authList,
			}

			tx := types.NewTx(setCodeTx)
			
			// Sign the transaction if we expect success (to avoid "invalid sender" error)
			if !tt.expectError {
				signedTx, signErr := types.SignTx(tx, signer, key)
				if signErr != nil {
					t.Fatalf("Failed to sign transaction: %v", signErr)
				}
				tx = signedTx
			}
			
			err := ValidateTransaction(tx, header, signer, opts)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s[0:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		(len(s) > len(substr) && func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()))
}