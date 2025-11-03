package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

const (
	// BackupFileExtension is the extension used for backup files
	BackupFileExtension = ".bak"
	// MetadataFileExtension is the extension used for metadata files
	MetadataFileExtension = ".meta"
	// IndexFileName is the name of the backup index file
	IndexFileName = "index.json"
)

// LocalStorage implements StorageBackend for local filesystem storage
type LocalStorage struct {
	config types.LocalConfig
	logger *utils.Logger
}

// NewLocalStorage creates a new local storage backend
func NewLocalStorage(config types.LocalConfig, logger *utils.Logger) *LocalStorage {
	return &LocalStorage{
		config: config,
		logger: logger,
	}
}

// Initialize sets up the local storage backend
func (ls *LocalStorage) Initialize(ctx context.Context) error {
	// Create the backup directory if it doesn't exist
	if err := utils.EnsureDir(ls.config.Path); err != nil {
		return fmt.Errorf("failed to create backup directory %s: %w", ls.config.Path, err)
	}

	// Set proper permissions on the backup directory (user only)
	if err := os.Chmod(ls.config.Path, 0700); err != nil {
		return fmt.Errorf("failed to set permissions on backup directory: %w", err)
	}

	ls.logger.Info("Local storage initialized at %s", ls.config.Path)
	return nil
}

// Store saves backup data to the local filesystem
func (ls *LocalStorage) Store(ctx context.Context, key string, data []byte, metadata *types.BackupMetadata) error {
	// Generate file paths
	backupPath := filepath.Join(ls.config.Path, key+BackupFileExtension)
	metadataPath := filepath.Join(ls.config.Path, key+MetadataFileExtension)

	// Calculate checksum if not provided
	if metadata.Checksum == "" {
		metadata.Checksum = utils.CalculateChecksumBytes(data)
	}

	// Update metadata
	metadata.FilePath = backupPath
	metadata.Size = int64(len(data))
	metadata.StorageType = ls.GetType()

	// Write backup data atomically
	if err := utils.AtomicWrite(backupPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write backup file %s: %w", backupPath, err)
	}

	// Write metadata atomically
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		// Clean up backup file on metadata error
		_ = os.Remove(backupPath)
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := utils.AtomicWrite(metadataPath, metadataBytes, 0600); err != nil {
		// Clean up backup file on metadata error
		_ = os.Remove(backupPath)
		return fmt.Errorf("failed to write metadata file %s: %w", metadataPath, err)
	}

	// Update backup index
	if err := ls.updateIndex(ctx, metadata); err != nil {
		ls.logger.Warn("Failed to update backup index: %v", err)
		// Don't fail the operation if index update fails
	}

	ls.logger.Info("Backup stored successfully: %s (size: %d bytes, checksum: %s)",
		key, metadata.Size, metadata.Checksum[:8])

	return nil
}

// Retrieve gets backup data from the local filesystem
func (ls *LocalStorage) Retrieve(ctx context.Context, key string) ([]byte, *types.BackupMetadata, error) {
	backupPath := filepath.Join(ls.config.Path, key+BackupFileExtension)
	metadataPath := filepath.Join(ls.config.Path, key+MetadataFileExtension)

	// Check if backup file exists
	if !utils.FileExists(backupPath) {
		return nil, nil, fmt.Errorf("backup file not found: %s", key)
	}

	// Read metadata
	metadata, err := ls.readMetadata(metadataPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read metadata for %s: %w", key, err)
	}

	// Read backup data
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read backup file %s: %w", backupPath, err)
	}

	// Validate checksum
	actualChecksum := utils.CalculateChecksumBytes(data)
	if actualChecksum != metadata.Checksum {
		return nil, nil, fmt.Errorf("checksum mismatch for backup %s: expected %s, got %s",
			key, metadata.Checksum, actualChecksum)
	}

	ls.logger.Debug("Backup retrieved successfully: %s", key)
	return data, metadata, nil
}

// List returns all available backups in the local storage
func (ls *LocalStorage) List(ctx context.Context) ([]*types.BackupMetadata, error) {
	var backups []*types.BackupMetadata

	// Read directory contents
	entries, err := os.ReadDir(ls.config.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return backups, nil // Return empty list if directory doesn't exist
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Process each metadata file
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), MetadataFileExtension) {
			continue
		}

		metadataPath := filepath.Join(ls.config.Path, entry.Name())
		metadata, err := ls.readMetadata(metadataPath)
		if err != nil {
			ls.logger.Warn("Failed to read metadata file %s: %v", entry.Name(), err)
			continue
		}

		backups = append(backups, metadata)
	}

	// Sort by timestamp (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// Delete removes a backup from the local filesystem
func (ls *LocalStorage) Delete(ctx context.Context, key string) error {
	backupPath := filepath.Join(ls.config.Path, key+BackupFileExtension)
	metadataPath := filepath.Join(ls.config.Path, key+MetadataFileExtension)

	// Remove backup file
	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove backup file %s: %w", backupPath, err)
	}

	// Remove metadata file
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove metadata file %s: %w", metadataPath, err)
	}

	// Update backup index
	if err := ls.removeFromIndex(ctx, key); err != nil {
		ls.logger.Warn("Failed to update backup index after deletion: %v", err)
		// Don't fail the operation if index update fails
	}

	ls.logger.Info("Backup deleted successfully: %s", key)
	return nil
}

// Exists checks if a backup exists in the local filesystem
func (ls *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	backupPath := filepath.Join(ls.config.Path, key+BackupFileExtension)
	return utils.FileExists(backupPath), nil
}

// GetType returns the storage backend type identifier
func (ls *LocalStorage) GetType() string {
	return "local"
}

// Cleanup performs any necessary cleanup operations
func (ls *LocalStorage) Cleanup(ctx context.Context) error {
	// For local storage, cleanup is handled by retention policies
	// This method is here for interface compliance
	return nil
}

// readMetadata reads and parses a metadata file
func (ls *LocalStorage) readMetadata(path string) (*types.BackupMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var metadata types.BackupMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// updateIndex updates the backup index with new metadata
func (ls *LocalStorage) updateIndex(ctx context.Context, metadata *types.BackupMetadata) error {
	indexPath := filepath.Join(ls.config.Path, IndexFileName)

	// Read existing index or create new one
	var index types.BackupIndex
	if utils.FileExists(indexPath) {
		data, err := os.ReadFile(indexPath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &index); err != nil {
			return err
		}
	} else {
		index = types.BackupIndex{
			Version: "1.0",
			Backups: make(map[string]*types.BackupMetadata),
		}
	}

	// Update index
	index.Backups[metadata.ID] = metadata
	index.LastSync = time.Now()

	// Write index atomically
	indexData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return utils.AtomicWrite(indexPath, indexData, 0600)
}

// removeFromIndex removes a backup from the index
func (ls *LocalStorage) removeFromIndex(ctx context.Context, key string) error {
	indexPath := filepath.Join(ls.config.Path, IndexFileName)

	if !utils.FileExists(indexPath) {
		return nil // Index doesn't exist, nothing to remove
	}

	// Read existing index
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}

	var index types.BackupIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return err
	}

	// Remove from index
	delete(index.Backups, key)
	index.LastSync = time.Now()

	// Write updated index
	indexData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return utils.AtomicWrite(indexPath, indexData, 0600)
}
