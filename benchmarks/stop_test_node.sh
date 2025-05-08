#!/bin/bash

# Check if PID file exists
if [ ! -f "benchmarks/test_node.pid" ]; then
    echo "No test node seems to be running (pid file not found)."
    exit 0
fi

# Read the PID
NODE_PID=$(cat benchmarks/test_node.pid)

# Check if the process is still running
if ! ps -p $NODE_PID > /dev/null; then
    echo "Process with PID $NODE_PID not found. It may have already terminated."
    rm -f benchmarks/test_node.pid
    exit 0
fi

# Kill the process
echo "Stopping node with PID $NODE_PID..."
kill $NODE_PID

# Wait for a graceful shutdown
echo "Waiting for node to terminate..."
for i in {1..10}; do
    if ! ps -p $NODE_PID > /dev/null; then
        echo "Node stopped successfully."
        rm -f benchmarks/test_node.pid
        exit 0
    fi
    sleep 1
done

# If still running, force kill
if ps -p $NODE_PID > /dev/null; then
    echo "Node is still running. Sending SIGKILL..."
    kill -9 $NODE_PID
    sleep 1
    if ! ps -p $NODE_PID > /dev/null; then
        echo "Node terminated with SIGKILL."
    else
        echo "Failed to terminate node. Please manually kill PID $NODE_PID."
    fi
fi

rm -f benchmarks/test_node.pid 