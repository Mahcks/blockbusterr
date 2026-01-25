package jobs

import (
	"context"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// JobConfig contains common configuration for all jobs
type JobConfig struct {
	JobName             string // e.g., "trending_movies", "popular_shows"
	MediaType           string // "movie" or "show"
	Mode                string // "direct" or "jellyseerr"
	MinimumAvailability string // For direct mode only - "announced", "in_cinemas", or "released"
	Monitor             string // For Radarr direct mode - "movieOnly", "movieAndCollection", or "none"
	Limit               int
	Period              string // For watched/collected/played jobs
}

// ScoreInfo holds scoring information for a media item
type ScoreInfo struct {
	Score float64
	Rank  int
}

// MovieFetcher is a function type that fetches movies from Trakt
type MovieFetcher func(ctx context.Context, trakt *integrations.Trakt, limit int, period string) ([]integrations.Movie, error)

// ShowFetcher is a function type that fetches shows from Trakt
type ShowFetcher func(ctx context.Context, trakt *integrations.Trakt, limit int, period string) ([]integrations.Show, error)

// MovieExtractor extracts Movie from wrapped types (e.g., TrendingMovie.Movie)
type MovieExtractor func(any) integrations.Movie

// ShowExtractor extracts Show from wrapped types (e.g., TrendingShow.Show)
type ShowExtractor func(any) integrations.Show

// DetermineMode checks job-specific mode, falls back to global mode, defaults to "direct"
func DetermineMode(jobMode, globalMode string) string {
	if jobMode != "" {
		return jobMode
	}
	if globalMode != "" {
		return globalMode
	}
	return "direct"
}

// JobExecutor provides the common execution flow for all jobs
type JobExecutor struct {
	Config   *config.Config
	Database *database.Database
	DryRun   bool
}
