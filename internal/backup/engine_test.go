package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

// MockStorageBackend implements StorageBackend for testing
type MockStorageBackend struct {
	backups map[string][]byte
	metadata map[string]*types.BackupMetadata
	storageType string
	shouldFail bool
}

func NewMockStorageBackend(storageType string) *MockStorageBackend {
	return &MockStorageBackend{
		backups: make(map[string][]byte),
		metadata: make(map[string]*types.BackupMetadata),
		storageType: storageType,
	}
}

func (m *MockStorageBackend) Store(ctx context.Context, key string, data []byte, metadata *types.BackupMetadata) error {
	if m.shouldFail {
		return &types.TfSafeError{Code: "STORAGE_ERROR", Message: "Mock storage failure"}
	}
	m.backups[key] = data
	m.metadata[key] = metadata
	return nil
}

func (m *MockStorageBackend) Retrieve(ctx context.Context, key string) ([]byte, *types.BackupMetadata, error) {
	if m.shouldFail {
		return nil, nil, &types.TfSafeError{Code: "STORAGE_ERROR", Message: "Mock storage failure"}
	}
	data, exists := m.backups[key]
	if !exists {
		return nil, nil, &types.TfSafeError{Code: "BACKUP_NOT_FOUND", Message: "Backup not found"}
	}
	return data, m.metadata[key], nil
}

func (m *MockStorageBackend) List(ctx context.Context) ([]*types.BackupMetadata, error) {
	if m.shouldFail {
		return nil, &types.TfSafeError{Code: "STORAGE_ERROR", Message: "Mock storage failure"}
	}
	var result []*types.BackupMetadata
	for _, metadata := range m.metadata {
		result = append(result, metadata)
	}
	return result, nil
}

func (m *MockStorageBackend) Delete(ctx context.Context, key string) error {
	if m.shouldFail {
		return &types.TfSafeError{Code: "STORAGE_ERROR", Message: "Mock storage failure"}
	}
	delete(m.backups, key)
	delete(m.metadata, key)
	return nil
}

func (m *MockStorageBackend) Exists(ctx context.Context, key string) (bool, error) {
	if m.shouldFail {
		return false, &types.TfSafeError{Code: "STORAGE_ERROR", Message: "Mock storage failure"}
	}
	_, exists := m.backups[key]
	return exists, nil
}

func (m *MockStorageBackend) GetType() string {
	return m.storageType
}

func (m *MockStorageBackend) Initialize(ctx context.Context) error {
	return nil
}

func (m *MockStorageBackend) Cleanup(ctx context.Context) error {
	return nil
}

func (m *MockStorageBackend) SetShouldFail(fail bool) {
	m.shouldFail = fail
}

func TestEngine_CreateBackup(t *testing.T) {
	// Create temporary directory and state file
	tempDir, err := os.MkdirTemp("", "tf-safe-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	stateContent := `{"version": 4, "terraform_version": "1.0.0", "serial": 1}`
	stateFile := filepath.Join(tempDir, "terraform.tfstate")
	if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
		t.Fatalf("Failed to create state file: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create mock storage and engine
	mockStorage := NewMockStorageBackend("local")
	config := &types.Config{
		Local: types.LocalConfig{
			Enabled: true,
			Path:    ".tfstate_snapshots",
		},
		Remote: types.RemoteConfig{
			Enabled: false,
		},
	}
	logger := utils.NewLogger(utils.LogLevelInfo)
	engine := NewEngine(mockStorage, config, logger)

	ctx := context.Background()
	opts := types.BackupOptions{
		StateFilePath: stateFile,
	}

	// Test backup creation
	metadata, err := engine.CreateBackup(ctx, opts)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	if metadata == nil {
		t.Fatal("Expected backup metadata but got nil")
	}

	if metadata.Size != int64(len(stateContent)) {
		t.Errorf("Expected backup size %d, got %d", len(stateContent), metadata.Size)
	}

	if metadata.Checksum == "" {
		t.Error("Expected backup checksum but got empty string")
	}

	// Verify backup was stored
	exists, err := mockStorage.Exists(ctx, metadata.ID)
	if err != nil {
		t.Fatalf("Failed to check backup existence: %v", err)
	}
	if !exists {
		t.Error("Backup was not stored in mock storage")
	}
}

func TestEngine_CreateBackup_MissingStateFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf-safe-backup-missing-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	mockStorage := NewMockStorageBackend("local")
	config := &types.Config{}
	logger := utils.NewLogger(utils.LogLevelInfo)
	engine := NewEngine(mockStorage, config, logger)

	ctx := context.Background()
	opts := types.BackupOptions{
		StateFilePath: "nonexistent.tfstate",
		Force:         false,
	}

	// Test backup creation with missing file
	_, err = engine.CreateBackup(ctx, opts)
	if err == nil {
		t.Error("Expected error for missing state file but got none")
	}

	// Test with force flag
	opts.Force = true
	metadata, err := engine.CreateBackup(ctx, opts)
	if err != nil {
		t.Fatalf("Failed to create backup with force flag: %v", err)
	}

	if metadata.Size != 0 {
		t.Errorf("Expected backup size 0 for missing file, got %d", metadata.Size)
	}
}

func TestEngine_ListBackups(t *testing.T) {
	mockStorage := NewMockStorageBackend("local")
	config := &types.Config{}
	logger := utils.NewLogger(utils.LogLevelInfo)
	engine := NewEngine(mockStorage, config, logger)

	ctx := context.Background()

	// Add some test backups to mock storage
	now := time.Now().UTC()
	testBackups := []*types.BackupMetadata{
		{
			ID:        "terraform.tfstate.2023-01-01T10:00:00Z",
			Timestamp: now.Add(-2 * time.Hour),
			Size:      100,
			Checksum:  "checksum1",
		},
		{
			ID:        "terraform.tfstate.2023-01-01T12:00:00Z",
			Timestamp: now,
			Size:      200,
			Checksum:  "checksum2",
		},
	}

	for _, backup := range testBackups {
		mockStorage.metadata[backup.ID] = backup
	}

	// Test listing backups
	backups, err := engine.ListBackups(ctx)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) != 2 {
		t.Errorf("Expected 2 backups, got %d", len(backups))
	}

	// Verify backups are sorted by timestamp (newest first)
	if len(backups) >= 2 && backups[0].Timestamp.Before(backups[1].Timestamp) {
		t.Error("Backups should be sorted by timestamp (newest first)")
	}
}

func TestEngine_ValidateBackup(t *testing.T) {
	mockStorage := NewMockStorageBackend("local")
	config := &types.Config{}
	logger := utils.NewLogger(utils.LogLevelInfo)
	engine := NewEngine(mockStorage, config, logger)

	ctx := context.Background()
	testData := []byte("test backup data")
	backupID := "test-backup-id"

	// Store valid backup
	metadata := &types.BackupMetadata{
		ID:       backupID,
		Size:     int64(len(testData)),
		Checksum: utils.CalculateChecksumBytes(testData),
	}
	_ = mockStorage.Store(ctx, backupID, testData, metadata)

	// Test validation of valid backup
	err := engine.ValidateBackup(ctx, backupID)
	if err != nil {
		t.Errorf("Validation failed for valid backup: %v", err)
	}

	// Test validation of corrupted backup (wrong checksum)
	corruptedMetadata := &types.BackupMetadata{
		ID:       "corrupted-backup",
		Size:     int64(len(testData)),
		Checksum: "wrong-checksum",
	}
	_ = mockStorage.Store(ctx, "corrupted-backup", testData, corruptedMetadata)

	err = engine.ValidateBackup(ctx, "corrupted-backup")
	if err == nil {
		t.Error("Expected validation error for corrupted backup but got none")
	}

	// Test validation of non-existent backup
	err = engine.ValidateBackup(ctx, "non-existent")
	if err == nil {
		t.Error("Expected validation error for non-existent backup but got none")
	}
}

func TestEngine_CleanupOldBackups(t *testing.T) {
	mockStorage := NewMockStorageBackend("local")
	config := &types.Config{
		Retention: types.RetentionConfig{
			LocalCount: 4, // Keep 4 backups (must be > MinimumRetentionCount of 3)
		},
	}
	logger := utils.NewLogger(utils.LogLevelInfo)
	engine := NewEngine(mockStorage, config, logger)

	ctx := context.Background()
	now := time.Now().UTC()

	// Add 6 backups (should delete 2, keep 4)
	testBackups := []*types.BackupMetadata{
		{
			ID:        "backup-1",
			Timestamp: now.Add(-6 * time.Hour),
			Size:      100,
		},
		{
			ID:        "backup-2",
			Timestamp: now.Add(-5 * time.Hour),
			Size:      100,
		},
		{
			ID:        "backup-3",
			Timestamp: now.Add(-4 * time.Hour),
			Size:      100,
		},
		{
			ID:        "backup-4",
			Timestamp: now.Add(-3 * time.Hour),
			Size:      100,
		},
		{
			ID:        "backup-5",
			Timestamp: now.Add(-2 * time.Hour),
			Size:      100,
		},
		{
			ID:        "backup-6",
			Timestamp: now.Add(-1 * time.Hour),
			Size:      100,
		},
	}

	for _, backup := range testBackups {
		mockStorage.metadata[backup.ID] = backup
		mockStorage.backups[backup.ID] = []byte("test data")
	}

	// Test cleanup
	err := engine.CleanupOldBackups(ctx)
	if err != nil {
		t.Fatalf("Failed to cleanup old backups: %v", err)
	}

	// Verify 4 backups remain (the newest ones)
	remainingBackups, err := mockStorage.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list remaining backups: %v", err)
	}

	if len(remainingBackups) != 4 {
		t.Errorf("Expected 4 backups after cleanup, got %d", len(remainingBackups))
	}
}

func TestEngine_WithRemoteStorage(t *testing.T) {
	localStorage := NewMockStorageBackend("local")
	remoteStorage := NewMockStorageBackend("s3")
	config := &types.Config{
		Remote: types.RemoteConfig{
			Enabled: true,
		},
	}
	logger := utils.NewLogger(utils.LogLevelInfo)
	engine := NewEngineWithRemote(localStorage, remoteStorage, config, logger)

	// Create temporary state file
	tempDir, err := os.MkdirTemp("", "tf-safe-remote-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	stateContent := `{"version": 4}`
	stateFile := filepath.Join(tempDir, "terraform.tfstate")
	if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
		t.Fatalf("Failed to create state file: %v", err)
	}

	ctx := context.Background()
	opts := types.BackupOptions{
		StateFilePath: stateFile,
	}

	// Test backup creation with remote storage
	metadata, err := engine.CreateBackup(ctx, opts)
	if err != nil {
		t.Fatalf("Failed to create backup with remote storage: %v", err)
	}

	// Verify backup exists in both local and remote storage
	localExists, err := localStorage.Exists(ctx, metadata.ID)
	if err != nil {
		t.Fatalf("Failed to check local storage: %v", err)
	}
	if !localExists {
		t.Error("Backup not found in local storage")
	}

	remoteExists, err := remoteStorage.Exists(ctx, metadata.ID)
	if err != nil {
		t.Fatalf("Failed to check remote storage: %v", err)
	}
	if !remoteExists {
		t.Error("Backup not found in remote storage")
	}
}