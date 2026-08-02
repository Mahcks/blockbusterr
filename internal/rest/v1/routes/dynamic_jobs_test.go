package routes

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/global"
)

type dynamicJobsTestContext struct {
	context.Context
	cfg *config.Config
}

func (c dynamicJobsTestContext) Config() *config.Config       { return c.cfg }
func (c dynamicJobsTestContext) Database() *database.Database { return nil }
func (c dynamicJobsTestContext) Metadata() global.Metadata    { return global.Metadata{Version: "test"} }
func (c dynamicJobsTestContext) ReloadConfig() error          { return nil }

func TestValidateDynamicJobSources(t *testing.T) {
	cfg := &config.Config{}
	cfg.Trakt.ClientID = "configured"
	cfg.TMDB.APIKey = "configured"
	cfg.Simkl.ClientID = "configured"
	tests := []struct {
		name    string
		job     config.DynamicJob
		wantErr bool
	}{
		{name: "TMDB trending", job: config.DynamicJob{Name: "Trending", Type: "trending", Source: "tmdb", MediaType: "movie", Limit: 50}},
		{name: "Simkl watched", job: config.DynamicJob{Name: "Watched", Type: "watched", Source: "simkl", MediaType: "show", Limit: 50, Period: "weekly"}},
		{name: "TMDB collected unsupported", job: config.DynamicJob{Name: "Collected", Type: "collected", Source: "tmdb", MediaType: "movie", Limit: 50, Period: "weekly"}, wantErr: true},
		{name: "Simkl yearly unsupported", job: config.DynamicJob{Name: "Watched", Type: "watched", Source: "simkl", MediaType: "movie", Limit: 50, Period: "yearly"}, wantErr: true},
		{name: "Simkl limit", job: config.DynamicJob{Name: "Trending", Type: "trending", Source: "simkl", MediaType: "movie", Limit: 501}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateDynamicJob(cfg, test.job)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateDynamicJob() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestValidateDynamicJobRejectsUnconfiguredProvider(t *testing.T) {
	job := config.DynamicJob{Name: "Trending", Type: "trending", Source: "simkl", MediaType: "movie", Limit: 50}
	if err := validateDynamicJob(&config.Config{}, job); err == nil {
		t.Fatal("expected unconfigured provider error")
	}
}

func TestValidateDynamicJobRejectsInvalidCustomFilters(t *testing.T) {
	cfg := &config.Config{}
	cfg.TMDB.APIKey = "configured"
	job := config.DynamicJob{Name: "Trending", Type: "trending", Source: "tmdb", MediaType: "movie", Limit: 50, UseCustomFilters: true}
	job.Filters.Movies.BlacklistedMinYear = 2025
	job.Filters.Movies.BlacklistedMaxYear = 2000
	if err := validateDynamicJob(cfg, job); err == nil {
		t.Fatal("expected invalid custom filter range error")
	}
}

func TestCustomizeRulesCreatesAndAssignsJobSpecificCopy(t *testing.T) {
	cfg := &config.Config{ConfigFilePath: filepath.Join(t.TempDir(), "config.yaml")}
	cfg.Filters.Movies.MinRating = 7
	cfg.MigrateRuleSets()
	cfg.Jobs.List = []config.DynamicJob{{ID: "job-1", Name: "Trending Simkl", MediaType: "movie", RuleSetID: config.DefaultMoviesRuleSetID}}
	app := fiber.New()
	AddDynamicJobsRoutes(app.Group("/v1"), dynamicJobsTestContext{Context: context.Background(), cfg: cfg})

	response, err := app.Test(httptest.NewRequest("POST", "/v1/jobs/job-1/customize-rules", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusCreated)
	}
	var result struct {
		RuleSet config.RuleSet    `json:"rule_set"`
		Job     config.DynamicJob `json:"job"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.RuleSet.ID == config.DefaultMoviesRuleSetID || result.Job.RuleSetID != result.RuleSet.ID {
		t.Fatalf("job assignment = %q, cloned rule set = %q", result.Job.RuleSetID, result.RuleSet.ID)
	}
	if result.RuleSet.Name != "Trending Simkl Rules" || result.RuleSet.Movies == nil || result.RuleSet.Movies.MinRating != 7 {
		t.Fatalf("unexpected cloned rule set: %+v", result.RuleSet)
	}
}
