package db

import (
	"context"
	"database/sql"
	"fmt"
)

type OmbiSettings struct {
	ID              int            `db:"id"`                // Primary key with auto-increment
	APIKey          sql.NullString `db:"api_key"`           // API key required to make requests to Ombi
	URL             sql.NullString `db:"url"`               // Base URL for the Ombi server
	UserID          sql.NullString `db:"user_id"`           // User ID to use for Ombi
	Language        sql.NullString `db:"language"`          // Language for the Ombi server
	MovieQuality    sql.NullInt32  `db:"movie_quality"`     // Movie quality profile to use for Ombi
	MovieRootFolder sql.NullInt32  `db:"movie_root_folder"` // The root folder to use for Ombi
	ShowQuality     sql.NullInt32  `db:"show_quality"`      // Show quality profile to use for Ombi
	ShowRootFolder  sql.NullInt32  `db:"show_root_folder"`  // The root folder to use for Ombi
}

var ErrNoOmbiSettings = fmt.Errorf("no ombi settings found")

func (q *Queries) GetOmbiSettings(ctx context.Context) (OmbiSettings, error) {
	var settings OmbiSettings
	query := `
		SELECT id, api_key, url, user_id, language, movie_quality, movie_root_folder, show_quality, show_root_folder
		FROM ombi
		LIMIT 1;
	`

	row := q.db.QueryRowContext(ctx, query)
	err := row.Scan(
		&settings.ID,
		&settings.APIKey,
		&settings.URL,
		&settings.UserID,
		&settings.Language,
		&settings.MovieQuality,
		&settings.MovieRootFolder,
		&settings.ShowQuality,
		&settings.ShowRootFolder,
	)
	if err != nil {
		// Handle the case where there are no settings
		if err == sql.ErrNoRows {
			return settings, ErrNoOmbiSettings
		}

		// Return any other error that occurred during the query
		return settings, err
	}

	return settings, nil
}
