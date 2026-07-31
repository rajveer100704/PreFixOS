package eviction

import (
	"container/heap"
	"sync"
	"time"

	"prefixos/internal/interfaces"
	"prefixos/internal/memory"
	"prefixos/internal/radix"
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

var _ interfaces.EvictionPolicy = (*TreeLRU)(nil)

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

// SelectVictim selects the oldest leaf node ID for eviction.
func (l *TreeLRU) SelectVictim() (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.queue.Len() == 0 {
		return 0, ErrEmptyCache
	}
	node := heap.Pop(&l.queue).(*radix.RadixNode)
	return node.ID, nil
}

// OnInsert adds a newly created node to the TreeLRU queue.
func (l *TreeLRU) OnInsert(nodeID uint64, depth int, cost float64) {}

// OnAccess updates the usage timestamp of a node.
func (l *TreeLRU) OnAccess(nodeID uint64, hitCount uint64) {}

// OnDelete removes a node from the TreeLRU queue.
func (l *TreeLRU) OnDelete(nodeID uint64) {}

// MarkAsLeaf adds a newly created or transitioned leaf node to the eviction queue.
func (l *TreeLRU) MarkAsLeaf(node *radix.RadixNode) {
	l.mu.Lock()
	defer l.mu.Unlock()
	heap.Push(&l.queue, node)
}

// UpdateUsage adjusts the priority of a leaf node if its LastUsed time changes.
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
func (l *TreeLRU) Evict() bool {
	var oldest *radix.RadixNode

	for {
		l.mu.Lock()
		if l.queue.Len() == 0 {
			l.mu.Unlock()
			return false
		}
		oldest = heap.Pop(&l.queue).(*radix.RadixNode)
		l.mu.Unlock()

		oldest.RLock()
		isLeaf := len(oldest.Children) == 0
		oldest.RUnlock()

		if isLeaf {
			break
		}
	}

	var parent *radix.RadixNode
	for {
		parent = oldest.Parent
		if parent == nil {
			return false
		}

		parent.Lock()
		oldest.Lock()

		if oldest.Parent == parent {
			break
		}

		oldest.Unlock()
		parent.Unlock()
	}

	defer parent.Unlock()
	defer oldest.Unlock()

	if len(oldest.Children) > 0 {
		return false
	}

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

	if len(parent.Children) == 0 {
		parent.IsLeaf = true
		l.MarkAsLeaf(parent)
	}

	_ = l.bm.Free(oldest.BlockIDs)
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
