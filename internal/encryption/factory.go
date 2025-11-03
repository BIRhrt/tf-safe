package encryption

import (
	"context"
	"fmt"

	"tf-safe/pkg/types"
)

// Factory implements EncryptionFactory interface
type Factory struct{}

// NewFactory creates a new encryption factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateAES creates an AES encryption provider with a passphrase
func (f *Factory) CreateAES(passphrase string) (EncryptionProvider, error) {
	provider, err := NewAESProvider(passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES provider: %w", err)
	}
	return provider, nil
}

// CreateKMS creates a KMS encryption provider
func (f *Factory) CreateKMS(keyID string, region string) (EncryptionProvider, error) {
	provider, err := NewKMSProvider(keyID, region)
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS provider: %w", err)
	}
	return provider, nil
}

// CreateFromConfig creates an encryption provider based on configuration
func (f *Factory) CreateFromConfig(ctx context.Context, config types.EncryptionConfig) (EncryptionProvider, error) {
	switch config.Provider {
	case "aes":
		if config.Passphrase == "" {
			return nil, fmt.Errorf("passphrase is required for AES encryption")
		}
		provider, err := f.CreateAES(config.Passphrase)
		if err != nil {
			return nil, err
		}
		if err := provider.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize AES provider: %w", err)
		}
		return provider, nil

	case "kms":
		if config.KMSKeyID == "" {
			return nil, fmt.Errorf("KMS key ID is required for KMS encryption")
		}
		// Extract region from key ID if it's an ARN, otherwise use default
		region := extractRegionFromKMSKey(config.KMSKeyID)
		if region == "" {
			region = "us-east-1" // Default region
		}
		
		provider, err := f.CreateKMS(config.KMSKeyID, region)
		if err != nil {
			return nil, err
		}
		if err := provider.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize KMS provider: %w", err)
		}
		return provider, nil

	case "passphrase":
		// Same as AES for backward compatibility
		if config.Passphrase == "" {
			return nil, fmt.Errorf("passphrase is required for passphrase encryption")
		}
		provider, err := f.CreateAES(config.Passphrase)
		if err != nil {
			return nil, err
		}
		if err := provider.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize passphrase provider: %w", err)
		}
		return provider, nil

	case "none", "":
		return NewNoOpProvider(), nil

	default:
		return nil, fmt.Errorf("unsupported encryption provider: %s", config.Provider)
	}
}

// extractRegionFromKMSKey extracts AWS region from KMS key ARN
func extractRegionFromKMSKey(keyID string) string {
	// If it's an ARN, extract region: arn:aws:kms:region:account:key/key-id
	if len(keyID) > 20 && keyID[:8] == "arn:aws:" {
		parts := []string{}
		start := 0
		for i, char := range keyID {
			if char == ':' {
				parts = append(parts, keyID[start:i])
				start = i + 1
			}
		}
		// Add the last part
		if start < len(keyID) {
			parts = append(parts, keyID[start:])
		}
		
		// ARN format: arn:aws:kms:region:account:key/key-id
		// parts[0] = "arn", parts[1] = "aws", parts[2] = "kms", parts[3] = "region"
		if len(parts) >= 4 {
			return parts[3]
		}
	}
	return ""
}