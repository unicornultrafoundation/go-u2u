# SFC Contract Memory and CPU Profiling Tools

This directory contains tools for profiling the memory and CPU usage of the SFC contract's handleSealEpoch function.

## Quick Start

The easiest way to run the complete profiling workflow:

```bash
./benchmarks/run_all.sh
```

This script will:
1. Start a test node with profiling enabled
2. Run the profiling tests on the SealEpoch function
3. Analyze the collected profiles
4. Stop the test node

## Manual Workflow

You can also run each step manually:

### Step 1: Start a node with profiling enabled

```bash
./benchmarks/start_test_node.sh
```

### Step 2: Run the profiling

```bash
./benchmarks/profile_sealepoch.sh
```

### Step 3: Analyze the profiles

```bash
./benchmarks/analyze_profiles.sh
```

### Step 4: Stop the node

```bash
./benchmarks/stop_test_node.sh
```

## Viewing Profile Results

After running the analysis, you can find the results in:
- `profiles/analysis/` - Text and SVG reports
- `profiles/` - Raw profile data

For interactive analysis with a web UI:

```bash
go tool pprof -http=:8080 profiles/cpu.pprof
go tool pprof -http=:8080 profiles/allocs.pprof
go tool pprof -http=:8080 -base=profiles/heap_before.pprof profiles/heap_after.pprof
```

## Troubleshooting

1. **"Cannot access pprof server"**: Ensure the node is running with profiling enabled with the `--pprof` flag.

2. **"unrecognized profile format"**: The profile file is likely empty or corrupted. Check that the pprof endpoint is working correctly.

3. **"debug_traceCall returned an error"**: The RPC call to trigger SealEpoch may have failed. You might need to modify the data payload in `profile_sealepoch.sh` or manually trigger a SealEpoch operation.

## Files

- `start_test_node.sh` - Starts a U2U node with profiling enabled
- `stop_test_node.sh` - Stops the test node
- `profile_sealepoch.sh` - Collects profiles before/during/after SealEpoch execution
- `analyze_profiles.sh` - Analyzes the collected profiles
- `run_all.sh` - Runs the complete workflow 