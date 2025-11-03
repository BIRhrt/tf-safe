package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tf-safe/internal/storage"
	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

const (
	// DefaultStateFileName is the default Terraform state file name
	DefaultStateFileName = "terraform.tfstate"
	// BackupIDTimeFormat is the format used for backup IDs
	BackupIDTimeFormat = "2006-01-02T15:04:05Z"
)

// Engine implements the BackupEngine interface
type Engine struct {
	localStorage  storage.StorageBackend
	remoteStorage storage.StorageBackend
	config        *types.Config
	logger        *utils.Logger
}

// NewEngine creates a new backup engine
func NewEngine(localStorage storage.StorageBackend, config *types.Config, logger *utils.Logger) *Engine {
	return &Engine{
		localStorage: localStorage,
		config:       config,
		logger:       logger,
	}
}

// NewEngineWithRemote creates a new backup engine with remote storage support
func NewEngineWithRemote(localStorage, remoteStorage storage.StorageBackend, config *types.Config, logger *utils.Logger) *Engine {
	return &Engine{
		localStorage:  localStorage,
		remoteStorage: remoteStorage,
		config:        config,
		logger:        logger,
	}
}

// CreateBackup creates a new backup with the given options
func (e *Engine) CreateBackup(ctx context.Context, opts types.BackupOptions) (*types.BackupMetadata, error) {
	// Detect state file if not provided
	stateFilePath := opts.StateFilePath
	if stateFilePath == "" {
		var err error
		stateFilePath, err = e.detectStateFile()
		if err != nil {
			return nil, fmt.Errorf("failed to detect state file: %w", err)
		}
	}

	// Check if state file exists
	if !utils.FileExists(stateFilePath) {
		if !opts.Force {
			return nil, fmt.Errorf("state file not found: %s", stateFilePath)
		}
		e.logger.Warn("State file not found, creating empty backup: %s", stateFilePath)
	}

	// Read state file data
	var stateData []byte
	var err error
	if utils.FileExists(stateFilePath) {
		stateData, err = os.ReadFile(stateFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read state file %s: %w", stateFilePath, err)
		}
	}

	// Generate backup metadata
	now := time.Now().UTC()
	backupID := e.generateBackupID(now)

	metadata := &types.BackupMetadata{
		ID:          backupID,
		Timestamp:   now,
		Size:        int64(len(stateData)),
		Checksum:    utils.CalculateChecksumBytes(stateData),
		StorageType: e.localStorage.GetType(),
		Encrypted:   false, // Will be set by encryption layer if enabled
		FilePath:    stateFilePath,
	}

	// Store backup using local storage backend
	if err := e.localStorage.Store(ctx, backupID, stateData, metadata); err != nil {
		return nil, fmt.Errorf("failed to store backup locally: %w", err)
	}

	// Store backup using remote storage backend if configured
	if e.remoteStorage != nil && e.config.Remote.Enabled {
		// Create a copy of metadata for remote storage
		remoteMetadata := *metadata
		if err := e.remoteStorage.Store(ctx, backupID, stateData, &remoteMetadata); err != nil {
			e.logger.Error("Failed to store backup remotely: %v", err)
			// Don't fail the entire operation if remote storage fails
			// The backup is still available locally
		} else {
			e.logger.Info("Backup stored remotely: %s", backupID)
		}
	}

	e.logger.Info("Backup created successfully: %s from %s", backupID, stateFilePath)
	return metadata, nil
}

// ListBackups returns all available backups from both local and remote storage
func (e *Engine) ListBackups(ctx context.Context) ([]*types.BackupMetadata, error) {
	var allBackups []*types.BackupMetadata
	backupMap := make(map[string]*types.BackupMetadata)

	// Get local backups
	localBackups, err := e.localStorage.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list local backups: %w", err)
	}

	// Add local backups to map
	for _, backup := range localBackups {
		backupMap[backup.ID] = backup
	}

	// Get remote backups if configured
	if e.remoteStorage != nil && e.config.Remote.Enabled {
		remoteBackups, err := e.remoteStorage.List(ctx)
		if err != nil {
			e.logger.Warn("Failed to list remote backups: %v", err)
			// Continue with local backups only
		} else {
			// Add remote backups to map, preferring local versions if they exist
			for _, backup := range remoteBackups {
				if existing, exists := backupMap[backup.ID]; exists {
					// If local version exists, add remote info to it
					existing.FilePath = fmt.Sprintf("%s, %s", existing.FilePath, backup.FilePath)
				} else {
					// Add remote-only backup
					backupMap[backup.ID] = backup
				}
			}
		}
	}

	// Convert map to slice
	for _, backup := range backupMap {
		allBackups = append(allBackups, backup)
	}

	// Sort by timestamp (newest first)
	for i := 0; i < len(allBackups)-1; i++ {
		for j := i + 1; j < len(allBackups); j++ {
			if allBackups[i].Timestamp.Before(allBackups[j].Timestamp) {
				allBackups[i], allBackups[j] = allBackups[j], allBackups[i]
			}
		}
	}

	e.logger.Debug("Listed %d backups (%d local, %d total)", len(localBackups), len(allBackups))
	return allBackups, nil
}

// CleanupOldBackups removes old backups according to retention policies
func (e *Engine) CleanupOldBackups(ctx context.Context) error {
	// Apply retention policy for local backups
	localDeletedCount, err := e.cleanupLocalBackups(ctx)
	if err != nil {
		return fmt.Errorf("failed to cleanup local backups: %w", err)
	}

	// Apply retention policy for remote backups if configured
	remoteDeletedCount := 0
	if e.remoteStorage != nil && e.config.Remote.Enabled {
		count, err := e.cleanupRemoteBackups(ctx)
		if err != nil {
			e.logger.Warn("Failed to cleanup remote backups: %v", err)
			// Don't fail the entire operation if remote cleanup fails
		} else {
			remoteDeletedCount = count
		}
	}

	totalDeleted := localDeletedCount + remoteDeletedCount
	if totalDeleted > 0 {
		e.logger.Info("Cleanup completed: deleted %d local and %d remote backups",
			localDeletedCount, remoteDeletedCount)
	} else {
		e.logger.Debug("No backups needed cleanup")
	}

	return nil
}

// cleanupLocalBackups applies retention policy to local backups
func (e *Engine) cleanupLocalBackups(ctx context.Context) (int, error) {
	// Get local backups
	localBackups, err := e.localStorage.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list local backups: %w", err)
	}

	// Apply local retention policy
	retentionManager := NewRetentionManager(e.config.Retention, e.logger)
	toDelete, err := retentionManager.ApplyLocalRetentionPolicy(ctx, localBackups)
	if err != nil {
		return 0, fmt.Errorf("failed to apply local retention policy: %w", err)
	}

	// Delete old local backups
	deletedCount := 0
	for _, backup := range toDelete {
		if err := e.localStorage.Delete(ctx, backup.ID); err != nil {
			e.logger.Error("Failed to delete local backup %s: %v", backup.ID, err)
			continue
		}
		deletedCount++
		e.logger.Info("Deleted old local backup: %s (timestamp: %s)",
			backup.ID, backup.Timestamp.Format(time.RFC3339))
	}

	return deletedCount, nil
}

// cleanupRemoteBackups applies retention policy to remote backups
func (e *Engine) cleanupRemoteBackups(ctx context.Context) (int, error) {
	// Get remote backups
	remoteBackups, err := e.remoteStorage.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list remote backups: %w", err)
	}

	// Apply remote retention policy
	retentionManager := NewRetentionManager(e.config.Retention, e.logger)
	toDelete, err := retentionManager.ApplyRemoteRetentionPolicy(ctx, remoteBackups)
	if err != nil {
		return 0, fmt.Errorf("failed to apply remote retention policy: %w", err)
	}

	// Delete old remote backups
	deletedCount := 0
	for _, backup := range toDelete {
		if err := e.remoteStorage.Delete(ctx, backup.ID); err != nil {
			e.logger.Error("Failed to delete remote backup %s: %v", backup.ID, err)
			continue
		}
		deletedCount++
		e.logger.Info("Deleted old remote backup: %s (timestamp: %s)",
			backup.ID, backup.Timestamp.Format(time.RFC3339))
	}

	return deletedCount, nil
}

// GetBackupMetadata returns metadata for a specific backup
func (e *Engine) GetBackupMetadata(ctx context.Context, backupID string) (*types.BackupMetadata, error) {
	// Try local storage first
	_, metadata, err := e.localStorage.Retrieve(ctx, backupID)
	if err == nil {
		return metadata, nil
	}

	// If not found locally and remote storage is configured, try remote
	if e.remoteStorage != nil && e.config.Remote.Enabled {
		_, remoteMetadata, remoteErr := e.remoteStorage.Retrieve(ctx, backupID)
		if remoteErr == nil {
			return remoteMetadata, nil
		}
		e.logger.Debug("Backup %s not found in remote storage: %v", backupID, remoteErr)
	}

	return nil, fmt.Errorf("failed to get backup metadata for %s: %w", backupID, err)
}

// ValidateBackup validates the integrity of a backup
func (e *Engine) ValidateBackup(ctx context.Context, backupID string) error {
	// Try to validate local backup first
	localErr := e.validateBackupFromStorage(ctx, backupID, e.localStorage, "local")
	if localErr == nil {
		return nil
	}

	// If local validation failed and remote storage is configured, try remote
	if e.remoteStorage != nil && e.config.Remote.Enabled {
		remoteErr := e.validateBackupFromStorage(ctx, backupID, e.remoteStorage, "remote")
		if remoteErr == nil {
			return nil
		}
		e.logger.Debug("Remote backup validation failed for %s: %v", backupID, remoteErr)
	}

	return fmt.Errorf("backup validation failed for %s: %w", backupID, localErr)
}

// validateBackupFromStorage validates a backup from a specific storage backend
func (e *Engine) validateBackupFromStorage(ctx context.Context, backupID string, storage storage.StorageBackend, storageType string) error {
	data, metadata, err := storage.Retrieve(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to retrieve backup from %s storage: %w", storageType, err)
	}

	// Validate checksum
	actualChecksum := utils.CalculateChecksumBytes(data)
	if actualChecksum != metadata.Checksum {
		return fmt.Errorf("backup %s is corrupted in %s storage: checksum mismatch (expected %s, got %s)",
			backupID, storageType, metadata.Checksum, actualChecksum)
	}

	// Validate size
	if int64(len(data)) != metadata.Size {
		return fmt.Errorf("backup %s is corrupted in %s storage: size mismatch (expected %d, got %d)",
			backupID, storageType, metadata.Size, len(data))
	}

	e.logger.Debug("Backup validation successful in %s storage: %s", storageType, backupID)
	return nil
}

// detectStateFile detects the Terraform state file in the current directory
func (e *Engine) detectStateFile() (string, error) {
	// Check current directory for terraform.tfstate
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	stateFilePath := filepath.Join(currentDir, DefaultStateFileName)

	// Check if the default state file exists
	if utils.FileExists(stateFilePath) {
		e.logger.Debug("Detected state file: %s", stateFilePath)
		return stateFilePath, nil
	}

	// Look for any .tfstate files in the current directory
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return "", fmt.Errorf("failed to read current directory: %w", err)
	}

	var stateFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tfstate") {
			stateFiles = append(stateFiles, entry.Name())
		}
	}

	if len(stateFiles) == 0 {
		return "", fmt.Errorf("no Terraform state files found in current directory")
	}

	if len(stateFiles) == 1 {
		detectedPath := filepath.Join(currentDir, stateFiles[0])
		e.logger.Debug("Detected state file: %s", detectedPath)
		return detectedPath, nil
	}

	// Multiple state files found, prefer terraform.tfstate if it exists
	for _, file := range stateFiles {
		if file == DefaultStateFileName {
			detectedPath := filepath.Join(currentDir, file)
			e.logger.Debug("Detected default state file: %s", detectedPath)
			return detectedPath, nil
		}
	}

	// Return the first one found
	detectedPath := filepath.Join(currentDir, stateFiles[0])
	e.logger.Warn("Multiple state files found, using: %s", detectedPath)
	return detectedPath, nil
}

// generateBackupID generates a unique backup ID based on timestamp
func (e *Engine) generateBackupID(timestamp time.Time) string {
	// Format: terraform.tfstate.YYYY-MM-DDTHH:MM:SSZ
	return fmt.Sprintf("terraform.tfstate.%s", timestamp.Format(BackupIDTimeFormat))
}
