package database

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/mahcks/blockbusterr/pkg/enums"
)

func TestDeliveryMemorySurvivesActivityDeletion(t *testing.T) {
	db, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Now().Add(-time.Hour)
	if err := db.LogActivity(ActivityLog{Timestamp: now, MediaType: "movie", Title: "Delivered", TMDBID: 42, Status: "added"}); err != nil {
		t.Fatal(err)
	}
	if err := db.LogActivity(ActivityLog{Timestamp: now, MediaType: "movie", Title: "Preview", TMDBID: 43, Status: "added", Message: "[DRY RUN] Would add"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClearActivityHistory(); err != nil {
		t.Fatal(err)
	}
	if _, found, err := db.LatestSuccessfulDelivery("movie", 42, 0); err != nil || !found {
		t.Fatalf("delivery memory found=%v err=%v", found, err)
	}
	if _, found, err := db.LatestSuccessfulDelivery("movie", 43, 0); err != nil || found {
		t.Fatalf("dry-run memory found=%v err=%v", found, err)
	}
	if _, err := db.ClearActivityHistory(true); err != nil {
		t.Fatal(err)
	}
	if _, found, err := db.LatestSuccessfulDelivery("movie", 42, 0); err != nil || found {
		t.Fatalf("cleared memory found=%v err=%v", found, err)
	}
}

func TestDeliveryMemoryBackfillRunsOnce(t *testing.T) {
	dir := t.TempDir()
	raw, err := sql.Open("sqlite3", filepath.Join(dir, "blockbusterr.db"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = raw.Exec(`CREATE TABLE activity_logs (
		id INTEGER PRIMARY KEY, timestamp DATETIME, job_type TEXT NOT NULL, media_type TEXT NOT NULL,
		title TEXT NOT NULL, year INTEGER, tmdb_id INTEGER, tvdb_id INTEGER, imdb_id TEXT,
		score REAL, status TEXT NOT NULL, message TEXT
	); INSERT INTO activity_logs (timestamp, job_type, media_type, title, tmdb_id, status) VALUES (?, 'legacy', 'movie', 'Legacy', 99, 'added')`, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}

	db, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, column := range []string{"run_id", "job_id", "poster_url", "language", "rank", "filter_details"} {
		var exists int
		if err := db.db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('activity_logs') WHERE name = ?", column).Scan(&exists); err != nil || exists != 1 {
			t.Fatalf("column %s exists=%d err=%v", column, exists, err)
		}
	}
	if _, found, err := db.LatestSuccessfulDelivery("movie", 99, 0); err != nil || !found {
		t.Fatalf("backfill found=%v err=%v", found, err)
	}
	if _, err := db.ClearActivityHistory(true); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = New(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, found, err := db.LatestSuccessfulDelivery("movie", 99, 0); err != nil || found {
		t.Fatalf("cleared memory restored on restart: found=%v err=%v", found, err)
	}
}

func TestActivityQueryHasNoHistoryCapAndDedupesBeforePagination(t *testing.T) {
	db, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.db.Exec(`WITH RECURSIVE rows(value) AS (
		VALUES(1) UNION ALL SELECT value + 1 FROM rows WHERE value < 50001
	) INSERT INTO activity_logs (timestamp, run_id, job_id, job_type, media_type, title, language, year, tmdb_id, status)
	SELECT datetime('now', '-' || value || ' minutes'), value, 'noise', 'Noise', 'movie', 'Noise ' || value, 'en', 2026, value, 'rejected' FROM rows`)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := db.LogActivity(ActivityLog{Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), RunID: 42, JobID: "target", JobType: "Target", MediaType: "movie", Title: "Ancient Needle", Language: "fr", Year: 1999, TMDBID: 60000, Status: "added"}); err != nil {
			t.Fatal(err)
		}
	}
	runID := int64(42)
	page, err := db.QueryActivity(ActivityQuery{Status: "added", MediaType: "movie", Job: "target", Language: "fr", Search: "needle", RunID: &runID, Dedupe: true, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Groups) != 1 || page.Groups[0].Count != 2 || len(page.Groups[0].History) != 2 {
		t.Fatalf("page=%+v", page)
	}
}

func TestLatestSuccessfulDeliveriesBatchesPreviewSizedLists(t *testing.T) {
	db, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.db.Exec(`WITH RECURSIVE rows(value) AS (
		VALUES(1) UNION ALL SELECT value + 1 FROM rows WHERE value < 500
	) INSERT INTO delivery_memory (media_key, media_type, tmdb_id, delivered_at, outcome)
	SELECT 'movie:tmdb:' || value, 'movie', value, datetime('now'), 'added' FROM rows`)
	if err != nil {
		t.Fatal(err)
	}
	identities := make([]DeliveryIdentity, 500)
	for index := range identities {
		identities[index] = DeliveryIdentity{MediaType: "movie", TMDBID: index + 1}
	}
	deliveries, err := db.LatestSuccessfulDeliveries(identities)
	if err != nil || len(deliveries) != 500 {
		t.Fatalf("deliveries=%d err=%v", len(deliveries), err)
	}
}

func TestRetentionPreservesRunningHistoryAndDeliveryMemory(t *testing.T) {
	db, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	old := time.Now().AddDate(0, 0, -100)
	if err := db.LogActivity(ActivityLog{Timestamp: old, MediaType: "movie", Title: "Delivered", TMDBID: 7, Status: "added"}); err != nil {
		t.Fatal(err)
	}
	completedRun, err := db.StartJobRun("old", "Old", "movie", "direct", old)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CompleteJobRun(completedRun, old.Add(time.Minute), string(enums.JobRunStatusCompleted), 0, 0, 0, 0, 0, 0, 0, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := db.StartJobRun("running", "Running", "movie", "direct", old); err != nil {
		t.Fatal(err)
	}
	completedCycle, err := db.StartSelectionCycle(old)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveSelectionCycleItems(completedCycle, []SelectionCycleItem{{MediaKey: "movie:tmdb:7", JobID: "old", JobIDs: []string{"old"}, Sources: []string{"tmdb"}, Reason: enums.SelectionReasonWinner}}); err != nil {
		t.Fatal(err)
	}
	if err := db.CompleteSelectionCycle(completedCycle, enums.SelectionCycleCompleted, 1, 0, 1, 0, 0, ""); err != nil {
		t.Fatal(err)
	}
	runningCycle, err := db.StartSelectionCycle(old)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveSelectionCycleItems(runningCycle, []SelectionCycleItem{{MediaKey: "movie:tmdb:8", JobID: "running", JobIDs: []string{"running"}, Sources: []string{"tmdb"}, Reason: enums.SelectionReasonWinner}}); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ClearOldLogs(90); err != nil {
		t.Fatal(err)
	}
	if runs, err := db.GetRecentJobRuns(10, ""); err != nil || len(runs) != 1 || runs[0].Status != "running" {
		t.Fatalf("runs=%+v err=%v", runs, err)
	}
	if cycles, err := db.GetRecentSelectionCycles(10); err != nil || len(cycles) != 1 || cycles[0].Status != enums.SelectionCycleRunning {
		t.Fatalf("cycles=%+v err=%v", cycles, err)
	}
	var runningItems int
	if err := db.db.QueryRow("SELECT COUNT(*) FROM selection_cycle_items WHERE cycle_id = ?", runningCycle).Scan(&runningItems); err != nil || runningItems != 1 {
		t.Fatalf("running items=%d err=%v", runningItems, err)
	}
	if _, found, err := db.LatestSuccessfulDelivery("movie", 7, 0); err != nil || !found {
		t.Fatalf("memory found=%v err=%v", found, err)
	}
}
