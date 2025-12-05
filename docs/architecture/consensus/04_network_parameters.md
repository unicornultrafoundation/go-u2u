# Network Parameters

This document describes the key parameters that govern the U2U network's consensus behavior, block production, gas economics, and validator configuration. These parameters are defined in the codebase and determine how the network operates under various conditions.

## Time and Block Parameters

### Block Time

U2U does not use a fixed block time like some blockchains. Instead, block timing emerges from the consensus process:

- **Atropos Time**: Each block's time is determined by the median time of its Atropos event across validators
- **Minimum Interval**: Blocks are skipped if they're empty and within the maximum skip period (1 second on mainnet/testnet)
- **Target**: Under normal conditions, blocks are produced approximately every 1-2 seconds

This flexible timing allows the network to adapt to transaction load while maintaining consistent block production. During high load, blocks are produced more frequently. During low load, empty blocks may be skipped to reduce unnecessary state updates.

### Epoch Duration

Epochs are time-bounded periods during which a consistent validator set and economic rules apply. Epochs are sealed when either:

- **Maximum Epoch Gas** (300,000,000) is consumed, or
- **Maximum Epoch Duration** (7 minutes) is reached

These limits ensure that:
- Validator sets are regularly updated
- Rewards are distributed in a timely manner
- Network upgrades can be applied at epoch boundaries

The epoch duration parameter (`MaxEpochDuration`) is defined in `u2u/rules.go` and can be adjusted for different network configurations (e.g., testnets may use shorter epochs).

### Empty Block Handling

To avoid creating unnecessary blocks during low transaction periods, the system can skip empty blocks if:

- The block contains no transactions
- The block contains no cheaters
- The time since the last block is less than `MaxEmptyBlockSkipPeriod` (1 second on mainnet/testnet)

This mechanism reduces storage and processing overhead while maintaining the ability to produce blocks quickly when transactions are available.

## Gas and Fee Parameters

### Minimum Gas Price

The network enforces a minimum gas price to prevent spam and ensure transaction fees have meaningful value:

- **Mainnet/Testnet**: 1 Gwei (1e9 wei) per gas unit
- **Enforcement**: Transactions below this price are rejected

The minimum gas price is defined in `u2u/rules.go` as part of `EconomyRules.MinGasPrice`. The gas price oracle can propose higher prices based on network demand, but cannot go below this floor.

### Gas Limits

Several gas limits govern event and block construction:

**Event Gas Limits**:
- **Base Event Gas**: 28,000 gas (cost of including an event)
- **Maximum Event Gas**: 10,000,000 + 28,000 gas (cap on gas per event)
- **Parent Gas**: 2,400 gas per parent reference
- **Extra Data Gas**: 25 gas per byte of extra data

**Block Gas Limits**:
- **Maximum Block Gas**: 20,500,000 gas (technical hard limit)
- **Effective Limit**: Gas power allocation rules typically govern actual block size

These limits ensure that:
- Events cannot consume unbounded resources
- Blocks remain within reasonable size limits
- The network can process blocks efficiently

### Gas Power Allocation

Validators receive gas power allocation that determines their capacity to include transactions:

**Long-Term Allocation**:
- **Allocation Rate**: 100 × DefaultEventGas per second
- **Maximum Accumulation Period**: 60 minutes
- **Startup Period**: 5 seconds
- **Minimum Startup Gas**: 20 × DefaultEventGas

**Short-Term Allocation**:
- **Allocation Rate**: 2× the long-term rate (for burst capacity)
- **Maximum Accumulation Period**: Reduced by factor of 12
- **Startup Period**: Half of long-term period

This dual-window system allows validators to:
- Maintain steady transaction inclusion capacity
- Handle burst traffic when needed
- Start participating immediately without waiting for allocation buildup

## Validator Parameters

### Validator Count

The maximum number of validators is not hardcoded in the consensus layer but is determined by:

- **SFC Contract Rules**: The staking contract defines minimum stake and validator selection criteria
- **Consensus Limits**: The upstream Helios PoS implementation may enforce practical limits
- **Network Performance**: More validators increase security but may impact performance

The actual validator count is reflected in `EpochState.Validators.Len()`, which is updated during epoch sealing based on the current SFC state.

### Blocks Missed Tolerance

The system tracks validator liveness through:

- **Block Missed Slack**: 50 blocks (default tolerance before penalties may apply)
- **Uptime Tracking**: Validators' online time and last activity are recorded
- **Performance Metrics**: Validators' participation in events and blocks

These metrics enable:
- Detection of validator downtime
- Application of downtime penalties during epoch sealing
- Fair reward distribution based on participation

The exact penalty mechanism for missed blocks is determined by the SFC contract and epoch sealing logic, which can implement various policies (reduced rewards, temporary jailing, stake slashing, etc.).

## Network Protocol Parameters

### Message Size Limits

To prevent resource exhaustion and ensure network stability:

- **Maximum Message Size**: 10 MiB (defined in `native/event_serializer.go`)
- **Semaphore Limits**: Control concurrent message processing
  - Messages: 1,000 concurrent, 30 MiB total
  - Events: 10,000 concurrent, 30 MiB total
  - Block Votes: 5,000 concurrent, 15 MiB total

These limits ensure that:
- Individual messages cannot overwhelm validators
- Network resources are fairly distributed
- The system can handle high transaction volumes

### Propagation Settings

The gossip protocol uses several parameters to balance latency and throughput:

- **Progress Broadcast Period**: 10 seconds (how often validators advertise their progress)
- **Initial Transaction Hashes**: Up to 20,000 hashes sent to new peers
- **Random Transaction Hashes**: 128 hashes sent periodically to random peers

These settings ensure that:
- New peers can quickly synchronize
- Transaction availability is efficiently propagated
- Network bandwidth is used effectively

### DAG Stream Configuration

The DAG streaming protocol uses:

- **Chunk Size**: Configurable based on network conditions
- **Parallel Downloads**: Multiple chunks can be downloaded simultaneously
- **Session Timeout**: 4 minutes maximum for stream sessions

These parameters allow validators to:
- Efficiently synchronize the DAG
- Handle network interruptions gracefully
- Balance download speed with resource usage

## Summary

U2U's network parameters are designed to:

1. **Balance Performance and Security**: Parameters ensure fast block production while maintaining strong security guarantees
2. **Prevent Resource Exhaustion**: Limits on message sizes and concurrent processing protect validators
3. **Enable Fair Resource Distribution**: Gas power allocation ensures all validators can participate
4. **Support Network Evolution**: Parameters can be adjusted at epoch boundaries for network upgrades

The following table summarizes key parameters:

| Parameter | Value | Source |
|-----------|-------|--------|
| Max Empty Block Skip Period | 1 second | `u2u/rules.go` |
| Max Block Gas | 20,500,000 | `u2u/rules.go` |
| Max Epoch Gas | 300,000,000 | `u2u/rules.go` |
| Max Epoch Duration | 7 minutes | `u2u/rules.go` |
| Min Gas Price | 1 Gwei | `u2u/rules.go` |
| Base Event Gas | 28,000 | `u2u/rules.go` |
| Max Event Gas | 10,000,028 | `u2u/rules.go` |
| Block Missed Slack | 50 blocks | `u2u/rules.go` |
| Max Message Size | 10 MiB | `native/event_serializer.go` |
| Progress Broadcast Period | 10 seconds | `gossip/config.go` |

These parameters work together to create a network that is fast, secure, and efficient while maintaining flexibility for future improvements and upgrades.

