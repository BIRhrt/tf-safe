package encryption

import "context"

// EncryptionProvider defines the interface for encryption implementations
type EncryptionProvider interface {
	// Encrypt encrypts the provided data
	Encrypt(ctx context.Context, data []byte) ([]byte, error)
	
	// Decrypt decrypts the provided encrypted data
	Decrypt(ctx context.Context, encryptedData []byte) ([]byte, error)
	
	// GetKeyInfo returns information about the encryption key
	GetKeyInfo() KeyInfo
	
	// Initialize sets up the encryption provider
	Initialize(ctx context.Context) error
}

// KeyInfo contains information about encryption keys
type KeyInfo struct {
	Type        string `json:"type"`
	KeyID       string `json:"key_id,omitempty"`
	Algorithm   string `json:"algorithm"`
	KeySize     int    `json:"key_size"`
	Description string `json:"description,omitempty"`
}

// EncryptionFactory creates encryption provider instances
type EncryptionFactory interface {
	CreateAES(passphrase string) (EncryptionProvider, error)
	CreateKMS(keyID string, region string) (EncryptionProvider, error)
}