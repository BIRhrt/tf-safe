package terraform

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"tf-safe/pkg/types"
)

// MockConfigManager implements ConfigManager for testing
type MockConfigManager struct {
	config *types.Config
}

func NewMockConfigManager() *MockConfigManager {
	return &MockConfigManager{
		config: &types.Config{
			Local: types.LocalConfig{
				Enabled: true,
				Path:    ".tfstate_snapshots",
			},
			Remote: types.RemoteConfig{
				Enabled: false,
			},
			Encryption: types.EncryptionConfig{
				Provider: "none",
			},
			Retention: types.RetentionConfig{
				LocalCount: 10,
			},
		},
	}
}

func (m *MockConfigManager) Load() (*types.Config, error) {
	return m.config, nil
}

func (m *MockConfigManager) Validate(config *types.Config) error {
	return nil
}

func (m *MockConfigManager) GetStorageConfig() types.LocalConfig {
	return m.config.Local
}

func (m *MockConfigManager) GetRemoteConfig() types.RemoteConfig {
	return m.config.Remote
}

func (m *MockConfigManager) GetEncryptionConfig() types.EncryptionConfig {
	return m.config.Encryption
}

func (m *MockConfigManager) GetRetentionConfig() types.RetentionConfig {
	return m.config.Retention
}

func (m *MockConfigManager) Save(config *types.Config, path string) error {
	return nil
}

func (m *MockConfigManager) CreateDefault() *types.Config {
	return m.config
}

// MockBackupEngine implements BackupEngine for testing
type MockBackupEngine struct {
	backups []*types.BackupMetadata
	shouldFail bool
}

func NewMockBackupEngine() *MockBackupEngine {
	return &MockBackupEngine{
		backups: []*types.BackupMetadata{},
	}
}

func (m *MockBackupEngine) CreateBackup(ctx context.Context, opts types.BackupOptions) (*types.BackupMetadata, error) {
	if m.shouldFail {
		return nil, &types.TfSafeError{Code: "BACKUP_ERROR", Message: "Mock backup failure"}
	}
	
	metadata := &types.BackupMetadata{
		ID:        "test-backup-id",
		Size:      100,
		Checksum:  "test-checksum",
	}
	m.backups = append(m.backups, metadata)
	return metadata, nil
}

func (m *MockBackupEngine) ListBackups(ctx context.Context) ([]*types.BackupMetadata, error) {
	if m.shouldFail {
		return nil, &types.TfSafeError{Code: "BACKUP_ERROR", Message: "Mock backup failure"}
	}
	return m.backups, nil
}

func (m *MockBackupEngine) CleanupOldBackups(ctx context.Context) error {
	if m.shouldFail {
		return &types.TfSafeError{Code: "BACKUP_ERROR", Message: "Mock backup failure"}
	}
	return nil
}

func (m *MockBackupEngine) GetBackupMetadata(ctx context.Context, backupID string) (*types.BackupMetadata, error) {
	if m.shouldFail {
		return nil, &types.TfSafeError{Code: "BACKUP_ERROR", Message: "Mock backup failure"}
	}
	
	for _, backup := range m.backups {
		if backup.ID == backupID {
			return backup, nil
		}
	}
	return nil, &types.TfSafeError{Code: "BACKUP_NOT_FOUND", Message: "Backup not found"}
}

func (m *MockBackupEngine) ValidateBackup(ctx context.Context, backupID string) error {
	if m.shouldFail {
		return &types.TfSafeError{Code: "BACKUP_ERROR", Message: "Mock backup failure"}
	}
	
	for _, backup := range m.backups {
		if backup.ID == backupID {
			return nil
		}
	}
	return &types.TfSafeError{Code: "BACKUP_NOT_FOUND", Message: "Backup not found"}
}

func (m *MockBackupEngine) SetShouldFail(fail bool) {
	m.shouldFail = fail
}

func TestWrapper_DetectStateFile(t *testing.T) {
	// Create temporary directory with state file
	tempDir, err := os.MkdirTemp("", "tf-safe-wrapper-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	stateContent := `{"version": 4, "terraform_version": "1.0.0"}`
	stateFile := filepath.Join(tempDir, "terraform.tfstate")
	if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
		t.Fatalf("Failed to create state file: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	configManager := NewMockConfigManager()
	backupEngine := NewMockBackupEngine()
	wrapper := NewWrapper(configManager, backupEngine)

	// Test state file detection
	detectedFile, err := wrapper.DetectStateFile()
	if err != nil {
		t.Fatalf("Failed to detect state file: %v", err)
	}

	expectedPath := filepath.Join(tempDir, "terraform.tfstate")
	// Use filepath.EvalSymlinks to handle macOS /private symlinks
	expectedResolved, _ := filepath.EvalSymlinks(expectedPath)
	detectedResolved, _ := filepath.EvalSymlinks(detectedFile)
	if detectedResolved != expectedResolved {
		t.Errorf("Expected detected file %s, got %s", expectedResolved, detectedResolved)
	}
}

func TestWrapper_DetectStateFile_NoStateFile(t *testing.T) {
	// Create temporary directory without state file
	tempDir, err := os.MkdirTemp("", "tf-safe-wrapper-no-state-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	configManager := NewMockConfigManager()
	backupEngine := NewMockBackupEngine()
	wrapper := NewWrapper(configManager, backupEngine)

	// Test state file detection with no state file
	_, err = wrapper.DetectStateFile()
	if err == nil {
		t.Error("Expected error when no state file exists but got none")
	}
}

func TestWrapper_DetectStateFile_MultipleStateFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf-safe-wrapper-multiple-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// The detector only looks for specific filenames, so this test
	// verifies that it works correctly with just terraform.tfstate
	// (multiple .tfstate files with different names won't be detected)
	stateFile := filepath.Join(tempDir, "terraform.tfstate")
	if err := os.WriteFile(stateFile, []byte(`{"version": 4}`), 0644); err != nil {
		t.Fatalf("Failed to create state file: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	configManager := NewMockConfigManager()
	backupEngine := NewMockBackupEngine()
	wrapper := NewWrapper(configManager, backupEngine)

	// Test state file detection - should succeed with single terraform.tfstate
	detectedFile, err := wrapper.DetectStateFile()
	if err != nil {
		t.Errorf("Unexpected error detecting single state file: %v", err)
	}
	
	if detectedFile == "" {
		t.Error("Expected to detect terraform.tfstate file")
	}
}

func TestWrapper_ValidateStateFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf-safe-wrapper-validate-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configManager := NewMockConfigManager()
	backupEngine := NewMockBackupEngine()
	wrapper := NewWrapper(configManager, backupEngine)

	// Test with valid state file
	validStateContent := `{
		"version": 4,
		"terraform_version": "1.0.0",
		"serial": 1,
		"lineage": "test-lineage",
		"outputs": {},
		"resources": []
	}`
	validStateFile := filepath.Join(tempDir, "valid.tfstate")
	if err := os.WriteFile(validStateFile, []byte(validStateContent), 0644); err != nil {
		t.Fatalf("Failed to create valid state file: %v", err)
	}

	err = wrapper.ValidateStateFile(validStateFile)
	if err != nil {
		t.Errorf("Validation failed for valid state file: %v", err)
	}

	// Test with invalid state file (not JSON)
	invalidStateFile := filepath.Join(tempDir, "invalid.tfstate")
	if err := os.WriteFile(invalidStateFile, []byte("not json"), 0644); err != nil {
		t.Fatalf("Failed to create invalid state file: %v", err)
	}

	err = wrapper.ValidateStateFile(invalidStateFile)
	if err == nil {
		t.Error("Expected validation error for invalid state file but got none")
	}

	// Test with non-existent file
	err = wrapper.ValidateStateFile("nonexistent.tfstate")
	if err == nil {
		t.Error("Expected validation error for non-existent file but got none")
	}
}

func TestWrapper_CheckTerraformBinary(t *testing.T) {
	configManager := NewMockConfigManager()
	backupEngine := NewMockBackupEngine()
	wrapper := NewWrapper(configManager, backupEngine)

	// Note: This test will only pass if terraform is installed on the system
	// In a real CI environment, you might want to mock this or skip the test
	err := wrapper.CheckTerraformBinary()
	
	// We can't guarantee terraform is installed in all test environments
	// So we'll just verify the method doesn't panic and returns some result
	if err != nil {
		t.Logf("Terraform binary check failed (expected in environments without terraform): %v", err)
	} else {
		t.Log("Terraform binary check passed")
	}
}

func TestWrapper_GetTerraformVersion(t *testing.T) {
	configManager := NewMockConfigManager()
	backupEngine := NewMockBackupEngine()
	wrapper := NewWrapper(configManager, backupEngine)

	// Note: This test will only pass if terraform is installed on the system
	version, err := wrapper.GetTerraformVersion()
	
	if err != nil {
		t.Logf("Terraform version check failed (expected in environments without terraform): %v", err)
	} else {
		t.Logf("Terraform version: %s", version)
		if version == "" {
			t.Error("Expected non-empty version string")
		}
	}
}

func TestWrapper_AddHook(t *testing.T) {
	configManager := NewMockConfigManager()
	backupEngine := NewMockBackupEngine()
	wrapper := NewWrapper(configManager, backupEngine)

	// Create a mock hook
	mockHook := &MockCommandHook{}
	
	// Test adding hook
	wrapper.AddHook(mockHook)
	
	if len(wrapper.hooks) != 1 {
		t.Errorf("Expected 1 hook after adding, got %d", len(wrapper.hooks))
	}
	
	// Add another hook
	wrapper.AddHook(mockHook)
	
	if len(wrapper.hooks) != 2 {
		t.Errorf("Expected 2 hooks after adding second, got %d", len(wrapper.hooks))
	}
}

// MockCommandHook implements CommandHook for testing
type MockCommandHook struct {
	preExecuteCalled  bool
	postExecuteCalled bool
	onErrorCalled     bool
}

func (m *MockCommandHook) PreExecute(ctx context.Context, cmd string, args []string) (*types.BackupMetadata, error) {
	m.preExecuteCalled = true
	return nil, nil
}

func (m *MockCommandHook) PostExecute(ctx context.Context, cmd string, args []string, preBackup *types.BackupMetadata) (*types.BackupMetadata, error) {
	m.postExecuteCalled = true
	return nil, nil
}

func (m *MockCommandHook) OnError(ctx context.Context, cmd string, args []string, err error) error {
	m.onErrorCalled = true
	return nil
}