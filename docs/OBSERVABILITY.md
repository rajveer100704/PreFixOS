# Observability, Metrics & Telemetry - PrefixOS

This document specifies the Prometheus metrics dictionary, OpenTelemetry tracing spans, Zap structured logging formats, and Go `pprof` profiling handlers for PrefixOS.

---

## 1. Prometheus Metrics Dictionary (`/metrics`)

| Metric Name | Type | Description | Labels |
| :--- | :--- | :--- | :--- |
| `prefixos_cache_hits_total` | Counter | Total number of token prefix cache hits | `tenant_id`, `shard_id` |
| `prefixos_cache_misses_total` | Counter | Total number of token prefix cache misses | `tenant_id`, `shard_id` |
| `prefixos_lookup_latency_seconds` | Histogram | Token prefix lookup latency distribution | `status`, `shard_id` |
| `prefixos_insert_latency_seconds` | Histogram | Token insertion latency distribution | `status`, `shard_id` |
| `prefixos_prefix_reuse_ratio` | Gauge | Ratio of cached prefix tokens to total tokens | `tenant_id` |
| `prefixos_ttft_savings_percent` | Gauge | Estimated Time-To-First-Token latency reduction % | `tenant_id` |
| `prefixos_slab_blocks_allocated` | Gauge | Number of active allocated slab blocks | `size_class` |
| `prefixos_slab_fragmentation_ratio` | Gauge | Memory slab fragmentation ratio (0.0 to 1.0) | None |
| `prefixos_lock_contention_wait_seconds`| Histogram | Mutex wait time histogram per tree shard | `shard_id` |

---

## 2. Profiling Endpoints (`/debug/pprof/`)

PrefixOS exposes standard Go runtime profiling endpoints for diagnosing CPU utilization, heap allocations, and lock contention:

- `GET /debug/pprof/profile?seconds=30`: CPU Profile
- `GET /debug/pprof/heap`: Memory Heap Allocation Profile
- `GET /debug/pprof/mutex`: Mutex Lock Contention Profile
- `GET /debug/pprof/goroutine`: Active Goroutine Stack Traces
