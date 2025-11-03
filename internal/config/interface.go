package config

import "tf-safe/pkg/types"

// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	// Load loads configuration from all sources and merges them
	Load() (*types.Config, error)
	
	// Validate validates the configuration for correctness
	Validate(config *types.Config) error
	
	// GetStorageConfig returns the storage configuration
	GetStorageConfig() types.LocalConfig
	
	// GetRemoteConfig returns the remote storage configuration
	GetRemoteConfig() types.RemoteConfig
	
	// GetEncryptionConfig returns the encryption configuration
	GetEncryptionConfig() types.EncryptionConfig
	
	// GetRetentionConfig returns the retention configuration
	GetRetentionConfig() types.RetentionConfig
	
	// Save saves the configuration to a file
	Save(config *types.Config, path string) error
	
	// CreateDefault creates a default configuration
	CreateDefault() *types.Config
}

// ConfigSource represents a source of configuration data
type ConfigSource interface {
	// Load loads configuration from this source
	Load() (*types.Config, error)
	
	// GetPriority returns the priority of this source (higher = more important)
	GetPriority() int
	
	// GetName returns a human-readable name for this source
	GetName() string
}