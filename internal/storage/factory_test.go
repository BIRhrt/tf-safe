package storage

import (
	"testing"

	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

func TestFactory_CreateLocal(t *testing.T) {
	logger := utils.NewLogger(utils.LogLevelInfo)
	factory := NewStorageFactory(logger)

	config := types.LocalConfig{
		Enabled:        true,
		Path:           "/tmp/test-backups",
		RetentionCount: 5,
	}

	storage, err := factory.CreateLocal(config)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	if storage == nil {
		t.Fatal("Expected storage instance but got nil")
	}

	if storage.GetType() != "local" {
		t.Errorf("Expected storage type 'local', got '%s'", storage.GetType())
	}
}

func TestFactory_CreateLocal_Disabled(t *testing.T) {
	logger := utils.NewLogger(utils.LogLevelInfo)
	factory := NewStorageFactory(logger)

	config := types.LocalConfig{
		Enabled: false, // Disabled should cause error
		Path:    "/tmp/test-backups",
	}

	_, err := factory.CreateLocal(config)
	if err == nil {
		t.Error("Expected error for disabled local storage but got none")
	}
}

func TestFactory_CreateS3(t *testing.T) {
	logger := utils.NewLogger(utils.LogLevelInfo)
	factory := NewStorageFactory(logger)

	config := types.RemoteConfig{
		Enabled:  true,
		Provider: "s3",
		Bucket:   "test-bucket",
		Region:   "us-west-2",
		Prefix:   "backups/",
	}

	storage, err := factory.CreateS3(config)
	if err != nil {
		t.Fatalf("Failed to create S3 storage: %v", err)
	}

	if storage == nil {
		t.Fatal("Expected storage instance but got nil")
	}

	if storage.GetType() != "s3" {
		t.Errorf("Expected storage type 's3', got '%s'", storage.GetType())
	}
}

func TestFactory_CreateS3_MissingBucket(t *testing.T) {
	logger := utils.NewLogger(utils.LogLevelInfo)
	factory := NewStorageFactory(logger)

	config := types.RemoteConfig{
		Enabled:  true,
		Provider: "s3",
		Bucket:   "", // Missing bucket should cause error
		Region:   "us-west-2",
	}

	_, err := factory.CreateS3(config)
	if err == nil {
		t.Error("Expected error for missing bucket but got none")
	}
}

func TestFactory_CreateS3_MissingRegion(t *testing.T) {
	logger := utils.NewLogger(utils.LogLevelInfo)
	factory := NewStorageFactory(logger)

	config := types.RemoteConfig{
		Enabled:  true,
		Provider: "s3",
		Bucket:   "test-bucket",
		Region:   "", // Missing region should cause error
	}

	_, err := factory.CreateS3(config)
	if err == nil {
		t.Error("Expected error for missing region but got none")
	}
}

func TestFactory_CreateS3_UnsupportedProvider(t *testing.T) {
	logger := utils.NewLogger(utils.LogLevelInfo)
	factory := NewStorageFactory(logger)

	config := types.RemoteConfig{
		Enabled:  true,
		Provider: "gcs", // Unsupported provider
		Bucket:   "test-bucket",
		Region:   "us-west-2",
	}

	_, err := factory.CreateS3(config)
	if err == nil {
		t.Error("Expected error for unsupported provider but got none")
	}
}