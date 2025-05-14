# Profiling the SFC Contract's handleSealEpoch Function

This guide explains how to profile the memory and CPU usage of the `handleSealEpoch` function in the SFC contract implementation.

## Prerequisites

- Go installed on your system
- Access to a running U2U node with HTTP-RPC and pprof endpoints enabled

## Quick Start

1. Start your node with profiling enabled:

```bash
# Start a node with profiling (already configured in demo/start.sh)
./demo/start.sh

# Or manually start with these flags:
./build/demo_u2u --pprof --pprof.addr="127.0.0.1" --pprof.port=6060 [other flags]
```

2. Run the profiling script:

```bash
./profile_sealepoch.sh
```

3. Analyze the collected profiles:

```bash
./analyze_profiles.sh
```

## Understanding the Profiles

The scripts will collect and analyze several types of profiles:

1. **CPU Profile**: Shows where the CPU is spending time
2. **Heap Profile**: Shows memory allocations at the time of collection
3. **Allocs Profile**: Shows cumulative memory allocations
4. **Differential Heap Profile**: Shows heap growth between before and after SealEpoch

## Interactive Analysis

For a more detailed analysis using the interactive web UI:

```bash
# Analyze CPU usage
go tool pprof -http=:8080 ./profiles/cpu.pprof

# Analyze memory allocations
go tool pprof -http=:8080 ./profiles/allocs.pprof

# Analyze memory growth during SealEpoch
go tool pprof -http=:8080 -base=./profiles/heap_before.pprof ./profiles/heap_after.pprof
```

## What to Look For

When optimizing the `handleSealEpoch` function, focus on:

1. **Hot spots in CPU profiles**: Functions consuming the most CPU time
2. **Memory allocations**: Large or frequent allocations that could be cached/reused
3. **Memory leaks**: Memory that grows significantly between before/after profiles
4. **Expensive operations**: Hash calculations, large data structure manipulations

## Understanding SFC Memory Patterns

Common memory optimization opportunities in the SFC contract:

1. **Big.Int allocations**: Reuse big.Int objects from a pool rather than creating new ones
2. **Byte slice allocations**: Use a byte slice pool for operations like hashing and encoding
3. **Cached calculations**: Cache frequently accessed values like epoch snapshot slots
4. **Reduced storage operations**: Minimize state reads/writes where possible

## File Structure

- `profile_sealepoch.sh`: Collects profiles before/during/after SealEpoch execution
- `analyze_profiles.sh`: Generates text and SVG reports from the collected profiles
- `profiles/`: Directory where raw profiles are stored
- `profiles/analysis/`: Directory where analysis reports are stored

## Customization

To profile different operations, modify the transaction data in `profile_sealepoch.sh` to call different contract methods. 