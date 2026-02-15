package jobs

import (
	"context"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// RunAnticipatedMovies fetches most anticipated movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunAnticipatedMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.AnticipatedMovies.Mode, cfg.Jobs.Mode)

	executor := &MovieJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:             "anticipated_movies",
		MediaType:           "movie",
		Mode:                mode,
		MinimumAvailability: cfg.Jobs.AnticipatedMovies.MinimumAvailability,
		Monitor:             cfg.Jobs.AnticipatedMovies.Monitor,
		Limit:               cfg.Jobs.AnticipatedMovies.Limit,
	}, fetchAnticipatedMovies)
}

// RunCollectedMovies fetches most collected movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunCollectedMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.CollectedMovies.Mode, cfg.Jobs.Mode)

	executor := &MovieJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:             "collected_movies",
		MediaType:           "movie",
		Mode:                mode,
		MinimumAvailability: cfg.Jobs.CollectedMovies.MinimumAvailability,
		Monitor:             cfg.Jobs.CollectedMovies.Monitor,
		Limit:               cfg.Jobs.CollectedMovies.Limit,
		Period:              cfg.Jobs.CollectedMovies.Period,
	}, fetchCollectedMovies)
}

// RunFavoritedMovies fetches favorited movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunFavoritedMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.FavoritedMovies.Mode, cfg.Jobs.Mode)

	executor := &MovieJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:             "favorited_movies",
		MediaType:           "movie",
		Mode:                mode,
		MinimumAvailability: cfg.Jobs.FavoritedMovies.MinimumAvailability,
		Monitor:             cfg.Jobs.FavoritedMovies.Monitor,
		Limit:               cfg.Jobs.FavoritedMovies.Limit,
		Period:              cfg.Jobs.FavoritedMovies.Period,
	}, fetchFavoritedMovies)
}

// RunPlayedMovies fetches most played movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunPlayedMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.PlayedMovies.Mode, cfg.Jobs.Mode)

	executor := &MovieJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:             "played_movies",
		MediaType:           "movie",
		Mode:                mode,
		MinimumAvailability: cfg.Jobs.PlayedMovies.MinimumAvailability,
		Monitor:             cfg.Jobs.PlayedMovies.Monitor,
		Limit:               cfg.Jobs.PlayedMovies.Limit,
		Period:              cfg.Jobs.PlayedMovies.Period,
	}, fetchPlayedMovies)
}

// RunPopularMovies fetches popular movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunPopularMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.PopularMovies.Mode, cfg.Jobs.Mode)

	executor := &MovieJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:             "popular_movies",
		MediaType:           "movie",
		Mode:                mode,
		MinimumAvailability: cfg.Jobs.PopularMovies.MinimumAvailability,
		Monitor:             cfg.Jobs.PopularMovies.Monitor,
		Limit:               cfg.Jobs.PopularMovies.Limit,
	}, fetchPopularMovies)
}

// RunTrendingMovies fetches trending movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunTrendingMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Determine mode - check job-specific mode first, then fall back to global
	mode := DetermineMode(cfg.Jobs.TrendingMovies.Mode, cfg.Jobs.Mode)

	// Create executor
	executor := &MovieJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	// Define fetcher function for trending movies
	fetcher := func(ctx context.Context, trakt *integrations.Trakt, limit int, _ string) ([]integrations.Movie, error) {
		trendingMovies, err := trakt.GetTrendingMovies(ctx, limit)
		if err != nil {
			return nil, err
		}
		// Extract Movie from TrendingMovie wrapper
		movies := make([]integrations.Movie, len(trendingMovies))
		for i, tm := range trendingMovies {
			movies[i] = tm.Movie
		}
		return movies, nil
	}

	// Execute the job
	executor.Execute(ctx, JobConfig{
		JobName:             "trending_movies",
		MediaType:           "movie",
		Mode:                mode,
		MinimumAvailability: cfg.Jobs.TrendingMovies.MinimumAvailability,
		Monitor:             cfg.Jobs.TrendingMovies.Monitor,
		Limit:               cfg.Jobs.TrendingMovies.Limit,
	}, fetcher)
}

// RunWatchedMovies fetches most watched movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunWatchedMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.WatchedMovies.Mode, cfg.Jobs.Mode)

	executor := &MovieJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:             "watched_movies",
		MediaType:           "movie",
		Mode:                mode,
		MinimumAvailability: cfg.Jobs.WatchedMovies.MinimumAvailability,
		Monitor:             cfg.Jobs.WatchedMovies.Monitor,
		Limit:               cfg.Jobs.WatchedMovies.Limit,
		Period:              cfg.Jobs.WatchedMovies.Period,
	}, fetchWatchedMovies)
}

// RunBoxOffice fetches box office movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunBoxOffice(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.BoxOffice.Mode, cfg.Jobs.Mode)

	executor := &MovieJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:             "box_office",
		MediaType:           "movie",
		Mode:                mode,
		MinimumAvailability: cfg.Jobs.BoxOffice.MinimumAvailability,
		Monitor:             cfg.Jobs.BoxOffice.Monitor,
		Limit:               cfg.Jobs.BoxOffice.Limit,
	}, fetchBoxOfficeMovies)
}

// RunSmartPopularMovies fetches popular movies and applies adaptive rating thresholds
func RunSmartPopularMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.SmartPopularMovies.Mode, cfg.Jobs.Mode)

	executor := &SmartMovieJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, SmartJobConfig{
		JobID:               "legacy_smart_popular_movies",
		JobName:             "smart_popular_movies",
		MediaType:           "movie",
		Mode:                mode,
		MinimumAvailability: cfg.Jobs.SmartPopularMovies.MinimumAvailability,
		Monitor:             cfg.Jobs.SmartPopularMovies.Monitor,
		Limit:               cfg.Jobs.SmartPopularMovies.Limit,
		BaseMinRating:       cfg.Jobs.SmartPopularMovies.BaseMinRating,
		AdjustmentFactor:    cfg.Jobs.SmartPopularMovies.AdjustmentFactor,
	}, fetchPopularMovies)
}
