package radix

import (
	"sync"
	"time"
)

// RadixNode represents a node in the 64-way sharded Radix Tree.
type RadixNode struct {
	mu sync.RWMutex

	ID       uint64
	Tokens   []int32
	BlockIDs []int32
	Children map[int32]*RadixNode
	IsLeaf   bool
	LastUsed time.Time
	Parent   *RadixNode
}

// NewRadixNode creates a new initialized RadixNode.
func NewRadixNode(id uint64, tokens []int32, blockIDs []int32, parent *RadixNode) *RadixNode {
	return &RadixNode{
		ID:       id,
		Tokens:   tokens,
		BlockIDs: blockIDs,
		Children: make(map[int32]*RadixNode),
		IsLeaf:   true,
		LastUsed: time.Now(),
		Parent:   parent,
	}
}

func (n *RadixNode) Lock()    { n.mu.Lock() }
func (n *RadixNode) Unlock()  { n.mu.Unlock() }
func (n *RadixNode) RLock()   { n.mu.RLock() }
func (n *RadixNode) RUnlock() { n.mu.RUnlock() }
