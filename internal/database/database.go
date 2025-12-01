package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

type ActivityLog struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	JobType   string    `json:"job_type"`
	MediaType string    `json:"media_type"` // "movie" or "show"
	Title     string    `json:"title"`
	Year      int       `json:"year"`
	TMDBID    int       `json:"tmdb_id,omitempty"`
	TVDBID    int       `json:"tvdb_id,omitempty"`
	IMDBID    string    `json:"imdb_id,omitempty"`
	PosterURL string    `json:"poster_url,omitempty"`
	Status    string    `json:"status"` // "added", "failed", "skipped"
	Message   string    `json:"message,omitempty"`
}

func New(dataDir string) (*Database, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "blockbusterr.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	database := &Database{db: db}

	// Initialize schema
	if err := database.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return database, nil
}

func (d *Database) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS activity_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		job_type TEXT NOT NULL,
		media_type TEXT NOT NULL,
		title TEXT NOT NULL,
		year INTEGER,
		tmdb_id INTEGER,
		tvdb_id INTEGER,
		imdb_id TEXT,
		poster_url TEXT,
		status TEXT NOT NULL,
		message TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_activity_timestamp ON activity_logs(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_activity_job_type ON activity_logs(job_type);
	CREATE INDEX IF NOT EXISTS idx_activity_status ON activity_logs(status);
	`

	_, err := d.db.Exec(schema)
	if err != nil {
		return err
	}

	// Migrate existing tables if needed (add new columns)
	// Check if imdb_id column exists
	var columnCount int
	err = d.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('activity_logs') WHERE name='imdb_id'").Scan(&columnCount)
	if err != nil {
		return err
	}

	if columnCount == 0 {
		// Add imdb_id and poster_url columns to existing table
		_, err = d.db.Exec("ALTER TABLE activity_logs ADD COLUMN imdb_id TEXT")
		if err != nil {
			return err
		}
		_, err = d.db.Exec("ALTER TABLE activity_logs ADD COLUMN poster_url TEXT")
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// LogActivity adds a new activity log entry
func (d *Database) LogActivity(log ActivityLog) error {
	query := `
		INSERT INTO activity_logs (timestamp, job_type, media_type, title, year, tmdb_id, tvdb_id, imdb_id, poster_url, status, message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(query,
		log.Timestamp,
		log.JobType,
		log.MediaType,
		log.Title,
		log.Year,
		log.TMDBID,
		log.TVDBID,
		log.IMDBID,
		log.PosterURL,
		log.Status,
		log.Message,
	)

	return err
}

// GetRecentActivity retrieves recent activity logs
func (d *Database) GetRecentActivity(limit int) ([]ActivityLog, error) {
	query := `
		SELECT id, timestamp, job_type, media_type, title, year, tmdb_id, tvdb_id, imdb_id, poster_url, status, message
		FROM activity_logs
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []ActivityLog
	for rows.Next() {
		var log ActivityLog
		var tmdbID, tvdbID sql.NullInt64
		var message, imdbID, posterURL sql.NullString

		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&log.JobType,
			&log.MediaType,
			&log.Title,
			&log.Year,
			&tmdbID,
			&tvdbID,
			&imdbID,
			&posterURL,
			&log.Status,
			&message,
		)
		if err != nil {
			return nil, err
		}

		if tmdbID.Valid {
			log.TMDBID = int(tmdbID.Int64)
		}
		if tvdbID.Valid {
			log.TVDBID = int(tvdbID.Int64)
		}
		if imdbID.Valid {
			log.IMDBID = imdbID.String
		}
		if posterURL.Valid {
			log.PosterURL = posterURL.String
		}
		if message.Valid {
			log.Message = message.String
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetActivityStats returns statistics about activity
func (d *Database) GetActivityStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total added
	var totalAdded int
	err := d.db.QueryRow("SELECT COUNT(*) FROM activity_logs WHERE status = 'added'").Scan(&totalAdded)
	if err != nil {
		return nil, err
	}
	stats["total_added"] = totalAdded

	// Movies added
	var moviesAdded int
	err = d.db.QueryRow("SELECT COUNT(*) FROM activity_logs WHERE status = 'added' AND media_type = 'movie'").Scan(&moviesAdded)
	if err != nil {
		return nil, err
	}
	stats["movies_added"] = moviesAdded

	// Shows added
	var showsAdded int
	err = d.db.QueryRow("SELECT COUNT(*) FROM activity_logs WHERE status = 'added' AND media_type = 'show'").Scan(&showsAdded)
	if err != nil {
		return nil, err
	}
	stats["shows_added"] = showsAdded

	// Recent activity (last 24 hours)
	var recentAdded int
	err = d.db.QueryRow("SELECT COUNT(*) FROM activity_logs WHERE status = 'added' AND timestamp > datetime('now', '-24 hours')").Scan(&recentAdded)
	if err != nil {
		return nil, err
	}
	stats["added_last_24h"] = recentAdded

	return stats, nil
}

// ClearOldLogs removes logs older than the specified number of days
func (d *Database) ClearOldLogs(daysToKeep int) (int64, error) {
	query := "DELETE FROM activity_logs WHERE timestamp < datetime('now', '-' || ? || ' days')"
	result, err := d.db.Exec(query, daysToKeep)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
