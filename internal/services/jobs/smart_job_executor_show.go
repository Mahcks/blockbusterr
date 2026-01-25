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

// SmartShowJobExecutor handles execution of smart popular show jobs with adaptive rating thresholds
type SmartShowJobExecutor struct {
	Config        *config.Config
	Database      *database.Database
	DryRun        bool
	lastDecisions *JobRunDecisions
}

// Execute runs a smart show job with adaptive rating thresholds
func (e *SmartShowJobExecutor) Execute(
	ctx context.Context,
	jobConfig SmartJobConfig,
	fetcher ShowFetcher,
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

	// Fetch shows using the provided fetcher
	shows, err := fetcher(ctx, traktClient, jobConfig.Limit, "")
	if err != nil {
		log.Errorf("Failed to fetch %s from Trakt: %v", jobConfig.JobName, err)
		return
	}

	log.Infof("Found %d shows from Trakt for %s", len(shows), jobConfig.JobName)

	// Calculate popularity percentiles for this set
	percentiles := filters.CalculateShowPopularityPercentiles(shows)

	// Initialize decision tracking
	runDecisions := &JobRunDecisions{
		JobName:    jobConfig.JobName,
		RunTime:    time.Now(),
		TotalFound: len(shows),
		Decisions:  make([]ContentDecision, 0),
	}
	e.lastDecisions = runDecisions

	// Apply adaptive filters with decision tracking
	filteredShows, scoreMap, showDecisions := e.evaluateShowsWithAdaptiveFilters(
		shows,
		percentiles,
		jobConfig,
	)
	runDecisions.Decisions = showDecisions
	runDecisions.PassedFilters = len(filteredShows)

	// Convert to regular JobConfig for execution
	regularJobConfig := JobConfig{
		JobName:   jobConfig.JobName,
		MediaType: jobConfig.MediaType,
		Mode:      jobConfig.Mode,
		Limit:     jobConfig.Limit,
	}

	// Route to appropriate handler based on mode
	showExecutor := &ShowJobExecutor{
		Config:        e.Config,
		Database:      e.Database,
		DryRun:        e.DryRun,
		lastDecisions: runDecisions,
	}

	if jobConfig.Mode == "jellyseerr" {
		showExecutor.executeShowsJellyseerr(ctx, regularJobConfig, filteredShows, scoreMap)
	} else {
		showExecutor.executeShowsDirect(ctx, regularJobConfig, filteredShows, scoreMap)
	}

	runDecisions.Completed = true
}

// evaluateShowsWithAdaptiveFilters evaluates shows with adaptive rating thresholds
func (e *SmartShowJobExecutor) evaluateShowsWithAdaptiveFilters(
	shows []integrations.Show,
	percentiles map[int]float64,
	jobConfig SmartJobConfig,
) ([]integrations.Show, map[int]ScoreInfo, []ContentDecision) {
	decisions := make([]ContentDecision, 0, len(shows))
	passedShows := make([]integrations.Show, 0)

	for _, show := range shows {
		decision := ContentDecision{
			Title:       show.Title,
			Year:        show.Year,
			MediaType:   "tv",
			TMDBID:      show.IDs.TMDB,
			TVDBID:      show.IDs.TVDB,
			IMDBID:      show.IDs.IMDB,
			Rating:      show.Rating,
			Votes:       show.Votes,
			EvaluatedAt: time.Now(),
		}

		// Get poster URL
		if show.IDs.TMDB > 0 {
			decision.PosterURL = GetTMDBPosterURL(e.Config, show.IDs.TMDB, "tv")
		}

		// Get popularity percentile for this show
		percentile, hasPercentile := percentiles[show.IDs.TMDB]
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
		filterResult := filters.ShowPassesAdaptiveFilters(
			show,
			e.Config.Filters.Shows,
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
			passedShows = append(passedShows, show)
			decision.Action = "passed_filters"
			decision.ActionReason = fmt.Sprintf(
				"Passed adaptive filters (percentile: %.0f%%, threshold: %.1f)",
				percentile*100,
				adaptiveMinRating,
			)

			log.Debugf("'%s (%d)' - PASS (rating: %.1f, popularity: %.0f%%, threshold: %.1f)",
				show.Title, show.Year, show.Rating, percentile*100, adaptiveMinRating)
		} else {
			decision.Action = "rejected"
			decision.ActionReason = fmt.Sprintf(
				"%s (percentile: %.0f%%, threshold: %.1f)",
				filterResult.Reason,
				percentile*100,
				adaptiveMinRating,
			)

			log.Debugf("'%s (%d)' - REJECT: %s (popularity: %.0f%%, threshold: %.1f)",
				show.Title, show.Year, filterResult.Reason, percentile*100, adaptiveMinRating)
		}

		decisions = append(decisions, decision)
	}

	// Calculate scores and ranks
	scoreMap := ScoreAndRankShows(passedShows, e.Config)

	// Update decisions with scores
	for i := range decisions {
		if decisions[i].PassedFilters {
			if scoreInfo, ok := scoreMap[decisions[i].TVDBID]; ok {
				decisions[i].Score = scoreInfo.Score
				decisions[i].Rank = scoreInfo.Rank
			}
		}
	}

	// Log rejected items to database
	if e.Database != nil {
		for _, decision := range decisions {
			if !decision.PassedFilters {
				posterURL := GetTMDBPosterURL(e.Config, decision.TMDBID, "tv")
				err := e.Database.LogActivity(database.ActivityLog{
					Timestamp:     time.Now(),
					JobType:       jobConfig.JobName,
					MediaType:     "tv",
					Title:         decision.Title,
					Year:          decision.Year,
					TMDBID:        decision.TMDBID,
					TVDBID:        decision.TVDBID,
					IMDBID:        decision.IMDBID,
					PosterURL:     posterURL,
					Score:         decision.Score,
					Rank:          0,
					Status:        "rejected",
					Message:       decision.ActionReason,
					FilterDetails: FilterChecksToJSON(decision.FilterChecks),
				})
				if err != nil {
					slog.Error("Failed to log activity for show", "title", decision.Title, "year", decision.Year, "err", err)
				}
			}
		}
	}

	log.Infof("Adaptive filter evaluation: %d found, %d passed filters, %d rejected",
		len(shows), len(passedShows), len(shows)-len(passedShows))

	return passedShows, scoreMap, decisions
}
