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

package vm

import (
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
	"github.com/unicornultrafoundation/go-u2u/core/state"
	"github.com/unicornultrafoundation/go-u2u/core/types"
	"github.com/unicornultrafoundation/go-u2u/params"
)

// Mock StateDB for testing
type mockStateDB struct {
	state    map[common.Address]map[common.Hash]common.Hash
	code     map[common.Address][]byte
	balances map[common.Address]*big.Int
	nonces   map[common.Address]uint64
}

func newMockStateDB() *mockStateDB {
	return &mockStateDB{
		state:    make(map[common.Address]map[common.Hash]common.Hash),
		code:     make(map[common.Address][]byte),
		balances: make(map[common.Address]*big.Int),
		nonces:   make(map[common.Address]uint64),
	}
}

func (m *mockStateDB) GetState(addr common.Address, key common.Hash) common.Hash {
	if m.state[addr] == nil {
		return common.Hash{}
	}
	return m.state[addr][key]
}

func (m *mockStateDB) SetState(addr common.Address, key common.Hash, value common.Hash) {
	if m.state[addr] == nil {
		m.state[addr] = make(map[common.Hash]common.Hash)
	}
	m.state[addr][key] = value
}

func (m *mockStateDB) GetCode(addr common.Address) []byte {
	return m.code[addr]
}

func (m *mockStateDB) SetCode(addr common.Address, code []byte) {
	m.code[addr] = code
}

func (m *mockStateDB) GetBalance(addr common.Address) *big.Int {
	if balance := m.balances[addr]; balance != nil {
		return balance
	}
	return big.NewInt(0)
}

func (m *mockStateDB) SetBalance(addr common.Address, balance *big.Int) {
	m.balances[addr] = balance
}

func (m *mockStateDB) GetNonce(addr common.Address) uint64 {
	return m.nonces[addr]
}

func (m *mockStateDB) SetNonce(addr common.Address, nonce uint64) {
	m.nonces[addr] = nonce
}

// Implement other StateDB methods as no-ops for testing
func (m *mockStateDB) CreateAccount(common.Address)                                        {}
func (m *mockStateDB) SubBalance(common.Address, *big.Int)                                 {}
func (m *mockStateDB) AddBalance(common.Address, *big.Int)                                 {}
func (m *mockStateDB) GetCodeHash(common.Address) common.Hash                              { return common.Hash{} }
func (m *mockStateDB) GetCodeSize(common.Address) int                                      { return 0 }
func (m *mockStateDB) GetRefund() uint64                                                   { return 0 }
func (m *mockStateDB) GetCommittedState(common.Address, common.Hash) common.Hash          { return common.Hash{} }
func (m *mockStateDB) GetTransientState(common.Address, common.Hash) common.Hash          { return common.Hash{} }
func (m *mockStateDB) SetTransientState(common.Address, common.Hash, common.Hash)         {}
func (m *mockStateDB) Suicide(common.Address) bool                                        { return false }
func (m *mockStateDB) HasSuicided(common.Address) bool                                    { return false }
func (m *mockStateDB) Exist(common.Address) bool                                          { return true }
func (m *mockStateDB) Empty(common.Address) bool                                          { return false }
func (m *mockStateDB) AddRefund(uint64)                                                   {}
func (m *mockStateDB) SubRefund(uint64)                                                   {}
func (m *mockStateDB) AddAddressToAccessList(addr common.Address)                         {}
func (m *mockStateDB) AddSlotToAccessList(addr common.Address, slot common.Hash)          {}
func (m *mockStateDB) AddressInAccessList(addr common.Address) bool                       { return false }
func (m *mockStateDB) SlotInAccessList(addr common.Address, slot common.Hash) (bool, bool) { return false, false }
func (m *mockStateDB) Snapshot() int                                                      { return 0 }
func (m *mockStateDB) RevertToSnapshot(int)                                               {}
func (m *mockStateDB) AddLog(*types.Log)                                                  {}
func (m *mockStateDB) GetLogs(common.Hash, uint64, common.Hash) []*types.Log             { return nil }

func TestDelegationResolver(t *testing.T) {
	stateDB := newMockStateDB()
	resolver := NewDelegationResolver(stateDB)

	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Set some code at codeAddr
	testCode := []byte{0x60, 0x00, 0x60, 0x00, 0xf3} // Simple return bytecode
	stateDB.SetCode(codeAddr, testCode)

	// Test no delegation
	finalAddr, code, err := resolver.ResolveDelegatedCode(addr1)
	if err != nil {
		t.Errorf("ResolveDelegatedCode() error = %v", err)
	}
	if finalAddr != addr1 {
		t.Errorf("expected final address %v, got %v", addr1, finalAddr)
	}
	if len(code) != 0 {
		t.Errorf("expected no code, got %v", code)
	}

	// Set delegation from addr1 to codeAddr
	delegationKey := common.BytesToHash(append([]byte("EIP7702_DELEGATION_"), addr1.Bytes()...))
	stateDB.SetState(common.HexToAddress("0x7702"), delegationKey, codeAddr.Hash())

	// Test delegation resolution
	finalAddr, code, err = resolver.ResolveDelegatedCode(addr1)
	if err != nil {
		t.Errorf("ResolveDelegatedCode() error = %v", err)
	}
	if finalAddr != codeAddr {
		t.Errorf("expected final address %v, got %v", codeAddr, finalAddr)
	}
	if len(code) != len(testCode) {
		t.Errorf("expected code length %d, got %d", len(testCode), len(code))
	}

	// Test delegation check
	if !resolver.CheckDelegation(addr1) {
		t.Errorf("expected delegation to be found for addr1")
	}
	if resolver.CheckDelegation(addr2) {
		t.Errorf("expected no delegation for addr2")
	}

	// Test direct delegation retrieval
	directDelegation := resolver.GetDirectDelegation(addr1)
	if directDelegation == nil {
		t.Errorf("expected direct delegation for addr1")
	}
	if *directDelegation != codeAddr {
		t.Errorf("expected direct delegation to %v, got %v", codeAddr, *directDelegation)
	}

	directDelegation = resolver.GetDirectDelegation(addr2)
	if directDelegation != nil {
		t.Errorf("expected no direct delegation for addr2, got %v", *directDelegation)
	}
}

func TestDelegationResolverChain(t *testing.T) {
	stateDB := newMockStateDB()
	resolver := NewDelegationResolver(stateDB)

	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Set some code at final address
	testCode := []byte{0x60, 0x00, 0x60, 0x00, 0xf3}
	stateDB.SetCode(codeAddr, testCode)

	// Set delegation chain: addr1 -> addr2 -> codeAddr
	delegationKey1 := common.BytesToHash(append([]byte("EIP7702_DELEGATION_"), addr1.Bytes()...))
	stateDB.SetState(common.HexToAddress("0x7702"), delegationKey1, addr2.Hash())

	delegationKey2 := common.BytesToHash(append([]byte("EIP7702_DELEGATION_"), addr2.Bytes()...))
	stateDB.SetState(common.HexToAddress("0x7702"), delegationKey2, codeAddr.Hash())

	// Test chain resolution
	finalAddr, code, err := resolver.ResolveDelegatedCode(addr1)
	if err != nil {
		t.Errorf("ResolveDelegatedCode() error = %v", err)
	}
	if finalAddr != codeAddr {
		t.Errorf("expected final address %v, got %v", codeAddr, finalAddr)
	}
	if len(code) != len(testCode) {
		t.Errorf("expected code length %d, got %d", len(testCode), len(code))
	}
}

func TestEnhancedEVM(t *testing.T) {
	stateDB := newMockStateDB()
	chainConfig := params.AllEthashProtocolChanges
	
	// Create EVM context
	blockCtx := BlockContext{
		CanTransfer: func(StateDB, common.Address, *big.Int) bool { return true },
		Transfer:    func(StateDB, common.Address, common.Address, *big.Int) {},
		GetHash:     func(uint64) common.Hash { return common.Hash{} },
		Coinbase:    common.Address{},
		GasLimit:    8000000,
		BlockNumber: big.NewInt(1),
		Time:        big.NewInt(1),
		Difficulty:  big.NewInt(1),
		BaseFee:     big.NewInt(1000000000),
	}

	txCtx := TxContext{
		Origin:   common.Address{},
		GasPrice: big.NewInt(1000000000),
	}

	vmConfig := Config{}

	// Create enhanced EVM
	evm := NewEnhancedEVM(blockCtx, txCtx, stateDB, chainConfig, vmConfig)

	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Set some code at codeAddr
	testCode := []byte{0x60, 0x00, 0x60, 0x00, 0xf3}
	stateDB.SetCode(codeAddr, testCode)

	// Test getting code without delegation
	code := evm.GetCodeWithDelegation(addr1)
	if len(code) != 0 {
		t.Errorf("expected no code without delegation, got %v", code)
	}

	// Set delegation
	delegationKey := common.BytesToHash(append([]byte("EIP7702_DELEGATION_"), addr1.Bytes()...))
	stateDB.SetState(common.HexToAddress("0x7702"), delegationKey, codeAddr.Hash())

	// Test getting code with delegation
	code = evm.GetCodeWithDelegation(addr1)
	if len(code) != len(testCode) {
		t.Errorf("expected code length %d with delegation, got %d", len(testCode), len(code))
	}

	// Test delegation resolver access
	resolver := evm.GetDelegationResolver()
	if resolver == nil {
		t.Errorf("expected delegation resolver, got nil")
	}
}

func TestSetCodeEVMContext(t *testing.T) {
	chainID := big.NewInt(1)
	ctx := NewSetCodeEVMContext(chainID)

	// Test initial state
	if !ctx.IsDelegationEnabled() {
		t.Errorf("expected delegation to be enabled by default")
	}

	if ctx.GetChainID().Cmp(chainID) != 0 {
		t.Errorf("expected chain ID %v, got %v", chainID, ctx.GetChainID())
	}

	// Test disabling delegation
	ctx.SetDelegationEnabled(false)
	if ctx.IsDelegationEnabled() {
		t.Errorf("expected delegation to be disabled")
	}

	// Test re-enabling delegation
	ctx.SetDelegationEnabled(true)
	if !ctx.IsDelegationEnabled() {
		t.Errorf("expected delegation to be enabled again")
	}
}

func TestEnhancedEVMCallMethods(t *testing.T) {
	stateDB := newMockStateDB()
	chainConfig := params.AllEthashProtocolChanges
	
	caller := common.HexToAddress("0x1111111111111111111111111111111111111111")
	target := common.HexToAddress("0x2222222222222222222222222222222222222222")
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Set up balances
	stateDB.SetBalance(caller, big.NewInt(1000000000000000000))
	stateDB.SetBalance(target, big.NewInt(1000000000000000000))

	blockCtx := BlockContext{
		CanTransfer: func(db StateDB, addr common.Address, amount *big.Int) bool {
			return db.GetBalance(addr).Cmp(amount) >= 0
		},
		Transfer: func(db StateDB, from, to common.Address, amount *big.Int) {
			db.SubBalance(from, amount)
			db.AddBalance(to, amount)
		},
		GetHash:     func(uint64) common.Hash { return common.Hash{} },
		Coinbase:    common.Address{},
		GasLimit:    8000000,
		BlockNumber: big.NewInt(1),
		Time:        big.NewInt(1),
		Difficulty:  big.NewInt(1),
		BaseFee:     big.NewInt(1000000000),
	}

	txCtx := TxContext{
		Origin:   caller,
		GasPrice: big.NewInt(1000000000),
	}

	vmConfig := Config{}

	evm := NewEnhancedEVM(blockCtx, txCtx, stateDB, chainConfig, vmConfig)

	// Set simple return bytecode at codeAddr
	testCode := []byte{0x60, 0x00, 0x60, 0x00, 0xf3} // PUSH1 0x00 PUSH1 0x00 RETURN
	stateDB.SetCode(codeAddr, testCode)

	// Test call without delegation
	callerRef := AccountRef(caller)
	ret, leftOverGas, err := evm.CallWithDelegation(callerRef, target, []byte{}, 21000, big.NewInt(0))
	if err != nil {
		t.Errorf("CallWithDelegation() without delegation error = %v", err)
	}
	_ = ret
	_ = leftOverGas

	// Set delegation from target to codeAddr
	delegationKey := common.BytesToHash(append([]byte("EIP7702_DELEGATION_"), target.Bytes()...))
	stateDB.SetState(common.HexToAddress("0x7702"), delegationKey, codeAddr.Hash())

	// Test call with delegation
	ret, leftOverGas, err = evm.CallWithDelegation(callerRef, target, []byte{}, 21000, big.NewInt(0))
	if err != nil {
		t.Errorf("CallWithDelegation() with delegation error = %v", err)
	}

	// Test DELEGATECALL with delegation
	ret, leftOverGas, err = evm.DelegateCallWithDelegation(callerRef, target, []byte{}, 21000)
	if err != nil {
		t.Errorf("DelegateCallWithDelegation() with delegation error = %v", err)
	}

	// Test STATICCALL with delegation
	ret, leftOverGas, err = evm.StaticCallWithDelegation(callerRef, target, []byte{}, 21000)
	if err != nil {
		t.Errorf("StaticCallWithDelegation() with delegation error = %v", err)
	}
}

// Benchmark tests
func BenchmarkDelegationResolution(b *testing.B) {
	stateDB := newMockStateDB()
	resolver := NewDelegationResolver(stateDB)

	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Set delegation
	delegationKey := common.BytesToHash(append([]byte("EIP7702_DELEGATION_"), addr1.Bytes()...))
	stateDB.SetState(common.HexToAddress("0x7702"), delegationKey, codeAddr.Hash())

	// Set some code
	testCode := []byte{0x60, 0x00, 0x60, 0x00, 0xf3}
	stateDB.SetCode(codeAddr, testCode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.ResolveDelegatedCode(addr1)
	}
}

func BenchmarkEnhancedEVMCodeRetrieval(b *testing.B) {
	stateDB := newMockStateDB()
	chainConfig := params.AllEthashProtocolChanges
	
	blockCtx := BlockContext{
		CanTransfer: func(StateDB, common.Address, *big.Int) bool { return true },
		Transfer:    func(StateDB, common.Address, common.Address, *big.Int) {},
		GetHash:     func(uint64) common.Hash { return common.Hash{} },
		Coinbase:    common.Address{},
		GasLimit:    8000000,
		BlockNumber: big.NewInt(1),
		Time:        big.NewInt(1),
		Difficulty:  big.NewInt(1),
		BaseFee:     big.NewInt(1000000000),
	}

	txCtx := TxContext{
		Origin:   common.Address{},
		GasPrice: big.NewInt(1000000000),
	}

	vmConfig := Config{}

	evm := NewEnhancedEVM(blockCtx, txCtx, stateDB, chainConfig, vmConfig)

	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Set delegation and code
	delegationKey := common.BytesToHash(append([]byte("EIP7702_DELEGATION_"), addr1.Bytes()...))
	stateDB.SetState(common.HexToAddress("0x7702"), delegationKey, codeAddr.Hash())

	testCode := []byte{0x60, 0x00, 0x60, 0x00, 0xf3}
	stateDB.SetCode(codeAddr, testCode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evm.GetCodeWithDelegation(addr1)
	}
}