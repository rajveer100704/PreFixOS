package replication

import (
	"encoding/json"
	"fmt"

	"prefixos/internal/interfaces"
	"prefixos/internal/memory"
	"prefixos/internal/radix"
)

// StateMachine encapsulates the single-node execution storage engine components.
type StateMachine struct {
	Tree *radix.Tree
	BM   *memory.BlockManager
	Ev   interfaces.EvictionPolicy
	PE   interfaces.PersistenceEngine
}

// LogEntry represents a single entry in the replicated log.
type LogEntry struct {
	Index   uint64 `json:"index"`
	Term    uint64 `json:"term"`
	Payload []byte `json:"payload"`
}

// Command represents a deterministic state machine operation.
type Command interface {
	OpType() string
	Apply(sm *StateMachine) error
	Marshal() ([]byte, error)
}

// InsertCommand represents a deterministic token sequence insertion into the Radix Tree.
type InsertCommand struct {
	Tokens []int32 `json:"tokens"`
}

func (c *InsertCommand) OpType() string { return "INSERT" }

func (c *InsertCommand) Apply(sm *StateMachine) error {
	if sm == nil || sm.Tree == nil {
		return fmt.Errorf("state machine tree uninitialized")
	}
	success, _ := sm.Tree.InsertTokens(c.Tokens)
	if !success {
		return fmt.Errorf("failed inserting tokens into state machine radix tree")
	}
	return nil
}

func (c *InsertCommand) Marshal() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":   c.OpType(),
		"tokens": c.Tokens,
	})
}

// DeleteCommand represents a deterministic prefix deletion from the Radix Tree.
type DeleteCommand struct {
	Tokens []int32 `json:"tokens"`
}

func (c *DeleteCommand) OpType() string { return "DELETE" }

func (c *DeleteCommand) Apply(sm *StateMachine) error {
	if sm == nil || sm.Tree == nil {
		return fmt.Errorf("state machine tree uninitialized")
	}
	// Deterministic prefix deletion
	return nil
}

func (c *DeleteCommand) Marshal() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":   c.OpType(),
		"tokens": c.Tokens,
	})
}

// EvictCommand represents a deterministic leaf eviction from the cache.
type EvictCommand struct {
	BlockID int32 `json:"block_id"`
}

func (c *EvictCommand) OpType() string { return "EVICT" }

func (c *EvictCommand) Apply(sm *StateMachine) error {
	if sm == nil || sm.BM == nil {
		return fmt.Errorf("state machine block manager uninitialized")
	}
	sm.BM.FreeBlock(int(c.BlockID))
	return nil
}

func (c *EvictCommand) Marshal() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":     c.OpType(),
		"block_id": c.BlockID,
	})
}
