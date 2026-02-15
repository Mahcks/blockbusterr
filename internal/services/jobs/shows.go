package jobs

import (
	"context"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
)

// RunAnticipatedShows fetches anticipated TV shows from Trakt and adds them to Sonarr or requests via Jellyseerr
func RunAnticipatedShows(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.AnticipatedShows.Mode, cfg.Jobs.Mode)

	executor := &ShowJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:   "anticipated_shows",
		MediaType: "show",
		Mode:      mode,
		Monitor:   cfg.Jobs.AnticipatedShows.Monitor,
		Limit:     cfg.Jobs.AnticipatedShows.Limit,
	}, fetchAnticipatedShows)
}

// RunCollectedShows fetches collected TV shows from Trakt and adds them to Sonarr or requests via Jellyseerr
func RunCollectedShows(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.CollectedShows.Mode, cfg.Jobs.Mode)

	executor := &ShowJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:   "collected_shows",
		MediaType: "show",
		Mode:      mode,
		Monitor:   cfg.Jobs.CollectedShows.Monitor,
		Limit:     cfg.Jobs.CollectedShows.Limit,
		Period:    cfg.Jobs.CollectedShows.Period,
	}, fetchCollectedShows)
}

// RunFavoritedShows fetches favorited TV shows from Trakt and adds them to Sonarr or requests via Jellyseerr
func RunFavoritedShows(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.FavoritedShows.Mode, cfg.Jobs.Mode)

	executor := &ShowJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:   "favorited_shows",
		MediaType: "show",
		Mode:      mode,
		Monitor:   cfg.Jobs.FavoritedShows.Monitor,
		Limit:     cfg.Jobs.FavoritedShows.Limit,
		Period:    cfg.Jobs.FavoritedShows.Period,
	}, fetchFavoritedShows)
}

// RunPlayedShows fetches played TV shows from Trakt and adds them to Sonarr or requests via Jellyseerr
func RunPlayedShows(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.PlayedShows.Mode, cfg.Jobs.Mode)

	executor := &ShowJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:   "played_shows",
		MediaType: "show",
		Mode:      mode,
		Monitor:   cfg.Jobs.PlayedShows.Monitor,
		Limit:     cfg.Jobs.PlayedShows.Limit,
		Period:    cfg.Jobs.PlayedShows.Period,
	}, fetchPlayedShows)
}

// RunTrendingShows fetches trending TV shows from Trakt and adds them to Sonarr or requests via Jellyseerr
func RunTrendingShows(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.TrendingShows.Mode, cfg.Jobs.Mode)

	executor := &ShowJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:   "trending_shows",
		MediaType: "show",
		Mode:      mode,
		Monitor:   cfg.Jobs.TrendingShows.Monitor,
		Limit:     cfg.Jobs.TrendingShows.Limit,
	}, fetchTrendingShows)
}

// RunPopularShows fetches popular TV shows from Trakt and adds them to Sonarr or requests via Jellyseerr
func RunPopularShows(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.PopularShows.Mode, cfg.Jobs.Mode)

	executor := &ShowJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:   "popular_shows",
		MediaType: "show",
		Mode:      mode,
		Monitor:   cfg.Jobs.PopularShows.Monitor,
		Limit:     cfg.Jobs.PopularShows.Limit,
	}, fetchPopularShows)
}

// RunWatchedShows fetches watched TV shows from Trakt and adds them to Sonarr or requests via Jellyseerr
func RunWatchedShows(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.WatchedShows.Mode, cfg.Jobs.Mode)

	executor := &ShowJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, JobConfig{
		JobName:   "watched_shows",
		MediaType: "show",
		Mode:      mode,
		Monitor:   cfg.Jobs.WatchedShows.Monitor,
		Period:    cfg.Jobs.WatchedShows.Period,
	}, fetchWatchedShows)
}

// RunSmartPopularShows fetches popular shows and applies adaptive rating thresholds
func RunSmartPopularShows(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	mode := DetermineMode(cfg.Jobs.SmartPopularShows.Mode, cfg.Jobs.Mode)

	executor := &SmartShowJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}

	executor.Execute(ctx, SmartJobConfig{
		JobID:            "legacy_smart_popular_shows",
		JobName:          "smart_popular_shows",
		MediaType:        "show",
		Mode:             mode,
		Monitor:          cfg.Jobs.SmartPopularShows.Monitor,
		Limit:            cfg.Jobs.SmartPopularShows.Limit,
		BaseMinRating:    cfg.Jobs.SmartPopularShows.BaseMinRating,
		AdjustmentFactor: cfg.Jobs.SmartPopularShows.AdjustmentFactor,
	}, fetchPopularShows)
}
