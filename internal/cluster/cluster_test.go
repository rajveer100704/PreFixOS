package cluster

import (
	"context"
	"testing"
)

func TestCluster_HashRingAndRouting(t *testing.T) {
	ring := NewHashRing(10)

	// Add shards
	ring.AddShard("shard-a")
	ring.AddShard("shard-b")
	ring.AddShard("shard-c")

	// Map key tokens
	tokens1 := []int32{10, 20, 30}
	shard1, err := ring.GetShard(tokens1)
	if err != nil || shard1 == "" {
		t.Fatalf("failed routing tokens1: %v", err)
	}

	tokens2 := []int32{99, 88, 77}
	shard2, err := ring.GetShard(tokens2)
	if err != nil || shard2 == "" {
		t.Fatalf("failed routing tokens2: %v", err)
	}

	// Router leader resolution
	router := NewRouter(ring)
	router.RegisterShardLeader("shard-a", "127.0.0.1:50051")
	router.RegisterShardLeader("shard-b", "127.0.0.1:50052")
	router.RegisterShardLeader("shard-c", "127.0.0.1:50053")

	ctx := context.Background()
	targetShard, leaderAddr, err := router.RouteRequest(ctx, tokens1)
	if err != nil {
		t.Fatalf("router.RouteRequest failed: %v", err)
	}
	if targetShard == "" || leaderAddr == "" {
		t.Fatalf("expected valid shard and leader address, got shard=%s, leader=%s", targetShard, leaderAddr)
	}
}

func TestCluster_ShardManagerMultiRaft(t *testing.T) {
	dataDir := t.TempDir()
	sm := NewShardManager(dataDir)
	defer sm.Close()

	// Create Multi-Raft shards
	nodeA, err := sm.CreateShard("shard-a", 1, []int64{1, 2})
	if err != nil || nodeA == nil {
		t.Fatalf("failed creating shard-a: %v", err)
	}

	nodeB, err := sm.CreateShard("shard-b", 1, []int64{1, 2})
	if err != nil || nodeB == nil {
		t.Fatalf("failed creating shard-b: %v", err)
	}

	retrieved, err := sm.GetShard("shard-a")
	if err != nil || retrieved == nil {
		t.Fatalf("failed retrieving shard-a: %v", err)
	}
}
