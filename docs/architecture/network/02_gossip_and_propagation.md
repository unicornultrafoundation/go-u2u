# Gossip Protocol and Data Propagation

This document describes how transactions and events are announced, fetched, and broadcast across peers.

## Message Types (Gossip)

- **ProgressMsg**: Shares DAG/block progress to drive sync decisions.
- **NewEvmTxHashesMsg**: Announces transaction hashes to peers.
- **EvmTxsMsg**: Sends full transactions.
- **NewEventHashesMsg**: Announces event hashes.
- **EventsMsg**: Sends full events (DAG payloads).

All messages are bounded by `protocolMaxMsgSize` and semaphore limits.

## Broadcast Pipeline

1. **Local Processing**: When a tx/event is accepted, it is marked known and optionally broadcast.
2. **Peer Queues**: Each peer has an async writer queue capped by `MaxQueuedItems`/`MaxQueuedSize` (`gossip/peer.go`).
3. **Deduplication**: Known tx/event hashes tracked per peer; duplicates are not re-sent.
4. **Semaphore Guards**: Message-level semaphores throttle concurrent bytes (`handler.msgSemaphore`).
5. **Backpressure**: If semaphores or queues fill, messages are skipped to avoid blocking node internals.

## Transaction Propagation

- **Announce First**: Send hashes (`NewEvmTxHashesMsg`); peers request missing txs.
- **Fetch Path**: `txFetcher` schedules requested txs; `handleTxs` delivers to pool.
- **Pool Gate**: Propagation is blocked until node is sufficiently synced (`syncStatus.AcceptTxs()`).
- **Validation**: Tx pool validates before acceptance; underpriced/invalid txs dropped.
- **Turn-Based Selection**: Inclusion in events uses fair turn rules (see transaction docs).

## Event (DAG) Propagation

- **Announce First**: Send event hashes (`NewEventHashesMsg`); peers request unknown events.
- **Fetch Path**: `dagFetcher` schedules event requests; `handleEvents` enqueues into DAG processor.
- **Ordering**: Events may be delivered out-of-order; DAG processor handles parent checks and buffering.
- **Parent Checks**: `CheckParents/CheckParentless` enforce structure; bad parents can ban peers.
- **Lamport Guard**: Handler limits how far ahead announced events can be vs local highest Lamport, to avoid DoS.

## Flow Control & Limits

- **Message Size**: Hard cap per message (`protocolMaxMsgSize`).
- **Semaphores**: Limits for messages, events, block votes, etc. (`gossip/config.go`).
- **Queue Caps**: Per-peer queue length and byte caps (`PeerCacheConfig`).
- **Timeouts**: DAG stream sessions have timeouts; fetchers have scheduling deadlines.

## Progress & Sync Interaction

- **ProgressMsg** drives which data to request/offer.
- **Sync Gate**: Tx/events may be ignored if node not ready (e.g., during snap sync).
- **Force Syncing**: High Lamport announces can trigger DAG leecher to force sync.

## Error Handling

- **Decoding Errors**: Malformed messages cause disconnect with error response.
- **Ban Conditions**: Invalid events (parent check failures, size violations) can ban peers.
- **Drop on Overload**: When semaphores or queues are saturated, messages are dropped gracefully.

## Implementation References

- Handler/message loop: `gossip/handler.go`
- Peer broadcast queues: `gossip/peer.go`
- Fetchers: `gossip/protocols/dag/dagstreamleecher`, `gossip/protocols/snap/snapstream/snapleecher`, `gossip/itemsfetcher`
- Tx handling: `gossip/handler.go` (tx sections), `evmcore/tx_pool.go`
- Event checks: `eventcheck/*`
- Config: `gossip/config.go`


