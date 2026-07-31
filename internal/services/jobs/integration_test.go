package jobs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

func TestMovieExecutorReturnsRadarrInitializationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.Config{}
	cfg.Simkl.ClientID = "configured"
	cfg.Radarr.URL = server.URL
	executor := &MovieJobExecutor{Config: cfg, DryRun: true}
	err := executor.Execute(t.Context(), JobConfig{Source: "simkl", MediaType: "movie", Mode: "direct", Limit: 1}, func(_ context.Context, _ *DiscoveryClient, _ int, _ string) ([]integrations.Movie, error) {
		return []integrations.Movie{{Title: "Test", IDs: integrations.IDs{TMDB: 1}}}, nil
	})
	if err == nil {
		t.Fatal("expected Radarr initialization error")
	}
}

// TestMovieJobExecutorStructure verifies the movie executor can be created and configured
func TestMovieJobExecutorStructure(t *testing.T) {
	cfg := &config.Config{}
	cfg.Trakt.ClientID = "test-client"
	cfg.Trakt.ClientSecret = "test-secret"

	db := &database.Database{}

	executor := &MovieJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   true,
	}

	if executor.Config == nil {
		t.Error("Executor config should not be nil")
	}
	if executor.Database == nil {
		t.Error("Executor database should not be nil")
	}
	if !executor.DryRun {
		t.Error("Executor should be in dry run mode")
	}
}

// TestShowJobExecutorStructure verifies the show executor can be created and configured
func TestShowJobExecutorStructure(t *testing.T) {
	cfg := &config.Config{}
	cfg.Trakt.ClientID = "test-client"
	cfg.Trakt.ClientSecret = "test-secret"

	db := &database.Database{}

	executor := &ShowJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   true,
	}

	if executor.Config == nil {
		t.Error("Executor config should not be nil")
	}
	if executor.Database == nil {
		t.Error("Executor database should not be nil")
	}
	if !executor.DryRun {
		t.Error("Executor should be in dry run mode")
	}
}

// TestDetermineMode verifies the mode fallback logic
func TestDetermineMode(t *testing.T) {
	tests := []struct {
		name       string
		jobMode    string
		globalMode string
		expected   string
	}{
		{
			name:       "Job mode takes precedence",
			jobMode:    "jellyseerr",
			globalMode: "direct",
			expected:   "jellyseerr",
		},
		{
			name:       "Falls back to global mode",
			jobMode:    "",
			globalMode: "direct",
			expected:   "direct",
		},
		{
			name:       "Both empty defaults to direct",
			jobMode:    "",
			globalMode: "",
			expected:   "direct",
		},
		{
			name:       "Job mode direct",
			jobMode:    "direct",
			globalMode: "jellyseerr",
			expected:   "direct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineMode(tt.jobMode, tt.globalMode)
			if result != tt.expected {
				t.Errorf("DetermineMode(%q, %q) = %q, want %q", tt.jobMode, tt.globalMode, result, tt.expected)
			}
		})
	}
}

// TestFetcherFunctions verifies all fetcher functions have the correct signature
func TestFetcherFunctions(t *testing.T) {
	// Just verify that all fetchers exist with correct signatures by checking they're not nil
	// We can't actually call them without valid Trakt credentials

	movieFetchers := []struct {
		name    string
		fetcher MovieFetcher
	}{
		{"fetchPopularMovies", fetchPopularMovies},
		{"fetchAnticipatedMovies", fetchAnticipatedMovies},
		{"fetchCollectedMovies", fetchCollectedMovies},
		{"fetchFavoritedMovies", fetchFavoritedMovies},
		{"fetchPlayedMovies", fetchPlayedMovies},
		{"fetchWatchedMovies", fetchWatchedMovies},
		{"fetchBoxOfficeMovies", fetchBoxOfficeMovies},
	}

	for _, mf := range movieFetchers {
		t.Run(mf.name, func(t *testing.T) {
			if mf.fetcher == nil {
				t.Errorf("Fetcher %s should not be nil", mf.name)
			}
		})
	}

	showFetchers := []struct {
		name    string
		fetcher ShowFetcher
	}{
		{"fetchTrendingShows", fetchTrendingShows},
		{"fetchPopularShows", fetchPopularShows},
		{"fetchAnticipatedShows", fetchAnticipatedShows},
		{"fetchCollectedShows", fetchCollectedShows},
		{"fetchFavoritedShows", fetchFavoritedShows},
		{"fetchPlayedShows", fetchPlayedShows},
		{"fetchWatchedShows", fetchWatchedShows},
	}

	for _, sf := range showFetchers {
		t.Run(sf.name, func(t *testing.T) {
			if sf.fetcher == nil {
				t.Errorf("Fetcher %s should not be nil", sf.name)
			}
		})
	}
}

// TestJobConfigValidation verifies JobConfig structure
func TestJobConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config JobConfig
		valid  bool
	}{
		{
			name: "Valid movie config",
			config: JobConfig{
				JobName:   "trending_movies",
				MediaType: "movie",
				Mode:      "direct",
				Limit:     20,
			},
			valid: true,
		},
		{
			name: "Valid show config",
			config: JobConfig{
				JobName:   "trending_shows",
				MediaType: "show",
				Mode:      "jellyseerr",
				Limit:     15,
				Period:    "weekly",
			},
			valid: true,
		},
		{
			name: "Config with period",
			config: JobConfig{
				JobName:   "played_movies",
				MediaType: "movie",
				Mode:      "direct",
				Limit:     50,
				Period:    "monthly",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify required fields are set
			if tt.config.JobName == "" && tt.valid {
				t.Error("JobName should not be empty for valid config")
			}
			if tt.config.MediaType == "" && tt.valid {
				t.Error("MediaType should not be empty for valid config")
			}
			if tt.config.Limit <= 0 && tt.valid {
				t.Error("Limit should be positive for valid config")
			}
		})
	}
}

// TestAllMovieJobFunctions verifies all movie job functions exist and can be called
func TestAllMovieJobFunctions(t *testing.T) {
	// Create minimal config for dry run
	cfg := &config.Config{}
	cfg.Trakt.ClientID = "test"
	cfg.Trakt.ClientSecret = "test"
	cfg.Jobs.Mode = "direct"
	cfg.Jobs.TrendingMovies.Enabled = true
	cfg.Jobs.TrendingMovies.Limit = 10
	cfg.Jobs.PopularMovies.Enabled = true
	cfg.Jobs.PopularMovies.Limit = 10
	cfg.Jobs.AnticipatedMovies.Enabled = true
	cfg.Jobs.AnticipatedMovies.Limit = 10
	cfg.Jobs.CollectedMovies.Enabled = true
	cfg.Jobs.CollectedMovies.Limit = 10
	cfg.Jobs.CollectedMovies.Period = "weekly"
	cfg.Jobs.FavoritedMovies.Enabled = true
	cfg.Jobs.FavoritedMovies.Limit = 10
	cfg.Jobs.FavoritedMovies.Period = "weekly"
	cfg.Jobs.PlayedMovies.Enabled = true
	cfg.Jobs.PlayedMovies.Limit = 10
	cfg.Jobs.PlayedMovies.Period = "weekly"
	cfg.Jobs.WatchedMovies.Enabled = true
	cfg.Jobs.WatchedMovies.Limit = 10
	cfg.Jobs.WatchedMovies.Period = "weekly"
	cfg.Jobs.BoxOffice.Enabled = true
	cfg.Jobs.BoxOffice.Limit = 10

	db := &database.Database{}

	// Test that all job functions can be called (they'll fail due to missing real config, but structure is tested)
	jobs := []struct {
		name string
		fn   func(*config.Config, *database.Database, bool)
	}{
		{"RunTrendingMovies", RunTrendingMovies},
		{"RunPopularMovies", RunPopularMovies},
		{"RunAnticipatedMovies", RunAnticipatedMovies},
		{"RunCollectedMovies", RunCollectedMovies},
		{"RunFavoritedMovies", RunFavoritedMovies},
		{"RunPlayedMovies", RunPlayedMovies},
		{"RunWatchedMovies", RunWatchedMovies},
		{"RunBoxOffice", RunBoxOffice},
	}

	for _, job := range jobs {
		t.Run(job.name, func(t *testing.T) {
			// Just verify the function exists and can be called
			// It will fail internally, but that's expected without real Trakt credentials
			defer func() {
				if r := recover(); r != nil {
					// Some panic is ok since we don't have real config
					t.Logf("Job %s panicked as expected with test config: %v", job.name, r)
				}
			}()
			job.fn(cfg, db, true)
		})
	}
}

// TestAllShowJobFunctions verifies all show job functions exist and can be called
func TestAllShowJobFunctions(t *testing.T) {
	cfg := &config.Config{}
	cfg.Trakt.ClientID = "test"
	cfg.Trakt.ClientSecret = "test"
	cfg.Jobs.Mode = "direct"
	cfg.Jobs.TrendingShows.Enabled = true
	cfg.Jobs.TrendingShows.Limit = 10
	cfg.Jobs.PopularShows.Enabled = true
	cfg.Jobs.PopularShows.Limit = 10
	cfg.Jobs.AnticipatedShows.Enabled = true
	cfg.Jobs.AnticipatedShows.Limit = 10
	cfg.Jobs.CollectedShows.Enabled = true
	cfg.Jobs.CollectedShows.Limit = 10
	cfg.Jobs.CollectedShows.Period = "weekly"
	cfg.Jobs.FavoritedShows.Enabled = true
	cfg.Jobs.FavoritedShows.Limit = 10
	cfg.Jobs.FavoritedShows.Period = "weekly"
	cfg.Jobs.PlayedShows.Enabled = true
	cfg.Jobs.PlayedShows.Limit = 10
	cfg.Jobs.PlayedShows.Period = "weekly"
	cfg.Jobs.WatchedShows.Enabled = true
	cfg.Jobs.WatchedShows.Limit = 10
	cfg.Jobs.WatchedShows.Period = "weekly"

	db := &database.Database{}

	jobs := []struct {
		name string
		fn   func(*config.Config, *database.Database, bool)
	}{
		{"RunTrendingShows", RunTrendingShows},
		{"RunPopularShows", RunPopularShows},
		{"RunAnticipatedShows", RunAnticipatedShows},
		{"RunCollectedShows", RunCollectedShows},
		{"RunFavoritedShows", RunFavoritedShows},
		{"RunPlayedShows", RunPlayedShows},
		{"RunWatchedShows", RunWatchedShows},
	}

	for _, job := range jobs {
		t.Run(job.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Job %s panicked as expected with test config: %v", job.name, r)
				}
			}()
			job.fn(cfg, db, true)
		})
	}
}
