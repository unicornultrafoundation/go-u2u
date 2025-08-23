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
	"errors"
	"math/big"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/common/hexutil"
	"github.com/unicornultrafoundation/go-u2u/crypto"
)

//go:generate gencodec -type AuthorizationTuple -out gen_authorization_tuple.go
//go:generate gencodec -type SetCodeTx -field-override setCodeTxMarshaling -out gen_setcode_tx.go

// AuthorizationTuple represents a single authorization in EIP-7702.
// It contains the authority address, chain ID, nonce, and signature.
type AuthorizationTuple struct {
	ChainID   *big.Int        `json:"chainId"   gencodec:"required"`
	Address   common.Address  `json:"address"   gencodec:"required"`
	Nonce     uint64          `json:"nonce"     gencodec:"required"`
	V         *big.Int        `json:"v"         gencodec:"required"`
	R         *big.Int        `json:"r"         gencodec:"required"`
	S         *big.Int        `json:"s"         gencodec:"required"`
}

// AuthorizationList is a list of authorization tuples for EIP-7702.
type AuthorizationList []AuthorizationTuple

// SetCodeTx is the data of EIP-7702 set code transactions.
// It allows EOAs to delegate their account behavior to smart contracts.
type SetCodeTx struct {
	ChainID           *big.Int            // destination chain ID
	Nonce             uint64              // nonce of sender account
	GasTipCap         *big.Int            // a.k.a. maxPriorityFeePerGas
	GasFeeCap         *big.Int            // a.k.a. maxFeePerGas
	Gas               uint64              // gas limit
	To                *common.Address     `rlp:"nil"` // nil means contract creation
	Value             *big.Int            // wei amount
	Data              []byte              // contract invocation input data
	AccessList        AccessList          // EIP-2930 access list
	AuthorizationList AuthorizationList   // EIP-7702 authorization list

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`
}

// copy creates a deep copy of the transaction data and initializes all fields.
func (tx *SetCodeTx) copy() TxData {
	cpy := &SetCodeTx{
		Nonce: tx.Nonce,
		To:    tx.To, // TODO: copy pointed-to address
		Data:  common.CopyBytes(tx.Data),
		Gas:   tx.Gas,
		// These are copied below.
		AccessList:        make(AccessList, len(tx.AccessList)),
		AuthorizationList: make(AuthorizationList, len(tx.AuthorizationList)),
		Value:             new(big.Int),
		ChainID:           new(big.Int),
		GasTipCap:         new(big.Int),
		GasFeeCap:         new(big.Int),
		V:                 new(big.Int),
		R:                 new(big.Int),
		S:                 new(big.Int),
	}
	copy(cpy.AccessList, tx.AccessList)
	copy(cpy.AuthorizationList, tx.AuthorizationList)
	
	if tx.Value != nil {
		cpy.Value.Set(tx.Value)
	}
	if tx.ChainID != nil {
		cpy.ChainID.Set(tx.ChainID)
	}
	if tx.GasTipCap != nil {
		cpy.GasTipCap.Set(tx.GasTipCap)
	}
	if tx.GasFeeCap != nil {
		cpy.GasFeeCap.Set(tx.GasFeeCap)
	}
	if tx.V != nil {
		cpy.V.Set(tx.V)
	}
	if tx.R != nil {
		cpy.R.Set(tx.R)
	}
	if tx.S != nil {
		cpy.S.Set(tx.S)
	}
	return cpy
}

// txType returns the transaction type.
func (tx *SetCodeTx) txType() byte {
	return SetCodeTxType
}

// chainID returns the destination chain ID.
func (tx *SetCodeTx) chainID() *big.Int {
	return tx.ChainID
}

// accessList returns the access list.
func (tx *SetCodeTx) accessList() AccessList {
	return tx.AccessList
}

// data returns the contract invocation input data.
func (tx *SetCodeTx) data() []byte {
	return tx.Data
}

// gas returns the gas limit.
func (tx *SetCodeTx) gas() uint64 {
	return tx.Gas
}

// gasPrice returns the gas price, which for EIP-1559 transactions is the gas fee cap.
func (tx *SetCodeTx) gasPrice() *big.Int {
	return tx.GasFeeCap
}

// gasTipCap returns the gas tip cap.
func (tx *SetCodeTx) gasTipCap() *big.Int {
	return tx.GasTipCap
}

// gasFeeCap returns the gas fee cap.
func (tx *SetCodeTx) gasFeeCap() *big.Int {
	return tx.GasFeeCap
}

// value returns the wei amount.
func (tx *SetCodeTx) value() *big.Int {
	return tx.Value
}

// nonce returns the sender account nonce.
func (tx *SetCodeTx) nonce() uint64 {
	return tx.Nonce
}

// to returns the recipient address of the transaction.
func (tx *SetCodeTx) to() *common.Address {
	return tx.To
}

// rawSignatureValues returns the signature values.
func (tx *SetCodeTx) rawSignatureValues() (v, r, s *big.Int) {
	return tx.V, tx.R, tx.S
}

// setSignatureValues sets the signature values.
func (tx *SetCodeTx) setSignatureValues(chainID, v, r, s *big.Int) {
	tx.ChainID, tx.V, tx.R, tx.S = chainID, v, r, s
}

// authorizationList returns the authorization list for EIP-7702.
func (tx *SetCodeTx) authorizationList() AuthorizationList {
	return tx.AuthorizationList
}

// NewSetCodeTx creates a new EIP-7702 transaction.
func NewSetCodeTx(chainID *big.Int, nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasTipCap, gasFeeCap *big.Int, data []byte, accessList AccessList, authList AuthorizationList) *Transaction {
	return NewTx(&SetCodeTx{
		ChainID:           chainID,
		Nonce:             nonce,
		To:                to,
		Value:             amount,
		Gas:               gasLimit,
		GasTipCap:         gasTipCap,
		GasFeeCap:         gasFeeCap,
		Data:              data,
		AccessList:        accessList,
		AuthorizationList: authList,
		V:                 new(big.Int),
		R:                 new(big.Int),
		S:                 new(big.Int),
	})
}

// EIP-7702 validation errors
var (
	ErrInvalidAuthorization     = errors.New("invalid authorization")
	ErrInvalidAuthorizationSig  = errors.New("invalid authorization signature")
	ErrInvalidAuthChainID       = errors.New("invalid authorization chain ID")
	ErrInvalidAuthNonce         = errors.New("invalid authorization nonce")
	ErrDuplicateAuthorization   = errors.New("duplicate authorization")
)

// ValidateAuthorization validates a single authorization tuple according to EIP-7702.
func (auth *AuthorizationTuple) ValidateAuthorization() error {
	// Basic validation: check if required fields are present
	if auth.ChainID == nil {
		return ErrInvalidAuthChainID
	}
	if auth.V == nil || auth.R == nil || auth.S == nil {
		return ErrInvalidAuthorizationSig
	}
	
	// Validate signature values are in valid range
	if auth.V.Sign() < 0 || auth.R.Sign() < 0 || auth.S.Sign() < 0 {
		return ErrInvalidAuthorizationSig
	}
	
	// Check if R and S are valid (not zero and within valid range)
	if auth.R.Sign() == 0 || auth.S.Sign() == 0 {
		return ErrInvalidAuthorizationSig
	}
	
	return nil
}

// RecoverAuthority recovers the authority address from the authorization signature.
func (auth *AuthorizationTuple) RecoverAuthority() (common.Address, error) {
	// Validate authorization first
	if err := auth.ValidateAuthorization(); err != nil {
		return common.Address{}, err
	}
	
	// Create the authorization message hash according to EIP-7702
	// The message format is: keccak256(MAGIC || rlp([chainId, address, nonce]))
	// where MAGIC = 0x05 (EIP-7702 magic byte)
	magicByte := byte(0x05)
	
	// Create authorization payload for signing
	payload := []interface{}{
		auth.ChainID,
		auth.Address,
		auth.Nonce,
	}
	
	// RLP encode the payload
	hash := rlpHash([]interface{}{magicByte, payload})
	
	// Recover the public key from the signature
	pubKey, err := crypto.SigToPub(hash.Bytes(), auth.signatureBytes())
	if err != nil {
		return common.Address{}, ErrInvalidAuthorizationSig
	}
	
	// Derive address from public key
	return crypto.PubkeyToAddress(*pubKey), nil
}

// signatureBytes returns the authorization signature as a byte slice.
func (auth *AuthorizationTuple) signatureBytes() []byte {
	// Convert V, R, S to signature bytes format
	sig := make([]byte, 65)
	copy(sig[32-len(auth.R.Bytes()):32], auth.R.Bytes())
	copy(sig[64-len(auth.S.Bytes()):64], auth.S.Bytes())
	sig[64] = byte(auth.V.Uint64())
	return sig
}

// ValidateAuthorizationList validates all authorizations in the list.
func (authList AuthorizationList) ValidateAuthorizationList() error {
	// Check for empty list (this might be allowed depending on use case)
	if len(authList) == 0 {
		return nil // Empty list is valid
	}
	
	// Track seen authorities to prevent duplicates
	seen := make(map[common.Address]bool)
	
	for i, auth := range authList {
		// Validate individual authorization
		if err := auth.ValidateAuthorization(); err != nil {
			return err
		}
		
		// Recover authority address
		authority, err := auth.RecoverAuthority()
		if err != nil {
			return err
		}
		
		// Check for duplicate authorities
		if seen[authority] {
			return ErrDuplicateAuthorization
		}
		seen[authority] = true
		
		// Additional validation: ensure authorization is for the correct authority
		// The authority derived from signature should match what's expected
		_ = i // Use i if needed for detailed error reporting
	}
	
	return nil
}

// SignAuthorization signs an authorization tuple with the given private key.
func (auth *AuthorizationTuple) SignAuthorization(privateKey *ecdsa.PrivateKey) error {
	// Create the authorization message hash
	magicByte := byte(0x05)
	payload := []interface{}{
		auth.ChainID,
		auth.Address,
		auth.Nonce,
	}
	
	hash := rlpHash([]interface{}{magicByte, payload})
	
	// Sign the hash
	sig, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return err
	}
	
	// Extract V, R, S from signature
	auth.R = new(big.Int).SetBytes(sig[:32])
	auth.S = new(big.Int).SetBytes(sig[32:64])
	auth.V = new(big.Int).SetUint64(uint64(sig[64]))
	
	return nil
}

// setCodeTxMarshaling provides field type overrides for gencodec.
type setCodeTxMarshaling struct {
	ChainID           *hexutil.Big
	Nonce             hexutil.Uint64
	GasTipCap         *hexutil.Big
	GasFeeCap         *hexutil.Big
	Gas               hexutil.Uint64
	Value             *hexutil.Big
	Data              hexutil.Bytes
	V                 *hexutil.Big
	R                 *hexutil.Big
	S                 *hexutil.Big
}
// EIP-7702 delegation constants and helper functions

// DelegationPrefix is used by code to denote the account is delegating to another account.
var DelegationPrefix = []byte{0xef, 0x01, 0x00}

// ParseDelegation tries to parse the address from a delegation slice.
// Returns the delegated address and true if the code is a valid delegation, false otherwise.
func ParseDelegation(code []byte) (common.Address, bool) {
	if len(code) != 23 {
		return common.Address{}, false
	}
	// Check delegation prefix
	if code[0] != 0xef || code[1] != 0x01 || code[2] != 0x00 {
		return common.Address{}, false
	}
	return common.BytesToAddress(code[3:]), true
}

// AddressToDelegation adds the delegation prefix to the specified address.
// This creates the delegation code that should be stored at the authority's address.
func AddressToDelegation(addr common.Address) []byte {
	result := make([]byte, 23)
	copy(result[:3], DelegationPrefix)
	copy(result[3:], addr.Bytes())
	return result
}
