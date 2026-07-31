# Failure Modes & Fault Taxonomy - PrefixOS

This document catalogs failure modes, network partition behaviors, and split-brain resolution protocols.

---

## 1. Fault Taxonomy Matrix

| Failure Mode | Detection Mechanism | System Behavior | Recovery Action |
| :--- | :--- | :--- | :--- |
| **Follower Crash** | Raft Heartbeat Timeout (`> 150ms`) | Cluster continues operating with $N-1$ nodes | Node restarts and syncs missing log entries |
| **Leader Crash** | Heartbeat Missed | Follower initiates election after `150–300ms` randomized timeout | New leader elected within `< 1.5s` |
| **Network Partition** | Raft Quorum Check | Majority partition continues operating; Minority partition enters Read-Only state | Heals automatically when partition resolves |
| **Disk Write Out-of-Space** | OS I/O Error | Rejects new `Insert` requests, continues serving `MatchPrefix` reads | Free disk space or rotate WAL logs |
