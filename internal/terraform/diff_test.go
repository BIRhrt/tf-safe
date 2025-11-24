package terraform

import (
	"testing"
)

func TestParseState(t *testing.T) {
	stateJSON := `{
		"version": 4,
		"terraform_version": "1.0.0",
		"serial": 1,
		"lineage": "test-lineage",
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"provider": "provider.aws",
				"instances": [
					{
						"attributes": {
							"id": "i-1234",
							"instance_type": "t2.micro"
						}
					}
				]
			}
		]
	}`

	state, err := ParseState([]byte(stateJSON))
	if err != nil {
		t.Fatalf("ParseState() error = %v", err)
	}

	if state.Version != 4 {
		t.Errorf("Version = %d, want 4", state.Version)
	}
	if state.TerraformVersion != "1.0.0" {
		t.Errorf("TerraformVersion = %s, want 1.0.0", state.TerraformVersion)
	}
	if len(state.Resources) != 1 {
		t.Errorf("Resources count = %d, want 1", len(state.Resources))
	}
}

func TestCompareStates_NoChanges(t *testing.T) {
	stateJSON := `{
		"version": 4,
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"provider": "provider.aws",
				"instances": [{"attributes": {"id": "i-1234"}}]
			}
		]
	}`

	state1, _ := ParseState([]byte(stateJSON))
	state2, _ := ParseState([]byte(stateJSON))

	diff := CompareStates(state1, state2)

	if diff.HasDrift() {
		t.Error("Expected no drift, but changes detected")
	}
	if diff.Summary.TotalChanges != 0 {
		t.Errorf("TotalChanges = %d, want 0", diff.Summary.TotalChanges)
	}
}

func TestCompareStates_AddedResource(t *testing.T) {
	before := `{
		"version": 4,
		"resources": []
	}`

	after := `{
		"version": 4,
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"provider": "provider.aws",
				"instances": [{"attributes": {"id": "i-1234"}}]
			}
		]
	}`

	stateBefore, _ := ParseState([]byte(before))
	stateAfter, _ := ParseState([]byte(after))

	diff := CompareStates(stateBefore, stateAfter)

	if !diff.HasDrift() {
		t.Error("Expected drift, but no changes detected")
	}
	if diff.Summary.ResourcesAdded != 1 {
		t.Errorf("ResourcesAdded = %d, want 1", diff.Summary.ResourcesAdded)
	}
	if len(diff.Added) != 1 {
		t.Errorf("Added resources = %d, want 1", len(diff.Added))
	}
	if diff.Added[0].Type != "aws_instance" {
		t.Errorf("Added resource type = %s, want aws_instance", diff.Added[0].Type)
	}
}

func TestCompareStates_RemovedResource(t *testing.T) {
	before := `{
		"version": 4,
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"provider": "provider.aws",
				"instances": [{"attributes": {"id": "i-1234"}}]
			}
		]
	}`

	after := `{
		"version": 4,
		"resources": []
	}`

	stateBefore, _ := ParseState([]byte(before))
	stateAfter, _ := ParseState([]byte(after))

	diff := CompareStates(stateBefore, stateAfter)

	if !diff.HasDrift() {
		t.Error("Expected drift, but no changes detected")
	}
	if diff.Summary.ResourcesRemoved != 1 {
		t.Errorf("ResourcesRemoved = %d, want 1", diff.Summary.ResourcesRemoved)
	}
	if len(diff.Removed) != 1 {
		t.Errorf("Removed resources = %d, want 1", len(diff.Removed))
	}
}

func TestCompareStates_ModifiedResource(t *testing.T) {
	before := `{
		"version": 4,
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"provider": "provider.aws",
				"instances": [{"attributes": {"id": "i-1234", "instance_type": "t2.micro"}}]
			}
		]
	}`

	after := `{
		"version": 4,
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"provider": "provider.aws",
				"instances": [{"attributes": {"id": "i-1234", "instance_type": "t2.small"}}]
			}
		]
	}`

	stateBefore, _ := ParseState([]byte(before))
	stateAfter, _ := ParseState([]byte(after))

	diff := CompareStates(stateBefore, stateAfter)

	if !diff.HasDrift() {
		t.Error("Expected drift, but no changes detected")
	}
	if diff.Summary.ResourcesChanged != 1 {
		t.Errorf("ResourcesChanged = %d, want 1", diff.Summary.ResourcesChanged)
	}
	if len(diff.Modified) != 1 {
		t.Errorf("Modified resources = %d, want 1", len(diff.Modified))
	}

	// Check attribute diff
	if len(diff.Modified[0].AttributeDiff) == 0 {
		t.Error("Expected attribute changes, got none")
	}
	if _, exists := diff.Modified[0].AttributeDiff["instance_type"]; !exists {
		t.Error("Expected instance_type to be in attribute diff")
	}
}

func TestCompareStates_MultipleChanges(t *testing.T) {
	before := `{
		"version": 4,
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "web1",
				"provider": "provider.aws",
				"instances": [{"attributes": {"id": "i-1111"}}]
			},
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "web2",
				"provider": "provider.aws",
				"instances": [{"attributes": {"id": "i-2222"}}]
			}
		]
	}`

	after := `{
		"version": 4,
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "web2",
				"provider": "provider.aws",
				"instances": [{"attributes": {"id": "i-2222"}}]
			},
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "web3",
				"provider": "provider.aws",
				"instances": [{"attributes": {"id": "i-3333"}}]
			}
		]
	}`

	stateBefore, _ := ParseState([]byte(before))
	stateAfter, _ := ParseState([]byte(after))

	diff := CompareStates(stateBefore, stateAfter)

	if diff.Summary.ResourcesAdded != 1 {
		t.Errorf("ResourcesAdded = %d, want 1", diff.Summary.ResourcesAdded)
	}
	if diff.Summary.ResourcesRemoved != 1 {
		t.Errorf("ResourcesRemoved = %d, want 1", diff.Summary.ResourcesRemoved)
	}
	if diff.Summary.TotalChanges != 2 {
		t.Errorf("TotalChanges = %d, want 2", diff.Summary.TotalChanges)
	}
}

func TestFormatSummary(t *testing.T) {
	tests := []struct {
		name     string
		diff     *StateDiff
		expected string
	}{
		{
			name: "no changes",
			diff: &StateDiff{
				Summary: DiffSummary{HasChanges: false},
			},
			expected: "No changes detected",
		},
		{
			name: "only added",
			diff: &StateDiff{
				Summary: DiffSummary{
					HasChanges:     true,
					ResourcesAdded: 2,
				},
			},
			expected: "2 added",
		},
		{
			name: "multiple changes",
			diff: &StateDiff{
				Summary: DiffSummary{
					HasChanges:       true,
					ResourcesAdded:   1,
					ResourcesChanged: 2,
					ResourcesRemoved: 1,
				},
			},
			expected: "1 added, 2 changed, 1 removed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.diff.FormatSummary()
			if result != tt.expected {
				t.Errorf("FormatSummary() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestResourceAddress(t *testing.T) {
	tests := []struct {
		name     string
		resource *StateResource
		expected string
	}{
		{
			name: "no module",
			resource: &StateResource{
				Mode: "managed",
				Type: "aws_instance",
				Name: "web",
			},
			expected: "managed.aws_instance.web",
		},
		{
			name: "with module",
			resource: &StateResource{
				Mode:   "managed",
				Type:   "aws_instance",
				Name:   "web",
				Module: "module.networking",
			},
			expected: "module.networking.managed.aws_instance.web",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resourceAddress(tt.resource)
			if result != tt.expected {
				t.Errorf("resourceAddress() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCompareAttributes(t *testing.T) {
	before := map[string]interface{}{
		"id":            "i-1234",
		"instance_type": "t2.micro",
		"tags": map[string]interface{}{
			"Name": "web-server",
		},
	}

	after := map[string]interface{}{
		"id":            "i-1234",
		"instance_type": "t2.small",  // Changed
		"ami":           "ami-12345", // Added
		// tags removed
	}

	changes := compareAttributes(before, after)

	// Should have 3 changes: instance_type modified, tags removed, ami added
	if len(changes) != 3 {
		t.Errorf("Expected 3 attribute changes, got %d", len(changes))
	}

	// Check instance_type change
	if change, exists := changes["instance_type"]; exists {
		if change.Before != "t2.micro" || change.After != "t2.small" {
			t.Error("instance_type change not detected correctly")
		}
	} else {
		t.Error("instance_type change not found")
	}

	// Check tags removal
	if change, exists := changes["tags"]; exists {
		if change.After != nil {
			t.Error("tags should be set to nil (removed)")
		}
	} else {
		t.Error("tags removal not detected")
	}

	// Check ami addition
	if change, exists := changes["ami"]; exists {
		if change.Before != nil {
			t.Error("ami should have nil before value (added)")
		}
	} else {
		t.Error("ami addition not detected")
	}
}
