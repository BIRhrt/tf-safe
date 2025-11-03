package types

import (
	"fmt"
	"time"
)

// TfSafeError represents a structured error with additional context
type TfSafeError struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *TfSafeError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewTfSafeError creates a new TfSafeError
func NewTfSafeError(code, message, details string) *TfSafeError {
	return &TfSafeError{
		Code:      code,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// Common error codes
const (
	ErrCodeConfig     = "CONFIG_ERROR"
	ErrCodeStorage    = "STORAGE_ERROR"
	ErrCodeEncryption = "ENCRYPTION_ERROR"
	ErrCodeTerraform  = "TERRAFORM_ERROR"
	ErrCodeValidation = "VALIDATION_ERROR"
	ErrCodeNotFound   = "NOT_FOUND"
	ErrCodePermission = "PERMISSION_ERROR"
)