package notifications

import (
	"context"
	"fmt"
	"sync"
	"time"

	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

// Notifier is the interface that all notification channels must implement
type Notifier interface {
	// Send sends a notification event
	Send(ctx context.Context, event types.NotificationEvent) error
	// Name returns the name of the notifier
	Name() string
}

// Manager handles the distribution of notifications to configured channels
type Manager struct {
	config    types.NotificationConfig
	notifiers []Notifier
	logger    *utils.Logger
	mu        sync.RWMutex
}

// NewManager creates a new notification manager
func NewManager(cfg types.NotificationConfig, logger *utils.Logger) *Manager {
	manager := &Manager{
		config:    cfg,
		notifiers: make([]Notifier, 0),
		logger:    logger,
	}

	if !cfg.Enabled {
		return manager
	}

	// Initialize configured notifiers
	if cfg.Slack.Enabled {
		manager.Register(NewSlackNotifier(cfg.Slack))
	}

	if cfg.Webhook.Enabled {
		manager.Register(NewWebhookNotifier(cfg.Webhook))
	}

	if cfg.Email.Enabled {
		manager.Register(NewEmailNotifier(cfg.Email))
	}

	return manager
}

// Register adds a new notifier to the manager
func (m *Manager) Register(n Notifier) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifiers = append(m.notifiers, n)
	m.logger.Debug("Registered notifier: %s", n.Name())
}

// Notify sends an event to all registered notifiers
func (m *Manager) Notify(ctx context.Context, event types.NotificationEvent) {
	if !m.config.Enabled {
		return
	}

	// Check if event type is enabled
	if !m.isEventEnabled(event.Type) {
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var wg sync.WaitGroup
	for _, n := range m.notifiers {
		wg.Add(1)
		go func(notifier Notifier) {
			defer wg.Done()
			if err := notifier.Send(ctx, event); err != nil {
				m.logger.Error("Failed to send notification via %s: %v", notifier.Name(), err)
			} else {
				m.logger.Debug("Sent notification via %s", notifier.Name())
			}
		}(n)
	}
	wg.Wait()
}

// isEventEnabled checks if the event type is configured for notifications
func (m *Manager) isEventEnabled(eventType types.NotificationType) bool {
	if len(m.config.Events) == 0 {
		// If no events specified, default to all (or could be none, but usually all is safer default if enabled)
		// However, DefaultNotificationConfig sets defaults. If user explicitly sets empty list, maybe they want none?
		// Let's assume if enabled but no events listed, we send all.
		return true
	}

	for _, e := range m.config.Events {
		if e == eventType {
			return true
		}
	}
	return false
}

// Helper methods to create common events

func (m *Manager) NotifyBackupSuccess(backupID string, size int64) {
	m.Notify(context.Background(), types.NotificationEvent{
		Type:      types.NotificationBackupSuccess,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Backup %s created successfully", backupID),
		Metadata: map[string]interface{}{
			"backup_id": backupID,
			"size":      size,
		},
	})
}

func (m *Manager) NotifyBackupFailed(err error) {
	m.Notify(context.Background(), types.NotificationEvent{
		Type:      types.NotificationBackupFailed,
		Timestamp: time.Now(),
		Message:   "Backup operation failed",
		Error:     err.Error(),
	})
}

func (m *Manager) NotifyRestoreSuccess(backupID string) {
	m.Notify(context.Background(), types.NotificationEvent{
		Type:      types.NotificationRestoreSuccess,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Successfully restored from backup %s", backupID),
		Metadata: map[string]interface{}{
			"backup_id": backupID,
		},
	})
}

func (m *Manager) NotifyRestoreFailed(backupID string, err error) {
	m.Notify(context.Background(), types.NotificationEvent{
		Type:      types.NotificationRestoreFailed,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Failed to restore from backup %s", backupID),
		Metadata: map[string]interface{}{
			"backup_id": backupID,
		},
		Error: err.Error(),
	})
}

func (m *Manager) NotifyDriftDetected(backupID string, changes string) {
	m.Notify(context.Background(), types.NotificationEvent{
		Type:      types.NotificationDriftDetected,
		Timestamp: time.Now(),
		Message:   "Configuration drift detected",
		Metadata: map[string]interface{}{
			"baseline_backup": backupID,
			"changes":         changes,
		},
	})
}
