# Target System Architecture - PrefixOS

> [!NOTE]
> **Document Status**: Target System Architecture (v1.0+ Production Specification).  
> For the current implementation baseline, see [ARCHITECTURE.md](file:///c:/Users/BIT/PreFixOS/RadixKV/docs/ARCHITECTURE.md).

**Scope**: Control Plane Engine for Distributed LLM KV Cache Optimization  

---

## 1. High-Level System Architecture

PrefixOS decouples the **Control Plane** (Metadata, Prefix Indexing, Memory Slab Allocation, Eviction, Cluster Replication) from the **Data Plane** (CUDA VRAM Attention Tensors).

```
                      +---------------------------------------+
                      |         LLM Serving Cluster           |
                      |   (vLLM / SGLang / llama.cpp Nodes)   |
                      +-------------------+-------------------+
                                          |
                        gRPC Streaming / REST / IPC
                                          v
                      +---------------------------------------+
                      |        PrefixOS Gateway Layer         |
                      |  - REST / gRPC API / Admin Endpoints  |
                      |  - TLS / mTLS & API Key Authentication|
                      |  - Tokenizer Middleware (tiktoken)    |
                      +-------------------+-------------------+
                                          |
                                          v
                      +---------------------------------------+
                      |     Core PrefixOS Engine Cluster      |
                      |                                       |
                      |  +---------------------------------+  |
                      |  | 64-Way Sharded RCU Radix Tree   |  |
                      |  +---------------------------------+  |
                      |  | Slab Allocator + NUMA Controller|  |
                      |  +---------------------------------+  |
                      |  | Pluggable Eviction Framework    |  |
                      |  +---------------------------------+  |
                      |  | Raft Replication & Hash Ring    |  |
                      |  +---------------------------------+  |
                      |  | Compressed WAL & Snapshot Engine|  |
                      |  +---------------------------------+  |
                      +-------------------+-------------------+
                                          |
                                          v
                      +---------------------------------------+
                      |      Observability & Telemetry        |
                      |  - Prometheus Metrics (/metrics)      |
                      |  - OpenTelemetry Tracing              |
                      |  - Go runtime pprof (/debug/pprof)    |
                      +---------------------------------------+
```

---

## 2. Core Subsystem Architecture

### 2.1 Decoupled Interface Layer (`internal/interfaces/`)
PrefixOS isolates domain abstractions into pure Go interfaces:
- `MemoryAllocator`: Manages zero-allocation slab blocks, free-list recycling, NUMA node binding, and background compaction.
- `RadixTreeEngine`: Manages 64-way lock-sharded tree routing, epoch-based RCU lock-free readers, and path/prefix compression.
- `EvictionPolicy`: Selects victim leaf nodes based on cost-aware scoring (hit frequency, tree depth, token length, recency).
- `PersistenceEngine`: Manages append-only WAL writes with zstd compression and Copy-On-Write background snapshots.
- `ReplicationEngine`: Manages Raft consensus, leader election, and dynamic consistent hash ring rebalancing.

---

## 3. Data Flow & Communication Models

1. **Control Plane Lookup**: Client sends token sequence IDs (`MatchPrefix`). Gateway routes query to the target Shard Radix Tree.
2. **Lock-Free Read**: Reader goroutines traverse tree nodes using RCU atomic version counters, returning matched token length and physical `BlockIDs`.
3. **Block Handle Return**: PrefixOS returns array of physical `BlockIDs` and virtual memory offsets to the inference engine.
4. **Data Plane Execution**: Inference engine loads precomputed attention tensors directly from shared IPC memory or GPU VRAM corresponding to `BlockIDs`, eliminating redundant prefill computation.
