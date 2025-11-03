package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tf-safe/internal/backup"
	"tf-safe/internal/config"
	"tf-safe/internal/restore"
	"tf-safe/internal/storage"
	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore [backup-id]",
	Short: "Restore a previous Terraform state backup",
	Long: `Restore a previous Terraform state backup by specifying the backup ID.
	
Use 'tf-safe list' to see available backups and their IDs.
A backup of the current state will be created before restoration unless --no-backup is specified.

Examples:
  tf-safe restore terraform.tfstate.2025-10-28T11:50:27Z
  tf-safe restore terraform.tfstate.2025-10-28T11:50:27Z -t custom.tfstate
  tf-safe restore terraform.tfstate.2025-10-28T11:50:27Z --force
  tf-safe restore terraform.tfstate.2025-10-28T11:50:27Z --no-backup`,
	Args: cobra.ExactArgs(1),
	RunE: runRestoreCommand,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	
	// Add restore-specific flags
	restoreCmd.Flags().StringP("target", "t", "terraform.tfstate", "Target file path for restoration")
	restoreCmd.Flags().BoolP("force", "f", false, "Force restore without confirmation")
	restoreCmd.Flags().Bool("no-backup", false, "Skip creating backup before restore")
}

func runRestoreCommand(cmd *cobra.Command, args []string) error {
	backupID := args[0]
	
	// Get flags
	targetPath, err := cmd.Flags().GetString("target")
	if err != nil {
		return fmt.Errorf("failed to get target flag: %w", err)
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("failed to get force flag: %w", err)
	}
	noBackup, err := cmd.Flags().GetBool("no-backup")
	if err != nil {
		return fmt.Errorf("failed to get no-backup flag: %w", err)
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}
	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return fmt.Errorf("failed to get dry-run flag: %w", err)
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

	// Create restore engine
	restoreEngine := restore.NewEngine(localStorage, backupEngine, cfg, logger)

	// Validate backup exists and get metadata
	fmt.Print("Validating backup... ")
	if err := restoreEngine.ValidateBackup(ctx, backupID); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("backup validation failed: %w", err)
	}
	fmt.Println("OK")

	// Get backup metadata for display
	metadata, err := backupEngine.GetBackupMetadata(ctx, backupID)
	if err != nil {
		return fmt.Errorf("failed to get backup metadata: %w", err)
	}

	// Display backup information
	fmt.Printf("\nBackup Information:\n")
	fmt.Printf("  ID:        %s\n", metadata.ID)
	fmt.Printf("  Timestamp: %s\n", metadata.Timestamp.Format(time.RFC3339))
	fmt.Printf("  Size:      %d bytes\n", metadata.Size)
	fmt.Printf("  Checksum:  %s\n", metadata.Checksum)
	fmt.Printf("  Storage:   %s\n", metadata.StorageType)

	// Check if target file exists and warn user
	targetExists := utils.FileExists(targetPath)
	if targetExists {
		fmt.Printf("\nWarning: Target file '%s' exists and will be overwritten.\n", targetPath)
		if !noBackup {
			fmt.Printf("A backup will be created before restoration.\n")
		} else {
			fmt.Printf("No backup will be created (--no-backup specified).\n")
		}
	}

	// Confirmation prompt unless force is specified
	if !force && !dryRun {
		fmt.Printf("\nDo you want to proceed with the restore? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Restore cancelled.")
			return nil
		}
	}

	// Create restore options
	opts := types.RestoreOptions{
		BackupID:     backupID,
		TargetPath:   targetPath,
		CreateBackup: !noBackup && targetExists,
		Force:        force,
	}

	if dryRun {
		logger.Info("DRY RUN: Would restore backup with options: %+v", opts)
		return nil
	}

	// Perform restore
	fmt.Print("Restoring backup... ")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := restoreEngine.RestoreBackup(ctx, opts); err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("restore operation failed: %w", err)
	}

	fmt.Println("SUCCESS")

	fmt.Printf("\nRestore completed successfully:\n")
	fmt.Printf("  Backup ID: %s\n", backupID)
	fmt.Printf("  Target:    %s\n", targetPath)
	fmt.Printf("  Size:      %d bytes\n", metadata.Size)

	return nil
}