package memory

import (
	"errors"
	"sync"
	"sync/atomic"

	"prefixos/internal/interfaces"
)

var (
	ErrOutOfMemory = errors.New("out of memory blocks")
	ErrInvalidID   = errors.New("invalid memory block ID")
)

// BlockManager acts as our production Slab Allocator, handling zero-allocation
// pooling of fixed-size Blocks and tracking fragmentation/NUMA node bindings.
type BlockManager struct {
	mu          sync.Mutex
	pool        []*Block
	freeIDs     []int32
	allocated   map[int32]bool
	capacity    int64
	numaNodeID  int
	allocCount  int64
}

// NewBlockManager initializes a manager with a pre-allocated pool of blocks.
func NewBlockManager(capacity int) *BlockManager {
	pool := make([]*Block, capacity)
	freeIDs := make([]int32, capacity)
	allocated := make(map[int32]bool, capacity)

	for i := 0; i < capacity; i++ {
		id := int32(i)
		pool[i] = &Block{
			ID:           i,
			MemoryOffset: int64(i) * BlockSize * 4, // simulated memory offset in bytes
		}
		freeIDs[i] = id
	}

	return &BlockManager{
		pool:       pool,
		freeIDs:    freeIDs,
		allocated:  allocated,
		capacity:   int64(capacity),
		numaNodeID: 0,
	}
}

// Ensure BlockManager implements MemoryAllocator, NUMAManager, Defragmenter
var (
	_ interfaces.MemoryAllocator = (*BlockManager)(nil)
	_ interfaces.NUMAManager     = (*BlockManager)(nil)
	_ interfaces.Defragmenter     = (*BlockManager)(nil)
)

// Allocate requests 'count' blocks from the pre-allocated pool.
func (m *BlockManager) Allocate(count int) ([]int32, error) {
	if count <= 0 {
		return nil, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.freeIDs) < count {
		return nil, ErrOutOfMemory
	}

	result := make([]int32, count)
	for i := 0; i < count; i++ {
		id := m.freeIDs[len(m.freeIDs)-1]
		m.freeIDs = m.freeIDs[:len(m.freeIDs)-1]

		m.allocated[id] = true
		block := m.pool[id]

		// Zero out tokens
		for t := range block.Tokens {
			block.Tokens[t] = 0
		}

		result[i] = id
		atomic.AddInt64(&m.allocCount, 1)
	}

	return result, nil
}

// Free returns a list of blockIDs back to the free pool.
func (m *BlockManager) Free(blockIDs []int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, id := range blockIDs {
		if id < 0 || int(id) >= len(m.pool) {
			return ErrInvalidID
		}

		if m.allocated[id] {
			delete(m.allocated, id)
			m.freeIDs = append(m.freeIDs, id)
			atomic.AddInt64(&m.allocCount, -1)
		}
	}

	return nil
}

// Single block helper methods for internal direct tree usage
func (m *BlockManager) AllocateBlock() (*Block, error) {
	ids, err := m.Allocate(1)
	if err != nil {
		return nil, err
	}
	return m.pool[ids[0]], nil
}

func (m *BlockManager) FreeBlock(id int) {
	_ = m.Free([]int32{int32(id)})
}

func (m *BlockManager) GetBlock(id int32) (*Block, error) {
	if id < 0 || int(id) >= len(m.pool) {
		return nil, ErrInvalidID
	}
	return m.pool[id], nil
}

// FragmentationRatio calculates the ratio of non-contiguous free blocks.
func (m *BlockManager) FragmentationRatio() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.fragmentationRatioLocked()
}

func (m *BlockManager) fragmentationRatioLocked() float64 {
	if len(m.freeIDs) <= 1 {
		return 0.0
	}

	gaps := 0
	for i := 0; i < len(m.freeIDs)-1; i++ {
		if m.freeIDs[i+1] != m.freeIDs[i]+1 && m.freeIDs[i+1] != m.freeIDs[i]-1 {
			gaps++
		}
	}

	return float64(gaps) / float64(len(m.freeIDs))
}

// Stats returns a snapshot summary of memory allocation metrics.
func (m *BlockManager) Stats() interfaces.AllocatorStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	allocated := int64(len(m.allocated))
	free := m.capacity - allocated

	return interfaces.AllocatorStats{
		TotalBlocks:     m.capacity,
		AllocatedBlocks: allocated,
		FreeBlocks:      free,
		Fragmentation:   m.fragmentationRatioLocked(),
	}
}

// BindNode sets NUMA node affinity for memory allocations.
func (m *BlockManager) BindNode(nodeID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.numaNodeID = nodeID
	return nil
}

// GetNodeForBlock returns the assigned NUMA node ID for a given block.
func (m *BlockManager) GetNodeForBlock(blockID int32) int {
	return m.numaNodeID
}

// Compact defragmentates the free pool ordering for contiguous allocations.
func (m *BlockManager) Compact() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Sort free list by ID ascending to maximize contiguous block reuse
	freed := 0
	for i := 0; i < len(m.freeIDs); i++ {
		for j := i + 1; j < len(m.freeIDs); j++ {
			if m.freeIDs[i] > m.freeIDs[j] {
				m.freeIDs[i], m.freeIDs[j] = m.freeIDs[j], m.freeIDs[i]
				freed++
			}
		}
	}
	return freed, nil
}
