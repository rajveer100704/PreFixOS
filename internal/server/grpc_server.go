package server

import (
	"context"

	"prefixos/internal/interfaces"
	"prefixos/internal/memory"
	"prefixos/internal/radix"
	pb "prefixos/proto/v1"
)

// PrefixOSServer implements the PrefixOSService gRPC interface.
type PrefixOSServer struct {
	pb.UnimplementedPrefixOSServiceServer
	tree *radix.Tree
	bm   *memory.BlockManager
	ev   interfaces.EvictionPolicy
	pe   interfaces.PersistenceEngine
}

// NewPrefixOSServer creates a new instance of PrefixOSServer wiring memory, tree, eviction, and persistence.
func NewPrefixOSServer(tree *radix.Tree, bm *memory.BlockManager, ev interfaces.EvictionPolicy, pe interfaces.PersistenceEngine) *PrefixOSServer {
	return &PrefixOSServer{
		tree: tree,
		bm:   bm,
		ev:   ev,
		pe:   pe,
	}
}

// MatchPrefix finds the longest shared sequence of tokens in the Radix Tree.
func (s *PrefixOSServer) MatchPrefix(ctx context.Context, req *pb.MatchPrefixRequest) (*pb.MatchPrefixResponse, error) {
	matchedLen, blockIDs := s.tree.FindLongestPrefix(req.Tokens)

	pbBlockIDs := make([]int32, len(blockIDs))
	copy(pbBlockIDs, blockIDs)

	return &pb.MatchPrefixResponse{
		MatchedLength: int32(matchedLen),
		BlockIds:      pbBlockIDs,
	}, nil
}

// Insert allocates blocks and adds the new tokens into the Radix Tree.
func (s *PrefixOSServer) Insert(ctx context.Context, req *pb.InsertRequest) (*pb.InsertResponse, error) {
	success, blockIDs := s.tree.InsertTokens(req.Tokens)

	var pbBlockIDs []int32
	if success {
		pbBlockIDs = make([]int32, len(blockIDs))
		copy(pbBlockIDs, blockIDs)

		if s.pe != nil {
			_ = s.pe.AppendWAL(interfaces.WALEntry{
				Type:    1, // Insert
				Payload: []byte("insert_op"),
			})
		}
	}

	return &pb.InsertResponse{
		Success:         success,
		AllocatedBlocks: pbBlockIDs,
	}, nil
}
