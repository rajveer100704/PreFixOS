package persistence

import (
	"os"
	"path/filepath"
	"testing"

	"prefixos/internal/interfaces"
	"prefixos/internal/memory"
	"prefixos/internal/radix"
)

func TestWAL_AppendAndRead(t *testing.T) {
	tmpDir := t.TempDir()
	walPath := filepath.Join(tmpDir, "test.wal")

	wal, err := NewWALManager(walPath, true)
	if err != nil {
		t.Fatalf("failed to create WAL: %v", err)
	}

	entry1 := interfaces.WALEntry{
		Type:    OpInsert,
		Payload: []byte("token_sequence_1"),
	}
	entry2 := interfaces.WALEntry{
		Type:    OpEvict,
		Payload: []byte("token_sequence_2"),
	}

	if err := wal.AppendWAL(entry1); err != nil {
		t.Fatalf("failed to append entry1: %v", err)
	}
	if err := wal.AppendWAL(entry2); err != nil {
		t.Fatalf("failed to append entry2: %v", err)
	}

	records, maxSeq, err := wal.ReadAllRecords()
	if err != nil {
		t.Fatalf("failed to read WAL records: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if maxSeq != 2 {
		t.Fatalf("expected maxSeq 2, got %d", maxSeq)
	}
	if string(records[0].Payload) != "token_sequence_1" {
		t.Errorf("expected payload token_sequence_1, got %s", records[0].Payload)
	}

	wal.Close()
}

func TestWAL_CorruptChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	walPath := filepath.Join(tmpDir, "corrupt.wal")

	wal, _ := NewWALManager(walPath, true)
	_ = wal.AppendWAL(interfaces.WALEntry{Type: OpInsert, Payload: []byte("valid_data_for_checksum_test")})
	wal.Close()

	// Corrupt payload byte
	data, _ := os.ReadFile(walPath)
	if len(data) > 15 {
		data[15] ^= 0xFF // Mutate byte inside payload
		_ = os.WriteFile(walPath, data, 0644)
	}

	walReader, err := NewWALManager(walPath, false)
	if err != nil {
		t.Fatalf("failed opening WAL reader: %v", err)
	}
	defer walReader.Close()

	_, _, err = walReader.ReadAllRecords()
	if err == nil {
		t.Fatal("expected checksum error on corrupt WAL file")
	}
}

func TestSnapshot_CreateAndRestore(t *testing.T) {
	tmpDir := t.TempDir()
	bm := memory.NewBlockManager(100)
	tree := radix.NewTree(bm)

	tokens := []int32{10, 20, 30, 40}
	tree.InsertTokens(tokens)

	snapMgr, err := NewSnapshotManager(tmpDir, tree, bm)
	if err != nil {
		t.Fatalf("failed creating SnapshotManager: %v", err)
	}

	snapID, err := snapMgr.CreateSnapshot()
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	// Restore state into new tree instance
	newBM := memory.NewBlockManager(100)
	newTree := radix.NewTree(newBM)
	newSnapMgr, _ := NewSnapshotManager(tmpDir, newTree, newBM)

	if err := newSnapMgr.RestoreFromSnapshot(snapID); err != nil {
		t.Fatalf("RestoreFromSnapshot failed: %v", err)
	}

	matchedLen, _ := newTree.MatchPrefix(tokens)
	if matchedLen != 4 {
		t.Fatalf("expected matched length 4 after restore, got %d", matchedLen)
	}
}

func TestPersistenceEngine_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	bm := memory.NewBlockManager(100)
	tree := radix.NewTree(bm)

	engine, err := NewEngine(tmpDir, tree, bm, true)
	if err != nil {
		t.Fatalf("failed creating Engine: %v", err)
	}

	entry := interfaces.WALEntry{Type: OpInsert, Payload: []byte("test_mutation")}
	if err := engine.AppendWAL(entry); err != nil {
		t.Fatalf("AppendWAL failed: %v", err)
	}

	snapID, err := engine.CreateSnapshot()
	if err != nil || snapID == "" {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	lastSeq, err := engine.Recover()
	if err != nil {
		t.Fatalf("Recover failed: %v", err)
	}
	if lastSeq != 1 {
		t.Errorf("expected lastSeq 1, got %d", lastSeq)
	}

	engine.Close()
}
