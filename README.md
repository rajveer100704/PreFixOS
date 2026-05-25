# RadixKV

> A high-performance, prefix-aware KV cache orchestrator for LLMs. Uses a concurrent Radix Tree and O(1) Min-Heap eviction to eliminate redundant token processing in multi-agent AI systems.

![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)
![gRPC](https://img.shields.io/badge/gRPC-Ready-244c5a?style=for-the-badge)
![Architecture](https://img.shields.io/badge/Architecture-Control_Plane-blue?style=for-the-badge)

RadixKV solves the memory wall in highly-concurrent Large Language Model deployments. By sharing identical prompt prefixes (like system prompts) across thousands of parallel branches, it drastically reduces VRAM usage and Time-To-First-Token (TTFT). It leverages physical memory paging concepts (similar to PagedAttention) while maintaining massive concurrency via lock-coupling.

## Documentation
- [Architecture Deep Dive](docs/ARCHITECTURE.md) - Learn about the Control Plane, Radix Tree, and memory semantics.
- [gRPC API Reference](docs/API.md) - Learn how to interface your Inference Engine with RadixKV.

## Core Features
- **Zero-Allocation Memory Layer**: Simulates VRAM paging via a block-based memory manager. RadixKV acts as a Control Plane generating `MemoryOffsets` for your C++ Data Plane.
- **Concurrent Radix Tree**: Uses fine-grained lock-coupling semantics to safely mutate the tree while maintaining massive parallelism for token prefix lookups.
- **O(1) Tree-Aware LRU Eviction**: A highly optimized Garbage Collector utilizing a Min-Heap. It exclusively targets conversational leaf nodes, pruning dead branches from the bottom up while automatically preserving shared foundations.

## Project Structure
```text
radixkv/
├── cmd/
│   └── server/                 # Entry point: initializes memory, tree, and starts gRPC server
├── internal/
│   ├── memory/                 # Control Plane metadata pointers and MemoryOffset definitions
│   ├── radix/                  # Thread-safe insertion and prefix matching logic
│   └── eviction/               # O(1) Min-Heap LRU targeting leaf nodes
├── proto/                      # The gRPC API definitions
├── docs/                       # Architectural and API documentation
└── benchmarks/                 # Simulates agent tree-search vs standard cache
```

## Setup & Running

### 1. Compile Protobufs
You must generate the Go stubs from the protobuf file before running the server.
```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/kvcache.proto
```

### 2. Run the Server
```bash
go run cmd/server/main.go
```

### 3. Run the Benchmarks
Observe the ~78% memory footprint savings of the Radix Tree structure against a naive flat KV cache:
```bash
go test -bench=. ./benchmarks/...
```
