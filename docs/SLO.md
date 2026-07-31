# Service Level Objectives (SLO) & Numerical Targets - PrefixOS

This document specifies the target Service Level Objectives (SLOs), performance SLAs, and quantitative metrics for PrefixOS.

---

## 1. Quantitative Performance Objectives

| Operational Metric | Target Metric (p50) | Target Metric (p95) | Target Metric (p99) | Benchmark Condition |
| :--- | :--- | :--- | :--- | :--- |
| **Lookup Latency** | `< 45 µs` | `< 90 µs` | `< 150 µs` | 64 concurrent reader goroutines, 4K prompt length |
| **Insert Latency** | `< 75 µs` | `< 120 µs` | `< 250 µs` | Node split + block allocation |
| **Batch Lookup Latency** | `< 120 µs` | `< 250 µs` | `< 400 µs` | Batch size = 32 requests |
| **Throughput** | `> 5,500,000 ops/sec` | `> 4,800,000 ops/sec` | `> 4,000,000 ops/sec` | Synthetic Zipfian workload |
| **Memory Allocation Overhead** | `< 12%` | `< 15%` | `< 18%` | Compared to raw slice storage |
| **Recovery Time** | `< 4.5 seconds` | `< 6.0 seconds` | `< 8.0 seconds` | Cold start from 10M token snapshot |
| **Replication Lag** | `< 50 ms` | `< 80 ms` | `< 120 ms` | 3-node Raft cluster over gRPC |

---

## 2. Resource Utilization Limits

- **CPU Overhead**: `< 5%` total CPU usage during steady-state lookup operations.
- **Go GC Pauses**: Zero GC pauses attributable to prefix tree nodes (achieved via Slab block pre-allocation and struct recycling).
- **Lock Contention**: Maximum mutex wait time `< 10 µs` due to 64-way shard partitioning and RCU lock-free readers.

---

## 3. Availability & Failure Metrics

- **System Uptime**: `99.99%` availability in 3-node Raft cluster mode.
- **Max Unplanned Downtime**: `< 5 minutes / year`.
- **Raft Leader Election Time**: `< 1.5 seconds` on leader failure.
