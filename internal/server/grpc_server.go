package server

import (
	"context"

	"radixkv/internal/radix"
	pb "radixkv/proto"
)

// KVCacheServer implements the KVCacheService gRPC interface.
type KVCacheServer struct {
	pb.UnimplementedKVCacheServiceServer
	tree *radix.Tree
}

// NewKVCacheServer creates a new instance of KVCacheServer.
func NewKVCacheServer(tree *radix.Tree) *KVCacheServer {
	return &KVCacheServer{
		tree: tree,
	}
}

// MatchPrefix finds the longest shared sequence of tokens in the Radix Tree.
func (s *KVCacheServer) MatchPrefix(ctx context.Context, req *pb.MatchPrefixRequest) (*pb.MatchPrefixResponse, error) {
	matchedLen, blockIDs := s.tree.FindLongestPrefix(req.Tokens)

	// Convert []int to []int32 for protobuf compatibility
	pbBlockIDs := make([]int32, len(blockIDs))
	for i, id := range blockIDs {
		pbBlockIDs[i] = int32(id)
	}

	return &pb.MatchPrefixResponse{
		MatchedLength: int32(matchedLen),
		BlockIds:      pbBlockIDs,
	}, nil
}

// Insert allocates blocks and adds the new tokens into the Radix Tree.
func (s *KVCacheServer) Insert(ctx context.Context, req *pb.InsertRequest) (*pb.InsertResponse, error) {
	success, blockIDs := s.tree.Insert(req.Tokens)

	var pbBlockIDs []int32
	if success {
		pbBlockIDs = make([]int32, len(blockIDs))
		for i, id := range blockIDs {
			pbBlockIDs[i] = int32(id)
		}
	}

	return &pb.InsertResponse{
		Success:  success,
		BlockIds: pbBlockIDs,
	}, nil
}
