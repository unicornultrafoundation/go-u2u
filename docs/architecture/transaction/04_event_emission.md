# Event Emission Logic

Event emission is the process by which validators create and broadcast events containing transactions. This document explains when validators emit events, how they select parents, construct events, and control emission timing.

## Emission Overview

Event emission occurs when a validator:

1. **Determines Readiness**: Checks if conditions are met to create an event
2. **Selects Parents**: Chooses parent events for the DAG
3. **Builds Event**: Constructs event structure with transactions and consensus data
4. **Validates Event**: Ensures event is valid before broadcasting
5. **Broadcasts Event**: Sends event to peers for propagation

The emission process is implemented in `gossip/emitter/emitter.go` and related files.

## Emission Timing

Validators don't emit events continuously. Instead, emission is controlled by several factors:

### Minimum Interval

Validators must wait at least `Min` interval (configurable, typically ~100ms) between events:

- **Rate Limiting**: Prevents validators from creating events too frequently
- **Network Stability**: Reduces network load and improves stability
- **Efficiency**: Ensures events have meaningful content

### Maximum Interval

Validators must emit within `Max` interval (configurable, typically ~1s) to maintain participation:

- **Liveness Requirement**: Ensures validators stay active
- **Block Missed Prevention**: Prevents validators from missing too many blocks
- **Network Progress**: Ensures network continues to make progress

### Adaptive Intervals

Emission intervals are adjusted based on:

- **Efficiency Metric**: How much the event advances consensus
- **Gas Power**: Available gas power affects emission timing
- **Network Conditions**: Busy rate and network load influence timing
- **Stake Ratio**: Validators with different stake ratios have different timing

### Efficiency-Based Timing

The emission interval is adjusted by an **efficiency metric** that estimates how much an event advances consensus:

```
AdjustedInterval = BaseInterval × (1 / EfficiencyMetric)
```

Higher efficiency events can be emitted more frequently, while lower efficiency events are spaced out.

## Parent Selection

Each event must reference parent events in the DAG. Parent selection determines which events to reference:

### Self-Parent

The **self-parent** is the validator's most recent event:

- **Sequential Chain**: Maintains validator's sequential event chain
- **Gas Power Continuity**: Ensures gas power state is continuous
- **Required**: Always included as first parent (if exists)

### Other Parents

Additional parents are selected using search strategies:

1. **Payload Strategy**: Selects parents based on transaction content
2. **Random Strategy**: Randomly selects from available heads
3. **Quorum Strategy**: Selects parents that advance consensus (FC or Quorum indexer)

### Parent Selection Process

The `chooseParents` function:

1. **Gets Self-Parent**: Retrieves validator's last event in current epoch
2. **Gets Heads**: Finds all events with no descendants (DAG heads)
3. **Applies Strategies**: Uses search strategies to select additional parents
4. **Validates**: Ensures selected parents are valid and don't create forks

### Maximum Parents

Events can reference up to `maxParents` parents (configurable, typically 5-10):

- **DAG Connectivity**: More parents improve DAG connectivity
- **Consensus Progress**: Better parent selection advances consensus faster
- **Network Efficiency**: Balances connectivity with event size

## Event Construction

Event construction builds the complete event structure:

### Base Fields

- **Version**: Event serialization version (0 or 1)
- **Epoch**: Current epoch number
- **Sequence**: Sequential number from validator (self-parent sequence + 1)
- **Creator**: Validator ID creating the event
- **Parents**: References to parent events
- **Lamport**: Logical timestamp (max parent Lamport + 1)
- **Creation Time**: Wall-clock time of creation

### Consensus Fields

- **Gas Power Left**: Remaining gas power after previous events
- **Gas Power Used**: Gas consumed by transactions in this event
- **Median Time**: Median time across validators for this event
- **Previous Epoch Hash**: Hash of previous epoch state (if applicable)

### Payload

- **Transactions**: Selected transactions from the pool
- **Block Votes**: Votes on block records (LLR layer)
- **Epoch Vote**: Vote on epoch record (LLR layer)
- **Misbehaviour Proofs**: Proofs of validator misbehaviour (if any)
- **Extra Data**: Optional arbitrary bytes (e.g., node version)

### Construction Process

1. **Initialize Event**: Create `MutableEventPayload` with base fields
2. **Set Parents**: Add parent references
3. **Set Timing**: Set creation time and Lamport timestamp
4. **Add LLR Votes**: Add block and epoch votes
5. **Build Gas Power**: Calculate and set gas power left
6. **Add Transactions**: Select and add transactions
7. **Set Payload Hash**: Calculate payload hash
8. **Sign Event**: Sign event with validator's private key
9. **Build Final Event**: Create immutable `EventPayload`

## Emission Control

Several mechanisms control when events can be emitted:

### Pre-Emission Checks

Before creating an event, validators check:

- **Validator Status**: Must be an active validator in current epoch
- **Sync Status**: Must be synced to emit (not reindexing)
- **Gas Power**: Must have sufficient gas power (or event will be empty)
- **No Fork**: Must not have created a fork (multiple self-parents)

### Emission Permission

The `isAllowedToEmit` function determines if emission is allowed:

#### Time-Based Checks

- **Minimum Time**: Must have passed minimum interval since last emission
- **Maximum Time**: Must emit within maximum interval (enforced)
- **Adjusted Time**: Efficiency-adjusted time must have passed

#### Gas Power Checks

- **Emergency Threshold**: If gas power below threshold and decreasing, emission forbidden
- **Low Power Slowdown**: Emission slowed when gas power is low
- **No Txs Threshold**: No transactions included if gas power too low

#### Idle Checks

- **Idle Time**: If no transactions to confirm, must wait longer
- **Confirming Interval**: Special interval for confirming transactions

#### Efficiency Checks

- **Efficiency Metric**: Event must have sufficient efficiency to advance consensus
- **Adjusted Intervals**: Intervals adjusted based on efficiency metric

### Forced Emission

Emission is forced (allowed even if conditions not met) when:

- **Maximum Time Exceeded**: Too much time since last emission
- **Block Missed Threshold**: Approaching block missed threshold
- **Network Progress**: Network needs progress even if conditions suboptimal

## Conflict Prevention

Several mechanisms prevent conflicts and double-signing:

### Double-Sign Prevention

- **Last Event Tracking**: Tracks last emitted event ID
- **Database Check**: Verifies no event already exists at same sequence
- **Permanent Lock**: Locks permanently if double-sign detected
- **File Persistence**: Writes last event ID to file for crash recovery

### Fork Prevention

- **Single Self-Parent**: Only one self-parent allowed per event
- **Fork Detection**: Detects if multiple self-parents exist
- **Emission Blocking**: Blocks emission if fork detected

### Transaction Conflict Prevention

- **Originated Txs Buffer**: Tracks transactions already included
- **Conflict Check**: Prevents including same transaction twice
- **Sender Tracking**: Tracks transactions per sender to prevent conflicts

## Event Validation

Before broadcasting, events are validated:

### Structural Validation

- **Field Validity**: All fields must be valid
- **Parent Existence**: All parent events must exist
- **Sequence Continuity**: Sequence must be self-parent sequence + 1
- **Time Validity**: Creation time must be valid

### Gas Power Validation

- **Gas Power Calculation**: Gas power left must be correctly calculated
- **Gas Power Usage**: Gas power used must not exceed available
- **Gas Power Continuity**: Gas power must be continuous from self-parent

### Consensus Validation

- **Epoch Validity**: Event must be for current epoch
- **Validator Validity**: Creator must be active validator
- **LLR Votes**: Block and epoch votes must be valid

### Signature Validation

- **Signature Verification**: Event signature must be valid
- **Creator Match**: Signature must match creator validator
- **Hash Verification**: Signature must be over correct hash

## Broadcasting

After validation, events are broadcast to peers:

### Broadcast Process

1. **Process Locally**: Event is processed and added to local DAG
2. **Persist State**: Last emitted event ID is written to file
3. **Broadcast**: Event is sent to all connected peers
4. **Track Metrics**: Emission metrics are updated

### Propagation

- **Gossip Protocol**: Event propagates through gossip protocol
- **Peer Selection**: Events sent to all connected peers
- **Deduplication**: Peers filter duplicate events
- **Validation**: Peers validate events before accepting

## Performance Optimizations

Several optimizations improve emission performance:

### Caching

- **Sorted Transactions**: Transaction list is cached
- **Parent Selection**: Parent selection strategies are cached
- **Gas Power Calculation**: Gas power calculations are optimized

### Lazy Evaluation

- **Transaction Selection**: Transactions selected only when needed
- **Parent Resolution**: Parents resolved only when needed
- **Validation**: Validation performed only when necessary

### Batch Operations

- **Transaction Batching**: Multiple transactions processed together
- **Parent Batching**: Multiple parents selected together
- **Validation Batching**: Validations performed in batches

## Configuration

Event emission is configured through emitter settings:

- **EmitIntervals**: Minimum and maximum emission intervals
- **MaxParents**: Maximum number of parents per event
- **Validator**: Validator ID and public key
- **TxsCacheInvalidation**: Transaction cache invalidation timeout
- **EmergencyThreshold**: Emergency gas power threshold
- **NoTxsThreshold**: Threshold below which no transactions included

These settings balance emission frequency, network load, and consensus progress.

## Summary

Event emission provides:

- **Controlled Timing**: Adaptive intervals based on efficiency and conditions
- **Smart Parent Selection**: Strategies for optimal parent choice
- **Complete Construction**: Full event structure with all required fields
- **Conflict Prevention**: Mechanisms to prevent double-signing and forks
- **Validation**: Comprehensive validation before broadcasting
- **Efficient Propagation**: Optimized broadcasting to peers

The emission system ensures validators create events efficiently while maintaining network security and consensus progress.

For implementation details, see `gossip/emitter/emitter.go`, `gossip/emitter/control.go`, `gossip/emitter/parents.go`, and related files.



