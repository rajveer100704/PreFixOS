package radix

import (
	"testing"

	"radixkv/internal/memory"
)

func TestTreeInsertAndFind(t *testing.T) {
	bm := memory.NewBlockManager(100) // 100 blocks = 1600 tokens
	tree := NewTree(bm)

	// Insert "A B C D E"
	seq1 := []int32{1, 2, 3, 4, 5}
	success, blocks1 := tree.Insert(seq1)
	if !success || len(blocks1) != 1 { // 5 tokens fit in 1 block
		t.Fatalf("Failed to insert seq1")
	}

	// FindLongestPrefix for "A B C"
	matchLen, blocks := tree.FindLongestPrefix([]int32{1, 2, 3})
	if matchLen != 3 {
		t.Errorf("Expected matchLen 3, got %d", matchLen)
	}

	// Insert "A B X Y" -> should split at "A B"
	seq2 := []int32{1, 2, 10, 11}
	success, blocks2 := tree.Insert(seq2)
	if !success {
		t.Fatalf("Failed to insert seq2")
	}
	if len(blocks2) != 1 { // "X Y" fits in 1 block
		t.Errorf("Expected 1 block allocated for seq2, got %d", len(blocks2))
	}

	// FindLongestPrefix for "A B X Y Z"
	matchLen, _ = tree.FindLongestPrefix([]int32{1, 2, 10, 11, 12})
	if matchLen != 4 {
		t.Errorf("Expected matchLen 4, got %d", matchLen)
	}
	
	// Ensure root has 1 child ("A B" prefix)
	if len(tree.Root.Children) != 1 {
		t.Errorf("Expected 1 child at root after split, got %d", len(tree.Root.Children))
	}
}
