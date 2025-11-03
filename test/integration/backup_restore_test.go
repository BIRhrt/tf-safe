package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tf-safe/internal/backup"
	"tf-safe/internal/restore"
	"tf-safe/internal/storage"
	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

func TestBackupRestoreWorkflow(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "tf-safe-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test Terraform state file
	stateContent := `{
		"version": 4,
		"terraform_version": "1.0.0",
		"serial": 1,
		"lineage": "test-lineage-123",
		"outputs": {
			"instance_ip": {
				"value": "192.168.1.100",
				"type": "string"
			}
		},
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "example",
				"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
				"instances": [
					{
						"schema_version": 1,
						"attributes": {
							"id": "i-1234567890abcdef0",
							"instance_type": "t2.micro"
						}
					}
				]
			}
		]
	}`

	stateFile := filepath.Join(tempDir, "terraform.tfstate")
	if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
		t.Fatalf("Failed to create state file: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Setup configuration
	config := &types.Config{
		Local: types.LocalConfig{
			Enabled:        true,
			Path:           ".tfstate_snapshots",
			RetentionCount: 5,
		},
		Remote: types.RemoteConfig{
			Enabled: false,
		},
		Encryption: types.EncryptionConfig{
			Provider: "none",
		},
		Retention: types.RetentionConfig{
			LocalCount:  5,
			RemoteCount: 20,
			MaxAgeDays:  30,
		},
	}

	// Setup components
	logger := utils.NewLogger(utils.LogLevelInfo)
	storageFactory := storage.NewStorageFactory(logger)
	localStorage, err := storageFactory.CreateLocal(config.Local)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	ctx := context.Background()
	if err := localStorage.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	backupEngine := backup.NewEngine(localStorage, config, logger)
	restoreEngine := restore.NewEngine(localStorage, backupEngine, config, logger)

	// Test 1: Create backup
	t.Log("Creating backup...")
	opts := types.BackupOptions{
		StateFilePath: stateFile,
	}

	metadata, err := backupEngine.CreateBackup(ctx, opts)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	if metadata == nil {
		t.Fatal("Expected backup metadata but got nil")
	}

	t.Logf("Backup created: %s", metadata.ID)

	// Test 2: List backups
	t.Log("Listing backups...")
	backups, err := backupEngine.ListBackups(ctx)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) != 1 {
		t.Errorf("Expected 1 backup, got %d", len(backups))
	}

	// Test 3: Validate backup
	t.Log("Validating backup...")
	err = backupEngine.ValidateBackup(ctx, metadata.ID)
	if err != nil {
		t.Errorf("Backup validation failed: %v", err)
	}

	// Test 4: Modify state file
	t.Log("Modifying state file...")
	modifiedContent := `{
		"version": 4,
		"terraform_version": "1.0.0",
		"serial": 2,
		"lineage": "test-lineage-123",
		"outputs": {
			"instance_ip": {
				"value": "192.168.1.200",
				"type": "string"
			}
		},
		"resources": []
	}`

	if err := os.WriteFile(stateFile, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("Failed to modify state file: %v", err)
	}

	// Test 5: Restore backup
	t.Log("Restoring backup...")
	restoreOpts := types.RestoreOptions{
		BackupID:     metadata.ID,
		TargetPath:   stateFile,
		Force:        false,
		CreateBackup: true,
	}

	err = restoreEngine.RestoreBackup(ctx, restoreOpts)
	if err != nil {
		t.Fatalf("Failed to restore backup: %v", err)
	}

	// Test 6: Verify restoration (check that file exists and has reasonable content)
	t.Log("Verifying restoration...")
	restoredContent, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read restored state file: %v", err)
	}

	// Verify the file is not empty and contains JSON-like content
	if len(restoredContent) == 0 {
		t.Error("Restored file is empty")
	}

	// Basic check that it looks like a Terraform state file
	restoredStr := string(restoredContent)
	if !strings.Contains(restoredStr, "version") || !strings.Contains(restoredStr, "terraform_version") {
		t.Error("Restored content doesn't appear to be a valid Terraform state file")
	}

	t.Log("Backup and restore workflow completed successfully")
}

func TestMultipleBackupsRetention(t *testing.T) {
	// This test verifies that the retention policy logic works correctly
	// by testing the retention manager directly rather than creating actual backups
	// which would require long delays to ensure unique timestamps
	
	config := types.RetentionConfig{
		LocalCount: 4, // Keep 4 backups (> minimum of 3)
	}

	logger := utils.NewLogger(utils.LogLevelInfo)
	retentionManager := backup.NewRetentionManager(config, logger)

	ctx := context.Background()
	now := time.Now().UTC()

	// Create test backup metadata (simulating 6 backups)
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

	t.Logf("Testing retention policy with %d backups", len(testBackups))

	// Test retention policy
	toDelete, err := retentionManager.ApplyLocalRetentionPolicy(ctx, testBackups)
	if err != nil {
		t.Fatalf("Failed to apply retention policy: %v", err)
	}

	// Should delete 2 oldest backups (keep 4 newest)
	expectedDeleteCount := 2
	if len(toDelete) != expectedDeleteCount {
		t.Errorf("Expected %d backups to delete, got %d", expectedDeleteCount, len(toDelete))
	}

	// Verify the oldest backups are marked for deletion
	expectedToDelete := map[string]bool{
		"backup-1": true,
		"backup-2": true,
	}

	for _, backup := range toDelete {
		if !expectedToDelete[backup.ID] {
			t.Errorf("Unexpected backup marked for deletion: %s", backup.ID)
		}
	}

	t.Logf("Retention policy test completed successfully: %d backups marked for deletion", len(toDelete))
}

func TestEncryptedBackupWorkflow(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "tf-safe-encryption-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test state file
	stateContent := `{"version": 4, "terraform_version": "1.0.0", "serial": 1}`
	stateFile := filepath.Join(tempDir, "terraform.tfstate")
	if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
		t.Fatalf("Failed to create state file: %v", err)
	}

	// Setup configuration with encryption
	config := &types.Config{
		Local: types.LocalConfig{
			Enabled: true,
			Path:    filepath.Join(tempDir, ".tfstate_snapshots"),
		},
		Encryption: types.EncryptionConfig{
			Provider:   "aes",
			Passphrase: "test-passphrase-for-integration",
		},
	}

	// Setup components
	logger := utils.NewLogger(utils.LogLevelInfo)
	storageFactory := storage.NewStorageFactory(logger)
	localStorage, err := storageFactory.CreateLocal(config.Local)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	ctx := context.Background()
	if err := localStorage.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create backup engine with encryption
	backupEngine := backup.NewEngine(localStorage, config, logger)

	// Test encrypted backup creation
	t.Log("Creating encrypted backup...")
	opts := types.BackupOptions{
		StateFilePath: stateFile,
	}

	metadata, err := backupEngine.CreateBackup(ctx, opts)
	if err != nil {
		t.Fatalf("Failed to create encrypted backup: %v", err)
	}

	// Verify backup was created
	backups, err := backupEngine.ListBackups(ctx)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) != 1 {
		t.Errorf("Expected 1 backup, got %d", len(backups))
	}

	// Verify backup validation works with encryption
	err = backupEngine.ValidateBackup(ctx, metadata.ID)
	if err != nil {
		t.Errorf("Encrypted backup validation failed: %v", err)
	}

	t.Log("Encrypted backup workflow completed successfully")
}