package db

import (
	"context"
	"database/sql"
	"fmt"
)

type OMDb struct {
	ID     int            `db:"id"`
	APIKey sql.NullString `db:"api_key"`
}

func (q *Queries) GetOMDbSettings(ctx context.Context) (OMDb, error) {
	var omdb OMDb

	query := `
		SELECT id, api_key
		FROM omdb
		LIMIT 1;
	`

	row := q.db.QueryRowContext(ctx, query)
	err := row.Scan(
		&omdb.ID,
		&omdb.APIKey,
	)
	if err != nil {
		return omdb, err
	}

	return omdb, nil
}

func (q *Queries) UpdateOMDbSettings(ctx context.Context, apiKey string) error {
	query := `UPDATE omdb SET api_key = $1 WHERE id = 1`

	_, err := q.db.ExecContext(ctx, query, apiKey)
	if err != nil {
		return fmt.Errorf("error updating trakt settings: %v", err)
	}

	return nil
}
