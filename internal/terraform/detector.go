package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultStateDetector implements the StateDetector interface
type DefaultStateDetector struct{}

// NewStateDetector creates a new state detector instance
func NewStateDetector() StateDetector {
	return &DefaultStateDetector{}
}

// FindStateFiles finds all Terraform state files in a directory
func (d *DefaultStateDetector) FindStateFiles(dir string) ([]string, error) {
	var stateFiles []string

	// Common state file names
	stateFileNames := []string{
		"terraform.tfstate",
		"terraform.tfstate.backup",
	}

	for _, fileName := range stateFileNames {
		filePath := filepath.Join(dir, fileName)
		if _, err := os.Stat(filePath); err == nil {
			// Only include the main state file, not backup
			if fileName == "terraform.tfstate" {
				stateFiles = append(stateFiles, filePath)
			}
		}
	}

	return stateFiles, nil
}

// IsValidStateFile checks if a file is a valid Terraform state file
func (d *DefaultStateDetector) IsValidStateFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log error but don't override the main error
		}
	}()

	// Try to parse as JSON
	var stateData map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&stateData); err != nil {
		return false, nil // Not valid JSON, so not a valid state file
	}

	// Check for required Terraform state fields
	requiredFields := []string{"version", "terraform_version", "serial"}
	for _, field := range requiredFields {
		if _, exists := stateData[field]; !exists {
			return false, nil
		}
	}

	// Check if it has resources or outputs (optional but common)
	if resources, exists := stateData["resources"]; exists {
		if _, ok := resources.([]interface{}); !ok {
			return false, nil
		}
	}

	return true, nil
}

// GetStateFileInfo returns information about a state file
func (d *DefaultStateDetector) GetStateFileInfo(path string) (*StateFileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open state file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log error but don't override the main error
		}
	}()

	// Get file stats
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	// Parse state file
	var stateData map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&stateData); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	info := &StateFileInfo{
		Path:    path,
		Size:    stat.Size(),
		ModTime: stat.ModTime().Unix(),
	}

	// Extract Terraform version
	if tfVersion, exists := stateData["terraform_version"]; exists {
		if version, ok := tfVersion.(string); ok {
			info.TerraformVersion = version
		}
	}

	// Extract serial number
	if serial, exists := stateData["serial"]; exists {
		if serialNum, ok := serial.(float64); ok {
			info.Serial = int64(serialNum)
		}
	}

	// Extract lineage
	if lineage, exists := stateData["lineage"]; exists {
		if lineageStr, ok := lineage.(string); ok {
			info.Lineage = lineageStr
		}
	}

	return info, nil
}