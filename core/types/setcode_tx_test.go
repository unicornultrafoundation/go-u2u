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
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/rlp"
)

// Test key and addresses for EIP-7702 testing
var (
	testKey7702, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	testAddr7702    = crypto.PubkeyToAddress(testKey7702.PublicKey)
	testChainID7702 = big.NewInt(1)
	testNonce7702   = uint64(42)
	codeAddr7702    = common.HexToAddress("0x1234567890123456789012345678901234567890")
)

func TestAuthorizationTupleValidation(t *testing.T) {
	tests := []struct {
		name    string
		auth    *AuthorizationTuple
		wantErr bool
	}{
		{
			name: "valid authorization",
			auth: &AuthorizationTuple{
				ChainID: big.NewInt(1),
				Address: codeAddr7702,
				Nonce:   42,
				V:       big.NewInt(27),
				R:       big.NewInt(1),
				S:       big.NewInt(1),
			},
			wantErr: false,
		},
		{
			name: "nil chain ID",
			auth: &AuthorizationTuple{
				ChainID: nil,
				Address: codeAddr7702,
				Nonce:   42,
				V:       big.NewInt(27),
				R:       big.NewInt(1),
				S:       big.NewInt(1),
			},
			wantErr: true,
		},
		{
			name: "nil signature values",
			auth: &AuthorizationTuple{
				ChainID: big.NewInt(1),
				Address: codeAddr7702,
				Nonce:   42,
				V:       nil,
				R:       big.NewInt(1),
				S:       big.NewInt(1),
			},
			wantErr: true,
		},
		{
			name: "zero R value",
			auth: &AuthorizationTuple{
				ChainID: big.NewInt(1),
				Address: codeAddr7702,
				Nonce:   42,
				V:       big.NewInt(27),
				R:       big.NewInt(0),
				S:       big.NewInt(1),
			},
			wantErr: true,
		},
		{
			name: "zero S value",
			auth: &AuthorizationTuple{
				ChainID: big.NewInt(1),
				Address: codeAddr7702,
				Nonce:   42,
				V:       big.NewInt(27),
				R:       big.NewInt(1),
				S:       big.NewInt(0),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.auth.ValidateAuthorization()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAuthorization() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthorizationSigning(t *testing.T) {
	auth := &AuthorizationTuple{
		ChainID: testChainID7702,
		Address: codeAddr7702,
		Nonce:   testNonce7702,
	}

	// Sign the authorization
	err := auth.SignAuthorization(testKey7702)
	if err != nil {
		t.Fatalf("SignAuthorization() error = %v", err)
	}

	// Validate the authorization
	err = auth.ValidateAuthorization()
	if err != nil {
		t.Errorf("ValidateAuthorization() error = %v", err)
	}

	// Recover the authority address
	recoveredAddr, err := auth.RecoverAuthority()
	if err != nil {
		t.Fatalf("RecoverAuthority() error = %v", err)
	}

	// Check if the recovered address matches the test address
	if recoveredAddr != testAddr7702 {
		t.Errorf("RecoverAuthority() = %v, want %v", recoveredAddr, testAddr7702)
	}
}

func TestAuthorizationListValidation(t *testing.T) {
	// Create a valid authorization
	auth1 := &AuthorizationTuple{
		ChainID: testChainID7702,
		Address: codeAddr7702,
		Nonce:   testNonce7702,
	}
	auth1.SignAuthorization(testKey7702)

	// Create another key and authorization
	testKey2, _ := crypto.GenerateKey()
	auth2 := &AuthorizationTuple{
		ChainID: testChainID7702,
		Address: common.HexToAddress("0x9876543210987654321098765432109876543210"),
		Nonce:   testNonce7702 + 1,
	}
	auth2.SignAuthorization(testKey2)

	tests := []struct {
		name    string
		authList AuthorizationList
		wantErr bool
	}{
		{
			name:     "empty list",
			authList: AuthorizationList{},
			wantErr:  false,
		},
		{
			name:     "single valid authorization",
			authList: AuthorizationList{*auth1},
			wantErr:  false,
		},
		{
			name:     "multiple valid authorizations",
			authList: AuthorizationList{*auth1, *auth2},
			wantErr:  false,
		},
		{
			name:     "duplicate authorization",
			authList: AuthorizationList{*auth1, *auth1},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.authList.ValidateAuthorizationList()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAuthorizationList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetCodeTxCreation(t *testing.T) {
	// Create authorization list
	auth := &AuthorizationTuple{
		ChainID: testChainID7702,
		Address: codeAddr7702,
		Nonce:   testNonce7702,
	}
	auth.SignAuthorization(testKey7702)
	authList := AuthorizationList{*auth}

	// Create access list
	accessList := AccessList{
		{Address: common.HexToAddress("0x1234"), StorageKeys: []common.Hash{common.HexToHash("0x5678")}},
	}

	// Create SetCodeTx
	tx := NewSetCodeTx(
		testChainID7702,
		testNonce7702,
		&codeAddr7702,
		big.NewInt(1000),
		21000,
		big.NewInt(1000000000), // gasTipCap
		big.NewInt(2000000000), // gasFeeCap
		[]byte("test data"),
		accessList,
		authList,
	)

	// Verify transaction type
	if tx.Type() != SetCodeTxType {
		t.Errorf("Type() = %v, want %v", tx.Type(), SetCodeTxType)
	}

	// Verify authorization list
	retrievedAuthList := tx.AuthorizationList()
	if len(retrievedAuthList) != 1 {
		t.Errorf("AuthorizationList() length = %v, want %v", len(retrievedAuthList), 1)
	}

	// Verify basic transaction fields
	if tx.ChainId().Cmp(testChainID7702) != 0 {
		t.Errorf("ChainId() = %v, want %v", tx.ChainId(), testChainID7702)
	}

	if tx.Nonce() != testNonce7702 {
		t.Errorf("Nonce() = %v, want %v", tx.Nonce(), testNonce7702)
	}

	if tx.Gas() != 21000 {
		t.Errorf("Gas() = %v, want %v", tx.Gas(), 21000)
	}
}

func TestSetCodeTxRLPEncoding(t *testing.T) {
	// Create authorization list
	auth := &AuthorizationTuple{
		ChainID: testChainID7702,
		Address: codeAddr7702,
		Nonce:   testNonce7702,
	}
	auth.SignAuthorization(testKey7702)
	authList := AuthorizationList{*auth}

	// Create SetCodeTx
	originalTx := NewSetCodeTx(
		testChainID7702,
		testNonce7702,
		&codeAddr7702,
		big.NewInt(1000),
		21000,
		big.NewInt(1000000000), // gasTipCap
		big.NewInt(2000000000), // gasFeeCap
		[]byte("test data"),
		AccessList{},
		authList,
	)

	// Encode transaction
	encoded, err := rlp.EncodeToBytes(originalTx)
	if err != nil {
		t.Fatalf("rlp.EncodeToBytes() error = %v", err)
	}

	// Decode transaction
	var decodedTx Transaction
	err = rlp.DecodeBytes(encoded, &decodedTx)
	if err != nil {
		t.Fatalf("rlp.DecodeBytes() error = %v", err)
	}

	// Verify decoded transaction
	if decodedTx.Type() != SetCodeTxType {
		t.Errorf("Decoded transaction type = %v, want %v", decodedTx.Type(), SetCodeTxType)
	}

	if decodedTx.ChainId().Cmp(testChainID7702) != 0 {
		t.Errorf("Decoded ChainId() = %v, want %v", decodedTx.ChainId(), testChainID7702)
	}

	if decodedTx.Nonce() != testNonce7702 {
		t.Errorf("Decoded Nonce() = %v, want %v", decodedTx.Nonce(), testNonce7702)
	}

	// Verify authorization list
	decodedAuthList := decodedTx.AuthorizationList()
	if len(decodedAuthList) != 1 {
		t.Errorf("Decoded AuthorizationList() length = %v, want %v", len(decodedAuthList), 1)
	}
}

func TestSetCodeTxBinaryEncoding(t *testing.T) {
	// Create authorization list
	auth := &AuthorizationTuple{
		ChainID: testChainID7702,
		Address: codeAddr7702,
		Nonce:   testNonce7702,
	}
	auth.SignAuthorization(testKey7702)
	authList := AuthorizationList{*auth}

	// Create SetCodeTx
	originalTx := NewSetCodeTx(
		testChainID7702,
		testNonce7702,
		&codeAddr7702,
		big.NewInt(1000),
		21000,
		big.NewInt(1000000000), // gasTipCap
		big.NewInt(2000000000), // gasFeeCap
		[]byte("test data"),
		AccessList{},
		authList,
	)

	// Test MarshalBinary
	encoded, err := originalTx.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error = %v", err)
	}

	// Test UnmarshalBinary
	var decodedTx Transaction
	err = decodedTx.UnmarshalBinary(encoded)
	if err != nil {
		t.Fatalf("UnmarshalBinary() error = %v", err)
	}

	// Verify the decoded transaction
	if decodedTx.Type() != SetCodeTxType {
		t.Errorf("Decoded transaction type = %v, want %v", decodedTx.Type(), SetCodeTxType)
	}

	// Verify that the hash is preserved
	originalHash := originalTx.Hash()
	decodedHash := decodedTx.Hash()
	if originalHash != decodedHash {
		t.Errorf("Hash mismatch: original = %v, decoded = %v", originalHash, decodedHash)
	}
}

// Benchmark tests for performance validation
func BenchmarkAuthorizationValidation(b *testing.B) {
	auth := &AuthorizationTuple{
		ChainID: testChainID7702,
		Address: codeAddr7702,
		Nonce:   testNonce7702,
	}
	auth.SignAuthorization(testKey7702)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auth.ValidateAuthorization()
	}
}

func BenchmarkAuthorizationSigning(b *testing.B) {
	auth := &AuthorizationTuple{
		ChainID: testChainID7702,
		Address: codeAddr7702,
		Nonce:   testNonce7702,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auth.SignAuthorization(testKey7702)
	}
}

func BenchmarkAuthorizationRecovery(b *testing.B) {
	auth := &AuthorizationTuple{
		ChainID: testChainID7702,
		Address: codeAddr7702,
		Nonce:   testNonce7702,
	}
	auth.SignAuthorization(testKey7702)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auth.RecoverAuthority()
	}
}

func BenchmarkSetCodeTxEncoding(b *testing.B) {
	auth := &AuthorizationTuple{
		ChainID: testChainID7702,
		Address: codeAddr7702,
		Nonce:   testNonce7702,
	}
	auth.SignAuthorization(testKey7702)
	authList := AuthorizationList{*auth}

	tx := NewSetCodeTx(
		testChainID7702,
		testNonce7702,
		&codeAddr7702,
		big.NewInt(1000),
		21000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		AccessList{},
		authList,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx.MarshalBinary()
	}
}

func BenchmarkSetCodeTxDecoding(b *testing.B) {
	auth := &AuthorizationTuple{
		ChainID: testChainID7702,
		Address: codeAddr7702,
		Nonce:   testNonce7702,
	}
	auth.SignAuthorization(testKey7702)
	authList := AuthorizationList{*auth}

	tx := NewSetCodeTx(
		testChainID7702,
		testNonce7702,
		&codeAddr7702,
		big.NewInt(1000),
		21000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		AccessList{},
		authList,
	)

	encoded, _ := tx.MarshalBinary()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decodedTx Transaction
		decodedTx.UnmarshalBinary(encoded)
	}
}

// Helper function to create test authorizations
func createTestAuthorization(key *ecdsa.PrivateKey, chainID *big.Int, addr common.Address, nonce uint64) *AuthorizationTuple {
	auth := &AuthorizationTuple{
		ChainID: chainID,
		Address: addr,
		Nonce:   nonce,
	}
	auth.SignAuthorization(key)
	return auth
}

// Test helper function for creating sample SetCodeTx
func createTestSetCodeTx() *Transaction {
	auth := createTestAuthorization(testKey7702, testChainID7702, codeAddr7702, testNonce7702)
	authList := AuthorizationList{*auth}

	return NewSetCodeTx(
		testChainID7702,
		testNonce7702,
		&codeAddr7702,
		big.NewInt(1000),
		21000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		AccessList{},
		authList,
	)
}