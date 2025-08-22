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

	"github.com/unicornultrafoundation/go-u2u/common"
)

// EIP-7702 Constants as per specification
const (
	// MaxAuthorizationListSize limits the number of authorizations per transaction
	MaxAuthorizationListSize = 256
	
	// MaxDelegationDepth prevents infinite delegation loops
	MaxDelegationDepth = 16
	
	// AuthorizationBaseGas is the base gas cost for each authorization
	AuthorizationBaseGas = 3000
	
	// AuthorizationSignatureGas is the additional gas cost for signature verification
	AuthorizationSignatureGas = 3000
	
	// DelegationStorageAddress is the system address where delegations are stored
	DelegationStorageAddress = "0x7702000000000000000000000000000000000000"
	
	// DelegationKeyPrefix is the prefix for delegation storage keys
	DelegationKeyPrefix = "EIP7702_DELEGATION_"
)

// EIP-7702 validation errors
var (
	ErrAuthorizationListTooLarge   = fmt.Errorf("authorization list exceeds maximum size of %d", MaxAuthorizationListSize)
	ErrInvalidAuthorizationChainID = fmt.Errorf("invalid chain ID in authorization")
	ErrInvalidAuthorizationNonce   = fmt.Errorf("invalid nonce in authorization")
	ErrInvalidAuthorizationAddress = fmt.Errorf("invalid address in authorization")
	ErrZeroSignatureValue          = fmt.Errorf("signature R or S value is zero")
	ErrSetCodeTxInvalidChainID     = fmt.Errorf("SetCode transaction chain ID mismatch")
	ErrSetCodeTxInvalidGasParams   = fmt.Errorf("SetCode transaction invalid gas parameters")
)

// EIP7702Validator provides validation functions for EIP-7702 compliance
type EIP7702Validator struct {
	chainID *big.Int
}

// NewEIP7702Validator creates a new EIP-7702 validator
func NewEIP7702Validator(chainID *big.Int) *EIP7702Validator {
	return &EIP7702Validator{
		chainID: chainID,
	}
}

// ValidateSetCodeTransaction validates a SetCode transaction for EIP-7702 compliance
func (v *EIP7702Validator) ValidateSetCodeTransaction(tx *SetCodeTx) error {
	// Validate chain ID
	if tx.ChainID == nil || tx.ChainID.Cmp(v.chainID) != 0 {
		return ErrSetCodeTxInvalidChainID
	}

	// Validate gas parameters
	if err := v.validateGasParameters(tx); err != nil {
		return err
	}

	// Validate authorization list
	if err := v.validateAuthorizationList(tx.AuthorizationList); err != nil {
		return err
	}

	// Validate each authorization
	for i, auth := range tx.AuthorizationList {
		if err := v.validateAuthorization(&auth, i); err != nil {
			return fmt.Errorf("authorization %d: %w", i, err)
		}
	}

	return nil
}

// validateGasParameters validates gas-related parameters
func (v *EIP7702Validator) validateGasParameters(tx *SetCodeTx) error {
	// Gas must be positive
	if tx.Gas == 0 {
		return fmt.Errorf("gas limit cannot be zero")
	}

	// Gas fee cap must be positive
	if tx.GasFeeCap == nil || tx.GasFeeCap.Sign() <= 0 {
		return fmt.Errorf("gas fee cap must be positive")
	}

	// Gas tip cap must be positive
	if tx.GasTipCap == nil || tx.GasTipCap.Sign() <= 0 {
		return fmt.Errorf("gas tip cap must be positive")
	}

	// Gas fee cap must be >= gas tip cap
	if tx.GasFeeCap.Cmp(tx.GasTipCap) < 0 {
		return fmt.Errorf("gas fee cap %v less than gas tip cap %v", tx.GasFeeCap, tx.GasTipCap)
	}

	// Validate intrinsic gas
	intrinsicGas := v.calculateIntrinsicGas(tx)
	if tx.Gas < intrinsicGas {
		return fmt.Errorf("gas limit %d below intrinsic gas %d", tx.Gas, intrinsicGas)
	}

	return nil
}

// validateAuthorizationList validates the authorization list structure
func (v *EIP7702Validator) validateAuthorizationList(authList AuthorizationList) error {
	// Check maximum size
	if len(authList) > MaxAuthorizationListSize {
		return ErrAuthorizationListTooLarge
	}

	// Check for duplicates
	seen := make(map[common.Address]bool)
	for i, auth := range authList {
		// Recover authority address for duplicate checking
		authority, err := auth.RecoverAuthority()
		if err != nil {
			return fmt.Errorf("authorization %d: failed to recover authority: %w", i, err)
		}

		if seen[authority] {
			return fmt.Errorf("authorization %d: duplicate authorization for authority %s", i, authority.Hex())
		}
		seen[authority] = true
	}

	return nil
}

// validateAuthorization validates a single authorization tuple
func (v *EIP7702Validator) validateAuthorization(auth *AuthorizationTuple, index int) error {
	// Validate chain ID
	if auth.ChainID == nil || auth.ChainID.Cmp(v.chainID) != 0 {
		return fmt.Errorf("%w: expected %v, got %v", ErrInvalidAuthorizationChainID, v.chainID, auth.ChainID)
	}

	// Validate address (should not be zero address)
	if auth.Address == (common.Address{}) {
		return ErrInvalidAuthorizationAddress
	}

	// Validate signature components
	if auth.V == nil || auth.R == nil || auth.S == nil {
		return fmt.Errorf("missing signature components")
	}

	// Check for zero R or S values
	if auth.R.Sign() == 0 || auth.S.Sign() == 0 {
		return ErrZeroSignatureValue
	}

	// Validate signature format
	if err := auth.ValidateAuthorization(); err != nil {
		return fmt.Errorf("invalid authorization signature: %w", err)
	}

	return nil
}

// calculateIntrinsicGas calculates the intrinsic gas cost for a SetCode transaction
func (v *EIP7702Validator) calculateIntrinsicGas(tx *SetCodeTx) uint64 {
	// Base transaction cost
	gas := uint64(21000)

	// Data cost
	if len(tx.Data) > 0 {
		nz := 0
		for _, b := range tx.Data {
			if b != 0 {
				nz++
			}
		}
		gas += uint64(nz)*16 + uint64(len(tx.Data)-nz)*4
	}

	// Access list cost
	gas += uint64(len(tx.AccessList)) * 2400
	for _, al := range tx.AccessList {
		gas += uint64(len(al.StorageKeys)) * 1900
	}

	// Authorization list cost
	gas += uint64(len(tx.AuthorizationList)) * (AuthorizationBaseGas + AuthorizationSignatureGas)

	return gas
}

// ValidateAuthorizationSignature validates an authorization signature independently
func ValidateAuthorizationSignature(auth *AuthorizationTuple, chainID *big.Int) error {
	validator := NewEIP7702Validator(chainID)
	return validator.validateAuthorization(auth, 0)
}

// ValidateAuthorizationListSize validates the authorization list size limit
func ValidateAuthorizationListSize(authList AuthorizationList) error {
	if len(authList) > MaxAuthorizationListSize {
		return ErrAuthorizationListTooLarge
	}
	return nil
}

// ValidateDelegationDepth validates that delegation depth doesn't exceed limits
func ValidateDelegationDepth(depth int) error {
	if depth > MaxDelegationDepth {
		return fmt.Errorf("delegation depth %d exceeds maximum %d", depth, MaxDelegationDepth)
	}
	return nil
}

// IsValidDelegationAddress checks if an address can be used for delegation
func IsValidDelegationAddress(addr common.Address) bool {
	// Zero address is not valid for delegation
	if addr == (common.Address{}) {
		return false
	}

	// System addresses (0x1-0xff) should not be used for delegation
	if addr[0] == 0 && addr[1] == 0 && addr[2] == 0 && addr[3] == 0 &&
		addr[4] == 0 && addr[5] == 0 && addr[6] == 0 && addr[7] == 0 &&
		addr[8] == 0 && addr[9] == 0 && addr[10] == 0 && addr[11] == 0 &&
		addr[12] == 0 && addr[13] == 0 && addr[14] == 0 && addr[15] == 0 &&
		addr[16] == 0 && addr[17] == 0 && addr[18] == 0 && addr[19] <= 255 {
		return false
	}

	return true
}

// GetDelegationStorageKey returns the storage key for a delegation mapping
func GetDelegationStorageKey(authority common.Address) common.Hash {
	return common.BytesToHash(append([]byte(DelegationKeyPrefix), authority.Bytes()...))
}

// GetDelegationStorageAddress returns the system address for delegation storage
func GetDelegationStorageAddress() common.Address {
	return common.HexToAddress(DelegationStorageAddress)
}

// EIP7702Compliance contains compliance check functions
type EIP7702Compliance struct{}

// CheckTransactionCompliance performs comprehensive EIP-7702 compliance check
func (c *EIP7702Compliance) CheckTransactionCompliance(tx *SetCodeTx, chainID *big.Int) []error {
	var errors []error
	validator := NewEIP7702Validator(chainID)

	// Validate transaction
	if err := validator.ValidateSetCodeTransaction(tx); err != nil {
		errors = append(errors, err)
	}

	// Check authorization list size
	if err := ValidateAuthorizationListSize(tx.AuthorizationList); err != nil {
		errors = append(errors, err)
	}

	// Check delegation addresses
	for i, auth := range tx.AuthorizationList {
		if !IsValidDelegationAddress(auth.Address) {
			errors = append(errors, fmt.Errorf("authorization %d: invalid delegation address %s", i, auth.Address.Hex()))
		}
	}

	return errors
}

// EstimateGas estimates the gas requirement for a SetCode transaction
func (c *EIP7702Compliance) EstimateGas(tx *SetCodeTx) uint64 {
	validator := NewEIP7702Validator(tx.ChainID)
	return validator.calculateIntrinsicGas(tx)
}