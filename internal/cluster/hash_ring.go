package cluster

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
)

// HashRing implements a consistent hash ring with virtual nodes for request routing.
type HashRing struct {
	mu           sync.RWMutex
	vnodesPerNode int
	ring         []uint32
	vnodeMap     map[uint32]string
	shards       map[string]bool
}

// NewHashRing creates a new consistent hash ring instance.
func NewHashRing(vnodes int) *HashRing {
	return &HashRing{
		vnodesPerNode: vnodes,
		vnodeMap:     make(map[uint32]string),
		shards:       make(map[string]bool),
	}
}

func hashKey(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32()
}

// AddShard adds a new shard ID to the consistent hash ring.
func (hr *HashRing) AddShard(shardID string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if hr.shards[shardID] {
		return
	}
	hr.shards[shardID] = true

	for i := 0; i < hr.vnodesPerNode; i++ {
		vnodeKey := shardID + "#" + strconv.Itoa(i)
		hash := hashKey(vnodeKey)
		hr.ring = append(hr.ring, hash)
		hr.vnodeMap[hash] = shardID
	}

	sort.Slice(hr.ring, func(i, j int) bool { return hr.ring[i] < hr.ring[j] })
}

// RemoveShard removes a shard from the hash ring.
func (hr *HashRing) RemoveShard(shardID string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if !hr.shards[shardID] {
		return
	}
	delete(hr.shards, shardID)

	var newRing []uint32
	for _, hash := range hr.ring {
		if hr.vnodeMap[hash] == shardID {
			delete(hr.vnodeMap, hash)
		} else {
			newRing = append(newRing, hash)
		}
	}
	hr.ring = newRing
}

// GetShard maps a token sequence to its target Shard ID on the ring.
func (hr *HashRing) GetShard(tokens []int32) (string, error) {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if len(hr.ring) == 0 {
		return "", fmt.Errorf("hash ring is empty")
	}

	keyStr := fmt.Sprintf("%v", tokens)
	hash := hashKey(keyStr)

	idx := sort.Search(len(hr.ring), func(i int) bool { return hr.ring[i] >= hash })
	if idx == len(hr.ring) {
		idx = 0
	}

	return hr.vnodeMap[hr.ring[idx]], nil
}
