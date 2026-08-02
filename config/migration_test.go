package config

import "testing"

func TestMigrateLegacyJobsPreservesSettings(t *testing.T) {
	cfg := &Config{}
	cfg.Jobs.TrendingMovies.Enabled = true
	cfg.Jobs.TrendingMovies.Limit = 75
	cfg.Jobs.TrendingMovies.SyncInterval = "6h"
	cfg.Jobs.TrendingMovies.Mode = "jellyseerr"
	cfg.Jobs.TrendingMovies.Monitor = "movieAndCollection"

	migrated, err := cfg.MigrateLegacyJobs()
	if err != nil {
		t.Fatal(err)
	}
	if len(migrated) != 1 || len(cfg.Jobs.List) != 1 {
		t.Fatalf("migrated = %d, jobs = %d, want 1 each", len(migrated), len(cfg.Jobs.List))
	}
	job := cfg.Jobs.List[0]
	if job.ID != "trending_movies" || job.RuleSetID != DefaultMoviesRuleSetID || job.Limit != 75 || job.SyncInterval != "6h" || job.Mode != "jellyseerr" || job.Monitor != "movieAndCollection" {
		t.Fatalf("settings not preserved: %+v", job)
	}
	if cfg.Jobs.TrendingMovies.Enabled {
		t.Fatal("legacy job remained enabled")
	}
}

func TestMigrateLegacyJobsRollsBackOnCollision(t *testing.T) {
	cfg := &Config{}
	cfg.Jobs.TrendingMovies.Enabled = true
	cfg.Jobs.List = []DynamicJob{{ID: "trending_movies", Type: "popular", MediaType: "show"}}

	if _, err := cfg.MigrateLegacyJobs(); err == nil {
		t.Fatal("expected ID collision error")
	}
	if !cfg.Jobs.TrendingMovies.Enabled || len(cfg.Jobs.List) != 1 || cfg.Jobs.List[0].Type != "popular" {
		t.Fatalf("migration changed config after failure: %+v", cfg.Jobs)
	}
}
