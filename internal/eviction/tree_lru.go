package eviction

import (
	"container/heap"
	"sync"
	"time"

	"radixkv/internal/memory"
	"radixkv/internal/radix"
)

// LeafQueue implements heap.Interface and holds RadixNodes.
// It acts as an O(1) Min-Heap Priority Queue prioritizing the oldest LastUsed timestamps.
type LeafQueue []*radix.RadixNode

func (q LeafQueue) Len() int { return len(q) }

func (q LeafQueue) Less(i, j int) bool {
	return q[i].LastUsed.Before(q[j].LastUsed)
}

func (q LeafQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *LeafQueue) Push(x interface{}) {
	node := x.(*radix.RadixNode)
	*q = append(*q, node)
}

func (q *LeafQueue) Pop() interface{} {
	old := *q
	n := len(old)
	node := old[n-1]
	old[n-1] = nil // Avoid memory leak
	*q = old[0 : n-1]
	return node
}

// TreeLRU implements an O(1) Tree-Aware LRU eviction policy.
// It exclusively targets leaf nodes, preserving valuable shared prefixes.
type TreeLRU struct {
	mu    sync.Mutex
	queue LeafQueue
	tree  *radix.Tree
	bm    *memory.BlockManager
}

// NewTreeLRU initializes a new Tree-Aware LRU garbage collector with a Min-Heap.
func NewTreeLRU(tree *radix.Tree, bm *memory.BlockManager) *TreeLRU {
	lru := &TreeLRU{
		queue: make(LeafQueue, 0),
		tree:  tree,
		bm:    bm,
	}
	heap.Init(&lru.queue)
	return lru
}

// MarkAsLeaf adds a newly created or transitioned leaf node to the eviction queue.
// The Radix Tree should call this when appending a new leaf or truncating a node.
func (l *TreeLRU) MarkAsLeaf(node *radix.RadixNode) {
	l.mu.Lock()
	defer l.mu.Unlock()
	heap.Push(&l.queue, node)
}

// UpdateUsage adjusts the priority of a leaf node if its LastUsed time changes.
// To keep concurrency simple, we do a quick linear scan to fix the heap.
func (l *TreeLRU) UpdateUsage(node *radix.RadixNode) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, n := range l.queue {
		if n == node {
			heap.Fix(&l.queue, i)
			return
		}
	}
}

// Evict pops the oldest node from the O(1) Min-Heap and evicts it.
// If evicting a leaf leaves its parent with no children, the parent becomes a leaf
// and is dynamically pushed into the Min-Heap.
func (l *TreeLRU) Evict() bool {
	var oldest *radix.RadixNode

	// Loop to pop a valid leaf candidate without holding the global lock
	for {
		l.mu.Lock()
		if l.queue.Len() == 0 {
			l.mu.Unlock()
			return false // Nothing to evict
		}
		oldest = heap.Pop(&l.queue).(*radix.RadixNode)
		l.mu.Unlock()

		// Edge case: Node might have gained children after being pushed.
		oldest.mu.RLock()
		isLeaf := len(oldest.Children) == 0
		oldest.mu.RUnlock()

		if isLeaf {
			break
		}
		// Discard the invalid leaf and keep searching
	}

	// Safely lock parent and child in top-down order to prevent deadlocks
	var parent *radix.RadixNode
	for {
		parent = oldest.Parent
		if parent == nil {
			return false // Root cannot be evicted
		}
		
		parent.mu.Lock()
		oldest.mu.Lock()
		
		if oldest.Parent == parent {
			break // Safe to proceed
		}
		
		// Parent changed due to concurrent node split, unlock and retry
		oldest.mu.Unlock()
		parent.mu.Unlock()
	}

	defer parent.mu.Unlock()
	defer oldest.mu.Unlock()

	// Final verification under write lock
	if len(oldest.Children) > 0 {
		return false // Abort eviction, already discarded from heap
	}

	// Delete the oldest node from its parent
	var deletedKey int32
	found := false
	for k, v := range parent.Children {
		if v == oldest {
			deletedKey = k
			found = true
			break
		}
	}

	if found {
		delete(parent.Children, deletedKey)
	}

	// If parent is now childless, it becomes a leaf!
	if len(parent.Children) == 0 {
		parent.IsLeaf = true
		l.MarkAsLeaf(parent)
	}

	// Free VRAM metadata pointers
	for _, blockID := range oldest.BlockIDs {
		l.bm.Free(blockID)
	}

	return true
}

// RunBackgroundGarbageCollector runs a background loop to free memory when triggered.
func (l *TreeLRU) RunBackgroundGarbageCollector(interval time.Duration, trigger func() bool) {
	go func() {
		for {
			time.Sleep(interval)
			for trigger() {
				if !l.Evict() {
					break
				}
			}
		}
	}()
}
