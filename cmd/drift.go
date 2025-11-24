package cmd

import (
	"context"
	"fmt"
	"os"

	"tf-safe/internal/audit"
	"tf-safe/internal/backup"
	"tf-safe/internal/config"
	"tf-safe/internal/notifications"
	"tf-safe/internal/storage"
	"tf-safe/internal/terraform"
	"tf-safe/internal/utils"
	"tf-safe/pkg/types"

	"github.com/spf13/cobra"
)

// driftDetectCmd represents the drift-detect command
var driftDetectCmd = &cobra.Command{
	Use:   "drift-detect",
	Short: "Detect configuration drift from last backup",
	Long: `Detect configuration drift by comparing current Terraform state with the most recent backup.

This command helps identify unexpected changes in your infrastructure by comparing
the current state file with the last backup. It's useful for detecting manual changes
or configuration drift in production environments.

Examples:
  tf-safe drift-detect                           # Compare with latest backup
  tf-safe drift-detect --tags "env=prod"        # Compare with latest prod backup
  tf-safe drift-detect --format json            # Output as JSON
  tf-safe drift-detect --backup-id <id>         # Compare with specific backup`,
	RunE: runDriftDetectCommand,
}

func init() {
	rootCmd.AddCommand(driftDetectCmd)

	// Add drift-detect-specific flags
	driftDetectCmd.Flags().StringP("format", "f", "text", "Output format (text, json, yaml)")
	driftDetectCmd.Flags().Bool("no-color", false, "Disable color output")
	driftDetectCmd.Flags().String("tags", "", "Filter baseline backup by tags")
	driftDetectCmd.Flags().String("backup-id", "", "Compare with specific backup ID (instead of latest)")
	driftDetectCmd.Flags().String("state-file", "terraform.tfstate", "Path to current state file")
}

func runDriftDetectCommand(cmd *cobra.Command, args []string) error {
	// Get flags
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return fmt.Errorf("failed to get format flag: %w", err)
	}
	noColor, err := cmd.Flags().GetBool("no-color")
	if err != nil {
		return fmt.Errorf("failed to get no-color flag: %w", err)
	}
	tagsStr, err := cmd.Flags().GetString("tags")
	if err != nil {
		return fmt.Errorf("failed to get tags flag: %w", err)
	}
	backupID, err := cmd.Flags().GetString("backup-id")
	if err != nil {
		return fmt.Errorf("failed to get backup-id flag: %w", err)
	}
	stateFile, err := cmd.Flags().GetString("state-file")
	if err != nil {
		return fmt.Errorf("failed to get state-file flag: %w", err)
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Validate format
	validFormats := []string{"text", "json", "yaml"}
	if !contains(validFormats, format) {
		return fmt.Errorf("invalid format '%s'. Valid formats: text, json, yaml", format)
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

	// Create storage backend
	localStorage := storage.NewLocalStorage(cfg.Local, logger)

	// Initialize storage backend
	ctx := context.Background()
	if err := localStorage.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize local storage: %w", err)
	}

	// Create backup engine
	backupEngine := backup.NewEngine(localStorage, cfg, logger)

	// Read current state file
	currentData, err := os.ReadFile(stateFile)
	if err != nil {
		return fmt.Errorf("failed to read current state file '%s': %w", stateFile, err)
	}

	// Determine baseline backup
	var baselineData []byte
	var baselineID string

	if backupID != "" {
		// Use specified backup
		baselineID = backupID
		baselineData, _, err = localStorage.Retrieve(ctx, backupID)
		if err != nil {
			return fmt.Errorf("failed to retrieve backup %s: %w", backupID, err)
		}
	} else {
		// Find latest backup (optionally filtered by tags)
		searchEngine := backup.NewSearchEngine(backupEngine)

		var tagFilter types.Tags
		if tagsStr != "" {
			tagFilter, err = types.ParseTagString(tagsStr)
			if err != nil {
				return fmt.Errorf("invalid tags format: %w", err)
			}
		}

		metadata, err := searchEngine.FindLatestByTags(ctx, tagFilter)
		if err != nil {
			return fmt.Errorf("failed to find baseline backup: %w", err)
		}

		baselineID = metadata.ID
		baselineData, _, err = localStorage.Retrieve(ctx, baselineID)
		if err != nil {
			return fmt.Errorf("failed to retrieve baseline backup: %w", err)
		}
	}

	// Parse states
	baselineState, err := terraform.ParseState(baselineData)
	if err != nil {
		return fmt.Errorf("failed to parse baseline state: %w", err)
	}

	currentState, err := terraform.ParseState(currentData)
	if err != nil {
		return fmt.Errorf("failed to parse current state: %w", err)
	}

	// Compare states
	diff := terraform.CompareStates(baselineState, currentState)

	// Display comparison info
	fmt.Printf("Comparing:\n")
	fmt.Printf("  Baseline: %s\n", baselineID)
	fmt.Printf("  Current:  %s\n\n", stateFile)

	// Format and display output
	formatter := terraform.NewDiffFormatter(diff)

	switch format {
	case "json":
		output, err := formatter.FormatJSON(true)
		if err != nil {
			return err
		}
		fmt.Println(output)

	case "yaml":
		output, err := formatter.FormatYAML()
		if err != nil {
			return err
		}
		fmt.Print(output)

	default: // text
		colorize := !noColor
		fmt.Print(formatter.FormatText(colorize))
	}

	// Exit with code 1 if drift detected (useful for CI/CD)
	if diff.HasDrift() {
		fmt.Printf("\n⚠️  Configuration drift detected!\n")

		// Send notification
		notifier := notifications.NewManager(cfg.Notifications, logger)
		notifier.NotifyDriftDetected(baselineID, diff.FormatSummary())

		// Log audit entry
		auditLogger := audit.NewLogger(cfg.Audit)
		auditLogger.Log(types.AuditActionDriftCheck, types.AuditStatusFailed, map[string]interface{}{
			"baseline": baselineID,
			"current":  stateFile,
			"drift":    true,
			"summary":  diff.FormatSummary(),
		}, nil)

		os.Exit(1)
	} else {
		fmt.Printf("\n✓ No configuration drift detected\n")

		// Log audit entry
		auditLogger := audit.NewLogger(cfg.Audit)
		auditLogger.Log(types.AuditActionDriftCheck, types.AuditStatusSuccess, map[string]interface{}{
			"baseline": baselineID,
			"current":  stateFile,
			"drift":    false,
		}, nil)
	}

	return nil
}
