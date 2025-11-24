package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"tf-safe/pkg/types"
)

// SearchOptions contains options for searching backups
type SearchOptions struct {
	Tags          types.Tags // Filter by tags (AND logic - all must match)
	SearchContent string     // Search for content in state files
	Limit         int        // Limit number of results
}

// SearchEngine provides backup search capabilities
type SearchEngine struct {
	engine *Engine
}

// NewSearchEngine creates a new search engine
func NewSearchEngine(engine *Engine) *SearchEngine {
	return &SearchEngine{
		engine: engine,
	}
}

// Search searches backups based on provided options
func (s *SearchEngine) Search(ctx context.Context, opts SearchOptions) ([]*types.BackupMetadata, error) {
	// Get all backups
	allBackups, err := s.engine.ListBackups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	var results []*types.BackupMetadata

	// Apply filters
	for _, backup := range allBackups {
		if s.matchesFilters(ctx, backup, opts) {
			results = append(results, backup)

			// Apply limit
			if opts.Limit > 0 && len(results) >= opts.Limit {
				break
			}
		}
	}

	return results, nil
}

// matchesFilters checks if a backup matches the search criteria
func (s *SearchEngine) matchesFilters(ctx context.Context, backup *types.BackupMetadata, opts SearchOptions) bool {
	// Filter by tags
	if len(opts.Tags) > 0 {
		if backup.Tags == nil || !backup.Tags.Matches(opts.Tags) {
			return false
		}
	}

	// Filter by content search
	if opts.SearchContent != "" {
		if !s.matchesContent(ctx, backup, opts.SearchContent) {
			return false
		}
	}

	return true
}

// matchesContent checks if backup content contains the search string
func (s *SearchEngine) matchesContent(ctx context.Context, backup *types.BackupMetadata, searchStr string) bool {
	// Try to retrieve backup data
	var data []byte
	var err error

	// Try local storage first
	if s.engine.localStorage != nil {
		data, _, err = s.engine.localStorage.Retrieve(ctx, backup.ID)
		if err == nil {
			return s.searchInStateData(data, searchStr)
		}
	}

	// Try remote storage if local failed
	if s.engine.remoteStorage != nil {
		data, _, err = s.engine.remoteStorage.Retrieve(ctx, backup.ID)
		if err == nil {
			return s.searchInStateData(data, searchStr)
		}
	}

	// If we can't retrieve the data, don't match
	return false
}

// searchInStateData searches within Terraform state file data
func (s *SearchEngine) searchInStateData(data []byte, searchStr string) bool {
	searchLower := strings.ToLower(searchStr)

	// Try to parse as JSON state file
	var state TerraformState
	if err := json.Unmarshal(data, &state); err != nil {
		// If not valid JSON, do simple string search
		return strings.Contains(strings.ToLower(string(data)), searchLower)
	}

	// Search in resources
	for _, resource := range state.Resources {
		// Search in resource type
		if strings.Contains(strings.ToLower(resource.Type), searchLower) {
			return true
		}

		// Search in resource name
		if strings.Contains(strings.ToLower(resource.Name), searchLower) {
			return true
		}

		// Search in module
		if strings.Contains(strings.ToLower(resource.Module), searchLower) {
			return true
		}

		// Search in provider
		if strings.Contains(strings.ToLower(resource.Provider), searchLower) {
			return true
		}

		// Search in instances (convert to string and search)
		for _, instance := range resource.Instances {
			instanceJSON, _ := json.Marshal(instance)
			if strings.Contains(strings.ToLower(string(instanceJSON)), searchLower) {
				return true
			}
		}
	}

	return false
}

// FindLatestByTags finds the most recent backup matching the given tags
func (s *SearchEngine) FindLatestByTags(ctx context.Context, tags types.Tags) (*types.BackupMetadata, error) {
	results, err := s.Search(ctx, SearchOptions{
		Tags:  tags,
		Limit: 1, // We only need the latest
	})
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no backups found matching tags: %s", tags.String())
	}

	// Since ListBackups returns backups in reverse chronological order,
	// the first result is the latest
	return results[0], nil
}

// TerraformState represents a simplified Terraform state structure for searching
type TerraformState struct {
	Version   int                 `json:"version"`
	Resources []TerraformResource `json:"resources"`
}

// TerraformResource represents a resource in Terraform state
type TerraformResource struct {
	Type      string                   `json:"type"`
	Name      string                   `json:"name"`
	Provider  string                   `json:"provider"`
	Module    string                   `json:"module,omitempty"`
	Instances []map[string]interface{} `json:"instances"`
}

// FilterBackupsByTags filters a list of backups by tags
func FilterBackupsByTags(backups []*types.BackupMetadata, tags types.Tags) []*types.BackupMetadata {
	if len(tags) == 0 {
		return backups
	}

	var filtered []*types.BackupMetadata
	for _, backup := range backups {
		if backup.Tags != nil && backup.Tags.Matches(tags) {
			filtered = append(filtered, backup)
		}
	}
	return filtered
}

// SortBackupsByTimestamp sorts backups by timestamp (newest first)
func SortBackupsByTimestamp(backups []*types.BackupMetadata) []*types.BackupMetadata {
	// Create a copy to avoid modifying the original
	sorted := make([]*types.BackupMetadata, len(backups))
	copy(sorted, backups)

	// Simple bubble sort (sufficient for typical backup counts)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Timestamp.Before(sorted[j].Timestamp) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// GroupBackupsByTag groups backups by a specific tag key
func GroupBackupsByTag(backups []*types.BackupMetadata, tagKey string) map[string][]*types.BackupMetadata {
	groups := make(map[string][]*types.BackupMetadata)

	for _, backup := range backups {
		tagValue := "<untagged>"

		if backup.Tags != nil {
			if val, exists := backup.Tags[tagKey]; exists {
				tagValue = val
			}
		}

		groups[tagValue] = append(groups[tagValue], backup)
	}

	return groups
}
