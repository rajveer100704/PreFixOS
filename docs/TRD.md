# Technical Requirements Document (TRD) - PrefixOS

**System**: PrefixOS  
**Language**: Go 1.22+  
**Architecture**: Control Plane for LLM KV Cache Optimization  
**Status**: Specification (Status: ✅ Active)

---

## 1. System Architecture Overview

PrefixOS operates as an out-of-band **Control Plane**. It maintains metadata indexing, prefix matching, token-to-block mapping, and memory lifecycle management, while leaving physical attention weight tensor ops to the Data Plane (e.g., vLLM CUDA backend).

```
                      +-----------------------------+
                      |   Inference Engine Client   |
                      |   (vLLM / SGLang / Python)  |
                      +--------------+--------------+
                                     |
                          gRPC / REST / SDK
                                     v
                      +-----------------------------+
                      |     PrefixOS Control Plane   |
                      |  - REST / gRPC API Gateway  |
                      |  - Sharded Radix Tree       |
                      |  - Slab Memory Allocator    |
                      |  - Pluggable Eviction Engine|
                      |  - Raft Cluster Replication |
                      +--------------+--------------+
                                     |
                            Memory Offsets API
                                     v
                      +-----------------------------+
                      |      Data Plane Memory      |
                      |   (Shared IPC / GPU VRAM)   |
                      +-----------------------------+
```

---

## 2. Software Design & Package Contracts

### 2.1 Interface Decoupling (`internal/interfaces/`)

All core subsystems MUST implement decoupled domain interfaces defined under `internal/interfaces/`:

```go
// MemoryAllocator governs zero-allocation block management
type MemoryAllocator interface {
    Allocate(count int) ([]int32, error)
    Free(blockIDs []int32) error
    FragmentationRatio() float64
    Stats() AllocatorStats
}

// RadixTreeEngine governs concurrent prefix matching & storage
type RadixTreeEngine interface {
    MatchPrefix(tokens []int32) (matchedLength int, blockIDs []int32)
    Insert(tokens []int32, blockIDs []int32) error
    EvictLeaf(leafNodeID uint64) ([]int32, error)
}

// EvictionPolicy governs victim selection
type EvictionPolicy interface {
    SelectVictim() (nodeID uint64, err error)
    OnAccess(nodeID uint64, hitCount uint64)
    OnInsert(nodeID uint64, depth int, cost float64)
}

// PersistenceEngine governs durability & crash recovery
type PersistenceEngine interface {
    AppendWAL(entry WALEntry) error
    CreateSnapshot() (snapshotID string, err error)
    Recover() (lastSequence uint64, err error)
}
```

---

## 3. Concurrency & Threading Model

### 3.1 Lock-Sharded Radix Tree
- The main tree is split into 64 lock-sharded sub-trees based on hash of the root token sequence:
  $$\text{ShardIndex} = \text{Hash}(\text{Tokens}[0..M]) \pmod{64}$$
- Eliminates global tree locks, bound to single-threaded CPU cache lines.

### 3.2 Epoch-Based RCU (Read-Copy-Update) Readers
- Reader threads traverse nodes without acquisition of write mutexes.
- Nodes utilize versioned atomic counters (`uint64`).
- Memory reclamation for deleted/split nodes is deferred via epoch reclamation (`epoch.Sync()`), ensuring lock-free read throughput exceeding $5,000,000\text{ ops/sec}$.

---

## 4. Storage & Persistence Contracts

### 4.1 Write-Ahead Log (WAL)
- Format: Binary append-only log file (`prefixos.wal`).
- Layout per entry:
  `[Magic Header (4B)][Sequence (8B)][Type (1B)][Payload Length (4B)][Payload (zstd compressed)][CRC32 Checksum (4B)]`
- Sync Strategy: Configurable (`ALWAYS`, `EVERY_SEC`, `NEVER`).

### 4.2 Copy-On-Write Snapshots
- Periodic background snapshots freeze an epoch tree reference and dump the serialized memory map into `snapshots/snap-<seq>.bin`.

---

## 5. Network & Clustering Protocol

### 5.1 Protocol Buffers & gRPC Streaming
- Schema: `proto/v1/prefixos.proto`.
- Supports unary (`MatchPrefix`, `Insert`) and bidirectional streaming (`MatchPrefixStream`, `InsertStream`) for high-throughput pipeline tokenization.

### 5.2 Raft Consensus & Memberlist
- Cluster nodes use Raft consensus for leader election, membership tracking, and WAL entry replication over gRPC.
