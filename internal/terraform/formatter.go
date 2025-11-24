package terraform

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// DiffFormatter provides different output formats for state diffs
type DiffFormatter struct {
	diff *StateDiff
}

// NewDiffFormatter creates a new diff formatter
func NewDiffFormatter(diff *StateDiff) *DiffFormatter {
	return &DiffFormatter{diff: diff}
}

// FormatText returns a human-readable text format of the diff
func (f *DiffFormatter) FormatText(colorize bool) string {
	if !f.diff.HasDrift() {
		if colorize {
			return color("✓ No changes detected - states are identical\n", colorGreen)
		}
		return "✓ No changes detected - states are identical\n"
	}

	var sb strings.Builder

	// Summary header
	summary := f.diff.FormatSummary()
	if colorize {
		sb.WriteString(fmt.Sprintf("\n%s\n\n", color(fmt.Sprintf("Changes detected: %s", summary), colorCyan)))
	} else {
		sb.WriteString(fmt.Sprintf("\nChanges detected: %s\n\n", summary))
	}

	// Added resources
	if len(f.diff.Added) > 0 {
		header := fmt.Sprintf("│ Added (%d):", len(f.diff.Added))
		if colorize {
			sb.WriteString(color(header, colorGreen))
		} else {
			sb.WriteString(header)
		}
		sb.WriteString("\n")

		for _, change := range f.diff.Added {
			symbol := "  + "
			if colorize {
				symbol = color(symbol, colorGreen)
			}
			sb.WriteString(fmt.Sprintf("%s%s\n", symbol, change.Address))

			// Show key attributes if available
			if len(change.After) > 0 {
				f.formatAttributes(&sb, change.After, "    ", colorize)
			}
			sb.WriteString("\n")
		}
	}

	// Modified resources
	if len(f.diff.Modified) > 0 {
		header := fmt.Sprintf("│ Modified (%d):", len(f.diff.Modified))
		if colorize {
			sb.WriteString(color(header, colorYellow))
		} else {
			sb.WriteString(header)
		}
		sb.WriteString("\n")

		for _, change := range f.diff.Modified {
			symbol := "  ~ "
			if colorize {
				symbol = color(symbol, colorYellow)
			}
			sb.WriteString(fmt.Sprintf("%s%s\n", symbol, change.Address))

			// Show attribute changes
			if len(change.AttributeDiff) > 0 {
				f.formatAttributeDiff(&sb, change.AttributeDiff, "    ", colorize)
			}
			sb.WriteString("\n")
		}
	}

	// Removed resources
	if len(f.diff.Removed) > 0 {
		header := fmt.Sprintf("│ Removed (%d):", len(f.diff.Removed))
		if colorize {
			sb.WriteString(color(header, colorRed))
		} else {
			sb.WriteString(header)
		}
		sb.WriteString("\n")

		for _, change := range f.diff.Removed {
			symbol := "  - "
			if colorize {
				symbol = color(symbol, colorRed)
			}
			sb.WriteString(fmt.Sprintf("%s%s\n", symbol, change.Address))

			// Show key attributes if available
			if len(change.Before) > 0 {
				f.formatAttributes(&sb, change.Before, "    ", colorize)
			}
			sb.WriteString("\n")
		}
	}

	// Summary footer
	sb.WriteString(strings.Repeat("─", 60))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Summary: %s\n", summary))

	return sb.String()
}

// FormatJSON returns a JSON format of the diff
func (f *DiffFormatter) FormatJSON(pretty bool) (string, error) {
	var data []byte
	var err error

	if pretty {
		data, err = json.MarshalIndent(f.diff, "", "  ")
	} else {
		data, err = json.Marshal(f.diff)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal diff to JSON: %w", err)
	}

	return string(data), nil
}

// FormatYAML returns a YAML format of the diff
func (f *DiffFormatter) FormatYAML() (string, error) {
	data, err := yaml.Marshal(f.diff)
	if err != nil {
		return "", fmt.Errorf("failed to marshal diff to YAML: %w", err)
	}
	return string(data), nil
}

// FormatCompact returns a compact one-line summary
func (f *DiffFormatter) FormatCompact() string {
	if !f.diff.HasDrift() {
		return "No changes"
	}
	return f.diff.FormatSummary()
}

// formatAttributes formats a map of attributes for display
func (f *DiffFormatter) formatAttributes(sb *strings.Builder, attrs map[string]interface{}, indent string, colorize bool) {
	// Get sorted keys for consistent output
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		// Skip internal/computed attributes
		if !strings.HasPrefix(k, "_") && k != "id" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Limit to first 5 attributes
	maxAttrs := 5
	for i, key := range keys {
		if i >= maxAttrs {
			remaining := len(keys) - maxAttrs
			sb.WriteString(fmt.Sprintf("%s... and %d more attributes\n", indent, remaining))
			break
		}

		value := formatValue(attrs[key])
		sb.WriteString(fmt.Sprintf("%s%s = %s\n", indent, key, value))
	}
}

// formatAttributeDiff formats attribute changes for display
func (f *DiffFormatter) formatAttributeDiff(sb *strings.Builder, attrDiff map[string]AttributeChange, indent string, colorize bool) {
	// Get sorted keys
	keys := make([]string, 0, len(attrDiff))
	for k := range attrDiff {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		change := attrDiff[key]

		if change.Before == nil && change.After != nil {
			// Added attribute
			symbol := "+ "
			if colorize {
				symbol = color(symbol, colorGreen)
			}
			sb.WriteString(fmt.Sprintf("%s%s%s = %s\n", indent, symbol, key, formatValue(change.After)))
		} else if change.Before != nil && change.After == nil {
			// Removed attribute
			symbol := "- "
			if colorize {
				symbol = color(symbol, colorRed)
			}
			sb.WriteString(fmt.Sprintf("%s%s%s = %s\n", indent, symbol, key, formatValue(change.Before)))
		} else {
			// Modified attribute
			symbol := "~ "
			if colorize {
				symbol = color(symbol, colorYellow)
			}
			sb.WriteString(fmt.Sprintf("%s%s%s: %s → %s\n", indent, symbol, key,
				formatValue(change.Before), formatValue(change.After)))
		}
	}
}

// formatValue formats a value for display
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		if len(v) > 50 {
			return fmt.Sprintf("\"%s...\"", v[:47])
		}
		return fmt.Sprintf("\"%s\"", v)
	case nil:
		return "<null>"
	case bool:
		return fmt.Sprintf("%t", v)
	case float64:
		return fmt.Sprintf("%.0f", v)
	case map[string]interface{}:
		return "{...}"
	case []interface{}:
		return fmt.Sprintf("[%d items]", len(v))
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// color wraps text in ANSI color codes
func color(text string, colorCode string) string {
	return colorCode + text + colorReset
}
