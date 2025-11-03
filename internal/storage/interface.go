package storage

import (
	"context"
	"tf-safe/pkg/types"
)

// StorageBackend defines the interface for all storage implementations
type StorageBackend interface {
	// Store saves backup data to the storage backend
	Store(ctx context.Context, key string, data []byte, metadata *types.BackupMetadata) error

	// Retrieve gets backup data from the storage backend
	Retrieve(ctx context.Context, key string) ([]byte, *types.BackupMetadata, error)

	// List returns all available backups in the storage backend
	List(ctx context.Context) ([]*types.BackupMetadata, error)

	// Delete removes a backup from the storage backend
	Delete(ctx context.Context, key string) error

	// Exists checks if a backup exists in the storage backend
	Exists(ctx context.Context, key string) (bool, error)

	// GetType returns the storage backend type identifier
	GetType() string

	// Initialize sets up the storage backend
	Initialize(ctx context.Context) error

	// Cleanup performs any necessary cleanup operations
	Cleanup(ctx context.Context) error
}

// StorageFactory creates storage backend instances
type StorageFactory interface {
	CreateLocal(config types.LocalConfig) (StorageBackend, error)
	CreateS3(config types.RemoteConfig) (StorageBackend, error)
}
