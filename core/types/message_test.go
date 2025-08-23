package types

import (
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
)

// TestMessageSetCodeAuthorizations tests the new SetCodeAuthorizations method.
func TestMessageSetCodeAuthorizations(t *testing.T) {
	// Create test authorization list
	auth1 := AuthorizationTuple{
		ChainID: big.NewInt(1),
		Address: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Nonce:   1,
		V:       big.NewInt(27),
		R:       big.NewInt(12345),
		S:       big.NewInt(67890),
	}
	
	authList := AuthorizationList{auth1}
	
	// Create message with authorization list
	msg := NewMessage(
		common.HexToAddress("0xabcd"),
		&common.Address{},
		0,
		big.NewInt(1000),
		21000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		big.NewInt(1000000000),
		[]byte{},
		AccessList{},
		authList,
		false,
	)
	
	// Test SetCodeAuthorizations method
	retrievedAuthList := msg.SetCodeAuthorizations()
	if len(retrievedAuthList) != 1 {
		t.Errorf("Expected 1 authorization, got %d", len(retrievedAuthList))
	}
	
	if retrievedAuthList[0].Address != auth1.Address {
		t.Errorf("Authorization address mismatch")
	}
	
	if retrievedAuthList[0].ChainID.Cmp(auth1.ChainID) != 0 {
		t.Errorf("Authorization ChainID mismatch")
	}
}

// TestMessageSetCodeAuthorizationsEmpty tests the method with empty authorization list.
func TestMessageSetCodeAuthorizationsEmpty(t *testing.T) {
	// Create message without authorization list
	msg := NewMessage(
		common.HexToAddress("0xabcd"),
		&common.Address{},
		0,
		big.NewInt(1000),
		21000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		big.NewInt(1000000000),
		[]byte{},
		AccessList{},
		nil, // Empty authorization list
		false,
	)
	
	// Test SetCodeAuthorizations method
	retrievedAuthList := msg.SetCodeAuthorizations()
	if retrievedAuthList != nil {
		t.Errorf("Expected nil authorization list, got %v", retrievedAuthList)
	}
}
