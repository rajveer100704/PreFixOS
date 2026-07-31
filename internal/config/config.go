package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the complete PrefixOS engine configuration parameters.
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Memory      MemoryConfig      `yaml:"memory"`
	Radix       RadixConfig       `yaml:"radix"`
	Eviction    EvictionConfig    `yaml:"eviction"`
	Persistence PersistenceConfig `yaml:"persistence"`
}

// ServerConfig configures networking and admin endpoints.
type ServerConfig struct {
	GRPCPort int    `yaml:"grpc_port"`
	HTTPPort int    `yaml:"http_port"`
	Host     string `yaml:"host"`
}

// MemoryConfig configures the zero-allocation slab allocator.
type MemoryConfig struct {
	TotalBlocks    int    `yaml:"total_blocks"`
	BlockSize      int    `yaml:"block_size"`
	NUMANodeID     int    `yaml:"numa_node_id"`
	DefragInterval string `yaml:"defrag_interval"`
}

// RadixConfig configures the 64-way sharded prefix tree.
type RadixConfig struct {
	ShardCount int `yaml:"shard_count"`
}

// EvictionConfig configures cache eviction policies.
type EvictionConfig struct {
	Policy        string  `yaml:"policy"` // "lru", "slru"
	Watermark     float64 `yaml:"watermark"`
	SLRUProbation float64 `yaml:"slru_probation_ratio"`
}

// PersistenceConfig configures WAL and snapshotting.
type PersistenceConfig struct {
	Enabled       bool   `yaml:"enabled"`
	WALPath       string `yaml:"wal_path"`
	SnapshotPath  string `yaml:"snapshot_path"`
	SyncImmediate bool   `yaml:"sync_immediate"`
}

// DefaultConfig returns safe production default configurations.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			GRPCPort: 50051,
			HTTPPort: 8080,
			Host:     "0.0.0.0",
		},
		Memory: MemoryConfig{
			TotalBlocks: 65536, // 64k blocks * 16 tokens = ~1M tokens
			BlockSize:   16,
			NUMANodeID:  0,
		},
		Radix: RadixConfig{
			ShardCount: 64,
		},
		Eviction: EvictionConfig{
			Policy:        "lru",
			Watermark:     0.90,
			SLRUProbation: 0.20,
		},
		Persistence: PersistenceConfig{
			Enabled:       true,
			WALPath:       "data/prefixos.wal",
			SnapshotPath:  "data/snapshots",
			SyncImmediate: false,
		},
	}
}

// LoadConfig loads and parses a YAML configuration file, falling back to defaults for missing fields.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate verifies that configuration settings are within valid bounds.
func (c *Config) Validate() error {
	if c.Memory.TotalBlocks <= 0 {
		return fmt.Errorf("memory.total_blocks must be positive")
	}
	if c.Memory.BlockSize <= 0 {
		return fmt.Errorf("memory.block_size must be positive")
	}
	if c.Radix.ShardCount <= 0 || c.Radix.ShardCount&(c.Radix.ShardCount-1) != 0 {
		return fmt.Errorf("radix.shard_count must be a power of 2")
	}
	if c.Eviction.Watermark <= 0 || c.Eviction.Watermark > 1.0 {
		return fmt.Errorf("eviction.watermark must be between 0.0 and 1.0")
	}
	return nil
}
