# ADR-002: Pre-Allocated Slab Memory Pool & NUMA Node Binding

**Status**: Accepted  
**Date**: 2026-07-31  
**Deciders**: Lead Systems Engineer  

---

## Context & Problem Statement
Standard dynamic slice allocations in Go (`make([]Token, len)`) cause frequent OS memory requests and trigger Go GC runtime mark-sweep pauses under high throughput, introducing unpredictable latency tails.

---

## Decision Outcome
**Chosen Option**: Implement a custom **Slab Memory Allocator** (`internal/memory/slab.go`) that pre-allocates fixed-size block pools bound to local CPU NUMA nodes at startup.

### Consequences
- **Positive**: Zero runtime Go GC pauses attributable to prefix tree blocks. Sub-12% memory overhead.
- **Negative**: Requires explicit free-list recycling management and background defragmentation compaction.
