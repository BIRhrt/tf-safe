package integration

import (
	"os"
	"path/filepath"
	"testing"

	"tf-safe/internal/config"
	"tf-safe/pkg/types"
)

func TestConfigurationMerging(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "tf-safe-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create project-level config
	projectConfig := `
local:
  enabled: true
  path: ".tfstate_snapshots"
  retention_count: 8

remote:
  enabled: true
  provider: "s3"
  bucket: "project-bucket"
  region: "us-east-1"

encryption:
  provider: "aes"
  passphrase: "project-passphrase"
`

	projectConfigPath := filepath.Join(tempDir, ".tf-safe.yaml")
	if err := os.WriteFile(projectConfigPath, []byte(projectConfig), 0644); err != nil {
		t.Fatalf("Failed to create project config: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Test configuration loading and merging
	manager := config.NewManager()
	manager.AddSource(config.NewFileSource(".tf-safe.yaml", 20, "project config"))

	loadedConfig, err := manager.Load()
	if err != nil {
		t.Fatalf("Failed to load configuration: %v", err)
	}

	// Verify project-level settings override defaults
	if loadedConfig.Local.RetentionCount != 8 {
		t.Errorf("Expected retention count 8, got %d", loadedConfig.Local.RetentionCount)
	}

	if loadedConfig.Remote.Bucket != "project-bucket" {
		t.Errorf("Expected bucket 'project-bucket', got '%s'", loadedConfig.Remote.Bucket)
	}

	if loadedConfig.Remote.Region != "us-east-1" {
		t.Errorf("Expected region 'us-east-1', got '%s'", loadedConfig.Remote.Region)
	}

	if loadedConfig.Encryption.Passphrase != "project-passphrase" {
		t.Errorf("Expected passphrase 'project-passphrase', got '%s'", loadedConfig.Encryption.Passphrase)
	}

	// Test configuration validation
	err = manager.Validate(loadedConfig)
	if err != nil {
		t.Errorf("Configuration validation failed: %v", err)
	}

	t.Log("Configuration merging test completed successfully")
}

func TestConfigurationValidation(t *testing.T) {
	manager := config.NewManager()

	testCases := []struct {
		name        string
		config      *types.Config
		expectError bool
		description string
	}{
		{
			name: "Valid minimal config",
			config: &types.Config{
				Local: types.LocalConfig{
					Enabled:        true,
					Path:           ".tfstate_snapshots",
					RetentionCount: 5,
				},
				Encryption: types.EncryptionConfig{
					Provider: "none",
				},
				Retention: types.RetentionConfig{
					LocalCount:  5,
					RemoteCount: 10,
					MaxAgeDays:  30,
				},
			},
			expectError: false,
			description: "Minimal valid configuration should pass validation",
		},
		{
			name: "Invalid retention count",
			config: &types.Config{
				Local: types.LocalConfig{
					Enabled:        true,
					Path:           ".tfstate_snapshots",
					RetentionCount: 0, // Invalid
				},
				Retention: types.RetentionConfig{
					LocalCount: 0, // Invalid
				},
			},
			expectError: true,
			description: "Zero retention count should fail validation",
		},
		{
			name: "S3 config without bucket",
			config: &types.Config{
				Remote: types.RemoteConfig{
					Enabled:  true,
					Provider: "s3",
					Bucket:   "", // Missing bucket
					Region:   "us-west-2",
				},
			},
			expectError: true,
			description: "S3 configuration without bucket should fail validation",
		},
		{
			name: "AES encryption without passphrase",
			config: &types.Config{
				Encryption: types.EncryptionConfig{
					Provider:   "aes",
					Passphrase: "", // Missing passphrase
				},
			},
			expectError: true,
			description: "AES encryption without passphrase should fail validation",
		},
		{
			name: "Valid S3 and encryption config",
			config: &types.Config{
				Local: types.LocalConfig{
					Enabled:        true,
					Path:           ".tfstate_snapshots",
					RetentionCount: 10,
				},
				Remote: types.RemoteConfig{
					Enabled:  true,
					Provider: "s3",
					Bucket:   "my-backup-bucket",
					Region:   "us-west-2",
					Prefix:   "terraform-backups/",
				},
				Encryption: types.EncryptionConfig{
					Provider:   "aes",
					Passphrase: "secure-passphrase-123",
				},
				Retention: types.RetentionConfig{
					LocalCount:  10,
					RemoteCount: 50,
					MaxAgeDays:  90,
				},
			},
			expectError: false,
			description: "Complete valid configuration should pass validation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.Validate(tc.config)
			
			if tc.expectError && err == nil {
				t.Errorf("%s: Expected validation error but got none", tc.description)
			}
			
			if !tc.expectError && err != nil {
				t.Errorf("%s: Unexpected validation error: %v", tc.description, err)
			}
		})
	}

	t.Log("Configuration validation test completed successfully")
}

func TestConfigurationSaveLoad(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "tf-safe-config-save-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	manager := config.NewManager()

	// Create test configuration
	originalConfig := &types.Config{
		Local: types.LocalConfig{
			Enabled:        true,
			Path:           ".tfstate_snapshots",
			RetentionCount: 15,
		},
		Remote: types.RemoteConfig{
			Enabled:  true,
			Provider: "s3",
			Bucket:   "test-save-bucket",
			Region:   "eu-west-1",
			Prefix:   "backups/",
		},
		Encryption: types.EncryptionConfig{
			Provider:   "aes",
			Passphrase: "test-save-passphrase",
		},
		Retention: types.RetentionConfig{
			LocalCount:  15,
			RemoteCount: 100,
			MaxAgeDays:  180,
		},
	}

	// Save configuration
	configPath := filepath.Join(tempDir, "test-config.yaml")
	err = manager.Save(originalConfig, configPath)
	if err != nil {
		t.Fatalf("Failed to save configuration: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Configuration file was not created")
	}

	// Load configuration back
	loadManager := config.NewManager()
	loadManager.AddSource(config.NewFileSource(configPath, 20, "test config"))

	loadedConfig, err := loadManager.Load()
	if err != nil {
		t.Fatalf("Failed to load saved configuration: %v", err)
	}

	// Verify loaded configuration matches original
	if loadedConfig.Local.RetentionCount != originalConfig.Local.RetentionCount {
		t.Errorf("Local retention count mismatch: expected %d, got %d",
			originalConfig.Local.RetentionCount, loadedConfig.Local.RetentionCount)
	}

	if loadedConfig.Remote.Bucket != originalConfig.Remote.Bucket {
		t.Errorf("Remote bucket mismatch: expected %s, got %s",
			originalConfig.Remote.Bucket, loadedConfig.Remote.Bucket)
	}

	if loadedConfig.Encryption.Passphrase != originalConfig.Encryption.Passphrase {
		t.Errorf("Encryption passphrase mismatch: expected %s, got %s",
			originalConfig.Encryption.Passphrase, loadedConfig.Encryption.Passphrase)
	}

	if loadedConfig.Retention.MaxAgeDays != originalConfig.Retention.MaxAgeDays {
		t.Errorf("Max age days mismatch: expected %d, got %d",
			originalConfig.Retention.MaxAgeDays, loadedConfig.Retention.MaxAgeDays)
	}

	t.Log("Configuration save/load test completed successfully")
}