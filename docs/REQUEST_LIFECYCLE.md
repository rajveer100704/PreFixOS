# Request Lifecycle & Flow Traces - PrefixOS

This document details the step-by-step request flow trace and sequence logic for critical operational paths in PrefixOS.

---

## 1. Request Flow Scenarios

PrefixOS defines sequence flows for 11 core request lifecycles:

1. **Cache Hit Path**: Client sends `MatchPrefix` request -> API Gateway validates -> Shard Hash Router maps to target Shard -> RCU Lock-Free Reader traverses tree -> Returns `matched_length` & `BlockIDs` -> Telemetry records cache hit & TTFT savings.
2. **Cache Miss & Storage Path**: Token sequence misses prefix match -> `Insert` allocates physical blocks from Slab Allocator -> Inserts node into Shard Radix Tree -> Writes entry to WAL -> Replicates via Raft -> Returns assigned `BlockIDs`.
3. **Adaptive Eviction Path**: Slab Memory pressure exceeds threshold -> Eviction Manager queries `EvictionPolicy.SelectVictim()` -> Leaf node with lowest score selected -> Node deleted -> Memory blocks freed to Slab Free-List.
4. **Snapshot Creation Path**: Background timer triggers -> COWs tree reference -> Dumps serialized snapshot to disk -> Truncates WAL.
5. **Node Crash & Recovery Path**: Server restarts -> Loads latest Snapshot binary into memory -> Replays pending WAL entries -> Restores Shard Radix Tree.
6. **Raft Cluster Replication Path**: Leader receives `Insert` -> Appends to local Raft Log -> Broadcasts `AppendEntries` RPC to Followers -> Commits upon majority ACK.
7. **Leader Failure & Failover Path**: Leader heartbeat times out -> Follower initiates Raft election -> Obtains quorum -> Becomes new Leader -> Re-routes client traffic.
8. **Node Join & Rebalance Path**: New node registers -> Consistent Hash Ring adds virtual nodes -> Rebalances token prefix shards -> Syncs state snapshot.
9. **Node Leave Path**: Node leaves cluster -> Hash Ring updates -> Neighboring nodes claim orphaned shards.
10. **Concurrent Insert Path**: Parallel goroutines invoke `Insert` on separate shards -> Processed concurrently without lock contention.
11. **Concurrent Lookup Path**: Parallel goroutines invoke `MatchPrefix` -> Traverse RCU nodes lock-free in sub-50µs latency.
