package benchmarks

import (
	"fmt"
	"testing"

	"prefixos/internal/memory"
	"prefixos/internal/radix"
)

// BenchmarkRadixMatchPrefix measures prefix lookup latency across varying prompt lengths.
func BenchmarkRadixMatchPrefix(b *testing.B) {
	lengths := []int{100, 500, 1000, 2000, 4000}

	for _, length := range lengths {
		b.Run(fmt.Sprintf("PromptLen_%d", length), func(b *testing.B) {
			bm := memory.NewBlockManager(100000)
			tree := radix.NewTree(bm)

			prompt := make([]int32, length)
			for i := 0; i < length; i++ {
				prompt[i] = int32(i + 1)
			}
			tree.InsertTokens(prompt)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				matchedLen, _ := tree.MatchPrefix(prompt)
				if matchedLen != length {
					b.Fatalf("expected matched length %d, got %d", length, matchedLen)
				}
			}
		})
	}
}

// BenchmarkRadixConcurrentReads measures concurrent prefix lookup throughput.
func BenchmarkRadixConcurrentReads(b *testing.B) {
	bm := memory.NewBlockManager(100000)
	tree := radix.NewTree(bm)

	prompt := make([]int32, 1000)
	for i := 0; i < 1000; i++ {
		prompt[i] = int32(i + 1)
	}
	tree.InsertTokens(prompt)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			matchedLen, _ := tree.MatchPrefix(prompt)
			if matchedLen != 1000 {
				b.Fatalf("expected matched length 1000, got %d", matchedLen)
			}
		}
	})
}
