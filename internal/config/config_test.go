package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("expected non-nil default config")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config validation failed: %v", err)
	}
	if cfg.Server.GRPCPort != 50051 {
		t.Errorf("expected GRPC port 50051, got %d", cfg.Server.GRPCPort)
	}
	if cfg.Radix.ShardCount != 64 {
		t.Errorf("expected 64 shards, got %d", cfg.Radix.ShardCount)
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	cfg, err := LoadConfig("non_existent_file.yaml")
	if err != nil {
		t.Fatalf("expected fallback to default config, got error: %v", err)
	}
	if cfg.Memory.TotalBlocks != 65536 {
		t.Errorf("expected default total blocks 65536, got %d", cfg.Memory.TotalBlocks)
	}
}

func TestLoadConfig_CustomYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := []byte(`
server:
  grpc_port: 9090
  http_port: 9091
  host: "127.0.0.1"
memory:
  total_blocks: 1024
  block_size: 16
  numa_node_id: 1
radix:
  shard_count: 32
eviction:
  policy: "slru"
  watermark: 0.85
  slru_probation_ratio: 0.25
persistence:
  enabled: true
  wal_path: "/tmp/prefixos.wal"
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatalf("failed to write tmp config file: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load custom config: %v", err)
	}

	if cfg.Server.GRPCPort != 9090 {
		t.Errorf("expected gRPC port 9090, got %d", cfg.Server.GRPCPort)
	}
	if cfg.Memory.TotalBlocks != 1024 {
		t.Errorf("expected 1024 blocks, got %d", cfg.Memory.TotalBlocks)
	}
	if cfg.Radix.ShardCount != 32 {
		t.Errorf("expected 32 shards, got %d", cfg.Radix.ShardCount)
	}
	if cfg.Eviction.Policy != "slru" {
		t.Errorf("expected slru policy, got %s", cfg.Eviction.Policy)
	}
}

func TestValidationErrors(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Memory.TotalBlocks = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for total_blocks <= 0")
	}

	cfg = DefaultConfig()
	cfg.Radix.ShardCount = 63 // Not a power of 2
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for non-power-of-2 shard count")
	}
}
