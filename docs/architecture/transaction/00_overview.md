# Transaction Pool and Event Emission Documentation Overview

This directory contains comprehensive documentation about how transactions flow through the U2U network, from user submission through event creation and inclusion in the consensus DAG.

## Documentation Structure

The transaction documentation is organized into several documents that progressively build understanding from transaction pool architecture to event emission mechanics:

### [01_transaction_pool.md](./01_transaction_pool.md)
**Transaction Pool Architecture**

This document explains how the transaction pool manages incoming transactions:

- **Pool Structure**: Pending vs queued transactions, account-based organization
- **Transaction Validation**: Multi-layer validation including signature, nonce, gas, and state checks
- **Transaction Lifecycle**: From submission through inclusion in events
- **Pool Management**: Eviction policies, replacement rules, and capacity limits
- **Local Transactions**: Special handling for locally submitted transactions

**Best for**: Developers integrating with U2U, those understanding transaction handling, or debugging transaction pool issues.

### [02_transaction_selection.md](./02_transaction_selection.md)
**Transaction Selection and Prioritization**

This document details how validators select transactions for inclusion in events:

- **Price-Based Ordering**: How gas price determines transaction priority
- **Nonce Handling**: Sequential transaction ordering per account
- **Selection Criteria**: Gas power limits, conflict detection, and turn-based selection
- **Transaction Caching**: Performance optimizations for transaction selection
- **Fairness Mechanisms**: Preventing validator monopolization of transaction inclusion

**Best for**: Validators optimizing event creation, developers understanding transaction ordering, or those interested in transaction economics.

### [03_gas_power_allocation.md](./03_gas_power_allocation.md)
**Gas Power Allocation System**

This document explains how validators receive and use gas power:

- **Allocation Mechanisms**: Long-term, short-term, and startup gas power allocation
- **Stake-Based Distribution**: How validator stake determines gas power capacity
- **Accumulation Rules**: Time-based gas power accumulation and maximum limits
- **Usage Tracking**: How gas power is consumed and tracked across events
- **Refund System**: Gas refunds and their role in gas power management

**Best for**: Validators managing their gas power, stakers understanding validator economics, or those studying the gas power system.

### [04_event_emission.md](./04_event_emission.md)
**Event Emission Logic**

This document covers how validators create and emit events:

- **Emission Timing**: When validators are allowed to create events
- **Parent Selection**: How validators choose parent events for the DAG
- **Event Construction**: Building event structure with transactions and consensus data
- **Emission Control**: Rate limiting, efficiency metrics, and adaptive intervals
- **Conflict Prevention**: Mechanisms to prevent double-signing and forks

**Best for**: Validators setting up nodes, developers implementing validators, or those studying event creation mechanics.

## Key Concepts Across Documents

Several concepts appear throughout the documentation and are essential to understanding transaction flow:

- **Transaction Pool**: The in-memory structure holding pending and queued transactions
- **Gas Power**: Validator capacity to include transactions in events, allocated based on stake
- **Event Emission**: The process by which validators create and broadcast events containing transactions
- **Transaction Selection**: The algorithm that determines which transactions are included in each event
- **Nonce**: Sequential transaction counter per account ensuring transaction ordering
- **Gas Price**: Fee per unit of gas, used for transaction prioritization

## Transaction Lifecycle

A transaction on U2U goes through the following stages:

1. **Submission**: User submits transaction via RPC API
2. **Validation**: Transaction pool validates signature, nonce, gas, and state
3. **Queuing**: Valid transaction enters pending or queued pool
4. **Selection**: Validator selects transaction based on price, nonce, and availability
5. **Inclusion**: Transaction added to event payload
6. **Emission**: Validator creates and broadcasts event
7. **Consensus**: Event propagates through DAG and becomes part of consensus
8. **Finalization**: Transaction included in finalized block

## Reading Paths

Depending on your role and goals, different reading paths may be most effective:

**For Developers/Integrators:**
1. Start with `01_transaction_pool.md` for transaction handling basics
2. Read `02_transaction_selection.md` to understand transaction ordering
3. Reference `04_event_emission.md` for event creation details

**For Validators:**
1. Read `03_gas_power_allocation.md` to understand gas power management
2. Focus on `04_event_emission.md` for event creation mechanics
3. Review `02_transaction_selection.md` for transaction selection optimization

**For Researchers:**
1. Read all documents in sequence (01 → 02 → 03 → 04)
2. Pay special attention to `03_gas_power_allocation.md` for economic mechanisms
3. Cross-reference with consensus documentation for complete picture

## Implementation References

Each document references specific code locations in the U2U codebase. Key areas include:

- **Transaction Pool**: `evmcore/tx_pool.go`, `evmcore/tx_list.go`
- **Transaction Selection**: `gossip/emitter/txs.go`, `gossip/emitter/emitter.go`
- **Gas Power**: `eventcheck/gaspowercheck/`, `native/gas_power_left.go`
- **Event Emission**: `gossip/emitter/emitter.go`, `gossip/emitter/control.go`
- **Parent Selection**: `gossip/emitter/parents.go`
- **Network Rules**: `u2u/rules.go`

## Related Documentation

For additional context, see:

- The [consensus documentation](../consensus/) for how events become blocks
- The [staking documentation](../consensus/03_staking_and_validators.md) for validator economics
- The [network parameters](../consensus/04_network_parameters.md) for configuration values

## Summary

U2U's transaction handling system provides:

- **Efficient Pooling**: Fast transaction validation and organization
- **Fair Selection**: Price-based prioritization with conflict prevention
- **Stake-Based Capacity**: Gas power allocation proportional to validator stake
- **Adaptive Emission**: Event creation timing based on network conditions and efficiency

The documentation in this directory provides comprehensive coverage of how transactions flow from user submission through event creation, enabling fast, secure, and efficient transaction processing.



