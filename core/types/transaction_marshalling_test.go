// Copyright 2024 The go-u2u Authors
// This file is part of the go-u2u library.

package types

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
)

// TestAuthorizationTupleJSONMarshalling tests JSON marshalling and unmarshalling of AuthorizationTuple
func TestAuthorizationTupleJSONMarshalling(t *testing.T) {
	auth := AuthorizationTuple{
		ChainID: big.NewInt(1),
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Nonce:   42,
		V:       big.NewInt(27),
		R:       big.NewInt(12345),
		S:       big.NewInt(67890),
	}

	// Test marshalling
	jsonData, err := json.Marshal(auth)
	if err != nil {
		t.Fatalf("Failed to marshal AuthorizationTuple: %v", err)
	}

	t.Logf("AuthorizationTuple JSON: %s", string(jsonData))

	// Test unmarshalling
	var unmarshalled AuthorizationTuple
	if err := json.Unmarshal(jsonData, &unmarshalled); err != nil {
		t.Fatalf("Failed to unmarshal AuthorizationTuple: %v", err)
	}

	// Verify fields
	if auth.ChainID.Cmp(unmarshalled.ChainID) != 0 {
		t.Errorf("ChainID mismatch: expected %v, got %v", auth.ChainID, unmarshalled.ChainID)
	}
	if auth.Address != unmarshalled.Address {
		t.Errorf("Address mismatch: expected %v, got %v", auth.Address, unmarshalled.Address)
	}
	if auth.Nonce != unmarshalled.Nonce {
		t.Errorf("Nonce mismatch: expected %v, got %v", auth.Nonce, unmarshalled.Nonce)
	}
	if auth.V.Cmp(unmarshalled.V) != 0 {
		t.Errorf("V mismatch: expected %v, got %v", auth.V, unmarshalled.V)
	}
	if auth.R.Cmp(unmarshalled.R) != 0 {
		t.Errorf("R mismatch: expected %v, got %v", auth.R, unmarshalled.R)
	}
	if auth.S.Cmp(unmarshalled.S) != 0 {
		t.Errorf("S mismatch: expected %v, got %v", auth.S, unmarshalled.S)
	}
}

// TestSetCodeTxJSONStructure tests that SetCodeTx can be marshalled without errors
func TestSetCodeTxJSONStructure(t *testing.T) {
	// Test that SetCodeTx struct can be marshalled to JSON
	to := common.HexToAddress("0x1234567890123456789012345678901234567890")
	authList := AuthorizationList{
		{
			ChainID: big.NewInt(1),
			Address: common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdef"),
			Nonce:   99,
			V:       big.NewInt(28),
			R:       big.NewInt(54321),
			S:       big.NewInt(98765),
		},
	}

	setCodeTx := &SetCodeTx{
		ChainID:           big.NewInt(1),
		Nonce:             10,
		GasTipCap:         big.NewInt(1000000000),
		GasFeeCap:         big.NewInt(2000000000),
		Gas:               21000,
		To:                &to,
		Value:             big.NewInt(0),
		Data:              []byte("test"),
		AccessList:        AccessList{},
		AuthorizationList: authList,
		V:                 big.NewInt(27),
		R:                 big.NewInt(1),
		S:                 big.NewInt(1),
	}

	// Test direct marshalling of SetCodeTx
	jsonData, err := json.Marshal(setCodeTx)
	if err != nil {
		t.Fatalf("Failed to marshal SetCodeTx: %v", err)
	}

	t.Logf("SetCodeTx JSON: %s", string(jsonData))

	// Verify it contains the expected fields
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// Check that authorization list is present
	if authListData, exists := jsonMap["AuthorizationList"]; !exists {
		t.Error("Authorization list not found in JSON")
	} else {
		authArray, ok := authListData.([]interface{})
		if !ok {
			t.Error("Authorization list is not an array")
		} else if len(authArray) != 1 {
			t.Errorf("Expected 1 authorization, got %d", len(authArray))
		}
	}
}
