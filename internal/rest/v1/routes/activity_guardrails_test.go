package routes

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/global"
)

func setupActivityTestApp(t *testing.T) (*fiber.App, *database.Database) {
	t.Helper()

	db, err := database.New(t.TempDir())
	if err != nil {
		t.Fatalf("database.New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	cfg := &config.Config{}
	gctx := global.New(context.Background(), cfg, db, "test", "test", nil)

	app := fiber.New()
	RegisterActivityRoutes(app.Group("/v1"), gctx)
	return app, db
}

func TestActivityLogsRejectsInvalidStatusFilter(t *testing.T) {
	t.Parallel()

	app, _ := setupActivityTestApp(t)
	req := httptest.NewRequest("GET", "/v1/activity/logs?status=not_a_status", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestClearAllActivityHistoryRequiresConfirmationAndClearsRuns(t *testing.T) {
	t.Parallel()

	app, db := setupActivityTestApp(t)
	if err := db.LogActivity(database.ActivityLog{Timestamp: time.Now(), JobType: "test", MediaType: "movie", Title: "Test", Status: "skipped"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.StartJobRun("job-a", "Job A", "movie", "direct", time.Now()); err != nil {
		t.Fatal(err)
	}

	resp, err := app.Test(httptest.NewRequest("DELETE", "/v1/activity/logs?scope=all", nil), -1)
	if err != nil || resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("unconfirmed clear status = %d, error = %v", resp.StatusCode, err)
	}
	resp, err = app.Test(httptest.NewRequest("DELETE", "/v1/activity/logs?scope=all&confirm=CLEAR", nil), -1)
	if err != nil || resp.StatusCode != fiber.StatusOK {
		t.Fatalf("confirmed clear status = %d, error = %v", resp.StatusCode, err)
	}
	if runs, err := db.GetRecentJobRuns(10, ""); err != nil || len(runs) != 0 {
		t.Fatalf("remaining runs = %d, error = %v", len(runs), err)
	}
	stats, err := db.GetActivityStats()
	if err != nil || stats["total_skipped"] != 0 {
		t.Fatalf("activity stats after clear = %#v, error = %v", stats, err)
	}
}

func TestActivityBlockWritesUniversalTitleException(t *testing.T) {
	db, err := database.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte("version: test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{ConfigFilePath: configPath}
	gctx := global.New(context.Background(), cfg, db, "test", "test", nil)
	app := fiber.New()
	RegisterActivityRoutes(app.Group("/v1"), gctx)
	if err := db.LogActivity(database.ActivityLog{Timestamp: time.Now(), JobType: "test", MediaType: "movie", Title: "Blocked", TMDBID: 42, Status: "rejected"}); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		resp, err := app.Test(httptest.NewRequest("POST", "/v1/activity/1/block", nil), -1)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d", resp.StatusCode)
		}
	}
	current := gctx.Config()
	if got := current.TitleExceptions.BlockedMovieTMDBIDs; len(got) != 1 || got[0] != 42 {
		t.Fatalf("universal blocks = %v, want [42]", got)
	}
	if len(current.Filters.Movies.BlacklistedTMDBIds) != 0 {
		t.Fatalf("legacy filters were mutated: %v", current.Filters.Movies.BlacklistedTMDBIds)
	}
}

func TestActivityLogsRejectsInvalidMediaFilter(t *testing.T) {
	t.Parallel()

	app, _ := setupActivityTestApp(t)
	req := httptest.NewRequest("GET", "/v1/activity/logs?media=tv", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestActivityLogsRejectsInvalidRunID(t *testing.T) {
	t.Parallel()

	app, _ := setupActivityTestApp(t)
	req := httptest.NewRequest("GET", "/v1/activity/logs?run_id=abc", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func TestActivityLogsNormalizesStatusAndMediaFilters(t *testing.T) {
	t.Parallel()

	app, db := setupActivityTestApp(t)
	err := db.LogActivity(database.ActivityLog{
		Timestamp: time.Now(),
		JobType:   "test",
		MediaType: "movie",
		Title:     "Test Movie",
		Year:      2026,
		Status:    "added",
	})
	if err != nil {
		t.Fatalf("LogActivity() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/v1/activity/logs?status=AdDeD&media=MoViE&dedupe=false", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var body struct {
		Logs []struct {
			Log database.ActivityLog `json:"Log"`
		} `json:"logs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("json decode error = %v", err)
	}
	if len(body.Logs) == 0 {
		t.Fatal("expected at least one log in response")
	}
	if body.Logs[0].Log.Status != "added" || body.Logs[0].Log.MediaType != "movie" {
		t.Fatalf("unexpected first log: status=%q media=%q", body.Logs[0].Log.Status, body.Logs[0].Log.MediaType)
	}
}

func TestActivityLogsYesterdayExcludesToday(t *testing.T) {
	t.Parallel()

	app, db := setupActivityTestApp(t)
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	err := db.LogActivity(database.ActivityLog{
		Timestamp: now,
		JobType:   "test",
		MediaType: "movie",
		Title:     "Today Movie",
		Year:      2026,
		Status:    "added",
	})
	if err != nil {
		t.Fatalf("LogActivity(today) error = %v", err)
	}

	err = db.LogActivity(database.ActivityLog{
		Timestamp: yesterday,
		JobType:   "test",
		MediaType: "movie",
		Title:     "Yesterday Movie",
		Year:      2026,
		Status:    "added",
	})
	if err != nil {
		t.Fatalf("LogActivity(yesterday) error = %v", err)
	}

	req := httptest.NewRequest("GET", "/v1/activity/logs?date_range=yesterday&dedupe=false", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var body struct {
		Logs []struct {
			Log database.ActivityLog `json:"Log"`
		} `json:"logs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("json decode error = %v", err)
	}
	for _, item := range body.Logs {
		if item.Log.Title == "Today Movie" {
			t.Fatal("today log should not be included for date_range=yesterday")
		}
	}
}

func TestActivityRunsAreOrderedNewestFirstWhenStartedAtMatches(t *testing.T) {
	t.Parallel()

	app, db := setupActivityTestApp(t)
	startedAt := time.Now().Truncate(time.Second)

	firstRunID, err := db.StartJobRun("job-a", "Job A", "movie", "direct", startedAt)
	if err != nil {
		t.Fatalf("StartJobRun(first) error = %v", err)
	}
	if err := db.CompleteJobRun(firstRunID, startedAt.Add(10*time.Second), "completed", 1, 1, 1, 0, 0, 0, 0, ""); err != nil {
		t.Fatalf("CompleteJobRun(first) error = %v", err)
	}

	secondRunID, err := db.StartJobRun("job-b", "Job B", "show", "direct", startedAt)
	if err != nil {
		t.Fatalf("StartJobRun(second) error = %v", err)
	}
	if err := db.CompleteJobRun(secondRunID, startedAt.Add(20*time.Second), "completed", 2, 2, 1, 0, 0, 1, 0, ""); err != nil {
		t.Fatalf("CompleteJobRun(second) error = %v", err)
	}

	req := httptest.NewRequest("GET", "/v1/activity/runs?limit=2", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var body struct {
		Runs []database.JobRun `json:"runs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("json decode error = %v", err)
	}
	if len(body.Runs) != 2 {
		t.Fatalf("runs length = %d, want 2", len(body.Runs))
	}
	if body.Runs[0].ID != secondRunID {
		t.Fatalf("first run id = %d, want %d", body.Runs[0].ID, secondRunID)
	}
	if body.Runs[1].ID != firstRunID {
		t.Fatalf("second run id = %d, want %d", body.Runs[1].ID, firstRunID)
	}
}
