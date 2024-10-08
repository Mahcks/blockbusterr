package db

import (
	"context"
	"database/sql"
	"fmt"
)

type SonarrSettings struct {
	ID           int            `db:"id"`            // Primary key with auto-increment
	APIKey       sql.NullString `db:"api_key"`       // API key required to make requests to Sonarr
	URL          sql.NullString `db:"url"`           // Base URL for the Sonarr server
	Language     sql.NullString `db:"language"`      // ???
	Quality      sql.NullInt32  `db:"quality"`       // Quality profile to use for Sonarr
	RootFolder   sql.NullInt32  `db:"root_folder"`   // The root folder to use for Sonarr
	SeasonFolder sql.NullBool   `db:"season_folder"` // Whether to use season folders
}

var ErrNoSonarrSettings = fmt.Errorf("no sonarr settings found")

func (q *Queries) GetSonarrSettings(ctx context.Context) (SonarrSettings, error) {
	var settings SonarrSettings
	query := `
		SELECT id, api_key, url, language, quality, root_folder, season_folder
		FROM sonarr
		LIMIT 1;
	`

	row := q.db.QueryRowContext(ctx, query)
	err := row.Scan(
		&settings.ID,
		&settings.APIKey,
		&settings.URL,
		&settings.Language,
		&settings.Quality,
		&settings.RootFolder,
		&settings.SeasonFolder,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle the case where there are no settings
			return settings, ErrNoSonarrSettings
		}
		return settings, err
	}

	return settings, nil
}

func (q *Queries) UpdateSonarrSettings(ctx context.Context, apiKey, url, language string, quality, rootFolder int32, seasonFolder bool) error {
	query := `
		UPDATE sonarr
		SET api_key = $1, url = $2, language = $3, quality = $4, root_folder = $5, season_folder = $6
		WHERE id = 1;
	`

	_, err := q.db.ExecContext(ctx, query, apiKey, url, language, quality, rootFolder, seasonFolder)
	if err != nil {
		return err
	}

	return nil
}

func (q *Queries) CreateSonarrSettings(ctx context.Context, apiKey, url, language string, quality, rootFolder int32, seasonFolder bool) error {
	query := `
		INSERT INTO sonarr (api_key, url, language, quality, root_folder, season_folder)
		VALUES ($1, $2, $3, $4, $5, $6);
	`

	_, err := q.db.ExecContext(ctx, query, apiKey, url, language, quality, rootFolder, seasonFolder)
	if err != nil {
		return err
	}

	return nil
}

func (q *Queries) GetShowInterval(ctx context.Context) (sql.NullInt32, error) {
	var interval sql.NullInt32

	err := q.db.QueryRowContext(ctx, `
		SELECT interval
		FROM show_settings
		LIMIT 1;
	`).Scan(&interval)
	if err != nil {
		if err == sql.ErrNoRows {
			return interval, fmt.Errorf("no show interval found")
		}
		return interval, err
	}

	return interval, nil
}
