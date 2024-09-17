package notifications

import (
	"encoding/json"
	"fmt"

	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

type NotificationProvider interface {
	SendNotification(notificationType structures.NotificationType, payload json.RawMessage) error
}

type NotificationManager struct {
	providers []NotificationProvider
}

// NewNotificationManager loads the notification settings from the database and initializes the providers
func NewNotificationManager(gctx global.Context) (*NotificationManager, error) {
	// Fetch notification settings from the database
	settings, err := gctx.Crate().SQL.Queries().GetNotificationSettings(gctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load notification settings: %w", err)
	}

	// Initialize the NotificationManager
	var providers []NotificationProvider

	// Check if Discord notifications are enabled and add the provider
	if settings.Platform == "discord" && settings.Enabled {
		discordNotifier := NewDiscordNotification(settings.WebhookURL, "Blockbusterr")
		providers = append(providers, discordNotifier)
	}

	// Add other providers based on settings.Platform (Slack, etc.) as needed

	return &NotificationManager{providers: providers}, nil
}

// SendNotification sends notifications to all enabled providers
func (m *NotificationManager) SendNotification(notificationType structures.NotificationType, payload json.RawMessage) error {
	for _, provider := range m.providers {
		if err := provider.SendNotification(notificationType, payload); err != nil {
			fmt.Printf("Failed to send notification via provider: %v\n", err)
		}
	}
	return nil
}
