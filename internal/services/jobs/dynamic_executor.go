package jobs

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

// DynamicJobExecutor executes jobs based on DynamicJob configuration
type DynamicJobExecutor struct {
	Config      *config.Config
	Database    *database.Database
	DryRun      bool
	ListSources ListSourceRegistry
}

// Execute runs a dynamic job, routing to the appropriate executor based on job type and media
func (e *DynamicJobExecutor) Execute(ctx context.Context, job config.DynamicJob) error {
	effectiveConfig, err := configForJob(e.Config, job)
	if err != nil {
		return err
	}
	effectiveExecutor := *e
	effectiveExecutor.Config = effectiveConfig
	e = &effectiveExecutor

	// Validate job type
	typeDef, ok := GetJobTypeDefinition(job.Type)
	if !ok {
		return fmt.Errorf("unknown job type: %s", job.Type)
	}

	// Validate media type support
	if !SupportsMediaType(job.Type, job.MediaType) {
		return fmt.Errorf("job type %s does not support media type %s", job.Type, job.MediaType)
	}
	if job.Type != string(enums.JobTypeList) && !SupportsSource(job.Type, job.Source) {
		return fmt.Errorf("job type %s does not support source %s", job.Type, job.Source)
	}
	discovery, err := e.discoveryForJob(job)
	if err != nil {
		return err
	}
	if job.Source == "simkl" && job.Type == "watched" && job.Period != "weekly" && job.Period != "monthly" {
		return fmt.Errorf("simkl most watched jobs support weekly or monthly periods")
	}
	if job.Source == "simkl" && job.Limit > 500 {
		return fmt.Errorf("simkl jobs cannot exceed 500 items")
	}

	// Determine effective mode
	mode := DetermineMode(job.Mode, e.Config.Jobs.Mode)

	// Build JobConfig from DynamicJob
	jobName := job.Name
	if job.Source == "simkl" && !strings.Contains(strings.ToLower(jobName), "simkl") {
		jobName += " (Simkl)"
	}
	jobConfig := JobConfig{
		JobID:               job.ID,
		JobName:             jobName,
		Source:              job.Source,
		MediaType:           job.MediaType,
		Mode:                mode,
		MinimumAvailability: job.MinimumAvailability,
		Monitor:             job.Monitor,
		Limit:               job.Limit,
		DeliveryLimit:       job.DeliveryLimit,
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
		return e.executeMovieJob(ctx, job, jobConfig, discovery)
	case "show":
		return e.executeShowJob(ctx, job, jobConfig, discovery)
	default:
		return fmt.Errorf("unsupported media type: %s", job.MediaType)
	}
}

func configForJob(cfg *config.Config, job config.DynamicJob) (*config.Config, error) {
	_, rules, err := cfg.ResolveRuleSet(job)
	if err != nil {
		return nil, err
	}
	effective := *cfg
	effective.Filters = rules
	return &effective, nil
}

// executeMovieJob executes a movie job using the appropriate executor
func (e *DynamicJobExecutor) executeMovieJob(ctx context.Context, job config.DynamicJob, jobConfig JobConfig, discovery *DiscoveryClient) error {
	// Get the appropriate fetcher based on job type
	fetcher := e.getMovieFetcher(job)
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
			Source:              jobConfig.Source,
			MediaType:           jobConfig.MediaType,
			Mode:                jobConfig.Mode,
			MinimumAvailability: jobConfig.MinimumAvailability,
			Monitor:             jobConfig.Monitor,
			Limit:               jobConfig.Limit,
			DeliveryLimit:       jobConfig.DeliveryLimit,
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

		return executor.Execute(ctx, smartConfig, fetcher)
	} else {
		// Use standard MovieJobExecutor
		executor := &MovieJobExecutor{
			Config:    e.Config,
			Database:  e.Database,
			DryRun:    e.DryRun,
			discovery: discovery,
		}
		return executor.Execute(ctx, jobConfig, fetcher)
	}
}

// executeShowJob executes a show job using the appropriate executor
func (e *DynamicJobExecutor) executeShowJob(ctx context.Context, job config.DynamicJob, jobConfig JobConfig, discovery *DiscoveryClient) error {
	// Get the appropriate fetcher based on job type
	fetcher := e.getShowFetcher(job)
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
			Source:           jobConfig.Source,
			MediaType:        jobConfig.MediaType,
			Mode:             jobConfig.Mode,
			Monitor:          jobConfig.Monitor,
			Limit:            jobConfig.Limit,
			DeliveryLimit:    jobConfig.DeliveryLimit,
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

		return executor.Execute(ctx, smartConfig, fetcher)
	} else {
		// Use standard ShowJobExecutor
		executor := &ShowJobExecutor{
			Config:    e.Config,
			Database:  e.Database,
			DryRun:    e.DryRun,
			discovery: discovery,
		}
		return executor.Execute(ctx, jobConfig, fetcher)
	}
}

// getMovieFetcher returns the appropriate fetcher function for a movie job type
func (e *DynamicJobExecutor) getMovieFetcher(job config.DynamicJob) MovieFetcher {
	switch job.Type {
	case string(enums.JobTypeList):
		return func(ctx context.Context, discovery *DiscoveryClient, limit int, _ string) ([]integrations.Movie, error) {
			return discovery.GetListMovies(ctx, *job.List, limit)
		}
	case "trending":
		return func(ctx context.Context, discovery *DiscoveryClient, limit int, _ string) ([]integrations.Movie, error) {
			trendingMovies, err := discovery.GetTrendingMovies(ctx, limit)
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
		log.Warnf("Unknown movie job type: %s", job.Type)
		return nil
	}
}

// getShowFetcher returns the appropriate fetcher function for a show job type
func (e *DynamicJobExecutor) getShowFetcher(job config.DynamicJob) ShowFetcher {
	switch job.Type {
	case string(enums.JobTypeList):
		return func(ctx context.Context, discovery *DiscoveryClient, limit int, _ string) ([]integrations.Show, error) {
			return discovery.GetListShows(ctx, *job.List, limit)
		}
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
		log.Warnf("Unknown show job type: %s", job.Type)
		return nil
	}
}

func (e *DynamicJobExecutor) discoveryForJob(job config.DynamicJob) (*DiscoveryClient, error) {
	if job.Type != string(enums.JobTypeList) {
		return NewDiscoveryClient(e.Config, job.Source)
	}
	if job.List == nil {
		return nil, fmt.Errorf("list locator is required")
	}
	if err := ValidateListSourceLocator(job.Source, *job.List); err != nil {
		return nil, err
	}
	adapter := e.ListSources[job.Source]
	if e.ListSources == nil {
		var err error
		adapter, err = configuredListSource(e.Config, job.Source)
		if err != nil {
			return nil, err
		}
	}
	if adapter == nil {
		return nil, fmt.Errorf("%s list adapter is unavailable", job.Source)
	}
	return newListDiscoveryClient(job.Source, adapter), nil
}

// RunDynamicJob is a convenience function to run a dynamic job
func RunDynamicJob(ctx context.Context, cfg *config.Config, db *database.Database, job config.DynamicJob, dryRun bool) error {
	executor := &DynamicJobExecutor{
		Config:   cfg,
		Database: db,
		DryRun:   dryRun,
	}
	return executor.Execute(ctx, job)
}
