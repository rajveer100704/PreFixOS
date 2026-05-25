package memory

import (
	"errors"
	"sync"
)

var ErrOutOfMemory = errors.New("out of memory blocks")

// BlockManager acts as our VRAM simulator, handling the allocation
// and freeing of fixed-size Blocks.
type BlockManager struct {
	mu      sync.Mutex
	pool    []*Block
	freeIDs []int
}

// NewBlockManager initializes a manager with a pre-allocated set of blocks.
func NewBlockManager(capacity int) *BlockManager {
	pool := make([]*Block, capacity)
	freeIDs := make([]int, capacity)
	for i := 0; i < capacity; i++ {
		pool[i] = &Block{ID: i}
		freeIDs[i] = i
	}
	return &BlockManager{
		pool:    pool,
		freeIDs: freeIDs,
	}
}

// Allocate requests a single block from the pre-allocated pool.
func (m *BlockManager) Allocate() (*Block, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.freeIDs) == 0 {
		return nil, ErrOutOfMemory
	}

	// Pop a free block ID
	id := m.freeIDs[len(m.freeIDs)-1]
	m.freeIDs = m.freeIDs[:len(m.freeIDs)-1]

	block := m.pool[id]

	// Zero out tokens to avoid leaking previous context
	for i := range block.Tokens {
		block.Tokens[i] = 0
	}

	return block, nil
}

// Free returns a block to the manager's free list.
func (m *BlockManager) Free(id int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Push the ID back to the free list
	m.freeIDs = append(m.freeIDs, id)
}
