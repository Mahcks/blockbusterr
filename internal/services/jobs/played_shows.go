package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/filters"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// RunPlayedShows fetches played TV shows from Trakt and adds them to Sonarr or requests via Jellyseerr
func RunPlayedShows(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dryRun {
		log.Info("Starting played shows job (DRY RUN)")
	} else {
		log.Info("Starting played shows job")
	}

	// Determine mode - check job-specific mode first, then fall back to global
	mode := cfg.Jobs.PlayedShows.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct" // Default to direct if not specified
	}

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch played shows from Trakt
	playedShows, err := traktClient.GetPlayedShows(ctx, cfg.Jobs.PlayedShows.Period, cfg.Jobs.PlayedShows.Limit)
	if err != nil {
		log.Errorf("Failed to fetch played shows from Trakt: %v", err)
		return
	}

	log.Infof("Found %d played shows from Trakt", len(playedShows))

	// Apply filters - extract Show slice, filter, and convert back
	shows := make([]integrations.Show, len(playedShows))
	for i, ps := range playedShows {
		shows[i] = ps.Show
	}
	filteredShows := filters.FilterShows(shows, cfg.Filters.Shows)
	if len(filteredShows) < len(shows) {
		log.Infof("Filtered out %d shows, %d remaining", len(shows)-len(filteredShows), len(filteredShows))
	}
	// Convert back to PlayedShow slice
	filteredPlayed := make([]integrations.PlayedShow, 0, len(filteredShows))
	for _, show := range filteredShows {
		for _, ps := range playedShows {
			if ps.Show.IDs.TVDB == show.IDs.TVDB {
				filteredPlayed = append(filteredPlayed, ps)
				break
			}
		}
	}
	playedShows = filteredPlayed

	// Calculate scores and ranks if scoring is enabled
	scoreMap := ScoreAndRankShows(filteredShows, cfg)

	// Route to appropriate handler based on mode
	if mode == "jellyseerr" {
		runPlayedShowsJellyseerr(ctx, cfg, db, playedShows, scoreMap, dryRun)
	} else {
		runPlayedShowsDirect(ctx, cfg, db, playedShows, scoreMap, dryRun)
	}
}

// runPlayedShowsDirect adds shows directly to Sonarr
func runPlayedShowsDirect(ctx context.Context, cfg *config.Config, db *database.Database, playedShows []integrations.PlayedShow, scoreMap map[int]struct {
	Score float64
	Rank  int
}, dryRun bool,
) {
	sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
		BaseURL: cfg.Sonarr.URL,
		APIKey:  cfg.Sonarr.APIKey,
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

	for _, played := range playedShows {
		// Lookup series in Sonarr to get TVDB ID
		lookupResults, err := sonarrClient.LookupSeries(ctx, fmt.Sprintf("trakt:%d", played.Show.IDs.Trakt))
		if err != nil || len(lookupResults) == 0 {
			log.Errorf("Failed to lookup show '%s (%d)' in Sonarr: %v", played.Show.Title, played.Show.Year, err)
			failed++
			continue
		}

		series := lookupResults[0]

		// Check if lookup returned a series that's already in Sonarr (has an ID assigned)
		if series.ID > 0 {
			log.Debugf("Skipping '%s (%d)' - already in Sonarr (ID: %d)", played.Show.Title, played.Show.Year, series.ID)
			skipped++
			continue
		}

		// Double-check against cached TVDB IDs
		if series.TvdbID > 0 && existingTVDBIDs[series.TvdbID] {
			log.Debugf("Skipping '%s (%d)' - already in Sonarr (TVDB: %d)", played.Show.Title, played.Show.Year, series.TvdbID)
			skipped++
			continue
		}

		// Warn if TVDB ID is missing
		if series.TvdbID == 0 {
			log.Warnf("Show '%s (%d)' has no TVDB ID, may cause issues", played.Show.Title, played.Show.Year)
		}

		// Configure series for Sonarr
		series.QualityProfileID = cfg.Sonarr.QualityProfile
		series.Monitored = true
		series.RootFolderPath = cfg.Sonarr.RootFolder
		series.AddOptions = &integrations.SonarrAddOptions{
			SearchForMissingEpisodes: true,
		}

		// Add series to Sonarr (or simulate in dry-run mode)
		if dryRun {
			log.Infof("[DRY RUN] Would add played show '%s (%d)' to Sonarr", played.Show.Title, played.Show.Year)
			added++
		} else {
			addedSeries, err := sonarrClient.AddSeries(ctx, series)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Show '%s (%d)' already exists in Sonarr", played.Show.Title, played.Show.Year)
					skipped++
				} else {
					log.Errorf("Failed to add show '%s (%d)' to Sonarr: %v", played.Show.Title, played.Show.Year, err)
					failed++
					// Log failed activity
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "played_shows",
						MediaType: "show",
						Title:     played.Show.Title,
						Year:      played.Show.Year,
						TVDBID:    series.TvdbID,
						IMDBID:    played.Show.IDs.IMDB,
						Status:    "failed",
						Message:   err.Error(),
					})
				}
				continue
			}

			log.Infof("Added played show '%s (%d)' to Sonarr (ID: %d)", addedSeries.Title, addedSeries.Year, addedSeries.ID)
			added++
			// Log successful activity
			db.LogActivity(database.ActivityLog{
				Timestamp: time.Now(),
				JobType:   "played_shows",
				MediaType: "show",
				Title:     addedSeries.Title,
				Year:      addedSeries.Year,
				TVDBID:    addedSeries.TvdbID,
				IMDBID:    addedSeries.ImdbID,
				PosterURL: fmt.Sprintf("https://www.thetvdb.com/series/%d", addedSeries.TvdbID),
				Status:    "added",
			})
		}

		// Mark as existing to avoid duplicates within this job run
		existingTVDBIDs[series.TvdbID] = true
	}

	log.Infof("Played shows job completed - Added: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}

// runPlayedShowsJellyseerr requests shows via Jellyseerr
func runPlayedShowsJellyseerr(ctx context.Context, cfg *config.Config, db *database.Database, playedShows []integrations.PlayedShow, scoreMap map[int]struct {
	Score float64
	Rank  int
}, dryRun bool,
) {
	jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
		URL:             cfg.Jellyseerr.URL,
		APIKey:          cfg.Jellyseerr.APIKey,
		UserID:          cfg.Jellyseerr.UserID,
		RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
		RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
	})

	sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
		BaseURL: cfg.Sonarr.URL,
		APIKey:  cfg.Sonarr.APIKey,
	})

	// Request shows via Jellyseerr
	added := 0
	skipped := 0
	failed := 0

	for _, played := range playedShows {
		// Need to lookup the series to get TVDB ID
		lookupResults, err := sonarrClient.LookupSeries(ctx, fmt.Sprintf("trakt:%d", played.Show.IDs.Trakt))
		if err != nil || len(lookupResults) == 0 {
			log.Errorf("Failed to lookup show '%s (%d)' in Sonarr: %v", played.Show.Title, played.Show.Year, err)
			failed++
			continue
		}

		series := lookupResults[0]

		if played.Show.IDs.TMDB == 0 {
			log.Warnf("Skipping '%s (%d)' - missing TMDB ID", played.Show.Title, played.Show.Year)
			skipped++
			continue
		}

		// Check if show already exists in Jellyseerr
		mediaInfo, err := jellyseerrClient.GetShowInfo(played.Show.IDs.TMDB)
		if err != nil {
			log.Debugf("Failed to check Jellyseerr status for '%s (%d)': %v", played.Show.Title, played.Show.Year, err)
		} else if mediaInfo.HasMediaInfo() {
			log.Debugf("Skipping '%s (%d)' - already requested/available in Jellyseerr", played.Show.Title, played.Show.Year)
			skipped++
			continue
		}

		// Request show via Jellyseerr (or simulate in dry-run mode)
		if dryRun {
			log.Infof("[DRY RUN] Would request played show '%s (%d)' via Jellyseerr", played.Show.Title, played.Show.Year)
			added++
		} else {
			result, err := jellyseerrClient.RequestShow(played.Show.IDs.TMDB)
			if err != nil {
				log.Errorf("Failed to request show '%s (%d)' via Jellyseerr: %v", played.Show.Title, played.Show.Year, err)
				failed++
				// Log failure to database
				if db != nil {
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "played_shows",
						MediaType: "show",
						Title:     played.Show.Title,
						Year:      played.Show.Year,
						TVDBID:    series.TvdbID,
						IMDBID:    played.Show.IDs.IMDB,
						Status:    "failed",
						Message:   err.Error(),
					})
				}
				continue
			}

			if result.IsAlreadyRequested() {
				log.Debugf("Show '%s (%d)' already requested in Jellyseerr", played.Show.Title, played.Show.Year)
				skipped++
			} else {
				log.Infof("Requested played show '%s (%d)' via Jellyseerr (Request ID: %d)", played.Show.Title, played.Show.Year, result.ID)
				added++

				// Log success to database
				if db != nil {
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "played_shows",
						MediaType: "show",
						Title:     played.Show.Title,
						Year:      played.Show.Year,
						TVDBID:    series.TvdbID,
						IMDBID:    played.Show.IDs.IMDB,
						PosterURL: fmt.Sprintf("https://www.thetvdb.com/series/%d", series.TvdbID),
						Status:    "requested",
						Message:   fmt.Sprintf("Jellyseerr request ID: %d", result.ID),
					})
				}
			}
		}
	}

	log.Infof("Played shows job completed (Jellyseerr mode) - Requested: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}
