# ADR-004: gRPC-Based Raft Consensus Engine for Multi-Node Replication

**Status**: Accepted  
**Date**: 2026-07-31  
**Deciders**: Distributed Systems Lead  

---

## Context & Problem Statement
Single-node cache engines present a single point of failure (SPOF) and cannot scale across enterprise LLM serving clusters.

---

## Decision Outcome
**Chosen Option**: Integrate a gRPC-based **Raft Consensus Engine** (`internal/replication/raft.go`) paired with a Consistent Hash Ring for leader election, multi-node log replication, and partition resilience.

### Consequences
- **Positive**: Guaranteed linearizable log replication, automated leader failover in `< 1.5s`, and high availability.
- **Negative**: Adds network consensus latency on write/insert operations.
