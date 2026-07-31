package eviction

import (
	"container/list"
	"sync"

	"prefixos/internal/interfaces"
)

// SLRUEvictionPolicy implements Segmented LRU (Probationary + Protected queues).
type SLRUEvictionPolicy struct {
	mu           sync.Mutex
	probationMap map[uint64]*list.Element
	protectedMap map[uint64]*list.Element
	probation    *list.List
	protected    *list.List
	protectedCap int
}

var _ interfaces.EvictionPolicy = (*SLRUEvictionPolicy)(nil)

// NewSLRUEvictionPolicy initializes an SLRU eviction manager with specified protected capacity.
func NewSLRUEvictionPolicy(protectedCap int) *SLRUEvictionPolicy {
	if protectedCap <= 0 {
		protectedCap = 1000
	}
	return &SLRUEvictionPolicy{
		probationMap: make(map[uint64]*list.Element),
		protectedMap: make(map[uint64]*list.Element),
		probation:    list.New(),
		protected:    list.New(),
		protectedCap: protectedCap,
	}
}

// OnInsert adds a new item to the Probation queue.
func (s *SLRUEvictionPolicy) OnInsert(nodeID uint64, depth int, cost float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := &lruItem{nodeID: nodeID, depth: depth, cost: cost}
	elem := s.probation.PushFront(item)
	s.probationMap[nodeID] = elem
}

// OnAccess promotes an item from Probation to Protected queue on access.
func (s *SLRUEvictionPolicy) OnAccess(nodeID uint64, hitCount uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If in protected, refresh recency
	if elem, exists := s.protectedMap[nodeID]; exists {
		s.protected.MoveToFront(elem)
		return
	}

	// If in probation, promote to protected
	if elem, exists := s.probationMap[nodeID]; exists {
		s.probation.Remove(elem)
		delete(s.probationMap, nodeID)

		pElem := s.protected.PushFront(elem.Value)
		s.protectedMap[nodeID] = pElem

		// Overflow protected back to probation if over capacity
		if s.protected.Len() > s.protectedCap {
			demoteElem := s.protected.Back()
			s.protected.Remove(demoteElem)
			demotedItem := demoteElem.Value.(*lruItem)
			delete(s.protectedMap, demotedItem.nodeID)

			newProbElem := s.probation.PushFront(demotedItem)
			s.probationMap[demotedItem.nodeID] = newProbElem
		}
	}
}

// OnDelete removes nodeID from both probation and protected queues.
func (s *SLRUEvictionPolicy) OnDelete(nodeID uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if elem, exists := s.probationMap[nodeID]; exists {
		s.probation.Remove(elem)
		delete(s.probationMap, nodeID)
	}
	if elem, exists := s.protectedMap[nodeID]; exists {
		s.protected.Remove(elem)
		delete(s.protectedMap, nodeID)
	}
}

// SelectVictim selects victim first from Probation queue, then from Protected.
func (s *SLRUEvictionPolicy) SelectVictim() (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if elem := s.probation.Back(); elem != nil {
		item := elem.Value.(*lruItem)
		s.probation.Remove(elem)
		delete(s.probationMap, item.nodeID)
		return item.nodeID, nil
	}

	if elem := s.protected.Back(); elem != nil {
		item := elem.Value.(*lruItem)
		s.protected.Remove(elem)
		delete(s.protectedMap, item.nodeID)
		return item.nodeID, nil
	}

	return 0, ErrEmptyCache
}
