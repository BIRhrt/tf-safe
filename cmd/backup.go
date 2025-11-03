package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"tf-safe/internal/backup"
	"tf-safe/internal/config"
	"tf-safe/internal/storage"
	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup [state-file]",
	Short: "Create a manual backup of the Terraform state file",
	Long: `Create a manual backup of the current Terraform state file.
	
This command will detect the terraform.tfstate file in the current directory
and create a timestamped backup according to your configuration settings.

Examples:
  tf-safe backup                    # Backup detected state file
  tf-safe backup terraform.tfstate # Backup specific state file
  tf-safe backup -f                 # Force backup even if no state file exists
  tf-safe backup -d "Pre-deploy"   # Add description to backup`,
	RunE: runBackupCommand,
}

func init() {
	rootCmd.AddCommand(backupCmd)
	
	// Add backup-specific flags
	backupCmd.Flags().StringP("description", "d", "", "Description for the backup")
	backupCmd.Flags().BoolP("force", "f", false, "Force backup even if no state file exists")
}

func runBackupCommand(cmd *cobra.Command, args []string) error {
	// Get flags
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return fmt.Errorf("failed to get description flag: %w", err)
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("failed to get force flag: %w", err)
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

	// Determine state file path
	var stateFilePath string
	if len(args) > 0 {
		stateFilePath = args[0]
	}

	// Create backup options
	opts := types.BackupOptions{
		StateFilePath: stateFilePath,
		Description:   description,
		Force:         force,
	}

	if dryRun {
		logger.Info("DRY RUN: Would create backup with options: %+v", opts)
		return nil
	}

	// Show progress indicator
	fmt.Print("Creating backup... ")

	// Create backup
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	metadata, err := backupEngine.CreateBackup(ctx, opts)
	if err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("backup creation failed: %w", err)
	}

	fmt.Println("SUCCESS")

	// Display backup information
	fmt.Printf("\nBackup created successfully:\n")
	fmt.Printf("  ID:        %s\n", metadata.ID)
	fmt.Printf("  Timestamp: %s\n", metadata.Timestamp.Format(time.RFC3339))
	fmt.Printf("  Size:      %d bytes\n", metadata.Size)
	fmt.Printf("  Checksum:  %s\n", metadata.Checksum)
	fmt.Printf("  Storage:   %s\n", metadata.StorageType)
	if metadata.Encrypted {
		fmt.Printf("  Encrypted: Yes\n")
	}
	if description != "" {
		fmt.Printf("  Description: %s\n", description)
	}

	// Apply retention policies
	if !dryRun {
		fmt.Print("Applying retention policies... ")
		if err := backupEngine.CleanupOldBackups(ctx); err != nil {
			fmt.Println("WARNING")
			logger.Warn("Failed to apply retention policies: %v", err)
		} else {
			fmt.Println("DONE")
		}
	}

	return nil
}