#!/bin/bash

echo "Setting up bootnode connections..."
echo ""

# Wait for node 1 to be ready
echo "Waiting for Node 1 to start..."
for i in {1..30}; do
    if curl -s -X POST http://localhost:8545 -H "Content-Type: application/json" \
       -d '{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}' > /dev/null 2>&1; then
        echo "✅ Node 1 is ready!"
        break
    fi
    echo "  Waiting... ($i/30)"
    sleep 2
done

# Get Node 1's enode
echo ""
echo "Getting Node 1's enode address..."
ENODE=$(curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' | \
  jq -r '.result.enode')

if [ "$ENODE" = "null" ] || [ -z "$ENODE" ]; then
    echo "❌ Failed to get enode from Node 1"
    exit 1
fi

echo "Node 1 enode: $ENODE"

# Update node scripts with the actual enode
echo ""
echo "Updating bootnode configuration in other node scripts..."
for i in $(seq 2 3); do
    sed -i.bak "s|enode://REPLACE_WITH_NODE1_ENODE@127.0.0.1:30303|$ENODE|g" "start-node$i.sh"
    echo "✅ Updated start-node$i.sh"
done

echo ""
echo "🎉 Bootnode setup complete! You can now start the other nodes:"
for i in $(seq 2 3); do
    HTTP_PORT=$((8544 + $i))
    echo "  ./start-node$i.sh   # Node $i - http://localhost:$HTTP_PORT"
done
