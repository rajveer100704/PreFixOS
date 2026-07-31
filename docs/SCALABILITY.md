# Scalability & Multi-Node Cluster Sizing - PrefixOS

This document details horizontal scaling dynamics, shard rebalancing, and cluster throughput scaling in PrefixOS.

---

## 1. Horizontal Scaling Dynamics

PrefixOS scales read throughput linearly with node count:

$$\text{ClusterThroughput} = N_{\text{nodes}} \times \text{Throughput}_{\text{single\_node}} \times (1 - \text{ReplicationOverhead})$$

- **Read Scaling**: Adding follower nodes increases read throughput linearly (`> 15M ops/sec` across 3 nodes).
- **Write Scaling**: Write sharding routes token prefix branches across the 64-way Consistent Hash Ring.
