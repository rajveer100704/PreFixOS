# ADR-003: Pluggable Eviction Framework & Adaptive Cost-Aware Scoring

**Status**: Accepted  
**Date**: 2026-07-31  
**Deciders**: ML Infrastructure Lead  

---

## Context & Problem Statement
Standard LRU eviction policies evict shared root system prompts when leaf requests burst, causing severe cache trashing across all dependent agent requests.

---

## Decision Outcome
**Chosen Option**: Define a pluggable `EvictionPolicy` interface (`internal/interfaces/eviction.go`) with an **Adaptive Cost-Aware Eviction** default algorithm that computes multi-factor priority scores ($\frac{\text{HitCount} \times \text{Depth} \times \text{Cost}}{\text{Age} + 1}$) and exclusively targets childless leaf nodes bottom-up.

### Consequences
- **Positive**: Completely prevents root system prompt eviction. Preserves shared foundations and maximizes cache hit ratio.
- **Negative**: Requires maintaining leaf tracking queues and scoring metadata.
