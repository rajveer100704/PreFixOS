# Product Requirements Document (PRD) - PrefixOS

**Project Name**: PrefixOS  
**Tagline**: High-Performance Distributed Prefix-Aware Cache Engine for Large Language Model Serving  
**Status**: Milestone v1.0 Specification (Status: ✅ Active Specification)  
**Attribution**: Inspired by the RadixKV project and re-architected into a production-grade distributed prefix-aware cache for LLM inference.

---

## 1. Executive Summary & Vision

Large Language Model (LLM) inference servers encounter severe memory capacity limits and bandwidth bottlenecks—commonly referred to as the **KV Cache Memory Wall**. In multi-turn chat applications, agentic workflows, long-context retrieval, and tree-search decoding, up to 80% of prompt tokens are identical system prompts, developer instructions, or prefix context.

Existing flat KV cache management schemes store duplicate prompt attention tensors independently for every incoming request. **PrefixOS** eliminates this redundancy by orchestrating a concurrent, prefix-aware control plane. By matching token prefixes across concurrent requests using a sharded Radix Tree and managing zero-copy virtual memory blocks, PrefixOS maximizes prompt KV cache reuse, drastically reduces Time-To-First-Token (TTFT) latency, and increases GPU batch throughput by 3x–5x.

---

## 2. Problem Statement & Motivation

1. **VRAM Exhaustion**: KV cache memory scales linearly with sequence length and batch size ($O(B \times L)$). Storing duplicate system prompts for 1,000 concurrent agents depletes available GPU memory.
2. **TTFT Latency Overhead**: Recomputing attention Key/Value matrices for shared prefixes on every request wastes FLOPs and introduces significant prefill latency.
3. **Cache Trashing**: Standard LRU eviction algorithms evict root system prompts when leaf requests burst, causing catastrophic cache misses across all subsequent agent requests.
4. **Single-Node Bottlenecks**: Academic cache prototypes lack distributed replication, persistence, crash recovery, and multi-tenant isolation needed for enterprise LLM serving fleets (vLLM, TensorRT-LLM, SGLang).

---

## 3. Key Target Personas & Use Cases

- **ML Infrastructure Engineers**: Seeking a zero-allocation, distributed control-plane cache to plug into inference clusters (vLLM / SGLang / llama.cpp).
- **Enterprise AI Platform Teams**: Requiring strict multi-tenant memory quota isolation, prometheus telemetry, and high availability.
- **Systems & Distributed Systems Engineers**: Evaluating lock-free data structures, Raft consensus replication, custom slab allocation, and zero-copy memory management.

---

## 4. Functional Requirements

### 4.1 Token Prefix Matching & Storage
- **FR-01**: Must perform $O(K)$ token prefix matching over a sharded Radix Tree (where $K$ is the matching token sequence length).
- **FR-02**: Must return ordered memory block handles (`BlockIDs`) corresponding to cached physical memory offsets.
- **FR-03**: Must support thread-safe concurrent insertion of new token branches with node splitting.

### 4.2 Zero-Allocation Memory Management
- **FR-04**: Pre-allocated Slab Memory Allocator managing virtual blocks without runtime OS memory requests or Go GC pauses.
- **FR-05**: NUMA-aware block allocation and memory quota enforcement per tenant.
- **FR-06**: Background memory compaction and fragmentation tracking.

### 4.3 Pluggable & Adaptive Eviction
- **FR-07**: Support pluggable eviction strategies via a unified `EvictionPolicy` interface (Adaptive Cost-Aware, ARC, TinyLFU, LFU, LRU).
- **FR-08**: Adaptive policy MUST target leaf nodes bottom-up, preventing root prompt eviction.

### 4.4 Enterprise Persistence & Crash Recovery
- **FR-09**: Write-Ahead Log (WAL) with zstd compression and CRC32 checksum validation for every state mutation.
- **FR-10**: Copy-On-Write background snapshots and fast parallel startup recovery.

### 4.5 Distributed Consensus & High Availability
- **FR-11**: Multi-node replication utilizing gRPC-based Raft consensus for leader election and log replication.
- **FR-12**: Dynamic consistent hash ring with virtual nodes for partition resilience and node rebalancing.

---

## 5. Non-Functional Requirements & Key Performance Indicators (KPIs)

| Metric | Target / Requirement |
| :--- | :--- |
| **Lookup Latency (p50)** | `< 50 µs` |
| **Lookup Latency (p99)** | `< 150 µs` |
| **Insert Latency (p50)** | `< 80 µs` |
| **Lookup Throughput** | `> 5,000,000 requests / sec` |
| **TTFT Latency Reduction** | `Up to 75% reduction on shared prefix hits` |
| **GPU Memory Savings** | `Up to 78% reduction in KV cache footprint` |
| **Startup Crash Recovery** | `< 5 seconds for 10M token nodes` |
| **Replication Lag** | `< 100 ms across cluster nodes` |

---

## 6. Staged Milestone Roadmap

### v1.0 MVP Core (Active Release Target)
- Slab Memory Allocator & Fragmentation Controller
- 64-way Lock-Sharded Radix Tree with RCU lock-free readers
- Adaptive Cost-Aware Eviction Engine
- REST & gRPC API Server
- Prometheus Telemetry & Benchmark Suite

### v1.5 Enterprise Persistence
- Compressed WAL Engine & Copy-On-Write Snapshots
- Parallel Startup Recovery Engine
- Multi-stage Docker image & Go/Python SDKs

### v2.0 Distributed Cluster
- Raft Consensus Leader Election & Replication
- Consistent Hash Ring & Node Rebalancing
- Network Partition & Split-Brain Handler

### v2.5 Ecosystem Adapters
- Native IPC Adapters for vLLM, SGLang, and llama.cpp
- OpenAI-compatible Proxy Gateway

### v3.0+ Hardware Acceleration (Research)
- RDMA Transport & GPU Direct Storage (GDS) Integration
- SmartNIC / DPU Control Plane Offload
