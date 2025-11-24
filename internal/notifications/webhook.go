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

// WebhookNotifier implements Notifier for generic webhooks
type WebhookNotifier struct {
	config types.WebhookConfig
	client *http.Client
}

// NewWebhookNotifier creates a new Webhook notifier
func NewWebhookNotifier(cfg types.WebhookConfig) *WebhookNotifier {
	return &WebhookNotifier{
		config: cfg,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WebhookNotifier) Name() string {
	return "Webhook"
}

func (w *WebhookNotifier) Send(ctx context.Context, event types.NotificationEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	method := w.config.Method
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequestWithContext(ctx, method, w.config.URL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "tf-safe/1.0")

	// Add custom headers
	for k, v := range w.config.Headers {
		req.Header.Set(k, v)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status: %d", resp.StatusCode)
	}

	return nil
}
