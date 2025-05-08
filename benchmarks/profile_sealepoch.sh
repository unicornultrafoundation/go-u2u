#!/bin/bash

# Define profiling output directory
PROFILE_DIR="./profiles"
mkdir -p $PROFILE_DIR

# Set the node's HTTP endpoint - adjust port if needed
NODE_HTTP="http://localhost:8545"
PPROF_HTTP="http://localhost:6060"

# Check if the pprof endpoint is accessible
echo "Checking if pprof server is running at $PPROF_HTTP..."
curl -s -f "$PPROF_HTTP/debug/pprof/" > /dev/null
if [ $? -ne 0 ]; then
    echo "ERROR: Cannot access pprof server at $PPROF_HTTP"
    echo "Please ensure the node is running with the --pprof flag enabled."
    echo "Example: ./build/demo_u2u --pprof --pprof.addr=\"127.0.0.1\" --pprof.port=6060 [other flags]"
    exit 1
fi

echo "pprof server is running. Starting profiling..."

# Start collecting heap profile before SealEpoch
echo "Collecting initial heap profile..."
curl -s "$PPROF_HTTP/debug/pprof/heap" > "$PROFILE_DIR/heap_before.pprof"

# Verify that the heap profile is valid
if [ ! -s "$PROFILE_DIR/heap_before.pprof" ]; then
    echo "ERROR: Failed to collect initial heap profile."
    exit 1
fi

# Start CPU profiling
echo "Starting CPU profiling for 30 seconds..."
curl -s "$PPROF_HTTP/debug/pprof/profile?seconds=30" > "$PROFILE_DIR/cpu.pprof" &
CPU_PID=$!

echo "Triggering SealEpoch operation..."

# Use admin.exec to directly call the SFC contract's sealEpoch function
# This is more reliable than debug_traceCall for actually executing the contract function
echo "Calling admin.exec to trigger sealEpoch..."
SEAL_RESULT=$(curl -s -X POST -H "Content-Type: application/json" --data '{
  "jsonrpc":"2.0",
  "method":"admin.exec",
  "params":["SFC.sealEpoch([[], [], [], [], 0])"],
  "id":1
}' "$NODE_HTTP")

# Check if the admin.exec call was successful
if echo "$SEAL_RESULT" | grep -q "error"; then
    echo "WARNING: admin.exec returned an error:"
    echo "$SEAL_RESULT" | grep -o '"error":[^}]*'
    echo "Trying alternative method: direct contract call..."
    
    # Attempt direct contract call as fallback
    # Get the current account to use as sender
    ACCOUNTS_RESULT=$(curl -s -X POST -H "Content-Type: application/json" --data '{
      "jsonrpc":"2.0",
      "method":"eth.accounts",
      "params":[],
      "id":1
    }' "$NODE_HTTP")
    
    ACCOUNT=$(echo "$ACCOUNTS_RESULT" | grep -o '"result":\["[^"]*' | cut -d'"' -f4)
    
    if [ -z "$ACCOUNT" ]; then
        echo "ERROR: Failed to get accounts. Cannot proceed with direct contract call."
    else
        echo "Using account: $ACCOUNT for transaction"
        
        # Send transaction to the SFC contract
        TX_RESULT=$(curl -s -X POST -H "Content-Type: application/json" --data "{
          \"jsonrpc\":\"2.0\",
          \"method\":\"eth.sendTransaction\",
          \"params\":[{
            \"from\": \"$ACCOUNT\",
            \"to\": \"0x0000000000000000000000000000000000000400\",
            \"data\": \"0x592fe0c000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000016000000000000000000000000000000000000000000000000000000000322adc3a00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000001c00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000\"
          }],
          \"id\":1
        }" "$NODE_HTTP")
        
        echo "Transaction result: $TX_RESULT"
    fi
else
    echo "admin.exec completed successfully."
    echo "Result: $SEAL_RESULT"
fi

echo "Waiting for CPU profiling to complete..."
wait $CPU_PID

# Verify CPU profile was collected
if [ ! -s "$PROFILE_DIR/cpu.pprof" ]; then
    echo "WARNING: CPU profile appears to be empty."
fi

# Collect heap profile after SealEpoch
echo "Collecting final heap profile..."
curl -s "$PPROF_HTTP/debug/pprof/heap" > "$PROFILE_DIR/heap_after.pprof"

# Verify heap profile was collected
if [ ! -s "$PROFILE_DIR/heap_after.pprof" ]; then
    echo "WARNING: Final heap profile appears to be empty."
fi

# Collect other useful profiles
echo "Collecting additional profiles..."
curl -s "$PPROF_HTTP/debug/pprof/goroutine" > "$PROFILE_DIR/goroutine.pprof"
curl -s "$PPROF_HTTP/debug/pprof/block" > "$PROFILE_DIR/block.pprof"
curl -s "$PPROF_HTTP/debug/pprof/mutex" > "$PROFILE_DIR/mutex.pprof"
curl -s "$PPROF_HTTP/debug/pprof/allocs" > "$PROFILE_DIR/allocs.pprof"

echo "Profiling completed. Profiles saved to $PROFILE_DIR/"
echo "To analyze the profiles, run:"
echo "  ./benchmarks/analyze_profiles.sh"
echo
echo "Or for manual analysis:"
echo "  go tool pprof -http=:8080 $PROFILE_DIR/cpu.pprof"
echo "  go tool pprof -http=:8080 $PROFILE_DIR/heap_after.pprof"
echo "  go tool pprof -http=:8080 -base=$PROFILE_DIR/heap_before.pprof $PROFILE_DIR/heap_after.pprof" 