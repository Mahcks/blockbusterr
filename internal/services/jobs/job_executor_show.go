package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/filters"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// ShowJobExecutor handles execution of show jobs with unified logic
type ShowJobExecutor struct {
	Config        *config.Config
	Database      *database.Database
	DryRun        bool
	lastDecisions *JobRunDecisions // Store last run decisions for API access
}

// Execute runs a show job with the given configuration and fetcher function
func (e *ShowJobExecutor) Execute(
	ctx context.Context,
	jobConfig JobConfig,
	fetcher ShowFetcher,
) {
	if e.DryRun {
		log.Infof("Starting %s job (DRY RUN)", jobConfig.JobName)
	} else {
		log.Infof("Starting %s job", jobConfig.JobName)
	}

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     e.Config.Trakt.ClientID,
		ClientSecret: e.Config.Trakt.ClientSecret,
	})

	// Fetch shows using the provided fetcher
	shows, err := fetcher(ctx, traktClient, jobConfig.Limit, jobConfig.Period)
	if err != nil {
		log.Errorf("Failed to fetch %s from Trakt: %v", jobConfig.JobName, err)
		return
	}

	log.Infof("Found %d shows from Trakt for %s", len(shows), jobConfig.JobName)

	// Initialize decision tracking for this run
	runDecisions := &JobRunDecisions{
		JobName:    jobConfig.JobName,
		RunTime:    time.Now(),
		TotalFound: len(shows),
		Decisions:  make([]ContentDecision, 0),
	}
	e.lastDecisions = runDecisions

	// Apply filters with detailed decision tracking
	filteredShows, scoreMap, showDecisions := e.evaluateShowsWithDecisions(shows, jobConfig)
	runDecisions.Decisions = showDecisions
	runDecisions.PassedFilters = len(filteredShows)

	// Route to appropriate handler based on mode
	if jobConfig.Mode == "jellyseerr" {
		e.executeShowsJellyseerr(ctx, jobConfig, filteredShows, scoreMap)
	} else {
		e.executeShowsDirect(ctx, jobConfig, filteredShows, scoreMap)
	}

	// Mark run as completed and update final stats
	runDecisions.Completed = true
}

// executeShowsDirect adds shows directly to Sonarr
func (e *ShowJobExecutor) executeShowsDirect(
	ctx context.Context,
	jobConfig JobConfig,
	shows []integrations.Show,
	scoreMap map[int]ScoreInfo,
) {
	sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
		BaseURL: e.Config.Sonarr.URL,
		APIKey:  e.Config.Sonarr.APIKey,
	})

	// Get existing series from Sonarr for deduplication
	existingSeries, err := sonarrClient.GetSeries(ctx)
	if err != nil {
		log.Errorf("Failed to fetch existing series from Sonarr: %v", err)
		return
	}

	// Create a map of existing TVDB IDs for fast lookup
	existingTVDBIDs := make(map[int]bool)
	for _, series := range existingSeries {
		existingTVDBIDs[series.TvdbID] = true
	}

	// Add new shows to Sonarr
	added := 0
	skipped := 0
	failed := 0

	for _, show := range shows {
		// Try to lookup series in Sonarr - prefer TVDB ID if available, otherwise use title
		var lookupResults []integrations.SonarrSeries
		var err error

		if show.IDs.TVDB > 0 {
			lookupResults, err = sonarrClient.LookupSeries(ctx, fmt.Sprintf("tvdb:%d", show.IDs.TVDB))
			if err != nil || len(lookupResults) == 0 {
				log.Warnf("TVDB lookup failed for '%s (%d)', trying title search", show.Title, show.Year)
				lookupResults, err = sonarrClient.LookupSeries(ctx, show.Title)
			}
		} else {
			lookupResults, err = sonarrClient.LookupSeries(ctx, show.Title)
		}

		if err != nil || len(lookupResults) == 0 {
			log.Errorf("Failed to lookup show '%s (%d)' in Sonarr: %v", show.Title, show.Year, err)
			failed++
			continue
		}

		series := lookupResults[0]

		// Check if lookup returned a series that's already in Sonarr (has an ID assigned)
		if series.ID > 0 {
			log.Debugf("Skipping '%s (%d)' - already in Sonarr (ID: %d)", show.Title, show.Year, series.ID)
			skipped++
			continue
		}

		// Double-check against cached TVDB IDs
		if series.TvdbID > 0 && existingTVDBIDs[series.TvdbID] {
			log.Debugf("Skipping '%s (%d)' - already in Sonarr (TVDB: %d)", show.Title, show.Year, series.TvdbID)
			skipped++
			continue
		}

		// Warn if TVDB ID is missing
		if series.TvdbID == 0 {
			log.Warnf("Show '%s (%d)' has no TVDB ID, may cause issues", show.Title, show.Year)
		}

		monitor := jobConfig.Monitor
		if monitor == "" {
			monitor = e.Config.Sonarr.Monitor
		}
		if monitor == "" {
			monitor = "all"
		}

		// Configure series for Sonarr
		series.QualityProfileID = e.Config.Sonarr.QualityProfile
		series.Monitored = true
		series.RootFolderPath = e.Config.Sonarr.RootFolder
		series.AddOptions = &integrations.SonarrAddOptions{
			SearchForMissingEpisodes: true,
			Monitor:                  monitor,
		}

		// Add series to Sonarr (or simulate in dry-run mode)
		if e.DryRun {
			log.Infof("[DRY RUN] Would add %s show '%s (%d)' to Sonarr", jobConfig.JobName, show.Title, show.Year)
			added++
		} else {
			addedSeries, err := sonarrClient.AddSeries(ctx, series)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Show '%s (%d)' already exists in Sonarr", show.Title, show.Year)
					skipped++
				} else {
					log.Errorf("Failed to add show '%s (%d)' to Sonarr: %v", show.Title, show.Year, err)
					failed++
					// Log failed activity
					if e.Database != nil {
						scoreInfo := scoreMap[series.TvdbID]
						err := e.Database.LogActivity(database.ActivityLog{
							Timestamp: time.Now(),
							JobType:   jobConfig.JobName,
							MediaType: "show",
							Title:     show.Title,
							Year:      show.Year,
							TVDBID:    series.TvdbID,
							IMDBID:    show.IDs.IMDB,
							Score:     scoreInfo.Score,
							Rank:      scoreInfo.Rank,
							Status:    "failed",
							Message:   err.Error(),
						})
						if err != nil {
							slog.Error("Failed to log activity for show", "title", show.Title, "year", show.Year, "err", err)
						}
					}
				}
				continue
			}

			log.Infof("Added %s show '%s (%d)' to Sonarr (ID: %d)", jobConfig.JobName, addedSeries.Title, addedSeries.Year, addedSeries.ID)
			added++

			// Log successful activity
			if e.Database != nil {
				posterURL := GetShowPosterURL(e.Config, addedSeries.TmdbID, addedSeries.TvdbID)
				scoreInfo := scoreMap[addedSeries.TvdbID]
				err := e.Database.LogActivity(database.ActivityLog{
					Timestamp: time.Now(),
					JobType:   jobConfig.JobName,
					MediaType: "show",
					Title:     addedSeries.Title,
					Year:      addedSeries.Year,
					TVDBID:    addedSeries.TvdbID,
					IMDBID:    addedSeries.ImdbID,
					PosterURL: posterURL,
					Score:     scoreInfo.Score,
					Rank:      scoreInfo.Rank,
					Status:    "added",
				})
				if err != nil {
					slog.Error("Failed to log activity for show", "title", addedSeries.Title, "year", addedSeries.Year, "err", err)
				}
			}
		}

		// Mark as existing to avoid duplicates within this job run
		existingTVDBIDs[series.TvdbID] = true
	}

	log.Infof("%s job completed - Added: %d, Skipped: %d, Failed: %d", jobConfig.JobName, added, skipped, failed)
}

// executeShowsJellyseerr requests shows via Jellyseerr
func (e *ShowJobExecutor) executeShowsJellyseerr(
	_ context.Context,
	jobConfig JobConfig,
	shows []integrations.Show,
	scoreMap map[int]ScoreInfo,
) {
	jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
		URL:             e.Config.Jellyseerr.URL,
		APIKey:          e.Config.Jellyseerr.APIKey,
		UserID:          e.Config.Jellyseerr.UserID,
		RequestEmail:    e.Config.Jellyseerr.RequestCredentials.Email,
		RequestPassword: e.Config.Jellyseerr.RequestCredentials.Password,
	})

	// Request new shows via Jellyseerr
	requested := 0
	skipped := 0
	failed := 0

	for _, show := range shows {
		// Jellyseerr uses TMDB IDs for TV shows, not TVDB
		// Check if show already exists in Jellyseerr
		mediaInfo, err := jellyseerrClient.GetShowInfo(show.IDs.TMDB)
		if err != nil {
			log.Debugf("Failed to check Jellyseerr status for '%s (%d)': %v", show.Title, show.Year, err)
		} else if mediaInfo.HasMediaInfo() {
			log.Debugf("Skipping '%s (%d)' - already requested/available in Jellyseerr", show.Title, show.Year)
			e.updateDecisionOutcome(show.IDs.TVDB, "skipped", "Already in Jellyseerr")
			skipped++
			continue
		}

		// Request show via Jellyseerr (or simulate in dry-run mode)
		if e.DryRun {
			log.Infof("[DRY RUN] Would request %s show '%s (%d)' via Jellyseerr", jobConfig.JobName, show.Title, show.Year)
			e.updateDecisionOutcome(show.IDs.TVDB, "requested", "[DRY RUN] Would be requested")
			requested++
		} else {
			result, err := jellyseerrClient.RequestShow(show.IDs.TMDB)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") || strings.Contains(err.Error(), "requested") {
					log.Debugf("Show '%s (%d)' already requested in Jellyseerr", show.Title, show.Year)
					e.updateDecisionOutcome(show.IDs.TVDB, "skipped", "Already requested")
					skipped++
				} else {
					log.Errorf("Failed to request show '%s (%d)' via Jellyseerr: %v", show.Title, show.Year, err)
					e.updateDecisionOutcome(show.IDs.TVDB, "failed", err.Error())
					failed++
					// Log failure to database
					if e.Database != nil {
						scoreInfo := scoreMap[show.IDs.TVDB]
						filterDetails := e.getFilterDetailsForShow(show.IDs.TVDB)
						err := e.Database.LogActivity(database.ActivityLog{
							Timestamp:     time.Now(),
							JobType:       jobConfig.JobName,
							MediaType:     "show",
							Title:         show.Title,
							Year:          show.Year,
							TVDBID:        show.IDs.TVDB,
							IMDBID:        show.IDs.IMDB,
							Score:         scoreInfo.Score,
							Rank:          scoreInfo.Rank,
							Status:        "failed",
							Message:       err.Error(),
							FilterDetails: filterDetails,
						})
						if err != nil {
							slog.Error("Failed to log activity for show", "title", show.Title, "year", show.Year, "err", err)
						}
					}
				}
				continue
			}

			if result.IsAlreadyRequested() {
				log.Debugf("Show '%s (%d)' already requested in Jellyseerr", show.Title, show.Year)
				e.updateDecisionOutcome(show.IDs.TVDB, "skipped", "Already requested")
				skipped++
			} else {
				log.Infof("Requested %s show '%s (%d)' via Jellyseerr (Request ID: %d)", jobConfig.JobName, show.Title, show.Year, result.ID)
				e.updateDecisionOutcome(show.IDs.TVDB, "requested", fmt.Sprintf("Jellyseerr request ID: %d", result.ID))
				requested++

				// Log success to database
				if e.Database != nil {
					posterURL := GetShowPosterURL(e.Config, show.IDs.TMDB, show.IDs.TVDB)
					scoreInfo := scoreMap[show.IDs.TVDB]
					filterDetails := e.getFilterDetailsForShow(show.IDs.TVDB)
					err := e.Database.LogActivity(database.ActivityLog{
						Timestamp:     time.Now(),
						JobType:       jobConfig.JobName,
						MediaType:     "show",
						Title:         show.Title,
						Year:          show.Year,
						TVDBID:        show.IDs.TVDB,
						IMDBID:        show.IDs.IMDB,
						PosterURL:     posterURL,
						Score:         scoreInfo.Score,
						Rank:          scoreInfo.Rank,
						Status:        "requested",
						Message:       fmt.Sprintf("Jellyseerr request ID: %d", result.ID),
						FilterDetails: filterDetails,
					})
					if err != nil {
						slog.Error("Failed to log activity for show", "title", show.Title, "year", show.Year, "err", err)
					}
				}
			}
		}
	}

	log.Infof("%s job completed - Requested: %d, Skipped: %d, Failed: %d", jobConfig.JobName, requested, skipped, failed)
}

// evaluateShowsWithDecisions evaluates all shows through filters and scoring, tracking detailed decisions
func (e *ShowJobExecutor) evaluateShowsWithDecisions(
	shows []integrations.Show,
	jobConfig JobConfig,
) ([]integrations.Show, map[int]ScoreInfo, []ContentDecision) {
	decisions := make([]ContentDecision, 0, len(shows))
	passedShows := make([]integrations.Show, 0)

	// Evaluate each show through filters
	for _, show := range shows {
		decision := ContentDecision{
			Title:       show.Title,
			Year:        show.Year,
			MediaType:   "show",
			TVDBID:      show.IDs.TVDB,
			TMDBID:      show.IDs.TMDB,
			IMDBID:      show.IDs.IMDB,
			Rating:      show.Rating,
			Votes:       show.Votes,
			EvaluatedAt: time.Now(),
		}

		// Get poster URL (tries TVDB first, then TMDB)
		decision.PosterURL = GetShowPosterURL(e.Config, show.IDs.TMDB, show.IDs.TVDB)

		// Run through filters and get detailed results
		filterResult := filters.ShowPassesFiltersDetailed(show, e.Config.Filters.Shows)
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
			decision.ActionReason = "Passed all filter checks"

			log.Debugf("'%s (%d)' - PASS all filters (rating: %.1f)", show.Title, show.Year, show.Rating)
		} else {
			decision.Action = "rejected"
			decision.ActionReason = filterResult.Reason

			log.Debugf("'%s (%d)' - FAIL: %s", show.Title, show.Year, filterResult.Reason)
		}

		decisions = append(decisions, decision)
	}

	// Calculate scores and ranks for shows that passed filters
	scoreMap := ScoreAndRankShows(passedShows, e.Config)

	// Update decisions with score and rank information
	for i := range decisions {
		if decisions[i].PassedFilters {
			if scoreInfo, ok := scoreMap[decisions[i].TVDBID]; ok {
				decisions[i].Score = scoreInfo.Score
				decisions[i].Rank = scoreInfo.Rank
			}
		}
	}

	// Log rejected items to database so they appear in activity log
	if e.Database != nil {
		for _, decision := range decisions {
			if !decision.PassedFilters {
				posterURL := GetShowPosterURL(e.Config, decision.TMDBID, decision.TVDBID)
				err := e.Database.LogActivity(database.ActivityLog{
					Timestamp:     time.Now(),
					JobType:       jobConfig.JobName,
					MediaType:     "show",
					Title:         decision.Title,
					Year:          decision.Year,
					TMDBID:        decision.TMDBID,
					TVDBID:        decision.TVDBID,
					IMDBID:        decision.IMDBID,
					PosterURL:     posterURL,
					Score:         decision.Score,
					Rank:          0, // Not ranked since it didn't pass filters
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

	log.Infof("Filter evaluation: %d found, %d passed filters, %d rejected",
		len(shows), len(passedShows), len(shows)-len(passedShows))

	return passedShows, scoreMap, decisions
}

// updateDecisionOutcome updates a decision with the final action taken
func (e *ShowJobExecutor) updateDecisionOutcome(tvdbID int, action, reason string) {
	if e.lastDecisions == nil {
		return
	}

	for i := range e.lastDecisions.Decisions {
		if e.lastDecisions.Decisions[i].TVDBID == tvdbID {
			e.lastDecisions.Decisions[i].Action = action
			e.lastDecisions.Decisions[i].ActionReason = reason

			// Update summary stats
			switch action {
			case "added":
				e.lastDecisions.Added++
			case "requested":
				e.lastDecisions.Requested++
			case "skipped":
				e.lastDecisions.Skipped++
			case "failed":
				e.lastDecisions.Failed++
			}
			break
		}
	}
}

// getFilterDetailsForShow retrieves the filter checks for a show from decisions
func (e *ShowJobExecutor) getFilterDetailsForShow(tvdbID int) string {
	if e.lastDecisions == nil {
		return ""
	}

	for _, decision := range e.lastDecisions.Decisions {
		if decision.TVDBID == tvdbID {
			return FilterChecksToJSON(decision.FilterChecks)
		}
	}

	return ""
}
