package interfaces

// AllocatorStats contains summary metrics for the Slab Allocator
type AllocatorStats struct {
	TotalBlocks     int64
	AllocatedBlocks int64
	FreeBlocks      int64
	Fragmentation   float64
}

// MemoryAllocator defines the contract for zero-allocation slab memory block management
type MemoryAllocator interface {
	Allocate(count int) ([]int32, error)
	Free(blockIDs []int32) error
	FragmentationRatio() float64
	Stats() AllocatorStats
}

// NUMAManager defines NUMA node binding operations
type NUMAManager interface {
	BindNode(nodeID int) error
	GetNodeForBlock(blockID int32) int
}

// Defragmenter defines memory compaction operations
type Defragmenter interface {
	Compact() (freedBlocks int, err error)
}
