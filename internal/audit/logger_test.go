package audit

import (
	"os"
	"path/filepath"
	"testing"

	"tf-safe/pkg/types"
)

func TestLogger_Log(t *testing.T) {
	// Create temp dir for test logs
	tmpDir, err := os.MkdirTemp("", "tf-safe-audit-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "audit.log")

	config := types.AuditConfig{
		Enabled: true,
		File:    logFile,
	}

	logger := NewLogger(config)

	// Test logging success
	err = logger.Log(types.AuditActionBackupCreate, types.AuditStatusSuccess, map[string]interface{}{"id": "123"}, nil)
	if err != nil {
		t.Errorf("Log() error = %v", err)
	}

	// Test logging failure
	err = logger.Log(types.AuditActionRestore, types.AuditStatusFailed, nil, os.ErrPermission)
	if err != nil {
		t.Errorf("Log() error = %v", err)
	}

	// Verify file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Should have 2 lines (plus empty line at end)
	lines := 0
	for _, line := range string(content) {
		if line == '\n' {
			lines++
		}
	}
	if lines != 2 {
		t.Errorf("Expected 2 log entries, got %d", lines)
	}

	// Verify entries
	entries, err := logger.GetEntries(10)
	if err != nil {
		t.Fatalf("GetEntries() error = %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	// Most recent first
	if entries[0].Action != types.AuditActionRestore {
		t.Errorf("Expected first entry action %s, got %s", types.AuditActionRestore, entries[0].Action)
	}
	if entries[0].Status != types.AuditStatusFailed {
		t.Errorf("Expected first entry status %s, got %s", types.AuditStatusFailed, entries[0].Status)
	}

	if entries[1].Action != types.AuditActionBackupCreate {
		t.Errorf("Expected second entry action %s, got %s", types.AuditActionBackupCreate, entries[1].Action)
	}
}

func TestLogger_Disabled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tf-safe-audit-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	config := types.AuditConfig{
		Enabled: false,
		File:    filepath.Join(tmpDir, "audit.log"),
	}

	logger := NewLogger(config)

	err = logger.Log(types.AuditActionBackupCreate, types.AuditStatusSuccess, nil, nil)
	if err != nil {
		t.Errorf("Log() error = %v", err)
	}

	// File should not exist
	if _, err := os.Stat(config.File); !os.IsNotExist(err) {
		t.Error("Log file should not exist when disabled")
	}
}
