package radix

import (
	"sync"
	"sync/atomic"
	"time"

	"prefixos/internal/interfaces"
	"prefixos/internal/memory"
)

// Tree represents a thread-safe Prefix Radix Tree.
type Tree struct {
	Root       *RadixNode
	mu         sync.RWMutex
	bm         *memory.BlockManager
	nodeIDSeq  uint64
}

// Ensure Tree implements interfaces.RadixTreeEngine
var _ interfaces.RadixTreeEngine = (*Tree)(nil)

// NewTree initializes a new Radix Tree.
func NewTree(bm *memory.BlockManager) *Tree {
	t := &Tree{
		bm: bm,
	}
	t.Root = NewRadixNode(t.nextNodeID(), []int32{}, []int32{}, nil)
	return t
}

func (t *Tree) nextNodeID() uint64 {
	return atomic.AddUint64(&t.nodeIDSeq, 1)
}

// allocateBlocksAndCopy allocates blocks from the manager and copies tokens into them.
func (t *Tree) allocateBlocksAndCopy(tokens []int32) ([]int32, error) {
	if len(tokens) == 0 {
		return nil, nil
	}

	numBlocks := (len(tokens) + memory.BlockSize - 1) / memory.BlockSize
	blockIDs, err := t.bm.Allocate(numBlocks)
	if err != nil {
		return nil, err
	}

	for i, id := range blockIDs {
		b, err := t.bm.GetBlock(id)
		if err != nil {
			_ = t.bm.Free(blockIDs)
			return nil, err
		}

		start := i * memory.BlockSize
		end := start + memory.BlockSize
		if end > len(tokens) {
			end = len(tokens)
		}

		copy(b.Tokens[:], tokens[start:end])
	}

	return blockIDs, nil
}

// FindLongestPrefix searches the tree for the longest matching prefix of tokens.
func (t *Tree) FindLongestPrefix(tokens []int32) (int, []int32) {
	return t.MatchPrefix(tokens)
}

// MatchPrefix implements interfaces.RadixTreeEngine
func (t *Tree) MatchPrefix(tokens []int32) (int, []int32) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	curr := t.Root
	matchedLen := 0
	var matchedBlocks []int32

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
			matchedBlocks = append(matchedBlocks, child.BlockIDs...)
		}

		child.LastUsed = time.Now()
		matchedLen += i
		child.mu.Unlock()

		if i < len(child.Tokens) {
			// Diverged inside child node
			break
		}

		curr = child
	}

	return matchedLen, matchedBlocks
}

// Insert implements interfaces.RadixTreeEngine
func (t *Tree) Insert(tokens []int32, blockIDs []int32) error {
	ok, _ := t.InsertTokens(tokens)
	if !ok {
		return memory.ErrOutOfMemory
	}
	return nil
}

// InsertTokens adds a sequence of tokens into the tree, allocating blocks as needed.
func (t *Tree) InsertTokens(tokens []int32) (bool, []int32) {
	if len(tokens) == 0 {
		return true, nil
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	t.Root.mu.Lock()
	curr := t.Root
	matchedLen := 0
	var allocatedBlocks []int32

	for matchedLen < len(tokens) {
		targetToken := tokens[matchedLen]
		child, exists := curr.Children[targetToken]

		if !exists {
			unmatched := tokens[matchedLen:]
			blocks, err := t.allocateBlocksAndCopy(unmatched)
			if err != nil {
				curr.mu.Unlock()
				return false, nil
			}

			newNode := NewRadixNode(t.nextNodeID(), unmatched, blocks, curr)
			curr.Children[targetToken] = newNode
			curr.IsLeaf = false
			allocatedBlocks = append(allocatedBlocks, blocks...)

			curr.mu.Unlock()
			return true, allocatedBlocks
		}

		child.mu.Lock()

		i := 0
		for i < len(child.Tokens) && matchedLen+i < len(tokens) {
			if child.Tokens[i] != tokens[matchedLen+i] {
				break
			}
			i++
		}

		if i == len(child.Tokens) {
			curr.mu.Unlock()
			matchedLen += i
			curr = child
			continue
		}

		// Split child node
		grandchildTokens := make([]int32, len(child.Tokens)-i)
		copy(grandchildTokens, child.Tokens[i:])

		grandchildBlocks, err := t.allocateBlocksAndCopy(grandchildTokens)
		if err != nil {
			child.mu.Unlock()
			curr.mu.Unlock()
			return false, nil
		}

		grandchild := NewRadixNode(t.nextNodeID(), grandchildTokens, grandchildBlocks, child)
		grandchild.Children = child.Children
		for _, gcChild := range grandchild.Children {
			gcChild.Parent = grandchild
		}
		grandchild.IsLeaf = child.IsLeaf
		grandchild.LastUsed = child.LastUsed

		child.Tokens = child.Tokens[:i]
		numKeptBlocks := (i + memory.BlockSize - 1) / memory.BlockSize

		if numKeptBlocks < len(child.BlockIDs) {
			_ = t.bm.Free(child.BlockIDs[numKeptBlocks:])
			child.BlockIDs = child.BlockIDs[:numKeptBlocks]
		}

		child.Children = make(map[int32]*RadixNode)
		child.Children[grandchildTokens[0]] = grandchild
		child.IsLeaf = false
		child.LastUsed = time.Now()

		unmatched := tokens[matchedLen+i:]
		if len(unmatched) > 0 {
			newChildBlocks, err := t.allocateBlocksAndCopy(unmatched)
			if err != nil {
				child.mu.Unlock()
				curr.mu.Unlock()
				return false, nil
			}

			newChild := NewRadixNode(t.nextNodeID(), unmatched, newChildBlocks, child)
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

// EvictLeaf implements interfaces.RadixTreeEngine
func (t *Tree) EvictLeaf(nodeID uint64) ([]int32, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var target *RadixNode
	var findLeaf func(n *RadixNode)
	findLeaf = func(n *RadixNode) {
		if target != nil {
			return
		}
		if n.ID == nodeID {
			target = n
			return
		}
		for _, c := range n.Children {
			findLeaf(c)
		}
	}

	findLeaf(t.Root)
	if target == nil || target.Parent == nil {
		return nil, nil
	}

	freed := make([]int32, len(target.BlockIDs))
	copy(freed, target.BlockIDs)
	_ = t.bm.Free(freed)

	// Remove target from parent's children
	for k, v := range target.Parent.Children {
		if v.ID == nodeID {
			delete(target.Parent.Children, k)
			break
		}
	}
	if len(target.Parent.Children) == 0 {
		target.Parent.IsLeaf = true
	}

	return freed, nil
}

// TreeIterator implements interfaces.Iterator
type TreeIterator struct {
	nodes []*RadixNode
	index int
}

func (it *TreeIterator) HasNext() bool {
	return it.index < len(it.nodes)
}

func (it *TreeIterator) Next() ([]int32, []int32) {
	if !it.HasNext() {
		return nil, nil
	}
	n := it.nodes[it.index]
	it.index++
	return n.Tokens, n.BlockIDs
}

// Iterator implements interfaces.RadixTreeEngine
func (t *Tree) Iterator() interfaces.Iterator {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var nodes []*RadixNode
	var collect func(n *RadixNode)
	collect = func(n *RadixNode) {
		if n != t.Root {
			nodes = append(nodes, n)
		}
		for _, c := range n.Children {
			collect(c)
		}
	}
	collect(t.Root)

	return &TreeIterator{nodes: nodes, index: 0}
}
