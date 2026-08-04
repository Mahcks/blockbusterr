package routes

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/config"
	"gopkg.in/yaml.v3"
)

func configRoutesTestApp(cfg *config.Config) *fiber.App {
	app := fiber.New()
	RegisterConfigRoutes(app, dynamicJobsTestContext{cfg: cfg})
	return app
}

func uploadConfig(t *testing.T, app *fiber.App, path string, data []byte) *http.Response {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("config", "portable.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", path, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func portableTestConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := &config.Config{ConfigFilePath: filepath.Join(t.TempDir(), "config.yaml")}
	cfg.TMDB.APIKey = "configured-tmdb-secret"
	cfg.Radarr.APIKey = "configured-radarr-secret"
	cfg.Filters.Movies.MinRating = 7
	cfg.MigrateRuleSets()
	rules := config.RuleSet{ID: "quality", Name: "Quality", Media: "movie", Revision: 1, Movies: &config.MovieFilters{MinRating: 8}}
	cfg.RuleSets = append(cfg.RuleSets, rules)
	cfg.Jobs.List = []config.DynamicJob{{ID: "job-1", Name: "Quality Movies", Enabled: true, Type: "popular", Source: "tmdb", MediaType: "movie", Limit: 20, RuleSetID: rules.ID}}
	cfg.TitleExceptions.BlockedMovieTMDBIDs = []int{123}
	return cfg
}

func TestShareableConfigRoundTripIncludesPoliciesAndPreservesCredentials(t *testing.T) {
	source := portableTestConfig(t)
	response, err := configRoutesTestApp(source).Test(httptest.NewRequest("GET", "/config/export", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	exported, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK || !bytes.Contains(exported, []byte("schema_version: 2")) || !bytes.Contains(exported, []byte("rule_sets:")) || bytes.Contains(exported, []byte("configured-radarr-secret")) {
		t.Fatalf("unexpected export status=%d body=%s", response.StatusCode, exported)
	}

	target := &config.Config{ConfigFilePath: filepath.Join(t.TempDir(), "config.yaml")}
	target.TMDB.APIKey = "target-tmdb-secret"
	target.Radarr.APIKey = "target-radarr-secret"
	response = uploadConfig(t, configRoutesTestApp(target), "/config/import", exported)
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status=%d body=%s", response.StatusCode, body)
	}
	if target.TMDB.APIKey != "target-tmdb-secret" || target.Radarr.APIKey != "target-radarr-secret" {
		t.Fatal("portable import changed credentials")
	}
	if len(target.Jobs.List) != 1 || target.Jobs.List[0].RuleSetID != "quality" {
		t.Fatalf("jobs not imported: %+v", target.Jobs.List)
	}
	if rules, ok := target.RuleSetByID("quality"); !ok || rules.Movies == nil || rules.Movies.MinRating != 8 {
		t.Fatalf("rule set not imported: %+v", rules)
	}
	if len(target.TitleExceptions.BlockedMovieTMDBIDs) != 1 || target.TitleExceptions.BlockedMovieTMDBIDs[0] != 123 {
		t.Fatalf("title exceptions not imported: %+v", target.TitleExceptions)
	}
}

func TestJobBundleImportRegeneratesIDsAndDisablesJob(t *testing.T) {
	source := portableTestConfig(t)
	response, err := configRoutesTestApp(source).Test(httptest.NewRequest("GET", "/config/jobs/job-1/export", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	exported, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	var bundle jobBundle
	if err := yaml.Unmarshal(exported, &bundle); err != nil {
		t.Fatal(err)
	}
	if bundle.Job.ID != "job-1" || bundle.RuleSet.ID != "quality" {
		t.Fatalf("unexpected bundle: %+v", bundle)
	}

	target := portableTestConfig(t)
	target.Jobs.List = nil
	response = uploadConfig(t, configRoutesTestApp(target), "/config/jobs/import", exported)
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != fiber.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status=%d body=%s", response.StatusCode, body)
	}
	if len(target.Jobs.List) != 1 {
		t.Fatalf("imported jobs=%d", len(target.Jobs.List))
	}
	job := target.Jobs.List[0]
	if job.ID == bundle.Job.ID || job.RuleSetID == bundle.RuleSet.ID || job.Enabled {
		t.Fatalf("unsafe imported job: %+v", job)
	}
	rules, ok := target.RuleSetByID(job.RuleSetID)
	if !ok || !strings.HasPrefix(rules.Name, "Quality") {
		t.Fatalf("imported rules missing: %+v", rules)
	}
}
