# Memory Layout & Allocation Model - PrefixOS

This document specifies the internal memory layout, Slab Allocator design, NUMA node binding, and defragmentation mechanics of PrefixOS.

---

## 1. Zero-Allocation Memory Hierarchy

PrefixOS eliminates Go GC runtime overhead by managing memory through a fixed-size Slab Allocator pool hierarchy:

```
Memory Pool (RAM / Huge Pages)
  └── Slab Classes (Size Classes: 16, 32, 64, 128 Tokens)
       └── Pages (Aligned Virtual Memory Blocks)
            └── Blocks (Physical Block Metadata Handle)
                 └── Tokens ([]int32 Token Sequence Pointer)
```

---

## 2. Block Metadata Layout

Each `Block` metadata struct is packed into 32 bytes to ensure cache-line alignment (2 blocks per 64-byte L1 cache line):

```go
type Block struct {
    ID          int32    // 4 Bytes: Physical Block ID
    SizeClass   uint16   // 2 Bytes: Size Class Index
    RefCount    uint16   // 2 Bytes: Active Reference Count
    VirtualOffset uint64 // 8 Bytes: IPC Virtual Memory Offset
    NextFree    int32    // 4 Bytes: Free-list Next Pointer
    _padding    [12]byte // 12 Bytes: Padding to align struct to 32 Bytes
}
```

---

## 3. NUMA Node Binding & Compaction

- **NUMA Allocation**: Memory slabs are bound to local CPU NUMA sockets using `mbind()` / Linux libnuma syscalls, avoiding cross-socket memory bus latencies.
- **Fragmentation Tracking**:
  $$\text{FragmentationRatio} = 1.0 - \left( \frac{\text{ContiguousFreeBlocks}}{\text{TotalFreeBlocks}} \right)$$
- **Background Defragmentation**: When `FragmentationRatio > 0.35`, the background compaction worker (`defrag.go`) compacts non-contiguous free blocks during idle cycles.
