# Technical Risk Register & Mitigations - PrefixOS

This document catalogs technical risks, performance edge cases, failure scenarios, and their corresponding mitigations in PrefixOS.

---

## 1. Technical Risk Register

| Technical Risk | Impact | Probability | Architectural Mitigation Strategy |
| :--- | :--- | :--- | :--- |
| **Lock Contention on Root Tree Nodes** | High | Medium | **Mitigation**: 64-way shard partitioning combined with Epoch-based RCU lock-free readers. Writes only lock the localized target shard. |
| **Memory Fragmentation in Free Lists** | Medium | High | **Mitigation**: Multi-size class Slab Allocator with background defragmentation compaction worker (`defrag.go`). |
| **Unbounded WAL Log Growth** | High | Medium | **Mitigation**: Background zstd log compression combined with periodic Copy-On-Write snapshots and WAL truncation. |
| **Split-Brain on Network Partition** | Critical | Low | **Mitigation**: Raft consensus quorum validation. Non-majority partitions automatically reject writes and transition to read-only state. |
| **Slow Recovery on Node Restart** | High | Low | **Mitigation**: Parallel WAL replay across multi-threaded workers combined with fast binary snapshot loading (`< 5s` for 10M tokens). |
| **Hot Shard Imbalance** | Medium | Medium | **Mitigation**: Consistent Hash Ring with virtual nodes dynamically redistributing token prefix shards across cluster nodes. |
| **Root System Prompt Eviction** | Critical | Low | **Mitigation**: Adaptive Cost-Aware Eviction policy exclusively targeting leaf nodes (`len(Children) == 0`) bottom-up. |

---

## 2. Risk Audit & Continuous Validation

- **Go Race Detector**: All integration tests MUST be executed with `go test -race ./...` in CI/CD.
- **Chaos Fault Injection**: Network partition scenarios simulated in `tests/chaos/` to verify Raft failover guarantees.
