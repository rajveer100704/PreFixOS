package benchmarks

import (
	"path/filepath"
	"testing"

	"prefixos/internal/interfaces"
	"prefixos/internal/memory"
	"prefixos/internal/persistence"
	"prefixos/internal/radix"
)

// BenchmarkWALAppend measures WAL append latency.
func BenchmarkWALAppend(b *testing.B) {
	tmpDir := b.TempDir()
	walPath := filepath.Join(tmpDir, "bench.wal")

	wal, err := persistence.NewWALManager(walPath, false)
	if err != nil {
		b.Fatal(err)
	}
	defer wal.Close()

	entry := interfaces.WALEntry{
		Type:    persistence.OpInsert,
		Payload: []byte("token_sequence_payload_data_for_wal_benchmark"),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := wal.AppendWAL(entry); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSnapshotCreation measures snapshot serialization throughput.
func BenchmarkSnapshotCreation(b *testing.B) {
	tmpDir := b.TempDir()
	bm := memory.NewBlockManager(10000)
	tree := radix.NewTree(bm)

	for i := 0; i < 100; i++ {
		tokens := make([]int32, 50)
		for j := 0; j < 50; j++ {
			tokens[j] = int32(i*50 + j)
		}
		tree.InsertTokens(tokens)
	}

	snapMgr, err := persistence.NewSnapshotManager(tmpDir, tree, bm)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := snapMgr.CreateSnapshot()
		if err != nil {
			b.Fatal(err)
		}
	}
}
