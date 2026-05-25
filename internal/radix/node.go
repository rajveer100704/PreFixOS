package radix

import (
	"sync"
	"time"
)

// RadixNode represents a node in the Radix Tree.
// It uses a reader-writer mutex to allow concurrent reads on shared paths
// while protecting writes during splits or appends.
type RadixNode struct {
	mu sync.RWMutex

	Tokens   []int32
	BlockIDs []int
	Children map[int32]*RadixNode
	IsLeaf   bool
	LastUsed time.Time
	Parent   *RadixNode
}

// NewRadixNode creates a new initialized RadixNode.
func NewRadixNode(tokens []int32, blockIDs []int, parent *RadixNode) *RadixNode {
	return &RadixNode{
		Tokens:   tokens,
		BlockIDs: blockIDs,
		Children: make(map[int32]*RadixNode),
		IsLeaf:   true,
		LastUsed: time.Now(),
		Parent:   parent,
	}
}
