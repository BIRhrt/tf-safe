package notifications

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"tf-safe/pkg/types"
)

// EmailNotifier implements Notifier for email
type EmailNotifier struct {
	config types.EmailConfig
}

// NewEmailNotifier creates a new Email notifier
func NewEmailNotifier(cfg types.EmailConfig) *EmailNotifier {
	return &EmailNotifier{
		config: cfg,
	}
}

func (e *EmailNotifier) Name() string {
	return "Email"
}

func (e *EmailNotifier) Send(ctx context.Context, event types.NotificationEvent) error {
	if len(e.config.Recipients) == 0 {
		return nil
	}

	subject := fmt.Sprintf("[tf-safe] %s: %s", getTitleForEventType(event.Type), event.Message)
	body := buildEmailBody(event)

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", strings.Join(e.config.Recipients, ","), subject, body))

	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPHost)
	addr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)

	// Note: In a real production environment, we might want to use a more robust library
	// or handle TLS explicitly, but net/smtp is sufficient for basic usage.
	err := smtp.SendMail(addr, auth, e.config.From, e.config.Recipients, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func buildEmailBody(event types.NotificationEvent) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Event: %s\n", event.Type))
	sb.WriteString(fmt.Sprintf("Time: %s\n", event.Timestamp.Format(time.RFC1123)))
	sb.WriteString(fmt.Sprintf("Message: %s\n", event.Message))

	if event.Error != "" {
		sb.WriteString(fmt.Sprintf("\nError:\n%s\n", event.Error))
	}

	if len(event.Metadata) > 0 {
		sb.WriteString("\nMetadata:\n")
		for k, v := range event.Metadata {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
		}
	}

	return sb.String()
}
