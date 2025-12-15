# P2P Transport, Handshake, and Discovery

This document explains how U2U nodes connect, negotiate capabilities, and manage peers at the transport layer.

## Stack Overview

- **Transport**: RLPx-based encrypted TCP sessions (`p2p/peer.go`, `p2p/server.go`).
- **Capabilities**: Subprotocols advertised during handshake (e.g., `u2u/UP01`).
- **Discovery**: Kademlia-based node discovery (v4/v5) plus static bootnodes (`p2p/server.go`).
- **Peer Manager**: Limits, useless-peer filtering, and trusted peer bypasses (`gossip/handler.go`, `gossip/peer.go`).

## Node Startup & Discovery

1. **Server init**: Loads keys, sets listen addresses, configures discovery (`node/node.go`, `p2p/server.go`).
2. **Discovery**: Uses UDP listeners to exchange ENRs and find nodes (v4/v5). Respects `NoDiscovery`/`DiscoveryV5` flags.
3. **Dial Candidates**: Protocols can provide dial sources; bootnodes used if configured.
4. **NAT Handling**: Optional NAT mapping of UDP/TCP ports when exposed.

## Handshake Flow

1. **RLPx Handshake**: Establishes encrypted session and capability set.
2. **Protocol Handshake** (`gossip/handler.go`):
   - Sends `NetworkID`, local progress (DAG/blocks), and genesis hash.
   - Validates peer name/caps (filters non-U2U or incompatible peers).
   - Registers peer in gossip manager and DAG leecher.
3. **Capability Match**: Connection accepted only if protocol versions overlap; trusted peers bypass max-peer limits.

## Peer Management & Limits

- **Max Peers**: Enforced per handler; trusted peers can exceed limit.
- **Useless Peers**: Heuristic banning of peers without required caps/name pattern.
- **Queues**: Each peer has bounded send queues with data semaphores (`gossip/peer.go`).
- **Semaphores**: Message-level semaphores limit concurrent bytes (`handler.msgSemaphore`).
- **Message Size**: Enforced `protocolMaxMsgSize` guard; oversize messages dropped.

## Base Messaging Loop

- **Read Loop**: Per-peer reader dispatches messages to protocol handlers (`p2p/peer.go`).
- **Ping/Pong**: Liveness detection; disconnection on errors.
- **Subprotocol Dispatch**: Routes messages to gossip handler channels.
- **Discard & Metering**: Unknown or oversize messages discarded with metrics.

## Progress Tracking

- **Progress Messages**: Peers periodically share DAG/block progress (`ProgressMsg`).
- **Usage**: Guides what to request/offer (events, blocks, snapshots).
- **Sync Status**: Handler tracks sync state to accept/reject tx propagation when not ready.

## Security & Sanity Checks

- **Genesis Check**: Handshake includes genesis hash; mismatched peers disconnected.
- **Network ID**: Rejects peers on different networks.
- **Double-Sign Prevention**: Local emitter checks for self-forks; peers rejected on malformed events/txs.
- **Bans**: Misbehaving peers can be banned via discfilter.

## Key Implementation References

- Transport & discovery: `p2p/server.go`, `p2p/peer.go`, `p2p/enode/*`
- Node wiring: `node/node.go`, `node/api.go`
- Gossip handler: `gossip/handler.go`, `gossip/protocol.go`
- Peer abstraction: `gossip/peer.go`
- Config: `gossip/config.go`, `u2u/rules.go` (network limits)

