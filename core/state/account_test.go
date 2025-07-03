package state

import (
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
)

func TestAccountCmp(t *testing.T) {
	// Test data
	hash1 := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	hash2 := common.HexToHash("0x2234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	codeHash1 := []byte{0x01, 0x02, 0x03}
	codeHash2 := []byte{0x01, 0x02, 0x04}
	balance1 := big.NewInt(100)
	balance2 := big.NewInt(200)

	tests := []struct {
		name     string
		a        *Account
		b        *Account
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "a nil, b not nil",
			a:        nil,
			b:        &Account{Nonce: 1},
			expected: false,
		},
		{
			name:     "a not nil, b nil",
			a:        &Account{Nonce: 1},
			b:        nil,
			expected: false,
		},
		{
			name: "identical accounts",
			a: &Account{
				Nonce:    1,
				Balance:  big.NewInt(100),
				Root:     hash1,
				CodeHash: codeHash1,
			},
			b: &Account{
				Nonce:    1,
				Balance:  big.NewInt(100),
				Root:     hash1,
				CodeHash: codeHash1,
			},
			expected: true,
		},
		{
			name: "different nonce - a < b",
			a: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			b: &Account{
				Nonce:    2,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			expected: false,
		},
		{
			name: "different nonce - a > b",
			a: &Account{
				Nonce:    2,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			b: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			expected: false,
		},
		{
			name: "different balance - a < b",
			a: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			b: &Account{
				Nonce:    1,
				Balance:  balance2,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			expected: false,
		},
		{
			name: "different balance - a > b",
			a: &Account{
				Nonce:    1,
				Balance:  balance2,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			b: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			expected: false,
		},
		{
			name: "nil balance in a",
			a: &Account{
				Nonce:    1,
				Balance:  nil,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			b: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			expected: false,
		},
		{
			name: "nil balance in b",
			a: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			b: &Account{
				Nonce:    1,
				Balance:  nil,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			expected: false,
		},
		{
			name: "both balances nil",
			a: &Account{
				Nonce:    1,
				Balance:  nil,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			b: &Account{
				Nonce:    1,
				Balance:  nil,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			expected: true,
		},
		{
			name: "different root hash - a < b",
			a: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			b: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash2,
				CodeHash: codeHash1,
			},
			expected: false,
		},
		{
			name: "different root hash - a > b",
			a: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash2,
				CodeHash: codeHash1,
			},
			b: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			expected: false,
		},
		{
			name: "different code hash - a < b",
			a: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			b: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash2,
			},
			expected: false,
		},
		{
			name: "different code hash - a > b",
			a: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash2,
			},
			b: &Account{
				Nonce:    1,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Cmp(tt.b)
			if result != tt.expected {
				t.Errorf("Account.Cmp() = %t, expected %t", result, tt.expected)
			}
		})
	}
}

func TestAccountString(t *testing.T) {
	hash1 := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	codeHash1 := []byte{0x01, 0x02, 0x03}
	balance1 := big.NewInt(12345)

	tests := []struct {
		name     string
		account  *Account
		expected string
	}{
		{
			name:     "nil account",
			account:  nil,
			expected: "Account(nil)",
		},
		{
			name: "account with nil balance",
			account: &Account{
				Nonce:    42,
				Balance:  nil,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			expected: "Account{Nonce: 42, Balance: <nil>, Root: 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef, CodeHash: 010203}",
		},
		{
			name: "account with all fields",
			account: &Account{
				Nonce:    100,
				Balance:  balance1,
				Root:     hash1,
				CodeHash: codeHash1,
			},
			expected: "Account{Nonce: 100, Balance: 12345, Root: 0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef, CodeHash: 010203}",
		},
		{
			name: "account with zero balance",
			account: &Account{
				Nonce:    0,
				Balance:  big.NewInt(0),
				Root:     common.Hash{},
				CodeHash: []byte{},
			},
			expected: "Account{Nonce: 0, Balance: 0, Root: 0x0000000000000000000000000000000000000000000000000000000000000000, CodeHash: }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.account.String()
			if result != tt.expected {
				t.Errorf("Account.String() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
