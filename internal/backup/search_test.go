package backup

import (
	"testing"
	"time"

	"tf-safe/internal/storage"
	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

func TestSearchEngine_FilterByTags(t *testing.T) {
	// Create mock backups
	backups := []*types.BackupMetadata{
		{
			ID:        "backup-1",
			Timestamp: time.Now().Add(-1 * time.Hour),
			Tags: types.Tags{
				"env":  "prod",
				"team": "platform",
			},
		},
		{
			ID:        "backup-2",
			Timestamp: time.Now().Add(-2 * time.Hour),
			Tags: types.Tags{
				"env":  "staging",
				"team": "platform",
			},
		},
		{
			ID:        "backup-3",
			Timestamp: time.Now().Add(-3 * time.Hour),
			Tags: types.Tags{
				"env":  "prod",
				"team": "data",
			},
		},
		{
			ID:        "backup-4",
			Timestamp: time.Now().Add(-4 * time.Hour),
			Tags:      nil, // No tags
		},
	}

	tests := []struct {
		name      string
		filter    types.Tags
		wantCount int
		wantIDs   []string
	}{
		{
			name:      "filter by single tag",
			filter:    types.Tags{"env": "prod"},
			wantCount: 2,
			wantIDs:   []string{"backup-1", "backup-3"},
		},
		{
			name:      "filter by multiple tags (AND)",
			filter:    types.Tags{"env": "prod", "team": "platform"},
			wantCount: 1,
			wantIDs:   []string{"backup-1"},
		},
		{
			name:      "no matches",
			filter:    types.Tags{"env": "dev"},
			wantCount: 0,
			wantIDs:   []string{},
		},
		{
			name:      "empty filter returns all",
			filter:    types.Tags{},
			wantCount: 4,
			wantIDs:   []string{"backup-1", "backup-2", "backup-3", "backup-4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterBackupsByTags(backups, tt.filter)

			if len(filtered) != tt.wantCount {
				t.Errorf("FilterBackupsByTags() got %d backups, want %d", len(filtered), tt.wantCount)
			}

			// Check IDs
			gotIDs := make(map[string]bool)
			for _, b := range filtered {
				gotIDs[b.ID] = true
			}

			for _, wantID := range tt.wantIDs {
				if !gotIDs[wantID] {
					t.Errorf("FilterBackupsByTags() missing expected backup: %s", wantID)
				}
			}
		})
	}
}

func TestSortBackupsByTimestamp(t *testing.T) {
	now := time.Now()
	backups := []*types.BackupMetadata{
		{ID: "backup-2", Timestamp: now.Add(-2 * time.Hour)},
		{ID: "backup-4", Timestamp: now.Add(-4 * time.Hour)},
		{ID: "backup-1", Timestamp: now.Add(-1 * time.Hour)},
		{ID: "backup-3", Timestamp: now.Add(-3 * time.Hour)},
	}

	sorted := SortBackupsByTimestamp(backups)

	// Check order (newest first)
	expectedOrder := []string{"backup-1", "backup-2", "backup-3", "backup-4"}
	for i, expected := range expectedOrder {
		if sorted[i].ID != expected {
			t.Errorf("SortBackupsByTimestamp() position %d = %s, want %s", i, sorted[i].ID, expected)
		}
	}

	// Ensure original is not modified
	if backups[0].ID == "backup-1" {
		t.Error("SortBackupsByTimestamp() modified original slice")
	}
}

func TestGroupBackupsByTag(t *testing.T) {
	backups := []*types.BackupMetadata{
		{
			ID:   "backup-1",
			Tags: types.Tags{"env": "prod"},
		},
		{
			ID:   "backup-2",
			Tags: types.Tags{"env": "staging"},
		},
		{
			ID:   "backup-3",
			Tags: types.Tags{"env": "prod"},
		},
		{
			ID:   "backup-4",
			Tags: nil, // No tags
		},
		{
			ID:   "backup-5",
			Tags: types.Tags{"team": "platform"}, // Missing 'env' tag
		},
	}

	groups := GroupBackupsByTag(backups, "env")

	// Check prod group
	if len(groups["prod"]) != 2 {
		t.Errorf("GroupBackupsByTag() prod group = %d, want 2", len(groups["prod"]))
	}

	// Check staging group
	if len(groups["staging"]) != 1 {
		t.Errorf("GroupBackupsByTag() staging group = %d, want 1", len(groups["staging"]))
	}

	// Check <untagged> group (backup-4 has no tags, backup-5 is missing 'env' tag)
	if len(groups["<untagged>"]) != 2 {
		t.Errorf("GroupBackupsByTag() <untagged> group = %d, want 2", len(groups["<untagged>"]))
	}

	// Check total groups (prod, staging, <untagged>)
	if len(groups) != 3 {
		t.Errorf("GroupBackupsByTag() got %d groups, want 3", len(groups))
	}
}

func TestSearchInStateData(t *testing.T) {
	// Create a minimal search engine for testing
	cfg := &types.Config{}
	logger := utils.NewLogger(utils.LogLevelInfo)
	localStorage := storage.NewLocalStorage(types.LocalConfig{
		Enabled: true,
		Path:    ".test_snapshots",
	}, logger)
	engine := NewEngine(localStorage, cfg, logger)
	searchEngine := NewSearchEngine(engine)

	tests := []struct {
		name       string
		stateJSON  string
		searchTerm string
		want       bool
	}{
		{
			name: "find resource type",
			stateJSON: `{
				"version": 4,
				"resources": [
					{
						"type": "aws_instance",
						"name": "web",
						"provider": "provider.aws",
						"instances": []
					}
				]
			}`,
			searchTerm: "aws_instance",
			want:       true,
		},
		{
			name: "find resource name",
			stateJSON: `{
				"version": 4,
				"resources": [
					{
						"type": "aws_instance",
						"name": "web_server",
						"provider": "provider.aws",
						"instances": []
					}
				]
			}`,
			searchTerm: "web_server",
			want:       true,
		},
		{
			name: "no match",
			stateJSON: `{
				"version": 4,
				"resources": [
					{
						"type": "aws_instance",
						"name": "web",
						"provider": "provider.aws",
						"instances": []
					}
				]
			}`,
			searchTerm: "aws_s3_bucket",
			want:       false,
		},
		{
			name: "case insensitive search",
			stateJSON: `{
				"version": 4,
				"resources": [
					{
						"type": "AWS_Instance",
						"name": "web",
						"provider": "provider.aws",
						"instances": []
					}
				]
			}`,
			searchTerm: "aws_instance",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := searchEngine.searchInStateData([]byte(tt.stateJSON), tt.searchTerm)
			if got != tt.want {
				t.Errorf("searchInStateData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindLatestByTags(t *testing.T) {
	// This test would require mocking the storage layer
	// For now, we'll test the logic through integrated tests
	t.Skip("Requires mock storage - covered in integration tests")
}
