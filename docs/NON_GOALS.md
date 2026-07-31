# Explicit Non-Goals & Scope Boundaries - PrefixOS

To maintain project focus, architectural elegance, and high engineering quality, PrefixOS explicitly defines system boundaries and intentional non-goals.

---

## 1. Explicit Non-Goals

1. **Not an LLM Inference Engine**: PrefixOS is a **Control Plane Cache Engine**, not a matrix multiplication runtime. It does not compute attention tensors, execute matrix multiplications, or host CUDA kernels directly.
2. **Not a General-Purpose Vector Database**: PrefixOS matches exact token sequence ID prefixes ($O(K)$ matching). It does not compute semantic vector embeddings, nearest-neighbor searches (HNSW/IVF), or cosine similarities.
3. **Not a Distributed Filesystem**: PrefixOS manages in-memory token-to-block offsets for LLM KV caches. It does not act as a POSIX filesystem, block store, or blob storage engine.
4. **Not a Custom Tokenizer Implementation**: PrefixOS accepts token sequence IDs (`[]int32`). It delegates string tokenization to external tokenizers (e.g., `tiktoken` or HuggingFace tokenizers).
5. **Not a Web GUI / Management Console**: PrefixOS exports rich Prometheus metrics and OpenTelemetry traces for Grafana visualization. It intentionally omits an embedded web dashboard to avoid bloating binary size.
6. **Not a Service Mesh / Kubernetes Operator**: Deployment manifests (Helm / K8s StatefulSets) are provided under `deployments/`, but PrefixOS does not attempt to embed service mesh features (Istio/Linkerd) or custom operators.

---

## 2. Rationale for Boundaries

By tightly constraining scope to **Prefix-Aware Control Plane Cache Management**, PrefixOS guarantees sub-50µs lookup latencies, zero GC pauses, lock-free reader scalability, and clear production boundaries suitable for integration into vLLM, SGLang, and llama.cpp.
