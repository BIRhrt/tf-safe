package notifications

import (
	"context"
	"errors"
	"testing"
	"time"

	"tf-safe/internal/utils"
	"tf-safe/pkg/types"
)

// MockNotifier implements Notifier for testing
type MockNotifier struct {
	name       string
	SentEvents []types.NotificationEvent
	ShouldFail bool
}

func (m *MockNotifier) Name() string {
	return m.name
}

func (m *MockNotifier) Send(ctx context.Context, event types.NotificationEvent) error {
	if m.ShouldFail {
		return errors.New("mock failure")
	}
	m.SentEvents = append(m.SentEvents, event)
	return nil
}

func TestManager_Notify(t *testing.T) {
	logger := utils.NewLogger(utils.LogLevelDebug)

	tests := []struct {
		name          string
		config        types.NotificationConfig
		event         types.NotificationEvent
		expectedSends int
	}{
		{
			name: "disabled notifications",
			config: types.NotificationConfig{
				Enabled: false,
			},
			event: types.NotificationEvent{
				Type: types.NotificationBackupSuccess,
			},
			expectedSends: 0,
		},
		{
			name: "enabled notifications - all events",
			config: types.NotificationConfig{
				Enabled: true,
				Events:  []types.NotificationType{}, // Empty means all
			},
			event: types.NotificationEvent{
				Type: types.NotificationBackupSuccess,
			},
			expectedSends: 1,
		},
		{
			name: "enabled notifications - specific event match",
			config: types.NotificationConfig{
				Enabled: true,
				Events:  []types.NotificationType{types.NotificationBackupSuccess},
			},
			event: types.NotificationEvent{
				Type: types.NotificationBackupSuccess,
			},
			expectedSends: 1,
		},
		{
			name: "enabled notifications - specific event mismatch",
			config: types.NotificationConfig{
				Enabled: true,
				Events:  []types.NotificationType{types.NotificationBackupFailed},
			},
			event: types.NotificationEvent{
				Type: types.NotificationBackupSuccess,
			},
			expectedSends: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager(tt.config, logger)

			// Manually register mock notifier since NewManager only registers configured ones
			mock := &MockNotifier{name: "Mock"}
			manager.Register(mock)

			manager.Notify(context.Background(), tt.event)

			// Wait a bit for goroutines to finish (Notify uses goroutines)
			time.Sleep(10 * time.Millisecond)

			if len(mock.SentEvents) != tt.expectedSends {
				t.Errorf("Notify() sent %d events, want %d", len(mock.SentEvents), tt.expectedSends)
			}
		})
	}
}

func TestManager_HelperMethods(t *testing.T) {
	logger := utils.NewLogger(utils.LogLevelDebug)
	config := types.NotificationConfig{
		Enabled: true,
	}
	manager := NewManager(config, logger)
	mock := &MockNotifier{name: "Mock"}
	manager.Register(mock)

	// Test NotifyBackupSuccess
	manager.NotifyBackupSuccess("backup-1", 100)
	time.Sleep(10 * time.Millisecond)

	if len(mock.SentEvents) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(mock.SentEvents))
	}
	if mock.SentEvents[0].Type != types.NotificationBackupSuccess {
		t.Errorf("Expected event type %s, got %s", types.NotificationBackupSuccess, mock.SentEvents[0].Type)
	}

	// Test NotifyBackupFailed
	manager.NotifyBackupFailed(errors.New("test error"))
	time.Sleep(10 * time.Millisecond)

	if len(mock.SentEvents) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(mock.SentEvents))
	}
	if mock.SentEvents[1].Type != types.NotificationBackupFailed {
		t.Errorf("Expected event type %s, got %s", types.NotificationBackupFailed, mock.SentEvents[1].Type)
	}
	if mock.SentEvents[1].Error != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", mock.SentEvents[1].Error)
	}
}
