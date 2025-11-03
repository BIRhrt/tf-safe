package config

import "tf-safe/pkg/types"

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *types.Config {
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
		Commands: types.CommandsConfig{
			Apply: types.CommandConfig{
				AutoBackup: true,
			},
			Plan: types.CommandConfig{
				AutoBackup: false, // Plan doesn't modify state, so default to false
			},
			Destroy: types.CommandConfig{
				AutoBackup: true,
			},
		},
	}
}

// DefaultLocalConfig returns default local storage configuration
func DefaultLocalConfig() types.LocalConfig {
	return types.LocalConfig{
		Enabled:        true,
		Path:           ".tfstate_snapshots",
		RetentionCount: 10,
	}
}

// DefaultRemoteConfig returns default remote storage configuration
func DefaultRemoteConfig() types.RemoteConfig {
	return types.RemoteConfig{
		Provider: "s3",
		Bucket:   "",
		Region:   "us-west-2",
		Prefix:   "",
		Enabled:  false,
	}
}

// DefaultEncryptionConfig returns default encryption configuration
func DefaultEncryptionConfig() types.EncryptionConfig {
	return types.EncryptionConfig{
		Provider:   "aes",
		KMSKeyID:   "",
		Passphrase: "",
	}
}

// DefaultRetentionConfig returns default retention configuration
func DefaultRetentionConfig() types.RetentionConfig {
	return types.RetentionConfig{
		LocalCount:  10,
		RemoteCount: 50,
		MaxAgeDays:  90,
	}
}

// DefaultLoggingConfig returns default logging configuration
func DefaultLoggingConfig() types.LoggingConfig {
	return types.LoggingConfig{
		Level:  "info",
		Format: "text",
	}
}

// DefaultCommandsConfig returns default commands configuration
func DefaultCommandsConfig() types.CommandsConfig {
	return types.CommandsConfig{
		Apply: types.CommandConfig{
			AutoBackup: true,
		},
		Plan: types.CommandConfig{
			AutoBackup: false, // Plan doesn't modify state
		},
		Destroy: types.CommandConfig{
			AutoBackup: true,
		},
	}
}

// Constants for configuration values
const (
	// Default paths
	DefaultLocalPath     = ".tfstate_snapshots"
	DefaultConfigFile    = ".tf-safe.yaml"
	DefaultGlobalConfig  = "~/.tf-safe/config.yaml"
	
	// Default retention values
	MinRetentionCount    = 3
	DefaultLocalRetention = 10
	DefaultRemoteRetention = 50
	DefaultMaxAgeDays    = 90
	
	// Default encryption
	DefaultEncryptionProvider = "aes"
	
	// Default logging
	DefaultLogLevel      = "info"
	DefaultLogFormat     = "text"
	
	// Default remote storage
	DefaultS3Region      = "us-west-2"
	DefaultRemoteProvider = "s3"
)