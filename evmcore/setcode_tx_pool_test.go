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
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/crypto"
	"github.com/unicornultrafoundation/go-u2u/params"
)

func TestSetCodeTxPoolValidator(t *testing.T) {
	chainID := big.NewInt(1)
	validator := NewSetCodeTxPoolValidator(chainID)

	// Create test transaction
	key, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(key.PublicKey)
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	auth := &types.AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	tx := types.NewSetCodeTx(
		chainID,
		42,
		&codeAddr,
		big.NewInt(1000),
		50000,
		big.NewInt(1000000000), // gasTipCap
		big.NewInt(2000000000), // gasFeeCap
		[]byte("test data"),
		types.AccessList{},
		types.AuthorizationList{*auth},
	)

	// Create mock state
	statedb := state.NewDatabase(nil)
	state, _ := state.New(common.Hash{}, statedb, nil)
	
	// Set up account with sufficient balance
	state.SetBalance(from, big.NewInt(1000000000000000000)) // 1 ETH
	state.SetNonce(from, 42)

	// Create head
	head := &EvmHeader{
		Number:   big.NewInt(1000),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}

	// Create signer
	signer := types.NewEIP155Signer(chainID)

	// Create validation options
	opts := &ValidationOptions{
		Config:       &params.ChainConfig{},
		Accept:       1 << types.SetCodeTxType,
		MaxSize:      32 * 1024,
		MinTip:       big.NewInt(1),
		MinGasPrice:  big.NewInt(1000000000),
		PoolGasPrice: big.NewInt(1000000000),
	}

	err := validator.ValidateSetCodeTransaction(tx, head, signer, opts)
	if err != nil {
		t.Errorf("ValidateSetCodeTransaction() error = %v", err)
	}
}

func TestSetCodeTxPoolValidatorInsufficientBalance(t *testing.T) {
	chainID := big.NewInt(1)
	validator := NewSetCodeTxPoolValidator(chainID)

	key, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(key.PublicKey)
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	auth := &types.AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	tx := types.NewSetCodeTx(
		chainID,
		42,
		&codeAddr,
		big.NewInt(1000),
		50000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		types.AccessList{},
		types.AuthorizationList{*auth},
	)

	// Create mock state with insufficient balance
	statedb := state.NewDatabase(nil)
	state, _ := state.New(common.Hash{}, statedb, nil)
	
	// Set up account with insufficient balance
	state.SetBalance(from, big.NewInt(1000)) // Very small balance
	state.SetNonce(from, 42)

	head := &EvmHeader{
		Number:   big.NewInt(1000),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}

	signer := types.NewEIP155Signer(chainID)

	opts := &ValidationOptions{
		Config:       &params.ChainConfig{},
		Accept:       1 << types.SetCodeTxType,
		MaxSize:      32 * 1024,
		MinTip:       big.NewInt(1),
		MinGasPrice:  big.NewInt(1000000000),
		PoolGasPrice: big.NewInt(1000000000),
	}

	err := validator.ValidateSetCodeTransaction(tx, head, signer, opts)
	if err == nil {
		t.Errorf("expected validation error for insufficient balance but got none")
	}
}

func TestSetCodeTxPoolValidatorInvalidNonce(t *testing.T) {
	chainID := big.NewInt(1)
	validator := NewSetCodeTxPoolValidator(chainID)

	key, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(key.PublicKey)
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	auth := &types.AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	// Transaction with nonce 40, but account nonce is 42
	tx := types.NewSetCodeTx(
		chainID,
		40, // Lower than account nonce
		&codeAddr,
		big.NewInt(1000),
		50000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		types.AccessList{},
		types.AuthorizationList{*auth},
	)

	statedb := state.NewDatabase(nil)
	state, _ := state.New(common.Hash{}, statedb, nil)
	
	state.SetBalance(from, big.NewInt(1000000000000000000))
	state.SetNonce(from, 42) // Higher than transaction nonce

	head := &EvmHeader{
		Number:   big.NewInt(1000),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}

	signer := types.NewEIP155Signer(chainID)

	opts := &ValidationOptions{
		Config:       &params.ChainConfig{},
		Accept:       1 << types.SetCodeTxType,
		MaxSize:      32 * 1024,
		MinTip:       big.NewInt(1),
		MinGasPrice:  big.NewInt(1000000000),
		PoolGasPrice: big.NewInt(1000000000),
	}

	err := validator.ValidateSetCodeTransaction(tx, head, signer, opts)
	if err == nil {
		t.Errorf("expected validation error for invalid nonce but got none")
	}
}

func TestSetCodeTxPoolValidatorInvalidAuthorization(t *testing.T) {
	chainID := big.NewInt(1)
	validator := NewSetCodeTxPoolValidator(chainID)

	key, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(key.PublicKey)
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Create authorization with invalid signature
	auth := &types.AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   42,
		V:       big.NewInt(27),
		R:       big.NewInt(1),
		S:       big.NewInt(1),
	}

	tx := types.NewSetCodeTx(
		chainID,
		42,
		&codeAddr,
		big.NewInt(1000),
		50000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		types.AccessList{},
		types.AuthorizationList{*auth},
	)

	statedb := state.NewDatabase(nil)
	state, _ := state.New(common.Hash{}, statedb, nil)
	
	state.SetBalance(from, big.NewInt(1000000000000000000))
	state.SetNonce(from, 42)

	head := &EvmHeader{
		Number:   big.NewInt(1000),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}

	signer := types.NewEIP155Signer(chainID)

	opts := &ValidationOptions{
		Config:       &params.ChainConfig{},
		Accept:       1 << types.SetCodeTxType,
		MaxSize:      32 * 1024,
		MinTip:       big.NewInt(1),
		MinGasPrice:  big.NewInt(1000000000),
		PoolGasPrice: big.NewInt(1000000000),
	}

	err := validator.ValidateSetCodeTransaction(tx, head, signer, opts)
	if err == nil {
		t.Errorf("expected validation error for invalid authorization but got none")
	}
}

func TestSetCodeTxPoolValidatorGasLimit(t *testing.T) {
	chainID := big.NewInt(1)
	validator := NewSetCodeTxPoolValidator(chainID)

	key, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(key.PublicKey)
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	auth := &types.AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	// Transaction with gas higher than block limit
	tx := types.NewSetCodeTx(
		chainID,
		42,
		&codeAddr,
		big.NewInt(1000),
		9000000, // Higher than block gas limit
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		types.AccessList{},
		types.AuthorizationList{*auth},
	)

	statedb := state.NewDatabase(nil)
	state, _ := state.New(common.Hash{}, statedb, nil)
	
	state.SetBalance(from, big.NewInt(1000000000000000000))
	state.SetNonce(from, 42)

	head := &EvmHeader{
		Number:   big.NewInt(1000),
		GasLimit: 8000000, // Lower than transaction gas
		BaseFee:  big.NewInt(1000000000),
	}

	signer := types.NewEIP155Signer(chainID)

	opts := &ValidationOptions{
		Config:       &params.ChainConfig{},
		Accept:       1 << types.SetCodeTxType,
		MaxSize:      32 * 1024,
		MinTip:       big.NewInt(1),
		MinGasPrice:  big.NewInt(1000000000),
		PoolGasPrice: big.NewInt(1000000000),
	}

	err := validator.ValidateSetCodeTransaction(tx, head, signer, opts)
	if err == nil {
		t.Errorf("expected validation error for gas limit exceeded but got none")
	}
}

func TestSetCodeTxPoolValidatorNotAccepted(t *testing.T) {
	chainID := big.NewInt(1)
	validator := NewSetCodeTxPoolValidator(chainID)

	key, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(key.PublicKey)
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	auth := &types.AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	tx := types.NewSetCodeTx(
		chainID,
		42,
		&codeAddr,
		big.NewInt(1000),
		50000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		types.AccessList{},
		types.AuthorizationList{*auth},
	)

	statedb := state.NewDatabase(nil)
	state, _ := state.New(common.Hash{}, statedb, nil)
	
	state.SetBalance(from, big.NewInt(1000000000000000000))
	state.SetNonce(from, 42)

	head := &EvmHeader{
		Number:   big.NewInt(1000),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}

	signer := types.NewEIP155Signer(chainID)

	// Don't accept SetCode transactions
	opts := &ValidationOptions{
		Config:       &params.ChainConfig{},
		Accept:       1 << types.LegacyTxType, // Only accept legacy transactions
		MaxSize:      32 * 1024,
		MinTip:       big.NewInt(1),
		MinGasPrice:  big.NewInt(1000000000),
		PoolGasPrice: big.NewInt(1000000000),
	}

	err := validator.ValidateSetCodeTransaction(tx, head, signer, opts)
	if err == nil {
		t.Errorf("expected validation error for non-accepted transaction type but got none")
	}
}

// Benchmark tests
func BenchmarkSetCodeTxPoolValidation(b *testing.B) {
	chainID := big.NewInt(1)
	validator := NewSetCodeTxPoolValidator(chainID)

	key, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(key.PublicKey)
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	auth := &types.AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	tx := types.NewSetCodeTx(
		chainID,
		42,
		&codeAddr,
		big.NewInt(1000),
		50000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		types.AccessList{},
		types.AuthorizationList{*auth},
	)

	statedb := state.NewDatabase(nil)
	state, _ := state.New(common.Hash{}, statedb, nil)
	
	state.SetBalance(from, big.NewInt(1000000000000000000))
	state.SetNonce(from, 42)

	head := &EvmHeader{
		Number:   big.NewInt(1000),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}

	signer := types.NewEIP155Signer(chainID)

	opts := &ValidationOptions{
		Config:       &params.ChainConfig{},
		Accept:       1 << types.SetCodeTxType,
		MaxSize:      32 * 1024,
		MinTip:       big.NewInt(1),
		MinGasPrice:  big.NewInt(1000000000),
		PoolGasPrice: big.NewInt(1000000000),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.ValidateSetCodeTransaction(tx, head, signer, opts)
	}
}