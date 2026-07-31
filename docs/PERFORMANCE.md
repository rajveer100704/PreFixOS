# Performance Tuning & Optimization Guide - PrefixOS

This guide covers kernel tuning, Go runtime environment variables, CPU affinity, NUMA memory optimization settings, and empirical benchmark evidence for high-performance PrefixOS deployments.

---

## 1. Operating System & Kernel Tuning

```bash
# Increase maximum file descriptors
sysctl -w fs.file-max=2097152

# Enable Huge Pages (2MB pages)
sysctl -w vm.nr_hugepages=1024

# Set Go memory limit to avoid swap memory thrashing
export GOMEMLIMIT=12GiB
export GOGC=100
```

---

## 2. Core CPU Pinning

For sub-40µs p50 latencies:
```bash
# Pin PrefixOS process to specific CPU NUMA node 0
taskset -c 0-15 ./prefixos --config=configs/config.yaml
```

---

## 3. Benchmark History & Measured Empirical Evidence (Milestone 6)

### Benchmark Execution Commands

```bash
# Run full benchmark suite inside benchmarks directory
cd benchmarks && go test -bench=. -benchmem -cpuprofile=cpu.pprof -memprofile=mem.pprof
```

### Empirical Benchmark Baseline Results

> [!NOTE]
> **Test Environment**: Go microbenchmarks executed on a single process with warm cache on a local development node. Network socket overhead and gRPC transport latencies are measured separately in Milestone 7 integration tests.

| Subsystem Benchmark | Operation / Prompt Size | Latency per Op | Memory / Op | Allocations | Notes |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Slab Allocator** | Single block allocate/free | **141.4 ns** | 16 B | 1 alloc | Zero VRAM heap allocs |
| **Slab Allocator** | Batch (16 blocks) allocate/free | **1,833 ns** | 256 B | 1 alloc | Zero VRAM heap allocs |
| **Radix Engine** | MatchPrefix (100 tokens) | **97.97 ns** | 496 B | 1 alloc | Ultra-fast prefix lookup |
| **Radix Engine** | MatchPrefix (500 tokens) | **406.8 ns** | 2,048 B | 1 alloc | Sub-microsecond lookup |
| **Radix Engine** | MatchPrefix (1,000 tokens) | **774.8 ns** | 4,096 B | 1 alloc | Meets sub-1µs target |
| **Radix Engine** | MatchPrefix (2,000 tokens) | **1,504 ns** | 8,192 B | 1 alloc | Linear scaling with tokens |
| **Radix Engine** | MatchPrefix (4,000 tokens) | **3,146 ns** | 16,384 B | 1 alloc | 4k context window lookup |
| **Radix Engine** | Concurrent Read Throughput | **130.6 ns** | 4,096 B | 1 alloc | Parallel reader goroutines |
| **Persistence Engine**| WAL Append | **1,113 ns** | 96 B | 2 allocs | Append-only binary log + CRC32 |
| **Persistence Engine**| Snapshot Creation | **629.4 µs** | 4,372 B | 103 allocs | CoW binary snapshot checkpoint |

### Memory Reduction Metrics (`BenchmarkAgenticTreeSearch`)

- **Root System Prompt**: 2,000 tokens
- **Sub-Agents**: 50 concurrent agent branches (500 tokens each)
- **Naive Flat Cache Capacity**: 7,813 blocks (~125,000 tokens in VRAM)
- **PrefixOS Shared Cache Capacity**: 1,725 blocks (~27,600 tokens in VRAM)
- **VRAM Memory Reduction**: **77.92% net reduction**
