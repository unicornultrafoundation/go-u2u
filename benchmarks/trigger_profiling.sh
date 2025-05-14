#!/bin/bash

# This script starts a test node with profiling enabled, 
# calls the SealEpoch function, and then analyzes the profiles

# Configuration 
NODE_BIN="./build/demo_u2u"
PROFILE_DIR="./profiles"
DATADIR="./benchmarks/test_node.datadir"
mkdir -p $PROFILE_DIR
mkdir -p $DATADIR

# Use different ports to avoid conflicts
PPROF_PORT=6070
HTTP_PORT=8555

echo "Starting node with profiling enabled..."
# Start the node with profiling enabled - use separate datadir to avoid DB conflicts
$NODE_BIN --pprof --pprof.addr="127.0.0.1" --pprof.port=$PPROF_PORT --fakenet=1/1 \
    --datadir="$DATADIR" \
    --allow-insecure-unlock --http --http.addr="0.0.0.0" --http.port=$HTTP_PORT \
    --http.api="eth,debug,admin,web3,personal,net" &
NODE_PID=$!

# Wait for node to start
echo "Waiting for node to start..."
sleep 10  # Give it more time to initialize

# Verify node is running
curl -s http://localhost:$HTTP_PORT -X POST -H "Content-Type: application/json" --data '{"jsonrpc":"2.0","method":"web3_clientVersion","params":[],"id":1}' > /dev/null
if [ $? -ne 0 ]; then
    echo "ERROR: Node is not running. Aborting."
    kill $NODE_PID 2>/dev/null
    exit 1
fi
echo "Node started successfully!"

# Get an account to use for transactions
ACCOUNTS_RESULT=$(curl -s -X POST -H "Content-Type: application/json" --data '{
  "jsonrpc":"2.0",
  "method":"accounts",
  "params":[],
  "id":1
}' "http://localhost:$HTTP_PORT")

echo "Accounts result: $ACCOUNTS_RESULT"

# Try alternative method to get accounts
ALT_ACCOUNTS_RESULT=$(curl -s -X POST -H "Content-Type: application/json" --data '{
  "jsonrpc":"2.0",
  "method":"personal_listAccounts",
  "params":[],
  "id":1
}' "http://localhost:$HTTP_PORT")

echo "Alternative accounts result: $ALT_ACCOUNTS_RESULT"

# Extract account from the results
ACCOUNT=$(echo "$ACCOUNTS_RESULT $ALT_ACCOUNTS_RESULT" | grep -o '"result":\[\s*"[^"]*' | head -1 | sed 's/"result":\[\s*"//' | tr -d '"')

# If no account found, look for validator accounts
if [ -z "$ACCOUNT" ]; then
    # Try to find if there's a known genesis/validator account
    echo "No regular accounts found. Looking for validator accounts..."
    
    # Check for the default validator account that might be available in fakenet mode
    ACCOUNT="0x239fA7623354eC26520dE878B52f13Fe84b06971"
    
    echo "Using validator account: $ACCOUNT"
fi

# If still no account, try to create one
if [ -z "$ACCOUNT" ]; then
    echo "No accounts found. Trying to create a new account..."
    
    # Create a new account with empty password
    NEW_ACCOUNT_RESULT=$(curl -s -X POST -H "Content-Type: application/json" --data '{
      "jsonrpc":"2.0",
      "method":"personal_newAccount",
      "params":[""],
      "id":1
    }' "http://localhost:$HTTP_PORT")
    
    ACCOUNT=$(echo "$NEW_ACCOUNT_RESULT" | grep -o '"result":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$ACCOUNT" ]; then
        echo "ERROR: Failed to create account. Cannot proceed."
        kill $NODE_PID 2>/dev/null
        exit 1
    fi
fi

echo "Using account: $ACCOUNT"

# Unlock account
UNLOCK_RESULT=$(curl -s -X POST -H "Content-Type: application/json" --data "{
  \"jsonrpc\":\"2.0\",
  \"method\":\"personal_unlockAccount\",
  \"params\":[\"$ACCOUNT\", \"\", 300],
  \"id\":1
}" "http://localhost:$HTTP_PORT")

echo "Account unlock result: $UNLOCK_RESULT"

# Collect baseline profiles before SealEpoch
echo "Collecting baseline profiles..."
curl -s "http://localhost:$PPROF_PORT/debug/pprof/heap" > "$PROFILE_DIR/heap_before.pprof"

# Start CPU profiling 
echo "Starting CPU profiling..."
curl -s "http://localhost:$PPROF_PORT/debug/pprof/profile?seconds=60" > "$PROFILE_DIR/cpu.pprof" &
CPU_PID=$!

# Trigger SealEpoch operation
echo "Triggering SealEpoch operation..."

# Try admin.exec first (preferred method)
SEAL_RESULT=$(curl -s -X POST -H "Content-Type: application/json" --data '{
  "jsonrpc":"2.0",
  "method":"admin_exec",
  "params":["SFC.sealEpoch([[], [], [], [], 0])"],
  "id":1
}' "http://localhost:$HTTP_PORT")

if echo "$SEAL_RESULT" | grep -q "error"; then
    # Fall back to direct contract call
    echo "Admin exec failed, falling back to direct transaction..."
    TX_RESULT=$(curl -s -X POST -H "Content-Type: application/json" --data "{
      \"jsonrpc\":\"2.0\",
      \"method\":\"sendTransaction\",
      \"params\":[{
        \"from\": \"$ACCOUNT\",
        \"to\": \"0x0000000000000000000000000000000000000400\", 
        \"data\": \"0x592fe0c000000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000120000000000000000000000000000000000000000000000000000000000000016000000000000000000000000000000000000000000000000000000000322adc3a00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000001c00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000\"
      }],
      \"id\":1
    }" "http://localhost:$HTTP_PORT")
    
    echo "Transaction result: $TX_RESULT"
else
    echo "Admin exec successful."
fi

# Wait for CPU profiling to complete
echo "Waiting for CPU profiling to complete..."
wait $CPU_PID

# Collect heap profile after SealEpoch
echo "Collecting final profiles..."
curl -s "http://localhost:$PPROF_PORT/debug/pprof/heap" > "$PROFILE_DIR/heap_after.pprof"
curl -s "http://localhost:$PPROF_PORT/debug/pprof/goroutine" > "$PROFILE_DIR/goroutine.pprof"
curl -s "http://localhost:$PPROF_PORT/debug/pprof/block" > "$PROFILE_DIR/block.pprof"
curl -s "http://localhost:$PPROF_PORT/debug/pprof/mutex" > "$PROFILE_DIR/mutex.pprof"
curl -s "http://localhost:$PPROF_PORT/debug/pprof/allocs" > "$PROFILE_DIR/allocs.pprof"

# Shutdown the node
echo "Shutting down the node..."
kill $NODE_PID 2>/dev/null

# Wait for node to shut down
sleep 2

# Analyze profiles
echo "Analyzing profiles..."
cd benchmarks
bash ./analyze_profiles.sh

echo "Profiling complete! Results are available in $PROFILE_DIR/analysis/"
echo "To view interactive results, run:"
echo "go tool pprof -http=:8080 $PROFILE_DIR/cpu.pprof" 