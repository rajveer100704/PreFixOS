# Capacity Planning & Sizing Guide - PrefixOS

This document provides capacity planning guidelines, formulas, and memory consumption estimates for deploying PrefixOS in production.

---

## 1. Memory Calculation Formula

PrefixOS maintains Metadata (Block IDs, Node Pointers, Version Counters) in RAM, while physical attention tensors reside in GPU VRAM or shared IPC memory.

$$\text{Total RAM (Bytes)} = (\text{NumNodes} \times S_{\text{node}}) + (\text{NumBlocks} \times S_{\text{block\_meta}}) + \text{SlabPoolSize} + \text{WALBuffer}$$

Where:
- $S_{\text{node}} \approx 128 \text{ Bytes}$ (Radix Node metadata + 64-way shard version)
- $S_{\text{block\_meta}} \approx 32 \text{ Bytes}$ (Block ID, Offset, Reference Count)
- $\text{SlabPoolSize} = \text{Preallocated Slab Memory}$

---

## 2. Resource Estimation Matrix

| Token Count | Number of Nodes | RAM Needed (Control Plane) | VRAM Saved (Data Plane) | Recommended Shards | Recommended Nodes |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **100,000 Tokens** | ~5,000 | ~16 MB | ~400 MB | 16 Shards | 1 Node |
| **1,000,000 Tokens** | ~50,000 | ~120 MB | ~4.0 GB | 32 Shards | 1–3 Nodes |
| **10,000,000 Tokens** | ~500,000 | ~1.1 GB | ~40.0 GB | 64 Shards | 3 Nodes |
| **100,000,000 Tokens**| ~5,000,000 | ~10.8 GB | ~400.0 GB | 128 Shards | 5 Nodes |

---

## 3. Shard & Memory Limit Recommendations

- **Default Slab Size**: 1,000,000 blocks (each block tracking 16 tokens).
- **Max Tokens Per Node**: 50,000,000 tokens per Go process.
- **CPU Cores**: 4 vCPUs per 10M token footprint to maintain sub-50µs lookup latencies.
