package server

import (
	"path/filepath"
	"testing"

	"prefixos/internal/memory"
	"prefixos/internal/persistence"
	"prefixos/internal/radix"
)

func TestEndToEndDurabilityAndRecovery(t *testing.T) {
	dataDir := t.TempDir()

	// Phase 1: Initialize initial engine node
	bm1 := memory.NewBlockManager(1000)
	tree1 := radix.NewTree(bm1)
	pe1, err := persistence.NewEngine(dataDir, tree1, bm1, true)
	if err != nil {
		t.Fatalf("failed creating persistence engine: %v", err)
	}

	// Insert token prompt
	prompt1 := []int32{101, 102, 103, 104, 105}
	ok, blocks1 := tree1.InsertTokens(prompt1)
	if !ok || len(blocks1) == 0 {
		t.Fatalf("tree1 insert failed")
	}

	// Create snapshot
	snapID, err := pe1.CreateSnapshot()
	if err != nil || snapID == "" {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	pe1.Close()

	// Phase 2: Simulate node crash & restart state recovery
	bm2 := memory.NewBlockManager(1000)
	tree2 := radix.NewTree(bm2)
	snapMgr, err := persistence.NewSnapshotManager(filepath.Join(dataDir, "snapshots"), tree2, bm2)
	if err != nil {
		t.Fatalf("failed creating SnapshotManager: %v", err)
	}

	// Restore from snapshot
	if err := snapMgr.RestoreFromSnapshot(snapID); err != nil {
		t.Fatalf("RestoreFromSnapshot failed: %v", err)
	}

	// Verify matched prefix length after restart
	matchedLen, matchedBlocks := tree2.MatchPrefix(prompt1)
	if matchedLen != 5 {
		t.Fatalf("expected matched length 5 after restart, got %d", matchedLen)
	}
	if len(matchedBlocks) == 0 {
		t.Fatalf("expected physical block handles after restart, got empty")
	}
}
