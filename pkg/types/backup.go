package types

import "time"

// BackupMetadata contains information about a backup
type BackupMetadata struct {
	ID          string    `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum"`
	StorageType string    `json:"storage_type"`
	Encrypted   bool      `json:"encrypted"`
	FilePath    string    `json:"file_path"`
}

// BackupOptions contains options for creating backups
type BackupOptions struct {
	StateFilePath string
	Description   string
	Force         bool
}

// BackupIndex maintains an index of all backups
type BackupIndex struct {
	Version  string                     `json:"version"`
	Backups  map[string]*BackupMetadata `json:"backups"`
	LastSync time.Time                  `json:"last_sync"`
}

// RestoreOptions contains options for restoring backups
type RestoreOptions struct {
	BackupID      string
	TargetPath    string
	CreateBackup  bool
	Force         bool
}