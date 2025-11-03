package restore

import (
	"context"
	"tf-safe/pkg/types"
)

// RestoreEngine defines the interface for restore operations
type RestoreEngine interface {
	// RestoreBackup restores a backup to the specified location
	RestoreBackup(ctx context.Context, opts types.RestoreOptions) error
	
	// ValidateBackup validates a backup before restoration
	ValidateBackup(ctx context.Context, backupID string) error
	
	// CreatePreRestoreBackup creates a backup before performing restoration
	CreatePreRestoreBackup(ctx context.Context, targetPath string) (*types.BackupMetadata, error)
	
	// RollbackRestore rolls back a failed restore operation
	RollbackRestore(ctx context.Context, backupID string) error
}

// BackupValidator defines the interface for backup validation
type BackupValidator interface {
	// ValidateIntegrity validates the integrity of backup data
	ValidateIntegrity(ctx context.Context, data []byte, metadata *types.BackupMetadata) error
	
	// ValidateChecksum validates the checksum of backup data
	ValidateChecksum(ctx context.Context, data []byte, expectedChecksum string) error
	
	// ValidateMetadata validates backup metadata
	ValidateMetadata(ctx context.Context, metadata *types.BackupMetadata) error
}