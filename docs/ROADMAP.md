# Project Roadmap & Release Milestones - PrefixOS

This document outlines the release milestone roadmap and long-term research vision for PrefixOS.

---

## 1. Milestone Roadmap

### Phase 1: v1.0 MVP Core (Status: ✅ Active Milestone)
- [x] Slab Memory Allocator (`internal/memory/slab.go`) & Fragmentation Controller
- [x] 64-way Lock-Sharded Radix Tree with RCU Lock-Free Readers (`internal/radix/`)
- [x] Adaptive Cost-Aware Eviction Engine (`internal/eviction/adaptive.go`)
- [x] REST & gRPC API Server Gateway (`internal/api/`)
- [x] Prometheus Metrics Dictionary & Benchmark Suite
- [x] Production Documentation & ADR Package (`docs/`)

### Phase 2: v1.5 Enterprise Persistence (Status: 🚧 Planned)
- [ ] Compressed Write-Ahead Log (WAL) with zstd & CRC32 checksums
- [ ] Copy-On-Write Snapshots & Parallel Crash Recovery Engine
- [ ] Multi-stage Production Docker Image & Kubernetes Helm Chart
- [ ] Idiomatic Go & Python Client SDKs (`pkg/sdk/`)

### Phase 3: v2.0 Distributed Cluster (Status: 🚧 Planned)
- [ ] Raft Consensus Engine for Leader Election & Log Replication (`internal/replication/raft.go`)
- [ ] Dynamic Consistent Hash Ring with Virtual Nodes & Rebalancing
- [ ] Network Partition & Split-Brain Detection Protocols

### Phase 4: v2.5 Ecosystem Adapters (Status: 🚧 Planned)
- [ ] vLLM Direct IPC Adapter & Control Plane Integration
- [ ] SGLang RadixAttention gRPC Client Adapter
- [ ] llama.cpp KV Cache Control Plane Adapter

---

## 2. Long-Term Research Roadmap (v3.0+ Status: 🔬 Research)

1. **GPU Direct Storage (GDS)**: Bypass CPU host memory, streaming attention weight tensors directly between NVMe storage and GPU VRAM via NVIDIA cuFile APIs.
2. **RDMA Transport**: InfiniBand / RoCE (RDMA over Converged Ethernet) transport for sub-10µs node-to-node replication.
3. **eBPF-Based Telemetry**: Kernel-level eBPF probes for zero-overhead socket packet tracing and memory bus latency profiling.
4. **SmartNIC / DPU Offload**: Offloading control plane token prefix hashing to NVIDIA BlueField DPUs.
