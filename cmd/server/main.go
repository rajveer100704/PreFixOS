package main

import (
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"radixkv/internal/eviction"
	"radixkv/internal/memory"
	"radixkv/internal/radix"
	"radixkv/internal/server"
	pb "radixkv/proto"
)

func main() {
	log.Println("Bootstrapping RadixKV Engine...")

	// 1. Initialize Memory Manager (VRAM Simulator)
	// We allocate 65,536 blocks. At 16 tokens/block, this is ~1M tokens capacity.
	capacity := 65536
	bm := memory.NewBlockManager(capacity)

	// 2. Initialize the Radix Tree
	tree := radix.NewTree(bm)

	// 3. Initialize the Tree-Aware LRU Garbage Collector
	lru := eviction.NewTreeLRU(tree, bm)

	// Start the background GC. 
	// In production, the trigger would check bm.FreeBlocks() against a low-watermark.
	gcTrigger := func() bool {
		// Mock trigger: normally returns bm.Available() < threshold
		return false
	}
	lru.RunBackgroundGarbageCollector(1*time.Second, gcTrigger)

	// 4. Initialize gRPC Server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	kvServer := server.NewKVCacheServer(tree)
	pb.RegisterKVCacheServiceServer(grpcServer, kvServer)

	log.Printf("RadixKV gRPC Server listening on :50051 (Memory Capacity: %d blocks)", capacity)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
