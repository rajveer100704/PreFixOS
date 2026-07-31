package main

import (
	"context"
	"fmt"
	"log"

	sdk "prefixos/pkg/sdk/go"
)

func main() {
	fmt.Println("=== PrefixOS Typed SDK Client Example ===")

	// Initialize SDK Client
	opts := sdk.DefaultOptions("127.0.0.1:50051")
	client, err := sdk.NewClient(opts)
	if err != nil {
		log.Fatalf("failed connecting to PrefixOS: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// 1. Store System Prompt
	promptTokens := []int32{101, 102, 103, 104, 105, 106, 107, 108}
	fmt.Printf("1. Registering System Prompt (%d tokens)... ", len(promptTokens))

	success, blocks, err := client.Insert(ctx, promptTokens)
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
		return
	}
	fmt.Printf("SUCCESS! (Status=%v) Allocated Physical Block Handles: %v\n", success, blocks)

	// 2. Query Prefix Match for Sub-Agent
	subAgentTokens := []int32{101, 102, 103, 104, 105, 106, 107, 108, 9001, 9002}
	fmt.Printf("2. Querying Prefix Match for Sub-Agent (%d tokens)... ", len(subAgentTokens))

	matchedLen, matchedBlocks, err := client.MatchPrefix(ctx, subAgentTokens)
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
		return
	}
	fmt.Printf("SUCCESS! Matched Prefix Length: %d tokens, VRAM Block Handles: %v\n", matchedLen, matchedBlocks)
}
