# ADR-001: 64-Way Radix Tree Sharding & Epoch-Based RCU Readers

**Status**: Accepted  
**Date**: 2026-07-31  
**Deciders**: Lead Infrastructure Architect  

---

## Context & Problem Statement
The baseline RadixKV prototype utilized a single global `sync.RWMutex` lock protecting the entire Radix Tree. Under high concurrency (64+ worker threads), lock contention on the root node serialized reader goroutines, causing latency spikes (`> 5ms`) and limiting throughput.

---

## Considered Alternatives
1. **Single Lock RWMutex**: Simple, but severe lock contention under concurrent read workloads.
2. **Global Lock-Free SkipList**: Complex node splitting logic, poor cash locality for token prefixes.
3. **64-Way Sharded Radix Tree with Epoch-Based RCU Readers**: Partition the tree into 64 lock-sharded trees by token prefix hash, and use epoch-based Read-Copy-Update (RCU) for lock-free reader node traversal.

---

## Decision Outcome
**Chosen Option**: Option 3 (64-Way Sharded Radix Tree + Epoch-Based RCU Readers).

### Consequences
- **Positive**: Eliminates global read lock contention. Lookup latency reduced to `< 50µs` p50, supporting `> 5,000,000 ops/sec`.
- **Negative**: Requires careful memory reclamation logic (`epoch.Sync()`) when pruning or splitting nodes to prevent dangling pointer access by active readers.
