# Changelog - PrefixOS

All notable changes to PrefixOS are documented in this file.

---

## [1.0.0] - 2026-07-31 (Release Candidate)

### Added
- **Slab Memory Allocator**: Pre-allocated size-class block allocator with NUMA node binding and zero GC pauses.
- **64-Way Sharded RCU Radix Tree**: Lock-sharded tree engine featuring epoch-based RCU lock-free readers (`MatchPrefix`).
- **Pluggable Eviction Architecture**: Added `EvictionPolicy` interface with default **Adaptive Cost-Aware** leaf targeting algorithm.
- **API & Observability**: REST HTTP endpoints, gRPC v1 schema (`proto/v1/prefixos.proto`), Prometheus metrics dictionary, and `pprof` profiling handlers.
- **Documentation Suite**: Comprehensive technical specifications (PRD, TRD, SLO, CAPACITY, COMPATIBILITY, MIGRATION, RELEASE, NON_GOALS, ARCHITECTURE_PRINCIPLES, RISKS, ADRs 001–004).

### Changed
- Re-architected academic RadixKV prototype into production **PrefixOS** engine.
