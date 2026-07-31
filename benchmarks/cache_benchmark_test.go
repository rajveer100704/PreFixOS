package benchmarks

import (
	"fmt"
	"testing"

	"prefixos/internal/memory"
	"prefixos/internal/radix"
)

// BenchmarkAgenticTreeSearch simulates an LLM system cache doing Tree-of-Thought search.
// Scenario: 1 system prompt (2000 tokens) branching into 50 sub-agents,
// each generating 500 unique tokens.
func BenchmarkAgenticTreeSearch(b *testing.B) {
	promptTokens := 2000
	agents := 50
	generatedTokens := 500

	// 1. Calculate naive flat cache usage
	// Flat cache stores the entire sequence (Prompt + Generated) for each agent independently.
	naiveTotalTokens := agents * (promptTokens + generatedTokens)
	naiveBlocks := (naiveTotalTokens + memory.BlockSize - 1) / memory.BlockSize

	// 2. Simulate RadixKV
	capacity := 100000 // Ensure we have enough simulated VRAM blocks
	var radixAllocatedBlocks int

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		bm := memory.NewBlockManager(capacity)
		tree := radix.NewTree(bm)
		radixAllocatedBlocks = 0
		b.StartTimer()

		// A. Insert the shared system prompt
		rootPrompt := make([]int32, promptTokens)
		for j := 0; j < promptTokens; j++ {
			rootPrompt[j] = int32(j + 1)
		}

		_, blocks := tree.InsertTokens(rootPrompt)
		radixAllocatedBlocks += len(blocks)

		// B. Branch into 50 sub-agents
		for a := 0; a < agents; a++ {
			// Each agent's context is the root prompt + its unique 500 tokens
			agentTokens := make([]int32, promptTokens+generatedTokens)
			copy(agentTokens, rootPrompt)

			for j := 0; j < generatedTokens; j++ {
				// Offset by 10000 to ensure completely unique token suffixes per agent
				agentTokens[promptTokens+j] = int32(10000 + (a * 1000) + j)
			}

			_, blocks := tree.InsertTokens(agentTokens)
			radixAllocatedBlocks += len(blocks)
		}
	}

	b.StopTimer()

	// Print the result on the first iteration to display the metrics
	if b.N == 1 {
		fmt.Printf("\n--- VRAM Memory Usage Comparison ---\n")
		fmt.Printf("Scenario: 1 Root Prompt (%d tokens) -> %d Agents generating %d tokens each\n", promptTokens, agents, generatedTokens)
		fmt.Printf("Naive Flat Cache: %d blocks (approx %d tokens in VRAM)\n", naiveBlocks, naiveTotalTokens)
		fmt.Printf("RadixKV Cache:    %d blocks (approx %d tokens in VRAM)\n", radixAllocatedBlocks, radixAllocatedBlocks*memory.BlockSize)
		fmt.Printf("Memory Saved:     %.2f%%\n", float64(naiveBlocks-radixAllocatedBlocks)/float64(naiveBlocks)*100)
		fmt.Printf("------------------------------------\n\n")
	}
}
