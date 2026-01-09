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
	Score     float64   `json:"score,omitempty"` // Content score (0-1)
	Rank      int       `json:"rank,omitempty"`  // Rank among all items
	Status    string    `json:"status"`          // "added", "failed", "skipped"
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

	CREATE TABLE IF NOT EXISTS global_limits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		period TEXT NOT NULL,
		media_type TEXT NOT NULL,
		count INTEGER NOT NULL DEFAULT 0,
		reset_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_global_limits_period ON global_limits(period, media_type);
	CREATE INDEX IF NOT EXISTS idx_global_limits_reset ON global_limits(reset_at);
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

	// Check if score column exists
	err = d.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('activity_logs') WHERE name='score'").Scan(&columnCount)
	if err != nil {
		return err
	}

	if columnCount == 0 {
		// Add score and rank columns
		_, err = d.db.Exec("ALTER TABLE activity_logs ADD COLUMN score REAL")
		if err != nil {
			return err
		}
		_, err = d.db.Exec("ALTER TABLE activity_logs ADD COLUMN rank INTEGER")
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
		INSERT INTO activity_logs (timestamp, job_type, media_type, title, year, tmdb_id, tvdb_id, imdb_id, poster_url, score, rank, status, message)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		log.Score,
		log.Rank,
		log.Status,
		log.Message,
	)

	return err
}

// GetRecentActivity retrieves recent activity logs
func (d *Database) GetRecentActivity(limit int) ([]ActivityLog, error) {
	query := `
		SELECT id, timestamp, job_type, media_type, title, year, tmdb_id, tvdb_id, imdb_id, poster_url, score, rank, status, message
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
		var tmdbID, tvdbID, rank sql.NullInt64
		var message, imdbID, posterURL sql.NullString
		var score sql.NullFloat64

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
			&score,
			&rank,
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
		if score.Valid {
			log.Score = score.Float64
		}
		if rank.Valid {
			log.Rank = int(rank.Int64)
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

// GlobalLimit represents a limit tracking record
type GlobalLimit struct {
	ID        int64     `json:"id"`
	Period    string    `json:"period"`     // sync, daily, weekly, monthly
	MediaType string    `json:"media_type"` // movie, show
	Count     int       `json:"count"`
	ResetAt   time.Time `json:"reset_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetGlobalLimit retrieves the current limit record for a period and media type
func (d *Database) GetGlobalLimit(period string, mediaType string) (*GlobalLimit, error) {
	var limit GlobalLimit

	err := d.db.QueryRow(`
		SELECT id, period, media_type, count, reset_at, created_at, updated_at
		FROM global_limits
		WHERE period = ? AND media_type = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, period, mediaType).Scan(
		&limit.ID,
		&limit.Period,
		&limit.MediaType,
		&limit.Count,
		&limit.ResetAt,
		&limit.CreatedAt,
		&limit.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No record found
	}
	if err != nil {
		return nil, err
	}

	return &limit, nil
}

// IncrementGlobalLimit increments the count for a period and media type
// Creates a new record if one doesn't exist or if the current one has expired
func (d *Database) IncrementGlobalLimit(period string, mediaType string, resetAt time.Time) error {
	limit, err := d.GetGlobalLimit(period, mediaType)
	if err != nil {
		return err
	}

	now := time.Now()

	// If no limit exists or the current one has expired, create a new one
	if limit == nil || now.After(limit.ResetAt) {
		_, err = d.db.Exec(`
			INSERT INTO global_limits (period, media_type, count, reset_at, created_at, updated_at)
			VALUES (?, ?, 1, ?, ?, ?)
		`, period, mediaType, resetAt, now, now)
		return err
	}

	// Otherwise increment the existing count
	_, err = d.db.Exec(`
		UPDATE global_limits
		SET count = count + 1, updated_at = ?
		WHERE id = ?
	`, now, limit.ID)
	return err
}

// GetCurrentGlobalCount returns the current count for a period and media type
// Returns 0 if no record exists or if the record has expired
func (d *Database) GetCurrentGlobalCount(period string, mediaType string) (int, error) {
	limit, err := d.GetGlobalLimit(period, mediaType)
	if err != nil {
		return 0, err
	}

	if limit == nil {
		return 0, nil
	}

	// Check if limit has expired
	if time.Now().After(limit.ResetAt) {
		return 0, nil
	}

	return limit.Count, nil
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
