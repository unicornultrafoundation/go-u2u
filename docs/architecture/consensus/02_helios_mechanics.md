# Helios Consensus Mechanics

This document provides a deeper technical understanding of how the Helios consensus mechanism works at the implementation level. It explains the internal mechanics that transform a DAG of events into a linear blockchain while maintaining security and consistency.

## Event Structure

Events are the atomic units of consensus in the Helios DAG. Each event is a vertex in the graph that contains transactions, votes, or other consensus-related data. The event structure is defined in `native/event.go` and represents a rich data structure that encodes both its position in the DAG and its payload content.

### Event Components

An event consists of several key components:

**Base Event Data** (from `dag.BaseEvent`):
- **Epoch**: The epoch number in which this event was created
- **Lamport**: A logical timestamp that helps order events
- **Frame**: A logical round within the epoch
- **Creator**: The validator ID that created this event
- **Sequence**: The sequence number of this event from its creator
- **Parents**: References to previous events in the DAG
- **ID**: A unique identifier derived from the event's content

**Extended Event Data** (from `extEventData`):
- **Version**: Serialization version for forward compatibility
- **Network Fork ID**: Identifies the network fork version
- **Creation Time**: Wall-clock time when the event was created
- **Median Time**: Median time across validators for this event
- **Previous Epoch Hash**: Hash of the previous epoch state
- **Gas Power Left**: Remaining gas power available to the creator
- **Gas Power Used**: Gas consumed by this event
- **Extra Data**: Arbitrary bytes for future extensions

**Payload** (from `payloadData`):
- **Transactions**: User transactions included in this event
- **Block Votes**: Votes on block records (for LLR layer)
- **Epoch Vote**: Vote on epoch records
- **Misbehaviour Proofs**: Proofs of validator misbehaviour

### Event Serialization

Events are serialized using a compact format defined in `native/event_serializer.go`. This format is designed to minimize network bandwidth while preserving all necessary information for validation and consensus.

The serialization process:

1. Encodes base fields (epoch, Lamport, creator, sequence, frame)
2. Stores timing information (creation time and median time difference)
3. Compresses parent references using Lamport time differences
4. Includes optional fields (previous epoch hash, payload hash)
5. Appends the signature and full payload if present

This efficient encoding allows validators to quickly exchange event blocks even when the DAG contains thousands of events.

## The Ordering Engine

The core innovation of Helios is its ability to order events in a DAG without requiring sequential block production. This ordering is achieved through a multi-stage process that transforms the DAG into a linear blockchain.

### Roots

A root event is one that has been observed by a sufficient number of validators. The root detection algorithm (implemented in the upstream `go-helios` library) analyzes the DAG structure to identify events that have achieved sufficient reachability across the validator set.

Roots serve as synchronization points in the DAG. They indicate that enough validators have seen and acknowledged an event, making it a candidate for inclusion in the final blockchain.

### Clotho

A Clotho is a root event that has been confirmed by subsequent frames. The Clotho selection process involves checking that later frames contain roots that attest to having seen the candidate root. This multi-frame voting mechanism ensures that Clotho events have strong consensus support.

The Clotho determination logic is handled by the upstream Helios DAG engine. In this codebase, we see the results of Clotho selection through:

- Block votes and epoch votes embedded in event payloads
- Misbehaviour proofs that detect inconsistent voting
- Epoch and block records in the LLR (Lower-Level Record) layer

### Atropos

An Atropos is a Clotho event that has been fully finalized by the consensus algorithm. Each Atropos event anchors a block in the main chain. The Atropos selection ensures that all validators agree on which events should be included in each block.

When an Atropos is decided, the consensus engine calls back into the block processing logic through `utypes.ConsensusCallbacks` (implemented in `gossip/c_block_callbacks.go`). This callback receives:

- The Atropos event ID
- The set of cheaters detected through misbehaviour proofs
- The confirmed events associated with this block

The block processing logic then:

1. Determines the block time from the Atropos event's median time
2. Collects all confirmed events that contain transactions
3. Processes misbehaviour proofs to identify cheating validators
4. Executes transactions through the EVM
5. Creates and persists the final block

## Graph Traversal and Validation

Validators must traverse and validate the DAG to ensure events are consistent and safe before they influence consensus and state updates.

### Ancestor Traversal

Each event references its parents, creating a directed graph structure. Validators traverse this graph to:

- Verify that all parent events exist and are valid
- Compute reachability relationships between events
- Determine frame membership and root eligibility
- Detect cycles or other structural inconsistencies

The parent references are encoded efficiently in serialization: only the Lamport time difference and a truncated hash are stored for each parent, allowing validators to reconstruct the full parent IDs deterministically.

### Validation Pipeline

Events undergo multiple layers of validation before being accepted into the DAG:

**Light Check** (pre-buffering):
- Verifies the event's epoch matches the current epoch
- Rejects duplicate events
- Performs basic structural validation
- Checks epoch-specific rules

**Buffered Check** (with parents):
- Validates parent set structure and consistency
- Verifies gas power usage against allocation rules
- Ensures no cycles or structural violations

**Heavy Check** (optional):
- Performs expensive cryptographic validations
- Checks data availability
- Validates misbehaviour proofs

Once an event passes all validations, it is:

- Connected into the DAG store
- Processed by the consensus layer
- Broadcasted to other peers

### Misbehaviour Detection

The system detects validator misbehaviour through proofs embedded in event payloads. These proofs can detect:

- **Double-signing**: A validator signing two conflicting events, block votes, or epoch votes
- **Wrong votes**: Votes that contradict the canonical block or epoch record

When misbehaviour is detected:

1. The proof is extracted from the event payload
2. The cheating validator is identified
3. The validator is added to the epoch's cheater list
4. Economic penalties are applied during epoch sealing

This two-layer approach (detection in consensus, enforcement in SFC/staking) ensures that misbehaviour is reliably detected and punished.

## Block Construction

When an Atropos event is decided, the system constructs a block that includes:

- **Block Context**: Index, time, and Atropos event ID
- **Confirmed Events**: All events associated with this Atropos that contain transactions
- **Transactions**: Extracted and ordered from the confirmed events
- **State Roots**: Both EVM state root and SFC (staking) state root
- **Gas Usage**: Total gas consumed by the block
- **Skipped Transactions**: Transactions that couldn't be included due to gas limits

The block construction process (`gossip/c_block_callbacks.go`) ensures that:

- Blocks are skipped if they're empty and within the maximum skip period
- Events are sorted by Lamport time for deterministic ordering
- Gas limits are respected (events may be spilled to subsequent blocks)
- Both EVM and SFC states are updated atomically

## Summary

Helios consensus mechanics transform a DAG of events into a linear blockchain through:

1. **Event Creation**: Validators create events containing transactions
2. **DAG Propagation**: Events are asynchronously shared across the network
3. **Root Detection**: Events that reach enough validators become roots
4. **Clotho Selection**: Roots confirmed by later frames become Clotho
5. **Atropos Finalization**: Clotho events are finalized as block anchors
6. **Block Construction**: Blocks are built from Atropos events and their confirmed transactions

This process provides strong finality guarantees while maintaining high throughput through asynchronous processing. The DAG structure allows validators to work concurrently without waiting for sequential block production, significantly improving transaction speed compared to traditional blockchain architectures.

For implementation details, see `native/event.go`, `gossip/handler.go`, `gossip/c_block_callbacks.go`, and the upstream `go-helios` library documentation.

