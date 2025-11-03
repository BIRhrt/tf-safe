package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"tf-safe/pkg/types"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// Validator provides comprehensive configuration validation
type Validator struct {
	errors []ValidationError
}

// NewValidator creates a new configuration validator
func NewValidator() *Validator {
	return &Validator{
		errors: make([]ValidationError, 0),
	}
}

// ValidateConfig performs comprehensive validation of the configuration
func (v *Validator) ValidateConfig(config *types.Config) error {
	v.errors = make([]ValidationError, 0)
	
	v.validateLocalConfig(config.Local)
	v.validateRemoteConfig(config.Remote)
	v.validateEncryptionConfig(config.Encryption)
	v.validateRetentionConfig(config.Retention)
	v.validateLoggingConfig(config.Logging)
	
	if len(v.errors) > 0 {
		return v.buildValidationError()
	}
	
	return nil
}

// validateLocalConfig validates local storage configuration
func (v *Validator) validateLocalConfig(config types.LocalConfig) {
	if config.Enabled {
		// Validate path
		if config.Path == "" {
			v.addError("local.path", config.Path, "path is required when local storage is enabled")
		} else {
			// Check if path is valid
			if !isValidPath(config.Path) {
				v.addError("local.path", config.Path, "path contains invalid characters")
			}
			
			// Check if we can create the directory
			absPath, err := filepath.Abs(config.Path)
			if err != nil {
				v.addError("local.path", config.Path, "cannot resolve absolute path")
			} else {
				// Check if parent directory exists and is writable
				parentDir := filepath.Dir(absPath)
				if _, err := os.Stat(parentDir); os.IsNotExist(err) {
					v.addError("local.path", config.Path, "parent directory does not exist")
				} else if err := checkWritePermission(parentDir); err != nil {
					v.addError("local.path", config.Path, "parent directory is not writable")
				}
			}
		}
		
		// Validate retention count
		if config.RetentionCount < MinRetentionCount {
			v.addError("local.retention_count", config.RetentionCount, 
				fmt.Sprintf("must be at least %d", MinRetentionCount))
		}
		if config.RetentionCount > 1000 {
			v.addError("local.retention_count", config.RetentionCount, "must not exceed 1000")
		}
	}
}

// validateRemoteConfig validates remote storage configuration
func (v *Validator) validateRemoteConfig(config types.RemoteConfig) {
	if config.Enabled {
		// Validate provider
		validProviders := []string{"s3", "gcs", "azure"}
		if !contains(validProviders, config.Provider) {
			v.addError("remote.provider", config.Provider, 
				fmt.Sprintf("must be one of: %s", strings.Join(validProviders, ", ")))
		}
		
		// Validate bucket name
		if config.Bucket == "" {
			v.addError("remote.bucket", config.Bucket, "bucket name is required")
		} else if !isValidBucketName(config.Bucket, config.Provider) {
			v.addError("remote.bucket", config.Bucket, "invalid bucket name format")
		}
		
		// Provider-specific validation
		switch config.Provider {
		case "s3":
			if config.Region == "" {
				v.addError("remote.region", config.Region, "region is required for S3 provider")
			} else if !isValidAWSRegion(config.Region) {
				v.addError("remote.region", config.Region, "invalid AWS region format")
			}
		case "gcs":
			// GCS doesn't require region, but validate if provided
			if config.Region != "" && !isValidGCPRegion(config.Region) {
				v.addError("remote.region", config.Region, "invalid GCP region format")
			}
		}
		
		// Validate prefix if provided
		if config.Prefix != "" && !isValidPrefix(config.Prefix) {
			v.addError("remote.prefix", config.Prefix, "invalid prefix format")
		}
	}
}

// validateEncryptionConfig validates encryption configuration
func (v *Validator) validateEncryptionConfig(config types.EncryptionConfig) {
	validProviders := []string{"aes", "kms", "passphrase", "none"}
	if !contains(validProviders, config.Provider) {
		v.addError("encryption.provider", config.Provider, 
			fmt.Sprintf("must be one of: %s", strings.Join(validProviders, ", ")))
	}
	
	switch config.Provider {
	case "kms":
		if config.KMSKeyID == "" {
			v.addError("encryption.kms_key_id", config.KMSKeyID, "KMS key ID is required for KMS encryption")
		} else if !isValidKMSKeyID(config.KMSKeyID) {
			v.addError("encryption.kms_key_id", config.KMSKeyID, "invalid KMS key ID format")
		}
	case "passphrase":
		if config.Passphrase == "" {
			v.addError("encryption.passphrase", config.Passphrase, "passphrase is required for passphrase encryption")
		} else if len(config.Passphrase) < 8 {
			v.addError("encryption.passphrase", "***", "passphrase must be at least 8 characters long")
		}
	}
}

// validateRetentionConfig validates retention configuration
func (v *Validator) validateRetentionConfig(config types.RetentionConfig) {
	if config.LocalCount < MinRetentionCount {
		v.addError("retention.local_count", config.LocalCount, 
			fmt.Sprintf("must be at least %d", MinRetentionCount))
	}
	if config.LocalCount > 1000 {
		v.addError("retention.local_count", config.LocalCount, "must not exceed 1000")
	}
	
	if config.RemoteCount < 1 {
		v.addError("retention.remote_count", config.RemoteCount, "must be at least 1")
	}
	if config.RemoteCount > 10000 {
		v.addError("retention.remote_count", config.RemoteCount, "must not exceed 10000")
	}
	
	if config.MaxAgeDays < 1 {
		v.addError("retention.max_age_days", config.MaxAgeDays, "must be at least 1")
	}
	if config.MaxAgeDays > 3650 { // 10 years
		v.addError("retention.max_age_days", config.MaxAgeDays, "must not exceed 3650 days (10 years)")
	}
}

// validateLoggingConfig validates logging configuration
func (v *Validator) validateLoggingConfig(config types.LoggingConfig) {
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, config.Level) {
		v.addError("logging.level", config.Level, 
			fmt.Sprintf("must be one of: %s", strings.Join(validLevels, ", ")))
	}
	
	validFormats := []string{"json", "text"}
	if !contains(validFormats, config.Format) {
		v.addError("logging.format", config.Format, 
			fmt.Sprintf("must be one of: %s", strings.Join(validFormats, ", ")))
	}
}

// Helper functions

func (v *Validator) addError(field string, value interface{}, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	})
}

func (v *Validator) buildValidationError() error {
	messages := make([]string, len(v.errors))
	for i, err := range v.errors {
		messages[i] = err.Error()
	}
	return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(messages, "\n  - "))
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isValidPath(path string) bool {
	// Basic path validation - no null bytes, reasonable length
	if strings.Contains(path, "\x00") {
		return false
	}
	if len(path) > 4096 {
		return false
	}
	return true
}

func checkWritePermission(dir string) error {
	// Try to create a temporary file to test write permission
	tempFile := filepath.Join(dir, ".tf-safe-write-test")
	file, err := os.Create(tempFile)
	if err != nil {
		return err
	}
	_ = file.Close()
	_ = os.Remove(tempFile)
	return nil
}

func isValidBucketName(name, provider string) bool {
	switch provider {
	case "s3":
		return isValidS3BucketName(name)
	case "gcs":
		return isValidGCSBucketName(name)
	case "azure":
		return isValidAzureBlobName(name)
	default:
		return false
	}
}

func isValidS3BucketName(name string) bool {
	// S3 bucket naming rules
	if len(name) < 3 || len(name) > 63 {
		return false
	}
	
	// Must start and end with lowercase letter or number
	matched, err := regexp.MatchString(`^[a-z0-9].*[a-z0-9]$`, name)
	if err != nil || !matched {
		return false
	}
	
	// Can contain lowercase letters, numbers, hyphens, and periods
	matched, err = regexp.MatchString(`^[a-z0-9.-]+$`, name)
	if err != nil || !matched {
		return false
	}
	
	// Cannot contain consecutive periods or hyphens
	if strings.Contains(name, "..") || strings.Contains(name, "--") {
		return false
	}
	
	return true
}

func isValidGCSBucketName(name string) bool {
	// GCS bucket naming rules (simplified)
	if len(name) < 3 || len(name) > 63 {
		return false
	}
	matched, err := regexp.MatchString(`^[a-z0-9][a-z0-9._-]*[a-z0-9]$`, name)
	if err != nil {
		return false
	}
	return matched
}

func isValidAzureBlobName(name string) bool {
	// Azure blob container naming rules (simplified)
	if len(name) < 3 || len(name) > 63 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`, name)
	return matched
}

func isValidAWSRegion(region string) bool {
	// AWS region format validation
	matched, _ := regexp.MatchString(`^[a-z]{2}-[a-z]+-[0-9]+$`, region)
	return matched
}

func isValidGCPRegion(region string) bool {
	// GCP region format validation
	matched, _ := regexp.MatchString(`^[a-z]+-[a-z0-9]+-[a-z]$`, region)
	return matched
}

func isValidPrefix(prefix string) bool {
	// Prefix should not start with / and should be a valid path
	if strings.HasPrefix(prefix, "/") {
		return false
	}
	if strings.Contains(prefix, "//") {
		return false
	}
	return isValidPath(prefix)
}

func isValidKMSKeyID(keyID string) bool {
	// AWS KMS key ID can be:
	// - Key ID: 1234abcd-12ab-34cd-56ef-1234567890ab
	// - Key ARN: arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab
	// - Alias: alias/example-alias
	// - Alias ARN: arn:aws:kms:us-west-2:111122223333:alias/example-alias
	
	patterns := []string{
		`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, // Key ID
		`^arn:aws:kms:[a-z0-9-]+:[0-9]{12}:key/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, // Key ARN
		`^alias/[a-zA-Z0-9/_-]+$`, // Alias
		`^arn:aws:kms:[a-z0-9-]+:[0-9]{12}:alias/[a-zA-Z0-9/_-]+$`, // Alias ARN
	}
	
	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, keyID)
		if matched {
			return true
		}
	}
	
	return false
}