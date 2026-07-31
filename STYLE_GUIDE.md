# Go Coding & Architecture Style Guide - PrefixOS

This document defines coding conventions, package layout rules, error handling, and performance idioms for PrefixOS.

---

## 1. Core Idioms

1. **Interface Isolation**: Pure interfaces belong under `internal/interfaces/`. Implementations MUST NOT export un-interfaced global structs.
2. **Error Handling**: Wrap errors with context (`fmt.Errorf("slab allocate: %w", err)`). Never swallow errors silently or return dummy fallbacks.
3. **No Allocation Hot Paths**: Avoid standard slice creation (`make([]byte, size)`) inside hot lookup loops (`MatchPrefix`). Allocate from pre-warmed Slab pools.
4. **Concurrency Safety**: Run `go test -race ./...` before submitting any PR. Mutexes must strictly guard localized struct fields.
