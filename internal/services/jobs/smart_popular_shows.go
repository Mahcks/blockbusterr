package jobs

import (
	"context"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
)

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
		JobName:          "smart_popular_shows",
		MediaType:        "show",
		Mode:             mode,
		Limit:            cfg.Jobs.SmartPopularShows.Limit,
		BaseMinRating:    cfg.Jobs.SmartPopularShows.BaseMinRating,
		AdjustmentFactor: cfg.Jobs.SmartPopularShows.AdjustmentFactor,
	}, fetchPopularShows)
}
