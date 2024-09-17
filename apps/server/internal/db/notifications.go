package db

import "context"

type NotificationSettings struct {
	ID         int    `db:"id"`          // Primary key with auto-increment
	Platform   string `db:"platform"`    // Platform to send notifications to
	Enabled    bool   `db:"enabled"`     // Whether notifications are enabled for this platform
	WebhookURL string `db:"webhook_url"` // Webhook URL to send notifications to
}

func (q *Queries) GetNotificationSettings(ctx context.Context) (NotificationSettings, error) {
	var settings NotificationSettings
	query := `
		SELECT id, platform, enabled, webhook_url
		FROM notifications_config
		LIMIT 1;
	`

	row := q.db.QueryRowContext(ctx, query)
	err := row.Scan(
		&settings.ID,
		&settings.Platform,
		&settings.Enabled,
		&settings.WebhookURL,
	)
	if err != nil {
		return settings, err
	}

	return settings, nil
}
