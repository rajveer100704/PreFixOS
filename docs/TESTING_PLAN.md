# Testing Plan & Quality Assurance Matrix - PrefixOS

This document specifies the unit testing, integration testing, Go race detection, fuzzing, chaos fault injection, and load testing framework for PrefixOS.

---

## 1. Testing Framework Matrix

| Testing Level | Scope / Target | Execution Tool | Target Coverage / Metric |
| :--- | :--- | :--- | :--- |
| **Unit Tests** | Subsystem packages (`memory`, `radix`, `eviction`) | `go test -v ./internal/...` | `> 85% Code Coverage` |
| **Race Detection** | Concurrency safety under high goroutine load | `go test -race ./internal/...` | `0 Data Races` |
| **Fuzzing** | Tree insertion, node splits, token parsing | `go test -fuzz=FuzzRadixInsert` | `0 Crashes / Panics` |
| **Integration** | End-to-end API gateway, WAL, Snapshots | `go test -tags=integration ./tests/...` | `100% Core Flows Pass` |
| **Chaos Injection** | Network partition, Raft leader kill, disk drop | `tests/chaos/run_chaos.sh` | `Zero Data Loss / Quick Failover` |
| **Load Testing** | 10,000 concurrent gRPC requests | `ghz --insecure --proto=...` | `p99 < 150µs` |

---

## 2. Test Execution Commands

```bash
# Run all unit tests with race detector
go test -v -race -cover ./...

# Run tree insertion fuzz testing
go test -fuzz=FuzzTreeInsert -fuzztime=1m ./internal/radix/

# Run chaos fault injection suite
./scripts/run_chaos_tests.sh
```
