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

// MovieJobExecutor handles execution of movie jobs with unified logic
type MovieJobExecutor struct {
	Config        *config.Config
	Database      *database.Database
	DryRun        bool
	lastDecisions *JobRunDecisions // Store last run decisions for API access
}

// Execute runs a movie job with the given configuration and fetcher function
func (e *MovieJobExecutor) Execute(
	ctx context.Context,
	jobConfig JobConfig,
	fetcher MovieFetcher,
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

	// Fetch movies using the provided fetcher
	movies, err := fetcher(ctx, traktClient, jobConfig.Limit, jobConfig.Period)
	if err != nil {
		log.Errorf("Failed to fetch %s from Trakt: %v", jobConfig.JobName, err)
		return
	}

	log.Infof("Found %d movies from Trakt for %s", len(movies), jobConfig.JobName)

	// Initialize decision tracking for this run
	runDecisions := &JobRunDecisions{
		JobName:    jobConfig.JobName,
		RunTime:    time.Now(),
		TotalFound: len(movies),
		Decisions:  make([]ContentDecision, 0),
	}
	e.lastDecisions = runDecisions

	// Apply filters with detailed decision tracking
	filteredMovies, scoreMap, movieDecisions := e.evaluateMoviesWithDecisions(movies, jobConfig)
	runDecisions.Decisions = movieDecisions
	runDecisions.PassedFilters = len(filteredMovies)

	// Route to appropriate handler based on mode
	if jobConfig.Mode == "jellyseerr" {
		e.executeMoviesJellyseerr(ctx, jobConfig, filteredMovies, scoreMap)
	} else {
		e.executeMoviesDirect(ctx, jobConfig, filteredMovies, scoreMap)
	}

	// Mark run as completed and update final stats
	runDecisions.Completed = true
}

// executeMoviesDirect adds movies directly to Radarr
func (e *MovieJobExecutor) executeMoviesDirect(
	ctx context.Context,
	jobConfig JobConfig,
	movies []integrations.Movie,
	scoreMap map[int]ScoreInfo,
) {
	radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
		BaseURL: e.Config.Radarr.URL,
		APIKey:  e.Config.Radarr.APIKey,
	})

	// Get existing movies from Radarr for deduplication
	existingMovies, err := radarrClient.GetMovies(ctx)
	if err != nil {
		log.Errorf("Failed to fetch existing movies from Radarr: %v", err)
		return
	}

	// Create a map of existing TMDB IDs for fast lookup
	existingTMDBIDs := make(map[int]bool)
	for _, movie := range existingMovies {
		existingTMDBIDs[movie.TmdbID] = true
	}

	// Add new movies to Radarr
	added := 0
	skipped := 0
	failed := 0

	for _, movie := range movies {
		// Skip if movie already exists in Radarr
		if existingTMDBIDs[movie.IDs.TMDB] {
			log.Debugf("Skipping '%s (%d)' - already in Radarr", movie.Title, movie.Year)
			skipped++
			continue
		}

		minAvail := jobConfig.MinimumAvailability
		if minAvail == "" {
			minAvail = e.Config.Radarr.MinimumAvailability
		}
		if minAvail == "" {
			minAvail = "announced"
		}

		// Determine Radarr monitor setting - check job-specific, then global, then default to "movieOnly"
		monitorSetting := jobConfig.Monitor
		if monitorSetting == "" {
			monitorSetting = e.Config.Radarr.Monitor
		}
		if monitorSetting == "" {
			monitorSetting = "movieOnly"
		}

		// Determine if movie should be monitored based on monitor setting
		monitored := monitorSetting != "none"

		// Create movie object for Radarr
		radarrMovie := integrations.RadarrMovie{
			Title:               movie.Title,
			Year:                movie.Year,
			TmdbID:              movie.IDs.TMDB,
			QualityProfileID:    e.Config.Radarr.QualityProfile,
			Monitored:           monitored,
			MinimumAvailability: minAvail,
			RootFolderPath:      e.Config.Radarr.RootFolder,
			AddOptions: &integrations.RadarrAddOptions{
				SearchForMovie: monitored, // Only search if monitoring
				Monitor:        monitorSetting,
			},
		}

		// Add movie to Radarr (or simulate in dry-run mode)
		if e.DryRun {
			log.Infof("[DRY RUN] Would add %s movie '%s (%d)' to Radarr", jobConfig.JobName, movie.Title, movie.Year)
			added++
		} else {
			addedMovie, err := radarrClient.AddMovie(ctx, radarrMovie)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Movie '%s (%d)' already exists in Radarr", movie.Title, movie.Year)
					skipped++
				} else {
					log.Errorf("Failed to add movie '%s (%d)' to Radarr: %v", movie.Title, movie.Year, err)
					failed++
					// Log failure to database
					if e.Database != nil {
						scoreInfo := scoreMap[movie.IDs.TMDB]
						err := e.Database.LogActivity(database.ActivityLog{
							Timestamp: time.Now(),
							JobType:   jobConfig.JobName,
							MediaType: "movie",
							Title:     movie.Title,
							Year:      movie.Year,
							TMDBID:    movie.IDs.TMDB,
							IMDBID:    movie.IDs.IMDB,
							Score:     scoreInfo.Score,
							Rank:      scoreInfo.Rank,
							Status:    "failed",
							Message:   err.Error(),
						})
						if err != nil {
							slog.Error("Failed to log activity for movie", "title", movie.Title, "year", movie.Year, "err", err)
						}
					}
				}
				continue
			}

			log.Infof("Added %s movie '%s (%d)' to Radarr (ID: %d)", jobConfig.JobName, addedMovie.Title, addedMovie.Year, addedMovie.ID)
			added++

			// Log success to database
			if e.Database != nil {
				posterURL := GetTMDBPosterURL(e.Config, addedMovie.TmdbID, "movie")
				scoreInfo := scoreMap[addedMovie.TmdbID]
				err := e.Database.LogActivity(database.ActivityLog{
					Timestamp: time.Now(),
					JobType:   jobConfig.JobName,
					MediaType: "movie",
					Title:     addedMovie.Title,
					Year:      addedMovie.Year,
					TMDBID:    addedMovie.TmdbID,
					IMDBID:    addedMovie.ImdbID,
					PosterURL: posterURL,
					Score:     scoreInfo.Score,
					Rank:      scoreInfo.Rank,
					Status:    "added",
				})
				if err != nil {
					slog.Error("Failed to log activity for movie", "title", addedMovie.Title, "year", addedMovie.Year, "err", err)
				}
			}
		}

		// Mark as existing to avoid duplicates within this job run
		existingTMDBIDs[movie.IDs.TMDB] = true
	}

	log.Infof("%s job completed - Added: %d, Skipped: %d, Failed: %d", jobConfig.JobName, added, skipped, failed)
}

// executeMoviesJellyseerr requests movies via Jellyseerr
func (e *MovieJobExecutor) executeMoviesJellyseerr(
	_ context.Context,
	jobConfig JobConfig,
	movies []integrations.Movie,
	scoreMap map[int]ScoreInfo,
) {
	jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
		URL:             e.Config.Jellyseerr.URL,
		APIKey:          e.Config.Jellyseerr.APIKey,
		UserID:          e.Config.Jellyseerr.UserID,
		RequestEmail:    e.Config.Jellyseerr.RequestCredentials.Email,
		RequestPassword: e.Config.Jellyseerr.RequestCredentials.Password,
	})

	// Request new movies via Jellyseerr
	requested := 0
	skipped := 0
	failed := 0

	for _, movie := range movies {
		// Check if movie already exists in Jellyseerr
		mediaInfo, err := jellyseerrClient.GetMovieInfo(movie.IDs.TMDB)
		if err != nil {
			log.Debugf("Failed to check Jellyseerr status for '%s (%d)': %v", movie.Title, movie.Year, err)
		} else if mediaInfo.HasMediaInfo() {
			log.Debugf("Skipping '%s (%d)' - already requested/available in Jellyseerr", movie.Title, movie.Year)
			e.updateDecisionOutcome(movie.IDs.TMDB, "skipped", "Already in Jellyseerr")
			skipped++
			continue
		}

		// Request movie via Jellyseerr (or simulate in dry-run mode)
		if e.DryRun {
			log.Infof("[DRY RUN] Would request %s movie '%s (%d)' via Jellyseerr", jobConfig.JobName, movie.Title, movie.Year)
			e.updateDecisionOutcome(movie.IDs.TMDB, "requested", "[DRY RUN] Would be requested")
			requested++
		} else {
			result, err := jellyseerrClient.RequestMovie(movie.IDs.TMDB)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") || strings.Contains(err.Error(), "requested") {
					log.Debugf("Movie '%s (%d)' already requested in Jellyseerr", movie.Title, movie.Year)
					e.updateDecisionOutcome(movie.IDs.TMDB, "skipped", "Already requested")
					skipped++
				} else {
					log.Errorf("Failed to request movie '%s (%d)' via Jellyseerr: %v", movie.Title, movie.Year, err)
					e.updateDecisionOutcome(movie.IDs.TMDB, "failed", err.Error())
					failed++
					// Log failure to database
					if e.Database != nil {
						scoreInfo := scoreMap[movie.IDs.TMDB]
						filterDetails := e.getFilterDetailsForMovie(movie.IDs.TMDB)
						err := e.Database.LogActivity(database.ActivityLog{
							Timestamp:     time.Now(),
							JobType:       jobConfig.JobName,
							MediaType:     "movie",
							Title:         movie.Title,
							Year:          movie.Year,
							TMDBID:        movie.IDs.TMDB,
							IMDBID:        movie.IDs.IMDB,
							Score:         scoreInfo.Score,
							Rank:          scoreInfo.Rank,
							Status:        "failed",
							Message:       err.Error(),
							FilterDetails: filterDetails,
						})
						if err != nil {
							slog.Error("Failed to log activity for movie", "title", movie.Title, "year", movie.Year, "err", err)
						}
					}
				}
				continue
			}

			if result.IsAlreadyRequested() {
				log.Debugf("Movie '%s (%d)' already requested in Jellyseerr", movie.Title, movie.Year)
				e.updateDecisionOutcome(movie.IDs.TMDB, "skipped", "Already requested")
				skipped++
			} else {
				log.Infof("Requested %s movie '%s (%d)' via Jellyseerr (Request ID: %d)", jobConfig.JobName, movie.Title, movie.Year, result.ID)
				e.updateDecisionOutcome(movie.IDs.TMDB, "requested", fmt.Sprintf("Jellyseerr request ID: %d", result.ID))
				requested++

				// Log success to database
				if e.Database != nil {
					posterURL := GetTMDBPosterURL(e.Config, movie.IDs.TMDB, "movie")
					scoreInfo := scoreMap[movie.IDs.TMDB]
					filterDetails := e.getFilterDetailsForMovie(movie.IDs.TMDB)
					err := e.Database.LogActivity(database.ActivityLog{
						Timestamp:     time.Now(),
						JobType:       jobConfig.JobName,
						MediaType:     "movie",
						Title:         movie.Title,
						Year:          movie.Year,
						TMDBID:        movie.IDs.TMDB,
						IMDBID:        movie.IDs.IMDB,
						PosterURL:     posterURL,
						Score:         scoreInfo.Score,
						Rank:          scoreInfo.Rank,
						Status:        "requested",
						Message:       fmt.Sprintf("Jellyseerr request ID: %d", result.ID),
						FilterDetails: filterDetails,
					})
					if err != nil {
						slog.Error("Failed to log activity for movie", "title", movie.Title, "year", movie.Year, "err", err)
					}
				}
			}
		}
	}

	log.Infof("%s job completed - Requested: %d, Skipped: %d, Failed: %d", jobConfig.JobName, requested, skipped, failed)
}

// evaluateMoviesWithDecisions evaluates all movies through filters and scoring, tracking detailed decisions
func (e *MovieJobExecutor) evaluateMoviesWithDecisions(
	movies []integrations.Movie,
	jobConfig JobConfig,
) ([]integrations.Movie, map[int]ScoreInfo, []ContentDecision) {
	decisions := make([]ContentDecision, 0, len(movies))
	passedMovies := make([]integrations.Movie, 0)

	// Evaluate each movie through filters
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

		// Get poster URL if TMDB ID is available
		if movie.IDs.TMDB > 0 {
			decision.PosterURL = GetTMDBPosterURL(e.Config, movie.IDs.TMDB, "movie")
		}

		// Run through filters and get detailed results
		filterResult := filters.MoviePassesFiltersDetailed(movie, e.Config.Filters.Movies)
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
			decision.ActionReason = "Passed all filter checks"

			log.Debugf("'%s (%d)' - PASS all filters (rating: %.1f)", movie.Title, movie.Year, movie.Rating)
		} else {
			decision.Action = "rejected"
			decision.ActionReason = filterResult.Reason

			log.Debugf("'%s (%d)' - FAIL: %s", movie.Title, movie.Year, filterResult.Reason)
		}

		decisions = append(decisions, decision)
	}

	// Calculate scores and ranks for movies that passed filters
	scoreMap := ScoreAndRankMovies(passedMovies, e.Config)

	// Update decisions with score and rank information
	for i := range decisions {
		if decisions[i].PassedFilters {
			if scoreInfo, ok := scoreMap[decisions[i].TMDBID]; ok {
				decisions[i].Score = scoreInfo.Score
				decisions[i].Rank = scoreInfo.Rank
			}
		}
	}

	// Log rejected items to database so they appear in activity log
	if e.Database != nil {
		for _, decision := range decisions {
			if !decision.PassedFilters {
				posterURL := GetTMDBPosterURL(e.Config, decision.TMDBID, "movie")
				err := e.Database.LogActivity(database.ActivityLog{
					Timestamp:     time.Now(),
					JobType:       jobConfig.JobName,
					MediaType:     "movie",
					Title:         decision.Title,
					Year:          decision.Year,
					TMDBID:        decision.TMDBID,
					IMDBID:        decision.IMDBID,
					PosterURL:     posterURL,
					Score:         decision.Score,
					Rank:          0, // Not ranked since it didn't pass filters
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

	log.Infof("Filter evaluation: %d found, %d passed filters, %d rejected",
		len(movies), len(passedMovies), len(movies)-len(passedMovies))

	return passedMovies, scoreMap, decisions
}

// updateDecisionOutcome updates a decision with the final action taken
func (e *MovieJobExecutor) updateDecisionOutcome(tmdbID int, action, reason string) {
	if e.lastDecisions == nil {
		return
	}

	for i := range e.lastDecisions.Decisions {
		if e.lastDecisions.Decisions[i].TMDBID == tmdbID {
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

// getFilterDetailsForMovie retrieves the filter checks for a movie from decisions
func (e *MovieJobExecutor) getFilterDetailsForMovie(tmdbID int) string {
	if e.lastDecisions == nil {
		return ""
	}

	for _, decision := range e.lastDecisions.Decisions {
		if decision.TMDBID == tmdbID {
			return FilterChecksToJSON(decision.FilterChecks)
		}
	}

	return ""
}
