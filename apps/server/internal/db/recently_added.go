package db

import (
	"context"
	"errors"
	"time"

	"github.com/mattn/go-sqlite3"
)

type RecentlyAddedMedia struct {
	ID        int       `db:"id"`         // Primary key with auto-increment
	MediaType string    `db:"media_type"` // Type of media (MOVIE, SHOW)
	Title     string    `db:"title"`      // Title of the media
	Year      int       `db:"year"`       // Year the media was released
	Summary   string    `db:"summary"`    // Summary of the media
	IMDBID    string    `db:"imdb_id"`    // IMDb ID of the media
	Poster    string    `db:"poster"`     // URL to the poster of the media
	AddedAt   time.Time `db:"added_at"`   // Time the media was added
}

func (q *Queries) GetRecentlyAddedMedia(ctx context.Context, limit, offset int) ([]RecentlyAddedMedia, error) {
	var recentlyAddedList []RecentlyAddedMedia

	query := `
		SELECT id, media_type, title, year, summary, imdb_id, poster, added_at
		FROM recently_added
		ORDER BY added_at DESC
		LIMIT $1 OFFSET $2;
	`

	rows, err := q.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Ensure the rows are closed once we're done

	for rows.Next() {
		var recentlyAdded RecentlyAddedMedia
		err := rows.Scan(
			&recentlyAdded.ID,
			&recentlyAdded.MediaType,
			&recentlyAdded.Title,
			&recentlyAdded.Year,
			&recentlyAdded.Summary,
			&recentlyAdded.IMDBID,
			&recentlyAdded.Poster,
			&recentlyAdded.AddedAt,
		)
		if err != nil {
			return nil, err
		}

		// Add the scanned row to the list
		recentlyAddedList = append(recentlyAddedList, recentlyAdded)
	}

	// Check for errors encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return recentlyAddedList, nil
}

func (q *Queries) AddToRecentlyAddedMedia(ctx context.Context, mediaType string, title string, year int, summary string, imdbID string, poster string) error {
	query := `
		INSERT INTO recently_added (media_type, title, year, summary, imdb_id, poster)
		VALUES (?, ?, ?, ?, ?, ?);
	`

	_, err := q.db.ExecContext(ctx, query, mediaType, title, year, summary, imdbID, poster)
	if err != nil {
		// Check if it's a UNIQUE constraint violation
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint {
			// Ignore the UNIQUE constraint error
			return nil
		}
		// Return other errors
		return err
	}

	return nil
}
