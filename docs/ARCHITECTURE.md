# RadixKV Architecture

RadixKV is fundamentally designed as a **Control Plane** for LLM orchestration. This document details the core subsystems that make RadixKV capable of supporting 100,000+ concurrent agent branches.

## 1. Control Plane vs. Data Plane

RadixKV is built in Go to leverage high concurrency, quick development cycles, and native gRPC support. 
- **The Control Plane (RadixKV)**: Handles complex distributed locking, prefix matching, node splitting, and memory layout tracking.
- **The Data Plane (Inference Engine)**: Handles actual matrix multiplication on hardware (e.g., vLLM or a custom C++ CUDA backend).

RadixKV **does not** store or send raw heavy attention tensors (float16 matrices) over the network. It communicates purely via Metadata (`BlockIDs` and `MemoryOffsets`). These metadata pointers inform the C++ engine exactly where to read or write the pre-computed attention weights directly from shared IPC memory or CUDA-accessible memory pools.

## 2. The Concurrent Radix Tree
Standard standard cache eviction (LRU) would evict a heavily shared root prompt if it hasn't been accessed recently, causing massive cache misses for all dependent agents. RadixKV uses a Radix (Prefix) Tree.

- **Lock Coupling**: To avoid global write locks stalling readers, RadixKV uses localized `sync.RWMutex` lock-coupling. 
- **Traversal**: Readers use an `RLock` cascading down the nodes, allowing hundreds of LLM agents to simultaneously read the shared system prompt.
- **Node Splitting**: During an `Insert` divergence, the tree extracts the non-shared suffix, passes it to a new grandchild, truncates the parent, and physically frees only the truncated memory blocks.

## 3. Zero-Allocation Memory Manager
Rather than using standard Go slices that trigger garbage collection, RadixKV pre-allocates a massive pool of `Block` pointers. The `Allocate` and `Free` methods strictly push and pop from a free-list slice, guaranteeing O(1) allocation without OS memory requests or Go GC pauses.

## 4. O(1) Min-Heap Garbage Collection
The `TreeLRU` system perfectly prevents prompt hallucinations or cache misses on root prompts. 
- It uses a `container/heap` Priority Queue to maintain an $O(1)$ lookup of the oldest nodes.
- **Leaf-Only Targeting**: It strictly targets `len(Children) == 0` (leaf nodes).
- **Cascading Eviction**: If a leaf deletion leaves the parent childless, the parent dynamically transitions into a leaf and is pushed to the Min-Heap queue. This prunes dead tree branches from the bottom up.
