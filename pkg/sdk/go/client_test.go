package client

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	"prefixos/internal/memory"
	"prefixos/internal/radix"
	"prefixos/internal/server"
	pb "prefixos/proto/v1"
)

func startMockServer(t *testing.T) (string, func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	bm := memory.NewBlockManager(1000)
	tree := radix.NewTree(bm)

	grpcServer := grpc.NewServer()
	srv := server.NewPrefixOSServer(tree, bm, nil, nil)
	pb.RegisterPrefixOSServiceServer(grpcServer, srv)

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	stop := func() {
		grpcServer.Stop()
		lis.Close()
	}

	return lis.Addr().String(), stop
}

func TestSDKClient_EndToEnd(t *testing.T) {
	addr, stop := startMockServer(t)
	defer stop()

	opts := DefaultOptions(addr)
	opts.Timeout = 2 * time.Second
	c, err := NewClient(opts)
	if err != nil {
		t.Fatalf("failed creating SDK client: %v", err)
	}
	defer c.Close()

	ctx := context.Background()

	// 1. Insert token sequence
	tokens := []int32{1001, 1002, 1003, 1004}
	success, blocks, err := c.Insert(ctx, tokens)
	if err != nil {
		t.Fatalf("SDK Insert failed: %v", err)
	}
	if !success || len(blocks) == 0 {
		t.Fatalf("expected insert success and blocks, got %v, %v", success, blocks)
	}

	// 2. Match prefix query
	query := []int32{1001, 1002, 1003, 1004, 9999}
	matchedLen, matchedBlocks, err := c.MatchPrefix(ctx, query)
	if err != nil {
		t.Fatalf("SDK MatchPrefix failed: %v", err)
	}
	if matchedLen != 4 {
		t.Fatalf("expected matched length 4, got %d", matchedLen)
	}
	if len(matchedBlocks) == 0 {
		t.Fatalf("expected matched block IDs, got empty")
	}
}
