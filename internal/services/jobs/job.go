package jobs

import (
	"context"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// JobConfig contains common configuration for all jobs
type JobConfig struct {
	JobID               string // Unique identifier for the job
	JobName             string // e.g., "Anticipated Movies", "Popular Shows"
	Source              string // "trakt", "tmdb", or "simkl"
	MediaType           string // "movie" or "show"
	Mode                string // "direct" or "jellyseerr"
	MinimumAvailability string // For direct mode only - "announced", "in_cinemas", or "released"
	/*
		For Radarr: "movieOnly", "movieAndCollection", "none";

		For Sonarr: "all", "future", "missing", "existing", "pilot", "firstSeason", "latestSeason", "none"
	*/
	Monitor       string
	Limit         int
	DeliveryLimit int
	RepeatPolicy  string
	Period        string // For watched/collected/played jobs
	SeriesType    string // Sonarr series type: standard, daily, or anime
}

// FormatJobLabel returns a human-readable job label for logs.
func FormatJobLabel(jobID, jobName string) string {
	if jobName == "" {
		return jobID
	}
	if jobID == "" {
		return jobName
	}
	return jobName + " (" + jobID + ")"
}

// ScoreInfo holds scoring information for a media item
type ScoreInfo struct {
	Score float64
	Rank  int
}

// MovieFetcher fetches provider-neutral movies from a configured discovery source.
type MovieFetcher func(ctx context.Context, discovery *DiscoveryClient, limit int, period string) ([]integrations.Movie, error)

// ShowFetcher fetches provider-neutral shows from a configured discovery source.
type ShowFetcher func(ctx context.Context, discovery *DiscoveryClient, limit int, period string) ([]integrations.Show, error)

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
