package restore

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"tf-safe/internal/backup"
	"tf-safe/internal/storage"
	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

// Engine implements the RestoreEngine interface
type Engine struct {
	localStorage storage.StorageBackend
	backupEngine backup.BackupEngine
	config       *types.Config
	logger       *utils.Logger
}

// NewEngine creates a new restore engine
func NewEngine(localStorage storage.StorageBackend, backupEngine backup.BackupEngine, config *types.Config, logger *utils.Logger) *Engine {
	return &Engine{
		localStorage: localStorage,
		backupEngine: backupEngine,
		config:       config,
		logger:       logger,
	}
}

// RestoreBackup restores a backup to the specified location
func (e *Engine) RestoreBackup(ctx context.Context, opts types.RestoreOptions) error {
	e.logger.Info("Starting restore operation for backup: %s", opts.BackupID)

	// Validate backup exists and is intact
	if err := e.ValidateBackup(ctx, opts.BackupID); err != nil {
		return fmt.Errorf("backup validation failed: %w", err)
	}

	// Create pre-restore backup if requested
	var preRestoreBackup *types.BackupMetadata
	if opts.CreateBackup && utils.FileExists(opts.TargetPath) {
		var err error
		preRestoreBackup, err = e.CreatePreRestoreBackup(ctx, opts.TargetPath)
		if err != nil {
			return fmt.Errorf("failed to create pre-restore backup: %w", err)
		}
		e.logger.Info("Created pre-restore backup: %s", preRestoreBackup.ID)
	}

	// Retrieve backup data
	data, metadata, err := e.localStorage.Retrieve(ctx, opts.BackupID)
	if err != nil {
		return fmt.Errorf("failed to retrieve backup data: %w", err)
	}

	// Ensure target directory exists
	targetDir := filepath.Dir(opts.TargetPath)
	if err := utils.EnsureDir(targetDir); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Perform atomic restore
	if err := utils.AtomicWrite(opts.TargetPath, data, 0644); err != nil {
		// Attempt rollback if we have a pre-restore backup
		if preRestoreBackup != nil {
			e.logger.Error("Restore failed, attempting rollback to pre-restore backup")
			if rollbackErr := e.RollbackRestore(ctx, preRestoreBackup.ID); rollbackErr != nil {
				e.logger.Error("Rollback failed: %v", rollbackErr)
				return fmt.Errorf("restore failed and rollback failed: restore error: %w, rollback error: %v", err, rollbackErr)
			}
			e.logger.Info("Successfully rolled back to pre-restore state")
		}
		return fmt.Errorf("failed to write restored state file: %w", err)
	}

	e.logger.Info("Successfully restored backup %s to %s (size: %d bytes)", 
		opts.BackupID, opts.TargetPath, metadata.Size)

	return nil
}

// ValidateBackup validates a backup before restoration
func (e *Engine) ValidateBackup(ctx context.Context, backupID string) error {
	// Check if backup exists
	exists, err := e.localStorage.Exists(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to check backup existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("backup not found: %s", backupID)
	}

	// Use backup engine's validation
	if err := e.backupEngine.ValidateBackup(ctx, backupID); err != nil {
		return fmt.Errorf("backup integrity validation failed: %w", err)
	}

	e.logger.Debug("Backup validation successful: %s", backupID)
	return nil
}

// CreatePreRestoreBackup creates a backup before performing restoration
func (e *Engine) CreatePreRestoreBackup(ctx context.Context, targetPath string) (*types.BackupMetadata, error) {
	if !utils.FileExists(targetPath) {
		return nil, fmt.Errorf("target file does not exist: %s", targetPath)
	}

	// Create backup options for pre-restore backup
	opts := types.BackupOptions{
		StateFilePath: targetPath,
		Description:   fmt.Sprintf("Pre-restore backup created at %s", time.Now().Format(time.RFC3339)),
		Force:         false,
	}

	// Create the backup
	metadata, err := e.backupEngine.CreateBackup(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create pre-restore backup: %w", err)
	}

	return metadata, nil
}

// RollbackRestore rolls back a failed restore operation
func (e *Engine) RollbackRestore(ctx context.Context, backupID string) error {
	e.logger.Info("Rolling back restore operation using backup: %s", backupID)

	// Validate the rollback backup
	if err := e.ValidateBackup(ctx, backupID); err != nil {
		return fmt.Errorf("rollback backup validation failed: %w", err)
	}

	// Get the backup metadata to determine original path
	metadata, err := e.backupEngine.GetBackupMetadata(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to get rollback backup metadata: %w", err)
	}

	// Determine target path for rollback
	targetPath := metadata.FilePath
	if targetPath == "" {
		// Fallback to default state file name
		targetPath = "terraform.tfstate"
	}

	// Retrieve backup data
	data, _, err := e.localStorage.Retrieve(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to retrieve rollback backup data: %w", err)
	}

	// Restore the rollback backup
	if err := utils.AtomicWrite(targetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write rollback state file: %w", err)
	}

	e.logger.Info("Successfully rolled back to backup: %s", backupID)
	return nil
}

// Validator implements the BackupValidator interface
type Validator struct {
	logger *utils.Logger
}

// NewValidator creates a new backup validator
func NewValidator(logger *utils.Logger) *Validator {
	return &Validator{
		logger: logger,
	}
}

// ValidateIntegrity validates the integrity of backup data
func (v *Validator) ValidateIntegrity(ctx context.Context, data []byte, metadata *types.BackupMetadata) error {
	// Validate size
	if int64(len(data)) != metadata.Size {
		return fmt.Errorf("size mismatch: expected %d bytes, got %d bytes", metadata.Size, len(data))
	}

	// Validate checksum
	if err := v.ValidateChecksum(ctx, data, metadata.Checksum); err != nil {
		return fmt.Errorf("checksum validation failed: %w", err)
	}

	// Validate metadata
	if err := v.ValidateMetadata(ctx, metadata); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	v.logger.Debug("Backup integrity validation successful")
	return nil
}

// ValidateChecksum validates the checksum of backup data
func (v *Validator) ValidateChecksum(ctx context.Context, data []byte, expectedChecksum string) error {
	actualChecksum := utils.CalculateChecksumBytes(data)
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	v.logger.Debug("Checksum validation successful")
	return nil
}

// ValidateMetadata validates backup metadata
func (v *Validator) ValidateMetadata(ctx context.Context, metadata *types.BackupMetadata) error {
	if metadata.ID == "" {
		return fmt.Errorf("backup ID is empty")
	}

	if metadata.Checksum == "" {
		return fmt.Errorf("backup checksum is empty")
	}

	if metadata.Size < 0 {
		return fmt.Errorf("backup size is negative: %d", metadata.Size)
	}

	if metadata.Timestamp.IsZero() {
		return fmt.Errorf("backup timestamp is zero")
	}

	v.logger.Debug("Metadata validation successful")
	return nil
}