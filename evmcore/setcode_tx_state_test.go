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

func TestDelegationStateDB(t *testing.T) {
	// Create mock state
	statedb := state.NewDatabase(nil)
	state, _ := state.New(common.Hash{}, statedb, nil)
	
	delegationDB := NewDelegationStateDB(state)

	authority := common.HexToAddress("0x1111111111111111111111111111111111111111")
	codeAddr := common.HexToAddress("0x2222222222222222222222222222222222222222")

	// Initially no delegation
	result := delegationDB.GetDelegation(authority)
	if result != nil {
		t.Errorf("expected no delegation initially, got %v", result)
	}

	// Set delegation
	delegationDB.SetDelegation(authority, codeAddr)

	// Get delegation
	result = delegationDB.GetDelegation(authority)
	if result == nil {
		t.Errorf("expected delegation after setting, got nil")
	}
	if *result != codeAddr {
		t.Errorf("expected delegation to %v, got %v", codeAddr, *result)
	}

	// Remove delegation
	delegationDB.RemoveDelegation(authority)

	// Verify removal
	result = delegationDB.GetDelegation(authority)
	if result != nil {
		t.Errorf("expected no delegation after removal, got %v", result)
	}
}

func TestSetCodeStateTransition(t *testing.T) {
	// Create test setup
	chainID := big.NewInt(1)
	key, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(key.PublicKey)
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Create authorization
	auth := &types.AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	// Create SetCode transaction
	tx := types.NewSetCodeTx(
		chainID,
		42,
		&codeAddr,
		big.NewInt(1000),
		100000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		types.AccessList{},
		types.AuthorizationList{*auth},
	)

	// Create mock state
	statedb := state.NewDatabase(nil)
	stateDB, _ := state.New(common.Hash{}, statedb, nil)
	
	// Set up account with sufficient balance and correct nonce
	stateDB.SetBalance(from, big.NewInt(1000000000000000000)) // 1 ETH
	stateDB.SetNonce(from, 42)

	// Create gas pool
	gp := new(GasPool).AddGas(8000000)

	// Create header
	header := &EvmHeader{
		Number:   big.NewInt(1000),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}

	// Create chain config
	chainConfig := &params.ChainConfig{}

	// Create state transition processor
	st := NewSetCodeStateTransition(gp, stateDB, header, chainConfig, chainID)

	// Create signer
	signer := types.NewEIP155Signer(chainID)

	// Apply transaction
	receipt, err := st.ApplySetCodeTransaction(tx, signer)
	if err != nil {
		t.Fatalf("ApplySetCodeTransaction() error = %v", err)
	}

	// Verify receipt
	if receipt == nil {
		t.Errorf("expected receipt, got nil")
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		t.Errorf("expected successful receipt status, got %v", receipt.Status)
	}

	// Verify delegation was set
	delegation := st.delegationDB.GetDelegation(from)
	if delegation == nil {
		t.Errorf("expected delegation to be set")
	}
	if *delegation != codeAddr {
		t.Errorf("expected delegation to %v, got %v", codeAddr, *delegation)
	}

	// Verify nonce was incremented
	newNonce := stateDB.GetNonce(from)
	if newNonce != 43 {
		t.Errorf("expected nonce to be incremented to 43, got %v", newNonce)
	}
}

func TestSetCodeStateTransitionInsufficientBalance(t *testing.T) {
	chainID := big.NewInt(1)
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
		100000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		types.AccessList{},
		types.AuthorizationList{*auth},
	)

	statedb := state.NewDatabase(nil)
	stateDB, _ := state.New(common.Hash{}, statedb, nil)
	
	// Set up account with insufficient balance
	stateDB.SetBalance(from, big.NewInt(1000)) // Very small balance
	stateDB.SetNonce(from, 42)

	gp := new(GasPool).AddGas(8000000)

	header := &EvmHeader{
		Number:   big.NewInt(1000),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}

	chainConfig := &params.ChainConfig{}

	st := NewSetCodeStateTransition(gp, stateDB, header, chainConfig, chainID)
	signer := types.NewEIP155Signer(chainID)

	_, err := st.ApplySetCodeTransaction(tx, signer)
	if err == nil {
		t.Errorf("expected error for insufficient balance but got none")
	}
}

func TestSetCodeStateTransitionInsufficientGas(t *testing.T) {
	chainID := big.NewInt(1)
	key, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(key.PublicKey)
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	auth := &types.AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	// Transaction with insufficient gas
	tx := types.NewSetCodeTx(
		chainID,
		42,
		&codeAddr,
		big.NewInt(1000),
		1000, // Very low gas
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		types.AccessList{},
		types.AuthorizationList{*auth},
	)

	statedb := state.NewDatabase(nil)
	stateDB, _ := state.New(common.Hash{}, statedb, nil)
	
	stateDB.SetBalance(from, big.NewInt(1000000000000000000))
	stateDB.SetNonce(from, 42)

	gp := new(GasPool).AddGas(8000000)

	header := &EvmHeader{
		Number:   big.NewInt(1000),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}

	chainConfig := &params.ChainConfig{}

	st := NewSetCodeStateTransition(gp, stateDB, header, chainConfig, chainID)
	signer := types.NewEIP155Signer(chainID)

	_, err := st.ApplySetCodeTransaction(tx, signer)
	if err == nil {
		t.Errorf("expected error for insufficient gas but got none")
	}
}

func TestSetCodeStateTransitionWrongTransactionType(t *testing.T) {
	chainID := big.NewInt(1)
	key, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(key.PublicKey)

	// Create legacy transaction instead of SetCode transaction
	tx := types.NewTransaction(
		42,
		common.HexToAddress("0x1234567890123456789012345678901234567890"),
		big.NewInt(1000),
		21000,
		big.NewInt(1000000000),
		[]byte("test data"),
	)

	statedb := state.NewDatabase(nil)
	stateDB, _ := state.New(common.Hash{}, statedb, nil)
	
	stateDB.SetBalance(from, big.NewInt(1000000000000000000))
	stateDB.SetNonce(from, 42)

	gp := new(GasPool).AddGas(8000000)

	header := &EvmHeader{
		Number:   big.NewInt(1000),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}

	chainConfig := &params.ChainConfig{}

	st := NewSetCodeStateTransition(gp, stateDB, header, chainConfig, chainID)
	signer := types.NewEIP155Signer(chainID)

	_, err := st.ApplySetCodeTransaction(tx, signer)
	if err == nil {
		t.Errorf("expected error for wrong transaction type but got none")
	}
}

func TestSetCodeTransitionApplier(t *testing.T) {
	chainID := big.NewInt(1)
	key, _ := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(key.PublicKey)
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	auth := &types.AuthorizationTuple{
		ChainID: chainID,
		Address: codeAddr,
		Nonce:   42,
	}
	auth.SignAuthorization(key)

	// Create SetCode transaction
	setCodeTx := types.NewSetCodeTx(
		chainID,
		42,
		&codeAddr,
		big.NewInt(1000),
		100000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		types.AccessList{},
		types.AuthorizationList{*auth},
	)

	// Create legacy transaction
	legacyTx := types.NewTransaction(
		43,
		common.HexToAddress("0x1234567890123456789012345678901234567890"),
		big.NewInt(1000),
		21000,
		big.NewInt(1000000000),
		[]byte("legacy data"),
	)

	statedb := state.NewDatabase(nil)
	stateDB, _ := state.New(common.Hash{}, statedb, nil)
	
	stateDB.SetBalance(from, big.NewInt(1000000000000000000))
	stateDB.SetNonce(from, 42)

	gp := new(GasPool).AddGas(8000000)

	header := &EvmHeader{
		Number:   big.NewInt(1000),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}

	chainConfig := &params.ChainConfig{}

	applier := NewSetCodeTransitionApplier(gp, stateDB, header, chainConfig, chainID)
	signer := types.NewEIP155Signer(chainID)

	// Mock default applier for legacy transactions
	defaultApplierCalled := false
	defaultApplier := func(tx *types.Transaction, signer types.Signer) (*types.Receipt, error) {
		defaultApplierCalled = true
		return &types.Receipt{
			Status: types.ReceiptStatusSuccessful,
		}, nil
	}

	// Test SetCode transaction
	receipt, err := applier.ApplyTransaction(setCodeTx, signer, defaultApplier)
	if err != nil {
		t.Errorf("ApplyTransaction() for SetCode tx error = %v", err)
	}
	if receipt == nil {
		t.Errorf("expected receipt for SetCode transaction")
	}
	if defaultApplierCalled {
		t.Errorf("default applier should not be called for SetCode transaction")
	}

	// Reset flag
	defaultApplierCalled = false

	// Test legacy transaction
	receipt, err = applier.ApplyTransaction(legacyTx, signer, defaultApplier)
	if err != nil {
		t.Errorf("ApplyTransaction() for legacy tx error = %v", err)
	}
	if receipt == nil {
		t.Errorf("expected receipt for legacy transaction")
	}
	if !defaultApplierCalled {
		t.Errorf("default applier should be called for legacy transaction")
	}

	// Test delegation retrieval
	delegation := applier.GetDelegation(from)
	if delegation == nil {
		t.Errorf("expected delegation to be set")
	}
	if *delegation != codeAddr {
		t.Errorf("expected delegation to %v, got %v", codeAddr, *delegation)
	}
}

// Benchmark tests
func BenchmarkSetCodeStateTransition(b *testing.B) {
	chainID := big.NewInt(1)
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
		100000,
		big.NewInt(1000000000),
		big.NewInt(2000000000),
		[]byte("test data"),
		types.AccessList{},
		types.AuthorizationList{*auth},
	)

	signer := types.NewEIP155Signer(chainID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		
		// Reset state for each iteration
		statedb := state.NewDatabase(nil)
		stateDB, _ := state.New(common.Hash{}, statedb, nil)
		
		stateDB.SetBalance(from, big.NewInt(1000000000000000000))
		stateDB.SetNonce(from, 42)

		gp := new(GasPool).AddGas(8000000)

		header := &EvmHeader{
			Number:   big.NewInt(1000),
			GasLimit: 8000000,
			BaseFee:  big.NewInt(1000000000),
		}

		chainConfig := &params.ChainConfig{}

		st := NewSetCodeStateTransition(gp, stateDB, header, chainConfig, chainID)

		b.StartTimer()
		st.ApplySetCodeTransaction(tx, signer)
	}
}

func BenchmarkDelegationStateDBOperations(b *testing.B) {
	statedb := state.NewDatabase(nil)
	state, _ := state.New(common.Hash{}, statedb, nil)
	
	delegationDB := NewDelegationStateDB(state)

	authority := common.HexToAddress("0x1111111111111111111111111111111111111111")
	codeAddr := common.HexToAddress("0x2222222222222222222222222222222222222222")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		delegationDB.SetDelegation(authority, codeAddr)
		delegationDB.GetDelegation(authority)
		delegationDB.RemoveDelegation(authority)
	}
}