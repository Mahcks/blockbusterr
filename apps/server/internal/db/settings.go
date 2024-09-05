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

// InsertOrUpdateSetting inserts or updates a setting by key
func (q *Queries) InsertOrUpdateSetting(ctx context.Context, key string, value string, settingType string) error {
	// Parameterized query for upsert behavior (insert or update if exists)
	query := `
		INSERT INTO settings (key, value, type, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at;
	`

	// Execute the query with provided key, value, and type
	_, err := q.db.ExecContext(ctx, query, key, value, settingType)
	if err != nil {
		return fmt.Errorf("error inserting/updating setting: %v", err)
	}

	return nil
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

// DeleteSettingByKey deletes a setting from the database based on the key
func (q *Queries) DeleteSettingByKey(ctx context.Context, key string) error {
	// SQL query to delete a row where the key matches the provided key
	query := `DELETE FROM settings WHERE key = ?`

	// Execute the query with the provided context and key as a parameter
	result, err := q.db.ExecContext(ctx, query, key)
	if err != nil {
		return fmt.Errorf("error deleting setting: %v", err)
	}

	// Check how many rows were affected to ensure the delete was successful
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no setting found for key: %s", key)
	}

	return nil
}
