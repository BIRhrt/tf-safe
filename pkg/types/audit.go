package types

import (
	"time"
)

// AuditAction represents the type of action being audited
type AuditAction string

const (
	AuditActionBackupCreate AuditAction = "backup_create"
	AuditActionBackupDelete AuditAction = "backup_delete"
	AuditActionRestore      AuditAction = "restore"
	AuditActionPrune        AuditAction = "prune"
	AuditActionDriftCheck   AuditAction = "drift_check"
	AuditActionConfigUpdate AuditAction = "config_update"
)

// AuditStatus represents the status of the action
type AuditStatus string

const (
	AuditStatusSuccess AuditStatus = "success"
	AuditStatusFailed  AuditStatus = "failed"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Action    AuditAction            `json:"action"`
	Status    AuditStatus            `json:"status"`
	User      string                 `json:"user,omitempty"`     // System user who performed the action
	Command   string                 `json:"command,omitempty"`  // Full command line executed
	Details   map[string]interface{} `json:"details,omitempty"`  // Action-specific details
	Error     string                 `json:"error,omitempty"`    // Error message if failed
	Duration  string                 `json:"duration,omitempty"` // Duration of the operation
}

// AuditConfig holds configuration for audit logging
type AuditConfig struct {
	Enabled   bool   `mapstructure:"enabled" yaml:"enabled"`
	File      string `mapstructure:"file" yaml:"file"`           // Path to audit log file
	Format    string `mapstructure:"format" yaml:"format"`       // json or text
	Retention int    `mapstructure:"retention" yaml:"retention"` // Days to keep audit logs
}

// DefaultAuditConfig returns the default audit configuration
func DefaultAuditConfig() AuditConfig {
	return AuditConfig{
		Enabled:   true,
		File:      "tf-safe-audit.log",
		Format:    "json",
		Retention: 90,
	}
}
