package radix

import (
	"testing"

	"prefixos/internal/memory"
)

func TestRadixTree_BasicInsertAndMatch(t *testing.T) {
	bm := memory.NewBlockManager(1000)
	tree := NewTree(bm)

	tokens := []int32{101, 102, 103, 104, 105}
	ok, blocks := tree.InsertTokens(tokens)
	if !ok || len(blocks) == 0 {
		t.Fatalf("InsertTokens failed, ok: %v, blocks: %v", ok, blocks)
	}

	matchedLen, matchedBlocks := tree.MatchPrefix(tokens)
	if matchedLen != 5 {
		t.Fatalf("expected matched length 5, got %d", matchedLen)
	}
	if len(matchedBlocks) == 0 {
		t.Fatalf("expected matched blocks, got empty")
	}
}

func TestRadixTree_PrefixBranching(t *testing.T) {
	bm := memory.NewBlockManager(1000)
	tree := NewTree(bm)

	promptA := []int32{1, 2, 3, 4, 10}
	promptB := []int32{1, 2, 3, 4, 20}

	tree.InsertTokens(promptA)
	tree.InsertTokens(promptB)

	matchedLen, _ := tree.MatchPrefix([]int32{1, 2, 3, 4, 30})
	if matchedLen != 4 {
		t.Fatalf("expected common prefix matched length 4, got %d", matchedLen)
	}
}

func TestRadixTree_Iterator(t *testing.T) {
	bm := memory.NewBlockManager(1000)
	tree := NewTree(bm)

	tree.InsertTokens([]int32{10, 20, 30})
	it := tree.Iterator()

	count := 0
	for it.HasNext() {
		toks, _ := it.Next()
		if len(toks) > 0 {
			count++
		}
	}

	if count == 0 {
		t.Fatalf("expected non-zero iterator nodes")
	}
}
