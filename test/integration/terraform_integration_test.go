package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tf-safe/internal/backup"
	"tf-safe/internal/config"
	"tf-safe/internal/restore"
	"tf-safe/internal/storage"
	"tf-safe/internal/terraform"
	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

func TestTerraformStateDetection(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "tf-safe-terraform-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Test 1: No state file
	detector := terraform.NewStateDetector()
	stateFiles, err := detector.FindStateFiles(tempDir)
	if err != nil {
		t.Fatalf("Failed to find state files in empty dir: %v", err)
	}
	if len(stateFiles) != 0 {
		t.Errorf("Expected 0 state files in empty dir, got %d", len(stateFiles))
	}

	// Test 2: Create valid state file
	validStateContent := `{
		"version": 4,
		"terraform_version": "1.0.0",
		"serial": 1,
		"lineage": "test-lineage-integration",
		"outputs": {},
		"resources": []
	}`

	stateFile := filepath.Join(tempDir, "terraform.tfstate")
	if err := os.WriteFile(stateFile, []byte(validStateContent), 0644); err != nil {
		t.Fatalf("Failed to create state file: %v", err)
	}

	// Test state file detection
	stateFiles, err = detector.FindStateFiles(tempDir)
	if err != nil {
		t.Fatalf("Failed to find state files: %v", err)
	}
	if len(stateFiles) != 1 {
		t.Errorf("Expected 1 state file, got %d", len(stateFiles))
	}

	// Test state file validation
	isValid, err := detector.IsValidStateFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to validate state file: %v", err)
	}
	if !isValid {
		t.Error("Valid state file should be considered valid")
	}

	// Test 3: Create invalid state file
	invalidStateFile := filepath.Join(tempDir, "invalid.tfstate")
	if err := os.WriteFile(invalidStateFile, []byte("not json"), 0644); err != nil {
		t.Fatalf("Failed to create invalid state file: %v", err)
	}

	isValid, err = detector.IsValidStateFile(invalidStateFile)
	if err != nil {
		t.Fatalf("Failed to validate invalid state file: %v", err)
	}
	if isValid {
		t.Error("Invalid state file should not be considered valid")
	}

	t.Log("Terraform state detection test completed successfully")
}

func TestTerraformWrapperIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "tf-safe-wrapper-integration")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test state file
	stateContent := `{
		"version": 4,
		"terraform_version": "1.0.0",
		"serial": 1,
		"lineage": "wrapper-test-lineage",
		"outputs": {},
		"resources": []
	}`

	stateFile := filepath.Join(tempDir, "terraform.tfstate")
	if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
		t.Fatalf("Failed to create state file: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Setup configuration
	testConfig := &types.Config{
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
	}

	// Setup components
	logger := utils.NewLogger(utils.LogLevelInfo)
	configManager := config.NewManager()
	
	storageFactory := storage.NewStorageFactory(logger)
	localStorage, err := storageFactory.CreateLocal(testConfig.Local)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	ctx := context.Background()
	if err := localStorage.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	backupEngine := backup.NewEngine(localStorage, testConfig, logger)
	wrapper := terraform.NewWrapper(configManager, backupEngine)

	// Test state file detection
	detectedFile, err := wrapper.DetectStateFile()
	if err != nil {
		t.Fatalf("Failed to detect state file: %v", err)
	}

	expectedPath := filepath.Join(tempDir, "terraform.tfstate")
	// Handle potential symlink resolution on macOS
	expectedResolved, _ := filepath.EvalSymlinks(expectedPath)
	detectedResolved, _ := filepath.EvalSymlinks(detectedFile)
	
	if detectedResolved != expectedResolved {
		t.Errorf("Expected detected file %s, got %s", expectedResolved, detectedResolved)
	}

	// Test state file validation
	err = wrapper.ValidateStateFile(stateFile)
	if err != nil {
		t.Errorf("State file validation failed: %v", err)
	}

	// Test Terraform binary check (may fail in CI environments without Terraform)
	err = wrapper.CheckTerraformBinary()
	if err != nil {
		t.Logf("Terraform binary check failed (expected in environments without Terraform): %v", err)
	} else {
		t.Log("Terraform binary check passed")
		
		// If Terraform is available, test version detection
		version, err := wrapper.GetTerraformVersion()
		if err != nil {
			t.Errorf("Failed to get Terraform version: %v", err)
		} else {
			t.Logf("Detected Terraform version: %s", version)
		}
	}

	t.Log("Terraform wrapper integration test completed successfully")
}

func TestEndToEndBackupWithTerraform(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "tf-safe-e2e-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create initial state file
	initialState := `{
		"version": 4,
		"terraform_version": "1.0.0",
		"serial": 1,
		"lineage": "e2e-test-lineage",
		"outputs": {
			"environment": {
				"value": "development",
				"type": "string"
			}
		},
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"instances": [
					{
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
	if err := os.WriteFile(stateFile, []byte(initialState), 0644); err != nil {
		t.Fatalf("Failed to create initial state file: %v", err)
	}

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Setup complete tf-safe system
	testConfig := &types.Config{
		Local: types.LocalConfig{
			Enabled:        true,
			Path:           ".tfstate_snapshots",
			RetentionCount: 3,
		},
		Remote: types.RemoteConfig{
			Enabled: false,
		},
		Encryption: types.EncryptionConfig{
			Provider: "none",
		},
		Retention: types.RetentionConfig{
			LocalCount: 3,
		},
	}

	logger := utils.NewLogger(utils.LogLevelInfo)
	configManager := config.NewManager()
	
	storageFactory := storage.NewStorageFactory(logger)
	localStorage, err := storageFactory.CreateLocal(testConfig.Local)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	ctx := context.Background()
	if err := localStorage.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	backupEngine := backup.NewEngine(localStorage, testConfig, logger)
	wrapper := terraform.NewWrapper(configManager, backupEngine)

	// Step 1: Detect and validate initial state
	t.Log("Step 1: Detecting initial state...")
	detectedFile, err := wrapper.DetectStateFile()
	if err != nil {
		t.Fatalf("Failed to detect state file: %v", err)
	}

	err = wrapper.ValidateStateFile(detectedFile)
	if err != nil {
		t.Fatalf("Initial state validation failed: %v", err)
	}

	// Step 2: Create initial backup
	t.Log("Step 2: Creating initial backup...")
	opts := types.BackupOptions{
		StateFilePath: stateFile,
	}

	metadata1, err := backupEngine.CreateBackup(ctx, opts)
	if err != nil {
		t.Fatalf("Failed to create initial backup: %v", err)
	}

	// Step 3: Simulate Terraform operation (modify state)
	t.Log("Step 3: Simulating Terraform operation...")
	modifiedState := `{
		"version": 4,
		"terraform_version": "1.0.0",
		"serial": 2,
		"lineage": "e2e-test-lineage",
		"outputs": {
			"environment": {
				"value": "production",
				"type": "string"
			}
		},
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"instances": [
					{
						"attributes": {
							"id": "i-1234567890abcdef0",
							"instance_type": "t3.small"
						}
					}
				]
			},
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "db",
				"instances": [
					{
						"attributes": {
							"id": "i-0987654321fedcba0",
							"instance_type": "t3.medium"
						}
					}
				]
			}
		]
	}`

	if err := os.WriteFile(stateFile, []byte(modifiedState), 0644); err != nil {
		t.Fatalf("Failed to write modified state: %v", err)
	}

	// Step 4: Create post-operation backup (wait to ensure different timestamp)
	t.Log("Step 4: Creating post-operation backup...")
	time.Sleep(1 * time.Second) // Ensure different timestamp
	_, err = backupEngine.CreateBackup(ctx, opts)
	if err != nil {
		t.Fatalf("Failed to create post-operation backup: %v", err)
	}

	// Step 5: List and verify backups
	t.Log("Step 5: Listing and verifying backups...")
	backups, err := backupEngine.ListBackups(ctx)
	if err != nil {
		t.Fatalf("Failed to list backups: %v", err)
	}

	if len(backups) != 2 {
		t.Errorf("Expected 2 backups, got %d", len(backups))
	}

	// Verify both backups
	for _, backup := range backups {
		err = backupEngine.ValidateBackup(ctx, backup.ID)
		if err != nil {
			t.Errorf("Backup validation failed for %s: %v", backup.ID, err)
		}
	}

	// Step 6: Test restoration to previous state
	t.Log("Step 6: Testing restoration to previous state...")
	restoreEngine := restore.NewEngine(localStorage, backupEngine, testConfig, logger)
	
	restoreOpts := types.RestoreOptions{
		BackupID:     metadata1.ID,
		TargetPath:   stateFile,
		Force:        false,
		CreateBackup: true,
	}

	err = restoreEngine.RestoreBackup(ctx, restoreOpts)
	if err != nil {
		t.Fatalf("Failed to restore backup: %v", err)
	}

	// Verify restoration
	restoredContent, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read restored state: %v", err)
	}

	if string(restoredContent) != initialState {
		t.Error("Restored state doesn't match initial state")
	}

	t.Log("End-to-end backup workflow completed successfully")
}