package terraform

import (
	"context"
	"tf-safe/pkg/types"
)

// TerraformWrapper defines the interface for Terraform command wrapping
type TerraformWrapper interface {
	// ExecuteWithBackup executes a Terraform command with automatic backup hooks
	ExecuteWithBackup(ctx context.Context, cmd string, args []string) error
	
	// DetectStateFile detects the Terraform state file in the current directory
	DetectStateFile() (string, error)
	
	// ValidateStateFile validates that a file is a valid Terraform state file
	ValidateStateFile(path string) error
	
	// GetTerraformVersion returns the version of the Terraform binary
	GetTerraformVersion() (string, error)
	
	// CheckTerraformBinary checks if Terraform binary is available and compatible
	CheckTerraformBinary() error
}

// CommandHook defines the interface for pre/post operation hooks
type CommandHook interface {
	// PreExecute runs before Terraform command execution
	PreExecute(ctx context.Context, cmd string, args []string) (*types.BackupMetadata, error)
	
	// PostExecute runs after Terraform command execution
	PostExecute(ctx context.Context, cmd string, args []string, preBackup *types.BackupMetadata) (*types.BackupMetadata, error)
	
	// OnError runs when Terraform command execution fails
	OnError(ctx context.Context, cmd string, args []string, err error) error
}

// StateDetector defines the interface for Terraform state file detection
type StateDetector interface {
	// FindStateFiles finds all Terraform state files in a directory
	FindStateFiles(dir string) ([]string, error)
	
	// IsValidStateFile checks if a file is a valid Terraform state file
	IsValidStateFile(path string) (bool, error)
	
	// GetStateFileInfo returns information about a state file
	GetStateFileInfo(path string) (*StateFileInfo, error)
}

// StateFileInfo contains information about a Terraform state file
type StateFileInfo struct {
	Path            string `json:"path"`
	Size            int64  `json:"size"`
	ModTime         int64  `json:"mod_time"`
	TerraformVersion string `json:"terraform_version"`
	Serial          int64  `json:"serial"`
	Lineage         string `json:"lineage"`
}