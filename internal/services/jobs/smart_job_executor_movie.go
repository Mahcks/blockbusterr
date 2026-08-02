package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/filters"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// SmartJobConfig extends JobConfig with adaptive rating parameters
type SmartJobConfig struct {
	JobID               string
	JobName             string
	Source              string
	MediaType           string
	Mode                string
	MinimumAvailability string // Radarr only
	Monitor             string // For Radarr: "movieOnly", "movieAndCollection", "none"; For Sonarr: "all", "future", "missing", "existing", "pilot", "firstSeason", "latestSeason", "none"
	Limit               int
	BaseMinRating       float64
	AdjustmentFactor    float64
}

// SmartMovieJobExecutor handles execution of smart popular movie jobs with adaptive rating thresholds
type SmartMovieJobExecutor struct {
	Config        *config.Config
	Database      *database.Database
	DryRun        bool
	lastDecisions *JobRunDecisions
	currentRunID  int64
}

// Execute runs a smart movie job with adaptive rating thresholds
func (e *SmartMovieJobExecutor) Execute(
	ctx context.Context,
	jobConfig SmartJobConfig,
	fetcher MovieFetcher,
) error {
	if e.DryRun {
		log.Infof("Starting %s job (DRY RUN) with adaptive ratings", jobConfig.JobName)
	} else {
		log.Infof("Starting %s job with adaptive ratings", jobConfig.JobName)
	}

	if e.Database != nil {
		runID, err := e.Database.StartJobRun(jobConfig.JobID, jobConfig.JobName, jobConfig.MediaType, jobConfig.Mode, time.Now())
		if err != nil {
			log.Warnf("Failed to create job run for %s: %v", jobConfig.JobName, err)
		} else {
			e.currentRunID = runID
			defer func() { e.currentRunID = 0 }()
		}
	}

	discoveryClient, err := NewDiscoveryClient(e.Config, jobConfig.Source)
	if err != nil {
		log.Errorf("Failed to configure discovery source for %s: %v", jobConfig.JobName, err)
		return err
	}

	// Fetch movies using the provided fetcher
	movies, err := fetcher(ctx, discoveryClient, jobConfig.Limit, "")
	if err != nil {
		log.Errorf("Failed to fetch %s from %s: %v", jobConfig.JobName, discoveryClient.Source(), err)
		if e.Database != nil {
			_ = e.Database.LogActivity(database.ActivityLog{
				Timestamp: time.Now(),
				JobID:     jobConfig.JobID,
				RunID:     e.currentRunID,
				JobType:   jobConfig.JobName,
				MediaType: "movie",
				Title:     FormatJobLabel(jobConfig.JobID, jobConfig.JobName),
				Status:    "failed",
				Message:   fmt.Sprintf("Failed to fetch from %s: %v", discoveryClient.Source(), err),
			})
		}
		if e.Database != nil && e.currentRunID > 0 {
			_ = e.Database.CompleteJobRun(e.currentRunID, time.Now(), "failed", 0, 0, 0, 0, 0, 0, 1, err.Error())
		}
		return err
	}

	log.Infof("Found %d movies from %s for %s", len(movies), discoveryClient.Source(), jobConfig.JobName)

	// Calculate popularity percentiles for this set
	percentiles := filters.CalculateMoviePopularityPercentiles(movies)

	// Initialize decision tracking
	runDecisions := &JobRunDecisions{
		JobName:    jobConfig.JobName,
		RunTime:    time.Now(),
		TotalFound: len(movies),
		Decisions:  make([]ContentDecision, 0),
	}
	e.lastDecisions = runDecisions

	// Apply adaptive filters with decision tracking
	filteredMovies, scoreMap, movieDecisions := e.evaluateMoviesWithAdaptiveFilters(
		ctx,
		movies,
		percentiles,
		jobConfig,
	)
	runDecisions.Decisions = movieDecisions
	runDecisions.PassedFilters = len(filteredMovies)

	// Convert to regular JobConfig for execution
	regularJobConfig := JobConfig{
		JobID:               jobConfig.JobID,
		JobName:             jobConfig.JobName,
		Source:              jobConfig.Source,
		MediaType:           jobConfig.MediaType,
		Mode:                jobConfig.Mode,
		MinimumAvailability: jobConfig.MinimumAvailability,
		Monitor:             jobConfig.Monitor,
		Limit:               jobConfig.Limit,
	}

	// Route to appropriate handler based on mode
	movieExecutor := &MovieJobExecutor{
		Config:        e.Config,
		Database:      e.Database,
		DryRun:        e.DryRun,
		lastDecisions: runDecisions,
		currentRunID:  e.currentRunID,
	}

	var executionErr error
	if ctx.Err() != nil {
		executionErr = ctx.Err()
	} else if jobConfig.Mode == "jellyseerr" {
		movieExecutor.executeMoviesJellyseerr(ctx, regularJobConfig, filteredMovies, scoreMap)
		executionErr = ctx.Err()
	} else {
		executionErr = movieExecutor.executeMoviesDirect(ctx, regularJobConfig, filteredMovies, scoreMap)
	}
	if executionErr != nil {
		if e.Database != nil && e.currentRunID > 0 {
			_ = e.Database.CompleteJobRun(e.currentRunID, time.Now(), "failed", runDecisions.TotalFound, runDecisions.PassedFilters, 0, 0, 0, runDecisions.Rejected, 1, executionErr.Error())
		}
		return executionErr
	}

	runDecisions.Rejected = runDecisions.TotalFound - runDecisions.PassedFilters
	runDecisions.Completed = true
	if e.Database != nil && e.currentRunID > 0 {
		_ = e.Database.CompleteJobRun(
			e.currentRunID,
			time.Now(),
			"completed",
			runDecisions.TotalFound,
			runDecisions.PassedFilters,
			runDecisions.Added,
			runDecisions.Requested,
			runDecisions.Skipped,
			runDecisions.Rejected,
			runDecisions.Failed,
			"",
		)
	}
	return nil
}

// evaluateMoviesWithAdaptiveFilters evaluates movies with adaptive rating thresholds
func (e *SmartMovieJobExecutor) evaluateMoviesWithAdaptiveFilters(
	ctx context.Context,
	movies []integrations.Movie,
	percentiles map[int]float64,
	jobConfig SmartJobConfig,
) ([]integrations.Movie, map[int]ScoreInfo, []ContentDecision) {
	decisions := make([]ContentDecision, 0, len(movies))
	passedMovies := make([]integrations.Movie, 0)

	for _, movie := range movies {
		if ctx.Err() != nil {
			break
		}
		decision := ContentDecision{
			Title:       movie.Title,
			Year:        movie.Year,
			Language:    movie.Language,
			MediaType:   "movie",
			TMDBID:      movie.IDs.TMDB,
			IMDBID:      movie.IDs.IMDB,
			Rating:      movie.Rating,
			Votes:       movie.Votes,
			EvaluatedAt: time.Now(),
		}

		// Get poster URL
		if movie.IDs.TMDB > 0 {
			decision.PosterURL = GetTMDBPosterURL(e.Config, movie.IDs.TMDB, "movie")
		}

		// Get popularity percentile for this movie
		percentile, hasPercentile := percentiles[movie.IDs.TMDB]
		if !hasPercentile {
			percentile = 0.5 // Default to middle if not found
		}

		// Calculate adaptive rating threshold
		adaptiveMinRating := filters.CalculateAdaptiveRating(
			jobConfig.BaseMinRating,
			percentile,
			jobConfig.AdjustmentFactor,
		)

		// Run through adaptive filters
		filterResult := filters.MoviePassesAdaptiveFilters(
			movie,
			e.Config.Filters.Movies,
			adaptiveMinRating,
		)
		decision.PassedFilters = filterResult.Passed

		// Convert filter checks
		decision.FilterChecks = make([]FilterCheck, len(filterResult.Checks))
		for i, check := range filterResult.Checks {
			decision.FilterChecks[i] = FilterCheck{
				Name:    check.Name,
				Passed:  check.Passed,
				Message: check.Message,
			}
		}

		if filterResult.Passed {
			passedMovies = append(passedMovies, movie)
			decision.Action = "passed_filters"
			decision.ActionReason = fmt.Sprintf(
				"Passed adaptive filters (percentile: %.0f%%, threshold: %.1f)",
				percentile*100,
				adaptiveMinRating,
			)

			log.Debugf("'%s (%d)' - PASS (rating: %.1f, popularity: %.0f%%, threshold: %.1f)",
				movie.Title, movie.Year, movie.Rating, percentile*100, adaptiveMinRating)
		} else {
			decision.Action = "rejected"
			decision.ActionReason = fmt.Sprintf(
				"%s (percentile: %.0f%%, threshold: %.1f)",
				filters.Explain(filterResult),
				percentile*100,
				adaptiveMinRating,
			)

			log.Debugf("'%s (%d)' - REJECT: %s (popularity: %.0f%%, threshold: %.1f)",
				movie.Title, movie.Year, filterResult.Reason, percentile*100, adaptiveMinRating)
		}

		decisions = append(decisions, decision)
	}

	// Calculate scores and ranks
	scoreMap := ScoreAndRankMovies(passedMovies, e.Config)

	// Update decisions with scores
	for i := range decisions {
		if decisions[i].PassedFilters {
			if scoreInfo, ok := scoreMap[decisions[i].TMDBID]; ok {
				decisions[i].Score = scoreInfo.Score
				decisions[i].Rank = scoreInfo.Rank
			}
		}
	}

	// Log rejected items to database
	if e.Database != nil {
		for _, decision := range decisions {
			if !decision.PassedFilters {
				posterURL := GetTMDBPosterURL(e.Config, decision.TMDBID, "movie")
				err := e.Database.LogActivity(database.ActivityLog{
					Timestamp:     time.Now(),
					JobID:         jobConfig.JobID,
					RunID:         e.currentRunID,
					JobType:       jobConfig.JobName,
					MediaType:     "movie",
					Title:         decision.Title,
					Year:          decision.Year,
					Language:      decision.Language,
					TMDBID:        decision.TMDBID,
					IMDBID:        decision.IMDBID,
					PosterURL:     posterURL,
					Score:         decision.Score,
					Rank:          0,
					Status:        "rejected",
					Message:       decision.ActionReason,
					FilterDetails: FilterChecksToJSON(decision.FilterChecks),
				})
				if err != nil {
					slog.Error("Failed to log activity for movie", "title", decision.Title, "year", decision.Year, "err", err)
				}
			}
		}
	}

	log.Infof("Adaptive filter evaluation: %d found, %d passed filters, %d rejected",
		len(movies), len(passedMovies), len(movies)-len(passedMovies))

	return passedMovies, scoreMap, decisions
}
