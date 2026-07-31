# Migration Guide: RadixKV v1 to PrefixOS

This guide outlines the upgrade steps for migrating from the academic RadixKV prototype to the production-grade **PrefixOS** cache engine.

---

## 1. Overview of Key Changes

| Component | Academic RadixKV (v1) | Production PrefixOS (v1.0+) |
| :--- | :--- | :--- |
| **Package Layout** | `internal/memory`, `internal/radix` | `internal/interfaces/`, `internal/memory/`, `internal/radix/` |
| **Concurrency** | Single `sync.RWMutex` lock | 64-way Sharded Radix Tree + Epoch-based RCU Readers |
| **Memory Allocation** | Slice free-list | Slab Allocator + NUMA Aware + Compaction + Quotas |
| **Eviction** | Single Min-Heap LRU | Pluggable Policies (Adaptive, ARC, TinyLFU, LFU, LRU) |
| **Persistence** | None (RAM only) | Compressed WAL + Incremental Snapshots |
| **Clustering** | Single process | gRPC Raft Consensus + Consistent Hash Ring |
| **API** | Basic gRPC | gRPC Bidirectional Streaming + REST SSE + Admin APIs |

---

## 2. API Migration Code Example

### Legacy RadixKV Go Invocation:
```go
import "radixkv/internal/radix"

tree := radix.NewTree(memManager, evictionHeap)
matched, blockIDs := tree.MatchPrefix(tokens)
```

### Modern PrefixOS Go SDK Invocation:
```go
import "github.com/PrefixOS/PrefixOS/pkg/sdk/go"

client, err := prefixos.NewClient(prefixos.Config{
    Target: "localhost:50051",
    EnableTLS: true,
})
res, err := client.MatchPrefix(ctx, tokens)
```

---

## 3. Configuration & State Migration

1. **Config File**: Move inline config variables to `configs/config.yaml`.
2. **Data Persistence**: Initialize `storage.wal_path` and `storage.snapshot_dir` in `config.yaml` to ensure durability across restarts.
3. **Protobuf Stubs**: Update import paths from `radixkv/proto` to `github.com/PrefixOS/PrefixOS/proto/v1`.
