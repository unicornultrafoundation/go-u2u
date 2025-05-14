#!/bin/bash

# Build the U2U node if it doesn't exist
if [ ! -f "./build/demo_u2u" ]; then
    echo "Building U2U node..."
    go build -o ./build/demo_u2u ./cmd/u2u
    if [ $? -ne 0 ]; then
        echo "Failed to build U2U node."
        exit 1
    fi
fi

# Create a test data directory
TEST_DATADIR="./benchmarks/test_node.datadir"
mkdir -p "$TEST_DATADIR"

# Set up ports
PORT=30303
RPC_PORT=8545
WS_PORT=8546
PPROF_PORT=6060
METRICS_PORT=6061
PROMETHEUS_PORT=9090

echo "Initializing SFC state..."
./build/demo_u2u db dump-sfc --experimental --datadir "${TEST_DATADIR}" --verbosity 4

echo "Starting test node with profiling enabled:"
echo "  - HTTP RPC: http://localhost:$RPC_PORT"
echo "  - WebSocket: ws://localhost:$WS_PORT"
echo "  - pprof: http://localhost:$PPROF_PORT/debug/pprof/"
echo "  - Metrics: http://localhost:$METRICS_PORT/debug/metrics"
echo "  - Prometheus: http://localhost:$PROMETHEUS_PORT"

# Start the node with profiling and metrics enabled
./build/demo_u2u \
    --datadir="${TEST_DATADIR}" \
    --fakenet=1/1 \
    --port=${PORT} \
    --nat=extip:127.0.0.1 \
    --http --http.addr="127.0.0.1" --http.port=${RPC_PORT} --http.corsdomain="*" --http.api="eth,debug,net,admin,web3,personal,txpool,dag,sfc" \
    --ws --ws.addr="127.0.0.1" --ws.port=${WS_PORT} --ws.origins="*" --ws.api="eth,debug,net,admin,web3,personal,txpool,dag" \
    --pprof --pprof.addr="127.0.0.1" --pprof.port=${PPROF_PORT} \
    --metrics --metrics.expensive --metrics.addr="127.0.0.1" --metrics.port=${METRICS_PORT} \
    --verbosity=3 --sfc --monitor --monitor.prometheus.port=${PROMETHEUS_PORT} &

NODE_PID=$!

echo "Node started with PID: $NODE_PID"
echo "Press Ctrl+C to stop the node"
echo "To profile the node, run: ./benchmarks/profile_sealepoch.sh"

# Create a file to store the PID
echo $NODE_PID > benchmarks/test_node.pid

# Trap to kill the node when this script is terminated
trap "echo 'Stopping node...'; kill $NODE_PID; rm -f benchmarks/test_node.pid; echo 'Node stopped.'" INT TERM

# Wait for the node process to finish
wait $NODE_PID 