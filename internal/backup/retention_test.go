package backup

import (
	"context"
	"testing"
	"time"

	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

func TestRetentionManager_ApplyLocalRetentionPolicy(t *testing.T) {
	config := types.RetentionConfig{
		LocalCount:  4, // Must be > MinimumRetentionCount (3) to trigger count-based retention
		RemoteCount: 10,
		MaxAgeDays:  30,
	}
	logger := utils.NewLogger(utils.LogLevelInfo)
	manager := NewRetentionManager(config, logger)

	ctx := context.Background()
	now := time.Now().UTC()

	// Create test backups (6 backups, should keep 4 newest)
	backups := []*types.BackupMetadata{
		{
			ID:        "backup-1",
			Timestamp: now.Add(-6 * time.Hour), // oldest
			Size:      100,
		},
		{
			ID:        "backup-2",
			Timestamp: now.Add(-5 * time.Hour),
			Size:      100,
		},
		{
			ID:        "backup-3",
			Timestamp: now.Add(-4 * time.Hour),
			Size:      100,
		},
		{
			ID:        "backup-4",
			Timestamp: now.Add(-3 * time.Hour),
			Size:      100,
		},
		{
			ID:        "backup-5",
			Timestamp: now.Add(-2 * time.Hour),
			Size:      100,
		},
		{
			ID:        "backup-6",
			Timestamp: now.Add(-1 * time.Hour), // newest
			Size:      100,
		},
	}

	toDelete, err := manager.ApplyLocalRetentionPolicy(ctx, backups)
	if err != nil {
		t.Fatalf("Failed to apply local retention policy: %v", err)
	}

	// Should delete 2 oldest backups (backup-1 and backup-2)
	if len(toDelete) != 2 {
		t.Errorf("Expected 2 backups to delete, got %d", len(toDelete))
	}

	// Verify the oldest backups are marked for deletion
	expectedToDelete := map[string]bool{
		"backup-1": true,
		"backup-2": true,
	}

	for _, backup := range toDelete {
		if !expectedToDelete[backup.ID] {
			t.Errorf("Unexpected backup marked for deletion: %s", backup.ID)
		}
	}
}

func TestRetentionManager_ApplyLocalRetentionPolicy_MinimumRetention(t *testing.T) {
	// Test with retention count less than minimum (should keep at least 3)
	config := types.RetentionConfig{
		LocalCount: 1, // Less than minimum
	}
	logger := utils.NewLogger(utils.LogLevelInfo)
	manager := NewRetentionManager(config, logger)

	ctx := context.Background()
	now := time.Now().UTC()

	// Create only 3 backups (at minimum threshold)
	backups := []*types.BackupMetadata{
		{ID: "backup-1", Timestamp: now.Add(-3 * time.Hour)},
		{ID: "backup-2", Timestamp: now.Add(-2 * time.Hour)},
		{ID: "backup-3", Timestamp: now.Add(-1 * time.Hour)},
	}

	toDelete, err := manager.ApplyLocalRetentionPolicy(ctx, backups)
	if err != nil {
		t.Fatalf("Failed to apply local retention policy: %v", err)
	}

	// Should not delete any backups (at minimum retention)
	if len(toDelete) != 0 {
		t.Errorf("Expected 0 backups to delete (minimum retention), got %d", len(toDelete))
	}
}

func TestRetentionManager_ApplyRemoteRetentionPolicy(t *testing.T) {
	config := types.RetentionConfig{
		LocalCount:  5,
		RemoteCount: 4, // Must be > MinimumRetentionCount (3) to trigger count-based retention
		MaxAgeDays:  30,
	}
	logger := utils.NewLogger(utils.LogLevelInfo)
	manager := NewRetentionManager(config, logger)

	ctx := context.Background()
	now := time.Now().UTC()

	// Create test backups for remote storage (6 backups, should keep 4 newest)
	backups := []*types.BackupMetadata{
		{
			ID:        "remote-backup-1",
			Timestamp: now.Add(-6 * time.Hour),
			Size:      200,
		},
		{
			ID:        "remote-backup-2",
			Timestamp: now.Add(-5 * time.Hour),
			Size:      200,
		},
		{
			ID:        "remote-backup-3",
			Timestamp: now.Add(-4 * time.Hour),
			Size:      200,
		},
		{
			ID:        "remote-backup-4",
			Timestamp: now.Add(-3 * time.Hour),
			Size:      200,
		},
		{
			ID:        "remote-backup-5",
			Timestamp: now.Add(-2 * time.Hour),
			Size:      200,
		},
		{
			ID:        "remote-backup-6",
			Timestamp: now.Add(-1 * time.Hour),
			Size:      200,
		},
	}

	toDelete, err := manager.ApplyRemoteRetentionPolicy(ctx, backups)
	if err != nil {
		t.Fatalf("Failed to apply remote retention policy: %v", err)
	}

	// Should delete 2 oldest backups (keep 4 newest)
	if len(toDelete) != 2 {
		t.Errorf("Expected 2 backups to delete, got %d", len(toDelete))
	}

	// Verify the oldest backups are marked for deletion
	expectedToDelete := map[string]bool{
		"remote-backup-1": true,
		"remote-backup-2": true,
	}

	for _, backup := range toDelete {
		if !expectedToDelete[backup.ID] {
			t.Errorf("Unexpected backup marked for deletion: %s", backup.ID)
		}
	}
}

func TestRetentionManager_ApplyAgeBasedRetention(t *testing.T) {
	config := types.RetentionConfig{
		LocalCount:  10, // High count, so age should be the limiting factor
		RemoteCount: 10,
		MaxAgeDays:  7, // 7 days
	}
	logger := utils.NewLogger(utils.LogLevelInfo)
	manager := NewRetentionManager(config, logger)

	ctx := context.Background()
	now := time.Now().UTC()

	// Create backups with different ages
	backups := []*types.BackupMetadata{
		{
			ID:        "old-backup-1",
			Timestamp: now.Add(-10 * 24 * time.Hour), // 10 days old (should delete)
			Size:      100,
		},
		{
			ID:        "old-backup-2",
			Timestamp: now.Add(-8 * 24 * time.Hour), // 8 days old (should delete)
			Size:      100,
		},
		{
			ID:        "recent-backup-1",
			Timestamp: now.Add(-5 * 24 * time.Hour), // 5 days old (should keep)
			Size:      100,
		},
		{
			ID:        "recent-backup-2",
			Timestamp: now.Add(-2 * 24 * time.Hour), // 2 days old (should keep)
			Size:      100,
		},
		{
			ID:        "recent-backup-3",
			Timestamp: now.Add(-1 * time.Hour), // 1 hour old (should keep)
			Size:      100,
		},
	}

	toDelete, err := manager.ApplyLocalRetentionPolicy(ctx, backups)
	if err != nil {
		t.Fatalf("Failed to apply age-based retention policy: %v", err)
	}

	// Should delete 2 old backups (older than 7 days)
	if len(toDelete) != 2 {
		t.Errorf("Expected 2 old backups to delete, got %d", len(toDelete))
	}

	// Verify the old backups are marked for deletion
	expectedToDelete := map[string]bool{
		"old-backup-1": true,
		"old-backup-2": true,
	}

	for _, backup := range toDelete {
		if !expectedToDelete[backup.ID] {
			t.Errorf("Unexpected backup marked for deletion: %s", backup.ID)
		}
	}
}

func TestRetentionManager_NoBackupsToDelete(t *testing.T) {
	config := types.RetentionConfig{
		LocalCount:  10,
		RemoteCount: 10,
		MaxAgeDays:  30,
	}
	logger := utils.NewLogger(utils.LogLevelInfo)
	manager := NewRetentionManager(config, logger)

	ctx := context.Background()
	now := time.Now().UTC()

	// Create only 2 recent backups (less than retention count and within age limit)
	backups := []*types.BackupMetadata{
		{
			ID:        "recent-backup-1",
			Timestamp: now.Add(-2 * time.Hour),
			Size:      100,
		},
		{
			ID:        "recent-backup-2",
			Timestamp: now.Add(-1 * time.Hour),
			Size:      100,
		},
	}

	toDelete, err := manager.ApplyLocalRetentionPolicy(ctx, backups)
	if err != nil {
		t.Fatalf("Failed to apply retention policy: %v", err)
	}

	// Should not delete any backups
	if len(toDelete) != 0 {
		t.Errorf("Expected 0 backups to delete, got %d", len(toDelete))
	}
}

func TestRetentionManager_EmptyBackupList(t *testing.T) {
	config := types.RetentionConfig{
		LocalCount:  5,
		RemoteCount: 10,
		MaxAgeDays:  30,
	}
	logger := utils.NewLogger(utils.LogLevelInfo)
	manager := NewRetentionManager(config, logger)

	ctx := context.Background()

	// Test with empty backup list
	toDelete, err := manager.ApplyLocalRetentionPolicy(ctx, []*types.BackupMetadata{})
	if err != nil {
		t.Fatalf("Failed to apply retention policy to empty list: %v", err)
	}

	if len(toDelete) != 0 {
		t.Errorf("Expected 0 backups to delete from empty list, got %d", len(toDelete))
	}
}

func TestRetentionManager_SortBackupsByTimestamp(t *testing.T) {
	config := types.RetentionConfig{
		LocalCount: 4, // Must be > MinimumRetentionCount (3) to trigger count-based retention
	}
	logger := utils.NewLogger(utils.LogLevelInfo)
	manager := NewRetentionManager(config, logger)

	ctx := context.Background()
	now := time.Now().UTC()

	// Create 5 backups in random order
	backups := []*types.BackupMetadata{
		{
			ID:        "backup-middle",
			Timestamp: now.Add(-3 * time.Hour),
			Size:      100,
		},
		{
			ID:        "backup-newest",
			Timestamp: now.Add(-1 * time.Hour),
			Size:      100,
		},
		{
			ID:        "backup-oldest",
			Timestamp: now.Add(-5 * time.Hour),
			Size:      100,
		},
		{
			ID:        "backup-second-newest",
			Timestamp: now.Add(-2 * time.Hour),
			Size:      100,
		},
		{
			ID:        "backup-second-oldest",
			Timestamp: now.Add(-4 * time.Hour),
			Size:      100,
		},
	}

	toDelete, err := manager.ApplyLocalRetentionPolicy(ctx, backups)
	if err != nil {
		t.Fatalf("Failed to apply retention policy: %v", err)
	}

	// Should delete 1 backup (the oldest one)
	if len(toDelete) != 1 {
		t.Errorf("Expected 1 backup to delete, got %d", len(toDelete))
	}

	if len(toDelete) > 0 && toDelete[0].ID != "backup-oldest" {
		t.Errorf("Expected oldest backup to be deleted, got %s", toDelete[0].ID)
	}
}