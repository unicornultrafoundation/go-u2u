# Transaction Selection and Prioritization

Validators select transactions from the pool to include in events based on a sophisticated algorithm that balances fairness, efficiency, and economic incentives. This document explains how transactions are prioritized, selected, and included in events.

## Selection Overview

Transaction selection occurs when a validator creates a new event. The process:

1. **Retrieves Pending Transactions**: Gets all executable transactions from the pool
2. **Sorts by Price and Nonce**: Orders transactions by gas price, then by nonce
3. **Applies Selection Criteria**: Filters based on gas power, conflicts, and turn-based rules
4. **Includes in Event**: Adds selected transactions to the event payload

The selection process is implemented in `gossip/emitter/txs.go` and `gossip/emitter/emitter.go`.

## Price-Based Ordering

Transactions are primarily ordered by **gas price**, which determines the fee users are willing to pay:

### Gas Price Calculation

For different transaction types:

- **Legacy Transactions**: Use `gasPrice` directly
- **EIP-1559 Transactions**: Use `min(gasFeeCap, baseFee + gasTipCap)`
- **AccessList Transactions**: Use `gasPrice` directly

The effective gas price determines transaction priority in the selection queue.

### Price-Ordered Structure

The `TransactionsByPriceAndNonce` structure maintains transactions in price order:

- **Heap-Based**: Uses a heap to efficiently maintain price ordering
- **Per-Account Queues**: Each account's transactions are ordered by nonce
- **Global Ordering**: Highest price transactions are selected first

### Minimum Gas Price

Transactions must meet the network's minimum gas price (1 Gwei on mainnet/testnet) to be considered. This prevents spam and ensures transaction fees have meaningful value.

## Nonce Handling

Within each account, transactions are ordered by **nonce** to ensure sequential execution:

### Sequential Execution

- **Nonce Continuity**: Transactions must be executed in nonce order
- **Gap Handling**: Nonce gaps prevent later transactions from being selected
- **Account Isolation**: Each account's transactions are independent

### Nonce-Based Selection

When selecting transactions:

1. **Start with Lowest Nonce**: Always select the lowest nonce transaction first
2. **Progress Sequentially**: Move to next nonce only after current is selected
3. **Respect Gaps**: Cannot skip nonces, even if later transactions have higher prices

This ensures that transactions from the same account are always processed in order, maintaining state consistency.

## Selection Criteria

When a validator selects transactions for an event, several criteria are applied:

### Gas Power Availability

The validator must have sufficient **gas power** to include the transaction:

- **Gas Power Left**: Remaining gas power must exceed transaction gas limit
- **Maximum Event Gas**: Transaction cannot exceed maximum event gas limit (10,000,028 gas)
- **Cumulative Limit**: Total gas in event cannot exceed available gas power

If a transaction would exceed available gas power, it is skipped and the selection process continues with the next transaction.

### Conflict Detection

Transactions are checked for conflicts to prevent double-inclusion:

- **Already Originated**: Transaction from same sender already included in a connected event
- **Nonce Conflicts**: Cannot include multiple transactions with same nonce from same account
- **Pool Presence**: Transaction must still be in pool (not already included elsewhere)

The `originatedTxs` buffer tracks which transactions have been included to prevent conflicts.

### Turn-Based Selection

To prevent multiple validators from simultaneously including the same transaction, a **turn-based system** is used:

#### Turn Calculation

The turn is determined by:

1. **Transaction Time**: When the transaction was first seen (`txtime.Of(txHash)`)
2. **Round Index**: Calculated based on time elapsed divided by turn period (8 seconds)
3. **Validator Permutation**: Weighted permutation of validators based on round index
4. **Turn Assignment**: Each round assigns transactions to specific validators

#### Turn Validation

A transaction can only be included if:

- **Current Turn**: The validator's turn for this transaction's round
- **Stable Round**: Round is not about to change (within 1 second of change)
- **Nonce Grouping**: Transactions are grouped by nonce ranges (32 nonces per group)

This mechanism ensures fair distribution of transaction inclusion across validators while preventing race conditions.

### Epoch Rules

Transactions must comply with epoch-specific rules:

- **Epoch Validity**: Transaction must be valid for current epoch
- **Gas Price Rules**: Must meet epoch's minimum gas price
- **Transaction Type**: Must be a supported transaction type

Epoch rules are checked via `epochcheck.CheckTxs()` before selection.

## Transaction Caching

To optimize performance, the emitter caches sorted transactions:

### Cache Structure

The cache (`cache.sortedTxs`) stores:

- **Sorted Transactions**: Pre-sorted `TransactionsByPriceAndNonce` structure
- **Cache Time**: When the cache was built
- **Pool Block**: Block number when cache was built
- **Pool Count**: Number of transactions in pool at cache time

### Cache Invalidation

The cache is invalidated when:

- **Pool Updated**: New transactions added or removed
- **Block Advanced**: New block finalized (state may have changed)
- **Timeout**: Cache age exceeds invalidation period (configurable)

When invalidated, the cache is rebuilt from current pool state.

### Cache Benefits

Caching provides:

- **Reduced Sorting**: Avoids re-sorting transactions on every event creation
- **Performance**: Faster event creation, especially with large pools
- **Consistency**: Ensures consistent ordering during event creation

## Selection Process

The transaction selection process (`addTxs`) works as follows:

### Iteration

1. **Peek Next Transaction**: Get highest-priority transaction from sorted list
2. **Apply Filters**: Check all selection criteria
3. **Include or Skip**: Add transaction to event or skip to next
4. **Update Gas Power**: Track gas power usage
5. **Continue**: Repeat until gas power exhausted or no valid transactions

### Filtering Steps

For each transaction, filters are applied in order:

1. **Epoch Rules**: Check transaction validity for current epoch
2. **Gas Power**: Verify sufficient gas power available
3. **Conflict Check**: Ensure no conflicts with already-originated transactions
4. **Turn Check**: Verify it's the validator's turn for this transaction
5. **Pool Presence**: Confirm transaction still in pool

If any filter fails, the transaction is skipped and the next one is considered.

### Gas Power Tracking

As transactions are selected:

- **Gas Power Used**: Accumulated for all included transactions
- **Gas Power Left**: Decremented for each transaction
- **Maximum Limit**: Cannot exceed `maxGasPowerToUse()` limit

The `maxGasPowerToUse()` function calculates the maximum gas that can be used in an event based on:
- Available gas power
- Pending gas in network
- Smooth TPS thresholds
- Emergency thresholds

## Fairness Mechanisms

Several mechanisms ensure fair transaction inclusion:

### Per-Address Limits

The emitter limits transactions per address:

- **MaxTxsPerAddress**: Maximum transactions per sender (configurable)
- **Prevents Monopolization**: Prevents single accounts from dominating event space
- **Fair Distribution**: Ensures multiple users' transactions are included

### Turn-Based Distribution

The turn-based system ensures:

- **Validator Rotation**: Each validator gets turns for different transactions
- **Time-Based**: Turns rotate every 8 seconds
- **Weighted**: Validator selection is weighted by stake

### Gas Power Allocation

Gas power allocation (covered in detail in `03_gas_power_allocation.md`) ensures:

- **Stake-Based**: Validators with more stake get more gas power
- **Fair Distribution**: All validators can participate
- **Prevents Monopolization**: No single validator can dominate

## Performance Optimizations

Several optimizations improve selection performance:

### Sorted Transaction List

- **Pre-Sorted**: Transactions are sorted once and reused
- **Efficient Peek**: O(1) access to highest-priority transaction
- **Incremental Updates**: Only re-sort when necessary

### Early Termination

Selection stops early when:

- **Gas Power Exhausted**: No more gas power available
- **No Valid Transactions**: All remaining transactions fail filters
- **Maximum Reached**: Event gas limit reached

### Caching

- **Sorted List Cache**: Avoids re-sorting on every event
- **Pool State Cache**: Tracks pool state to detect changes
- **Time-Based Invalidation**: Cache expires after timeout

## Configuration

Transaction selection is configured through emitter settings:

- **MaxTxsPerAddress**: Maximum transactions per sender
- **TxsCacheInvalidation**: Cache invalidation timeout
- **NoTxsThreshold**: Gas power threshold below which no transactions are included
- **LimitedTpsThreshold**: Threshold for TPS smoothing
- **EmergencyThreshold**: Emergency gas power threshold

These settings balance performance, fairness, and network efficiency.

## Summary

Transaction selection provides:

- **Price-Based Prioritization**: Higher gas price transactions are selected first
- **Nonce Ordering**: Sequential execution within accounts
- **Fair Distribution**: Turn-based system prevents monopolization
- **Conflict Prevention**: Prevents double-inclusion and conflicts
- **Performance Optimization**: Caching and efficient algorithms

The selection process ensures that transactions are included fairly and efficiently while respecting economic incentives and network constraints.

For implementation details, see `gossip/emitter/txs.go`, `gossip/emitter/emitter.go`, and `core/types/transaction.go`.



