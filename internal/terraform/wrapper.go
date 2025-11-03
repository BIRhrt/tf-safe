package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"tf-safe/internal/backup"
	"tf-safe/internal/config"
	"tf-safe/pkg/types"
)

// Wrapper implements the TerraformWrapper interface
type Wrapper struct {
	configManager config.ConfigManager
	backupEngine  backup.BackupEngine
	stateDetector StateDetector
	hooks         []CommandHook
}

// NewWrapper creates a new Terraform wrapper instance
func NewWrapper(configManager config.ConfigManager, backupEngine backup.BackupEngine) *Wrapper {
	return &Wrapper{
		configManager: configManager,
		backupEngine:  backupEngine,
		stateDetector: NewStateDetector(),
		hooks:         []CommandHook{},
	}
}

// AddHook adds a command hook to the wrapper
func (w *Wrapper) AddHook(hook CommandHook) {
	w.hooks = append(w.hooks, hook)
}

// ExecuteWithBackup executes a Terraform command with automatic backup hooks
func (w *Wrapper) ExecuteWithBackup(ctx context.Context, cmd string, args []string) error {
	// Check if Terraform binary is available
	if err := w.CheckTerraformBinary(); err != nil {
		return fmt.Errorf("terraform binary check failed: %w", err)
	}

	// Detect state file (log warning but continue - some commands don't require state file)
	_, err := w.DetectStateFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not detect state file: %v\n", err)
	}

	var preBackup *types.BackupMetadata

	// Run pre-execution hooks
	for _, hook := range w.hooks {
		backup, err := hook.PreExecute(ctx, cmd, args)
		if err != nil {
			return fmt.Errorf("pre-execution hook failed: %w", err)
		}
		if backup != nil {
			preBackup = backup
		}
	}

	// Execute Terraform command
	terraformCmd := exec.CommandContext(ctx, "terraform", append([]string{cmd}, args...)...)
	terraformCmd.Stdout = os.Stdout
	terraformCmd.Stderr = os.Stderr
	terraformCmd.Stdin = os.Stdin

	err = terraformCmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}

		// Run error hooks
		for _, hook := range w.hooks {
			if hookErr := hook.OnError(ctx, cmd, args, err); hookErr != nil {
				fmt.Fprintf(os.Stderr, "Error hook failed: %v\n", hookErr)
			}
		}

		// Exit with the same code as Terraform
		os.Exit(exitCode)
	}

	// Run post-execution hooks
	for _, hook := range w.hooks {
		_, err := hook.PostExecute(ctx, cmd, args, preBackup)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Post-execution hook failed: %v\n", err)
		}
	}

	return nil
}

// DetectStateFile detects the Terraform state file in the current directory
func (w *Wrapper) DetectStateFile() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	stateFiles, err := w.stateDetector.FindStateFiles(cwd)
	if err != nil {
		return "", fmt.Errorf("failed to find state files: %w", err)
	}

	if len(stateFiles) == 0 {
		return "", fmt.Errorf("no terraform state file found in current directory")
	}

	if len(stateFiles) > 1 {
		return "", fmt.Errorf("multiple state files found: %v", stateFiles)
	}

	return stateFiles[0], nil
}

// ValidateStateFile validates that a file is a valid Terraform state file
func (w *Wrapper) ValidateStateFile(path string) error {
	isValid, err := w.stateDetector.IsValidStateFile(path)
	if err != nil {
		return fmt.Errorf("failed to validate state file: %w", err)
	}

	if !isValid {
		return fmt.Errorf("file is not a valid Terraform state file: %s", path)
	}

	return nil
}

// GetTerraformVersion returns the version of the Terraform binary
func (w *Wrapper) GetTerraformVersion() (string, error) {
	cmd := exec.Command("terraform", "version", "-json")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to plain version command
		cmd = exec.Command("terraform", "version")
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get terraform version: %w", err)
		}
		
		// Parse plain text version output
		versionRegex := regexp.MustCompile(`Terraform v(\d+\.\d+\.\d+)`)
		matches := versionRegex.FindStringSubmatch(string(output))
		if len(matches) < 2 {
			return "", fmt.Errorf("could not parse terraform version from output: %s", string(output))
		}
		return matches[1], nil
	}

	// Parse JSON version output
	var versionInfo struct {
		TerraformVersion string `json:"terraform_version"`
	}
	if err := json.Unmarshal(output, &versionInfo); err != nil {
		return "", fmt.Errorf("failed to parse terraform version JSON: %w", err)
	}

	return versionInfo.TerraformVersion, nil
}

// CheckTerraformBinary checks if Terraform binary is available and compatible
func (w *Wrapper) CheckTerraformBinary() error {
	// Check if terraform binary exists
	_, err := exec.LookPath("terraform")
	if err != nil {
		return fmt.Errorf("terraform binary not found in PATH: %w", err)
	}

	// Get version and check compatibility
	version, err := w.GetTerraformVersion()
	if err != nil {
		return fmt.Errorf("failed to get terraform version: %w", err)
	}

	// Check minimum version (0.12.0)
	if !isVersionCompatible(version, "0.12.0") {
		return fmt.Errorf("terraform version %s is not supported (minimum: 0.12.0)", version)
	}

	return nil
}

// isVersionCompatible checks if the given version meets the minimum requirement
func isVersionCompatible(version, minVersion string) bool {
	// Simple version comparison - in production, use a proper semver library
	versionParts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	minVersionParts := strings.Split(strings.TrimPrefix(minVersion, "v"), ".")

	if len(versionParts) < 3 || len(minVersionParts) < 3 {
		return false
	}

	// Compare major version
	if versionParts[0] > minVersionParts[0] {
		return true
	}
	if versionParts[0] < minVersionParts[0] {
		return false
	}

	// Compare minor version
	if versionParts[1] > minVersionParts[1] {
		return true
	}
	if versionParts[1] < minVersionParts[1] {
		return false
	}

	// Compare patch version
	return versionParts[2] >= minVersionParts[2]
}