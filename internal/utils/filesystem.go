package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// EnsureDir ensures that a directory exists, creating it if necessary
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := sourceFile.Close(); closeErr != nil {
			// Log error but don't override the main error
		}
	}()

	// Ensure destination directory exists
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := destFile.Close(); closeErr != nil {
			// Log error but don't override the main error
		}
	}()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// AtomicWrite writes data to a file atomically by writing to a temp file first
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	// Ensure directory exists
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}

	// Create temp file in same directory
	tempFile, err := os.CreateTemp(filepath.Dir(path), ".tmp-"+filepath.Base(path))
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()

	// Clean up temp file on error
	defer func() {
		if err != nil {
			_ = os.Remove(tempPath)
		}
	}()

	// Write data to temp file
	if _, err = tempFile.Write(data); err != nil {
		_ = tempFile.Close()
		return err
	}

	// Close temp file
	if err = tempFile.Close(); err != nil {
		return err
	}

	// Set permissions
	if err = os.Chmod(tempPath, perm); err != nil {
		return err
	}

	// Atomic rename
	return os.Rename(tempPath, path)
}

// CalculateChecksum calculates SHA256 checksum of a file
func CalculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log error but don't override the main error
		}
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// CalculateChecksumBytes calculates SHA256 checksum of byte data
func CalculateChecksumBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:])
}