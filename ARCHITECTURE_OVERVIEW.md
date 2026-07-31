# Architecture Overview - PrefixOS

PrefixOS is a **Production-Grade Distributed Prefix-Aware Cache Engine** for Large Language Model Serving frameworks (vLLM, SGLang, llama.cpp, TensorRT-LLM).

> **Attribution Note**: Inspired by the RadixKV architecture and re-architected into a production-grade distributed prefix-aware cache for LLM inference.

---

## Key Subsystems

```
                               PrefixOS Control Plane
                                         │
        ┌───────────────────┬────────────┴───────┬───────────────────┐
        │                   │                    │                   │
  Slab Memory         64-Way Sharded       Adaptive Cost       gRPC & REST
   Allocator            Radix Tree          Eviction API         Gateway
  (Zero Allocation)   (RCU Lock-Free)    (Leaf Target Priority)  (v1 Streaming)
```

1. **Slab Memory Allocator**: Pre-allocates fixed-size block pools bound to CPU NUMA sockets, guaranteeing sub-12% memory overhead and zero runtime GC pauses.
2. **64-Way Sharded RCU Tree**: Splits tree routing across 64 lock-sharded sub-trees, using epoch-based RCU readers for sub-50µs p50 lookup latencies.
3. **Adaptive Eviction Policy**: Multi-factor priority scoring algorithm ($\frac{\text{Hit} \times \text{Depth} \times \text{Cost}}{\text{Age} + 1}$) that targets childless leaf nodes bottom-up, preventing root system prompt eviction.
4. **Protobuf & REST Gateway**: Exposes gRPC Unary/Streaming APIs (`proto/v1/prefixos.proto`), REST HTTP endpoints, Prometheus telemetry, and operational Admin control APIs.

---

## Detailed Documentation
- [PRD (Product Requirements)](docs/PRD.md)
- [TRD (Technical Specs)](docs/TRD.md)
- [SLO & Latency SLAs](docs/SLO.md)
- [Capacity Planning](docs/CAPACITY.md)
- [Architecture Decision Records (ADR 001–004)](docs/ADR/)
