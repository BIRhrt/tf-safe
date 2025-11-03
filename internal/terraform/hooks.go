package terraform

import (
	"context"
	"fmt"
	"os"
	"time"

	"tf-safe/internal/backup"
	"tf-safe/internal/config"
	"tf-safe/pkg/types"
)

// BackupHook implements CommandHook to provide automatic backup functionality
type BackupHook struct {
	configManager config.ConfigManager
	backupEngine  backup.BackupEngine
	stateDetector StateDetector
}

// NewBackupHook creates a new backup hook instance
func NewBackupHook(configManager config.ConfigManager, backupEngine backup.BackupEngine) *BackupHook {
	return &BackupHook{
		configManager: configManager,
		backupEngine:  backupEngine,
		stateDetector: NewStateDetector(),
	}
}

// PreExecute runs before Terraform command execution
func (h *BackupHook) PreExecute(ctx context.Context, cmd string, args []string) (*types.BackupMetadata, error) {
	// Check if this command should trigger a backup
	if !h.shouldCreateBackup(cmd) {
		return nil, nil
	}

	// Load configuration
	config, err := h.configManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if automatic backups are enabled
	if !config.Local.Enabled {
		return nil, nil
	}

	// Find state file
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	stateFiles, err := h.stateDetector.FindStateFiles(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to find state files: %w", err)
	}

	if len(stateFiles) == 0 {
		// No state file found - this is not an error for some commands
		fmt.Fprintf(os.Stderr, "Warning: No state file found for pre-operation backup\n")
		return nil, nil
	}

	// Create backup
	backupOpts := types.BackupOptions{
		StateFilePath: stateFiles[0],
		Description:   fmt.Sprintf("Pre-%s backup at %s", cmd, time.Now().Format(time.RFC3339)),
		Force:         false,
	}

	backup, err := h.backupEngine.CreateBackup(ctx, backupOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create pre-operation backup: %w", err)
	}

	fmt.Printf("Created pre-operation backup: %s\n", backup.ID)
	return backup, nil
}

// PostExecute runs after Terraform command execution
func (h *BackupHook) PostExecute(ctx context.Context, cmd string, args []string, preBackup *types.BackupMetadata) (*types.BackupMetadata, error) {
	// Check if this command should trigger a backup
	if !h.shouldCreateBackup(cmd) {
		return nil, nil
	}

	// Load configuration
	config, err := h.configManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if automatic backups are enabled
	if !config.Local.Enabled {
		return nil, nil
	}

	// Find state file
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	stateFiles, err := h.stateDetector.FindStateFiles(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to find state files: %w", err)
	}

	if len(stateFiles) == 0 {
		// No state file found - this might be normal for destroy operations
		fmt.Fprintf(os.Stderr, "Warning: No state file found for post-operation backup\n")
		return nil, nil
	}

	// Create backup
	backupOpts := types.BackupOptions{
		StateFilePath: stateFiles[0],
		Description:   fmt.Sprintf("Post-%s backup at %s", cmd, time.Now().Format(time.RFC3339)),
		Force:         false,
	}

	backup, err := h.backupEngine.CreateBackup(ctx, backupOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create post-operation backup: %w", err)
	}

	fmt.Printf("Created post-operation backup: %s\n", backup.ID)

	// Apply retention policies
	if err := h.backupEngine.CleanupOldBackups(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to cleanup old backups: %v\n", err)
	}

	return backup, nil
}

// OnError runs when Terraform command execution fails
func (h *BackupHook) OnError(ctx context.Context, cmd string, args []string, err error) error {
	// Log the error but don't fail the operation
	fmt.Fprintf(os.Stderr, "Terraform command failed: %v\n", err)
	
	// Could implement additional error handling here, such as:
	// - Creating an error backup
	// - Sending notifications
	// - Rolling back changes
	
	return nil
}

// shouldCreateBackup determines if a backup should be created for the given command
func (h *BackupHook) shouldCreateBackup(cmd string) bool {
	// Load configuration to check command-specific settings
	config, err := h.configManager.Load()
	if err != nil {
		// If we can't load config, fall back to default behavior
		return h.isModifyingCommand(cmd)
	}

	// Check command-specific auto-backup settings
	switch cmd {
	case "apply":
		return config.Commands.Apply.AutoBackup
	case "plan":
		return config.Commands.Plan.AutoBackup
	case "destroy":
		return config.Commands.Destroy.AutoBackup
	default:
		// For other commands, use default behavior
		return h.isModifyingCommand(cmd)
	}
}

// isModifyingCommand checks if a command modifies Terraform state
func (h *BackupHook) isModifyingCommand(cmd string) bool {
	// Commands that modify state should trigger backups by default
	modifyingCommands := map[string]bool{
		"apply":   true,
		"destroy": true,
		"import":  true,
		"refresh": true,
		"taint":   true,
		"untaint": true,
	}

	return modifyingCommands[cmd]
}

// LoggingHook implements CommandHook to provide logging functionality
type LoggingHook struct {
	verbose bool
}

// NewLoggingHook creates a new logging hook instance
func NewLoggingHook(verbose bool) *LoggingHook {
	return &LoggingHook{
		verbose: verbose,
	}
}

// PreExecute logs before command execution
func (h *LoggingHook) PreExecute(ctx context.Context, cmd string, args []string) (*types.BackupMetadata, error) {
	if h.verbose {
		fmt.Printf("Executing terraform %s %v\n", cmd, args)
	}
	return nil, nil
}

// PostExecute logs after command execution
func (h *LoggingHook) PostExecute(ctx context.Context, cmd string, args []string, preBackup *types.BackupMetadata) (*types.BackupMetadata, error) {
	if h.verbose {
		fmt.Printf("Completed terraform %s %v\n", cmd, args)
	}
	return nil, nil
}

// OnError logs when command execution fails
func (h *LoggingHook) OnError(ctx context.Context, cmd string, args []string, err error) error {
	if h.verbose {
		fmt.Printf("Failed terraform %s %v: %v\n", cmd, args, err)
	}
	return nil
}