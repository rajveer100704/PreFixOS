package eviction

import (
	"container/list"
	"errors"
	"sync"

	"prefixos/internal/interfaces"
)

var ErrEmptyCache = errors.New("eviction cache is empty")

type lruItem struct {
	nodeID   uint64
	depth    int
	cost     float64
	hitCount uint64
}

// LRUEvictionPolicy implements interfaces.EvictionPolicy using a thread-safe doubly-linked list & hash map.
type LRUEvictionPolicy struct {
	mu       sync.Mutex
	items    map[uint64]*list.Element
	evictList *list.List
}

var _ interfaces.EvictionPolicy = (*LRUEvictionPolicy)(nil)

// NewLRUEvictionPolicy creates a new LRUEvictionPolicy instance.
func NewLRUEvictionPolicy() *LRUEvictionPolicy {
	return &LRUEvictionPolicy{
		items:     make(map[uint64]*list.Element),
		evictList: list.New(),
	}
}

// OnInsert registers a node in the LRU eviction queue.
func (p *LRUEvictionPolicy) OnInsert(nodeID uint64, depth int, cost float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if elem, exists := p.items[nodeID]; exists {
		p.evictList.MoveToFront(elem)
		item := elem.Value.(*lruItem)
		item.depth = depth
		item.cost = cost
		return
	}

	item := &lruItem{
		nodeID: nodeID,
		depth:  depth,
		cost:   cost,
	}
	elem := p.evictList.PushFront(item)
	p.items[nodeID] = elem
}

// OnAccess moves the node to the front of the LRU queue upon access.
func (p *LRUEvictionPolicy) OnAccess(nodeID uint64, hitCount uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if elem, exists := p.items[nodeID]; exists {
		p.evictList.MoveToFront(elem)
		item := elem.Value.(*lruItem)
		item.hitCount += hitCount
	}
}

// OnDelete removes a node from the eviction tracking system.
func (p *LRUEvictionPolicy) OnDelete(nodeID uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if elem, exists := p.items[nodeID]; exists {
		p.evictList.Remove(elem)
		delete(p.items, nodeID)
	}
}

// SelectVictim selects the least recently used leaf node (back of list) for eviction.
func (p *LRUEvictionPolicy) SelectVictim() (uint64, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	elem := p.evictList.Back()
	if elem == nil {
		return 0, ErrEmptyCache
	}

	item := elem.Value.(*lruItem)
	p.evictList.Remove(elem)
	delete(p.items, item.nodeID)

	return item.nodeID, nil
}
