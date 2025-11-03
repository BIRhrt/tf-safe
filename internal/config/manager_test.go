package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tf-safe/pkg/types"
)

func TestManager_Load(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "tf-safe-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test config file
	configContent := `
local:
  enabled: true
  path: ".tfstate_snapshots"
  retention_count: 5

remote:
  enabled: false
  provider: "s3"
  bucket: "test-bucket"
  region: "us-west-2"

encryption:
  provider: "aes"
  passphrase: "test-passphrase"

retention:
  local_count: 5
  remote_count: 20
  max_age_days: 30
`

	configPath := filepath.Join(tempDir, ".tf-safe.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	manager := NewManager()
	manager.AddSource(NewFileSource(".tf-safe.yaml", 20, "project config"))
	config, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded configuration
	if !config.Local.Enabled {
		t.Error("Expected local.enabled to be true")
	}
	if config.Local.Path != ".tfstate_snapshots" {
		t.Errorf("Expected local.path to be '.tfstate_snapshots', got '%s'", config.Local.Path)
	}
	if config.Local.RetentionCount != 5 {
		t.Errorf("Expected local.retention_count to be 5, got %d", config.Local.RetentionCount)
	}
	if config.Remote.Enabled {
		t.Error("Expected remote.enabled to be false")
	}
	if config.Encryption.Provider != "aes" {
		t.Errorf("Expected encryption.provider to be 'aes', got '%s'", config.Encryption.Provider)
	}
}

func TestManager_Validate(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name        string
		config      *types.Config
		expectError bool
	}{
		{
			name: "Valid configuration",
			config: &types.Config{
				Local: types.LocalConfig{
					Enabled:        true,
					Path:           ".tfstate_snapshots",
					RetentionCount: 5,
				},
				Remote: types.RemoteConfig{
					Enabled:  false,
					Provider: "s3",
				},
				Encryption: types.EncryptionConfig{
					Provider: "none",
				},
				Retention: types.RetentionConfig{
					LocalCount:  5,
					RemoteCount: 20,
					MaxAgeDays:  30,
				},
			},
			expectError: false,
		},
		{
			name: "Invalid retention count",
			config: &types.Config{
				Local: types.LocalConfig{
					Enabled:        true,
					Path:           ".tfstate_snapshots",
					RetentionCount: 0,
				},
				Retention: types.RetentionConfig{
					LocalCount: 0,
				},
			},
			expectError: true,
		},
		{
			name: "S3 without bucket",
			config: &types.Config{
				Remote: types.RemoteConfig{
					Enabled:  true,
					Provider: "s3",
					Bucket:   "",
				},
			},
			expectError: true,
		},
		{
			name: "AES without passphrase",
			config: &types.Config{
				Encryption: types.EncryptionConfig{
					Provider:   "aes",
					Passphrase: "",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.Validate(tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestManager_CreateDefault(t *testing.T) {
	manager := NewManager()
	config := manager.CreateDefault()

	if config == nil {
		t.Fatal("Expected default config but got nil")
	}

	// Verify default values
	if !config.Local.Enabled {
		t.Error("Expected default local.enabled to be true")
	}
	if config.Local.Path != ".tfstate_snapshots" {
		t.Errorf("Expected default local.path to be '.tfstate_snapshots', got '%s'", config.Local.Path)
	}
	if config.Local.RetentionCount != 10 {
		t.Errorf("Expected default local.retention_count to be 10, got %d", config.Local.RetentionCount)
	}
	if config.Remote.Enabled {
		t.Error("Expected default remote.enabled to be false")
	}
	if config.Encryption.Provider != "aes" {
		t.Errorf("Expected default encryption.provider to be 'aes', got '%s'", config.Encryption.Provider)
	}
}

func TestManager_Save(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf-safe-config-save-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager := NewManager()
	config := manager.CreateDefault()
	
	// Modify some values
	config.Local.RetentionCount = 15
	config.Encryption.Provider = "aes"
	config.Encryption.Passphrase = "test-passphrase"

	configPath := filepath.Join(tempDir, "test-config.yaml")
	err = manager.Save(config, configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load the saved config and verify
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	configStr := string(content)
	if !strings.Contains(configStr, "retention_count: 15") {
		t.Error("Saved config doesn't contain expected retention_count")
	}
	if !strings.Contains(configStr, "provider: aes") {
		t.Error("Saved config doesn't contain expected encryption provider")
	}
}