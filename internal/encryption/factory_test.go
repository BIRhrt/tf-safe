package encryption

import (
	"context"
	"testing"

	"tf-safe/pkg/types"
)

func TestFactory_CreateFromConfig(t *testing.T) {
	factory := NewFactory()
	ctx := context.Background()

	tests := []struct {
		name        string
		config      types.EncryptionConfig
		expectError bool
	}{
		{
			name: "AES encryption",
			config: types.EncryptionConfig{
				Provider:   "aes",
				Passphrase: "test-passphrase-123",
			},
			expectError: false,
		},
		{
			name: "Passphrase encryption (same as AES)",
			config: types.EncryptionConfig{
				Provider:   "passphrase",
				Passphrase: "test-passphrase-456",
			},
			expectError: false,
		},
		{
			name: "No encryption",
			config: types.EncryptionConfig{
				Provider: "none",
			},
			expectError: false,
		},
		{
			name: "Empty provider (defaults to none)",
			config: types.EncryptionConfig{
				Provider: "",
			},
			expectError: false,
		},
		{
			name: "AES without passphrase",
			config: types.EncryptionConfig{
				Provider: "aes",
			},
			expectError: true,
		},
		{
			name: "Unsupported provider",
			config: types.EncryptionConfig{
				Provider: "unsupported",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := factory.CreateFromConfig(ctx, tt.config)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if provider == nil {
				t.Error("Expected provider but got nil")
			}

			// Test basic functionality
			testData := []byte("test data")
			encrypted, err := provider.Encrypt(ctx, testData)
			if err != nil {
				t.Fatalf("Failed to encrypt: %v", err)
			}

			decrypted, err := provider.Decrypt(ctx, encrypted)
			if err != nil {
				t.Fatalf("Failed to decrypt: %v", err)
			}

			if string(decrypted) != string(testData) {
				t.Error("Decrypted data doesn't match original")
			}
		})
	}
}

func TestExtractRegionFromKMSKey(t *testing.T) {
	tests := []struct {
		keyID    string
		expected string
	}{
		{
			keyID:    "arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012",
			expected: "us-west-2",
		},
		{
			keyID:    "arn:aws:kms:eu-central-1:123456789012:key/abcdef12-3456-7890-abcd-ef1234567890",
			expected: "eu-central-1",
		},
		{
			keyID:    "12345678-1234-1234-1234-123456789012",
			expected: "",
		},
		{
			keyID:    "alias/my-key",
			expected: "",
		},
		{
			keyID:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.keyID, func(t *testing.T) {
			result := extractRegionFromKMSKey(tt.keyID)
			if result != tt.expected {
				t.Errorf("Expected region '%s', got '%s'", tt.expected, result)
			}
		})
	}
}