# Transaction Pool Architecture

The transaction pool is the entry point for all transactions in the U2U network. It manages incoming transactions, validates them, and organizes them for inclusion in events. This document explains the structure, validation, and management of the transaction pool.

## Pool Structure

The transaction pool (`TxPool`) maintains transactions in two main categories:

### Pending Transactions

Pending transactions are **executable** transactions that can be immediately applied to the current state:

- **Nonce Continuity**: The transaction's nonce matches the account's current nonce
- **State Validity**: The account has sufficient balance and gas
- **Ready for Execution**: All dependencies are satisfied

Pending transactions are organized by sender address in `pending map[common.Address]*txList`, where each account has its own list of transactions sorted by nonce.

### Queued Transactions

Queued transactions are **non-executable** transactions that cannot yet be applied:

- **Nonce Gaps**: The transaction's nonce is higher than the account's current nonce
- **Future Transactions**: Waiting for earlier transactions to be included
- **State Dependencies**: Require state changes from pending transactions

Queued transactions are stored in `queue map[common.Address]*txList` and automatically promoted to pending when their dependencies are satisfied.

### Transaction Lookup

All transactions are indexed in a global lookup structure (`all *txLookup`) that provides:

- **Hash-based Lookup**: Fast retrieval by transaction hash
- **Duplicate Detection**: Prevents the same transaction from being added twice
- **Local Transaction Tracking**: Special handling for locally submitted transactions

### Price-Ordered List

Transactions are also maintained in a price-ordered list (`priced *txPricedList`) that:

- **Prioritizes by Gas Price**: Higher gas price transactions are preferred
- **Enables Eviction**: Underpriced transactions can be removed when pool is full
- **Supports Replacement**: Allows replacing transactions with higher-priced versions

## Transaction Validation

Transactions undergo multi-layer validation before being accepted into the pool:

### Basic Validation

Basic validation checks transaction structure and format:

- **Transaction Type**: Supports Legacy, EIP-2930 (AccessList), and EIP-1559 (DynamicFee) transactions
- **Size Limits**: Maximum transaction size of 128KB (4 × 32KB slots)
- **Signature Validity**: Cryptographic signature verification
- **Chain ID**: Ensures transaction is for the correct network

### State Validation

State validation ensures the transaction can be executed:

- **Nonce Check**: Transaction nonce is valid (allows gaps for queued transactions)
- **Balance Check**: Account has sufficient balance for value + gas fees
- **Gas Limits**: Transaction gas limit is within acceptable bounds
- **Account Slots**: Respects per-account and global slot limits

### Price Validation

Price validation enforces minimum gas price requirements:

- **Minimum Gas Price**: Must meet network minimum (1 Gwei on mainnet/testnet)
- **Effective Min Tip**: Must meet effective minimum tip for EIP-1559 transactions
- **Local Transaction Exemption**: Locally submitted transactions bypass price checks

### Validation Flow

The validation process (`validateTx`) performs checks in this order:

1. **Basic Structure**: Transaction type, size, signature
2. **Price Check**: Gas price meets minimum requirements
3. **State Check**: Nonce, balance, and gas availability
4. **Pool Limits**: Account and global slot availability

If any check fails, the transaction is rejected with an appropriate error.

## Transaction Lifecycle

### Submission

Transactions enter the pool through:

- **RPC API**: Users submit via `eth_sendTransaction` or `eth_sendRawTransaction`
- **Network Propagation**: Transactions received from other peers
- **Local Submission**: Direct submission bypassing price checks

### Addition Process

When a transaction is added (`add` method):

1. **Duplicate Check**: Verify transaction isn't already in pool
2. **Validation**: Run full validation pipeline
3. **Pool Capacity**: Check if pool has space, evict underpriced if needed
4. **Replacement Check**: If nonce exists, check if replacement meets price bump (10%)
5. **Insertion**: Add to pending (if executable) or queue (if not)
6. **Indexing**: Add to lookup and price-ordered structures
7. **Journaling**: Save local transactions to disk for persistence

### Promotion

Queued transactions are automatically promoted to pending when:

- **Nonce Gap Filled**: Earlier transactions are included in blocks
- **State Updated**: Account state changes make transaction executable
- **Reorganization**: Chain reorganization makes previously queued transactions executable

The promotion process (`promoteExecutables`) runs periodically and after each new block.

### Removal

Transactions are removed from the pool when:

- **Inclusion**: Transaction is included in a finalized block
- **Eviction**: Pool is full and transaction is underpriced
- **Replacement**: Higher-priced transaction replaces it
- **Expiration**: Queued transaction exceeds lifetime (1 hour default)
- **Invalidation**: State changes make transaction invalid

## Pool Management

### Capacity Limits

The pool enforces several capacity limits to prevent resource exhaustion:

**Global Limits:**
- **Global Slots**: Maximum executable transactions (default: 2,560)
- **Global Queue**: Maximum queued transactions (default: 512)

**Per-Account Limits:**
- **Account Slots**: Maximum executable per account (default: 32)
- **Account Queue**: Maximum queued per account (default: 128)

**Transaction Lifetime:**
- **Queued Lifetime**: Maximum time in queue (default: 1 hour)

### Eviction Policy

When the pool reaches capacity:

1. **Local Protection**: Local transactions are never evicted
2. **Price-Based**: Underpriced transactions are removed first
3. **Age-Based**: Old queued transactions are evicted after lifetime
4. **Replacement Throttle**: Limits replacements between reorganization runs (25% of slots)

### Price Bump Requirement

To replace an existing transaction, the new transaction must:

- **Meet Price Bump**: Gas price must be at least 10% higher
- **Same Nonce**: Must target the same nonce as existing transaction
- **Valid Replacement**: Pass all validation checks

This prevents transaction spam and ensures replacements are economically meaningful.

## Local Transactions

Local transactions receive special treatment:

### Definition

A transaction is considered "local" if:

- **Direct Submission**: Submitted directly via RPC (not from network)
- **Marked Address**: Sender address is in the local allowlist
- **Previously Local**: Transaction from an address that previously submitted locally

### Special Handling

Local transactions:

- **Price Exemption**: Bypass minimum gas price requirements
- **Eviction Protection**: Never evicted due to low price
- **Journal Persistence**: Saved to disk for recovery after restart
- **Priority**: Treated preferentially in selection

### Journal Management

Local transactions are persisted to a journal file (`transactions.rlp`) that:

- **Survives Restarts**: Transactions are reloaded on node restart
- **Periodic Updates**: Journal is updated periodically (default: 1 hour)
- **Recovery**: Ensures local transactions aren't lost

## State Synchronization

The pool maintains synchronization with blockchain state:

### Chain Head Updates

When a new block is finalized:

1. **State Update**: Current state is updated to new block's state
2. **Transaction Removal**: Included transactions are removed from pool
3. **Nonce Updates**: Account nonces are updated based on included transactions
4. **Promotion**: Queued transactions are promoted if now executable
5. **Reorganization**: Handles chain reorganizations if they occur

### Reorganization Handling

During chain reorganizations:

1. **State Rollback**: Pool state is rolled back to common ancestor
2. **Transaction Revalidation**: All transactions are revalidated against new state
3. **Removal**: Invalid transactions are removed
4. **Promotion**: Previously queued transactions may become executable

The pool tracks changes between reorganizations to throttle replacements and prevent excessive churn.

## Configuration

Transaction pool behavior is configured through `TxPoolConfig`:

```go
type TxPoolConfig struct {
    Locals    []common.Address // Local addresses
    NoLocals  bool             // Disable local handling
    Journal   string           // Journal file path
    Rejournal time.Duration    // Journal update interval
    
    PriceLimit uint64 // Minimum gas price
    PriceBump  uint64 // Replacement price bump (percentage)
    
    AccountSlots uint64 // Per-account executable slots
    GlobalSlots uint64 // Global executable slots
    AccountQueue uint64 // Per-account queued slots
    GlobalQueue uint64 // Global queued slots
    
    Lifetime time.Duration // Queued transaction lifetime
}
```

Default values provide a balance between capacity and resource usage, but can be adjusted for specific use cases.

## Summary

The transaction pool provides:

- **Efficient Organization**: Separates executable and queued transactions
- **Comprehensive Validation**: Multi-layer validation ensures transaction safety
- **Fair Management**: Price-based prioritization with local transaction protection
- **State Synchronization**: Maintains consistency with blockchain state
- **Resource Protection**: Capacity limits prevent resource exhaustion

The pool serves as the foundation for transaction processing, ensuring that valid transactions are available for validators to include in events while maintaining network security and efficiency.

For implementation details, see `evmcore/tx_pool.go`, `evmcore/tx_list.go`, and related validation code.

