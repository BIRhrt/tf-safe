package cmd

import (
	"context"
	"fmt"
	"os"

	"tf-safe/internal/config"
	"tf-safe/internal/storage"
	"tf-safe/internal/terraform"
	"tf-safe/internal/utils"

	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff <backup-id-1> <backup-id-2>",
	Short: "Compare two Terraform state backups",
	Long: `Compare two Terraform state backups and show the differences.

This command parses the Terraform state files from two backups and displays
resource-level changes including added, modified, and removed resources.

Examples:
  tf-safe diff backup-1 backup-2                    # Compare two backups
  tf-safe diff backup-1 backup-2 --format json     # Output as JSON
  tf-safe diff backup-1 backup-2 --no-color        # Disable color output
  tf-safe diff backup-1 current                     # Compare backup with current state`,
	Args: cobra.ExactArgs(2),
	RunE: runDiffCommand,
}

func init() {
	rootCmd.AddCommand(diffCmd)

	// Add diff-specific flags
	diffCmd.Flags().StringP("format", "f", "text", "Output format (text, json, yaml)")
	diffCmd.Flags().Bool("no-color", false, "Disable color output")
	diffCmd.Flags().Bool("compact", false, "Show only summary without details")
}

func runDiffCommand(cmd *cobra.Command, args []string) error {
	backupID1 := args[0]
	backupID2 := args[1]

	// Get flags
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return fmt.Errorf("failed to get format flag: %w", err)
	}
	noColor, err := cmd.Flags().GetBool("no-color")
	if err != nil {
		return fmt.Errorf("failed to get no-color flag: %w", err)
	}
	compact, err := cmd.Flags().GetBool("compact")
	if err != nil {
		return fmt.Errorf("failed to get compact flag: %w", err)
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

	// Retrieve first backup
	var data1 []byte
	if backupID1 == "current" {
		// Read current state file
		data1, err = os.ReadFile("terraform.tfstate")
		if err != nil {
			return fmt.Errorf("failed to read current state file: %w", err)
		}
	} else {
		data1, _, err = localStorage.Retrieve(ctx, backupID1)
		if err != nil {
			return fmt.Errorf("failed to retrieve backup %s: %w", backupID1, err)
		}
	}

	// Retrieve second backup
	var data2 []byte
	if backupID2 == "current" {
		data2, err = os.ReadFile("terraform.tfstate")
		if err != nil {
			return fmt.Errorf("failed to read current state file: %w", err)
		}
	} else {
		data2, _, err = localStorage.Retrieve(ctx, backupID2)
		if err != nil {
			return fmt.Errorf("failed to retrieve backup %s: %w", backupID2, err)
		}
	}

	// Parse states
	state1, err := terraform.ParseState(data1)
	if err != nil {
		return fmt.Errorf("failed to parse state from %s: %w", backupID1, err)
	}

	state2, err := terraform.ParseState(data2)
	if err != nil {
		return fmt.Errorf("failed to parse state from %s: %w", backupID2, err)
	}

	// Compare states
	diff := terraform.CompareStates(state1, state2)

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
		if compact {
			fmt.Println(formatter.FormatCompact())
		} else {
			colorize := !noColor
			fmt.Print(formatter.FormatText(colorize))
		}
	}

	// Exit with code 1 if there are changes (useful for scripts)
	if diff.HasDrift() {
		os.Exit(1)
	}

	return nil
}
