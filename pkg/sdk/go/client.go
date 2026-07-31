package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "prefixos/proto/v1"
)

// ClientOptions configures the PrefixOS Go SDK Client.
type ClientOptions struct {
	Address      string
	Timeout      time.Duration
	MaxRetries   int
	RetryBackoff time.Duration
}

// Client wraps the gRPC client with connection management, retries, timeouts, and error handling.
type Client struct {
	conn       *grpc.ClientConn
	grpcClient pb.PrefixOSServiceClient
	opts       ClientOptions
}

// DefaultOptions returns safe client configuration defaults.
func DefaultOptions(address string) ClientOptions {
	return ClientOptions{
		Address:      address,
		Timeout:      5 * time.Second,
		MaxRetries:   3,
		RetryBackoff: 100 * time.Millisecond,
	}
}

// NewClient establishes a connection to a PrefixOS gRPC node.
func NewClient(opts ClientOptions) (*Client, error) {
	if opts.Address == "" {
		opts.Address = "127.0.0.1:50051"
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Second
	}
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = 3
	}

	conn, err := grpc.NewClient(
		opts.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to PrefixOS node at %s: %w", opts.Address, err)
	}

	return &Client{
		conn:       conn,
		grpcClient: pb.NewPrefixOSServiceClient(conn),
		opts:       opts,
	}, nil
}

// Close closes the underlying gRPC client connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// isRetryable determines if a gRPC error is transient and safe to retry.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		return true // Retry generic network errors
	}
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}

// MatchPrefix queries PrefixOS for the longest cached token sequence match.
func (c *Client) MatchPrefix(ctx context.Context, tokens []int32) (int, []int32, error) {
	req := &pb.MatchPrefixRequest{
		Tokens: tokens,
	}

	var resp *pb.MatchPrefixResponse
	var err error

	for attempt := 0; attempt <= c.opts.MaxRetries; attempt++ {
		callCtx, cancel := context.WithTimeout(ctx, c.opts.Timeout)
		resp, err = c.grpcClient.MatchPrefix(callCtx, req)
		cancel()

		if err == nil {
			return int(resp.MatchedLength), resp.BlockIds, nil
		}

		if !isRetryable(err) {
			return 0, nil, fmt.Errorf("non-retryable MatchPrefix error: %w", err)
		}

		if attempt < c.opts.MaxRetries {
			time.Sleep(c.opts.RetryBackoff * time.Duration(1<<attempt))
		}
	}

	return 0, nil, fmt.Errorf("MatchPrefix failed after %d retries: %w", c.opts.MaxRetries, err)
}

// Insert registers a token sequence into PrefixOS and allocates slab memory block handles.
func (c *Client) Insert(ctx context.Context, tokens []int32) (bool, []int32, error) {
	req := &pb.InsertRequest{
		Tokens: tokens,
	}

	var resp *pb.InsertResponse
	var err error

	for attempt := 0; attempt <= c.opts.MaxRetries; attempt++ {
		callCtx, cancel := context.WithTimeout(ctx, c.opts.Timeout)
		resp, err = c.grpcClient.Insert(callCtx, req)
		cancel()

		if err == nil {
			return resp.Success, resp.AllocatedBlocks, nil
		}

		if !isRetryable(err) {
			return false, nil, fmt.Errorf("non-retryable Insert error: %w", err)
		}

		if attempt < c.opts.MaxRetries {
			time.Sleep(c.opts.RetryBackoff * time.Duration(1<<attempt))
		}
	}

	return false, nil, fmt.Errorf("Insert failed after %d retries: %w", c.opts.MaxRetries, err)
}
