package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

func TestLocalStorage_Store(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "tf-safe-local-storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := types.LocalConfig{
		Enabled: true,
		Path:    tempDir,
	}

	logger := utils.NewLogger(utils.LogLevelInfo)
	storage := NewLocalStorage(config, logger)

	ctx := context.Background()
	if err := storage.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Test data
	testData := []byte("test backup data")
	backupID := "terraform.tfstate.2023-01-01T10:00:00Z"
	metadata := &types.BackupMetadata{
		ID:        backupID,
		Timestamp: time.Now().UTC(),
		Size:      int64(len(testData)),
		Checksum:  "test-checksum",
	}

	// Store backup
	_ = storage.Store(ctx, backupID, testData, metadata)

	// Verify backup file exists
	backupPath := filepath.Join(tempDir, backupID+".bak")
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}

	// Verify metadata file exists
	metadataPath := filepath.Join(tempDir, backupID+".meta")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Error("Metadata file was not created")
	}

	// Verify stored data
	storedData, err := os.ReadFile(backupPath)

	if string(storedData) != string(testData) {
		t.Errorf("Stored data doesn't match original. Got: %s, Want: %s", string(storedData), string(testData))
	}
}

func TestLocalStorage_Retrieve(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf-safe-local-retrieve-test")
	defer os.RemoveAll(tempDir)

	config := types.LocalConfig{
		Enabled: true,
		Path:    tempDir,
	}

	logger := utils.NewLogger(utils.LogLevelInfo)
	storage := NewLocalStorage(config, logger)

	ctx := context.Background()
	if err := storage.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Store a backup first
	testData := []byte("test backup data for retrieval")
	backupID := "terraform.tfstate.2023-01-01T11:00:00Z"
	originalMetadata := &types.BackupMetadata{
		ID:        backupID,
		Timestamp: time.Now().UTC(),
		Size:      int64(len(testData)),
		Checksum:  utils.CalculateChecksumBytes(testData),
	}

	_ = storage.Store(ctx, backupID, testData, originalMetadata)

	// Retrieve the backup
	retrievedData, retrievedMetadata, err := storage.Retrieve(ctx, backupID)
	if err != nil {
		t.Fatalf("Failed to retrieve backup: %v", err)
	}

	// Verify retrieved data
	if string(retrievedData) != string(testData) {
		t.Errorf("Retrieved data doesn't match original. Got: %s, Want: %s", string(retrievedData), string(testData))
	}

	// Verify retrieved metadata
	if retrievedMetadata.ID != originalMetadata.ID {
		t.Errorf("Retrieved metadata ID doesn't match. Got: %s, Want: %s", retrievedMetadata.ID, originalMetadata.ID)
	}
	if retrievedMetadata.Checksum != originalMetadata.Checksum {
		t.Errorf("Retrieved metadata checksum doesn't match. Got: %s, Want: %s", retrievedMetadata.Checksum, originalMetadata.Checksum)
	}
}

func TestLocalStorage_List(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf-safe-local-list-test")
	defer os.RemoveAll(tempDir)

	config := types.LocalConfig{
		Enabled: true,
		Path:    tempDir,
	}

	logger := utils.NewLogger(utils.LogLevelInfo)
	storage := NewLocalStorage(config, logger)

	ctx := context.Background()
	if err := storage.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Store multiple backups
	testBackups := []struct {
		id   string
		data string
	}{
		{"terraform.tfstate.2023-01-01T10:00:00Z", "backup data 1"},
		{"terraform.tfstate.2023-01-01T11:00:00Z", "backup data 2"},
		{"terraform.tfstate.2023-01-01T12:00:00Z", "backup data 3"},
	}

	for _, backup := range testBackups {
		data := []byte(backup.data)
		metadata := &types.BackupMetadata{
			ID:        backup.id,
			Timestamp: time.Now().UTC(),
			Size:      int64(len(data)),
			Checksum:  "checksum-" + backup.id,
		}
		_ = storage.Store(ctx, backup.id, data, metadata)
	}

	// List backups
	backups, err := storage.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) != len(testBackups) {
		t.Errorf("Expected %d backups, got %d", len(testBackups), len(backups))
	}

	// Verify all backup IDs are present
	foundIDs := make(map[string]bool)
	for _, backup := range backups {
		foundIDs[backup.ID] = true
	}

	for _, expectedBackup := range testBackups {
		if !foundIDs[expectedBackup.id] {
			t.Errorf("Expected backup ID %s not found in list", expectedBackup.id)
		}
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf-safe-local-delete-test")
	defer os.RemoveAll(tempDir)

	config := types.LocalConfig{
		Enabled: true,
		Path:    tempDir,
	}

	logger := utils.NewLogger(utils.LogLevelInfo)
	storage := NewLocalStorage(config, logger)

	ctx := context.Background()
	if err := storage.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Store a backup
	testData := []byte("test backup data for deletion")
	backupID := "terraform.tfstate.2023-01-01T13:00:00Z"
	metadata := &types.BackupMetadata{
		ID:        backupID,
		Timestamp: time.Now().UTC(),
		Size:      int64(len(testData)),
		Checksum:  "test-checksum-delete",
	}

	_ = storage.Store(ctx, backupID, testData, metadata)

	// Verify backup exists
	exists, err := storage.Exists(ctx, backupID)
	if err != nil {
		t.Fatalf("Failed to check backup existence: %v", err)
	}
	if !exists {
		t.Error("Backup should exist before deletion")
	}

	// Delete the backup
	_ = storage.Delete(ctx, backupID)

	// Verify backup no longer exists
	exists, _ = storage.Exists(ctx, backupID)
	if exists {
		t.Error("Backup should not exist after deletion")
	}

	// Verify files are actually deleted
	backupPath := filepath.Join(tempDir, backupID+".bak")
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("Backup file should be deleted")
	}

	metadataPath := filepath.Join(tempDir, backupID+".meta")
	if _, err := os.Stat(metadataPath); !os.IsNotExist(err) {
		t.Error("Metadata file should be deleted")
	}
}

func TestLocalStorage_Exists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf-safe-local-exists-test")
	defer os.RemoveAll(tempDir)

	config := types.LocalConfig{
		Enabled: true,
		Path:    tempDir,
	}

	logger := utils.NewLogger(utils.LogLevelInfo)
	storage := NewLocalStorage(config, logger)

	ctx := context.Background()
	if err := storage.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	backupID := "terraform.tfstate.2023-01-01T14:00:00Z"

	// Check non-existent backup
	exists, err := storage.Exists(ctx, backupID)
	if err != nil {
		t.Fatalf("Failed to check non-existent backup: %v", err)
	}
	if exists {
		t.Error("Non-existent backup should not exist")
	}

	// Store a backup
	testData := []byte("test backup data for exists check")
	metadata := &types.BackupMetadata{
		ID:        backupID,
		Timestamp: time.Now().UTC(),
		Size:      int64(len(testData)),
		Checksum:  "test-checksum-exists",
	}

	_ = storage.Store(ctx, backupID, testData, metadata)

	// Check existing backup
	exists, _ = storage.Exists(ctx, backupID)
	if !exists {
		t.Error("Existing backup should exist")
	}
}

func TestLocalStorage_GetType(t *testing.T) {
	config := types.LocalConfig{
		Enabled: true,
		Path:    "/tmp",
	}

	logger := utils.NewLogger(utils.LogLevelInfo)
	storage := NewLocalStorage(config, logger)

	storageType := storage.GetType()
	if storageType != "local" {
		t.Errorf("Expected storage type 'local', got '%s'", storageType)
	}
}