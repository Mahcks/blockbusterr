package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Settings represents a row in the `settings` table, using sql.NullXXX for nullable fields
type Settings struct {
	ID        int            `db:"id"`         // Primary key, auto-increment
	Key       string         `db:"key"`        // Unique key for the setting (non-nullable)
	Value     sql.NullString `db:"value"`      // Value of the setting (non-nullable)
	Type      sql.NullString `db:"type"`       // Type of the setting (non-nullable, defaults to 'text')
	UpdatedAt sql.NullTime   `db:"updated_at"` // Nullable DATETIME for when the setting was last updated
}

// GetSettingByKey fetches a setting by its key
func (q *Queries) GetSettingByKey(ctx context.Context, key string) (*Settings, error) {
	var setting Settings

	// Use a parameterized query to prevent SQL injection
	query := `SELECT id, key, value, type, updated_at FROM settings WHERE key = ?`

	// Execute the query with the provided context and key as a parameter
	err := q.db.QueryRowContext(ctx, query, key).Scan(
		&setting.ID,
		&setting.Key,
		&setting.Value,
		&setting.Type,
		&setting.UpdatedAt,
	)

	// Handle no rows found
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no setting found for key: %s", key)
	}

	// Handle other errors
	if err != nil {
		return nil, fmt.Errorf("error fetching setting: %v", err)
	}

	return &setting, nil
}
