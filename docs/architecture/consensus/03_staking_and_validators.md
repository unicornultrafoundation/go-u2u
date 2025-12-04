# Staking and Validators

U2U uses a proof-of-stake consensus mechanism where validators stake tokens to participate in block production and transaction validation. This document explains how validators operate within the consensus system, how staking is managed, and how misbehaviour is detected and penalized.

## Validator Lifecycle

Validators in the U2U network go through several stages during their lifecycle, from initial setup to active participation and potential deactivation.

### Becoming a Validator

To become a validator, a participant must:

1. **Stake tokens** through the SFC (Staking and Fee Contract), which is predeployed at address `0xfc00face00000000000000000000000000000000`
2. **Meet minimum stake requirements** as defined in the SFC contract
3. **Register their validator node** with the network

The SFC contract manages all staking operations, including:
- Validator creation and registration
- Delegation from token holders
- Stake withdrawal and unbonding
- Reward distribution

The consensus layer interacts with the SFC through a separate state root (`SfcStateRoot`) that tracks staking state independently from the main EVM state. This separation allows the consensus engine to efficiently query validator sets and staking information without scanning the entire EVM state.

### Active Participation

Once active, validators participate in consensus by:

1. **Creating event blocks**: Validators batch transactions into events and add them to their local DAG
2. **Validating events**: Validators validate events from other validators before incorporating them
3. **Voting on blocks and epochs**: Validators vote on block and epoch records through the LLR layer
4. **Maintaining uptime**: Validators must stay online to participate in consensus

The consensus system tracks validator performance through several metrics stored in `ValidatorBlockState` and `ValidatorEpochState` (defined in `native/iblockproc/decided_state.go`):

- **Last Event**: The most recent event created by the validator
- **Uptime**: Total time the validator has been online
- **Last Online Time**: When the validator was last considered online
- **Gas Power Left**: Remaining gas allocation for creating events
- **Last Block**: The most recent block in which the validator participated

### Validator State

Validator state is maintained at two levels:

**Block State** (`ValidatorBlockState`):
- Tracks validator activity within the current block
- Maintains gas power allocation and usage
- Records the last event and block participation

**Epoch State** (`ValidatorEpochState`):
- Tracks validator activity across an entire epoch
- Maintains gas refunds and epoch-level metrics
- Records the previous epoch's last event

These states are updated during block processing and epoch sealing, ensuring that validator performance is accurately tracked for reward distribution and slashing calculations.

## Staking Parameters

The consensus system enforces several parameters that affect validator economics and behavior:

### Gas Power Allocation

Validators receive gas power allocation that determines how many transactions they can include in their events. Gas power is allocated according to rules defined in `u2u/rules.go`:

- **Long-term allocation**: Provides steady gas power over a 60-minute window
- **Short-term allocation**: Provides faster allocation for burst capacity
- **Startup allocation**: Ensures new validators can immediately start creating events

These rules prevent validators from monopolizing transaction inclusion while ensuring fair resource distribution.

### Epoch Rules

Epochs have two main constraints:

- **Maximum Epoch Gas**: The total gas that can be consumed in an epoch (default: 300,000,000)
- **Maximum Epoch Duration**: The maximum time an epoch can last (default: 7 minutes)

When either limit is reached, the epoch is sealed, validator sets are updated, and rewards are distributed.

### Block Missed Tolerance

The system tracks missed blocks through `BlockMissedSlack` (default: 50 blocks). Validators that miss more than this threshold may face penalties during epoch sealing, though the exact penalty mechanism is determined by the SFC contract and epoch sealing logic.

## SFC and Consensus Interaction

The SFC (Staking and Fee Contract) is a critical component that bridges on-chain staking with consensus operations. The consensus layer maintains a separate state root for SFC data, allowing efficient access to staking information without scanning the entire EVM state.

### SFC State Root

Each block includes two state roots:

- **EVM State Root**: The root of the main EVM state trie
- **SFC State Root**: The root of the SFC staking state trie

This dual-root approach allows the consensus engine to:

- Quickly query validator sets and staking information
- Update staking state independently from EVM state
- Maintain efficient state synchronization

The SFC state root is stored and indexed in `gossip/evmstore/store_sfc.go`, allowing historical queries of staking state at any block.

### Epoch Sealing and Validator Updates

At epoch boundaries, the system:

1. Reads the current SFC state to determine the active validator set
2. Updates validator profiles with current staking information
3. Distributes rewards based on validator performance
4. Applies penalties for misbehaviour or downtime
5. Updates the consensus validator set for the new epoch

This process ensures that the consensus validator set always reflects the current on-chain staking state, maintaining alignment between staking and consensus.

## Slashing and Misbehaviour

The system detects and penalizes validator misbehaviour through a two-layer approach:

### Detection Layer

Misbehaviour is detected through proofs embedded in event payloads. These proofs can identify:

- **Event Double-Signing**: A validator creating two different events at the same sequence position
- **Block Vote Double-Signing**: A validator voting for two different block records
- **Epoch Vote Double-Signing**: A validator voting for two different epoch records
- **Wrong Votes**: Votes that contradict the canonical block or epoch record

When misbehaviour is detected:

1. The proof is extracted from the event payload
2. The cheating validator is identified
3. The validator is added to the epoch's cheater list (`EpochCheaters`)
4. The cheater list is persisted in block and epoch state

### Enforcement Layer

Economic penalties are applied during epoch sealing:

1. The SFC contract is queried for slashing rules
2. Validator stakes are reduced according to the severity of misbehaviour
3. Rewards are adjusted or withheld
4. Validators may be temporarily jailed or permanently removed

The exact penalty mechanism depends on the SFC contract implementation, which can be customized for different misbehaviour types and severities.

### Downtime Penalties

While explicit downtime slashing rules are not hardcoded in the consensus layer, the system tracks:

- Validator uptime and last online time
- Blocks missed by each validator
- Periods of inactivity

This information is available to the SFC contract and epoch sealing logic, which can apply penalties for excessive downtime or missed blocks. The `BlockMissedSlack` parameter (default: 50) defines the tolerance threshold before penalties may apply.

## Validator Rewards

Validators earn rewards for:

- Creating events and including transactions
- Validating events from other validators
- Participating in block and epoch voting
- Maintaining high uptime

Rewards are calculated and distributed during epoch sealing based on:

- Validator stake and delegation
- Gas refunds from transaction fees
- Performance metrics (uptime, events created, etc.)
- Penalties for misbehaviour or downtime

The exact reward calculation is handled by the SFC contract and epoch sealing logic, which can implement various reward distribution schemes.

## Summary

U2U's proof-of-stake system integrates staking with consensus through:

1. **SFC Contract**: Manages all staking operations on-chain
2. **Consensus State**: Tracks validator performance and participation
3. **Dual State Roots**: Separates EVM and staking state for efficiency
4. **Misbehaviour Detection**: Identifies cheating through event-embedded proofs
5. **Economic Enforcement**: Applies penalties through SFC and epoch sealing

This design ensures that validators are properly incentivized to act honestly while providing mechanisms to detect and penalize misbehaviour. The separation between detection (in consensus) and enforcement (in SFC) allows for flexible penalty mechanisms while maintaining strong security guarantees.

For implementation details, see `native/iblockproc/decided_state.go`, `gossip/evmstore/store_sfc.go`, `u2u/contracts/sfc/`, and the SFC contract sources in `gossip/contract/sfc100/`.

