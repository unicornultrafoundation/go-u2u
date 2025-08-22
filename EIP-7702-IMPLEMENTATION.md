# EIP-7702 Implementation Documentation

## Overview

This document describes the implementation of EIP-7702 "Set EOA account code" in the U2U blockchain. EIP-7702 allows externally owned accounts (EOAs) to temporarily delegate their code execution to another contract during a transaction.

## Architecture

### Core Components

1. **SetCodeTx Transaction Type** (`core/types/setcode_tx.go`)
   - New transaction type (type 3) for account delegation
   - Contains authorization list with signed authorizations
   - Supports EIP-1559 gas fee mechanism
   - RLP encoding/decoding support

2. **Authorization Processing** (`core/types/setcode_tx_processing.go`)
   - Delegation chain resolution with circular detection
   - Authorization signature verification
   - Nonce validation for delegation security

3. **Transaction Pool Integration** (`evmcore/setcode_tx_pool.go`)
   - Pool-level validation for SetCode transactions
   - State-aware validation with balance and nonce checking
   - Integration with existing validation framework

4. **State Transition Processing** (`evmcore/setcode_tx_state.go`)
   - Applies SetCode transactions to state
   - Manages delegation mappings in state database
   - Gas calculation and receipt generation

5. **EVM Delegation Support** (`core/vm/setcode_delegation.go`)
   - Enhanced EVM with delegation resolution
   - Delegation-aware CALL, DELEGATECALL, and STATICCALL
   - Integration with existing EVM infrastructure

## Data Structures

### AuthorizationTuple
```go
type AuthorizationTuple struct {
    ChainID *big.Int    // Chain ID for replay protection
    Address common.Address // Code address to delegate to
    Nonce   uint64      // Account nonce for replay protection
    V, R, S *big.Int    // ECDSA signature components
}
```

### SetCodeTx
```go
type SetCodeTx struct {
    ChainID           *big.Int
    Nonce             uint64
    To                *common.Address
    Value             *big.Int
    Gas               uint64
    GasFeeCap         *big.Int
    GasTipCap         *big.Int
    Data              []byte
    AccessList        AccessList
    AuthorizationList AuthorizationList
    V, R, S           *big.Int
}
```

## Processing Flow

### 1. Authorization Validation
```go
// Signature verification
authority, err := auth.RecoverAuthority()
if err != nil {
    return fmt.Errorf("invalid authorization signature")
}

// Nonce validation
currentNonce := getAccountNonce(authority)
if auth.Nonce != currentNonce {
    return fmt.Errorf("nonce mismatch")
}
```

### 2. Delegation Chain Resolution
```go
// Resolve delegation chain with circular detection
resolver := NewDelegationChainResolver()
finalAddress, err := resolver.ResolveDelegationChain(startAddr, getDelegation)
```

### 3. State Application
```go
// Store delegation in state
delegationKey := common.BytesToHash(append([]byte("EIP7702_DELEGATION_"), authority.Bytes()...))
statedb.SetState(common.HexToAddress("0x7702"), delegationKey, codeAddr.Hash())
```

### 4. EVM Execution
```go
// Resolve delegation during EVM calls
if delegationResolver.CheckDelegation(addr) {
    resolvedAddr, code, err := delegationResolver.ResolveDelegatedCode(addr)
    // Execute using delegated code
}
```

## Security Considerations

### 1. Circular Delegation Prevention
- Maximum delegation depth limit (16)
- Visited address tracking during resolution
- Self-delegation detection

### 2. Replay Protection
- Chain ID inclusion in authorization signatures
- Nonce-based replay protection
- Authority signature verification

### 3. Gas Calculation
```go
intrinsicGas = baseTxGas + 
               (dataSize * dataCostPerByte) + 
               (authCount * authorizationCost) +
               accessListGas
```

### 4. State Isolation
- Delegations stored in special system address (0x7702)
- Deterministic storage keys based on authority address
- Clean separation from regular account state

## Gas Costs

| Operation | Gas Cost |
|-----------|----------|
| Base Authorization | 3000 |
| Authorization Signature | 3000 |
| Delegation Resolution | Variable |
| Total per Authorization | 6000 |

## Integration Points

### Transaction Pool
- SetCodeTx type accepted in pool validation
- Extended Accept bitmask: `1 << types.SetCodeTxType`
- State-aware authorization validation

### State Database
- Delegation mappings at system address 0x7702
- Storage key: `keccak256("EIP7702_DELEGATION_" + authority.bytes)`
- Value: delegated code address

### EVM Enhancement
- Enhanced EVM wrapper for delegation support
- Delegation-aware call methods
- Backward compatibility with standard EVM

## Testing

### Test Coverage
1. **Unit Tests**
   - Authorization tuple validation
   - Delegation chain resolution
   - Gas calculation
   - RLP encoding/decoding

2. **Integration Tests**
   - Transaction pool validation
   - State transition processing
   - EVM delegation execution

3. **Security Tests**
   - Circular delegation detection
   - Invalid signature handling
   - Nonce mismatch scenarios

### Test Files
- `core/types/setcode_tx_test.go` - Basic SetCode transaction tests
- `core/types/setcode_tx_processing_test.go` - Processing logic tests
- `evmcore/setcode_tx_pool_test.go` - Pool validation tests
- `evmcore/setcode_tx_state_test.go` - State transition tests
- `core/vm/setcode_delegation_test.go` - EVM delegation tests

## Configuration

### Chain Configuration
EIP-7702 support can be enabled through chain configuration:
```go
// Future: Add EIP-7702 fork configuration
if config.IsPrague(blockNumber) {
    // Enable EIP-7702 support
}
```

### Gas Parameters
```go
const (
    AuthorizationBaseGas     = 3000
    AuthorizationSignatureGas = 3000
    MaxDelegationDepth       = 16
    MaxAuthorizationListSize = 256
)
```

## Limitations and Future Work

### Current Limitations
1. Basic EVM execution (placeholder implementation)
2. Limited precompile integration
3. No cross-client compatibility testing

### Future Enhancements
1. Full EVM integration with delegation context
2. Precompile delegation support
3. Performance optimizations
4. Cross-client test vectors

## References

- [EIP-7702 Specification](https://eips.ethereum.org/EIPS/eip-7702)
- [EIP-1559 Gas Fee Mechanism](https://eips.ethereum.org/EIPS/eip-1559)
- [EIP-2930 Access Lists](https://eips.ethereum.org/EIPS/eip-2930)

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2024 | Initial EIP-7702 implementation |