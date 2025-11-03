package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"tf-safe/internal/config"
	"tf-safe/pkg/types"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize tf-safe configuration",
	Long: `Initialize tf-safe configuration by creating a default .tf-safe.yaml file.
	
This command will create a configuration file with sensible defaults
and provide interactive prompts for common configuration options.

Examples:
  tf-safe init                    # Create default configuration
  tf-safe init -i                # Interactive configuration setup
  tf-safe init -f                # Force overwrite existing config
  tf-safe init -t minimal        # Use minimal template`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
	
	// Add init-specific flags
	initCmd.Flags().BoolP("interactive", "i", false, "Use interactive mode for configuration")
	initCmd.Flags().BoolP("force", "f", false, "Overwrite existing configuration file")
	initCmd.Flags().StringP("template", "t", "default", "Configuration template to use (default, minimal, enterprise)")
	initCmd.Flags().StringP("output", "o", ".tf-safe.yaml", "Output configuration file path")
}

func runInit(cmd *cobra.Command, args []string) error {
	interactive, err := cmd.Flags().GetBool("interactive")
	if err != nil {
		return fmt.Errorf("failed to get interactive flag: %w", err)
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("failed to get force flag: %w", err)
	}
	template, err := cmd.Flags().GetString("template")
	if err != nil {
		return fmt.Errorf("failed to get template flag: %w", err)
	}
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		return fmt.Errorf("failed to get output flag: %w", err)
	}
	
	// Check if config file already exists
	if _, statErr := os.Stat(output); statErr == nil && !force {
		return fmt.Errorf("configuration file %s already exists. Use --force to overwrite", output)
	}
	
	var cfg *types.Config
	
	if interactive {
		cfg, err = createInteractiveConfig()
		if err != nil {
			return fmt.Errorf("failed to create interactive configuration: %w", err)
		}
	} else {
		cfg, err = createTemplateConfig(template)
		if err != nil {
			return fmt.Errorf("failed to create template configuration: %w", err)
		}
	}
	
	// Save configuration
	manager := config.NewManager()
	if err := manager.Save(cfg, output); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	
	fmt.Printf("✅ Configuration file created: %s\n", output)
	
	// Validate the created configuration
	if err := manager.Validate(cfg); err != nil {
		fmt.Printf("⚠️  Warning: Configuration validation failed: %v\n", err)
		fmt.Println("Please review and correct the configuration file.")
	} else {
		fmt.Println("✅ Configuration is valid")
	}
	
	// Show next steps
	showNextSteps(output)
	
	return nil
}

func createTemplateConfig(template string) (*types.Config, error) {
	cfg, found := config.GetTemplate(template)
	if !found {
		templates := config.GetAvailableTemplates()
		var names []string
		for _, t := range templates {
			names = append(names, t.Name)
		}
		return nil, fmt.Errorf("unknown template: %s. Available templates: %s", template, strings.Join(names, ", "))
	}
	return cfg, nil
}



func createInteractiveConfig() (*types.Config, error) {
	fmt.Println("🔧 Interactive tf-safe configuration setup")
	fmt.Println("Press Enter to use default values shown in [brackets]")
	fmt.Println()
	
	reader := bufio.NewReader(os.Stdin)
	cfg := config.DefaultConfig()
	
	// Local storage configuration
	fmt.Println("📁 Local Storage Configuration")
	
	if enabled := promptBool(reader, "Enable local backups", cfg.Local.Enabled); enabled {
		cfg.Local.Enabled = true
		cfg.Local.Path = promptString(reader, "Local backup directory", cfg.Local.Path)
		cfg.Local.RetentionCount = promptInt(reader, "Number of local backups to keep", cfg.Local.RetentionCount)
	} else {
		cfg.Local.Enabled = false
	}
	
	fmt.Println()
	
	// Remote storage configuration
	fmt.Println("☁️  Remote Storage Configuration")
	
	if enabled := promptBool(reader, "Enable remote backups", cfg.Remote.Enabled); enabled {
		cfg.Remote.Enabled = true
		cfg.Remote.Provider = promptChoice(reader, "Remote storage provider", 
			[]string{"s3", "gcs", "azure"}, cfg.Remote.Provider)
		cfg.Remote.Bucket = promptString(reader, "Bucket name", cfg.Remote.Bucket)
		
		if cfg.Remote.Provider == "s3" {
			cfg.Remote.Region = promptString(reader, "AWS region", cfg.Remote.Region)
		}
		
		cfg.Remote.Prefix = promptString(reader, "Backup prefix (optional)", cfg.Remote.Prefix)
	} else {
		cfg.Remote.Enabled = false
	}
	
	fmt.Println()
	
	// Encryption configuration
	fmt.Println("🔐 Encryption Configuration")
	
	cfg.Encryption.Provider = promptChoice(reader, "Encryption provider", 
		[]string{"none", "aes", "kms", "passphrase"}, cfg.Encryption.Provider)
	
	switch cfg.Encryption.Provider {
	case "kms":
		cfg.Encryption.KMSKeyID = promptString(reader, "KMS Key ID or ARN", cfg.Encryption.KMSKeyID)
	case "passphrase":
		cfg.Encryption.Passphrase = promptPassword(reader, "Encryption passphrase")
	}
	
	fmt.Println()
	
	// Retention configuration
	fmt.Println("🗂️  Retention Configuration")
	
	cfg.Retention.LocalCount = promptInt(reader, "Local backup retention count", cfg.Retention.LocalCount)
	if cfg.Remote.Enabled {
		cfg.Retention.RemoteCount = promptInt(reader, "Remote backup retention count", cfg.Retention.RemoteCount)
	}
	cfg.Retention.MaxAgeDays = promptInt(reader, "Maximum backup age (days)", cfg.Retention.MaxAgeDays)
	
	fmt.Println()
	
	// Logging configuration
	fmt.Println("📝 Logging Configuration")
	
	cfg.Logging.Level = promptChoice(reader, "Log level", 
		[]string{"debug", "info", "warn", "error"}, cfg.Logging.Level)
	cfg.Logging.Format = promptChoice(reader, "Log format", 
		[]string{"text", "json"}, cfg.Logging.Format)
	
	return cfg, nil
}

func promptString(reader *bufio.Reader, prompt, defaultValue string) string {
	fmt.Printf("%s [%s]: ", prompt, defaultValue)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func promptBool(reader *bufio.Reader, prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}
	
	fmt.Printf("%s (y/n) [%s]: ", prompt, defaultStr)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	input = strings.TrimSpace(strings.ToLower(input))
	
	if input == "" {
		return defaultValue
	}
	
	return input == "y" || input == "yes"
}

func promptInt(reader *bufio.Reader, prompt string, defaultValue int) int {
	fmt.Printf("%s [%d]: ", prompt, defaultValue)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	input = strings.TrimSpace(input)
	
	if input == "" {
		return defaultValue
	}
	
	if value, err := strconv.Atoi(input); err == nil {
		return value
	}
	
	return defaultValue
}

func promptChoice(reader *bufio.Reader, prompt string, choices []string, defaultValue string) string {
	fmt.Printf("%s (%s) [%s]: ", prompt, strings.Join(choices, "/"), defaultValue)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}
	input = strings.TrimSpace(input)
	
	if input == "" {
		return defaultValue
	}
	
	// Validate choice
	for _, choice := range choices {
		if input == choice {
			return input
		}
	}
	
	return defaultValue
}

func promptPassword(reader *bufio.Reader, prompt string) string {
	fmt.Printf("%s: ", prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(input)
}

func showNextSteps(configPath string) {
	fmt.Println()
	fmt.Println("🚀 Next Steps:")
	fmt.Printf("1. Review the configuration file: %s\n", configPath)
	fmt.Println("2. Test your configuration: tf-safe backup")
	fmt.Println("3. List available backups: tf-safe list")
	fmt.Println("4. Use tf-safe with Terraform: tf-safe apply")
	fmt.Println()
	fmt.Println("📚 For more information, visit: https://github.com/BIRhrt/tf-safe")
}