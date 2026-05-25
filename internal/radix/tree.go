package radix

import (
	"sync"
	"time"

	"radixkv/internal/memory"
)

// Tree represents the prefix-aware Radix Tree for sharing tokens.
type Tree struct {
	Root *RadixNode
	mu   sync.RWMutex // Global tree lock for broader structural consistency if needed
	bm   *memory.BlockManager
}

// NewTree initializes a new Radix Tree with a root node.
func NewTree(bm *memory.BlockManager) *Tree {
	return &Tree{
		Root: NewRadixNode([]int32{}, []int{}, nil),
		bm:   bm,
	}
}

// allocateBlocksAndCopy allocates blocks from the manager and copies tokens into them.
func (t *Tree) allocateBlocksAndCopy(tokens []int32) ([]int, error) {
	if len(tokens) == 0 {
		return nil, nil
	}

	numBlocks := (len(tokens) + memory.BlockSize - 1) / memory.BlockSize
	blocks := make([]int, 0, numBlocks)

	for i := 0; i < numBlocks; i++ {
		b, err := t.bm.Allocate()
		if err != nil {
			// Rollback allocated blocks on OOM
			for _, allocated := range blocks {
				t.bm.Free(allocated)
			}
			return nil, err
		}
		blocks = append(blocks, b.ID)

		start := i * memory.BlockSize
		end := start + memory.BlockSize
		if end > len(tokens) {
			end = len(tokens)
		}

		copy(b.Tokens[:], tokens[start:end])
	}

	return blocks, nil
}

// FindLongestPrefix searches the tree for the longest matching prefix of tokens.
// Returns the matched length and the accumulated BlockIDs.
func (t *Tree) FindLongestPrefix(tokens []int32) (int, []int) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	curr := t.Root
	matchedLen := 0
	var blockIDs []int

	for matchedLen < len(tokens) {
		curr.mu.RLock()
		targetToken := tokens[matchedLen]
		child, exists := curr.Children[targetToken]
		curr.mu.RUnlock()

		if !exists {
			break
		}

		child.mu.Lock()
		i := 0
		for i < len(child.Tokens) && matchedLen+i < len(tokens) {
			if child.Tokens[i] != tokens[matchedLen+i] {
				break
			}
			i++
		}

		if i > 0 {
			blockIDs = append(blockIDs, child.BlockIDs...)
		}

		// Update LRU timestamp since this node was accessed
		child.LastUsed = time.Now()
		
		matchedLen += i
		child.mu.Unlock()

		if i < len(child.Tokens) {
			// Diverged inside the child node
			break
		}
		
		curr = child
	}

	return matchedLen, blockIDs
}

// Insert adds a sequence of tokens into the tree, splitting nodes if necessary.
// Uses lock-coupling to allow concurrent reads while writes are isolated to localized branches.
// Returns success (false if OOM) and a list of newly allocated BlockIDs.
func (t *Tree) Insert(tokens []int32) (bool, []int) {
	if len(tokens) == 0 {
		return true, nil
	}

	t.mu.RLock() // Protect root pointer
	defer t.mu.RUnlock()

	// Lock-coupling: start with root
	t.Root.mu.Lock()
	curr := t.Root
	matchedLen := 0
	var allocatedBlocks []int

	for matchedLen < len(tokens) {
		targetToken := tokens[matchedLen]
		child, exists := curr.Children[targetToken]

		if !exists {
			// Fast path: append new leaf
			unmatched := tokens[matchedLen:]
			blocks, err := t.allocateBlocksAndCopy(unmatched)
			if err != nil {
				curr.mu.Unlock()
				return false, nil // OOM
			}

			newNode := NewRadixNode(unmatched, blocks, curr)
			curr.Children[targetToken] = newNode
			curr.IsLeaf = false
			allocatedBlocks = append(allocatedBlocks, blocks...)
			
			curr.mu.Unlock()
			return true, allocatedBlocks
		}

		// Child exists, lock it before unlocking parent to ensure atomic traversal
		child.mu.Lock()

		// Find common prefix length
		i := 0
		for i < len(child.Tokens) && matchedLen+i < len(tokens) {
			if child.Tokens[i] != tokens[matchedLen+i] {
				break
			}
			i++
		}

		if i == len(child.Tokens) {
			// Matched entire child node, continue traversal
			curr.mu.Unlock() // unlock parent
			matchedLen += i
			curr = child
			continue
		}

		// We need to split the child node
		grandchildTokens := make([]int32, len(child.Tokens)-i)
		copy(grandchildTokens, child.Tokens[i:])

		grandchildBlocks, err := t.allocateBlocksAndCopy(grandchildTokens)
		if err != nil {
			child.mu.Unlock()
			curr.mu.Unlock()
			return false, nil // OOM
		}

		// Create grandchild
		grandchild := NewRadixNode(grandchildTokens, grandchildBlocks, child)
		grandchild.Children = child.Children
		for _, gcChild := range grandchild.Children {
			gcChild.Parent = grandchild // update parent pointers
		}
		grandchild.IsLeaf = child.IsLeaf
		grandchild.LastUsed = child.LastUsed // Preserve LRU state

		// Truncate the child
		child.Tokens = child.Tokens[:i]
		numKeptBlocks := (i + memory.BlockSize - 1) / memory.BlockSize
		// Free the unused blocks that belonged to the suffix of the child
		for j := numKeptBlocks; j < len(child.BlockIDs); j++ {
			t.bm.Free(child.BlockIDs[j])
		}
		child.BlockIDs = child.BlockIDs[:numKeptBlocks]
		
		// Reset child's children
		child.Children = make(map[int32]*RadixNode)
		child.Children[grandchildTokens[0]] = grandchild
		child.IsLeaf = false
		child.LastUsed = time.Now()

		// Insert the new tokens if any
		unmatched := tokens[matchedLen+i:]
		if len(unmatched) > 0 {
			newChildBlocks, err := t.allocateBlocksAndCopy(unmatched)
			if err != nil {
				// OOM. Tree split was successful but failed to append new tokens.
				child.mu.Unlock()
				curr.mu.Unlock()
				return false, nil
			}

			newChild := NewRadixNode(unmatched, newChildBlocks, child)
			child.Children[unmatched[0]] = newChild
			allocatedBlocks = append(allocatedBlocks, newChildBlocks...)
		}

		child.mu.Unlock()
		curr.mu.Unlock()
		return true, allocatedBlocks
	}

	curr.mu.Unlock()
	return true, allocatedBlocks
}
