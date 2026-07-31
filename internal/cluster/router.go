package cluster

import (
	"context"
	"fmt"
	"sync"
)

// ShardLeaderMap tracks current Raft leader addresses for each shard.
type ShardLeaderMap struct {
	mu      sync.RWMutex
	leaders map[string]string
}

// Router routes request payloads across Multi-Raft shards via Consistent Hashing.
type Router struct {
	ring      *HashRing
	leaderMap *ShardLeaderMap
}

// NewRouter creates a new cross-shard router.
func NewRouter(ring *HashRing) *Router {
	return &Router{
		ring: ring,
		leaderMap: &ShardLeaderMap{
			leaders: make(map[string]string),
		},
	}
}

// RegisterShardLeader updates the active leader address for a shard.
func (r *Router) RegisterShardLeader(shardID string, leaderAddr string) {
	r.leaderMap.mu.Lock()
	defer r.leaderMap.mu.Unlock()
	r.leaderMap.leaders[shardID] = leaderAddr
}

// RouteRequest determines the target Shard ID and Leader address for a given token sequence.
func (r *Router) RouteRequest(ctx context.Context, tokens []int32) (string, string, error) {
	shardID, err := r.ring.GetShard(tokens)
	if err != nil {
		return "", "", fmt.Errorf("failed routing key: %w", err)
	}

	r.leaderMap.mu.RLock()
	leaderAddr, exists := r.leaderMap.leaders[shardID]
	r.leaderMap.mu.RUnlock()

	if !exists {
		return shardID, "", fmt.Errorf("shard %s has no active leader", shardID)
	}

	return shardID, leaderAddr, nil
}
