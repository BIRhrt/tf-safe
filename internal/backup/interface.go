package backup

import (
	"context"
	"tf-safe/pkg/types"
)

// BackupEngine defines the interface for backup operations
type BackupEngine interface {
	// CreateBackup creates a new backup with the given options
	CreateBackup(ctx context.Context, opts types.BackupOptions) (*types.BackupMetadata, error)
	
	// ListBackups returns all available backups
	ListBackups(ctx context.Context) ([]*types.BackupMetadata, error)
	
	// CleanupOldBackups removes old backups according to retention policies
	CleanupOldBackups(ctx context.Context) error
	
	// GetBackupMetadata returns metadata for a specific backup
	GetBackupMetadata(ctx context.Context, backupID string) (*types.BackupMetadata, error)
	
	// ValidateBackup validates the integrity of a backup
	ValidateBackup(ctx context.Context, backupID string) error
}

// RetentionManager defines the interface for backup retention management
type RetentionManager interface {
	// ApplyRetentionPolicy applies retention policies to remove old backups (legacy method)
	ApplyRetentionPolicy(ctx context.Context, backups []*types.BackupMetadata) ([]*types.BackupMetadata, error)
	
	// ApplyLocalRetentionPolicy applies retention policies to local backups
	ApplyLocalRetentionPolicy(ctx context.Context, backups []*types.BackupMetadata) ([]*types.BackupMetadata, error)
	
	// ApplyRemoteRetentionPolicy applies retention policies to remote backups
	ApplyRemoteRetentionPolicy(ctx context.Context, backups []*types.BackupMetadata) ([]*types.BackupMetadata, error)
	
	// ShouldRetain determines if a backup should be retained
	ShouldRetain(backup *types.BackupMetadata, totalCount int) bool
	
	// GetRetentionConfig returns the current retention configuration
	GetRetentionConfig() types.RetentionConfig
}