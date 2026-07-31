package memory

import (
	"testing"
)

func TestBlockManager_AllocationAndFree(t *testing.T) {
	bm := NewBlockManager(100)

	// Initial stats
	stats := bm.Stats()
	if stats.TotalBlocks != 100 || stats.FreeBlocks != 100 || stats.AllocatedBlocks != 0 {
		t.Fatalf("unexpected initial stats: %+v", stats)
	}

	// Allocate 10 blocks
	ids, err := bm.Allocate(10)
	if err != nil {
		t.Fatalf("failed to allocate blocks: %v", err)
	}
	if len(ids) != 10 {
		t.Fatalf("expected 10 IDs, got %d", len(ids))
	}

	stats = bm.Stats()
	if stats.AllocatedBlocks != 10 || stats.FreeBlocks != 90 {
		t.Fatalf("expected 10 allocated, got %+v", stats)
	}

	// Free 5 blocks
	err = bm.Free(ids[:5])
	if err != nil {
		t.Fatalf("failed to free blocks: %v", err)
	}

	stats = bm.Stats()
	if stats.AllocatedBlocks != 5 || stats.FreeBlocks != 95 {
		t.Fatalf("expected 5 allocated, got %+v", stats)
	}
}

func TestBlockManager_OOM(t *testing.T) {
	bm := NewBlockManager(5)

	_, err := bm.Allocate(5)
	if err != nil {
		t.Fatalf("allocation of full capacity failed: %v", err)
	}

	// Try allocating 1 more block
	_, err = bm.Allocate(1)
	if err != ErrOutOfMemory {
		t.Fatalf("expected ErrOutOfMemory, got %v", err)
	}
}

func TestBlockManager_CompactAndNUMA(t *testing.T) {
	bm := NewBlockManager(50)

	if err := bm.BindNode(2); err != nil {
		t.Fatalf("BindNode failed: %v", err)
	}
	if bm.GetNodeForBlock(0) != 2 {
		t.Errorf("expected NUMA node 2, got %d", bm.GetNodeForBlock(0))
	}

	// Allocate and free in non-sequential order
	ids, err := bm.Allocate(10)
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}
	if err := bm.Free([]int32{ids[2], ids[5], ids[7]}); err != nil {
		t.Fatalf("Free failed: %v", err)
	}

	if _, err = bm.Compact(); err != nil {
		t.Fatalf("Compact failed: %v", err)
	}
}
