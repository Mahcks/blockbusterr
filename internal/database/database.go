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
	Source        string    `json:"source,omitempty"`
	MediaType     string    `json:"media_type"` // "movie" or "show"
	Title         string    `json:"title"`
	Language      string    `json:"language,omitempty"`
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

type DeliveryBudgetUsage struct {
	Movies int
	Shows  int
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

	CREATE TABLE IF NOT EXISTS delivery_budget_reservations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		reserved_at DATETIME NOT NULL,
		run_id INTEGER NOT NULL,
		job_id TEXT NOT NULL,
		media_type TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_delivery_budget_media_time ON delivery_budget_reservations(media_type, reserved_at);
	CREATE INDEX IF NOT EXISTS idx_delivery_budget_run ON delivery_budget_reservations(run_id);

	CREATE TABLE IF NOT EXISTS selection_cycles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		started_at DATETIME NOT NULL,
		finished_at DATETIME,
		status TEXT NOT NULL,
		movie_winners INTEGER NOT NULL DEFAULT 0,
		show_winners INTEGER NOT NULL DEFAULT 0,
		error_message TEXT
	);

	CREATE TABLE IF NOT EXISTS selection_cycle_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cycle_id INTEGER NOT NULL,
		media_key TEXT NOT NULL,
		job_id TEXT NOT NULL,
		job_ids TEXT NOT NULL,
		sources TEXT NOT NULL,
		score REAL NOT NULL,
		rank INTEGER NOT NULL,
		reason TEXT NOT NULL,
		FOREIGN KEY(cycle_id) REFERENCES selection_cycles(id)
	);

	CREATE INDEX IF NOT EXISTS idx_selection_cycle_items_cycle ON selection_cycle_items(cycle_id);

	`

	_, err := d.db.Exec(schema)
	if err != nil {
		return err
	}
	if _, err = d.db.Exec("DELETE FROM delivery_budget_reservations WHERE reserved_at < datetime('now', '-31 days')"); err != nil {
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
	if err := d.ensureActivityLogsColumn("language TEXT"); err != nil {
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

// TryReserveDelivery atomically claims one delivery slot. Zero means unlimited.
// The caller releases the reservation when the downstream delivery does not succeed.
func (d *Database) TryReserveDelivery(runID int64, jobID, mediaType string, perRunLimit, globalLimit int, since time.Time) (int64, string, error) {
	if perRunLimit <= 0 && globalLimit <= 0 {
		return 0, "", nil
	}

	tx, err := d.db.Begin()
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = tx.Rollback() }()

	if perRunLimit > 0 {
		var count int
		if err := tx.QueryRow("SELECT COUNT(*) FROM delivery_budget_reservations WHERE run_id = ?", runID).Scan(&count); err != nil {
			return 0, "", err
		}
		if count >= perRunLimit {
			return 0, "Job delivery limit reached", nil
		}
	}

	if globalLimit > 0 {
		var count int
		if err := tx.QueryRow("SELECT COUNT(*) FROM delivery_budget_reservations WHERE media_type = ? AND reserved_at >= ?", mediaType, since).Scan(&count); err != nil {
			return 0, "", err
		}
		if count >= globalLimit {
			return 0, "Global delivery limit reached", nil
		}
	}

	result, err := tx.Exec(
		"INSERT INTO delivery_budget_reservations (reserved_at, run_id, job_id, media_type) VALUES (?, ?, ?, ?)",
		time.Now(), runID, jobID, mediaType,
	)
	if err != nil {
		return 0, "", err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, "", err
	}
	if err := tx.Commit(); err != nil {
		return 0, "", err
	}
	return id, "", nil
}

func (d *Database) ReleaseDeliveryReservation(id int64) error {
	if id == 0 {
		return nil
	}
	_, err := d.db.Exec("DELETE FROM delivery_budget_reservations WHERE id = ?", id)
	return err
}

func (d *Database) CountDeliveriesSince(mediaType string, since time.Time) (int, error) {
	var count int
	err := d.db.QueryRow(
		"SELECT COUNT(*) FROM delivery_budget_reservations WHERE media_type = ? AND reserved_at >= ?",
		mediaType, since,
	).Scan(&count)
	return count, err
}

func (d *Database) GetDeliveryBudgetUsage(period string) (DeliveryBudgetUsage, error) {
	window := 24 * time.Hour
	if period == "weekly" {
		window = 7 * 24 * time.Hour
	} else if period == "monthly" {
		window = 30 * 24 * time.Hour
	}
	since := time.Now().Add(-window)
	movies, err := d.CountDeliveriesSince("movie", since)
	if err != nil {
		return DeliveryBudgetUsage{}, err
	}
	shows, err := d.CountDeliveriesSince("show", since)
	return DeliveryBudgetUsage{Movies: movies, Shows: shows}, err
}

func (d *Database) Close() error {
	return d.db.Close()
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
		INSERT INTO activity_logs (timestamp, run_id, job_id, job_type, media_type, title, language, year, tmdb_id, tvdb_id, imdb_id, poster_url, score, rank, status, message, filter_details)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(query, log.Timestamp, log.RunID, log.JobID, log.JobType, log.MediaType, log.Title, log.Language, log.Year,
		log.TMDBID, log.TVDBID, log.IMDBID, log.PosterURL, log.Score, log.Rank, log.Status, log.Message, log.FilterDetails)
	return err
}

// LatestSuccessfulDelivery returns the most recent time Blockbusterr delivered a title.
func (d *Database) LatestSuccessfulDelivery(mediaType string, tmdbID, tvdbID int) (time.Time, bool, error) {
	column, id := "tmdb_id", tmdbID
	if mediaType == string(enums.MediaTypeShow) {
		column, id = "tvdb_id", tvdbID
	}
	if id <= 0 {
		return time.Time{}, false, nil
	}
	var deliveredAt time.Time
	err := d.db.QueryRow(`SELECT timestamp FROM activity_logs WHERE media_type = ? AND `+column+` = ? AND status IN (?, ?) AND COALESCE(message, '') NOT LIKE '[DRY RUN]%' ORDER BY timestamp DESC LIMIT 1`,
		mediaType, id, enums.ActivityStatusAdded, enums.ActivityStatusRequested).Scan(&deliveredAt)
	if err == sql.ErrNoRows {
		return time.Time{}, false, nil
	}
	return deliveredAt, err == nil, err
}

// GetRecentActivityFiltered retrieves recent activity logs with optional filters
func (d *Database) GetRecentActivityFiltered(limit int, status, mediaType, jobType, language string) ([]ActivityLog, error) {
	if err := d.ensureActivityIdentityColumns(); err != nil {
		return nil, err
	}

	query := `
		SELECT id, timestamp, run_id, job_id, job_type, media_type, title, language, year, tmdb_id, tvdb_id, imdb_id, poster_url, score, rank, status, message, filter_details
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
	if language != "" {
		query += " AND language = ?"
		args = append(args, language)
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
		var message, imdbID, jobID, posterURL, filterDetails, languageValue sql.NullString
		var score sql.NullFloat64

		err := rows.Scan(
			&log.ID,
			&log.Timestamp,
			&runID,
			&jobID,
			&log.JobType,
			&log.MediaType,
			&log.Title,
			&languageValue,
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
		if languageValue.Valid {
			log.Language = languageValue.String
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

func (d *Database) GetActivityLanguages() ([]string, error) {
	rows, err := d.db.Query("SELECT DISTINCT language FROM activity_logs WHERE language IS NOT NULL AND language != '' ORDER BY language")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	languages := []string{}
	for rows.Next() {
		var language string
		if err := rows.Scan(&language); err != nil {
			return nil, err
		}
		languages = append(languages, language)
	}
	return languages, rows.Err()
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

// ClearOldLogs removes logs older than the specified number of days
func (d *Database) ClearOldLogs(daysToKeep int) (int64, error) {
	query := "DELETE FROM activity_logs WHERE timestamp < datetime('now', '-' || ? || ' days')"
	result, err := d.db.Exec(query, daysToKeep)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// ClearActivityHistory removes Activity Entries, Job Runs, and ranked-selection history.
func (d *Database) ClearActivityHistory() (int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var count int64
	for _, table := range []string{"activity_logs", "job_runs", "selection_cycle_items", "selection_cycles"} {
		result, err := tx.Exec("DELETE FROM " + table)
		if err != nil {
			return 0, err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		count += rows
	}
	return count, tx.Commit()
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
