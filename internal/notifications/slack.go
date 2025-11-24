package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tf-safe/pkg/types"
)

// SlackNotifier implements Notifier for Slack
type SlackNotifier struct {
	config types.SlackConfig
	client *http.Client
}

// NewSlackNotifier creates a new Slack notifier
func NewSlackNotifier(cfg types.SlackConfig) *SlackNotifier {
	return &SlackNotifier{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SlackNotifier) Name() string {
	return "Slack"
}

func (s *SlackNotifier) Send(ctx context.Context, event types.NotificationEvent) error {
	payload := map[string]interface{}{
		"username":   s.config.Username,
		"icon_emoji": s.config.IconEmoji,
		"channel":    s.config.Channel,
		"attachments": []map[string]interface{}{
			{
				"color":  getColorForEventType(event.Type),
				"title":  getTitleForEventType(event.Type),
				"text":   event.Message,
				"fields": buildSlackFields(event),
				"footer": "tf-safe",
				"ts":     event.Timestamp.Unix(),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack API returned status: %d", resp.StatusCode)
	}

	return nil
}

func getColorForEventType(eventType types.NotificationType) string {
	switch eventType {
	case types.NotificationBackupSuccess, types.NotificationRestoreSuccess:
		return "good" // Green
	case types.NotificationBackupFailed, types.NotificationRestoreFailed:
		return "danger" // Red
	case types.NotificationDriftDetected:
		return "warning" // Yellow
	default:
		return "#3AA3E3" // Blue
	}
}

func getTitleForEventType(eventType types.NotificationType) string {
	switch eventType {
	case types.NotificationBackupSuccess:
		return "Backup Successful"
	case types.NotificationBackupFailed:
		return "Backup Failed"
	case types.NotificationRestoreSuccess:
		return "Restore Successful"
	case types.NotificationRestoreFailed:
		return "Restore Failed"
	case types.NotificationDriftDetected:
		return "Drift Detected"
	default:
		return "Notification"
	}
}

func buildSlackFields(event types.NotificationEvent) []map[string]interface{} {
	fields := make([]map[string]interface{}, 0)

	if event.Error != "" {
		fields = append(fields, map[string]interface{}{
			"title": "Error",
			"value": event.Error,
			"short": false,
		})
	}

	for k, v := range event.Metadata {
		fields = append(fields, map[string]interface{}{
			"title": k,
			"value": fmt.Sprintf("%v", v),
			"short": true,
		})
	}

	return fields
}
