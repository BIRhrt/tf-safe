package terraform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateDetector_FindStateFiles(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "tf-safe-detector-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	detector := NewStateDetector()

	// Test with no state files
	stateFiles, err := detector.FindStateFiles(tempDir)
	if err != nil {
		t.Fatalf("Failed to find state files in empty dir: %v", err)
	}
	if len(stateFiles) != 0 {
		t.Errorf("Expected 0 state files in empty dir, got %d", len(stateFiles))
	}

	// Create terraform.tfstate file
	stateContent := `{
		"version": 4,
		"terraform_version": "1.0.0",
		"serial": 1,
		"lineage": "test-lineage",
		"outputs": {},
		"resources": []
	}`
	stateFile := filepath.Join(tempDir, "terraform.tfstate")
	if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
		t.Fatalf("Failed to create state file: %v", err)
	}

	// Test with single state file
	stateFiles, err = detector.FindStateFiles(tempDir)
	if err != nil {
		t.Fatalf("Failed to find state files: %v", err)
	}
	if len(stateFiles) != 1 {
		t.Errorf("Expected 1 state file, got %d", len(stateFiles))
	}
	if stateFiles[0] != stateFile {
		t.Errorf("Expected state file %s, got %s", stateFile, stateFiles[0])
	}

	// Create terraform.tfstate.backup file (should not be included)
	backupStateFile := filepath.Join(tempDir, "terraform.tfstate.backup")
	if err := os.WriteFile(backupStateFile, []byte(stateContent), 0644); err != nil {
		t.Fatalf("Failed to create backup state file: %v", err)
	}

	// Test with backup file present (should still only find main state file)
	stateFiles, err = detector.FindStateFiles(tempDir)
	if err != nil {
		t.Fatalf("Failed to find state files with backup present: %v", err)
	}
	if len(stateFiles) != 1 {
		t.Errorf("Expected 1 state file (backup should be ignored), got %d", len(stateFiles))
	}

	// Verify only the main state file is found
	if stateFiles[0] != stateFile {
		t.Errorf("Expected main state file %s, got %s", stateFile, stateFiles[0])
	}
}

func TestStateDetector_FindStateFiles_NonExistentDirectory(t *testing.T) {
	detector := NewStateDetector()

	// The detector doesn't return an error for non-existent directories,
	// it just returns an empty list since os.Stat fails silently
	stateFiles, err := detector.FindStateFiles("/non/existent/directory")
	if err != nil {
		t.Errorf("Unexpected error for non-existent directory: %v", err)
	}
	if len(stateFiles) != 0 {
		t.Errorf("Expected 0 state files for non-existent directory, got %d", len(stateFiles))
	}
}

func TestStateDetector_IsValidStateFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf-safe-detector-validate-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	detector := NewStateDetector()

	// Test with valid state file
	validStateContent := `{
		"version": 4,
		"terraform_version": "1.0.0",
		"serial": 1,
		"lineage": "test-lineage-123",
		"outputs": {},
		"resources": []
	}`
	validStateFile := filepath.Join(tempDir, "valid.tfstate")
	if err := os.WriteFile(validStateFile, []byte(validStateContent), 0644); err != nil {
		t.Fatalf("Failed to create valid state file: %v", err)
	}

	isValid, err := detector.IsValidStateFile(validStateFile)
	if err != nil {
		t.Fatalf("Failed to validate valid state file: %v", err)
	}
	if !isValid {
		t.Error("Valid state file should be considered valid")
	}

	// Test with invalid JSON
	invalidJSONFile := filepath.Join(tempDir, "invalid-json.tfstate")
	if err := os.WriteFile(invalidJSONFile, []byte("not json"), 0644); err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	isValid, err = detector.IsValidStateFile(invalidJSONFile)
	if err != nil {
		t.Fatalf("Failed to validate invalid JSON file: %v", err)
	}
	if isValid {
		t.Error("Invalid JSON file should not be considered valid")
	}

	// Test with missing required fields
	missingFieldsContent := `{
		"version": 4,
		"outputs": {}
	}`
	missingFieldsFile := filepath.Join(tempDir, "missing-fields.tfstate")
	if err := os.WriteFile(missingFieldsFile, []byte(missingFieldsContent), 0644); err != nil {
		t.Fatalf("Failed to create missing fields file: %v", err)
	}

	isValid, err = detector.IsValidStateFile(missingFieldsFile)
	if err != nil {
		t.Fatalf("Failed to validate missing fields file: %v", err)
	}
	if isValid {
		t.Error("File with missing required fields should not be considered valid")
	}

	// Test with non-existent file
	isValid, err = detector.IsValidStateFile("nonexistent.tfstate")
	if err == nil {
		t.Error("Expected error for non-existent file but got none")
	}
	if isValid {
		t.Error("Non-existent file should not be considered valid")
	}
}

func TestStateDetector_IsValidStateFile_EmptyFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf-safe-detector-empty-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	detector := NewStateDetector()

	// Test with empty file
	emptyFile := filepath.Join(tempDir, "empty.tfstate")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	isValid, err := detector.IsValidStateFile(emptyFile)
	if err != nil {
		t.Fatalf("Failed to validate empty file: %v", err)
	}
	if isValid {
		t.Error("Empty file should not be considered valid")
	}
}

func TestStateDetector_IsValidStateFile_OldVersion(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf-safe-detector-old-version-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	detector := NewStateDetector()

	// Test with old version (version 3)
	oldVersionContent := `{
		"version": 3,
		"terraform_version": "0.11.0",
		"serial": 1,
		"lineage": "test-lineage",
		"modules": []
	}`
	oldVersionFile := filepath.Join(tempDir, "old-version.tfstate")
	if err := os.WriteFile(oldVersionFile, []byte(oldVersionContent), 0644); err != nil {
		t.Fatalf("Failed to create old version file: %v", err)
	}

	isValid, err := detector.IsValidStateFile(oldVersionFile)
	if err != nil {
		t.Fatalf("Failed to validate old version file: %v", err)
	}
	if !isValid {
		t.Error("Old version state file should still be considered valid")
	}
}

func TestStateDetector_IsValidStateFile_WithResources(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tf-safe-detector-resources-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	detector := NewStateDetector()

	// Test with resources
	stateWithResourcesContent := `{
		"version": 4,
		"terraform_version": "1.0.0",
		"serial": 5,
		"lineage": "test-lineage-with-resources",
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
	stateWithResourcesFile := filepath.Join(tempDir, "with-resources.tfstate")
	if err := os.WriteFile(stateWithResourcesFile, []byte(stateWithResourcesContent), 0644); err != nil {
		t.Fatalf("Failed to create state file with resources: %v", err)
	}

	isValid, err := detector.IsValidStateFile(stateWithResourcesFile)
	if err != nil {
		t.Fatalf("Failed to validate state file with resources: %v", err)
	}
	if !isValid {
		t.Error("State file with resources should be considered valid")
	}
}