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

// RunPopularShows fetches popular TV shows from Trakt and adds them to Sonarr or requests via Jellyseerr
func RunPopularShows(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dryRun {
		log.Info("Starting popular shows job (DRY RUN)")
	} else {
		log.Info("Starting popular shows job")
	}

	// Determine mode - check job-specific mode first, then fall back to global
	mode := cfg.Jobs.PopularShows.Mode
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

	// Fetch popular shows from Trakt
	popularShows, err := traktClient.GetPopularShows(ctx, cfg.Jobs.PopularShows.Limit)
	if err != nil {
		log.Errorf("Failed to fetch popular shows from Trakt: %v", err)
		return
	}

	log.Infof("Found %d popular shows from Trakt", len(popularShows))

	// Apply filters
	filteredShows := filters.FilterShows(popularShows, cfg.Filters.Shows)
	if len(filteredShows) < len(popularShows) {
		log.Infof("Filtered out %d shows, %d remaining", len(popularShows)-len(filteredShows), len(filteredShows))
	}
	popularShows = filteredShows

	// Route to appropriate handler based on mode
	if mode == "jellyseerr" {
		runPopularShowsJellyseerr(ctx, cfg, db, popularShows, dryRun)
	} else {
		runPopularShowsDirect(ctx, cfg, db, popularShows, dryRun)
	}
}

// runPopularShowsDirect adds shows directly to Sonarr
func runPopularShowsDirect(ctx context.Context, cfg *config.Config, db *database.Database, popularShows []integrations.Show, dryRun bool) {
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

	for _, popular := range popularShows {
		// Sonarr uses TVDB ID, but Trakt provides it in the IDs
		// We need to lookup the series first to get the TVDB ID
		lookupResults, err := sonarrClient.LookupSeries(ctx, fmt.Sprintf("trakt:%d", popular.IDs.Trakt))
		if err != nil || len(lookupResults) == 0 {
			log.Errorf("Failed to lookup show '%s (%d)' in Sonarr: %v", popular.Title, popular.Year, err)
			failed++
			continue
		}

		series := lookupResults[0]

		// Check if lookup returned a series that's already in Sonarr (has an ID assigned)
		if series.ID > 0 {
			log.Debugf("Skipping '%s (%d)' - already in Sonarr (ID: %d)", popular.Title, popular.Year, series.ID)
			skipped++
			continue
		}

		// Double-check against cached TVDB IDs
		if series.TvdbID > 0 && existingTVDBIDs[series.TvdbID] {
			log.Debugf("Skipping '%s (%d)' - already in Sonarr (TVDB: %d)", popular.Title, popular.Year, series.TvdbID)
			skipped++
			continue
		}

		// Warn if TVDB ID is missing
		if series.TvdbID == 0 {
			log.Warnf("Show '%s (%d)' has no TVDB ID, may cause issues", popular.Title, popular.Year)
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
			log.Infof("[DRY RUN] Would add popular show '%s (%d)' to Sonarr", popular.Title, popular.Year)
			added++
		} else {
			addedSeries, err := sonarrClient.AddSeries(ctx, series)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Show '%s (%d)' already exists in Sonarr", popular.Title, popular.Year)
					skipped++
				} else {
					log.Errorf("Failed to add show '%s (%d)' to Sonarr: %v", popular.Title, popular.Year, err)
					failed++
					// Log failed activity
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "popular_shows",
						MediaType: "show",
						Title:     popular.Title,
						Year:      popular.Year,
						TVDBID:    series.TvdbID,
						IMDBID:    popular.IDs.IMDB,
						Status:    "failed",
						Message:   err.Error(),
					})
				}
				continue
			}

			log.Infof("Added popular show '%s (%d)' to Sonarr (ID: %d)", addedSeries.Title, addedSeries.Year, addedSeries.ID)
			added++
			// Log successful activity
			db.LogActivity(database.ActivityLog{
				Timestamp: time.Now(),
				JobType:   "popular_shows",
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

	log.Infof("Popular shows job completed - Added: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}

// runPopularShowsJellyseerr requests shows via Jellyseerr
func runPopularShowsJellyseerr(ctx context.Context, cfg *config.Config, db *database.Database, popularShows []integrations.Show, dryRun bool) {
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

	for _, popular := range popularShows {
		// Need to lookup the series to get TVDB ID
		lookupResults, err := sonarrClient.LookupSeries(ctx, fmt.Sprintf("trakt:%d", popular.IDs.Trakt))
		if err != nil || len(lookupResults) == 0 {
			log.Errorf("Failed to lookup show '%s (%d)' in Sonarr: %v", popular.Title, popular.Year, err)
			failed++
			continue
		}

		series := lookupResults[0]

		if popular.IDs.TMDB == 0 {
			log.Warnf("Skipping '%s (%d)' - missing TMDB ID", popular.Title, popular.Year)
			skipped++
			continue
		}

		// Check if show already exists in Jellyseerr
		mediaInfo, err := jellyseerrClient.GetShowInfo(popular.IDs.TMDB)
		if err != nil {
			log.Debugf("Failed to check Jellyseerr status for '%s (%d)': %v", popular.Title, popular.Year, err)
		} else if mediaInfo.HasMediaInfo() {
			log.Debugf("Skipping '%s (%d)' - already requested/available in Jellyseerr", popular.Title, popular.Year)
			skipped++
			continue
		}

		// Request show via Jellyseerr (or simulate in dry-run mode)
		if dryRun {
			log.Infof("[DRY RUN] Would request popular show '%s (%d)' via Jellyseerr", popular.Title, popular.Year)
			added++
		} else {
			result, err := jellyseerrClient.RequestShow(popular.IDs.TMDB)
			if err != nil {
				log.Errorf("Failed to request show '%s (%d)' via Jellyseerr: %v", popular.Title, popular.Year, err)
				failed++
				// Log failure to database
				if db != nil {
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "popular_shows",
						MediaType: "show",
						Title:     popular.Title,
						Year:      popular.Year,
						TVDBID:    series.TvdbID,
						IMDBID:    popular.IDs.IMDB,
						Status:    "failed",
						Message:   err.Error(),
					})
				}
				continue
			}

			if result.IsAlreadyRequested() {
				log.Debugf("Show '%s (%d)' already requested in Jellyseerr", popular.Title, popular.Year)
				skipped++
			} else {
				log.Infof("Requested popular show '%s (%d)' via Jellyseerr (Request ID: %d)", popular.Title, popular.Year, result.ID)
				added++

				// Log success to database
				if db != nil {
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "popular_shows",
						MediaType: "show",
						Title:     popular.Title,
						Year:      popular.Year,
						TVDBID:    series.TvdbID,
						IMDBID:    popular.IDs.IMDB,
						PosterURL: fmt.Sprintf("https://www.thetvdb.com/series/%d", series.TvdbID),
						Status:    "requested",
						Message:   fmt.Sprintf("Jellyseerr request ID: %d", result.ID),
					})
				}
			}
		}
	}

	log.Infof("Popular shows job completed (Jellyseerr mode) - Requested: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}
