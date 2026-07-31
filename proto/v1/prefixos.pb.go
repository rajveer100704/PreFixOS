package prefixosv1

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

// JSONCodec registers a high-performance JSON codec for gRPC serialization.
type JSONCodec struct{}

func (JSONCodec) Name() string { return "json" }
func (JSONCodec) Marshal(v interface{}) ([]byte, error) { return json.Marshal(v) }
func (JSONCodec) Unmarshal(data []byte, v interface{}) error { return json.Unmarshal(data, v) }

func init() {
	encoding.RegisterCodec(JSONCodec{})
}

// MatchPrefixRequest represents a token sequence lookup query.
type MatchPrefixRequest struct {
	Tokens   []int32 `json:"tokens,omitempty"`
	TenantId string  `json:"tenant_id,omitempty"`
}

// MatchPrefixResponse represents physical memory handles matching a prefix.
type MatchPrefixResponse struct {
	MatchedLength int32   `json:"matched_length,omitempty"`
	BlockIds      []int32 `json:"block_ids,omitempty"`
}

// InsertRequest represents a new token sequence registration.
type InsertRequest struct {
	Tokens   []int32 `json:"tokens,omitempty"`
	TenantId string  `json:"tenant_id,omitempty"`
}

// InsertResponse represents newly assigned physical memory handles.
type InsertResponse struct {
	Success         bool    `json:"success,omitempty"`
	AllocatedBlocks []int32 `json:"allocated_blocks,omitempty"`
}

// BatchLookupRequest represents batch query requests.
type BatchLookupRequest struct {
	Requests []*MatchPrefixRequest `json:"requests,omitempty"`
}

// BatchLookupResponse represents batch query responses.
type BatchLookupResponse struct {
	Responses []*MatchPrefixResponse `json:"responses,omitempty"`
}

// PrefixOSServiceClient is the client API for PrefixOSService.
type PrefixOSServiceClient interface {
	MatchPrefix(ctx context.Context, in *MatchPrefixRequest, opts ...grpc.CallOption) (*MatchPrefixResponse, error)
	Insert(ctx context.Context, in *InsertRequest, opts ...grpc.CallOption) (*InsertResponse, error)
}

type prefixOSServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPrefixOSServiceClient(cc grpc.ClientConnInterface) PrefixOSServiceClient {
	return &prefixOSServiceClient{cc}
}

func (c *prefixOSServiceClient) MatchPrefix(ctx context.Context, in *MatchPrefixRequest, opts ...grpc.CallOption) (*MatchPrefixResponse, error) {
	out := new(MatchPrefixResponse)
	callOpts := append([]grpc.CallOption{grpc.CallContentSubtype("json")}, opts...)
	err := c.cc.Invoke(ctx, "/prefixos.v1.PrefixOSService/MatchPrefix", in, out, callOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *prefixOSServiceClient) Insert(ctx context.Context, in *InsertRequest, opts ...grpc.CallOption) (*InsertResponse, error) {
	out := new(InsertResponse)
	callOpts := append([]grpc.CallOption{grpc.CallContentSubtype("json")}, opts...)
	err := c.cc.Invoke(ctx, "/prefixos.v1.PrefixOSService/Insert", in, out, callOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PrefixOSServiceServer is the server API for PrefixOSService.
type PrefixOSServiceServer interface {
	MatchPrefix(context.Context, *MatchPrefixRequest) (*MatchPrefixResponse, error)
	Insert(context.Context, *InsertRequest) (*InsertResponse, error)
	mustEmbedUnimplementedPrefixOSServiceServer()
}

type UnimplementedPrefixOSServiceServer struct{}

func (UnimplementedPrefixOSServiceServer) MatchPrefix(context.Context, *MatchPrefixRequest) (*MatchPrefixResponse, error) {
	return nil, nil
}
func (UnimplementedPrefixOSServiceServer) Insert(context.Context, *InsertRequest) (*InsertResponse, error) {
	return nil, nil
}
func (UnimplementedPrefixOSServiceServer) mustEmbedUnimplementedPrefixOSServiceServer() {}

func RegisterPrefixOSServiceServer(s grpc.ServiceRegistrar, srv PrefixOSServiceServer) {
	s.RegisterService(&PrefixOSService_ServiceDesc, srv)
}

var PrefixOSService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "prefixos.v1.PrefixOSService",
	HandlerType: (*PrefixOSServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MatchPrefix",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(MatchPrefixRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(PrefixOSServiceServer).MatchPrefix(ctx, in)
				}
				info := &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/prefixos.v1.PrefixOSService/MatchPrefix",
				}
				handler := func(ctx context.Context, req interface{}) (interface{}, error) {
					return srv.(PrefixOSServiceServer).MatchPrefix(ctx, req.(*MatchPrefixRequest))
				}
				return interceptor(ctx, in, info, handler)
			},
		},
		{
			MethodName: "Insert",
			Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
				in := new(InsertRequest)
				if err := dec(in); err != nil {
					return nil, err
				}
				if interceptor == nil {
					return srv.(PrefixOSServiceServer).Insert(ctx, in)
				}
				info := &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: "/prefixos.v1.PrefixOSService/Insert",
				}
				handler := func(ctx context.Context, req interface{}) (interface{}, error) {
					return srv.(PrefixOSServiceServer).Insert(ctx, req.(*InsertRequest))
				}
				return interceptor(ctx, in, info, handler)
			},
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/v1/prefixos.proto",
}
