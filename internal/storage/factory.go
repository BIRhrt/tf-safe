package storage

import (
	"fmt"

	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

// DefaultStorageFactory implements StorageFactory interface
type DefaultStorageFactory struct {
	logger *utils.Logger
}

// NewStorageFactory creates a new storage factory
func NewStorageFactory(logger *utils.Logger) StorageFactory {
	return &DefaultStorageFactory{
		logger: logger,
	}
}

// CreateLocal creates a local storage backend
func (f *DefaultStorageFactory) CreateLocal(config types.LocalConfig) (StorageBackend, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("local storage is disabled")
	}

	return NewLocalStorage(config, f.logger), nil
}

// CreateS3 creates an S3 storage backend
func (f *DefaultStorageFactory) CreateS3(config types.RemoteConfig) (StorageBackend, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("remote storage is disabled")
	}

	if config.Provider != "s3" {
		return nil, fmt.Errorf("unsupported remote storage provider: %s", config.Provider)
	}

	if config.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket name is required")
	}

	if config.Region == "" {
		return nil, fmt.Errorf("S3 region is required")
	}

	return NewS3Storage(config, f.logger), nil
}