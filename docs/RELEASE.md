# Release Strategy & Governance - PrefixOS

This document specifies the Semantic Versioning rules, release process, branching model, and support governance for PrefixOS.

---

## 1. Versioning Scheme

PrefixOS strictly adheres to **Semantic Versioning 2.0.0 (SemVer)**:
`MAJOR.MINOR.PATCH`

- **MAJOR**: Incompatible API breaking changes or storage schema migrations.
- **MINOR**: Backward-compatible feature additions (e.g., new eviction policy, new engine adapter).
- **PATCH**: Backward-compatible bug fixes, security patches, and performance optimizations.

---

## 2. Release Milestones & Lifecycle

```
[ v1.0 MVP Core ] ──> [ v1.5 Enterprise Persistence ] ──> [ v2.0 Distributed Cluster ]
        │                            │                               │
        ▼                            ▼                               ▼
Slab Allocator, Shards,     WAL, Snapshots, Recovery,      Raft Consensus, Hash Ring,
Adaptive Policy, REST       Docker, Go/Python SDKs         Multi-Node Partition Handling
```

| Milestone | Target Scope | Tag Format | LTS Support |
| :--- | :--- | :--- | :--- |
| **v1.0.0** | MVP Core (Slab, Sharded RCU Tree, Adaptive Eviction, REST, Metrics) | `v1.0.0` | 12 Months |
| **v1.5.0** | Durability (WAL Engine, Snapshot, Parallel Recovery, SDKs) | `v1.5.0` | 12 Months |
| **v2.0.0** | Distributed (Raft Consensus, Memberlist, Hash Ring Rebalancing) | `v2.0.0` | 18 Months (LTS) |
| **v2.5.0** | Adapters (vLLM, SGLang, llama.cpp Native IPC Adapters) | `v2.5.0` | 12 Months |
| **v3.0.0** | Hardware Acceleration (RDMA Transport, GPU Direct Storage) | `v3.0.0` | 24 Months (LTS) |

---

## 3. Git Branching Strategy

- `main`: Contains production-ready code. All commits represent tested, release candidate code.
- `develop`: Primary integration branch for upcoming minor releases.
- `feature/<name>`: Topic branches for new capabilities (must pass CI/CD before PR merge).
- `hotfix/<name>`: Emergency patch branches targeting active release tags.
