package types

import (
	"fmt"
	"strings"
)

// Config represents the complete tf-safe configuration
type Config struct {
	Local      LocalConfig      `yaml:"local" validate:"required"`
	Remote     RemoteConfig     `yaml:"remote"`
	Encryption EncryptionConfig `yaml:"encryption"`
	Retention  RetentionConfig  `yaml:"retention" validate:"required"`
	Logging    LoggingConfig    `yaml:"logging"`
	Commands   CommandsConfig   `yaml:"commands"`
}

// LocalConfig configures local storage settings
type LocalConfig struct {
	Enabled        bool   `yaml:"enabled"`
	Path           string `yaml:"path" validate:"required"`
	RetentionCount int    `yaml:"retention_count" validate:"min=1"`
}

// RemoteConfig configures remote storage settings
type RemoteConfig struct {
	Provider string `yaml:"provider" validate:"oneof=s3 gcs azure"`
	Bucket   string `yaml:"bucket"`
	Region   string `yaml:"region"`
	Prefix   string `yaml:"prefix"`
	Enabled  bool   `yaml:"enabled"`
}

// EncryptionConfig configures encryption settings
type EncryptionConfig struct {
	Provider   string `yaml:"provider" validate:"oneof=aes kms passphrase none"`
	KMSKeyID   string `yaml:"kms_key_id"`
	Passphrase string `yaml:"passphrase,omitempty"`
}

// RetentionConfig configures backup retention policies
type RetentionConfig struct {
	LocalCount  int `yaml:"local_count" validate:"min=3"`
	RemoteCount int `yaml:"remote_count" validate:"min=1"`
	MaxAgeDays  int `yaml:"max_age_days" validate:"min=1"`
}

// LoggingConfig configures logging settings
type LoggingConfig struct {
	Level  string `yaml:"level" validate:"oneof=debug info warn error"`
	Format string `yaml:"format" validate:"oneof=json text"`
}

// CommandsConfig configures command-specific settings
type CommandsConfig struct {
	Apply   CommandConfig `yaml:"apply"`
	Plan    CommandConfig `yaml:"plan"`
	Destroy CommandConfig `yaml:"destroy"`
}

// CommandConfig configures settings for individual commands
type CommandConfig struct {
	AutoBackup bool `yaml:"auto_backup"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	var errors []string

	// Validate local config
	if c.Local.Enabled {
		if c.Local.Path == "" {
			errors = append(errors, "local.path is required when local storage is enabled")
		}
		if c.Local.RetentionCount < 3 {
			errors = append(errors, "local.retention_count must be at least 3")
		}
	}

	// Validate remote config
	if c.Remote.Enabled {
		if c.Remote.Provider == "" {
			errors = append(errors, "remote.provider is required when remote storage is enabled")
		}
		if c.Remote.Bucket == "" {
			errors = append(errors, "remote.bucket is required when remote storage is enabled")
		}
		if c.Remote.Provider == "s3" && c.Remote.Region == "" {
			errors = append(errors, "remote.region is required for S3 provider")
		}
	}

	// Validate encryption config
	if c.Encryption.Provider == "kms" && c.Encryption.KMSKeyID == "" {
		errors = append(errors, "encryption.kms_key_id is required when using KMS encryption")
	}
	if c.Encryption.Provider == "passphrase" && c.Encryption.Passphrase == "" {
		errors = append(errors, "encryption.passphrase is required when using passphrase encryption")
	}

	// Validate retention config
	if c.Retention.LocalCount < 3 {
		errors = append(errors, "retention.local_count must be at least 3")
	}
	if c.Retention.RemoteCount < 1 {
		errors = append(errors, "retention.remote_count must be at least 1")
	}
	if c.Retention.MaxAgeDays < 1 {
		errors = append(errors, "retention.max_age_days must be at least 1")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}