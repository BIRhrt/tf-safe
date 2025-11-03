package config

import "tf-safe/pkg/types"

// ConfigTemplate represents a configuration template
type ConfigTemplate struct {
	Name        string
	Description string
	Config      *types.Config
}

// GetAvailableTemplates returns all available configuration templates
func GetAvailableTemplates() []ConfigTemplate {
	return []ConfigTemplate{
		{
			Name:        "default",
			Description: "Standard configuration with local backups and AES encryption",
			Config:      getDefaultTemplate(),
		},
		{
			Name:        "minimal",
			Description: "Minimal configuration with only local backups, no encryption",
			Config:      getMinimalTemplate(),
		},
		{
			Name:        "enterprise",
			Description: "Enterprise configuration with S3 remote storage and KMS encryption",
			Config:      getEnterpriseTemplate(),
		},
		{
			Name:        "local-only",
			Description: "Local-only configuration with enhanced retention",
			Config:      getLocalOnlyTemplate(),
		},
		{
			Name:        "cloud-native",
			Description: "Cloud-native configuration optimized for CI/CD pipelines",
			Config:      getCloudNativeTemplate(),
		},
	}
}

// GetTemplate returns a specific template by name
func GetTemplate(name string) (*types.Config, bool) {
	templates := GetAvailableTemplates()
	for _, template := range templates {
		if template.Name == name {
			return template.Config, true
		}
	}
	return nil, false
}

func getDefaultTemplate() *types.Config {
	return &types.Config{
		Local: types.LocalConfig{
			Enabled:        true,
			Path:           ".tfstate_snapshots",
			RetentionCount: 10,
		},
		Remote: types.RemoteConfig{
			Provider: "s3",
			Bucket:   "",
			Region:   "us-west-2",
			Prefix:   "",
			Enabled:  false,
		},
		Encryption: types.EncryptionConfig{
			Provider:   "aes",
			KMSKeyID:   "",
			Passphrase: "",
		},
		Retention: types.RetentionConfig{
			LocalCount:  10,
			RemoteCount: 50,
			MaxAgeDays:  90,
		},
		Logging: types.LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

func getMinimalTemplate() *types.Config {
	return &types.Config{
		Local: types.LocalConfig{
			Enabled:        true,
			Path:           ".tfstate_snapshots",
			RetentionCount: 5,
		},
		Remote: types.RemoteConfig{
			Enabled: false,
		},
		Encryption: types.EncryptionConfig{
			Provider: "none",
		},
		Retention: types.RetentionConfig{
			LocalCount:  5,
			RemoteCount: 10,
			MaxAgeDays:  30,
		},
		Logging: types.LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

func getEnterpriseTemplate() *types.Config {
	return &types.Config{
		Local: types.LocalConfig{
			Enabled:        true,
			Path:           ".tfstate_snapshots",
			RetentionCount: 20,
		},
		Remote: types.RemoteConfig{
			Provider: "s3",
			Bucket:   "your-terraform-backups",
			Region:   "us-west-2",
			Prefix:   "terraform-state/",
			Enabled:  true,
		},
		Encryption: types.EncryptionConfig{
			Provider: "kms",
			KMSKeyID: "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012",
		},
		Retention: types.RetentionConfig{
			LocalCount:  20,
			RemoteCount: 100,
			MaxAgeDays:  365,
		},
		Logging: types.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

func getLocalOnlyTemplate() *types.Config {
	return &types.Config{
		Local: types.LocalConfig{
			Enabled:        true,
			Path:           ".tfstate_snapshots",
			RetentionCount: 50,
		},
		Remote: types.RemoteConfig{
			Enabled: false,
		},
		Encryption: types.EncryptionConfig{
			Provider: "aes",
		},
		Retention: types.RetentionConfig{
			LocalCount:  50,
			RemoteCount: 0,
			MaxAgeDays:  180,
		},
		Logging: types.LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

func getCloudNativeTemplate() *types.Config {
	return &types.Config{
		Local: types.LocalConfig{
			Enabled:        false, // Disable local storage for CI/CD
			Path:           "",
			RetentionCount: 0,
		},
		Remote: types.RemoteConfig{
			Provider: "s3",
			Bucket:   "ci-terraform-backups",
			Region:   "us-west-2",
			Prefix:   "projects/",
			Enabled:  true,
		},
		Encryption: types.EncryptionConfig{
			Provider: "kms",
			KMSKeyID: "", // To be filled by user
		},
		Retention: types.RetentionConfig{
			LocalCount:  0,
			RemoteCount: 200,
			MaxAgeDays:  730, // 2 years
		},
		Logging: types.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

// GenerateExampleYAML returns a YAML string with example configuration and comments
func GenerateExampleYAML() string {
	return `# tf-safe configuration file
# This file configures tf-safe backup and restore behavior

# Local storage configuration
local:
  # Enable local backups (stored in filesystem)
  enabled: true
  
  # Directory to store local backups
  path: ".tfstate_snapshots"
  
  # Number of local backup versions to retain (minimum: 3)
  retention_count: 10

# Remote storage configuration
remote:
  # Enable remote backups (cloud storage)
  enabled: false
  
  # Storage provider: s3, gcs, or azure
  provider: "s3"
  
  # Bucket/container name for remote storage
  bucket: "your-terraform-backups"
  
  # Region for cloud storage (required for S3)
  region: "us-west-2"
  
  # Prefix for backup objects (optional)
  prefix: "terraform-state/"

# Encryption configuration
encryption:
  # Encryption provider: none, aes, kms, or passphrase
  provider: "aes"
  
  # KMS key ID or ARN (required for kms provider)
  kms_key_id: ""
  
  # Passphrase for encryption (required for passphrase provider)
  # Note: This will be stored in plaintext in the config file
  passphrase: ""

# Backup retention policies
retention:
  # Number of local backups to keep (minimum: 3)
  local_count: 10
  
  # Number of remote backups to keep (minimum: 1)
  remote_count: 50
  
  # Maximum age of backups in days (minimum: 1)
  max_age_days: 90

# Logging configuration
logging:
  # Log level: debug, info, warn, or error
  level: "info"
  
  # Log format: text or json
  format: "text"
`
}