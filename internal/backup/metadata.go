package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

// MetadataManager handles backup metadata operations
type MetadataManager struct {
	backupDir string
	logger    *utils.Logger
}

// NewMetadataManager creates a new metadata manager
func NewMetadataManager(backupDir string, logger *utils.Logger) *MetadataManager {
	return &MetadataManager{
		backupDir: backupDir,
		logger:    logger,
	}
}

// LoadIndex loads the backup index from disk
func (mm *MetadataManager) LoadIndex() (*types.BackupIndex, error) {
	indexPath := filepath.Join(mm.backupDir, "index.json")
	
	if !utils.FileExists(indexPath) {
		// Return empty index if file doesn't exist
		return &types.BackupIndex{
			Version: "1.0",
			Backups: make(map[string]*types.BackupMetadata),
		}, nil
	}

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup index: %w", err)
	}

	var index types.BackupIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse backup index: %w", err)
	}

	return &index, nil
}

// SaveIndex saves the backup index to disk
func (mm *MetadataManager) SaveIndex(index *types.BackupIndex) error {
	indexPath := filepath.Join(mm.backupDir, "index.json")
	
	// Update last sync time
	index.LastSync = time.Now()
	
	// Marshal index to JSON
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup index: %w", err)
	}

	// Write atomically
	if err := utils.AtomicWrite(indexPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write backup index: %w", err)
	}

	mm.logger.Debug("Backup index saved with %d entries", len(index.Backups))
	return nil
}

// AddBackup adds a backup to the index
func (mm *MetadataManager) AddBackup(metadata *types.BackupMetadata) error {
	index, err := mm.LoadIndex()
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	index.Backups[metadata.ID] = metadata
	
	if err := mm.SaveIndex(index); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	mm.logger.Debug("Added backup to index: %s", metadata.ID)
	return nil
}

// RemoveBackup removes a backup from the index
func (mm *MetadataManager) RemoveBackup(backupID string) error {
	index, err := mm.LoadIndex()
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	delete(index.Backups, backupID)
	
	if err := mm.SaveIndex(index); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	mm.logger.Debug("Removed backup from index: %s", backupID)
	return nil
}

// GetBackup retrieves backup metadata by ID
func (mm *MetadataManager) GetBackup(backupID string) (*types.BackupMetadata, error) {
	index, err := mm.LoadIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}

	backup, exists := index.Backups[backupID]
	if !exists {
		return nil, fmt.Errorf("backup not found: %s", backupID)
	}

	return backup, nil
}

// ListBackups returns all backups sorted by timestamp (newest first)
func (mm *MetadataManager) ListBackups() ([]*types.BackupMetadata, error) {
	index, err := mm.LoadIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}

	var backups []*types.BackupMetadata
	for _, backup := range index.Backups {
		backups = append(backups, backup)
	}

	// Sort by timestamp (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// ValidateIndex validates the backup index against actual files
func (mm *MetadataManager) ValidateIndex() error {
	index, err := mm.LoadIndex()
	if err != nil {
		return fmt.Errorf("failed to load index: %w", err)
	}

	var orphanedEntries []string
	var missingFiles []string

	// Check if indexed backups exist on disk
	for backupID := range index.Backups {
		backupPath := filepath.Join(mm.backupDir, backupID+".bak")
		if !utils.FileExists(backupPath) {
			orphanedEntries = append(orphanedEntries, backupID)
			mm.logger.Warn("Backup file missing for indexed entry: %s", backupID)
		}
	}

	// Check for backup files not in index
	entries, err := os.ReadDir(mm.backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".bak" {
			continue
		}

		backupID := entry.Name()[:len(entry.Name())-4] // Remove .bak extension
		if _, exists := index.Backups[backupID]; !exists {
			missingFiles = append(missingFiles, backupID)
			mm.logger.Warn("Backup file not indexed: %s", backupID)
		}
	}

	// Clean up orphaned entries
	if len(orphanedEntries) > 0 {
		mm.logger.Info("Cleaning up %d orphaned index entries", len(orphanedEntries))
		for _, backupID := range orphanedEntries {
			delete(index.Backups, backupID)
		}
		if err := mm.SaveIndex(index); err != nil {
			return fmt.Errorf("failed to save cleaned index: %w", err)
		}
	}

	if len(missingFiles) > 0 {
		mm.logger.Warn("Found %d backup files not in index - consider rebuilding index", len(missingFiles))
	}

	mm.logger.Debug("Index validation complete: %d orphaned entries cleaned, %d missing files found", 
		len(orphanedEntries), len(missingFiles))
	
	return nil
}

// RebuildIndex rebuilds the backup index from existing backup files
func (mm *MetadataManager) RebuildIndex() error {
	mm.logger.Info("Rebuilding backup index from existing files")
	
	entries, err := os.ReadDir(mm.backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	index := &types.BackupIndex{
		Version: "1.0",
		Backups: make(map[string]*types.BackupMetadata),
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".bak" {
			continue
		}

		backupID := entry.Name()[:len(entry.Name())-4] // Remove .bak extension
		backupPath := filepath.Join(mm.backupDir, entry.Name())
		metadataPath := filepath.Join(mm.backupDir, backupID+".meta")

		// Try to load existing metadata
		var metadata *types.BackupMetadata
		if utils.FileExists(metadataPath) {
			metadataData, err := os.ReadFile(metadataPath)
			if err == nil {
				if err := json.Unmarshal(metadataData, &metadata); err == nil {
					index.Backups[backupID] = metadata
					continue
				}
			}
		}

		// Generate metadata from file if metadata file is missing or corrupted
		fileInfo, err := entry.Info()
		if err != nil {
			mm.logger.Warn("Failed to get file info for %s: %v", entry.Name(), err)
			continue
		}

		// Calculate checksum
		checksum, err := utils.CalculateChecksum(backupPath)
		if err != nil {
			mm.logger.Warn("Failed to calculate checksum for %s: %v", entry.Name(), err)
			continue
		}

		metadata = &types.BackupMetadata{
			ID:          backupID,
			Timestamp:   fileInfo.ModTime(),
			Size:        fileInfo.Size(),
			Checksum:    checksum,
			StorageType: "local",
			Encrypted:   false,
			FilePath:    backupPath,
		}

		index.Backups[backupID] = metadata
		mm.logger.Debug("Rebuilt metadata for backup: %s", backupID)
	}

	if err := mm.SaveIndex(index); err != nil {
		return fmt.Errorf("failed to save rebuilt index: %w", err)
	}

	mm.logger.Info("Index rebuild complete: %d backups indexed", len(index.Backups))
	return nil
}