package routes

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/config"
)

func newConfigSaveTestApp(t *testing.T) (*fiber.App, *config.Config) {
	t.Helper()
	cfg := &config.Config{}
	cfg.ConfigFilePath = filepath.Join(t.TempDir(), "config.yaml")
	cfg.Jobs.SyncInterval = "1h"

	app := fiber.New()
	rg := NewRouteGroup(dynamicJobsTestContext{cfg: cfg})
	RegisterUIRoutes(rg, app)
	return app, cfg
}

func postConfigSave(t *testing.T, app *fiber.App, fields map[string]string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write field %s: %v", key, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest("POST", "/config/save", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	recorder := httptest.NewRecorder()
	recorder.Code = resp.StatusCode

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return recorder, payload
}

func TestConfigSaveRejectsInvalidScoringTotal(t *testing.T) {
	app, cfg := newConfigSaveTestApp(t)

	rec, payload := postConfigSave(t, app, map[string]string{
		"jobs.sync_interval":        "1h",
		"scoring.enabled":           "true",
		"scoring.rating_weight":     "0.6",
		"scoring.popularity_weight": "0.6",
		"scoring.recency_weight":    "0.1",
	})

	if rec.Code != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	errMsg, _ := payload["error"].(string)
	if errMsg == "" {
		t.Fatal("expected a structured error message")
	}

	// The live config must be untouched by a rejected save.
	if cfg.Scoring.Enabled {
		t.Error("live config was mutated despite validation failure")
	}
}

func TestConfigSaveAcceptsValidSubmission(t *testing.T) {
	app, cfg := newConfigSaveTestApp(t)

	rec, payload := postConfigSave(t, app, map[string]string{
		"jobs.sync_interval":        "2h",
		"jobs.mode":                 "direct",
		"radarr.url":                "http://localhost:7878",
		"radarr.api_key":            "secret",
		"jobs.global_limit_movies":  "12",
		"jobs.global_limit_shows":   "8",
		"jobs.global_period":        "weekly",
		"scoring.enabled":           "true",
		"scoring.rating_weight":     "0.6",
		"scoring.popularity_weight": "0.3",
		"scoring.recency_weight":    "0.1",
	})

	if rec.Code != fiber.StatusOK {
		t.Fatalf("expected 200, got %d: %v", rec.Code, payload)
	}
	if payload["success"] != true {
		t.Errorf("expected success=true, got %v", payload)
	}
	if cfg.Jobs.SyncInterval != "2h" {
		t.Errorf("sync interval not applied: %q", cfg.Jobs.SyncInterval)
	}
	if cfg.Radarr.URL != "http://localhost:7878" {
		t.Errorf("radarr url not applied: %q", cfg.Radarr.URL)
	}
	if cfg.Jobs.GlobalLimitMovies != 12 || cfg.Jobs.GlobalLimitShows != 8 || cfg.Jobs.GlobalPeriod != "weekly" {
		t.Errorf("delivery limits not applied: %+v", cfg.Jobs)
	}
}

func TestConfigSavePreservesBlankSecretsAndReplacesProvidedSecrets(t *testing.T) {
	app, cfg := newConfigSaveTestApp(t)
	cfg.TMDB.APIKey = "sentinel-tmdb"
	cfg.Radarr.APIKey = "sentinel-radarr"

	rec, payload := postConfigSave(t, app, map[string]string{
		"jobs.sync_interval":        "2h",
		"tmdb.api_key":              "",
		"radarr.api_key":            "replacement",
		"scoring.rating_weight":     "0.6",
		"scoring.popularity_weight": "0.3",
		"scoring.recency_weight":    "0.1",
	})
	if rec.Code != fiber.StatusOK {
		t.Fatalf("expected 200, got %d: %v", rec.Code, payload)
	}
	if cfg.TMDB.APIKey != "sentinel-tmdb" {
		t.Fatalf("blank field erased stored secret: %q", cfg.TMDB.APIKey)
	}
	if cfg.Radarr.APIKey != "replacement" {
		t.Fatalf("replacement was not stored: %q", cfg.Radarr.APIKey)
	}
}

func TestConfigJSONOmitsSecrets(t *testing.T) {
	cfg := &config.Config{}
	cfg.Trakt.ClientSecret = "sentinel"
	cfg.TMDB.APIKey = "sentinel"
	cfg.Radarr.APIKey = "sentinel"
	cfg.Jellyseerr.RequestCredentials.Password = "sentinel"

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("sentinel")) {
		t.Fatalf("ordinary config JSON exposed a credential: %s", data)
	}
}

func TestConfigSaveValidatesRankedSelection(t *testing.T) {
	app, cfg := newConfigSaveTestApp(t)
	fields := map[string]string{
		"jobs.sync_interval": "2h", "jobs.selection.enabled": "true", "jobs.selection.sync_interval": "24h",
		"jobs.selection.movie_limit": "5", "jobs.selection.show_limit": "0",
		"scoring.enabled": "true", "scoring.rating_weight": "0.6", "scoring.popularity_weight": "0.3", "scoring.recency_weight": "0.1",
	}
	rec, payload := postConfigSave(t, app, fields)
	if rec.Code != fiber.StatusOK {
		t.Fatalf("expected 200, got %d: %v", rec.Code, payload)
	}
	if !cfg.Jobs.Selection.Enabled || cfg.Jobs.Selection.MovieLimit != 5 || cfg.Jobs.Selection.SyncInterval != "24h" {
		t.Fatalf("selection not applied: %+v", cfg.Jobs.Selection)
	}

	fields["scoring.enabled"] = "false"
	if rec, _ := postConfigSave(t, app, fields); rec.Code != fiber.StatusBadRequest {
		t.Fatalf("expected scoring dependency rejection, got %d", rec.Code)
	}
}

func TestConfigSaveUpdatesRankedSelectionMembership(t *testing.T) {
	app, cfg := newConfigSaveTestApp(t)
	cfg.Jobs.List = []config.DynamicJob{
		{ID: "first", Name: "First", Enabled: true, MediaType: "movie", SelectionCycle: true},
		{ID: "second", Name: "Second", Enabled: true, MediaType: "movie"},
	}
	fields := map[string]string{
		"jobs.sync_interval": "2h", "jobs.selection.enabled": "true", "jobs.selection.sync_interval": "24h",
		"jobs.selection.movie_limit": "5", "jobs.selection.show_limit": "0", "jobs.selection.members_present": "true", "jobs.selection.members": "second",
		"scoring.enabled": "true", "scoring.rating_weight": "0.6", "scoring.popularity_weight": "0.3", "scoring.recency_weight": "0.1",
	}
	if rec, payload := postConfigSave(t, app, fields); rec.Code != fiber.StatusOK {
		t.Fatalf("expected 200, got %d: %v", rec.Code, payload)
	}
	if cfg.Jobs.List[0].SelectionCycle || !cfg.Jobs.List[1].SelectionCycle {
		t.Fatalf("membership not applied: %+v", cfg.Jobs.List)
	}
}
