# Consensus Documentation Overview

This directory contains comprehensive documentation about U2U's consensus mechanism, which is based on the Helios protocol—an Asynchronous Byzantine Fault Tolerance (aBFT) consensus algorithm combined with Directed Acyclic Graphs (DAGs).

## Documentation Structure

The consensus documentation is organized into four main documents that progressively build understanding from high-level concepts to detailed implementation specifics:

### [01_consensus_overview.md](./01_consensus_overview.md)
**Introduction to U2U Consensus**

This document provides the foundational understanding of U2U's consensus mechanism. It covers:

- **Byzantine Fault Tolerance**: Explains Practical Byzantine Fault Tolerance (PBFT) and how U2U innovates with Asynchronous Byzantine Fault Tolerance (aBFT)
- **Directed Acyclic Graphs**: Introduces DAGs as a data structure and how they enable concurrent event processing
- **Helios Consensus**: Describes how U2U combines proof-of-stake, DAG-based architecture, and aBFT to achieve fast, secure consensus
- **Key Concepts**: Defines fundamental terms including Events, Frames, Roots, Epochs, and Atropos
- **Transaction Lifecycle**: Traces a transaction from submission through finality, explaining each stage of the consensus process

**Best for**: Readers new to U2U consensus or those seeking a high-level understanding of how the system works.

### [02_helios_mechanics.md](./02_helios_mechanics.md)
**Deep Dive into Helios Implementation**

This document delves into the technical mechanics of how Helios transforms a DAG of events into a linear blockchain:

- **Event Structure**: Detailed breakdown of event components, serialization, and encoding
- **Ordering Engine**: Explains the multi-stage process of Root → Clotho → Atropos that orders events
- **Graph Traversal**: How validators validate and traverse the DAG structure
- **Validation Pipeline**: The layered validation system (light, buffered, heavy checks) that ensures event safety
- **Misbehaviour Detection**: How the system detects and handles validator cheating
- **Block Construction**: The process of building final blocks from Atropos events

**Best for**: Developers implementing consensus features, researchers studying the algorithm, or those needing to understand the internal mechanics.

### [03_staking_and_validators.md](./03_staking_and_validators.md)
**Proof-of-Stake and Validator Operations**

This document explains how validators participate in consensus through staking:

- **Validator Lifecycle**: From becoming a validator through active participation to potential deactivation
- **Staking Parameters**: Gas power allocation, epoch rules, and block missed tolerance
- **SFC Integration**: How the Staking and Fee Contract (SFC) interacts with consensus through dual state roots
- **Slashing and Misbehaviour**: Two-layer approach to detecting and penalizing validator misbehaviour
- **Validator Rewards**: How validators earn rewards and how they're calculated during epoch sealing

**Best for**: Validators setting up nodes, stakers delegating tokens, or those interested in the economic incentives of the network.

### [04_network_parameters.md](./04_network_parameters.md)
**Network Configuration and Parameters**

This document catalogs the key parameters that govern network behavior:

- **Time and Block Parameters**: Block timing, epoch duration, and empty block handling
- **Gas and Fee Parameters**: Minimum gas prices, gas limits, and gas power allocation rules
- **Validator Parameters**: Validator count limits, blocks missed tolerance, and performance tracking
- **Network Protocol Parameters**: Message size limits, propagation settings, and DAG stream configuration

**Best for**: Network operators, developers tuning network behavior, or those needing reference values for network parameters.

## Key Concepts Across Documents

Several concepts appear throughout the documentation and are essential to understanding U2U consensus:

- **Events**: The atomic units of consensus, containing transactions and consensus data
- **DAG**: The directed acyclic graph structure that allows concurrent event processing
- **Epochs**: Time-bounded periods with consistent validator sets and economic rules
- **Atropos**: Finalized events that anchor blocks in the main chain
- **aBFT**: Asynchronous Byzantine Fault Tolerance, enabling fast consensus without sequential block production
- **SFC**: The Staking and Fee Contract that manages validator stakes and rewards

## Reading Paths

Depending on your role and goals, different reading paths may be most effective:

**For New Users/Developers:**
1. Start with `01_consensus_overview.md` for foundational concepts
2. Read `03_staking_and_validators.md` to understand validator participation
3. Reference `04_network_parameters.md` as needed for specific values

**For Consensus Developers:**
1. Read `01_consensus_overview.md` for context
2. Deep dive into `02_helios_mechanics.md` for implementation details
3. Reference `04_network_parameters.md` for configuration values
4. Review `03_staking_and_validators.md` to understand validator integration

**For Validators/Operators:**
1. Start with `01_consensus_overview.md` for system understanding
2. Focus on `03_staking_and_validators.md` for operational details
3. Keep `04_network_parameters.md` handy for configuration reference

**For Researchers:**
1. Read all documents in sequence (01 → 02 → 03 → 04)
2. Pay special attention to `02_helios_mechanics.md` for algorithmic details
3. Cross-reference with codebase locations mentioned in each document

## Implementation References

Each document references specific code locations in the U2U codebase. Key areas include:

- **Event Structure**: `native/event.go`, `native/event_serializer.go`
- **Consensus Logic**: `gossip/handler.go`, `gossip/c_block_callbacks.go`
- **State Management**: `native/iblockproc/decided_state.go`
- **Staking Integration**: `gossip/evmstore/store_sfc.go`, `u2u/contracts/sfc/`
- **Network Rules**: `u2u/rules.go`
- **Protocol Implementation**: `gossip/protocol.go`, `gossip/emitter/*`

## Related Documentation

For additional context, see:

- The main [architecture documentation](../) for system-wide design
- The upstream [go-helios](https://github.com/ulrsoft/go-helios) library documentation for DAG engine details
- The SFC contract documentation for staking contract specifics

## Summary

U2U's consensus mechanism represents a significant innovation in blockchain consensus, combining:

- **Speed**: 1-2 second transaction finality through asynchronous processing
- **Security**: Strong finality guarantees through aBFT and proof-of-stake
- **Efficiency**: DAG-based architecture enables concurrent event processing
- **Flexibility**: Configurable parameters allow network evolution

The documentation in this directory provides comprehensive coverage of how these elements work together to create a fast, secure, and efficient consensus system.

