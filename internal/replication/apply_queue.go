package replication

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

// ApplyMsg is sent by Raft over the ApplyQueue to notify the state machine worker of committed entries.
type ApplyMsg struct {
	CommandValid bool
	Command      Command
	CommandIndex uint64
	CommandTerm  uint64
}

// ApplyQueue manages async state machine execution from committed Raft log entries.
type ApplyQueue struct {
	mu           sync.Mutex
	queue        chan ApplyMsg
	sm           *StateMachine
	ctx          context.Context
	cancel       context.CancelFunc
	appliedIndex uint64
}

// NewApplyQueue initializes an asynchronous apply queue worker loop.
func NewApplyQueue(sm *StateMachine, bufferSize int) *ApplyQueue {
	ctx, cancel := context.WithCancel(context.Background())
	aq := &ApplyQueue{
		queue:  make(chan ApplyMsg, bufferSize),
		sm:     sm,
		ctx:    ctx,
		cancel: cancel,
	}

	go aq.startWorker()
	return aq
}

// PushCommittedEntry pushes a committed Raft LogEntry onto the ApplyQueue channel.
func (aq *ApplyQueue) PushCommittedEntry(entry LogEntry) error {
	var cmd Command

	// Unmarshal payload type
	var raw map[string]interface{}
	if err := json.Unmarshal(entry.Payload, &raw); err != nil {
		return fmt.Errorf("failed unmarshaling command payload: %w", err)
	}

	opType, _ := raw["type"].(string)
	switch opType {
	case "INSERT":
		var ins InsertCommand
		if err := json.Unmarshal(entry.Payload, &ins); err == nil {
			cmd = &ins
		}
	case "EVICT":
		var ev EvictCommand
		if err := json.Unmarshal(entry.Payload, &ev); err == nil {
			cmd = &ev
		}
	default:
		cmd = &InsertCommand{}
	}

	aq.queue <- ApplyMsg{
		CommandValid: true,
		Command:      cmd,
		CommandIndex: entry.Index,
		CommandTerm:  entry.Term,
	}
	return nil
}

// startWorker processes committed messages sequentially on a background thread.
func (aq *ApplyQueue) startWorker() {
	for {
		select {
		case <-aq.ctx.Done():
			return
		case msg := <-aq.queue:
			if msg.CommandValid && msg.Command != nil {
				if err := msg.Command.Apply(aq.sm); err != nil {
					log.Printf("error applying command at index %d: %v", msg.CommandIndex, err)
				} else {
					aq.mu.Lock()
					aq.appliedIndex = msg.CommandIndex
					aq.mu.Unlock()
				}
			}
		}
	}
}

// GetAppliedIndex returns the highest log index applied to the state machine.
func (aq *ApplyQueue) GetAppliedIndex() uint64 {
	aq.mu.Lock()
	defer aq.mu.Unlock()
	return aq.appliedIndex
}

// Close gracefully stops the ApplyQueue worker loop.
func (aq *ApplyQueue) Close() {
	aq.cancel()
}
