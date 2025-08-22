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
	"github.com/unicornultrafoundation/go-u2u/params"
)

// EIP-7702 processing errors
var (
	ErrInvalidDelegationChain = fmt.Errorf("invalid delegation chain")
	ErrCircularDelegation     = fmt.Errorf("circular delegation detected")
	ErrDelegationDepthLimit   = fmt.Errorf("delegation depth limit exceeded")
	ErrInvalidCodeAddress     = fmt.Errorf("invalid code address for delegation")
	ErrAuthorizationExpired   = fmt.Errorf("authorization has expired")
)

// DelegationChainResolver handles delegation chain resolution for EIP-7702
type DelegationChainResolver struct {
	maxDepth int
	visited  map[common.Address]bool
}

// NewDelegationChainResolver creates a new delegation chain resolver
func NewDelegationChainResolver() *DelegationChainResolver {
	return &DelegationChainResolver{
		maxDepth: 16, // Maximum delegation depth to prevent infinite loops
		visited:  make(map[common.Address]bool),
	}
}

// ResolveDelegationChain resolves the final code address through delegation chain
func (dcr *DelegationChainResolver) ResolveDelegationChain(
	initialAddress common.Address,
	getDelegation func(common.Address) *common.Address,
) (common.Address, error) {
	currentAddress := initialAddress
	depth := 0

	// Reset visited map for each resolution
	for k := range dcr.visited {
		delete(dcr.visited, k)
	}

	for depth < dcr.maxDepth {
		// Check for circular delegation
		if dcr.visited[currentAddress] {
			return common.Address{}, ErrCircularDelegation
		}
		dcr.visited[currentAddress] = true

		// Get delegation for current address
		delegation := getDelegation(currentAddress)
		if delegation == nil {
			// No delegation found, return current address
			return currentAddress, nil
		}

		// Continue with delegation target
		currentAddress = *delegation
		depth++
	}

	return common.Address{}, ErrDelegationDepthLimit
}

// AuthorizationProcessor handles advanced authorization processing for EIP-7702
type AuthorizationProcessor struct {
	chainID *big.Int
}

// NewAuthorizationProcessor creates a new authorization processor
func NewAuthorizationProcessor(chainID *big.Int) *AuthorizationProcessor {
	return &AuthorizationProcessor{
		chainID: chainID,
	}
}

// ProcessAuthorizationList processes and validates an authorization list for transaction execution
func (ap *AuthorizationProcessor) ProcessAuthorizationList(
	authList AuthorizationList,
	currentBlockNumber *big.Int,
	getAccountNonce func(common.Address) uint64,
) (map[common.Address]common.Address, error) {
	if len(authList) == 0 {
		return make(map[common.Address]common.Address), nil
	}

	// Validate the authorization list first
	if err := authList.ValidateAuthorizationList(); err != nil {
		return nil, fmt.Errorf("authorization list validation failed: %w", err)
	}

	delegations := make(map[common.Address]common.Address)
	processedAuthorities := make(map[common.Address]bool)

	for i, auth := range authList {
		// Validate chain ID consistency
		if auth.ChainID.Cmp(ap.chainID) != 0 {
			return nil, fmt.Errorf("authorization %d: chain ID mismatch, expected %v, got %v", 
				i, ap.chainID, auth.ChainID)
		}

		// Recover authority address
		authority, err := auth.RecoverAuthority()
		if err != nil {
			return nil, fmt.Errorf("authorization %d: failed to recover authority: %w", i, err)
		}

		// Check for duplicate authorities (should be caught by ValidateAuthorizationList, but double-check)
		if processedAuthorities[authority] {
			return nil, fmt.Errorf("authorization %d: duplicate authority %s", i, authority.Hex())
		}
		processedAuthorities[authority] = true

		// Validate nonce
		currentNonce := getAccountNonce(authority)
		if auth.Nonce != currentNonce {
			return nil, fmt.Errorf("authorization %d: nonce mismatch for authority %s, expected %d, got %d",
				i, authority.Hex(), currentNonce, auth.Nonce)
		}

		// Add delegation mapping
		delegations[authority] = auth.Address
	}

	return delegations, nil
}

// ValidateAuthorizationForExecution validates authorization for execution context
func (ap *AuthorizationProcessor) ValidateAuthorizationForExecution(
	auth *AuthorizationTuple,
	authority common.Address,
	currentBlockNumber *big.Int,
) error {
	// Basic authorization validation
	if err := auth.ValidateAuthorization(); err != nil {
		return fmt.Errorf("basic validation failed: %w", err)
	}

	// Validate chain ID
	if auth.ChainID.Cmp(ap.chainID) != 0 {
		return fmt.Errorf("chain ID mismatch: expected %v, got %v", ap.chainID, auth.ChainID)
	}

	// Recover and verify authority
	recoveredAuthority, err := auth.RecoverAuthority()
	if err != nil {
		return fmt.Errorf("failed to recover authority: %w", err)
	}

	if recoveredAuthority != authority {
		return fmt.Errorf("authority mismatch: expected %s, recovered %s",
			authority.Hex(), recoveredAuthority.Hex())
	}

	return nil
}

// SetCodeTxProcessor handles SetCode transaction processing logic
type SetCodeTxProcessor struct {
	authProcessor *AuthorizationProcessor
	chainResolver *DelegationChainResolver
}

// NewSetCodeTxProcessor creates a new SetCode transaction processor
func NewSetCodeTxProcessor(chainID *big.Int) *SetCodeTxProcessor {
	return &SetCodeTxProcessor{
		authProcessor: NewAuthorizationProcessor(chainID),
		chainResolver: NewDelegationChainResolver(),
	}
}

// ValidateForPool validates SetCode transaction for transaction pool inclusion
func (stp *SetCodeTxProcessor) ValidateForPool(tx *SetCodeTx, currentChainID *big.Int) error {
	// Validate chain ID
	if tx.ChainID.Cmp(currentChainID) != 0 {
		return fmt.Errorf("chain ID mismatch: transaction has %v, expected %v", tx.ChainID, currentChainID)
	}

	// Validate gas limits
	if tx.Gas == 0 {
		return fmt.Errorf("gas limit cannot be zero")
	}

	// Validate EIP-1559 fee parameters
	if tx.GasFeeCap == nil || tx.GasTipCap == nil {
		return fmt.Errorf("EIP-1559 fee parameters cannot be nil")
	}

	if tx.GasFeeCap.Cmp(tx.GasTipCap) < 0 {
		return fmt.Errorf("gas fee cap (%v) cannot be less than gas tip cap (%v)", 
			tx.GasFeeCap, tx.GasTipCap)
	}

	// Validate authorization list (basic validation)
	if err := tx.AuthorizationList.ValidateAuthorizationList(); err != nil {
		return fmt.Errorf("authorization list validation failed: %w", err)
	}

	// Validate transaction size limits
	if len(tx.Data) > params.MaxCodeSize {
		return fmt.Errorf("transaction data too large: %d bytes (max %d)", len(tx.Data), params.MaxCodeSize)
	}

	return nil
}

// EstimateGas estimates gas consumption for SetCode transaction
func (stp *SetCodeTxProcessor) EstimateGas(tx *SetCodeTx, baseGas uint64) (uint64, error) {
	// Base gas cost
	gas := baseGas

	// Add cost for authorization list processing
	authListGas := stp.calculateAuthorizationListGas(tx.AuthorizationList)
	gas += authListGas

	// Add cost for data
	dataGas := stp.calculateDataGas(tx.Data)
	gas += dataGas

	// Add cost for access list
	accessListGas := stp.calculateAccessListGas(tx.AccessList)
	gas += accessListGas

	return gas, nil
}

// calculateAuthorizationListGas calculates gas cost for processing authorization list
func (stp *SetCodeTxProcessor) calculateAuthorizationListGas(authList AuthorizationList) uint64 {
	// Base cost per authorization for signature verification and processing
	const authorizationBaseGas = 3000
	const authorizationSignatureGas = 3000
	
	return uint64(len(authList)) * (authorizationBaseGas + authorizationSignatureGas)
}

// calculateDataGas calculates gas cost for transaction data
func (stp *SetCodeTxProcessor) calculateDataGas(data []byte) uint64 {
	const (
		txDataNonZeroGas = 16
		txDataZeroGas    = 4
	)

	var gas uint64
	for _, b := range data {
		if b == 0 {
			gas += txDataZeroGas
		} else {
			gas += txDataNonZeroGas
		}
	}
	return gas
}

// calculateAccessListGas calculates gas cost for access list
func (stp *SetCodeTxProcessor) calculateAccessListGas(accessList AccessList) uint64 {
	const (
		accessListAddressGas = 2400
		accessListSlotGas    = 1900
	)

	var gas uint64
	for _, entry := range accessList {
		gas += accessListAddressGas
		gas += uint64(len(entry.StorageKeys)) * accessListSlotGas
	}
	return gas
}

// ProcessDelegations processes delegations for a SetCode transaction
func (stp *SetCodeTxProcessor) ProcessDelegations(
	tx *SetCodeTx,
	currentBlockNumber *big.Int,
	getAccountNonce func(common.Address) uint64,
) (map[common.Address]common.Address, error) {
	return stp.authProcessor.ProcessAuthorizationList(
		tx.AuthorizationList,
		currentBlockNumber,
		getAccountNonce,
	)
}

// SetCodeTxGasCalculator provides gas calculation utilities for SetCode transactions
type SetCodeTxGasCalculator struct {
	processor *SetCodeTxProcessor
}

// NewSetCodeTxGasCalculator creates a new gas calculator for SetCode transactions
func NewSetCodeTxGasCalculator(chainID *big.Int) *SetCodeTxGasCalculator {
	return &SetCodeTxGasCalculator{
		processor: NewSetCodeTxProcessor(chainID),
	}
}

// IntrinsicGas calculates the intrinsic gas cost for a SetCode transaction
func (gc *SetCodeTxGasCalculator) IntrinsicGas(tx *SetCodeTx) (uint64, error) {
	// Base transaction gas
	const txGas = 21000

	// Start with base gas
	gas := uint64(txGas)

	// Add gas for authorization list
	authGas := gc.processor.calculateAuthorizationListGas(tx.AuthorizationList)
	gas += authGas

	// Add gas for data
	dataGas := gc.processor.calculateDataGas(tx.Data)
	gas += dataGas

	// Add gas for access list
	accessListGas := gc.processor.calculateAccessListGas(tx.AccessList)
	gas += accessListGas

	// Add gas for contract creation if To is nil
	if tx.To == nil {
		const contractCreationGas = 32000
		gas += contractCreationGas
	}

	return gas, nil
}

// ValidateGasLimit validates that the transaction gas limit is sufficient
func (gc *SetCodeTxGasCalculator) ValidateGasLimit(tx *SetCodeTx) error {
	intrinsicGas, err := gc.IntrinsicGas(tx)
	if err != nil {
		return fmt.Errorf("failed to calculate intrinsic gas: %w", err)
	}

	if tx.Gas < intrinsicGas {
		return fmt.Errorf("gas limit %d is below intrinsic gas requirement %d", 
			tx.Gas, intrinsicGas)
	}

	return nil
}