package interfaces

// EvictionPolicy defines pluggable eviction strategy contracts
type EvictionPolicy interface {
	SelectVictim() (nodeID uint64, err error)
	OnAccess(nodeID uint64, hitCount uint64)
	OnInsert(nodeID uint64, depth int, cost float64)
	OnDelete(nodeID uint64)
}
