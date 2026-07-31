# Benchmark Plan & Methodology - PrefixOS

This document specifies the benchmarking methodology, comparative workloads, synthetic Zipf distribution generators, and real token trace benchmarking for PrefixOS.

---

## 1. Comparative Benchmark Target Matrix

PrefixOS will be benchmarked against 6 baseline caching implementations:

1. **Naive Flat HashMap Cache**: Standard Go `map[string][]Token` with global RWMutex.
2. **Standard Trie**: Standard single-threaded non-sharded prefix Trie.
3. **RadixKV v1**: Academic single-lock prototype baseline.
4. **Redis KV Cache**: External Redis `v7.2` instance over local socket.
5. **LMCache**: Python-based LLM KV cache baseline.
6. **vLLM Prefix Cache**: vLLM built-in block manager prefix cache engine.

---

## 2. Workload Generators & Distributions

### 2.1 Synthetic Zipfian Distribution Generator
- Simulates realistic multi-agent LLM workloads where 20% of system prompts account for 80% of incoming request prefixes (skew parameter $\alpha = 1.2$).
- Token prompt lengths: 512, 1024, 2048, 4096 tokens.
- Concurrency levels: 1, 16, 64, 256, 1024 concurrent reader goroutines.

### 2.2 Real BPE Token Traces
- Uses `tiktoken` (`cl100k_base` / `llama3`) to tokenize real multi-turn chat datasets (LMSYS Chatbot Arena / ShareGPT dataset).

---

## 3. Execution Commands

```bash
# Run full comparative benchmark suite
go test -bench=. -benchmem ./benchmarks/...

# Run concurrent scaling benchmark
go test -bench=BenchmarkConcurrentMatchPrefix -cpu=1,4,16,64 ./benchmarks/...
```
