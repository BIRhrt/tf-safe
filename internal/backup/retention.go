package backup

import (
	"context"
	"sort"
	"time"

	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

const (
	// MinimumRetentionCount is the minimum number of backups to retain
	MinimumRetentionCount = 3
)

// RetentionManagerImpl implements the RetentionManager interface
type RetentionManagerImpl struct {
	config types.RetentionConfig
	logger *utils.Logger
}

// NewRetentionManager creates a new retention manager
func NewRetentionManager(config types.RetentionConfig, logger *utils.Logger) RetentionManager {
	return &RetentionManagerImpl{
		config: config,
		logger: logger,
	}
}

// ApplyLocalRetentionPolicy applies retention policies to local backups
func (rm *RetentionManagerImpl) ApplyLocalRetentionPolicy(ctx context.Context, backups []*types.BackupMetadata) ([]*types.BackupMetadata, error) {
	return rm.applyRetentionPolicy(ctx, backups, rm.config.LocalCount, "local")
}

// ApplyRemoteRetentionPolicy applies retention policies to remote backups
func (rm *RetentionManagerImpl) ApplyRemoteRetentionPolicy(ctx context.Context, backups []*types.BackupMetadata) ([]*types.BackupMetadata, error) {
	return rm.applyRetentionPolicy(ctx, backups, rm.config.RemoteCount, "remote")
}

// ApplyRetentionPolicy applies retention policies to remove old backups (legacy method)
func (rm *RetentionManagerImpl) ApplyRetentionPolicy(ctx context.Context, backups []*types.BackupMetadata) ([]*types.BackupMetadata, error) {
	return rm.ApplyLocalRetentionPolicy(ctx, backups)
}

// applyRetentionPolicy applies retention policies to remove old backups
func (rm *RetentionManagerImpl) applyRetentionPolicy(ctx context.Context, backups []*types.BackupMetadata, retentionCount int, storageType string) ([]*types.BackupMetadata, error) {
	rm.logger.Info("Starting %s retention policy analysis for %d backups", storageType, len(backups))
	
	if len(backups) <= MinimumRetentionCount {
		rm.logger.Info("Backup count (%d) is at or below minimum retention count (%d), no cleanup needed", 
			len(backups), MinimumRetentionCount)
		return nil, nil
	}

	// Sort backups by timestamp (newest first)
	sortedBackups := make([]*types.BackupMetadata, len(backups))
	copy(sortedBackups, backups)
	sort.Slice(sortedBackups, func(i, j int) bool {
		return sortedBackups[i].Timestamp.After(sortedBackups[j].Timestamp)
	})

	var toDelete []*types.BackupMetadata
	now := time.Now()

	rm.logger.Debug("%s retention configuration: Count=%d, MaxAgeDays=%d, MinimumCount=%d", 
		storageType, retentionCount, rm.config.MaxAgeDays, MinimumRetentionCount)

	// Apply count-based retention
	if retentionCount > MinimumRetentionCount && len(sortedBackups) > retentionCount {
		rm.logger.Debug("Applying count-based retention: keeping %d newest backups", retentionCount)
		// Keep the newest retentionCount backups, mark the rest for deletion
		for i := retentionCount; i < len(sortedBackups); i++ {
			backup := sortedBackups[i]
			if len(sortedBackups)-len(toDelete) > MinimumRetentionCount {
				toDelete = append(toDelete, backup)
				rm.logger.Debug("Marking backup for deletion (count policy): %s (timestamp: %s)", 
					backup.ID, backup.Timestamp.Format(time.RFC3339))
			}
		}
	}

	// Apply age-based retention
	if rm.config.MaxAgeDays > 0 {
		maxAge := time.Duration(rm.config.MaxAgeDays) * 24 * time.Hour
		rm.logger.Debug("Applying age-based retention: max age %v", maxAge)
		
		for _, backup := range sortedBackups {
			if rm.shouldDeleteByAge(backup, now) {
				// Only delete if we're not already marking it for deletion and we maintain minimum count
				alreadyMarked := false
				for _, marked := range toDelete {
					if marked.ID == backup.ID {
						alreadyMarked = true
						break
					}
				}
				
				if !alreadyMarked && len(sortedBackups)-len(toDelete) > MinimumRetentionCount {
					toDelete = append(toDelete, backup)
					age := now.Sub(backup.Timestamp)
					rm.logger.Debug("Marking backup for deletion (age policy): %s (age: %v, max: %v)", 
						backup.ID, age, maxAge)
				}
			}
		}
	}

	// Ensure we never delete more than we should to maintain minimum count
	if len(sortedBackups)-len(toDelete) < MinimumRetentionCount {
		// Sort toDelete by timestamp (oldest first) and remove some from deletion list
		sort.Slice(toDelete, func(i, j int) bool {
			return toDelete[i].Timestamp.Before(toDelete[j].Timestamp)
		})
		
		// Keep enough to maintain minimum count
		keepCount := MinimumRetentionCount - (len(sortedBackups) - len(toDelete))
		if keepCount > 0 && keepCount < len(toDelete) {
			rm.logger.Info("Adjusting deletion list to maintain minimum retention count of %d", MinimumRetentionCount)
			toDelete = toDelete[keepCount:]
		} else if keepCount >= len(toDelete) {
			rm.logger.Info("Cannot delete any backups without violating minimum retention count of %d", MinimumRetentionCount)
			toDelete = nil
		}
	}

	rm.logger.Info("Retention policy analysis complete: %d total backups, %d marked for deletion, %d will remain", 
		len(backups), len(toDelete), len(backups)-len(toDelete))
	
	// Log details of backups to be deleted
	if len(toDelete) > 0 {
		rm.logger.Info("Backups scheduled for deletion:")
		for _, backup := range toDelete {
			age := now.Sub(backup.Timestamp)
			rm.logger.Info("  - %s (age: %v, size: %d bytes)", backup.ID, age, backup.Size)
		}
	}
	
	return toDelete, nil
}

// ShouldRetain determines if a backup should be retained
func (rm *RetentionManagerImpl) ShouldRetain(backup *types.BackupMetadata, totalCount int) bool {
	// Always retain if we're at or below minimum count
	if totalCount <= MinimumRetentionCount {
		return true
	}

	// Check age-based retention
	if rm.shouldDeleteByAge(backup, time.Now()) {
		return false
	}

	return true
}

// GetRetentionConfig returns the current retention configuration
func (rm *RetentionManagerImpl) GetRetentionConfig() types.RetentionConfig {
	return rm.config
}

// shouldDeleteByAge determines if a backup should be deleted based on age
func (rm *RetentionManagerImpl) shouldDeleteByAge(backup *types.BackupMetadata, now time.Time) bool {
	if rm.config.MaxAgeDays <= 0 {
		return false // Age-based retention disabled
	}

	maxAge := time.Duration(rm.config.MaxAgeDays) * 24 * time.Hour
	age := now.Sub(backup.Timestamp)
	
	return age > maxAge
}