package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
	"tf-safe/pkg/types"
)

// Manager implements the ConfigManager interface
type Manager struct {
	sources []ConfigSource
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		sources: make([]ConfigSource, 0),
	}
}

// AddSource adds a configuration source to the manager
func (m *Manager) AddSource(source ConfigSource) {
	m.sources = append(m.sources, source)
}

// Load loads configuration from all sources and merges them according to priority
func (m *Manager) Load() (*types.Config, error) {
	// Start with default configuration
	config := DefaultConfig()
	
	// Sort sources by priority (lowest to highest)
	sortedSources := make([]ConfigSource, len(m.sources))
	copy(sortedSources, m.sources)
	sort.Slice(sortedSources, func(i, j int) bool {
		return sortedSources[i].GetPriority() < sortedSources[j].GetPriority()
	})
	
	// Merge configurations from each source
	for _, source := range sortedSources {
		sourceConfig, err := source.Load()
		if err != nil {
			// Log warning but continue with other sources
			continue
		}
		
		if sourceConfig != nil {
			config = mergeConfigs(config, sourceConfig)
		}
	}
	
	return config, nil
}

// Validate validates the configuration for correctness
func (m *Manager) Validate(config *types.Config) error {
	return config.Validate()
}

// GetStorageConfig returns the local storage configuration
func (m *Manager) GetStorageConfig() types.LocalConfig {
	config, _ := m.Load()
	return config.Local
}

// GetRemoteConfig returns the remote storage configuration
func (m *Manager) GetRemoteConfig() types.RemoteConfig {
	config, _ := m.Load()
	return config.Remote
}

// GetEncryptionConfig returns the encryption configuration
func (m *Manager) GetEncryptionConfig() types.EncryptionConfig {
	config, _ := m.Load()
	return config.Encryption
}

// GetRetentionConfig returns the retention configuration
func (m *Manager) GetRetentionConfig() types.RetentionConfig {
	config, _ := m.Load()
	return config.Retention
}

// Save saves the configuration to a file
func (m *Manager) Save(config *types.Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Write configuration file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}
	
	return nil
}

// CreateDefault creates a default configuration
func (m *Manager) CreateDefault() *types.Config {
	return DefaultConfig()
}

// mergeConfigs merges two configurations, with the second taking priority
func mergeConfigs(base, override *types.Config) *types.Config {
	result := *base // Copy base config
	
	// Merge local config
	if override.Local.Path != "" {
		result.Local.Path = override.Local.Path
	}
	if override.Local.RetentionCount > 0 {
		result.Local.RetentionCount = override.Local.RetentionCount
	}
	// Enabled is a boolean, so we need to check if it was explicitly set
	result.Local.Enabled = override.Local.Enabled
	
	// Merge remote config
	if override.Remote.Provider != "" {
		result.Remote.Provider = override.Remote.Provider
	}
	if override.Remote.Bucket != "" {
		result.Remote.Bucket = override.Remote.Bucket
	}
	if override.Remote.Region != "" {
		result.Remote.Region = override.Remote.Region
	}
	if override.Remote.Prefix != "" {
		result.Remote.Prefix = override.Remote.Prefix
	}
	result.Remote.Enabled = override.Remote.Enabled
	
	// Merge encryption config
	if override.Encryption.Provider != "" {
		result.Encryption.Provider = override.Encryption.Provider
	}
	if override.Encryption.KMSKeyID != "" {
		result.Encryption.KMSKeyID = override.Encryption.KMSKeyID
	}
	if override.Encryption.Passphrase != "" {
		result.Encryption.Passphrase = override.Encryption.Passphrase
	}
	
	// Merge retention config
	if override.Retention.LocalCount > 0 {
		result.Retention.LocalCount = override.Retention.LocalCount
	}
	if override.Retention.RemoteCount > 0 {
		result.Retention.RemoteCount = override.Retention.RemoteCount
	}
	if override.Retention.MaxAgeDays > 0 {
		result.Retention.MaxAgeDays = override.Retention.MaxAgeDays
	}
	
	// Merge logging config
	if override.Logging.Level != "" {
		result.Logging.Level = override.Logging.Level
	}
	if override.Logging.Format != "" {
		result.Logging.Format = override.Logging.Format
	}
	
	return &result
}

// FileSource represents a YAML configuration file source
type FileSource struct {
	path     string
	priority int
	name     string
}

// NewFileSource creates a new file-based configuration source
func NewFileSource(path string, priority int, name string) *FileSource {
	return &FileSource{
		path:     path,
		priority: priority,
		name:     name,
	}
}

// Load loads configuration from the file
func (f *FileSource) Load() (*types.Config, error) {
	// Expand home directory if needed
	path := f.path
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}
	
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil // File doesn't exist, return nil config
	}
	
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	
	// Parse YAML
	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}
	
	return &config, nil
}

// GetPriority returns the priority of this source
func (f *FileSource) GetPriority() int {
	return f.priority
}

// GetName returns the name of this source
func (f *FileSource) GetName() string {
	return f.name
}

// FlagSource represents CLI flag-based configuration source
type FlagSource struct {
	config   *types.Config
	priority int
}

// NewFlagSource creates a new flag-based configuration source
func NewFlagSource(config *types.Config, priority int) *FlagSource {
	return &FlagSource{
		config:   config,
		priority: priority,
	}
}

// Load returns the flag-based configuration
func (f *FlagSource) Load() (*types.Config, error) {
	return f.config, nil
}

// GetPriority returns the priority of this source
func (f *FlagSource) GetPriority() int {
	return f.priority
}

// GetName returns the name of this source
func (f *FlagSource) GetName() string {
	return "command-line flags"
}

// LoadConfiguration is a convenience function to load configuration with standard sources
func LoadConfiguration() (*types.Config, error) {
	manager := NewManager()
	
	// Add configuration sources in priority order (lowest to highest)
	// 1. Global configuration (priority 10)
	manager.AddSource(NewFileSource("~/.tf-safe/config.yaml", 10, "global config"))
	
	// 2. Project configuration (priority 20)
	manager.AddSource(NewFileSource(".tf-safe.yaml", 20, "project config"))
	
	// Note: CLI flags would be added with priority 30 when available
	
	config, err := manager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	
	// Validate the final configuration
	validator := NewValidator()
	if err := validator.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	
	return config, nil
}