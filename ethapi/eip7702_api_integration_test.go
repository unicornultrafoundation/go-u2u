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

package ethapi

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/unicornultrafoundation/go-helios/native/idx"
	"github.com/unicornultrafoundation/go-u2u/accounts"
	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/common/hexutil"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/core/vm"
	"github.com/unicornultrafoundation/go-u2u/ethdb"
	"github.com/unicornultrafoundation/go-u2u/evmcore/txtracer"
	notify "github.com/unicornultrafoundation/go-u2u/event"
	"github.com/unicornultrafoundation/go-u2u/evmcore"
	"github.com/unicornultrafoundation/go-u2u/params"
	"github.com/unicornultrafoundation/go-u2u/rpc"
)

// TestTransactionArgs_AuthorizationList tests TransactionArgs enhancement for EIP-7702
func TestTransactionArgs_AuthorizationList(t *testing.T) {
	tests := []struct {
		name        string
		args        TransactionArgs
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid authorization list",
			args: TransactionArgs{
				To:    &common.Address{1},
				Value: (*hexutil.Big)(big.NewInt(100)),
				AuthorizationList: &types.AuthorizationList{
					{
						ChainID: big.NewInt(1),
						Address: common.Address{2},
						Nonce:   1,
						V:       big.NewInt(27),
						R:       big.NewInt(1),
						S:       big.NewInt(1),
					},
				},
				MaxFeePerGas:         (*hexutil.Big)(big.NewInt(1000000000)),
				MaxPriorityFeePerGas: (*hexutil.Big)(big.NewInt(1000000000)),
			},
			expectError: false,
		},
		{
			name: "Authorization list without recipient",
			args: TransactionArgs{
				Value: (*hexutil.Big)(big.NewInt(100)),
				AuthorizationList: &types.AuthorizationList{
					{
						ChainID: big.NewInt(1),
						Address: common.Address{2},
						Nonce:   1,
						V:       big.NewInt(27),
						R:       big.NewInt(1),
						S:       big.NewInt(1),
					},
				},
			},
			expectError: true,
			errorMsg:    "EIP-7702 transactions cannot create contracts",
		},
		{
			name: "Empty authorization list",
			args: TransactionArgs{
				To:                &common.Address{1},
				Value:             (*hexutil.Big)(big.NewInt(100)),
				AuthorizationList: &types.AuthorizationList{},
			},
			expectError: true,
			errorMsg:    "EIP-7702 transactions cannot have empty authorization list",
		},
		{
			name: "Invalid signature values in authorization",
			args: TransactionArgs{
				To:    &common.Address{1},
				Value: (*hexutil.Big)(big.NewInt(100)),
				AuthorizationList: &types.AuthorizationList{
					{
						ChainID: big.NewInt(1),
						Address: common.Address{2},
						Nonce:   1,
						V:       big.NewInt(300), // Invalid V value
						R:       big.NewInt(1),
						S:       big.NewInt(1),
					},
				},
				MaxFeePerGas:         (*hexutil.Big)(big.NewInt(1000000000)),
				MaxPriorityFeePerGas: (*hexutil.Big)(big.NewInt(1000000000)),
			},
			expectError: true,
			errorMsg:    "invalid signature values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock backend for testing
			backend := &MockBackend{
				chainConfig: &params.ChainConfig{
					ChainID:       big.NewInt(1),
					LondonBlock:   big.NewInt(0),
					PhaethonBlock: big.NewInt(0),
				},
			}

			err := tt.args.setDefaults(context.Background(), backend)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestTransactionArgs_ToTransaction tests transaction creation with authorization list
func TestTransactionArgs_ToTransaction(t *testing.T) {
	authList := &types.AuthorizationList{
		{
			ChainID: big.NewInt(1),
			Address: common.Address{2},
			Nonce:   1,
			V:       big.NewInt(27),
			R:       big.NewInt(1),
			S:       big.NewInt(1),
		},
	}

	args := TransactionArgs{
		To:                   &common.Address{1},
		Value:                (*hexutil.Big)(big.NewInt(100)),
		Gas:                  func() *hexutil.Uint64 { v := hexutil.Uint64(21000); return &v }(),
		MaxFeePerGas:         (*hexutil.Big)(big.NewInt(1000000000)),
		MaxPriorityFeePerGas: (*hexutil.Big)(big.NewInt(1000000000)),
		Nonce:                func() *hexutil.Uint64 { v := hexutil.Uint64(1); return &v }(),
		ChainID:              (*hexutil.Big)(big.NewInt(1)),
		AuthorizationList:    authList,
	}

	tx := args.toTransaction()

	// Verify transaction type is SetCodeTx
	if tx.Type() != types.SetCodeTxType {
		t.Errorf("Expected transaction type %d, got %d", types.SetCodeTxType, tx.Type())
	}

	// Verify authorization list is preserved
	innerTx, ok := tx.Inner().(*types.SetCodeTx)
	if !ok {
		t.Fatalf("Failed to cast to SetCodeTx")
	}

	if len(innerTx.AuthorizationList) != len(*authList) {
		t.Errorf("Expected %d authorizations, got %d", len(*authList), len(innerTx.AuthorizationList))
	}

	// Verify authorization details
	if innerTx.AuthorizationList[0].Address != (*authList)[0].Address {
		t.Errorf("Authorization address mismatch")
	}
}

// TestValidateAuthorization tests the authorization validation function
func TestValidateAuthorization(t *testing.T) {
	chainID := big.NewInt(1)

	tests := []struct {
		name        string
		auth        types.AuthorizationTuple
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid authorization",
			auth: types.AuthorizationTuple{
				ChainID: big.NewInt(1),
				Address: common.Address{2},
				Nonce:   1,
				V:       big.NewInt(27),
				R:       big.NewInt(1),
				S:       big.NewInt(1),
			},
			expectError: false,
		},
		{
			name: "Chain ID mismatch",
			auth: types.AuthorizationTuple{
				ChainID: big.NewInt(2), // Different chain ID
				Address: common.Address{2},
				Nonce:   1,
				V:       big.NewInt(27),
				R:       big.NewInt(1),
				S:       big.NewInt(1),
			},
			expectError: true,
			errorMsg:    "authorization chain ID 2 does not match expected chain ID 1",
		},
		{
			name: "Invalid signature values",
			auth: types.AuthorizationTuple{
				ChainID: big.NewInt(1),
				Address: common.Address{2},
				Nonce:   1,
				V:       big.NewInt(300), // Invalid V
				R:       big.NewInt(1),
				S:       big.NewInt(1),
			},
			expectError: true,
			errorMsg:    "invalid signature values in authorization",
		},
		{
			name: "Nonce overflow",
			auth: types.AuthorizationTuple{
				ChainID: big.NewInt(1),
				Address: common.Address{2},
				Nonce:   ^uint64(0), // Maximum uint64
				V:       big.NewInt(27),
				R:       big.NewInt(1),
				S:       big.NewInt(1),
			},
			expectError: true,
			errorMsg:    "authorization nonce overflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAuthorization(&tt.auth, chainID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// MockBackend implements a minimal Backend interface for testing
type MockBackend struct {
	chainConfig *params.ChainConfig
	gasPrice    *big.Int
	nonce       uint64
}

func (m *MockBackend) ChainConfig() *params.ChainConfig {
	return m.chainConfig
}

func (m *MockBackend) SuggestGasTipCap(ctx context.Context, certainty uint64) *big.Int {
	if m.gasPrice != nil {
		return m.gasPrice
	}
	return big.NewInt(1000000000) // 1 gwei
}

func (m *MockBackend) MinGasPrice() *big.Int {
	return big.NewInt(0)
}

func (m *MockBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return m.nonce, nil
}

func (m *MockBackend) CurrentBlock() *evmcore.EvmBlock {
	header := &evmcore.EvmHeader{
		Number:   big.NewInt(100),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}
	return evmcore.NewEvmBlock(header, nil)
}

func (m *MockBackend) RPCGasCap() uint64 {
	return 50000000
}

func (m *MockBackend) AccountManager() *accounts.Manager {
	// Return a mock account manager for testing
	return accounts.NewManager(&accounts.Config{})
}

func (m *MockBackend) StateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *evmcore.EvmHeader, error) {
	// Return mock state and header for testing  
	header := &evmcore.EvmHeader{
		Number:   big.NewInt(100),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}
	return nil, header, nil
}

func (m *MockBackend) SfcStateAndHeaderByNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (*state.StateDB, *evmcore.EvmHeader, error) {
	// Return mock SFC state and header for testing  
	header := &evmcore.EvmHeader{
		Number:   big.NewInt(100),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}
	return nil, header, nil
}

func (m *MockBackend) BlockByHash(ctx context.Context, hash common.Hash) (*evmcore.EvmBlock, error) {
	// Return mock block for testing
	header := &evmcore.EvmHeader{
		Number:   big.NewInt(100),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}
	return evmcore.NewEvmBlock(header, nil), nil
}

func (m *MockBackend) BlockByNumber(ctx context.Context, number rpc.BlockNumber) (*evmcore.EvmBlock, error) {
	// Return mock block for testing
	header := &evmcore.EvmHeader{
		Number:   big.NewInt(100),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}
	return evmcore.NewEvmBlock(header, nil), nil
}

// Add all missing methods required by Backend interface
func (m *MockBackend) Progress() PeerProgress {
	return PeerProgress{}
}

func (m *MockBackend) EffectiveMinGasPrice(ctx context.Context) *big.Int {
	return big.NewInt(0)
}

func (m *MockBackend) ChainDb() ethdb.Database {
	return nil
}

func (m *MockBackend) ExtRPCEnabled() bool {
	return false
}

func (m *MockBackend) RPCTxFeeCap() float64 {
	return 1
}

func (m *MockBackend) RPCTimeout() time.Duration {
	return time.Second * 5
}

func (m *MockBackend) UnprotectedAllowed() bool {
	return false
}

func (m *MockBackend) CalcBlockExtApi() bool {
	return false
}

func (m *MockBackend) StateAtBlock(ctx context.Context, block *evmcore.EvmBlock, reexec uint64, base *state.StateDB, checkLive bool) (*state.StateDB, *state.StateDB, error) {
	return nil, nil, nil
}

func (m *MockBackend) StateAtTransaction(ctx context.Context, block *evmcore.EvmBlock, txIndex int, reexec uint64) (evmcore.Message, vm.BlockContext, *state.StateDB, error) {
	return nil, vm.BlockContext{}, nil, nil
}

func (m *MockBackend) HeaderByNumber(ctx context.Context, number rpc.BlockNumber) (*evmcore.EvmHeader, error) {
	header := &evmcore.EvmHeader{
		Number:   big.NewInt(100),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}
	return header, nil
}

func (m *MockBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*evmcore.EvmHeader, error) {
	header := &evmcore.EvmHeader{
		Number:   big.NewInt(100),
		GasLimit: 8000000,
		BaseFee:  big.NewInt(1000000000),
	}
	return header, nil
}

func (m *MockBackend) ResolveRpcBlockNumberOrHash(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (idx.Block, error) {
	return idx.Block(0), nil
}

func (m *MockBackend) GetReceiptsByNumber(ctx context.Context, number rpc.BlockNumber) (types.Receipts, error) {
	return nil, nil
}

func (m *MockBackend) GetTd(hash common.Hash) *big.Int {
	return big.NewInt(0)
}

func (m *MockBackend) GetEVM(ctx context.Context, msg evmcore.Message, state *state.StateDB, sfcState *state.StateDB, header *evmcore.EvmHeader, vmConfig *vm.Config) (*vm.EVM, func() error, error) {
	return nil, nil, nil
}

func (m *MockBackend) GetBlockContext(header *evmcore.EvmHeader) vm.BlockContext {
	return vm.BlockContext{}
}

func (m *MockBackend) MaxGasLimit() uint64 {
	return 8000000
}

func (m *MockBackend) TxTraceByHash(ctx context.Context, h common.Hash) (*[]txtracer.ActionTrace, error) {
	return nil, nil
}

func (m *MockBackend) TxTraceSave(ctx context.Context, h common.Hash, traces []byte) error {
	return nil
}

func (m *MockBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return nil
}

func (m *MockBackend) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, uint64, uint64, error) {
	return nil, 0, 0, nil
}

func (m *MockBackend) GetPoolTransactions() (types.Transactions, error) {
	return nil, nil
}

func (m *MockBackend) GetPoolTransaction(txHash common.Hash) *types.Transaction {
	return nil
}

func (m *MockBackend) Stats() (pending int, queued int) {
	return 0, 0
}

func (m *MockBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return nil, nil
}

func (m *MockBackend) TxPoolContentFrom(addr common.Address) (types.Transactions, types.Transactions) {
	return nil, nil
}

func (m *MockBackend) SubscribeNewTxsNotify(chan<- evmcore.NewTxsNotify) notify.Subscription {
	return nil
}

func (m *MockBackend) CurrentEpoch(ctx context.Context) idx.Epoch {
	return idx.Epoch(0)
}

func (m *MockBackend) GetDowntime(ctx context.Context, validatorID idx.ValidatorID) (*big.Int, error) {
	return big.NewInt(0), nil
}