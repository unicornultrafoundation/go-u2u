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
	"fmt"
	"math/big"
	"testing"

	"github.com/unicornultrafoundation/go-u2u/common"
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
func (m *mockStateDB) AddPreimage(common.Hash, []byte)                                   {}
func (m *mockStateDB) ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) error { return nil }
func (m *mockStateDB) GetStorageRoot(common.Address) common.Hash                         { return common.Hash{} }
func (m *mockStateDB) PrepareAccessList(sender common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {}

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
	finalAddr, code, gasUsed, err := resolver.ResolveDelegatedCode(addr1)
	if err != nil {
		t.Errorf("ResolveDelegatedCode() error = %v", err)
	}
	if finalAddr != addr1 {
		t.Errorf("expected final address %v, got %v", addr1, finalAddr)
	}
	if len(code) != 0 {
		t.Errorf("expected no code, got %v", code)
	}
	if gasUsed != 0 {
		t.Errorf("expected no gas used for non-delegation, got %v", gasUsed)
	}

	// Set delegation from addr1 to codeAddr using EIP-7702 delegation prefix
	delegationCode := types.AddressToDelegation(codeAddr)
	stateDB.SetCode(addr1, delegationCode)

	// Test delegation resolution
	finalAddr, code, gasUsed, err = resolver.ResolveDelegatedCode(addr1)
	if err != nil {
		t.Errorf("ResolveDelegatedCode() error = %v", err)
	}
	if finalAddr != codeAddr {
		t.Errorf("expected final address %v, got %v", codeAddr, finalAddr)
	}
	if len(code) != len(testCode) {
		t.Errorf("expected code length %d, got %d", len(testCode), len(code))
	}
	if gasUsed == 0 {
		t.Errorf("expected gas used for delegation resolution, got 0")
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

	// Set delegation chain: addr1 -> addr2 -> codeAddr using EIP-7702 delegation prefixes
	delegationCode1 := types.AddressToDelegation(addr2)
	stateDB.SetCode(addr1, delegationCode1)

	delegationCode2 := types.AddressToDelegation(codeAddr)
	stateDB.SetCode(addr2, delegationCode2)

	// Test chain resolution
	finalAddr, code, gasUsed, err := resolver.ResolveDelegatedCode(addr1)
	if err != nil {
		t.Errorf("ResolveDelegatedCode() error = %v", err)
	}
	if finalAddr != codeAddr {
		t.Errorf("expected final address %v, got %v", codeAddr, finalAddr)
	}
	if len(code) != len(testCode) {
		t.Errorf("expected code length %d, got %d", len(testCode), len(code))
	}
	// Should have used gas for resolving 2 delegations
	expectedGas := uint64(200) // 2 * 100 gas per delegation
	if gasUsed != expectedGas {
		t.Errorf("expected gas used %d, got %d", expectedGas, gasUsed)
	}
}

func TestEVMWithDelegation(t *testing.T) {
	stateDB := newMockStateDB()
	chainConfig := params.TestChainConfig
	
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

	// Create standard EVM with delegation support
	evm := NewEVM(blockCtx, txCtx, stateDB, nil, chainConfig, vmConfig)

	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Set some code at codeAddr
	testCode := []byte{0x60, 0x00, 0x60, 0x00, 0xf3}
	stateDB.SetCode(codeAddr, testCode)

	// Test getting code without delegation
	code := evm.StateDB.GetCode(addr1)
	if len(code) != 0 {
		t.Errorf("expected no code without delegation, got %v", code)
	}

	// Set delegation using EIP-7702 delegation prefix
	delegationCode := types.AddressToDelegation(codeAddr)
	stateDB.SetCode(addr1, delegationCode)

	// Test delegation resolution
	finalAddr, resolvedCode, gasUsed, err := evm.resolveDelegation(addr1)
	if err != nil {
		t.Errorf("resolveDelegation() error = %v", err)
	}
	if finalAddr != codeAddr {
		t.Errorf("expected final address %v, got %v", codeAddr, finalAddr)
	}
	if len(resolvedCode) != len(testCode) {
		t.Errorf("expected code length %d with delegation, got %d", len(testCode), len(resolvedCode))
	}
	if gasUsed == 0 {
		t.Errorf("expected gas used for delegation resolution, got 0")
	}

	// Test delegation caching - second call should use cache
	finalAddr2, resolvedCode2, gasUsed2, err2 := evm.resolveDelegation(addr1)
	if err2 != nil {
		t.Errorf("cached resolveDelegation() error = %v", err2)
	}
	if finalAddr2 != codeAddr {
		t.Errorf("cached: expected final address %v, got %v", codeAddr, finalAddr2)
	}
	if len(resolvedCode2) != len(testCode) {
		t.Errorf("cached: expected code length %d with delegation, got %d", len(testCode), len(resolvedCode2))
	}
	if gasUsed2 != 0 {
		t.Errorf("expected no gas used for cached resolution, got %d", gasUsed2)
	}
}

func TestCircularDelegationDetection(t *testing.T) {
	stateDB := newMockStateDB()
	resolver := NewDelegationResolver(stateDB)

	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	// Create circular delegation: addr1 -> addr2 -> addr1
	delegationCode1 := types.AddressToDelegation(addr2)
	stateDB.SetCode(addr1, delegationCode1)

	delegationCode2 := types.AddressToDelegation(addr1)
	stateDB.SetCode(addr2, delegationCode2)

	// Test circular delegation detection
	finalAddr, code, gasUsed, err := resolver.ResolveDelegatedCode(addr1)
	if err != nil {
		t.Errorf("ResolveDelegatedCode() error = %v", err)
	}
	// Should break the loop and return one of the addresses
	if finalAddr != addr1 && finalAddr != addr2 {
		t.Errorf("expected final address to be one of the circular addresses, got %v", finalAddr)
	}
	// Code should be delegation code since we stopped at a delegating address
	if len(code) == 0 {
		t.Errorf("expected delegation code for circular delegation, got empty code")
	}
	// Verify it's actually delegation code
	if _, isDelegation := types.ParseDelegation(code); !isDelegation {
		t.Errorf("expected delegation code for circular delegation resolution")
	}
	// Should have used some gas for delegation attempts
	if gasUsed == 0 {
		t.Errorf("expected gas used for circular delegation attempts, got 0")
	}
}

func TestMaxDelegationDepth(t *testing.T) {
	stateDB := newMockStateDB()
	resolver := NewDelegationResolver(stateDB)

	// Create a chain longer than max depth
	addresses := make([]common.Address, 15) // Longer than maxDelegationDepth (10)
	for i := 0; i < 15; i++ {
		addresses[i] = common.HexToAddress(fmt.Sprintf("0x%040d", i+1))
	}

	// Set up delegation chain
	for i := 0; i < 14; i++ {
		delegationCode := types.AddressToDelegation(addresses[i+1])
		stateDB.SetCode(addresses[i], delegationCode)
	}

	// Set actual code at the end
	testCode := []byte{0x60, 0x00, 0x60, 0x00, 0xf3}
	stateDB.SetCode(addresses[14], testCode)

	// Test delegation resolution stops at max depth
	finalAddr, code, gasUsed, err := resolver.ResolveDelegatedCode(addresses[0])
	if err != nil {
		t.Errorf("ResolveDelegatedCode() error = %v", err)
	}
	
	// Should have stopped at max depth, not reached the final address
	if finalAddr == addresses[14] {
		t.Errorf("delegation chain should have been limited by max depth")
	}
	// Code should be delegation code since we stopped at max depth
	if len(code) == 0 {
		t.Errorf("expected delegation code at max depth, got empty code")
	}
	
	// Should have used gas for up to max depth delegations
	maxExpectedGas := uint64(10 * 100) // maxDelegationDepth * delegationGas
	if gasUsed > maxExpectedGas {
		t.Errorf("gas used %d exceeds maximum expected %d", gasUsed, maxExpectedGas)
	}
}

// Benchmark tests
func BenchmarkDelegationResolution(b *testing.B) {
	stateDB := newMockStateDB()
	resolver := NewDelegationResolver(stateDB)

	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Set delegation using EIP-7702 delegation prefix
	delegationCode := types.AddressToDelegation(codeAddr)
	stateDB.SetCode(addr1, delegationCode)

	// Set some code
	testCode := []byte{0x60, 0x00, 0x60, 0x00, 0xf3}
	stateDB.SetCode(codeAddr, testCode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.ResolveDelegatedCode(addr1)
	}
}

func BenchmarkEVMDelegationResolution(b *testing.B) {
	stateDB := newMockStateDB()
	chainConfig := params.TestChainConfig
	
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

	evm := NewEVM(blockCtx, txCtx, stateDB, nil, chainConfig, vmConfig)

	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")

	// Set delegation and code using EIP-7702 delegation prefix
	delegationCode := types.AddressToDelegation(codeAddr)
	stateDB.SetCode(addr1, delegationCode)

	testCode := []byte{0x60, 0x00, 0x60, 0x00, 0xf3}
	stateDB.SetCode(codeAddr, testCode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evm.resolveDelegation(addr1)
	}
}