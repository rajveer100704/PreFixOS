package replication

import (
	"context"
	"fmt"
	"time"
)

// ReadIndexResult stores the commit index for a linearizable read request.
type ReadIndexResult struct {
	ReadIndex uint64
	Err       error
}

// ReadIndex requests the current committed index from the Leader to serve linearizable reads without log append.
func (rn *RaftNode) ReadIndex(ctx context.Context) (uint64, error) {
	rn.mu.Lock()
	if rn.role != Leader {
		rn.mu.Unlock()
		return 0, fmt.Errorf("not leader")
	}

	// Verify leader lease validity
	if time.Since(rn.lastHeartbeat) > rn.electionTimeout {
		rn.mu.Unlock()
		return 0, fmt.Errorf("leader lease expired")
	}

	commitIdx := rn.commitIndex
	rn.mu.Unlock()

	return commitIdx, nil
}
