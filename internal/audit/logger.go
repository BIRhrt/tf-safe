package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tf-safe/internal/utils"
	"tf-safe/pkg/types"

	"github.com/google/uuid"
)

// Logger handles writing audit logs
type Logger struct {
	config types.AuditConfig
	mu     sync.Mutex
}

// NewLogger creates a new audit logger
func NewLogger(cfg types.AuditConfig) *Logger {
	// Set default file if empty
	if cfg.File == "" {
		cfg.File = "tf-safe-audit.log"
	}

	// Ensure absolute path if not already
	if !filepath.IsAbs(cfg.File) {
		home, err := os.UserHomeDir()
		if err == nil {
			cfg.File = filepath.Join(home, ".tf-safe", cfg.File)
		}
	}

	return &Logger{
		config: cfg,
	}
}

// Log records an audit entry
func (l *Logger) Log(action types.AuditAction, status types.AuditStatus, details map[string]interface{}, err error) error {
	if !l.config.Enabled {
		return nil
	}

	currentUser, _ := user.Current()
	username := "unknown"
	if currentUser != nil {
		username = currentUser.Username
	}

	entry := types.AuditEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now().UTC(),
		Action:    action,
		Status:    status,
		User:      username,
		Command:   strings.Join(os.Args, " "),
		Details:   details,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	return l.writeEntry(entry)
}

// writeEntry appends the entry to the audit log file
func (l *Logger) writeEntry(entry types.AuditEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(l.config.File), 0755); err != nil {
		return fmt.Errorf("failed to create audit log directory: %w", err)
	}

	// Marshal entry to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	// Append to file
	f, err := os.OpenFile(l.config.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write audit entry: %w", err)
	}
	if _, err := f.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	return nil
}

// GetEntries returns audit entries, optionally filtered
func (l *Logger) GetEntries(limit int) ([]types.AuditEntry, error) {
	if !utils.FileExists(l.config.File) {
		return []types.AuditEntry{}, nil
	}

	// Read file
	content, err := os.ReadFile(l.config.File)
	if err != nil {
		return nil, fmt.Errorf("failed to read audit log file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var entries []types.AuditEntry

	// Parse lines in reverse order (newest first)
	count := 0
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		var entry types.AuditEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Skip malformed lines
			continue
		}

		entries = append(entries, entry)
		count++
		if limit > 0 && count >= limit {
			break
		}
	}

	return entries, nil
}
