package config

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestNewAutomaticallyMigratesAndBacksUpLegacyJobs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	original := []byte(`version: v1.5.0
trakt:
  client_id: preserved-client
jobs:
  sync_interval: "0 8 * * *"
  trending_movies:
    enabled: true
    limit: 37
filters:
  movies:
    blacklisted_genres: [Horror]
`)
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", dir)

	cfg, err := New("v2.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Jobs.List) != 1 || cfg.Jobs.List[0].ID != "trending_movies" || cfg.Jobs.List[0].Limit != 37 || cfg.Jobs.List[0].RuleSetID != DefaultMoviesRuleSetID {
		t.Fatalf("migrated jobs = %+v", cfg.Jobs.List)
	}
	if cfg.Jobs.TrendingMovies.Enabled || cfg.Trakt.ClientID != "preserved-client" {
		t.Fatalf("legacy state was not preserved safely: %+v", cfg)
	}
	backup, err := os.ReadFile(path + ".backup")
	if err != nil || !bytes.Equal(backup, original) {
		t.Fatalf("backup=%q err=%v, want original config", backup, err)
	}

	reloaded, err := New("v2.0.0")
	if err != nil || len(reloaded.Jobs.List) != 1 {
		t.Fatalf("second startup changed migration: jobs=%+v err=%v", reloaded.Jobs.List, err)
	}
}

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

func TestLegacyGenreFiltersSurviveMigrationSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	legacy := []byte(`version: v1.4.0
jobs:
  trending_movies:
    enabled: true
  trending_shows:
    enabled: true
filters:
  movies:
    blacklisted_genres: [Horror, Documentary]
  shows:
    blacklisted_genres: [Reality, Talk]
`)
	if err := os.WriteFile(path, legacy, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", dir)

	cfg, err := New("prod")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cfg.MigrateLegacyJobs(); err != nil {
		t.Fatal(err)
	}
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	reloaded, err := New("prod")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{
		"trending_movies": {"Horror", "Documentary"},
		"trending_shows":  {"Reality", "Talk"},
	}
	for _, job := range reloaded.Jobs.List {
		expected, ok := want[job.ID]
		if !ok {
			continue
		}
		_, filters, err := reloaded.ResolveRuleSet(job)
		if err != nil {
			t.Fatal(err)
		}
		actual := filters.Movies.BlacklistedGenres
		if job.MediaType == "show" {
			actual = filters.Shows.BlacklistedGenres
		}
		if !slices.Equal(actual, expected) {
			t.Fatalf("%s genres = %v, want %v", job.ID, actual, expected)
		}
		delete(want, job.ID)
	}
	if len(want) != 0 {
		t.Fatalf("migrated jobs missing after reload: %v", want)
	}
}
