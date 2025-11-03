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

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan [terraform-args...]",
	Short: "Terraform plan wrapper with automatic backups",
	Long: `Execute 'terraform plan' with automatic backup hooks.
	
This command creates a backup before running terraform plan and executes
the plan operation with all provided arguments.

All terraform plan arguments and flags are passed through unchanged.`,
	DisableFlagParsing: true, // Allow passing all args to terraform
	Run: func(cmd *cobra.Command, args []string) {
		if err := runPlanCommand(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runPlanCommand(args []string) error {
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
	if verbose, err := rootCmd.PersistentFlags().GetBool("verbose"); err == nil && verbose {
		loggingHook := terraform.NewLoggingHook(true)
		wrapper.AddHook(loggingHook)
	}

	// Execute terraform plan with backup hooks
	return wrapper.ExecuteWithBackup(ctx, "plan", args)
}

func init() {
	rootCmd.AddCommand(planCmd)
}