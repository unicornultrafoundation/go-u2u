# Gas Power Allocation System

Gas power is the capacity validators have to include transactions in events. It's allocated based on validator stake and accumulates over time. This document explains how gas power is calculated, allocated, and used.

## Gas Power Overview

Gas power represents a validator's ability to include transactions in events. Unlike traditional blockchains where block producers have fixed gas limits, U2U validators receive gas power allocation that:

- **Accumulates Over Time**: Gas power builds up continuously based on stake
- **Stake-Proportional**: Validators with more stake receive more gas power
- **Time-Based**: Allocation rate is per second, accumulating up to a maximum
- **Dual-Window System**: Uses both long-term and short-term allocation windows

## Allocation Mechanisms

U2U uses a dual-window gas power allocation system with three allocation types:

### Long-Term Allocation

Long-term allocation provides steady, predictable gas power:

- **Allocation Rate**: 100 × DefaultEventGas per second (typically 2,800,000 gas/second)
- **Maximum Accumulation Period**: 60 minutes
- **Maximum Gas Power**: `AllocPerSec × 3600` (accumulated over 60 minutes)
- **Startup Period**: 5 seconds
- **Minimum Startup Gas**: 20 × DefaultEventGas (typically 560,000 gas)

Long-term allocation ensures validators always have baseline capacity to create events.

### Short-Term Allocation

Short-term allocation provides burst capacity for high-demand periods:

- **Allocation Rate**: 2× the long-term rate (typically 5,600,000 gas/second)
- **Maximum Accumulation Period**: Reduced by factor of 12 (5 minutes)
- **Maximum Gas Power**: `AllocPerSec × 300` (accumulated over 5 minutes)
- **Startup Period**: Half of long-term period (2.5 seconds)
- **Minimum Startup Gas**: Same as long-term

Short-term allocation allows validators to handle traffic spikes while preventing long-term accumulation.

### Startup Allocation

Startup allocation ensures new validators can immediately participate:

- **Immediate Availability**: Available from first event creation
- **Minimum Guarantee**: Ensures validators can create at least one event
- **No Waiting Period**: No need to wait for accumulation

This prevents new validators from being unable to participate due to zero gas power.

## Stake-Based Distribution

Gas power allocation is proportional to validator stake:

### Calculation Formula

For a validator with stake `S` and total network stake `T`:

```
GasPowerPerSec = (AllocPerSec × S) / T
MaxGasPower = GasPowerPerSec × MaxAllocPeriod
```

### Proportional Allocation

- **Higher Stake = More Gas Power**: Validators with more stake receive proportionally more gas power
- **Fair Distribution**: All validators receive allocation proportional to their stake
- **No Minimum**: Even validators with small stakes receive allocation (though it may be minimal)

### Example

If a validator has 10% of total stake:
- Long-term allocation: 10% of 2,800,000 = 280,000 gas/second
- Maximum accumulation: 280,000 × 3600 = 1,008,000,000 gas
- Short-term allocation: 10% of 5,600,000 = 560,000 gas/second
- Maximum accumulation: 560,000 × 300 = 168,000,000 gas

## Accumulation Rules

Gas power accumulates continuously over time:

### Time-Based Accumulation

Gas power increases based on time elapsed since last event:

```
NewGasPower = PreviousGasPower + (TimeElapsed × AllocPerSec)
```

Where:
- `TimeElapsed`: Time since last event (or epoch start for first event)
- `AllocPerSec`: Validator's allocation rate (stake-proportional)

### Maximum Limits

Gas power cannot exceed maximum accumulation:

- **Long-Term Maximum**: `AllocPerSec × 3600` (60 minutes)
- **Short-Term Maximum**: `AllocPerSec × 300` (5 minutes)
- **Capped Accumulation**: Gas power stops accumulating at maximum

### Epoch Boundaries

At epoch boundaries:

- **Gas Refunds**: Unused gas power may be refunded (added to next epoch)
- **State Reset**: Gas power state is updated based on epoch sealing
- **Validator Updates**: New validator set may change allocation rates

## Usage Tracking

Gas power is tracked and consumed as transactions are included:

### Gas Power Left

Each event tracks `GasPowerLeft`, which represents:

- **Available Capacity**: Remaining gas power after previous events
- **Dual Values**: Separate values for long-term and short-term allocation
- **Minimum Selection**: The minimum of both values is used for transaction inclusion

### Gas Power Used

Each event tracks `GasPowerUsed`, which represents:

- **Consumed Capacity**: Total gas consumed by transactions in the event
- **Incremental**: Accumulated as transactions are added
- **Event Limit**: Cannot exceed `MaxEventGas` (10,000,028 gas)

### Calculation

When creating an event:

1. **Calculate Available**: `GasPowerLeft = PreviousGasPowerLeft + Accumulated - GasPowerUsed`
2. **Select Transactions**: Include transactions up to `GasPowerLeft.Min()`
3. **Update Used**: Track `GasPowerUsed` as transactions are added
4. **Update Left**: `GasPowerLeft = GasPowerLeft - GasPowerUsed`

## Gas Refunds

Unused gas power may be refunded at epoch boundaries:

### Refund Mechanism

- **Epoch Sealing**: Refunds calculated during epoch sealing
- **Validator State**: Stored in `ValidatorEpochState.GasRefund`
- **Next Epoch**: Refunds added to gas power at start of next epoch

### Refund Calculation

Refunds are calculated based on:

- **Unused Gas Power**: Gas power that wasn't consumed
- **Epoch Rules**: Refund rules defined in epoch sealing logic
- **Validator Performance**: May depend on validator participation

## Validation

Gas power allocation and usage are validated to prevent abuse:

### Allocation Validation

When validating an event:

1. **Calculate Expected**: Compute expected gas power based on time and stake
2. **Compare Actual**: Compare event's `GasPowerLeft` with expected
3. **Check Usage**: Verify `GasPowerLeft + GasPowerUsed = ExpectedGasPower`
4. **Reject if Invalid**: Event is rejected if gas power is miscalculated

### Validation Context

Validation uses:

- **Epoch State**: Current epoch's validator set and rules
- **Previous Event**: Self-parent event's gas power state
- **Time Information**: Event creation time and median time
- **Validator State**: Validator's previous epoch state (for first event)

## Configuration

Gas power allocation is configured in `u2u/rules.go`:

### Long-Term Configuration

```go
LongGasPower: {
    AllocPerSec:        100 * DefaultEventGas,  // 2,800,000 gas/second
    MaxAllocPeriod:     60 * time.Minute,        // 60 minutes
    MinEnsuredAlloc:    MaxEventGas,             // 10,000,028 gas
    StartupAllocPeriod: 5 * time.Second,         // 5 seconds
    MinStartupGas:      20 * DefaultEventGas,    // 560,000 gas
}
```

### Short-Term Configuration

```go
ShortGasPower: {
    AllocPerSec:        2 * LongGasPower.AllocPerSec,  // 5,600,000 gas/second
    MaxAllocPeriod:     LongGasPower.MaxAllocPeriod / 12,  // 5 minutes
    MinEnsuredAlloc:    MaxEventGas,                     // 10,000,028 gas
    StartupAllocPeriod: LongGasPower.StartupAllocPeriod / 2,  // 2.5 seconds
    MinStartupGas:      20 * DefaultEventGas,            // 560,000 gas
}
```

These parameters can be adjusted for different network configurations (testnets, mainnet, etc.).

## Usage Limits

Several limits govern gas power usage:

### Event Limits

- **Maximum Event Gas**: 10,000,028 gas per event
- **Base Event Gas**: 28,000 gas (cost of including an event)
- **Parent Gas**: 2,400 gas per parent reference
- **Extra Data Gas**: 25 gas per byte of extra data

### Network Limits

- **Maximum Block Gas**: 20,500,000 gas (technical hard limit)
- **Pending Gas Tracking**: System tracks pending gas to prevent over-allocation
- **Smooth TPS**: Gas power usage is smoothed to prevent TPS spikes

## Smooth TPS Mechanism

To prevent transaction throughput spikes, gas power usage is smoothed:

### Smoothing Logic

When gas power is low:

- **Threshold-Based**: Different thresholds for different behaviors
- **Reduced Usage**: Gas power usage is reduced when below thresholds
- **Gradual Increase**: Usage increases gradually as gas power recovers

### Thresholds

- **NoTxsThreshold**: Below this, no transactions are included
- **LimitedTpsThreshold**: Below this, TPS is limited
- **EmergencyThreshold**: Below this, emission is restricted

These thresholds ensure stable network performance even when validators have low gas power.

## Summary

Gas power allocation provides:

- **Stake-Based Capacity**: Validators receive gas power proportional to stake
- **Time-Based Accumulation**: Gas power builds up continuously over time
- **Dual-Window System**: Long-term and short-term allocation for flexibility
- **Fair Distribution**: All validators can participate based on their stake
- **Usage Tracking**: Precise tracking of gas power consumption
- **Validation**: Ensures gas power is used correctly

The gas power system ensures that validators have predictable capacity to include transactions while maintaining fairness and preventing resource monopolization.

For implementation details, see `eventcheck/gaspowercheck/`, `native/gas_power_left.go`, and `u2u/rules.go`.



