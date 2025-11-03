package encryption

import "context"

// NoOpProvider implements EncryptionProvider with no encryption (pass-through)
type NoOpProvider struct {
	keyInfo KeyInfo
}

// NewNoOpProvider creates a new no-op encryption provider
func NewNoOpProvider() *NoOpProvider {
	return &NoOpProvider{
		keyInfo: KeyInfo{
			Type:        "None",
			Algorithm:   "None",
			KeySize:     0,
			Description: "No encryption (pass-through)",
		},
	}
}

// Initialize is a no-op for the no-op provider
func (n *NoOpProvider) Initialize(ctx context.Context) error {
	return nil
}

// Encrypt returns the data unchanged
func (n *NoOpProvider) Encrypt(ctx context.Context, data []byte) ([]byte, error) {
	// Return a copy to prevent modification of original data
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

// Decrypt returns the data unchanged
func (n *NoOpProvider) Decrypt(ctx context.Context, encryptedData []byte) ([]byte, error) {
	// Return a copy to prevent modification of original data
	result := make([]byte, len(encryptedData))
	copy(result, encryptedData)
	return result, nil
}

// GetKeyInfo returns information about the no-op provider
func (n *NoOpProvider) GetKeyInfo() KeyInfo {
	return n.keyInfo
}