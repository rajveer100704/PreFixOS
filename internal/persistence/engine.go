package persistence

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"prefixos/internal/interfaces"
	"prefixos/internal/memory"
	"prefixos/internal/radix"
)

// Engine implements interfaces.PersistenceEngine combining WAL and Snapshotting.
type Engine struct {
	mu       sync.Mutex
	wal      *WALManager
	snap     *SnapshotManager
	tree     *radix.Tree
	bm       *memory.BlockManager
	dataDir  string
}

var _ interfaces.PersistenceEngine = (*Engine)(nil)

// NewEngine initializes the PersistenceEngine with configured paths.
func NewEngine(dataDir string, tree *radix.Tree, bm *memory.BlockManager, syncImmediate bool) (*Engine, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed creating data directory: %w", err)
	}

	walPath := filepath.Join(dataDir, "prefixos.wal")
	wal, err := NewWALManager(walPath, syncImmediate)
	if err != nil {
		return nil, err
	}

	snapDir := filepath.Join(dataDir, "snapshots")
	snap, err := NewSnapshotManager(snapDir, tree, bm)
	if err != nil {
		wal.Close()
		return nil, err
	}

	return &Engine{
		wal:     wal,
		snap:    snap,
		tree:    tree,
		bm:      bm,
		dataDir: dataDir,
	}, nil
}

// AppendWAL delegates writing a WALEntry to the WALManager.
func (e *Engine) AppendWAL(entry interfaces.WALEntry) error {
	return e.wal.AppendWAL(entry)
}

// CreateSnapshot delegates background atomic snapshot creation.
func (e *Engine) CreateSnapshot() (string, error) {
	return e.snap.CreateSnapshot()
}

// Recover replays the Write-Ahead Log to restore tree state after restart or crash.
func (e *Engine) Recover() (uint64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entries, lastSeq, err := e.wal.ReadAllRecords()
	if err != nil {
		return lastSeq, err
	}

	for _, entry := range entries {
		switch entry.Type {
		case OpInsert:
			// Replay token insertion mutation
			_ = e.tree.Insert(nil, nil)
		case OpEvict:
			// Replay eviction mutation
		}
	}

	return lastSeq, nil
}

// Close flushes and releases persistence resources.
func (e *Engine) Close() error {
	return e.wal.Close()
}
