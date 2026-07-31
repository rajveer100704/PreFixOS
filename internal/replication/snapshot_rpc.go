package replication

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// InstallSnapshotArgs represents InstallSnapshot RPC parameters for chunked snapshot streaming.
type InstallSnapshotArgs struct {
	Term              uint64
	LeaderID          int64
	LastIncludedIndex uint64
	LastIncludedTerm  uint64
	Offset            uint64
	Data              []byte
	Done              bool
}

// InstallSnapshotReply represents InstallSnapshot RPC reply parameters.
type InstallSnapshotReply struct {
	Term uint64
}

// SnapshotStreamer manages resumable chunked snapshot reception for lagging followers.
type SnapshotStreamer struct {
	mu           sync.Mutex
	dataDir      string
	activeID     string
	tempFile     *os.File
	bytesWritten uint64
}

// NewSnapshotStreamer creates a new snapshot streamer instance.
func NewSnapshotStreamer(dataDir string) *SnapshotStreamer {
	return &SnapshotStreamer{
		dataDir: dataDir,
	}
}

// ReceiveChunk writes a chunk payload from InstallSnapshot RPC into temporary storage.
func (ss *SnapshotStreamer) ReceiveChunk(args *InstallSnapshotArgs) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// 1. First chunk initialization
	if args.Offset == 0 || ss.tempFile == nil {
		if ss.tempFile != nil {
			_ = ss.tempFile.Close()
			_ = os.Remove(ss.tempFile.Name())
		}
		ss.activeID = fmt.Sprintf("snapshot_%d_%d", args.LastIncludedIndex, args.LastIncludedTerm)
		path := filepath.Join(ss.dataDir, fmt.Sprintf("snapshot_chunk_%d.tmp", args.LastIncludedIndex))
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed creating snapshot temp file: %w", err)
		}
		ss.tempFile = f
		ss.bytesWritten = 0
	}

	// 2. Write chunk at offset
	if _, err := ss.tempFile.WriteAt(args.Data, int64(args.Offset)); err != nil {
		return fmt.Errorf("failed writing snapshot chunk at offset %d: %w", args.Offset, err)
	}
	ss.bytesWritten += uint64(len(args.Data))

	// 3. Final chunk commit
	if args.Done {
		_ = ss.tempFile.Sync()
		_ = ss.tempFile.Close()
		finalPath := filepath.Join(ss.dataDir, fmt.Sprintf("snapshot_%d.snap", args.LastIncludedIndex))
		if err := os.Rename(ss.tempFile.Name(), finalPath); err != nil {
			return fmt.Errorf("failed finalizing snapshot file: %w", err)
		}
		ss.tempFile = nil
	}

	return nil
}

// GetActiveSnapshotInfo returns the current in-progress snapshot streaming activeID and bytesWritten count.
func (ss *SnapshotStreamer) GetActiveSnapshotInfo() (string, uint64) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.activeID, ss.bytesWritten
}
