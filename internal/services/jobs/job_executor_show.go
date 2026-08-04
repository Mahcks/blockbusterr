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
	currentRunID  int64
	discovery     *DiscoveryClient
}

// Execute runs a show job with the given configuration and fetcher function
func (e *ShowJobExecutor) Execute(
	ctx context.Context,
	jobConfig JobConfig,
	fetcher ShowFetcher,
) error {
	jobLabel := FormatJobLabel(jobConfig.JobID, jobConfig.JobName)
	if e.DryRun {
		log.Infof("Starting %s job (DRY RUN)", jobLabel)
	} else {
		log.Infof("Starting %s job", jobLabel)
	}

	if e.Database != nil {
		runID, err := e.Database.StartJobRun(jobConfig.JobID, jobConfig.JobName, jobConfig.MediaType, jobConfig.Mode, time.Now())
		if err != nil {
			log.Warnf("Failed to create job run for %s: %v", jobLabel, err)
		} else {
			e.currentRunID = runID
			defer func() { e.currentRunID = 0 }()
		}
	}

	discoveryClient := e.discovery
	if discoveryClient == nil {
		var err error
		discoveryClient, err = NewDiscoveryClient(e.Config, jobConfig.Source)
		if err != nil {
			log.Errorf("Failed to configure discovery source for %s: %v", jobLabel, err)
			return err
		}
	}

	// Fetch shows using the provided fetcher
	shows, err := fetcher(ctx, discoveryClient, jobConfig.Limit, jobConfig.Period)
	if err != nil {
		log.Errorf("Failed to fetch %s from %s: %v", jobLabel, discoveryClient.Source(), err)
		if e.Database != nil {
			_ = e.Database.LogActivity(database.ActivityLog{
				Timestamp: time.Now(),
				JobID:     jobConfig.JobID,
				RunID:     e.currentRunID,
				JobType:   jobConfig.JobName,
				MediaType: "show",
				Title:     jobLabel,
				Status:    "failed",
				Message:   fmt.Sprintf("Failed to fetch from %s: %v", discoveryClient.Source(), err),
			})
		}
		if e.Database != nil && e.currentRunID > 0 {
			_ = e.Database.CompleteJobRun(e.currentRunID, time.Now(), "failed", 0, 0, 0, 0, 0, 0, 1, err.Error())
		}
		return err
	}

	log.Infof("Found %d shows from %s for %s", len(shows), discoveryClient.Source(), jobLabel)

	// Initialize decision tracking for this run
	runDecisions := &JobRunDecisions{
		JobName:    jobConfig.JobName,
		RunTime:    time.Now(),
		TotalFound: len(shows),
		Decisions:  make([]ContentDecision, 0),
	}
	e.lastDecisions = runDecisions

	// Apply filters with detailed decision tracking
	filteredShows, scoreMap, showDecisions := e.evaluateShowsWithDecisions(ctx, shows, jobConfig)
	runDecisions.Decisions = showDecisions
	runDecisions.PassedFilters = len(filteredShows)

	// Route to appropriate handler based on mode
	var executionErr error
	if ctx.Err() != nil {
		executionErr = ctx.Err()
	} else if jobConfig.Mode == "jellyseerr" {
		e.executeShowsJellyseerr(ctx, jobConfig, filteredShows, scoreMap)
		executionErr = ctx.Err()
	} else {
		executionErr = e.executeShowsDirect(ctx, jobConfig, filteredShows, scoreMap)
	}
	if executionErr != nil {
		if e.Database != nil && e.currentRunID > 0 {
			_ = e.Database.CompleteJobRun(e.currentRunID, time.Now(), "failed", runDecisions.TotalFound, runDecisions.PassedFilters, 0, 0, 0, runDecisions.Rejected, 1, executionErr.Error())
		}
		return executionErr
	}

	// Mark run as completed and update final stats
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

// executeShowsDirect adds shows directly to Sonarr
func (e *ShowJobExecutor) executeShowsDirect(
	ctx context.Context,
	jobConfig JobConfig,
	shows []integrations.Show,
	scoreMap map[int]ScoreInfo,
) error {
	jobLabel := FormatJobLabel(jobConfig.JobID, jobConfig.JobName)
	sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
		BaseURL: e.Config.Sonarr.URL,
		APIKey:  e.Config.Sonarr.APIKey,
	})

	// Get existing series from Sonarr for deduplication
	existingSeries, err := sonarrClient.GetSeries(ctx)
	if err != nil {
		log.Errorf("Failed to fetch existing series from Sonarr: %v", err)
		return fmt.Errorf("failed to fetch existing series from Sonarr: %w", err)
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
	budget := newDeliveryBudget(e.Config, e.Database, jobConfig.JobID, e.currentRunID, jobConfig.DeliveryLimit, e.DryRun)

	for _, show := range shows {
		// Prefer the provider ID, but never trust Sonarr's first fuzzy result.
		var series integrations.SonarrSeries
		var found bool
		if show.IDs.TVDB > 0 {
			lookupResults, lookupErr := sonarrClient.LookupSeries(ctx, fmt.Sprintf("tvdb:%d", show.IDs.TVDB))
			if lookupErr == nil {
				series, found = matchingSonarrSeries(show, lookupResults)
			}
		}
		if !found {
			lookupResults, lookupErr := sonarrClient.LookupSeries(ctx, show.Title)
			if lookupErr != nil {
				log.Errorf("Failed to lookup show '%s (%d)' in Sonarr: %v", show.Title, show.Year, lookupErr)
				failed++
				continue
			}
			series, found = matchingSonarrSeries(show, lookupResults)
		}
		if !found {
			log.Errorf("Sonarr lookup returned no exact match for '%s (%d)'", show.Title, show.Year)
			failed++
			continue
		}

		// Check if lookup returned a series that's already in Sonarr (has an ID assigned)
		if series.ID > 0 {
			log.Debugf("Skipping '%s (%d)' - already in Sonarr (ID: %d)", show.Title, show.Year, series.ID)
			e.updateDecisionOutcome(show.IDs.TVDB, "skipped", "Already in Sonarr")
			if e.Database != nil {
				posterURL := GetShowPosterURL(e.Config, show.IDs.TMDB, show.IDs.TVDB)
				scoreInfo := scoreMap[show.IDs.TVDB]
				filterDetails := e.getFilterDetailsForShow(show.IDs.TVDB)
				err := e.Database.LogActivity(database.ActivityLog{
					Timestamp:     time.Now(),
					JobID:         jobConfig.JobID,
					RunID:         e.currentRunID,
					JobType:       jobConfig.JobName,
					MediaType:     "show",
					Title:         show.Title,
					Year:          show.Year,
					Language:      show.Language,
					TVDBID:        show.IDs.TVDB,
					IMDBID:        show.IDs.IMDB,
					PosterURL:     posterURL,
					Score:         scoreInfo.Score,
					Rank:          scoreInfo.Rank,
					Status:        "skipped",
					Message:       "Already in Sonarr",
					FilterDetails: filterDetails,
				})
				if err != nil {
					slog.Error("Failed to log activity for show", "title", show.Title, "year", show.Year, "err", err)
				}
			}
			skipped++
			continue
		}

		// Double-check against cached TVDB IDs
		if series.TvdbID > 0 && existingTVDBIDs[series.TvdbID] {
			log.Debugf("Skipping '%s (%d)' - already in Sonarr (TVDB: %d)", show.Title, show.Year, series.TvdbID)
			e.updateDecisionOutcome(show.IDs.TVDB, "skipped", "Already in Sonarr")
			if e.Database != nil {
				posterURL := GetShowPosterURL(e.Config, show.IDs.TMDB, show.IDs.TVDB)
				scoreInfo := scoreMap[show.IDs.TVDB]
				filterDetails := e.getFilterDetailsForShow(show.IDs.TVDB)
				err := e.Database.LogActivity(database.ActivityLog{
					Timestamp:     time.Now(),
					JobID:         jobConfig.JobID,
					RunID:         e.currentRunID,
					JobType:       jobConfig.JobName,
					MediaType:     "show",
					Title:         show.Title,
					Year:          show.Year,
					Language:      show.Language,
					TVDBID:        show.IDs.TVDB,
					IMDBID:        show.IDs.IMDB,
					PosterURL:     posterURL,
					Score:         scoreInfo.Score,
					Rank:          scoreInfo.Rank,
					Status:        "skipped",
					Message:       "Already in Sonarr",
					FilterDetails: filterDetails,
				})
				if err != nil {
					slog.Error("Failed to log activity for show", "title", show.Title, "year", show.Year, "err", err)
				}
			}
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

		reservationID, allowed, reason := budget.reserve("show")
		if !allowed {
			e.skipShowForBudget(jobConfig, show, scoreMap, reason)
			skipped++
			continue
		}

		// Add series to Sonarr (or simulate in dry-run mode)
		if e.DryRun {
			log.Infof("[DRY RUN] Would add %s show '%s (%d)' to Sonarr", jobLabel, show.Title, show.Year)
			e.updateDecisionOutcome(show.IDs.TVDB, "added", "[DRY RUN] Would be added to Sonarr")
			if e.Database != nil {
				posterURL := GetShowPosterURL(e.Config, show.IDs.TMDB, show.IDs.TVDB)
				scoreInfo := scoreMap[show.IDs.TVDB]
				filterDetails := e.getFilterDetailsForShow(show.IDs.TVDB)
				err := e.Database.LogActivity(database.ActivityLog{
					Timestamp:     time.Now(),
					JobID:         jobConfig.JobID,
					RunID:         e.currentRunID,
					JobType:       jobConfig.JobName,
					MediaType:     "show",
					Title:         show.Title,
					Year:          show.Year,
					Language:      show.Language,
					TVDBID:        show.IDs.TVDB,
					IMDBID:        show.IDs.IMDB,
					PosterURL:     posterURL,
					Score:         scoreInfo.Score,
					Rank:          scoreInfo.Rank,
					Status:        "added",
					Message:       "[DRY RUN] Would be added to Sonarr",
					FilterDetails: filterDetails,
				})
				if err != nil {
					slog.Error("Failed to log activity for show", "title", show.Title, "year", show.Year, "err", err)
				}
			}
			added++
		} else {
			addedSeries, err := sonarrClient.AddSeries(ctx, series)
			if err != nil {
				budget.release(reservationID)
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Show '%s (%d)' already exists in Sonarr", show.Title, show.Year)
					e.updateDecisionOutcome(show.IDs.TVDB, "skipped", "Already in Sonarr")
					if e.Database != nil {
						posterURL := GetShowPosterURL(e.Config, show.IDs.TMDB, show.IDs.TVDB)
						scoreInfo := scoreMap[show.IDs.TVDB]
						filterDetails := e.getFilterDetailsForShow(show.IDs.TVDB)
						err := e.Database.LogActivity(database.ActivityLog{
							Timestamp:     time.Now(),
							JobID:         jobConfig.JobID,
							RunID:         e.currentRunID,
							JobType:       jobConfig.JobName,
							MediaType:     "show",
							Title:         show.Title,
							Year:          show.Year,
							Language:      show.Language,
							TVDBID:        show.IDs.TVDB,
							IMDBID:        show.IDs.IMDB,
							PosterURL:     posterURL,
							Score:         scoreInfo.Score,
							Rank:          scoreInfo.Rank,
							Status:        "skipped",
							Message:       "Already in Sonarr",
							FilterDetails: filterDetails,
						})
						if err != nil {
							slog.Error("Failed to log activity for show", "title", show.Title, "year", show.Year, "err", err)
						}
					}
					skipped++
				} else {
					log.Errorf("Failed to add show '%s (%d)' to Sonarr: %v", show.Title, show.Year, err)
					failed++
					// Log failed activity
					if e.Database != nil {
						scoreInfo := scoreMap[series.TvdbID]
						err := e.Database.LogActivity(database.ActivityLog{
							Timestamp: time.Now(),
							JobID:     jobConfig.JobID,
							RunID:     e.currentRunID,
							JobType:   jobConfig.JobName,
							MediaType: "show",
							Title:     show.Title,
							Year:      show.Year,
							Language:  show.Language,
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
					JobID:     jobConfig.JobID,
					RunID:     e.currentRunID,
					JobType:   jobConfig.JobName,
					MediaType: "show",
					Title:     addedSeries.Title,
					Year:      addedSeries.Year,
					Language:  show.Language,
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
	return nil
}

func matchingSonarrSeries(show integrations.Show, results []integrations.SonarrSeries) (integrations.SonarrSeries, bool) {
	for _, series := range results {
		if show.IDs.TVDB > 0 && series.TvdbID == show.IDs.TVDB {
			return series, true
		}
	}
	for _, series := range results {
		if show.IDs.TMDB > 0 && series.TmdbID == show.IDs.TMDB {
			return series, true
		}
	}
	for _, series := range results {
		if strings.EqualFold(strings.TrimSpace(series.Title), strings.TrimSpace(show.Title)) && (show.Year == 0 || series.Year == 0 || series.Year == show.Year) {
			return series, true
		}
	}
	return integrations.SonarrSeries{}, false
}

// executeShowsJellyseerr requests shows via Jellyseerr
func (e *ShowJobExecutor) executeShowsJellyseerr(
	ctx context.Context,
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
	budget := newDeliveryBudget(e.Config, e.Database, jobConfig.JobID, e.currentRunID, jobConfig.DeliveryLimit, e.DryRun)

	for _, show := range shows {
		if ctx.Err() != nil {
			return
		}
		// Jellyseerr uses TMDB IDs for TV shows, not TVDB
		// Check if show already exists in Jellyseerr
		mediaInfo, err := jellyseerrClient.GetShowInfoContext(ctx, show.IDs.TMDB)
		if err != nil {
			log.Debugf("Failed to check Jellyseerr status for '%s (%d)': %v", show.Title, show.Year, err)
		} else if mediaInfo.HasMediaInfo() {
			log.Debugf("Skipping '%s (%d)' - already requested/available in Jellyseerr", show.Title, show.Year)
			e.updateDecisionOutcome(show.IDs.TVDB, "skipped", "Already in Jellyseerr")
			if e.Database != nil {
				posterURL := GetShowPosterURL(e.Config, show.IDs.TMDB, show.IDs.TVDB)
				scoreInfo := scoreMap[show.IDs.TVDB]
				filterDetails := e.getFilterDetailsForShow(show.IDs.TVDB)
				err := e.Database.LogActivity(database.ActivityLog{
					Timestamp:     time.Now(),
					JobID:         jobConfig.JobID,
					RunID:         e.currentRunID,
					JobType:       jobConfig.JobName,
					MediaType:     "show",
					Title:         show.Title,
					Year:          show.Year,
					Language:      show.Language,
					TVDBID:        show.IDs.TVDB,
					IMDBID:        show.IDs.IMDB,
					PosterURL:     posterURL,
					Score:         scoreInfo.Score,
					Rank:          scoreInfo.Rank,
					Status:        "skipped",
					Message:       "Already in Jellyseerr",
					FilterDetails: filterDetails,
				})
				if err != nil {
					slog.Error("Failed to log activity for show", "title", show.Title, "year", show.Year, "err", err)
				}
			}
			skipped++
			continue
		}

		reservationID, allowed, reason := budget.reserve("show")
		if !allowed {
			e.skipShowForBudget(jobConfig, show, scoreMap, reason)
			skipped++
			continue
		}

		// Request show via Jellyseerr (or simulate in dry-run mode)
		if e.DryRun {
			log.Infof("[DRY RUN] Would request %s show '%s (%d)' via Jellyseerr", jobConfig.JobName, show.Title, show.Year)
			e.updateDecisionOutcome(show.IDs.TVDB, "requested", "[DRY RUN] Would be requested")
			if e.Database != nil {
				posterURL := GetShowPosterURL(e.Config, show.IDs.TMDB, show.IDs.TVDB)
				scoreInfo := scoreMap[show.IDs.TVDB]
				filterDetails := e.getFilterDetailsForShow(show.IDs.TVDB)
				err := e.Database.LogActivity(database.ActivityLog{
					Timestamp:     time.Now(),
					JobID:         jobConfig.JobID,
					RunID:         e.currentRunID,
					JobType:       jobConfig.JobName,
					MediaType:     "show",
					Title:         show.Title,
					Year:          show.Year,
					Language:      show.Language,
					TVDBID:        show.IDs.TVDB,
					IMDBID:        show.IDs.IMDB,
					PosterURL:     posterURL,
					Score:         scoreInfo.Score,
					Rank:          scoreInfo.Rank,
					Status:        "requested",
					Message:       "[DRY RUN] Would be requested via Jellyseerr",
					FilterDetails: filterDetails,
				})
				if err != nil {
					slog.Error("Failed to log activity for show", "title", show.Title, "year", show.Year, "err", err)
				}
			}
			requested++
		} else {
			result, err := jellyseerrClient.RequestShowContext(ctx, show.IDs.TMDB)
			if err != nil {
				budget.release(reservationID)
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") || strings.Contains(err.Error(), "requested") {
					log.Debugf("Show '%s (%d)' already requested in Jellyseerr", show.Title, show.Year)
					e.updateDecisionOutcome(show.IDs.TVDB, "skipped", "Already requested")
					if e.Database != nil {
						posterURL := GetShowPosterURL(e.Config, show.IDs.TMDB, show.IDs.TVDB)
						scoreInfo := scoreMap[show.IDs.TVDB]
						filterDetails := e.getFilterDetailsForShow(show.IDs.TVDB)
						err := e.Database.LogActivity(database.ActivityLog{
							Timestamp:     time.Now(),
							JobID:         jobConfig.JobID,
							RunID:         e.currentRunID,
							JobType:       jobConfig.JobName,
							MediaType:     "show",
							Title:         show.Title,
							Year:          show.Year,
							Language:      show.Language,
							TVDBID:        show.IDs.TVDB,
							IMDBID:        show.IDs.IMDB,
							PosterURL:     posterURL,
							Score:         scoreInfo.Score,
							Rank:          scoreInfo.Rank,
							Status:        "skipped",
							Message:       "Already requested",
							FilterDetails: filterDetails,
						})
						if err != nil {
							slog.Error("Failed to log activity for show", "title", show.Title, "year", show.Year, "err", err)
						}
					}
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
							JobID:         jobConfig.JobID,
							RunID:         e.currentRunID,
							JobType:       jobConfig.JobName,
							MediaType:     "show",
							Title:         show.Title,
							Year:          show.Year,
							Language:      show.Language,
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
				budget.release(reservationID)
				log.Debugf("Show '%s (%d)' already requested in Jellyseerr", show.Title, show.Year)
				e.updateDecisionOutcome(show.IDs.TVDB, "skipped", "Already requested")
				if e.Database != nil {
					posterURL := GetShowPosterURL(e.Config, show.IDs.TMDB, show.IDs.TVDB)
					scoreInfo := scoreMap[show.IDs.TVDB]
					filterDetails := e.getFilterDetailsForShow(show.IDs.TVDB)
					err := e.Database.LogActivity(database.ActivityLog{
						Timestamp:     time.Now(),
						JobID:         jobConfig.JobID,
						RunID:         e.currentRunID,
						JobType:       jobConfig.JobName,
						MediaType:     "show",
						Title:         show.Title,
						Year:          show.Year,
						Language:      show.Language,
						TVDBID:        show.IDs.TVDB,
						IMDBID:        show.IDs.IMDB,
						PosterURL:     posterURL,
						Score:         scoreInfo.Score,
						Rank:          scoreInfo.Rank,
						Status:        "skipped",
						Message:       "Already requested",
						FilterDetails: filterDetails,
					})
					if err != nil {
						slog.Error("Failed to log activity for show", "title", show.Title, "year", show.Year, "err", err)
					}
				}
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
						JobID:         jobConfig.JobID,
						RunID:         e.currentRunID,
						JobType:       jobConfig.JobName,
						MediaType:     "show",
						Title:         show.Title,
						Year:          show.Year,
						Language:      show.Language,
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
	ctx context.Context,
	shows []integrations.Show,
	jobConfig JobConfig,
) ([]integrations.Show, map[int]ScoreInfo, []ContentDecision) {
	enrichShowCertifications(ctx, e.Config, shows)
	decisions := make([]ContentDecision, 0, len(shows))
	passedShows := make([]integrations.Show, 0)

	// Evaluate each show through filters
	for _, show := range shows {
		if ctx.Err() != nil {
			break
		}
		decision := ContentDecision{
			Title:       show.Title,
			Year:        show.Year,
			Language:    show.Language,
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
		filterResult := filters.ShowPassesRules(show, e.Config.Filters.Shows, e.Config.TitleExceptions)
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
			decision.ActionReason = filters.Explain(filterResult)

			log.Debugf("'%s (%d)' - PASS all filters (rating: %.1f)", show.Title, show.Year, show.Rating)
		} else {
			decision.Action = "rejected"
			decision.ActionReason = filters.Explain(filterResult)

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
					JobID:         jobConfig.JobID,
					RunID:         e.currentRunID,
					JobType:       jobConfig.JobName,
					MediaType:     "show",
					Title:         decision.Title,
					Year:          decision.Year,
					Language:      decision.Language,
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
