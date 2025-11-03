package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tf-safe/internal/backup"
	"tf-safe/internal/config"
	"tf-safe/internal/storage"
	"tf-safe/internal/terraform"
	"tf-safe/internal/utils"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply [terraform-args...]",
	Short: "Terraform apply wrapper with automatic backups",
	Long: `Execute 'terraform apply' with automatic pre and post-operation backups.
	
This command creates a backup before running terraform apply, executes the
apply operation with all provided arguments, and creates another backup
after successful completion.

All terraform apply arguments and flags are passed through unchanged.`,
	DisableFlagParsing: true, // Allow passing all args to terraform
	Run: func(cmd *cobra.Command, args []string) {
		if err := runApplyCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runApplyCommand(args []string) error {
	ctx := context.Background()

	// Initialize configuration manager
	configManager := config.NewManager()
	
	// Load configuration
	cfg, err := configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	logger := utils.NewLogger(utils.ParseLogLevel("info"))

	// Initialize storage backend
	storageBackend := storage.NewLocalStorage(cfg.Local, logger)
	
	// Initialize storage backend
	if err := storageBackend.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize backup engine
	backupEngine := backup.NewEngine(storageBackend, cfg, logger)

	// Initialize Terraform wrapper
	wrapper := terraform.NewWrapper(configManager, backupEngine)

	// Add backup hook
	backupHook := terraform.NewBackupHook(configManager, backupEngine)
	wrapper.AddHook(backupHook)

	// Add logging hook if verbose mode is enabled
	if verbose, _ := rootCmd.PersistentFlags().GetBool("verbose"); verbose {
		loggingHook := terraform.NewLoggingHook(true)
		wrapper.AddHook(loggingHook)
	}

	// Execute terraform apply with backup hooks
	return wrapper.ExecuteWithBackup(ctx, "apply", args)
}

func init() {
	rootCmd.AddCommand(applyCmd)
}