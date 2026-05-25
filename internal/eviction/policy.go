package eviction

// Policy defines the interface for cache eviction strategies.
type Policy interface {
	Evict() bool
}
