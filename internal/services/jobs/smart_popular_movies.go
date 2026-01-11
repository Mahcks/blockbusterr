package jobs

import (
	"context"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
)

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
		JobName:          "smart_popular_movies",
		MediaType:        "movie",
		Mode:             mode,
		Limit:            cfg.Jobs.SmartPopularMovies.Limit,
		BaseMinRating:    cfg.Jobs.SmartPopularMovies.BaseMinRating,
		AdjustmentFactor: cfg.Jobs.SmartPopularMovies.AdjustmentFactor,
	}, fetchPopularMovies)
}
