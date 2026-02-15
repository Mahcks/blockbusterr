package jobs

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// DynamicJobExecutor executes jobs based on DynamicJob configuration
type DynamicJobExecutor struct {
	Config   *config.Config
	Database *database.Database
	DryRun   bool
}

// Execute runs a dynamic job, routing to the appropriate executor based on job type and media
func (e *DynamicJobExecutor) Execute(ctx context.Context, job config.DynamicJob) error {
	// Validate job type
	typeDef, ok := GetJobTypeDefinition(job.Type)
	if !ok {
		return fmt.Errorf("unknown job type: %s", job.Type)
	}

	// Validate media type support
	if !SupportsMediaType(job.Type, job.MediaType) {
		return fmt.Errorf("job type %s does not support media type %s", job.Type, job.MediaType)
	}

	// Determine effective mode
	mode := DetermineMode(job.Mode, e.Config.Jobs.Mode)

	// Build JobConfig from DynamicJob
	jobConfig := JobConfig{
		JobID:               job.ID,
		JobName:             job.Name,
		MediaType:           job.MediaType,
		Mode:                mode,
		MinimumAvailability: job.MinimumAvailability,
		Monitor:             job.Monitor,
		Limit:               job.Limit,
		Period:              job.Period,
	}

	// Set default limit if not specified
	if jobConfig.Limit <= 0 {
		jobConfig.Limit = typeDef.DefaultLimit
	}

	// Enforce max limit
	if jobConfig.Limit > typeDef.MaxLimit {
		jobConfig.Limit = typeDef.MaxLimit
	}

	// Route to appropriate executor based on media type
	switch job.MediaType {
	case "movie":
		return e.executeMovieJob(ctx, job, jobConfig)
	case "show":
		return e.executeShowJob(ctx, job, jobConfig)
	default:
		return fmt.Errorf("unsupported media type: %s", job.MediaType)
	}
}

// executeMovieJob executes a movie job using the appropriate executor
func (e *DynamicJobExecutor) executeMovieJob(ctx context.Context, job config.DynamicJob, jobConfig JobConfig) error {
	// Get the appropriate fetcher based on job type
	fetcher := e.getMovieFetcher(job.Type)
	if fetcher == nil {
		return fmt.Errorf("no fetcher available for job type: %s", job.Type)
	}

	if job.Type == "smart_popular" {
		// Use SmartMovieJobExecutor for smart jobs
		executor := &SmartMovieJobExecutor{
			Config:   e.Config,
			Database: e.Database,
			DryRun:   e.DryRun,
		}

		// Build smart job config
		smartConfig := SmartJobConfig{
			JobID:               jobConfig.JobID,
			JobName:             jobConfig.JobName,
			MediaType:           jobConfig.MediaType,
			Mode:                jobConfig.Mode,
			MinimumAvailability: jobConfig.MinimumAvailability,
			Monitor:             jobConfig.Monitor,
			Limit:               jobConfig.Limit,
			BaseMinRating:       job.BaseMinRating,
			AdjustmentFactor:    job.AdjustmentFactor,
		}

		// Set default smart job parameters if not specified
		if smartConfig.BaseMinRating == 0 {
			smartConfig.BaseMinRating = 6.0
		}
		if smartConfig.AdjustmentFactor == 0 {
			smartConfig.AdjustmentFactor = 0.5
		}

		executor.Execute(ctx, smartConfig, fetcher)
	} else {
		// Use standard MovieJobExecutor
		executor := &MovieJobExecutor{
			Config:   e.Config,
			Database: e.Database,
			DryRun:   e.DryRun,
		}
		executor.Execute(ctx, jobConfig, fetcher)
	}

	return nil
}

// executeShowJob executes a show job using the appropriate executor
func (e *DynamicJobExecutor) executeShowJob(ctx context.Context, job config.DynamicJob, jobConfig JobConfig) error {
	// Get the appropriate fetcher based on job type
	fetcher := e.getShowFetcher(job.Type)
	if fetcher == nil {
		return fmt.Errorf("no fetcher available for job type: %s", job.Type)
	}

	if job.Type == "smart_popular" {
		// Use SmartShowJobExecutor for smart jobs
		executor := &SmartShowJobExecutor{
			Config:   e.Config,
			Database: e.Database,
			DryRun:   e.DryRun,
		}

		// Build smart job config
		smartConfig := SmartJobConfig{
			JobID:            jobConfig.JobID,
			JobName:          jobConfig.JobName,
			MediaType:        jobConfig.MediaType,
			Mode:             jobConfig.Mode,
			Monitor:          jobConfig.Monitor,
			Limit:            jobConfig.Limit,
			BaseMinRating:    job.BaseMinRating,
			AdjustmentFactor: job.AdjustmentFactor,
		}

		// Set default smart job parameters if not specified
		if smartConfig.BaseMinRating == 0 {
			smartConfig.BaseMinRating = 6.0
		}
		if smartConfig.AdjustmentFactor == 0 {
			smartConfig.AdjustmentFactor = 0.5
		}

		executor.Execute(ctx, smartConfig, fetcher)
	} else {
		// Use standard ShowJobExecutor
		executor := &ShowJobExecutor{
			Config:   e.Config,
			Database: e.Database,
			DryRun:   e.DryRun,
		}
		executor.Execute(ctx, jobConfig, fetcher)
	}

	return nil
}

// getMovieFetcher returns the appropriate fetcher function for a movie job type
func (e *DynamicJobExecutor) getMovieFetcher(jobType string) MovieFetcher {
	switch jobType {
	case "trending":
		return func(ctx context.Context, trakt *integrations.Trakt, limit int, _ string) ([]integrations.Movie, error) {
			trendingMovies, err := trakt.GetTrendingMovies(ctx, limit)
			if err != nil {
				return nil, err
			}
			movies := make([]integrations.Movie, len(trendingMovies))
			for i, tm := range trendingMovies {
				movies[i] = tm.Movie
			}
			return movies, nil
		}
	case "popular", "smart_popular":
		return fetchPopularMovies
	case "watched":
		return fetchWatchedMovies
	case "collected":
		return fetchCollectedMovies
	case "favorited":
		return fetchFavoritedMovies
	case "played":
		return fetchPlayedMovies
	case "anticipated":
		return fetchAnticipatedMovies
	case "box_office":
		return fetchBoxOfficeMovies
	default:
		log.Warnf("Unknown movie job type: %s", jobType)
		return nil
	}
}

// getShowFetcher returns the appropriate fetcher function for a show job type
func (e *DynamicJobExecutor) getShowFetcher(jobType string) ShowFetcher {
	switch jobType {
	case "trending":
		return fetchTrendingShows
	case "popular", "smart_popular":
		return fetchPopularShows
	case "watched":
		return fetchWatchedShows
	case "collected":
		return fetchCollectedShows
	case "favorited":
		return fetchFavoritedShows
	case "played":
		return fetchPlayedShows
	case "anticipated":
		return fetchAnticipatedShows
	default:
		log.Warnf("Unknown show job type: %s", jobType)
		return nil
	}
}

// RunDynamicJob is a convenience function to run a dynamic job
func RunDynamicJob(cfg *config.Config, db *database.Database, job config.DynamicJob, dryRun bool) error {
	executor := &DynamicJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}
	return executor.Execute(context.Background(), job)
}
