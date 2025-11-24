package terraform

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// StateDiff represents the difference between two Terraform states
type StateDiff struct {
	Added    []ResourceChange `json:"added"`
	Modified []ResourceChange `json:"modified"`
	Removed  []ResourceChange `json:"removed"`
	Summary  DiffSummary      `json:"summary"`
}

// ResourceChange represents a change to a resource
type ResourceChange struct {
	Address       string                     `json:"address"`
	Type          string                     `json:"type"`
	Name          string                     `json:"name"`
	Module        string                     `json:"module,omitempty"`
	Provider      string                     `json:"provider"`
	ChangeType    string                     `json:"change_type"` // "added", "modified", "removed"
	Before        map[string]interface{}     `json:"before,omitempty"`
	After         map[string]interface{}     `json:"after,omitempty"`
	AttributeDiff map[string]AttributeChange `json:"attribute_diff,omitempty"`
}

// AttributeChange represents a change to a specific attribute
type AttributeChange struct {
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

// DiffSummary provides high-level statistics about the diff
type DiffSummary struct {
	TotalChanges     int  `json:"total_changes"`
	ResourcesAdded   int  `json:"resources_added"`
	ResourcesChanged int  `json:"resources_changed"`
	ResourcesRemoved int  `json:"resources_removed"`
	HasChanges       bool `json:"has_changes"`
}

// TerraformState represents a Terraform state file
type TerraformState struct {
	Version          int               `json:"version"`
	TerraformVersion string            `json:"terraform_version"`
	Serial           int               `json:"serial"`
	Lineage          string            `json:"lineage"`
	Resources        []StateResource   `json:"resources"`
	Outputs          map[string]Output `json:"outputs,omitempty"`
}

// StateResource represents a resource in the state
type StateResource struct {
	Mode      string                   `json:"mode"`
	Type      string                   `json:"type"`
	Name      string                   `json:"name"`
	Provider  string                   `json:"provider"`
	Module    string                   `json:"module,omitempty"`
	Instances []map[string]interface{} `json:"instances"`
}

// Output represents a Terraform output value
type Output struct {
	Value     interface{} `json:"value"`
	Type      string      `json:"type,omitempty"`
	Sensitive bool        `json:"sensitive,omitempty"`
}

// ParseState parses a Terraform state file from JSON bytes
func ParseState(data []byte) (*TerraformState, error) {
	var state TerraformState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}
	return &state, nil
}

// CompareStates compares two Terraform states and returns the differences
func CompareStates(before, after *TerraformState) *StateDiff {
	diff := &StateDiff{
		Added:    []ResourceChange{},
		Modified: []ResourceChange{},
		Removed:  []ResourceChange{},
	}

	// Create maps for efficient lookup
	beforeResources := mapResources(before)
	afterResources := mapResources(after)

	// Find removed and modified resources
	for addr, beforeRes := range beforeResources {
		if afterRes, exists := afterResources[addr]; exists {
			// Resource exists in both - check if modified
			if isResourceModified(beforeRes, afterRes) {
				change := ResourceChange{
					Address:    addr,
					Type:       beforeRes.Type,
					Name:       beforeRes.Name,
					Module:     beforeRes.Module,
					Provider:   beforeRes.Provider,
					ChangeType: "modified",
					Before:     flattenInstances(beforeRes.Instances),
					After:      flattenInstances(afterRes.Instances),
				}
				change.AttributeDiff = compareAttributes(
					flattenInstances(beforeRes.Instances),
					flattenInstances(afterRes.Instances),
				)
				diff.Modified = append(diff.Modified, change)
			}
		} else {
			// Resource was removed
			diff.Removed = append(diff.Removed, ResourceChange{
				Address:    addr,
				Type:       beforeRes.Type,
				Name:       beforeRes.Name,
				Module:     beforeRes.Module,
				Provider:   beforeRes.Provider,
				ChangeType: "removed",
				Before:     flattenInstances(beforeRes.Instances),
			})
		}
	}

	// Find added resources
	for addr, afterRes := range afterResources {
		if _, exists := beforeResources[addr]; !exists {
			diff.Added = append(diff.Added, ResourceChange{
				Address:    addr,
				Type:       afterRes.Type,
				Name:       afterRes.Name,
				Module:     afterRes.Module,
				Provider:   afterRes.Provider,
				ChangeType: "added",
				After:      flattenInstances(afterRes.Instances),
			})
		}
	}

	// Sort for consistent output
	sort.Slice(diff.Added, func(i, j int) bool {
		return diff.Added[i].Address < diff.Added[j].Address
	})
	sort.Slice(diff.Modified, func(i, j int) bool {
		return diff.Modified[i].Address < diff.Modified[j].Address
	})
	sort.Slice(diff.Removed, func(i, j int) bool {
		return diff.Removed[i].Address < diff.Removed[j].Address
	})

	// Calculate summary
	diff.Summary = DiffSummary{
		ResourcesAdded:   len(diff.Added),
		ResourcesChanged: len(diff.Modified),
		ResourcesRemoved: len(diff.Removed),
		TotalChanges:     len(diff.Added) + len(diff.Modified) + len(diff.Removed),
		HasChanges:       len(diff.Added) > 0 || len(diff.Modified) > 0 || len(diff.Removed) > 0,
	}

	return diff
}

// mapResources creates a map of resources by their address
func mapResources(state *TerraformState) map[string]*StateResource {
	resources := make(map[string]*StateResource)
	for i := range state.Resources {
		res := &state.Resources[i]
		addr := resourceAddress(res)
		resources[addr] = res
	}
	return resources
}

// resourceAddress generates a unique address for a resource
func resourceAddress(res *StateResource) string {
	if res.Module != "" {
		return fmt.Sprintf("%s.%s.%s.%s", res.Module, res.Mode, res.Type, res.Name)
	}
	return fmt.Sprintf("%s.%s.%s", res.Mode, res.Type, res.Name)
}

// isResourceModified checks if a resource has been modified
func isResourceModified(before, after *StateResource) bool {
	beforeJSON, _ := json.Marshal(before.Instances)
	afterJSON, _ := json.Marshal(after.Instances)
	return string(beforeJSON) != string(afterJSON)
}

// flattenInstances flattens instance data into a single map
func flattenInstances(instances []map[string]interface{}) map[string]interface{} {
	if len(instances) == 0 {
		return make(map[string]interface{})
	}
	// For simplicity, return the first instance's attributes
	if attrs, ok := instances[0]["attributes"].(map[string]interface{}); ok {
		return attrs
	}
	return make(map[string]interface{})
}

// compareAttributes compares two attribute maps and returns differences
func compareAttributes(before, after map[string]interface{}) map[string]AttributeChange {
	changes := make(map[string]AttributeChange)

	// Find changed and removed attributes
	for key, beforeValue := range before {
		if afterValue, exists := after[key]; exists {
			// Compare values
			beforeJSON, _ := json.Marshal(beforeValue)
			afterJSON, _ := json.Marshal(afterValue)
			if string(beforeJSON) != string(afterJSON) {
				changes[key] = AttributeChange{
					Before: beforeValue,
					After:  afterValue,
				}
			}
		} else {
			// Attribute removed
			changes[key] = AttributeChange{
				Before: beforeValue,
				After:  nil,
			}
		}
	}

	// Find added attributes
	for key, afterValue := range after {
		if _, exists := before[key]; !exists {
			changes[key] = AttributeChange{
				Before: nil,
				After:  afterValue,
			}
		}
	}

	return changes
}

// HasDrift checks if there are any changes in the diff
func (d *StateDiff) HasDrift() bool {
	return d.Summary.HasChanges
}

// GetChangeCount returns the total number of changes
func (d *StateDiff) GetChangeCount() int {
	return d.Summary.TotalChanges
}

// FormatSummary returns a human-readable summary of changes
func (d *StateDiff) FormatSummary() string {
	if !d.HasDrift() {
		return "No changes detected"
	}

	parts := []string{}
	if d.Summary.ResourcesAdded > 0 {
		parts = append(parts, fmt.Sprintf("%d added", d.Summary.ResourcesAdded))
	}
	if d.Summary.ResourcesChanged > 0 {
		parts = append(parts, fmt.Sprintf("%d changed", d.Summary.ResourcesChanged))
	}
	if d.Summary.ResourcesRemoved > 0 {
		parts = append(parts, fmt.Sprintf("%d removed", d.Summary.ResourcesRemoved))
	}

	return strings.Join(parts, ", ")
}
