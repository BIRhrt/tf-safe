package encryption

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// AESProvider implements EncryptionProvider using AES-256-GCM
type AESProvider struct {
	key       []byte
	keyInfo   KeyInfo
	gcm       cipher.AEAD
	keySource string // "passphrase" or "generated"
}

// NewAESProvider creates a new AES encryption provider with a passphrase
func NewAESProvider(passphrase string) (*AESProvider, error) {
	if passphrase == "" {
		return nil, fmt.Errorf("passphrase cannot be empty")
	}

	// Generate a random salt for key derivation
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key using PBKDF2 with SHA-256
	key := pbkdf2.Key([]byte(passphrase), salt, 100000, 32, sha256.New)

	provider := &AESProvider{
		key:       key,
		keySource: "passphrase",
		keyInfo: KeyInfo{
			Type:        "AES",
			Algorithm:   "AES-256-GCM",
			KeySize:     256,
			Description: "AES-256-GCM with PBKDF2 key derivation",
		},
	}

	return provider, nil
}

// NewAESProviderWithKey creates a new AES encryption provider with a provided key
func NewAESProviderWithKey(key []byte) (*AESProvider, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes for AES-256")
	}

	provider := &AESProvider{
		key:       key,
		keySource: "provided",
		keyInfo: KeyInfo{
			Type:        "AES",
			Algorithm:   "AES-256-GCM",
			KeySize:     256,
			Description: "AES-256-GCM with provided key",
		},
	}

	return provider, nil
}

// GenerateAESProvider creates a new AES encryption provider with a randomly generated key
func GenerateAESProvider() (*AESProvider, error) {
	key := make([]byte, 32) // 256 bits
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	provider := &AESProvider{
		key:       key,
		keySource: "generated",
		keyInfo: KeyInfo{
			Type:        "AES",
			Algorithm:   "AES-256-GCM",
			KeySize:     256,
			Description: "AES-256-GCM with randomly generated key",
		},
	}

	return provider, nil
}

// Initialize sets up the AES-GCM cipher
func (a *AESProvider) Initialize(ctx context.Context) error {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	a.gcm = gcm
	return nil
}

// Encrypt encrypts data using AES-256-GCM
func (a *AESProvider) Encrypt(ctx context.Context, data []byte) ([]byte, error) {
	if a.gcm == nil {
		return nil, fmt.Errorf("encryption provider not initialized")
	}

	// Generate a random nonce
	nonce := make([]byte, a.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := a.gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-256-GCM
func (a *AESProvider) Decrypt(ctx context.Context, encryptedData []byte) ([]byte, error) {
	if a.gcm == nil {
		return nil, fmt.Errorf("encryption provider not initialized")
	}

	nonceSize := a.gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	// Decrypt the data
	plaintext, err := a.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// GetKeyInfo returns information about the encryption key
func (a *AESProvider) GetKeyInfo() KeyInfo {
	return a.keyInfo
}

// GetKey returns the encryption key (for testing purposes only)
func (a *AESProvider) GetKey() []byte {
	// Return a copy to prevent modification
	key := make([]byte, len(a.key))
	copy(key, a.key)
	return key
}