package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/filters"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// SmartJobConfig extends JobConfig with adaptive rating parameters
type SmartJobConfig struct {
	JobName          string
	MediaType        string
	Mode             string
	Limit            int
	BaseMinRating    float64
	AdjustmentFactor float64
}

// SmartMovieJobExecutor handles execution of smart popular movie jobs with adaptive rating thresholds
type SmartMovieJobExecutor struct {
	Config        *config.Config
	Database      *database.Database
	DryRun        bool
	lastDecisions *JobRunDecisions
}

// Execute runs a smart movie job with adaptive rating thresholds
func (e *SmartMovieJobExecutor) Execute(
	ctx context.Context,
	jobConfig SmartJobConfig,
	fetcher MovieFetcher,
) {
	if e.DryRun {
		log.Infof("Starting %s job (DRY RUN) with adaptive ratings", jobConfig.JobName)
	} else {
		log.Infof("Starting %s job with adaptive ratings", jobConfig.JobName)
	}

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     e.Config.Trakt.ClientID,
		ClientSecret: e.Config.Trakt.ClientSecret,
	})

	// Fetch movies using the provided fetcher
	movies, err := fetcher(ctx, traktClient, jobConfig.Limit, "")
	if err != nil {
		log.Errorf("Failed to fetch %s from Trakt: %v", jobConfig.JobName, err)
		return
	}

	log.Infof("Found %d movies from Trakt for %s", len(movies), jobConfig.JobName)

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
		movies,
		percentiles,
		jobConfig,
	)
	runDecisions.Decisions = movieDecisions
	runDecisions.PassedFilters = len(filteredMovies)

	// Convert to regular JobConfig for execution
	regularJobConfig := JobConfig{
		JobName:   jobConfig.JobName,
		MediaType: jobConfig.MediaType,
		Mode:      jobConfig.Mode,
		Limit:     jobConfig.Limit,
	}

	// Route to appropriate handler based on mode
	movieExecutor := &MovieJobExecutor{
		Config:        e.Config,
		Database:      e.Database,
		DryRun:        e.DryRun,
		lastDecisions: runDecisions,
	}

	if jobConfig.Mode == "jellyseerr" {
		movieExecutor.executeMoviesJellyseerr(ctx, regularJobConfig, filteredMovies, scoreMap)
	} else {
		movieExecutor.executeMoviesDirect(ctx, regularJobConfig, filteredMovies, scoreMap)
	}

	runDecisions.Completed = true
}

// evaluateMoviesWithAdaptiveFilters evaluates movies with adaptive rating thresholds
func (e *SmartMovieJobExecutor) evaluateMoviesWithAdaptiveFilters(
	movies []integrations.Movie,
	percentiles map[int]float64,
	jobConfig SmartJobConfig,
) ([]integrations.Movie, map[int]ScoreInfo, []ContentDecision) {
	decisions := make([]ContentDecision, 0, len(movies))
	passedMovies := make([]integrations.Movie, 0)

	for _, movie := range movies {
		decision := ContentDecision{
			Title:       movie.Title,
			Year:        movie.Year,
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
				filterResult.Reason,
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
				e.Database.LogActivity(database.ActivityLog{
					Timestamp:     time.Now(),
					JobType:       jobConfig.JobName,
					MediaType:     "movie",
					Title:         decision.Title,
					Year:          decision.Year,
					TMDBID:        decision.TMDBID,
					IMDBID:        decision.IMDBID,
					PosterURL:     posterURL,
					Score:         decision.Score,
					Rank:          0,
					Status:        "rejected",
					Message:       decision.ActionReason,
					FilterDetails: FilterChecksToJSON(decision.FilterChecks),
				})
			}
		}
	}

	log.Infof("Adaptive filter evaluation: %d found, %d passed filters, %d rejected",
		len(movies), len(passedMovies), len(movies)-len(passedMovies))

	return passedMovies, scoreMap, decisions
}

// GetLastDecisions returns the decisions from the last job run
func (e *SmartMovieJobExecutor) GetLastDecisions() *JobRunDecisions {
	return e.lastDecisions
}
