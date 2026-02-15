package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mahcks/blockbusterr/pkg/enums"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db   *sql.DB
	path string
}

type ActivityLog struct {
	ID            int64     `json:"id"`
	Timestamp     time.Time `json:"timestamp"`
	RunID         int64     `json:"run_id,omitempty"`
	JobID         string    `json:"job_id,omitempty"`
	JobType       string    `json:"job_type"`
	MediaType     string    `json:"media_type"` // "movie" or "show"
	Title         string    `json:"title"`
	Year          int       `json:"year"`
	TMDBID        int       `json:"tmdb_id,omitempty"`
	TVDBID        int       `json:"tvdb_id,omitempty"`
	IMDBID        string    `json:"imdb_id,omitempty"`
	PosterURL     string    `json:"poster_url,omitempty"`
	Score         float64   `json:"score,omitempty"` // Content score (0-1)
	Rank          int       `json:"rank,omitempty"`  // Rank among all items
	Status        string    `json:"status"`          // "added", "failed", "skipped"
	Message       string    `json:"message,omitempty"`
	FilterDetails string    `json:"filter_details,omitempty"` // JSON array of filter checks
}

type JobRun struct {
	ID           int64      `json:"id"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	DurationMs   int64      `json:"duration_ms"`
	JobID        string     `json:"job_id,omitempty"`
	JobName      string     `json:"job_name"`
	MediaType    string     `json:"media_type"` // "movie" or "show"
	Mode         string     `json:"mode"`       // "direct" or "jellyseerr"
	Status       string     `json:"status"`     // "running", "completed", "failed"
	TotalFound   int        `json:"total_found"`
	PassedFilter int        `json:"passed_filters"`
	Added        int        `json:"added"`
	Requested    int        `json:"requested"`
	Skipped      int        `json:"skipped"`
	Rejected     int        `json:"rejected"`
	Failed       int        `json:"failed"`
	ErrorMessage string     `json:"error_message,omitempty"`
}

type ActivityDailyCount struct {
	Day      string `json:"day"` // YYYY-MM-DD
	Added    int    `json:"added"`
	Rejected int    `json:"rejected"`
	Skipped  int    `json:"skipped"`
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

	database := &Database{db: db, path: dbPath}

	// Initialize schema
	if err := database.initSchema(); err != nil {
		schemaErr := err
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("failed to initialize schema (%v) and close db: %w", schemaErr, closeErr)
		}
		return nil, fmt.Errorf("failed to initialize schema: %w", schemaErr)
	}

	return database, nil
}

func (d *Database) initSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS activity_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			run_id INTEGER,
			job_id TEXT,
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

	CREATE TABLE IF NOT EXISTS job_runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		started_at DATETIME NOT NULL,
		finished_at DATETIME,
		duration_ms INTEGER NOT NULL DEFAULT 0,
		job_id TEXT,
		job_name TEXT NOT NULL,
		media_type TEXT NOT NULL,
		mode TEXT,
		status TEXT NOT NULL DEFAULT 'running',
		total_found INTEGER NOT NULL DEFAULT 0,
		passed_filters INTEGER NOT NULL DEFAULT 0,
		added INTEGER NOT NULL DEFAULT 0,
		requested INTEGER NOT NULL DEFAULT 0,
		skipped INTEGER NOT NULL DEFAULT 0,
		rejected INTEGER NOT NULL DEFAULT 0,
		failed INTEGER NOT NULL DEFAULT 0,
		error_message TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_job_runs_started_at ON job_runs(started_at DESC);
	CREATE INDEX IF NOT EXISTS idx_job_runs_job_id ON job_runs(job_id);
	CREATE INDEX IF NOT EXISTS idx_job_runs_status ON job_runs(status);

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

	// Check if filter_details column exists
	err = d.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('activity_logs') WHERE name='filter_details'").Scan(&columnCount)
	if err != nil {
		return err
	}

	if columnCount == 0 {
		// Add filter_details column
		_, err = d.db.Exec("ALTER TABLE activity_logs ADD COLUMN filter_details TEXT")
		if err != nil {
			return err
		}
	}

	// Ensure newer activity columns exist (robust against partially migrated DBs)
	if err := d.ensureActivityLogsColumn("run_id INTEGER"); err != nil {
		return err
	}
	if err := d.ensureActivityLogsColumn("job_id TEXT"); err != nil {
		return err
	}

	// Ensure job_id index exists for filtering/grouping performance
	_, err = d.db.Exec("CREATE INDEX IF NOT EXISTS idx_activity_job_id ON activity_logs(job_id)")
	if err != nil {
		return err
	}
	_, err = d.db.Exec("CREATE INDEX IF NOT EXISTS idx_activity_run_id ON activity_logs(run_id)")
	if err != nil {
		return err
	}

	// Ensure job_runs table/indexes exist for run timeline data
	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS job_runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			started_at DATETIME NOT NULL,
			finished_at DATETIME,
			duration_ms INTEGER NOT NULL DEFAULT 0,
			job_id TEXT,
			job_name TEXT NOT NULL,
			media_type TEXT NOT NULL,
			mode TEXT,
			status TEXT NOT NULL DEFAULT 'running',
			total_found INTEGER NOT NULL DEFAULT 0,
			passed_filters INTEGER NOT NULL DEFAULT 0,
			added INTEGER NOT NULL DEFAULT 0,
			requested INTEGER NOT NULL DEFAULT 0,
			skipped INTEGER NOT NULL DEFAULT 0,
			rejected INTEGER NOT NULL DEFAULT 0,
			failed INTEGER NOT NULL DEFAULT 0,
			error_message TEXT
		);
	`)
	if err != nil {
		return err
	}
	_, err = d.db.Exec("CREATE INDEX IF NOT EXISTS idx_job_runs_started_at ON job_runs(started_at DESC)")
	if err != nil {
		return err
	}
	_, err = d.db.Exec("CREATE INDEX IF NOT EXISTS idx_job_runs_job_id ON job_runs(job_id)")
	if err != nil {
		return err
	}
	_, err = d.db.Exec("CREATE INDEX IF NOT EXISTS idx_job_runs_status ON job_runs(status)")
	if err != nil {
		return err
	}

	return nil
}

func (d *Database) ensureActivityLogsColumn(definition string) error {
	_, err := d.db.Exec("ALTER TABLE activity_logs ADD COLUMN " + definition)
	if err != nil {
		// SQLite returns "duplicate column name: <name>" when it already exists.
		if strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
			return nil
		}
		return err
	}
	return nil
}

func (d *Database) ensureActivityIdentityColumns() error {
	if err := d.ensureActivityLogsColumn("run_id INTEGER"); err != nil {
		return err
	}
	if err := d.ensureActivityLogsColumn("job_id TEXT"); err != nil {
		return err
	}
	_, err := d.db.Exec("CREATE INDEX IF NOT EXISTS idx_activity_job_id ON activity_logs(job_id)")
	if err != nil {
		return err
	}
	_, err = d.db.Exec("CREATE INDEX IF NOT EXISTS idx_activity_run_id ON activity_logs(run_id)")
	if err != nil {
		return err
	}
	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) Path() string {
	return d.path
}

func (d *Database) GetActivityDebug() (map[string]any, error) {
	result := map[string]any{
		"db_path": d.path,
	}

	var activityCount int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM activity_logs").Scan(&activityCount); err != nil {
		return nil, err
	}
	result["activity_count"] = activityCount

	var hasJobRuns int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='job_runs'").Scan(&hasJobRuns); err != nil {
		return nil, err
	}
	result["has_job_runs"] = hasJobRuns > 0

	if hasJobRuns > 0 {
		var jobRunsCount int
		if err := d.db.QueryRow("SELECT COUNT(*) FROM job_runs").Scan(&jobRunsCount); err != nil {
			return nil, err
		}
		result["job_runs_count"] = jobRunsCount
	}

	var hasRunID int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('activity_logs') WHERE name='run_id'").Scan(&hasRunID); err != nil {
		return nil, err
	}
	result["has_run_id_column"] = hasRunID > 0

	var hasJobID int
	if err := d.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('activity_logs') WHERE name='job_id'").Scan(&hasJobID); err != nil {
		return nil, err
	}
	result["has_job_id_column"] = hasJobID > 0

	return result, nil
}

// LogActivity adds a new activity log entry
func (d *Database) LogActivity(log ActivityLog) error {
	if err := d.ensureActivityIdentityColumns(); err != nil {
		return err
	}

	query := `
		INSERT INTO activity_logs (timestamp, run_id, job_id, job_type, media_type, title, year, tmdb_id, tvdb_id, imdb_id, poster_url, score, rank, status, message, filter_details)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(query, log.Timestamp, log.RunID, log.JobID, log.JobType, log.MediaType, log.Title, log.Year,
		log.TMDBID, log.TVDBID, log.IMDBID, log.PosterURL, log.Score, log.Rank, log.Status, log.Message, log.FilterDetails)
	return err
}

// GetRecentActivity retrieves recent activity logs
func (d *Database) GetRecentActivity(limit int) ([]ActivityLog, error) {
	if err := d.ensureActivityIdentityColumns(); err != nil {
		return nil, err
	}

	query := `
		SELECT id, timestamp, run_id, job_id, job_type, media_type, title, year, tmdb_id, tvdb_id, imdb_id, poster_url, score, rank, status, message, filter_details
		FROM activity_logs
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var logs []ActivityLog
	for rows.Next() {
		var log ActivityLog
		var runID, tmdbID, tvdbID, rank sql.NullInt64
		var message, imdbID, jobID, posterURL, filterDetails sql.NullString
		var score sql.NullFloat64

		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&runID,
			&jobID,
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
			&filterDetails,
		)
		if err != nil {
			return nil, err
		}

		if tmdbID.Valid {
			log.TMDBID = int(tmdbID.Int64)
		}
		if runID.Valid {
			log.RunID = runID.Int64
		}
		if tvdbID.Valid {
			log.TVDBID = int(tvdbID.Int64)
		}
		if imdbID.Valid {
			log.IMDBID = imdbID.String
		}
		if jobID.Valid {
			log.JobID = jobID.String
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
		if filterDetails.Valid {
			log.FilterDetails = filterDetails.String
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetRecentActivityFiltered retrieves recent activity logs with optional filters
func (d *Database) GetRecentActivityFiltered(limit int, status, mediaType, jobType string) ([]ActivityLog, error) {
	if err := d.ensureActivityIdentityColumns(); err != nil {
		return nil, err
	}

	query := `
		SELECT id, timestamp, run_id, job_id, job_type, media_type, title, year, tmdb_id, tvdb_id, imdb_id, poster_url, score, rank, status, message, filter_details
		FROM activity_logs
		WHERE 1=1
	`
	args := []any{}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if mediaType != "" {
		query += " AND media_type = ?"
		args = append(args, mediaType)
	}
	if jobType != "" {
		query += " AND job_type = ?"
		args = append(args, jobType)
	}

	query += " ORDER BY timestamp DESC LIMIT ?"
	args = append(args, limit)

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var logs []ActivityLog
	for rows.Next() {
		var log ActivityLog
		var runID, tmdbID, tvdbID, rank sql.NullInt64
		var message, imdbID, jobID, posterURL, filterDetails sql.NullString
		var score sql.NullFloat64

		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&runID,
			&jobID,
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
			&filterDetails,
		)
		if err != nil {
			return nil, err
		}

		if tmdbID.Valid {
			log.TMDBID = int(tmdbID.Int64)
		}
		if runID.Valid {
			log.RunID = runID.Int64
		}
		if tvdbID.Valid {
			log.TVDBID = int(tvdbID.Int64)
		}
		if imdbID.Valid {
			log.IMDBID = imdbID.String
		}
		if jobID.Valid {
			log.JobID = jobID.String
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
		if filterDetails.Valid {
			log.FilterDetails = filterDetails.String
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetActivityStats returns statistics about activity
func (d *Database) GetActivityStats() (map[string]any, error) {
	stats := make(map[string]any)

	// Total added
	var totalAdded int
	err := d.db.QueryRow(
		"SELECT COUNT(*) FROM activity_logs WHERE status IN (?, ?)",
		enums.ActivityStatusAdded,
		enums.ActivityStatusRequested,
	).Scan(&totalAdded)
	if err != nil {
		return nil, err
	}
	stats["total_added"] = totalAdded

	// Total failed
	var totalFailed int
	err = d.db.QueryRow("SELECT COUNT(*) FROM activity_logs WHERE status = ?", enums.ActivityStatusFailed).Scan(&totalFailed)
	if err != nil {
		return nil, err
	}
	stats["total_failed"] = totalFailed

	// Total rejected by filters
	var totalRejected int
	err = d.db.QueryRow("SELECT COUNT(*) FROM activity_logs WHERE status = ?", enums.ActivityStatusRejected).Scan(&totalRejected)
	if err != nil {
		return nil, err
	}
	stats["total_rejected"] = totalRejected

	// Total skipped
	var totalSkipped int
	err = d.db.QueryRow("SELECT COUNT(*) FROM activity_logs WHERE status = ?", enums.ActivityStatusSkipped).Scan(&totalSkipped)
	if err != nil {
		return nil, err
	}
	stats["total_skipped"] = totalSkipped

	// Total movies
	var totalMovies int
	err = d.db.QueryRow("SELECT COUNT(*) FROM activity_logs WHERE media_type = ?", enums.MediaTypeMovie).Scan(&totalMovies)
	if err != nil {
		return nil, err
	}
	stats["total_movies"] = totalMovies

	// Total shows
	var totalShows int
	err = d.db.QueryRow("SELECT COUNT(*) FROM activity_logs WHERE media_type = ?", enums.MediaTypeShow).Scan(&totalShows)
	if err != nil {
		return nil, err
	}
	stats["total_shows"] = totalShows

	// Recent activity (last 24 hours)
	var recentAdded int
	err = d.db.QueryRow(
		"SELECT COUNT(*) FROM activity_logs WHERE status IN (?, ?) AND timestamp > datetime('now', '-24 hours')",
		enums.ActivityStatusAdded,
		enums.ActivityStatusRequested,
	).Scan(&recentAdded)
	if err != nil {
		return nil, err
	}
	stats["added_last_24h"] = recentAdded

	return stats, nil
}

// GetActivityDailyCounts returns per-day activity counters for the given lookback window.
// The result uses calendar-day keys in YYYY-MM-DD format.
func (d *Database) GetActivityDailyCounts(days int) (map[string]ActivityDailyCount, error) {
	if days <= 0 {
		days = 7
	}

	start := time.Now().AddDate(0, 0, -(days - 1))
	startDay := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())

	query := `
		SELECT
			date(timestamp) AS day,
			SUM(CASE WHEN status IN (?, ?) THEN 1 ELSE 0 END) AS added,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS rejected,
			SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS skipped
		FROM activity_logs
		WHERE timestamp >= ?
		GROUP BY day
		ORDER BY day ASC
	`

	rows, err := d.db.Query(
		query,
		enums.ActivityStatusAdded,
		enums.ActivityStatusRequested,
		enums.ActivityStatusRejected,
		enums.ActivityStatusSkipped,
		startDay,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make(map[string]ActivityDailyCount)
	for rows.Next() {
		var day string
		var added, rejected, skipped int
		if err := rows.Scan(&day, &added, &rejected, &skipped); err != nil {
			return nil, err
		}
		result[day] = ActivityDailyCount{
			Day:      day,
			Added:    added,
			Rejected: rejected,
			Skipped:  skipped,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
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

// GetActivityLogByID retrieves a single activity log entry by ID
func (d *Database) GetActivityLogByID(id int64) (*ActivityLog, error) {
	if err := d.ensureActivityIdentityColumns(); err != nil {
		return nil, err
	}

	query := `
		SELECT id, timestamp, run_id, job_id, job_type, media_type, title, year, tmdb_id, tvdb_id, imdb_id, poster_url, score, rank, status, message, filter_details
		FROM activity_logs
		WHERE id = ?
	`

	var log ActivityLog
	var runID, tmdbID, tvdbID, rank sql.NullInt64
	var message, imdbID, jobID, posterURL, filterDetails sql.NullString
	var score sql.NullFloat64

	err := d.db.QueryRow(query, id).Scan(
		&log.ID,
		&log.Timestamp,
		&runID,
		&jobID,
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
		&filterDetails,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if tmdbID.Valid {
		log.TMDBID = int(tmdbID.Int64)
	}
	if runID.Valid {
		log.RunID = runID.Int64
	}
	if tvdbID.Valid {
		log.TVDBID = int(tvdbID.Int64)
	}
	if imdbID.Valid {
		log.IMDBID = imdbID.String
	}
	if jobID.Valid {
		log.JobID = jobID.String
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
	if filterDetails.Valid {
		log.FilterDetails = filterDetails.String
	}

	return &log, nil
}

// StartJobRun inserts a new running job run and returns its ID.
func (d *Database) StartJobRun(jobID, jobName, mediaType, mode string, startedAt time.Time) (int64, error) {
	query := `
		INSERT INTO job_runs (started_at, job_id, job_name, media_type, mode, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := d.db.Exec(query, startedAt, jobID, jobName, mediaType, mode, enums.JobRunStatusRunning)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// CompleteJobRun marks a job run as completed/failed with final counters.
func (d *Database) CompleteJobRun(
	runID int64,
	finishedAt time.Time,
	status string,
	totalFound, passedFilters, added, requested, skipped, rejected, failed int,
	errorMessage string,
) error {
	if _, ok := enums.ParseJobRunStatus(status); !ok {
		return fmt.Errorf("invalid job run status: %s", status)
	}

	query := `
		UPDATE job_runs
		SET finished_at = ?,
			duration_ms = ?,
			status = ?,
			total_found = ?,
			passed_filters = ?,
			added = ?,
			requested = ?,
			skipped = ?,
			rejected = ?,
			failed = ?,
			error_message = ?
		WHERE id = ?
	`

	var durationMs int64
	var startedAt time.Time
	err := d.db.QueryRow("SELECT started_at FROM job_runs WHERE id = ?", runID).Scan(&startedAt)
	if err == nil {
		durationMs = finishedAt.Sub(startedAt).Milliseconds()
		durationMs = max(durationMs, 0)
	}

	_, err = d.db.Exec(
		query,
		finishedAt,
		durationMs,
		status,
		totalFound,
		passedFilters,
		added,
		requested,
		skipped,
		rejected,
		failed,
		errorMessage,
		runID,
	)
	return err
}

// GetRecentJobRuns returns recent job runs, optionally filtered by job ID.
func (d *Database) GetRecentJobRuns(limit int, jobID string) ([]JobRun, error) {
	query := `
		SELECT id, started_at, finished_at, duration_ms, job_id, job_name, media_type, mode, status,
			total_found, passed_filters, added, requested, skipped, rejected, failed, error_message
		FROM job_runs
	`
	args := []any{}

	if jobID != "" {
		query += " WHERE job_id = ?"
		args = append(args, jobID)
	}

	query += " ORDER BY started_at DESC, id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	runs := make([]JobRun, 0)
	for rows.Next() {
		var run JobRun
		var finishedAt sql.NullTime
		var jobIDVal, mode, errorMessage sql.NullString
		err := rows.Scan(
			&run.ID,
			&run.StartedAt,
			&finishedAt,
			&run.DurationMs,
			&jobIDVal,
			&run.JobName,
			&run.MediaType,
			&mode,
			&run.Status,
			&run.TotalFound,
			&run.PassedFilter,
			&run.Added,
			&run.Requested,
			&run.Skipped,
			&run.Rejected,
			&run.Failed,
			&errorMessage,
		)
		if err != nil {
			return nil, err
		}

		if finishedAt.Valid {
			t := finishedAt.Time
			run.FinishedAt = &t
		}
		if jobIDVal.Valid {
			run.JobID = jobIDVal.String
		}
		if mode.Valid {
			run.Mode = mode.String
		}
		if errorMessage.Valid {
			run.ErrorMessage = errorMessage.String
		}
		runs = append(runs, run)
	}

	return runs, rows.Err()
}

// UpdateActivityLogStatus updates the status and message of an activity log entry
func (d *Database) UpdateActivityLogStatus(id int64, status, message string) error {
	if _, ok := enums.ParseActivityStatus(status); !ok {
		return fmt.Errorf("invalid activity status: %s", status)
	}

	query := `
		UPDATE activity_logs
		SET status = ?, message = ?
		WHERE id = ?
	`

	_, err := d.db.Exec(query, status, message, id)
	return err
}
