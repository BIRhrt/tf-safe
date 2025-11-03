package encryption

import (
	"context"
	"testing"
)

func TestAESProvider_EncryptDecrypt(t *testing.T) {
	// Test with passphrase
	provider, err := NewAESProvider("test-passphrase-123")
	if err != nil {
		t.Fatalf("Failed to create AES provider: %v", err)
	}

	ctx := context.Background()
	if err := provider.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize AES provider: %v", err)
	}

	// Test data
	originalData := []byte("This is test data for encryption")

	// Encrypt
	encrypted, err := provider.Encrypt(ctx, originalData)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	// Verify encrypted data is different from original
	if string(encrypted) == string(originalData) {
		t.Error("Encrypted data should be different from original")
	}

	// Decrypt
	decrypted, err := provider.Decrypt(ctx, encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}

	// Verify decrypted data matches original
	if string(decrypted) != string(originalData) {
		t.Errorf("Decrypted data doesn't match original. Got: %s, Want: %s", string(decrypted), string(originalData))
	}
}

func TestAESProvider_KeyInfo(t *testing.T) {
	provider, err := NewAESProvider("test-passphrase")
	if err != nil {
		t.Fatalf("Failed to create AES provider: %v", err)
	}

	keyInfo := provider.GetKeyInfo()
	
	if keyInfo.Type != "AES" {
		t.Errorf("Expected key type 'AES', got '%s'", keyInfo.Type)
	}
	
	if keyInfo.Algorithm != "AES-256-GCM" {
		t.Errorf("Expected algorithm 'AES-256-GCM', got '%s'", keyInfo.Algorithm)
	}
	
	if keyInfo.KeySize != 256 {
		t.Errorf("Expected key size 256, got %d", keyInfo.KeySize)
	}
}

func TestAESProvider_WithGeneratedKey(t *testing.T) {
	provider, err := GenerateAESProvider()
	if err != nil {
		t.Fatalf("Failed to generate AES provider: %v", err)
	}

	ctx := context.Background()
	if err := provider.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize AES provider: %v", err)
	}

	// Test encryption/decryption
	originalData := []byte("Test data with generated key")
	
	encrypted, err := provider.Encrypt(ctx, originalData)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	decrypted, err := provider.Decrypt(ctx, encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}

	if string(decrypted) != string(originalData) {
		t.Errorf("Decrypted data doesn't match original")
	}
}

func TestAESProvider_InvalidData(t *testing.T) {
	provider, err := NewAESProvider("test-passphrase")
	if err != nil {
		t.Fatalf("Failed to create AES provider: %v", err)
	}

	ctx := context.Background()
	if err := provider.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize AES provider: %v", err)
	}

	// Test decryption with invalid data
	_, err = provider.Decrypt(ctx, []byte("invalid-encrypted-data"))
	if err == nil {
		t.Error("Expected error when decrypting invalid data")
	}

	// Test decryption with too short data
	_, err = provider.Decrypt(ctx, []byte("short"))
	if err == nil {
		t.Error("Expected error when decrypting too short data")
	}
}