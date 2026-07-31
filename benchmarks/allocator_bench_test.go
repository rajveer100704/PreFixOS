package benchmarks

import (
	"testing"

	"prefixos/internal/memory"
)

// BenchmarkSlabAllocation measures single and multi-block allocation latency.
func BenchmarkSlabAllocation(b *testing.B) {
	bm := memory.NewBlockManager(b.N + 1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ids, err := bm.Allocate(1)
		if err != nil {
			b.Fatal(err)
		}
		_ = bm.Free(ids)
	}
}

// BenchmarkSlabBatchAllocation measures batch block allocation latency.
func BenchmarkSlabBatchAllocation(b *testing.B) {
	bm := memory.NewBlockManager(b.N*16 + 1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ids, err := bm.Allocate(16)
		if err != nil {
			b.Fatal(err)
		}
		_ = bm.Free(ids)
	}
}
