package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"tf-safe/internal/backup"
	"tf-safe/internal/config"
	"tf-safe/internal/storage"
	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available backup versions",
	Long: `List all available backup versions with their timestamps, sizes, and storage locations.
	
This command shows both local and remote backups in chronological order,
along with metadata like file size, storage backend, and encryption status.

Examples:
  tf-safe list                    # List all backups in table format
  tf-safe list -f json           # List backups in JSON format
  tf-safe list -s local          # List only local backups
  tf-safe list --limit 10        # List only the 10 most recent backups`,
	RunE: runListCommand,
}

func init() {
	rootCmd.AddCommand(listCmd)
	
	// Add list-specific flags
	listCmd.Flags().StringP("format", "f", "table", "Output format (table, json, yaml)")
	listCmd.Flags().StringP("storage", "s", "all", "Filter by storage backend (local, remote, all)")
	listCmd.Flags().Int("limit", 0, "Limit number of results (0 = no limit)")
}

func runListCommand(cmd *cobra.Command, args []string) error {
	// Get flags
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return fmt.Errorf("failed to get format flag: %w", err)
	}
	storageFilter, err := cmd.Flags().GetString("storage")
	if err != nil {
		return fmt.Errorf("failed to get storage flag: %w", err)
	}
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return fmt.Errorf("failed to get limit flag: %w", err)
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Validate format
	validFormats := []string{"table", "json", "yaml"}
	if !contains(validFormats, format) {
		return fmt.Errorf("invalid format '%s'. Valid formats: %s", format, strings.Join(validFormats, ", "))
	}

	// Validate storage filter
	validStorageFilters := []string{"all", "local", "remote"}
	if !contains(validStorageFilters, storageFilter) {
		return fmt.Errorf("invalid storage filter '%s'. Valid filters: %s", storageFilter, strings.Join(validStorageFilters, ", "))
	}

	// Initialize logger
	logLevel := utils.LogLevelInfo
	if verbose {
		logLevel = utils.LogLevelDebug
	}
	logger := utils.NewLogger(logLevel)

	// Load configuration
	cfg, err := config.LoadConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate that local storage is enabled
	if !cfg.Local.Enabled {
		return fmt.Errorf("local storage is disabled in configuration")
	}

	// Create storage backend
	localStorage := storage.NewLocalStorage(cfg.Local, logger)
	
	// Initialize storage backend
	ctx := context.Background()
	if err := localStorage.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize local storage: %w", err)
	}

	// Create backup engine
	backupEngine := backup.NewEngine(localStorage, cfg, logger)

	// List backups
	backups, err := backupEngine.ListBackups(ctx)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	// Filter by storage type
	if storageFilter != "all" {
		filteredBackups := make([]*types.BackupMetadata, 0)
		for _, backup := range backups {
			if storageFilter == "local" && backup.StorageType == "local" {
				filteredBackups = append(filteredBackups, backup)
			} else if storageFilter == "remote" && backup.StorageType != "local" {
				filteredBackups = append(filteredBackups, backup)
			}
		}
		backups = filteredBackups
	}

	// Apply limit
	if limit > 0 && len(backups) > limit {
		backups = backups[:limit]
	}

	// Display results
	switch format {
	case "json":
		return displayJSON(backups)
	case "yaml":
		return displayYAML(backups)
	default:
		return displayTable(backups)
	}
}

func displayTable(backups []*types.BackupMetadata) error {
	if len(backups) == 0 {
		fmt.Println("No backups found.")
		return nil
	}

	// Print header
	fmt.Printf("%-35s %-20s %-10s %-10s %-10s %-10s\n", 
		"BACKUP ID", "TIMESTAMP", "SIZE", "STORAGE", "ENCRYPTED", "CHECKSUM")
	fmt.Printf("%-35s %-20s %-10s %-10s %-10s %-10s\n", 
		strings.Repeat("-", 35), strings.Repeat("-", 20), strings.Repeat("-", 10), 
		strings.Repeat("-", 10), strings.Repeat("-", 10), strings.Repeat("-", 10))

	// Print backup rows
	for _, backup := range backups {
		encrypted := "No"
		if backup.Encrypted {
			encrypted = "Yes"
		}

		// Format size
		sizeStr := formatSize(backup.Size)
		
		// Format timestamp
		timestampStr := backup.Timestamp.Format("2006-01-02 15:04:05")
		
		// Truncate checksum for display
		checksumStr := backup.Checksum
		if len(checksumStr) > 10 {
			checksumStr = checksumStr[:8] + ".."
		}

		fmt.Printf("%-35s %-20s %-10s %-10s %-10s %-10s\n",
			backup.ID, timestampStr, sizeStr, backup.StorageType, encrypted, checksumStr)
	}

	fmt.Printf("\nTotal: %d backup(s)\n", len(backups))
	return nil
}

func displayJSON(backups []*types.BackupMetadata) error {
	output := map[string]interface{}{
		"backups": backups,
		"total":   len(backups),
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	
	fmt.Println(string(data))
	return nil
}

func displayYAML(backups []*types.BackupMetadata) error {
	output := map[string]interface{}{
		"backups": backups,
		"total":   len(backups),
	}

	data, err := yaml.Marshal(output)
	if err != nil {
		return err
	}
	
	fmt.Print(string(data))
	return nil
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}