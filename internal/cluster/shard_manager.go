package cluster

import (
	"fmt"
	"path/filepath"
	"sync"

	"prefixos/internal/replication"
)

// ShardManager manages multiple independent Raft consensus groups on a single node (Multi-Raft).
type ShardManager struct {
	mu      sync.RWMutex
	baseDir string
	shards  map[string]*replication.RaftNode
}

// NewShardManager initializes a Multi-Raft shard manager instance.
func NewShardManager(baseDir string) *ShardManager {
	return &ShardManager{
		baseDir: baseDir,
		shards:  make(map[string]*replication.RaftNode),
	}
}

// CreateShard initializes a new Raft consensus group for a specified shard ID.
func (sm *ShardManager) CreateShard(shardID string, nodeID int64, peers []int64) (*replication.RaftNode, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if node, exists := sm.shards[shardID]; exists {
		return node, nil
	}

	shardDir := filepath.Join(sm.baseDir, "shards", shardID)
	node, err := replication.NewRaftNode(nodeID, peers, shardDir)
	if err != nil {
		return nil, fmt.Errorf("failed initializing shard %s: %w", shardID, err)
	}

	sm.shards[shardID] = node
	return node, nil
}

// GetShard returns the RaftNode managing a given shard ID.
func (sm *ShardManager) GetShard(shardID string) (*replication.RaftNode, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	node, exists := sm.shards[shardID]
	if !exists {
		return nil, fmt.Errorf("shard %s not found", shardID)
	}
	return node, nil
}

// Close gracefully closes all active Raft shard instances.
func (sm *ShardManager) Close() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for id, node := range sm.shards {
		_ = node.Close()
		delete(sm.shards, id)
	}
}
