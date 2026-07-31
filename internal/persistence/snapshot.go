package persistence

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"prefixos/internal/memory"
	"prefixos/internal/radix"
)

// SnapshotManager handles background atomic serialization and restoration of tree & memory state.
type SnapshotManager struct {
	mu           sync.Mutex
	dir          string
	tree         *radix.Tree
	bm           *memory.BlockManager
	lastSnapshot string
}

// NewSnapshotManager creates a snapshot manager targeting the specified directory.
func NewSnapshotManager(dir string, tree *radix.Tree, bm *memory.BlockManager) (*SnapshotManager, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed creating snapshot directory: %w", err)
	}
	return &SnapshotManager{
		dir:  dir,
		tree: tree,
		bm:   bm,
	}, nil
}

// CreateSnapshot serializes tree nodes and block manager state into a timestamped file.
func (sm *SnapshotManager) CreateSnapshot() (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	snapshotID := fmt.Sprintf("snap-%d.bin", time.Now().UnixNano())
	path := filepath.Join(sm.dir, snapshotID)

	tmpPath := path + ".tmp"
	file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", fmt.Errorf("failed opening snapshot file: %w", err)
	}

	// 1. Write Memory Allocator Stats Header
	stats := sm.bm.Stats()
	var header [16]byte
	binary.BigEndian.PutUint64(header[0:8], uint64(stats.TotalBlocks))
	binary.BigEndian.PutUint64(header[8:16], uint64(stats.AllocatedBlocks))
	if _, err := file.Write(header[:]); err != nil {
		file.Close()
		os.Remove(tmpPath)
		return "", err
	}

	// 2. Iterate Radix Tree and Serialize Nodes
	it := sm.tree.Iterator()
	for it.HasNext() {
		tokens, blockIDs := it.Next()

		// Write token length and tokens
		tokLenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(tokLenBuf, uint32(len(tokens)))
		if _, err := file.Write(tokLenBuf); err != nil {
			file.Close()
			os.Remove(tmpPath)
			return "", err
		}

		tokBuf := make([]byte, len(tokens)*4)
		for i, t := range tokens {
			binary.BigEndian.PutUint32(tokBuf[i*4:], uint32(t))
		}
		if _, err := file.Write(tokBuf); err != nil {
			file.Close()
			os.Remove(tmpPath)
			return "", err
		}

		// Write block ID count and block IDs
		blkLenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(blkLenBuf, uint32(len(blockIDs)))
		if _, err := file.Write(blkLenBuf); err != nil {
			file.Close()
			os.Remove(tmpPath)
			return "", err
		}

		blkBuf := make([]byte, len(blockIDs)*4)
		for i, b := range blockIDs {
			binary.BigEndian.PutUint32(blkBuf[i*4:], uint32(b))
		}
		if _, err := file.Write(blkBuf); err != nil {
			file.Close()
			os.Remove(tmpPath)
			return "", err
		}
	}

	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(tmpPath)
		return "", err
	}
	file.Close()

	if err := os.Rename(tmpPath, path); err != nil {
		return "", err
	}

	sm.lastSnapshot = snapshotID
	return snapshotID, nil
}

// RestoreFromSnapshot reads and restores state from a snapshot binary file.
func (sm *SnapshotManager) RestoreFromSnapshot(snapshotID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	path := filepath.Join(sm.dir, snapshotID)
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed opening snapshot file: %w", err)
	}
	defer file.Close()

	var header [16]byte
	if _, err := io.ReadFull(file, header[:]); err != nil {
		return fmt.Errorf("failed reading snapshot header: %w", err)
	}

	// Read serialized nodes until EOF
	for {
		var tokLenBuf [4]byte
		if _, err := io.ReadFull(file, tokLenBuf[:]); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		tokLen := binary.BigEndian.Uint32(tokLenBuf[:])

		tokBuf := make([]byte, tokLen*4)
		if _, err := io.ReadFull(file, tokBuf); err != nil {
			return err
		}
		tokens := make([]int32, tokLen)
		for i := 0; i < int(tokLen); i++ {
			tokens[i] = int32(binary.BigEndian.Uint32(tokBuf[i*4:]))
		}

		var blkLenBuf [4]byte
		if _, err := io.ReadFull(file, blkLenBuf[:]); err != nil {
			return err
		}
		blkLen := binary.BigEndian.Uint32(blkLenBuf[:])

		blkBuf := make([]byte, blkLen*4)
		if _, err := io.ReadFull(file, blkBuf); err != nil {
			return err
		}
		blockIDs := make([]int32, blkLen)
		for i := 0; i < int(blkLen); i++ {
			blockIDs[i] = int32(binary.BigEndian.Uint32(blkBuf[i*4:]))
		}

		// Insert restored sequence into tree
		_ = sm.tree.Insert(tokens, blockIDs)
	}

	return nil
}
