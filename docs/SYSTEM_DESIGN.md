# System Design Specification - PrefixOS

This document details the end-to-end system design, component boundaries, package structure, and lifecycle flows of PrefixOS.

---

## 1. System Component Map

```
PrefixOS Core Engine
 ├── API Layer (internal/api/)
 │    ├── REST Gateway (Gin/Net/HTTP)
 │    ├── gRPC Server (proto/v1/prefixos.proto)
 │    ├── Admin Operations Server (/admin/*)
 │    └── SSE Streaming Server
 │
 ├── Core Control Plane (internal/)
 │    ├── interfaces/        <-- Pure Go Interfaces
 │    ├── memory/            <-- Slab Allocator, NUMA, Defrag, Quotas
 │    ├── radix/             <-- 64-Way Sharded RCU Tree Engine
 │    ├── eviction/          <-- Adaptive, ARC, TinyLFU, LFU, LRU
 │    ├── persistence/       <-- WAL Engine, Snapshots, CRC32 Checksums
 │    ├── replication/       <-- Raft Consensus, Consistent Hash Ring
 │    ├── security/          <-- TLS/mTLS, API Keys, Rate Limiter
 │    ├── metrics/           <-- Prometheus, OpenTelemetry, pprof
 │    ├── tokenizer/         <-- Tiktoken / BPE Wrapper
 │    └── adapters/          <-- vLLM, SGLang, llama.cpp Adapters
 │
 └── Client Libraries (pkg/sdk/)
      ├── go/                <-- Native Go SDK Client
      ├── python/            <-- Python Client SDK
      └── rest/              <-- HTTP REST Client
```

---

## 2. Component Interactions & Dependencies

- `cmd/server` initializes `internal/config`, binds `internal/memory` Slab Allocator, and instantiates the 64-way `internal/radix` Tree.
- `internal/radix` registers leaf nodes with `internal/eviction` policy handlers upon node creation.
- `internal/api` endpoints delegate request validation to `internal/security` before invoking `internal/radix` methods.
- Write operations (`Insert`) publish mutations asynchronously to `internal/persistence` (WAL) and `internal/replication` (Raft Log).

---

## 3. High-Availability & Failure Domains

- **Node Failure**: If a Raft follower node crashes, the cluster continues serving reads and writes without interruption.
- **Leader Failure**: Raft triggers leader election within `< 1.5 seconds`. Clients automatically reconnect to the newly elected leader via the gRPC client load balancer.
- **Disk Corruption**: CRC32 checksum validation detects corrupted WAL entries during startup recovery, reverting safely to the latest valid snapshot.
