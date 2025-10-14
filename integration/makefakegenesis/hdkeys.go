package makefakegenesis

import (
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/sha512"
	"fmt"

	"github.com/tyler-smith/go-bip39"
	"github.com/unicornultrafoundation/go-helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/accounts"
	"github.com/unicornultrafoundation/go-u2u/crypto"
)

// DeriveKeyFromMnemonic derives a validator key from a mnemonic using BIP44 derivation path.
// Uses the path m/44'/60'/0'/0/validatorIndex where validatorIndex = validatorID - 1
// If mnemonic is empty, falls back to FakeKey for backward compatibility.
func DeriveKeyFromMnemonic(mnemonic string, validatorID idx.ValidatorID) (*ecdsa.PrivateKey, error) {
	// Fallback to deterministic keys if no mnemonic provided
	if mnemonic == "" {
		return FakeKey(validatorID), nil
	}

	// Validate mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic phrase")
	}

	// Generate seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "")

	// Derive key using standard Ethereum path: m/44'/60'/0'/0/index
	// validatorID starts at 1, but derivation index starts at 0
	derivationPath := accounts.DefaultBaseDerivationPath
	derivationPath[4] = uint32(validatorID - 1) // Set the address_index

	// Derive the key
	key, err := derivePrivateKey(seed, derivationPath)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key for validator %d: %w", validatorID, err)
	}

	return key, nil
}

// derivePrivateKey derives a private key from seed and derivation path
func derivePrivateKey(seed []byte, path accounts.DerivationPath) (*ecdsa.PrivateKey, error) {
	// Use the crypto package's HD wallet derivation
	// Based on go-ethereum's implementation
	masterKey, err := hdKeyFromSeed(seed)
	if err != nil {
		return nil, err
	}

	key := masterKey
	for _, n := range path {
		key, err = key.derive(n)
		if err != nil {
			return nil, err
		}
	}

	privateKey, err := crypto.ToECDSA(key.Key)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// hdKey represents a hierarchical deterministic key
type hdKey struct {
	Key       []byte
	ChainCode []byte
}

// hdKeyFromSeed creates a master key from seed
func hdKeyFromSeed(seed []byte) (*hdKey, error) {
	// Implementation based on BIP32
	// Use HMAC-SHA512 with key "Bitcoin seed"
	mac := hmac.New(sha512.New, []byte("Bitcoin seed"))
	mac.Write(seed)
	i := mac.Sum(nil)

	return &hdKey{
		Key:       i[:32],
		ChainCode: i[32:],
	}, nil
}

// derive derives a child key from parent
func (k *hdKey) derive(i uint32) (*hdKey, error) {
	// Implementation based on BIP32
	var data []byte
	if i >= 0x80000000 { // Hardened
		data = append([]byte{0x0}, k.Key...)
	} else {
		// For non-hardened, we need the public key
		privateKey, err := crypto.ToECDSA(k.Key)
		if err != nil {
			return nil, err
		}
		pubkey := crypto.CompressPubkey(&privateKey.PublicKey)
		data = pubkey
	}

	data = append(data, byte(i>>24), byte(i>>16), byte(i>>8), byte(i))

	mac := hmac.New(sha512.New, k.ChainCode)
	mac.Write(data)
	I := mac.Sum(nil)

	childKey := &hdKey{
		Key:       I[:32],
		ChainCode: I[32:],
	}

	return childKey, nil
}

// GenerateMnemonic generates a new random 24-word mnemonic for testing
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(256) // 256 bits = 24 words
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}
