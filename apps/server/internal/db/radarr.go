package db

import (
	"context"
	"database/sql"
	"fmt"
)

type RadarrSettings struct {
	ID                  int            `db:"id"`                   // Primary key with auto-increment
	APIKey              sql.NullString `db:"api_key"`              // API key required to make requests to Radarr
	URL                 sql.NullString `db:"url"`                  // Base URL for the Radarr server
	MinimumAvailability sql.NullString `db:"minimum_availability"` // Minimum availability setting ("announced", "in_cinemas", "released")
	Quality             sql.NullInt32  `db:"quality"`              // Quality profile to use for Radarr
	RootFolder          sql.NullInt32  `db:"root_folder"`          // The root folder to use for Radarr
}

func (q *Queries) GetRadarrSettings(ctx context.Context) (RadarrSettings, error) {
	var settings RadarrSettings
	query := `
		SELECT id, api_key, url, minimum_availability, quality, root_folder
		FROM radarr
		LIMIT 1;
	`

	row := q.db.QueryRowContext(ctx, query)
	err := row.Scan(
		&settings.ID,
		&settings.APIKey,
		&settings.URL,
		&settings.MinimumAvailability,
		&settings.Quality,
		&settings.RootFolder,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle the case where there are no settings
			return settings, fmt.Errorf("no Radarr settings found")
		}
		// Return any other error that occurred during the query
		return settings, err
	}

	return settings, nil
}

func (q *Queries) UpdateRadarrSettings(ctx context.Context, apiKey, url, minimumAvailability sql.NullString, quality, rootFolder sql.NullInt32) error {
	query := `
		UPDATE radarr
		SET api_key = $1, url = $2, minimum_availability = $3, quality = $4, root_folder = $5
		WHERE id = 1;
	`

	_, err := q.db.ExecContext(ctx, query, apiKey, url, minimumAvailability, quality, rootFolder)
	if err != nil {
		return err
	}

	return nil
}
