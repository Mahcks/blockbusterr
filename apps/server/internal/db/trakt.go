package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type TraktSettings struct {
	ID           int    `db:"id"`
	ClientID     string `db:"client_id"`
	ClientSecret string `db:"client_secret"`
}

var ErrNoTraktSettings = errors.New("no trakt settings found")

func (q *Queries) GetTraktSettings(ctx context.Context) (TraktSettings, error) {
	var settings TraktSettings

	query := `SELECT id, client_id, client_secret FROM trakt`

	err := q.db.QueryRowContext(ctx, query).Scan(
		&settings.ID,
		&settings.ClientID,
		&settings.ClientSecret,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TraktSettings{}, ErrNoTraktSettings
		} else {
			return TraktSettings{}, fmt.Errorf("error fetching trakt settings: %v", err)
		}
	}

	return settings, nil
}

func (q *Queries) UpdateTraktSettings(ctx context.Context, clientID, clientSecret string) error {
	query := `UPDATE trakt SET client_id = $1, client_secret = $2 WHERE id = 1`

	_, err := q.db.ExecContext(ctx, query, clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("error updating trakt settings: %v", err)
	}

	return nil
}

func (q *Queries) CreateTraktSettings(ctx context.Context, clientID, clientSecret string) error {
	query := `INSERT INTO trakt (client_id, client_secret) VALUES ($1, $2)`
	_, err := q.db.ExecContext(ctx, query, clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("error creating trakt settings: %v", err)
	}

	return nil
}
