# U2U Local Network Setup Guide

A complete step-by-step guide to set up a local U2U blockchain network with custom genesis and validator nodes.

---

## 📋 **Prerequisites**

- Go 1.21+ installed
- U2U binary built: `make u2u` (creates `./build/u2u`)
- Basic understanding of blockchain concepts

---

## 🎯 **Overview**

This guide covers 4 main steps:
1. **Genesis File Generation** - Using existing makegenesis tool
2. **Key Generation** - Generate validator keys and addresses
3. **Keystore Setup** - Configure validator and account keystores
4. **Node Startup** - Launch validator nodes

---

# STEP 1: Genesis File Generation

## 1.1 Using the Existing Genesis Generator

The project includes a built-in genesis generator at `cmd/makegenesis/main.go` with these CLI flags:

```bash
# Available options:
-output string      Output genesis file path (default "genesis.g")
-validators int     Number of validators (default 3)
-network string     Network name (default "testnet")
-networkid uint     Network ID (default fake network ID)
-balance string     Initial balance per validator in wei (default "1000000000000000000000000")
-stake string       Initial stake per validator in wei (default "1000000000000000000")
-epoch uint         Starting epoch number (default 2)
-block uint         Starting block number (default 1)
```

## 1.2 Generate Genesis File

```bash
cd cmd/makegenesis

# Generate with default settings (3 validators)
go run main.go -output local-genesis.g -network "Local U2U Network" -networkid 4439

# Or with custom settings
go run main.go \
  -output local-genesis.g \
  -validators 3 \
  -network "Local U2U Network" \
  -networkid 4439 \
  -balance "1000000000000000000000000" \
  -stake "1000000000000000000"
```

**Expected Output**:
```
INFO Creating genesis file validators=3 network="Local U2U Network" networkid=4439
INFO Genesis file created successfully file=local-genesis.g

=== Validator Information ===
Validator 1:
  Address: 0x239fA7623354eC26520dE878B52f13Fe84b06971
  PubKey:  0xc0048d505c351f4837cec72bce6f4254f5e4bc3f2c9a4816841db64319eee8b714ef9173fbf66d039b782624713791840846b2788d4b65a425adeba85a4b57efe0cd
  Balance: 1000000000000000000000000 wei
  Stake:   1000000000000000000 wei

Validator 2:
  Address: 0x02AFf1D0a9ed566E644f06FcFE7eFe00A3261D03
  PubKey:  0xc0043b4060fe18b3ae3a639e7e7b65a1ad01fb236a0dcf4ff4c8d7dd7e3ed4c4ef7a8c52e690a864ca953802f6f5b8e2e37adcfe97e1b740111a6ca782fc54efef11
  Balance: 1000000000000000000000000 wei
  Stake:   1000000000000000000 wei

Validator 3:
  Address: 0x83e573ad09147fc15dac762653a8EDaC9B2516d6
  PubKey:  0xc0045a463b88e6df3edad80dd667b80dcd9d4685706cbcc5879e3cfbfe27ebab3318b0ace95f2d7ae943748d4c6aa7970882a77d0e044196ac777f7a5202582778d2
  Balance: 1000000000000000000000000 wei
  Stake:   1000000000000000000 wei
```

**File Details**:
- Size: ~45KB
- Format: U2U fileshash (.g) format with compressed sections
- Sections: epochs (ers-1), blocks (brs-1), EVM state (evm-1)
- Network ID: 4439, starting at epoch 2, block 1

---

# STEP 2: Key Generation (Already Built-In!)

## 2.1 Keys Are Generated Automatically

**Important**: The `makegenesis` tool automatically displays validator information! You don't need a separate key generation step.

From the output above, note these **critical details**:

### Validator Public Key Format for U2U
The makegenesis tool now shows U2U format public keys with the **0xc0 type prefix**:

| Validator | U2U Format (from makegenesis & for startup scripts) |
|-----------|-----------------------------------------------------|
| 1 | `0xc0048d505c351f4837cec72bce6f4254f5e4bc3f2c9a4816841db64319eee8b714ef9173fbf66d039b782624713791840846b2788d4b65a425adeba85a4b57efe0cd` |
| 2 | `0xc0043b4060fe18b3ae3a639e7e7b65a1ad01fb236a0dcf4ff4c8d7dd7e3ed4c4ef7a8c52e690a864ca953802f6f5b8e2e37adcfe97e1b740111a6ca782fc54efef11` |
| 3 | `0xc0045a463b88e6df3edad80dd667b80dcd9d4685706cbcc5879e3cfbfe27ebab3318b0ace95f2d7ae943748d4c6aa7970882a77d0e044196ac777f7a5202582778d2` |

## 2.2 Optional: Standalone Key Display Tool

**Skip this step** - the makegenesis tool already provides all validator information needed.

---

# STEP 3: Keystore Setup

## 3.1 Use Existing Keystore Setup Tool

The project already includes `cmd/setup_validator_node/main.go` - no need to create it.

## 3.2 Setup Keystores for All Validators

```bash
# Create data directories
mkdir -p u2u-local/node1 u2u-local/node2 u2u-local/node3

# Setup keystores using existing tool
go run cmd/setup_validator_node/main.go 1 ./u2u-local/node1
go run cmd/setup_validator_node/main.go 2 ./u2u-local/node2
go run cmd/setup_validator_node/main.go 3 ./u2u-local/node3

# Create empty password files
touch u2u-local/password.txt
touch u2u-local/validator_password.txt
```

**Expected Output**:
```
Setting up validator node 1 in ./u2u-local/node1
Address: 0x239fA7623354eC26520dE878B52f13Fe84b06971
PubKey: 0x048d505c351f4837cec72bce6f4254f5e4bc3f2c9a4816841db64319eee8b714ef9173fbf66d039b782624713791840846b2788d4b65a425adeba85a4b57efe0cd
U2U PubKey: 0xc0048d505c351f4837cec72bce6f4254f5e4bc3f2c9a4816841db64319eee8b714ef9173fbf66d039b782624713791840846b2788d4b65a425adeba85a4b57efe0cd
✅ Imported to account keystore: 0x239fA7623354eC26520dE878B52f13Fe84b06971
✅ Added to validator keystore
✅ Regular keystore has address 0x239fA7623354eC26520dE878B52f13Fe84b06971
✅ Validator keystore has pubkey 0x048d505c351f4837cec72bce6f4254f5e4bc3f2c9a4816841db64319eee8b714ef9173fbf66d039b782624713791840846b2788d4b65a425adeba85a4b57efe0cd

🎯 Validator node 1 setup complete!
Use these flags:
  --validator.id 1
  --validator.pubkey 0xc0048d505c351f4837cec72bce6f4254f5e4bc3f2c9a4816841db64319eee8b714ef9173fbf66d039b782624713791840846b2788d4b65a425adeba85a4b57efe0cd
  --validator.password /path/to/empty_password_file.txt
  --unlock 0x239fA7623354eC26520dE878B52f13Fe84b06971
  --password /path/to/empty_password_file.txt
```

**Final Directory Structure**:
```
u2u-local/
├── node1/, node2/, node3/    # Validator data directories with keystores
├── password.txt              # Empty file for account unlock
├── validator_password.txt    # Empty file for validator unlock
└── (startup scripts - see next step)
```

---

# STEP 4: Node Startup Scripts

## 4.1 Generate Startup Scripts

The individual node startup scripts are created automatically using the start-local.sh script.

**From project root directory**:
```bash
./cmd/makegenesis/start-local.sh
```

**OR from the makegenesis directory**:
```bash
cd cmd/makegenesis
./start-local.sh
```

This will create:
- ✅ `u2u-local/start-node1.sh`, `u2u-local/start-node2.sh`, `u2u-local/start-node3.sh`
- ✅ `u2u-local/stop-all.sh`, `u2u-local/test-network.sh`
- ✅ All scripts contain the correct U2U format pubkeys with 0xc0 prefix

**Important**: The script automatically detects the project root and builds all paths correctly from either location.

## 4.2 Key Configuration Details

**Critical Configuration Points**:

1. **Validator PubKey Format**: `--validator.pubkey "0xc0048d505c..."`
   - ✅ **Includes 0xc0 prefix** (validator type marker)
   - ❌ NOT just `0x048d505c...` (raw pubkey)

2. **Dual Authentication**:
   - `--validator.password` + `--validator.pubkey` (for consensus)
   - `--unlock` + `--password` (for transactions)
   - `--allow-insecure-unlock` (for HTTP access)

3. **Genesis File**: `--genesis ./cmd/makegenesis/local-genesis.g --genesis.allowExperimental`

4. **Network Ports**:
   - Node 1: HTTP 8545, WS 8547, P2P 30303
   - Node 2: HTTP 8546, WS 8548, P2P 30304
   - Node 3: HTTP 8547, WS 8549, P2P 30305

5. **Available API Modules**:
   - **HTTP**: `eth,trace,debug,net,admin,web3,personal,txpool,dag,abft,sfc`
   - **WebSocket**: `eth,web3,net,sfc`
   - **Note**: `u2u` module is not available in this version

---

# STEP 5: Launch Your Network!

## 5.1 Start All Validators

**Terminal 1 - Bootstrap Node**:
```bash
./build/u2u \
  --datadir ./u2u-local/node1 \
  --genesis ./cmd/makegenesis/local-genesis.g \
  --genesis.allowExperimental \
  --validator.id 1 \
  --validator.pubkey "0xc0048d505c351f4837cec72bce6f4254f5e4bc3f2c9a4816841db64319eee8b714ef9173fbf66d039b782624713791840846b2788d4b65a425adeba85a4b57efe0cd" \
  --validator.password ./u2u-local/validator_password.txt \
  --unlock "0x239fA7623354eC26520dE878B52f13Fe84b06971" \
  --password ./u2u-local/password.txt \
  --allow-insecure-unlock \
  --verbosity 5 \
  --http \
  --http.addr="0.0.0.0" \
  --http.port=8545 \
  --http.corsdomain="*" \
  --http.vhosts=* \
  --http.api="eth,trace,debug,net,admin,web3,personal,txpool,dag" \
  --ws \
  --ws.addr "0.0.0.0" \
  --ws.port=8546 \
  --ws.api "eth,debug,net,web3,txpool,ftm,dag" \
  --ws.origins "*" \
  --enabletxtracer \
  --gcmode archive \
  --port 30303 \
  --txpool.nolocals
```

**Get Node 1's Enode** (wait 10 seconds after starting Node 1):
```bash
curl -s -X POST http://localhost:8545 -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' | jq -r '.result.enode'
# Example output: enode://b9734d8aafcb1f94425215646a100b88630907240644f9a9febb3ac2a953ecf18d681eb539346d9775db8170266fe7f20c0f9789418b7f164cbad716b962082a@127.0.0.1:30303
```

**Terminal 2 - Second Validator** (replace ENODE_FROM_NODE1 with actual enode):
```bash
./build/u2u \
  --datadir ./u2u-local/node2 \
  --genesis ./cmd/makegenesis/local-genesis.g \
  --genesis.allowExperimental \
  --validator.id 2 \
  --validator.pubkey "0xc0043b4060fe18b3ae3a639e7e7b65a1ad01fb236a0dcf4ff4c8d7dd7e3ed4c4ef7a8c52e690a864ca953802f6f5b8e2e37adcfe97e1b740111a6ca782fc54efef11" \
  --validator.password ./u2u-local/validator_password.txt \
  --unlock "0x02AFf1D0a9ed566E644f06FcFE7eFe00A3261D03" \
  --password ./u2u-local/password.txt \
  --allow-insecure-unlock \
  --bootnodes "ENODE_FROM_NODE1" \
  --verbosity 5 \
  --http \
  --http.addr="0.0.0.0" \
  --http.port=8547 \
  --http.corsdomain="*" \
  --http.vhosts=* \
  --http.api="eth,trace,debug,net,admin,web3,personal,txpool,dag" \
  --ws \
  --ws.addr "0.0.0.0" \
  --ws.port=8548 \
  --ws.api "eth,debug,net,web3,txpool,ftm,dag" \
  --ws.origins "*" \
  --enabletxtracer \
  --gcmode archive \
  --port 30304 \
  --txpool.nolocals
```

**Terminal 3 - Third Validator** (use same enode as bootnode):
```bash
./build/u2u \
  --datadir ./u2u-local/node3 \
  --genesis ./cmd/makegenesis/local-genesis.g \
  --genesis.allowExperimental \
  --validator.id 3 \
  --validator.pubkey "0xc0045a463b88e6df3edad80dd667b80dcd9d4685706cbcc5879e3cfbfe27ebab3318b0ace95f2d7ae943748d4c6aa7970882a77d0e044196ac777f7a5202582778d2" \
  --validator.password ./u2u-local/validator_password.txt \
  --unlock "0x83e573ad09147fc15dac762653a8EDaC9B2516d6" \
  --password ./u2u-local/password.txt \
  --allow-insecure-unlock \
  --port 30305 \
  --bootnodes "ENODE_FROM_NODE1" \
  --http --http.addr "127.0.0.1" --http.port 8547 \
  --http.api "eth,trace,debug,net,admin,web3,personal,txpool,dag,abft,sfc" \
  --ws --ws.addr "127.0.0.1" --ws.port 8549 \
  --ws.api "eth,web3,net,sfc" --ws.origins "*" \
  --verbosity 3
```

**Success Indicators**:
- ✅ `INFO Unlocked validator key pubkey=0xc0048d...`
- ✅ `INFO HTTP server started endpoint=127.0.0.1:8545`
- ✅ `INFO Applied genesis state name="Local U2U Network" id=4439`
- ✅ **Nodes Connected**: Check peer count with `curl -s -X POST http://localhost:8545 -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}'`
- ✅ **Consensus Starting**: `INFO Emitting is paused reason="recently connected"` (will start emitting after wait period)

## 5.2 Test Your Network

**API Test**:
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}'
# Expected: {"jsonrpc":"2.0","id":1,"result":"0x1"}
```

**Network Connectivity Test**:
```bash
# Check Node 1 peer count
curl -s -X POST http://localhost:8545 -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}'
# Expected: {"jsonrpc":"2.0","id":1,"result":"0x2"} (2 peers connected)

# Check Node 2 peer count
curl -s -X POST http://localhost:8546 -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}'
# Expected: {"jsonrpc":"2.0","id":1,"result":"0x2"} (2 peers connected)
```

**⚡ Important**: Consensus will pause initially with "recently connected" - this is normal. The network needs time to establish connections before validators start emitting blocks.

## 5.3 Stop the Network

```bash
pkill -f "u2u.*--datadir"
```

---

# ⚡ **Quick Reference**

## Complete Setup Commands

```bash
# 1. Generate Genesis
cd cmd/makegenesis
go run main.go -output local-genesis.g -network "Local U2U Network" -networkid 4439
cd ../..

# 2. Setup Keystores
mkdir -p u2u-local/node{1,2,3}
go run cmd/setup_validator_node/main.go 1 ./u2u-local/node1
go run cmd/setup_validator_node/main.go 2 ./u2u-local/node2
go run cmd/setup_validator_node/main.go 3 ./u2u-local/node3
touch u2u-local/{password.txt,validator_password.txt}

# 3. Launch Network (3 terminals)
# Terminal 1 - Bootstrap Node:
./build/u2u --datadir ./u2u-local/node1 --genesis ./cmd/makegenesis/local-genesis.g \
  --genesis.allowExperimental --validator.id 1 \
  --validator.pubkey "0xc0048d505c351f4837cec72bce6f4254f5e4bc3f2c9a4816841db64319eee8b714ef9173fbf66d039b782624713791840846b2788d4b65a425adeba85a4b57efe0cd" \
  --validator.password ./u2u-local/validator_password.txt --unlock "0x239fA7623354eC26520dE878B52f13Fe84b06971" \
  --password ./u2u-local/password.txt --allow-insecure-unlock --port 30303 \
  --http --http.addr "127.0.0.1" --http.port 8545 --http.api="eth,trace,debug,net,admin,web3,personal,txpool,dag" \
  --ws --ws.addr "127.0.0.1" --ws.port 8547 --ws.api "eth,web3,net,u2u"

# Get Node 1's enode (after 10 seconds):
curl -s -X POST http://localhost:8545 -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"admin_nodeInfo","params":[],"id":1}' | jq -r '.result.enode'

# Terminal 2 - Add --bootnodes with Node 1's enode
# Terminal 3 - Add --bootnodes with Node 1's enode

# 4. Test Network Connectivity
curl -X POST http://localhost:8545 -H "Content-Type: application/json" -d '{"jsonrpc":"2.0","method":"net_peerCount","params":[],"id":1}'
# Should return {"jsonrpc":"2.0","id":1,"result":"0x2"} (2 connected peers)

# 5. Stop
pkill -f "u2u.*--datadir"
```

## Network Info

| Item | Value |
|------|-------|
| **Network Name** | "Local U2U Network" |
| **Network ID** | 4439 |
| **Validators** | 3 (deterministic keys) |
| **APIs** | HTTP: 8545,8546,8547 / WS: 8547,8548,8549 |
| **Genesis** | `cmd/makegenesis/local-genesis.g` |

## Key Files

- **Genesis Generator**: `cmd/makegenesis/main.go`
- **Keystore Setup**: `cmd/setup_validator_node/main.go`
- **Node Scripts**: `u2u-local/start-node*.sh`
- **Network Control**: `u2u-local/stop-all.sh`

---

# 🛟 **Troubleshooting**

### "Failed to unlock validator key: key is not found"
- **Fix**: Ensure pubkey has `0xc0` prefix: `0xc0048d505c...` not `0x048d505c...`
- **Verify**: Check that keystore was created properly by running setup_validator_node again
- **Check**: Ensure the pubkey format matches exactly what setup_validator_node outputs as "U2U PubKey"

### "Account unlock with HTTP access is forbidden"
- **Fix**: Add `--allow-insecure-unlock` to startup script

### "Genesis file doesn't refer to any trusted preset"
- **Normal**: Expected warning for custom networks

### Database errors
- **Fix**: `rm -rf u2u-local/node*/ && <setup keystores again>`

### "Unavailable modules in HTTP API list unavailable=[u2u]"
- **Normal**: The `u2u` module is not available in this version
- **Solution**: Use available modules: `eth,trace,debug,net,admin,web3,personal,txpool,dag,abft,sfc`

---

# 🚀 **Success!**

Your **U2U Local Blockchain Network** is now running with:
- ✅ Custom genesis (not fakenet)
- ✅ 3 validator nodes with proper authentication
- ✅ Full HTTP/WebSocket APIs
- ✅ Ready for smart contracts and dApps

**Start developing on U2U! 🎉**