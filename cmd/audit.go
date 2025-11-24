package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"tf-safe/internal/audit"
	"tf-safe/internal/config"
	"tf-safe/pkg/types"

	"github.com/spf13/cobra"
)

// auditLogCmd represents the audit-log command
var auditLogCmd = &cobra.Command{
	Use:   "audit-log",
	Short: "View audit logs",
	Long: `View and filter audit logs of tf-safe operations.

This command displays a history of actions performed by tf-safe, including
backups, restores, and configuration changes.

Examples:
  tf-safe audit-log                     # Show recent logs
  tf-safe audit-log --limit 20          # Show last 20 entries
  tf-safe audit-log --action restore    # Filter by action`,
	RunE: runAuditLogCommand,
}

func init() {
	rootCmd.AddCommand(auditLogCmd)

	// Add flags
	auditLogCmd.Flags().IntP("limit", "n", 50, "Number of entries to show")
	auditLogCmd.Flags().String("action", "", "Filter by action type")
}

func runAuditLogCommand(cmd *cobra.Command, args []string) error {
	// Get flags
	limit, err := cmd.Flags().GetInt("limit")
	if err != nil {
		return fmt.Errorf("failed to get limit flag: %w", err)
	}
	actionFilter, err := cmd.Flags().GetString("action")
	if err != nil {
		return fmt.Errorf("failed to get action flag: %w", err)
	}

	// Load configuration
	cfg, err := config.LoadConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create audit logger
	auditLogger := audit.NewLogger(cfg.Audit)

	// Get entries
	entries, err := auditLogger.GetEntries(limit * 2) // Get more to allow for filtering
	if err != nil {
		return fmt.Errorf("failed to get audit entries: %w", err)
	}

	// Filter and display
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tACTION\tSTATUS\tUSER\tDETAILS")

	count := 0
	for _, entry := range entries {
		if actionFilter != "" && string(entry.Action) != actionFilter {
			continue
		}

		status := string(entry.Status)
		if entry.Status == types.AuditStatusSuccess {
			status = "✅ " + status
		} else {
			status = "❌ " + status
		}

		details := ""
		if entry.Error != "" {
			details = fmt.Sprintf("Error: %s", entry.Error)
		} else if len(entry.Details) > 0 {
			// Format key details
			if backupID, ok := entry.Details["backup_id"]; ok {
				details = fmt.Sprintf("Backup: %v", backupID)
			} else if baseline, ok := entry.Details["baseline_backup"]; ok {
				details = fmt.Sprintf("Baseline: %v", baseline)
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			entry.Timestamp.Local().Format(time.RFC3339),
			entry.Action,
			status,
			entry.User,
			details,
		)

		count++
		if count >= limit {
			break
		}
	}
	w.Flush()

	if count == 0 {
		fmt.Println("No audit logs found.")
	}

	return nil
}
