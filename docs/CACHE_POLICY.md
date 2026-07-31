# Pluggable Eviction Architecture & Policies - PrefixOS

This document specifies the pluggable eviction framework and scoring policies supported by PrefixOS.

---

## 1. Pluggable `EvictionPolicy` Interface

PrefixOS decouples eviction logic via `internal/interfaces/eviction.go`:

```go
type EvictionPolicy interface {
    SelectVictim() (nodeID uint64, err error)
    OnAccess(nodeID uint64, hitCount uint64)
    OnInsert(nodeID uint64, depth int, cost float64)
    OnDelete(nodeID uint64)
}
```

---

## 2. Supported Eviction Policies

### 2.1 Adaptive Cost-Aware Policy (Default)
Standard LRU policies evict heavily shared root prompts if they haven't been accessed recently, causing catastrophic cache misses. The Adaptive Cost-Aware policy computes a priority score for every node:

$$\text{PriorityScore} = \frac{\text{HitCount} \times \text{Depth} \times \text{RebuildCost}}{\text{RecencyAge} + 1}$$

- **Targeting**: Only targets leaf nodes (`len(Children) == 0`).
- **Victim Selection**: The leaf node with the lowest `PriorityScore` is selected for eviction.

### 2.2 Adaptive Replacement Cache (ARC)
Dynamically balances between Recency (LRU) and Frequency (LFU) using two ghost queues ($L_1$ and $L_2$).

### 2.3 TinyLFU
Frequency-based eviction utilizing a Count-Min Sketch estimator to maintain frequency histograms in $O(1)$ space.

### 2.4 Standard LRU & LFU
Fallback standard Least-Recently-Used and Least-Frequently-Used eviction policies for comparison testing.
