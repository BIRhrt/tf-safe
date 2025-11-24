package types

import (
	"time"
)

// NotificationType represents the type of notification event
type NotificationType string

const (
	NotificationBackupSuccess  NotificationType = "backup_success"
	NotificationBackupFailed   NotificationType = "backup_failed"
	NotificationRestoreSuccess NotificationType = "restore_success"
	NotificationRestoreFailed  NotificationType = "restore_failed"
	NotificationDriftDetected  NotificationType = "drift_detected"
)

// NotificationEvent represents a system event that triggers a notification
type NotificationEvent struct {
	Type      NotificationType       `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// NotificationConfig holds configuration for all notification channels
type NotificationConfig struct {
	Enabled bool               `mapstructure:"enabled" yaml:"enabled"`
	Slack   SlackConfig        `mapstructure:"slack" yaml:"slack"`
	Webhook WebhookConfig      `mapstructure:"webhook" yaml:"webhook"`
	Email   EmailConfig        `mapstructure:"email" yaml:"email"`
	Events  []NotificationType `mapstructure:"events" yaml:"events"` // Events to subscribe to
}

// SlackConfig holds configuration for Slack notifications
type SlackConfig struct {
	Enabled    bool   `mapstructure:"enabled" yaml:"enabled"`
	WebhookURL string `mapstructure:"webhook_url" yaml:"webhook_url"`
	Channel    string `mapstructure:"channel" yaml:"channel"`
	Username   string `mapstructure:"username" yaml:"username"`
	IconEmoji  string `mapstructure:"icon_emoji" yaml:"icon_emoji"`
}

// WebhookConfig holds configuration for generic webhook notifications
type WebhookConfig struct {
	Enabled bool              `mapstructure:"enabled" yaml:"enabled"`
	URL     string            `mapstructure:"url" yaml:"url"`
	Method  string            `mapstructure:"method" yaml:"method"` // Default: POST
	Headers map[string]string `mapstructure:"headers" yaml:"headers"`
}

// EmailConfig holds configuration for email notifications
type EmailConfig struct {
	Enabled    bool     `mapstructure:"enabled" yaml:"enabled"`
	SMTPHost   string   `mapstructure:"smtp_host" yaml:"smtp_host"`
	SMTPPort   int      `mapstructure:"smtp_port" yaml:"smtp_port"`
	Username   string   `mapstructure:"username" yaml:"username"`
	Password   string   `mapstructure:"password" yaml:"password"`
	From       string   `mapstructure:"from" yaml:"from"`
	Recipients []string `mapstructure:"recipients" yaml:"recipients"`
}

// DefaultNotificationConfig returns the default notification configuration
func DefaultNotificationConfig() NotificationConfig {
	return NotificationConfig{
		Enabled: false,
		Events: []NotificationType{
			NotificationBackupFailed,
			NotificationRestoreFailed,
			NotificationDriftDetected,
		},
		Slack: SlackConfig{
			Enabled:   false,
			Username:  "tf-safe",
			IconEmoji: ":shield:",
		},
		Webhook: WebhookConfig{
			Enabled: false,
			Method:  "POST",
		},
		Email: EmailConfig{
			Enabled:  false,
			SMTPPort: 587,
		},
	}
}
