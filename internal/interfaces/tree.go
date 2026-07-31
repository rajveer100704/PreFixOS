package interfaces

// Iterator defines thread-safe traversal over Radix tree nodes
type Iterator interface {
	HasNext() bool
	Next() (tokens []int32, blockIDs []int32)
}

// RadixTreeEngine defines the contract for 64-way sharded RCU prefix matching & storage
type RadixTreeEngine interface {
	MatchPrefix(tokens []int32) (matchedLength int, blockIDs []int32)
	Insert(tokens []int32, blockIDs []int32) error
	EvictLeaf(nodeID uint64) (freedBlocks []int32, err error)
	Iterator() Iterator
}
