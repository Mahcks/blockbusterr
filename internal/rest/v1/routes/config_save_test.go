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

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

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

func TestConfigSaveRejectsMalformedNumbers(t *testing.T) {
	app, cfg := newConfigSaveTestApp(t)

	rec, payload := postConfigSave(t, app, map[string]string{
		"jobs.sync_interval":       "1h",
		"jobs.global_limit_movies": "not-a-number",
	})

	if rec.Code != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if payload["error"] == nil {
		t.Fatal("expected a structured error message")
	}
	if cfg.Jobs.GlobalLimitMovies != 0 {
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
}
