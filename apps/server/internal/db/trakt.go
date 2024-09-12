package db

import (
	"context"
	"fmt"
)

type TraktSettings struct {
	ID           int    `db:"id"`
	ClientID     string `db:"client_id"`
	ClientSecret string `db:"client_secret"`
}

func (q *Queries) GetTraktSettings(ctx context.Context) (TraktSettings, error) {
	var settings TraktSettings

	query := `SELECT id, client_id, client_secret FROM trakt`

	err := q.db.QueryRowContext(ctx, query).Scan(
		&settings.ID,
		&settings.ClientID,
		&settings.ClientSecret,
	)
	if err != nil {
		fmt.Println(err)
		return TraktSettings{}, fmt.Errorf("error fetching trakt settings: %v", err)
	}

	return settings, nil
}
