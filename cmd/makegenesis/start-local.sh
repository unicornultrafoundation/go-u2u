#!/bin/bash

# Quick start script for local U2U network
# Usage: ./start-local.sh [number_of_validators]

set -e

# Configuration
NUM_VALIDATORS=${1:-3}
NETWORK_NAME="Local U2U Network"
NETWORK_ID=4439
BALANCE="1000000000000000000000000"  # 1M tokens
STAKE="1000000000000000000"          # 1 token
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BASE_DIR="$PROJECT_ROOT/u2u-local"
GENESIS_FILE="local-genesis.g"
U2U_BINARY="$PROJECT_ROOT/build/u2u"
MAKEGENESIS_BINARY="$PROJECT_ROOT/build/makegenesis"
SETUP_VALIDATOR_BINARY="$PROJECT_ROOT/build/setup_validator_node"

echo "=================================================="
echo "Starting Local U2U Network"
echo "=================================================="
echo "Validators: $NUM_VALIDATORS"
echo "Network ID: $NETWORK_ID"
echo "Data Dir:   $BASE_DIR"
echo "=================================================="
echo ""

# Build binaries if they don't exist
if [ ! -f "$U2U_BINARY" ] || [ ! -f "$MAKEGENESIS_BINARY" ] || [ ! -f "$SETUP_VALIDATOR_BINARY" ]; then
    echo "❌ Required binaries not found. Building all tools..."
    cd "$PROJECT_ROOT"
    make all
    echo "✅ All binaries built"
fi

# Ensure we're in the makegenesis directory
cd "$PROJECT_ROOT/cmd/makegenesis"
if [ ! -f "main.go" ]; then
    echo "❌ main.go not found in cmd/makegenesis directory."
    exit 1
fi

# Generate genesis file
echo ""
echo "📝 Generating genesis file..."
"$MAKEGENESIS_BINARY" \
    -output "$GENESIS_FILE" \
    -validators "$NUM_VALIDATORS" \
    -network "$NETWORK_NAME" \
    -networkid "$NETWORK_ID" \
    -balance "$BALANCE" \
    -stake "$STAKE"
echo "✅ Genesis file created: $GENESIS_FILE"

# Get validator keys
echo ""
echo "🔑 Validator keys:"
if [ -d "tools" ] && [ -f "tools/get_validator_keys.go" ]; then
    cd tools
    go run get_validator_keys.go -validators "$NUM_VALIDATORS"
    cd ..
else
    echo "⚠️  Validator key tool not found (optional)"
    echo "   Validator keys are deterministic - see QUICKSTART.md"
fi

# Create data directories and setup keystores
echo ""
echo "📁 Creating data directories and keystores..."
rm -rf "$BASE_DIR"
for i in $(seq 1 $NUM_VALIDATORS); do
    mkdir -p "$BASE_DIR/node$i"
    echo "  Created: $BASE_DIR/node$i"

    # Setup keystore for this validator
    echo "  Setting up keystore for validator $i..."
    "$SETUP_VALIDATOR_BINARY" $i "./u2u-local/node$i" > /dev/null 2>&1
    echo "  ✅ Keystore setup complete for validator $i"
done

# Create password files
touch "$BASE_DIR/password.txt"
touch "$BASE_DIR/validator_password.txt"
echo "✅ Data directories and keystores created"

# Create start scripts for each node
echo ""
echo "📜 Creating start scripts..."
for i in $(seq 1 $NUM_VALIDATORS); do
    PORT=$((30302 + $i))
    HTTP_PORT=$((8544 + $i))
    WS_PORT=$((8546 + $i))

    # Get validator info
    case $i in
        1)
            VALIDATOR_ADDR="0x239fA7623354eC26520dE878B52f13Fe84b06971"
            VALIDATOR_PUBKEY="0xc0048d505c351f4837cec72bce6f4254f5e4bc3f2c9a4816841db64319eee8b714ef9173fbf66d039b782624713791840846b2788d4b65a425adeba85a4b57efe0cd"
            ;;
        2)
            VALIDATOR_ADDR="0x02AFf1D0a9ed566E644f06FcFE7eFe00A3261D03"
            VALIDATOR_PUBKEY="0xc0043b4060fe18b3ae3a639e7e7b65a1ad01fb236a0dcf4ff4c8d7dd7e3ed4c4ef7a8c52e690a864ca953802f6f5b8e2e37adcfe97e1b740111a6ca782fc54efef11"
            ;;
        3)
            VALIDATOR_ADDR="0x83e573ad09147fc15dac762653a8EDaC9B2516d6"
            VALIDATOR_PUBKEY="0xc0045a463b88e6df3edad80dd667b80dcd9d4685706cbcc5879e3cfbfe27ebab3318b0ace95f2d7ae943748d4c6aa7970882a77d0e044196ac777f7a5202582778d2"
            ;;
        4)
            VALIDATOR_ADDR="0xFcF06fbf5505dF52E28fC907A0ec531E3bA06d18"
            VALIDATOR_PUBKEY="0xc0046636a452064bb2eaea645a705645b44646e6b3d5f8496821254952b98ea388c973290807ca2cedd8a39eea464678d1a0c5f5ff9bb2cfadee3b5a6131ef92cdea"
            ;;
        5)
            VALIDATOR_ADDR="0x0e1341A86EC53BefB038184ed7fa593A1b0bCE03"
            VALIDATOR_PUBKEY="0xc004719fd50e8b4efaab4e3b18f5a38bf635cdd5cb09434f0a319587fafc5774ce3a5762b0ded0c0525064dd07c0b8a9bb0102b3de66340044ffed2abd287b615849"
            ;;
    esac

    # Create bootnode configuration for nodes 2 and above
    if [ $i -eq 1 ]; then
        BOOTNODE_CONFIG=""
    else
        BOOTNODE_CONFIG="  --bootnodes \"enode://REPLACE_WITH_NODE1_ENODE@127.0.0.1:30303\" \\"
    fi

    cat > "$BASE_DIR/start-node$i.sh" << EOF
#!/bin/bash

echo "Starting Node $i..."
echo "  HTTP:  http://localhost:$HTTP_PORT"
echo "  WS:    ws://localhost:$WS_PORT"
echo "  P2P:   localhost:$PORT"
echo ""

"$U2U_BINARY" \\
  --datadir "$BASE_DIR/node$i" \\
  --genesis "$PROJECT_ROOT/cmd/makegenesis/$GENESIS_FILE" \\
  --genesis.allowExperimental \\
  --validator.id $i \\
  --validator.pubkey "$VALIDATOR_PUBKEY" \\
  --validator.password "$BASE_DIR/validator_password.txt" \\
  --unlock "$VALIDATOR_ADDR" \\
  --password "$BASE_DIR/password.txt" \\
  --allow-insecure-unlock \\
  --port $PORT \\
  --http \\
  --http.addr "127.0.0.1" \\
  --http.port $HTTP_PORT \\
  --http.api "eth,debug,net,admin,web3,personal,txpool,dag" \\
  --http.corsdomain "*" \\
  --ws \\
  --ws.addr "127.0.0.1" \\
  --ws.port $WS_PORT \\
  --ws.api "eth,web3,net,admin" \\
  --ws.origins "*" \\
  --verbosity 3 \\
  $BOOTNODE_CONFIG
EOF

    chmod +x "$BASE_DIR/start-node$i.sh"
    echo "  Created: $BASE_DIR/start-node$i.sh"
done

# Create stop script
cat > "$BASE_DIR/stop-all.sh" << 'EOF'
#!/bin/bash
echo "Stopping all U2U nodes..."
pkill -f "u2u.*--datadir"
echo "✅ All nodes stopped"
EOF
chmod +x "$BASE_DIR/stop-all.sh"
echo "  Created: $BASE_DIR/stop-all.sh"

# Create bootnode setup script
cat > "$BASE_DIR/setup-bootnodes.sh" << 'EOF'
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
for i in {2..3}; do
    sed -i.bak "s|enode://REPLACE_WITH_NODE1_ENODE@127.0.0.1:30303|$ENODE|g" "start-node$i.sh"
    echo "✅ Updated start-node$i.sh"
done

echo ""
echo "🎉 Bootnode setup complete! You can now start the other nodes:"
for i in {2..3}; do
    HTTP_PORT=$((8544 + $i))
    echo "  ./start-node$i.sh   # Node $i - http://localhost:$HTTP_PORT"
done
EOF
chmod +x "$BASE_DIR/setup-bootnodes.sh"
echo "  Created: $BASE_DIR/setup-bootnodes.sh"

# Create test script
cat > "$BASE_DIR/test-network.sh" << 'EOF'
#!/bin/bash
echo "Testing U2U Network..."
echo ""

echo "1. Checking Node 1 (port 8545)..."
curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' | jq .

echo ""
echo "2. Checking Network Version..."
curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}' | jq .

echo ""
echo "3. Checking Node 1 Peer Count..."
curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' | jq .

echo ""
echo "4. Checking Node 2 Peer Count..."
curl -s -X POST http://localhost:8546 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' | jq . 2>/dev/null || echo "Node 2 not running"

echo ""
echo "5. Checking Node 3 Peer Count..."
curl -s -X POST http://localhost:8547 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}' | jq . 2>/dev/null || echo "Node 3 not running"
EOF
chmod +x "$BASE_DIR/test-network.sh"
echo "  Created: $BASE_DIR/test-network.sh"

echo "✅ Scripts created"

# Print instructions
echo ""
echo "=================================================="
echo "✅ Local Network Setup Complete!"
echo "=================================================="
echo ""
echo "📋 Next Steps:"
echo ""
echo "1️⃣  Start Node 1 (bootstrap node):"
echo "   $BASE_DIR/start-node1.sh"
echo ""
echo "2️⃣  Setup bootnode connections (after Node 1 starts):"
echo "   cd $BASE_DIR && ./setup-bootnodes.sh"
echo ""
echo "3️⃣  In new terminals, start other nodes:"
for i in $(seq 2 $NUM_VALIDATORS); do
    echo "   $BASE_DIR/start-node$i.sh"
done
echo ""
echo "4️⃣  Test the network:"
echo "   $BASE_DIR/test-network.sh"
echo ""
echo "5️⃣  Stop all nodes:"
echo "   $BASE_DIR/stop-all.sh"
echo ""
echo "🌐 RPC Endpoints:"
for i in $(seq 1 $NUM_VALIDATORS); do
    HTTP_PORT=$((8544 + $i))
    echo "   Node $i: http://localhost:$HTTP_PORT"
done
echo ""
echo "📝 Genesis: $(pwd)/$GENESIS_FILE"
echo "📁 Data:    $BASE_DIR"
echo ""
echo "=================================================="
