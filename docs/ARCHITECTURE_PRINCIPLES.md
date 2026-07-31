# Core Architecture Principles - PrefixOS

This document outlines the core architectural principles governing all design, code contributions, and refactoring within PrefixOS.

---

## 1. The Seven Core Principles

### Principle 1: Interface-First Design
- Every major subsystem (`MemoryAllocator`, `RadixTreeEngine`, `EvictionPolicy`, `PersistenceEngine`, `ReplicationEngine`) MUST be declared as a pure Go interface under `internal/interfaces/`.
- Concrete implementations MUST NOT leak implementation details into consumer packages.

### Principle 2: Lock-Free Read Path
- The hot read path (`MatchPrefix`) MUST execute without acquiring exclusive write locks.
- Utilize 64-way shard partitioning combined with epoch-based RCU (Read-Copy-Update) node traversal to achieve sub-50µs p50 latencies under high concurrency.

### Principle 3: Zero-Allocation Memory Mechanics
- The control plane MUST NOT perform runtime OS memory allocations or trigger Go GC pauses during steady-state request processing.
- Memory block metadata MUST be managed via pre-allocated Slab pools and struct recycling.

### Principle 4: Fail-Fast Configuration & Strict Validation
- Invalid configurations (e.g., zero slab blocks, invalid shard counts, corrupt WAL paths) MUST cause the server to fail immediately at startup with detailed diagnostic logging.

### Principle 5: Backward-Compatible APIs
- Public Protobuf schemas under `proto/v1/` MUST maintain strict forward and backward compatibility. No field number re-use or breaking type changes without a major version increment (`proto/v2/`).

### Principle 6: Observability by Default
- Every subsystem MUST expose Prometheus metrics (counters, gauges, latency histograms) and OpenTelemetry tracing spans without requiring external plugins.

### Principle 7: Benchmark-Driven Optimization
- Architectural changes or performance claims MUST be backed by reproducible Go benchmarks (`go test -bench=.`) comparing against baseline implementations under realistic Zipf prompt distributions.
