package encryption

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

// KMSProvider implements EncryptionProvider using AWS KMS
type KMSProvider struct {
	client  *kms.Client
	keyID   string
	keyInfo KeyInfo
	region  string
}

// NewKMSProvider creates a new KMS encryption provider
func NewKMSProvider(keyID, region string) (*KMSProvider, error) {
	if keyID == "" {
		return nil, fmt.Errorf("KMS key ID cannot be empty")
	}
	if region == "" {
		return nil, fmt.Errorf("AWS region cannot be empty")
	}

	provider := &KMSProvider{
		keyID:  keyID,
		region: region,
		keyInfo: KeyInfo{
			Type:        "KMS",
			KeyID:       keyID,
			Algorithm:   "AWS-KMS",
			KeySize:     0, // KMS manages key size
			Description: fmt.Sprintf("AWS KMS encryption with key %s in region %s", keyID, region),
		},
	}

	return provider, nil
}

// Initialize sets up the KMS client and validates the key
func (k *KMSProvider) Initialize(ctx context.Context) error {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(k.region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	k.client = kms.NewFromConfig(cfg)

	// Validate KMS key access
	if err := k.validateKeyAccess(ctx); err != nil {
		return fmt.Errorf("KMS key validation failed: %w", err)
	}

	return nil
}

// validateKeyAccess verifies that we can access the KMS key
func (k *KMSProvider) validateKeyAccess(ctx context.Context) error {
	// Try to describe the key to verify access
	input := &kms.DescribeKeyInput{
		KeyId: aws.String(k.keyID),
	}

	output, err := k.client.DescribeKey(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to describe KMS key: %w", err)
	}

	// Check if key is enabled
	if output.KeyMetadata.KeyState != types.KeyStateEnabled {
		return fmt.Errorf("KMS key %s is not enabled (state: %s)", k.keyID, output.KeyMetadata.KeyState)
	}

	// Update key info with actual key details
	k.keyInfo.KeyID = *output.KeyMetadata.KeyId
	if output.KeyMetadata.Description != nil {
		k.keyInfo.Description = fmt.Sprintf("AWS KMS key: %s", *output.KeyMetadata.Description)
	}

	return nil
}

// Encrypt encrypts data using AWS KMS
func (k *KMSProvider) Encrypt(ctx context.Context, data []byte) ([]byte, error) {
	if k.client == nil {
		return nil, fmt.Errorf("KMS provider not initialized")
	}

	input := &kms.EncryptInput{
		KeyId:     aws.String(k.keyID),
		Plaintext: data,
	}

	output, err := k.client.Encrypt(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data with KMS: %w", err)
	}

	return output.CiphertextBlob, nil
}

// Decrypt decrypts data using AWS KMS
func (k *KMSProvider) Decrypt(ctx context.Context, encryptedData []byte) ([]byte, error) {
	if k.client == nil {
		return nil, fmt.Errorf("KMS provider not initialized")
	}

	input := &kms.DecryptInput{
		CiphertextBlob: encryptedData,
	}

	output, err := k.client.Decrypt(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data with KMS: %w", err)
	}

	return output.Plaintext, nil
}

// GetKeyInfo returns information about the KMS key
func (k *KMSProvider) GetKeyInfo() KeyInfo {
	return k.keyInfo
}

// IsAvailable checks if KMS service is available
func (k *KMSProvider) IsAvailable(ctx context.Context) bool {
	if k.client == nil {
		return false
	}

	// Try a simple operation to check availability
	_, err := k.client.ListKeys(ctx, &kms.ListKeysInput{
		Limit: aws.Int32(1),
	})

	return err == nil
}