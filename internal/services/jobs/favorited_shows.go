package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// RunFavoritedShows fetches favorited TV shows from Trakt and adds them to Sonarr or requests via Jellyseerr
func RunFavoritedShows(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dryRun {
		log.Info("Starting favorited shows job (DRY RUN)")
	} else {
		log.Info("Starting favorited shows job")
	}

	// Determine mode - check job-specific mode first, then fall back to global
	mode := cfg.Jobs.FavoritedShows.Mode
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

	// Fetch favorited shows from Trakt
	favoritedShows, err := traktClient.GetFavoritedShows(ctx, cfg.Jobs.FavoritedShows.Period, cfg.Jobs.FavoritedShows.Limit)
	if err != nil {
		log.Errorf("Failed to fetch favorited shows from Trakt: %v", err)
		return
	}

	log.Infof("Found %d favorited shows from Trakt", len(favoritedShows))

	// Route to appropriate handler based on mode
	if mode == "jellyseerr" {
		runFavoritedShowsJellyseerr(ctx, cfg, db, favoritedShows, dryRun)
	} else {
		runFavoritedShowsDirect(ctx, cfg, db, favoritedShows, dryRun)
	}
}

// runFavoritedShowsDirect adds shows directly to Sonarr
func runFavoritedShowsDirect(ctx context.Context, cfg *config.Config, db *database.Database, favoritedShows []integrations.FavoritedShow, dryRun bool) {
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

	for _, favorited := range favoritedShows {
		// Lookup series in Sonarr to get TVDB ID
		lookupResults, err := sonarrClient.LookupSeries(ctx, fmt.Sprintf("trakt:%d", favorited.Show.IDs.Trakt))
		if err != nil || len(lookupResults) == 0 {
			log.Errorf("Failed to lookup show '%s (%d)' in Sonarr: %v", favorited.Show.Title, favorited.Show.Year, err)
			failed++
			continue
		}

		series := lookupResults[0]

		// Check if lookup returned a series that's already in Sonarr (has an ID assigned)
		if series.ID > 0 {
			log.Debugf("Skipping '%s (%d)' - already in Sonarr (ID: %d)", favorited.Show.Title, favorited.Show.Year, series.ID)
			skipped++
			continue
		}

		// Double-check against cached TVDB IDs
		if series.TvdbID > 0 && existingTVDBIDs[series.TvdbID] {
			log.Debugf("Skipping '%s (%d)' - already in Sonarr (TVDB: %d)", favorited.Show.Title, favorited.Show.Year, series.TvdbID)
			skipped++
			continue
		}

		// Warn if TVDB ID is missing
		if series.TvdbID == 0 {
			log.Warnf("Show '%s (%d)' has no TVDB ID, may cause issues", favorited.Show.Title, favorited.Show.Year)
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
			log.Infof("[DRY RUN] Would add favorited show '%s (%d)' to Sonarr", favorited.Show.Title, favorited.Show.Year)
			added++
		} else {
			addedSeries, err := sonarrClient.AddSeries(ctx, series)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Show '%s (%d)' already exists in Sonarr", favorited.Show.Title, favorited.Show.Year)
					skipped++
				} else {
					log.Errorf("Failed to add show '%s (%d)' to Sonarr: %v", favorited.Show.Title, favorited.Show.Year, err)
					failed++
					// Log failed activity
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "favorited_shows",
						MediaType: "show",
						Title:     favorited.Show.Title,
						Year:      favorited.Show.Year,
						TVDBID:    series.TvdbID,
						IMDBID:    favorited.Show.IDs.IMDB,
						Status:    "failed",
						Message:   err.Error(),
					})
				}
				continue
			}

			log.Infof("Added favorited show '%s (%d)' to Sonarr (ID: %d)", addedSeries.Title, addedSeries.Year, addedSeries.ID)
			added++
			// Log successful activity
			db.LogActivity(database.ActivityLog{
				Timestamp: time.Now(),
				JobType:   "favorited_shows",
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

	log.Infof("Favorited shows job completed - Added: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}

// runFavoritedShowsJellyseerr requests shows via Jellyseerr
func runFavoritedShowsJellyseerr(ctx context.Context, cfg *config.Config, db *database.Database, favoritedShows []integrations.FavoritedShow, dryRun bool) {
	jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
		URL:    cfg.Jellyseerr.URL,
		APIKey: cfg.Jellyseerr.APIKey,
		UserID: cfg.Jellyseerr.UserID,
	})

	sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
		BaseURL: cfg.Sonarr.URL,
		APIKey:  cfg.Sonarr.APIKey,
	})

	// Request shows via Jellyseerr
	added := 0
	skipped := 0
	failed := 0

	for _, favorited := range favoritedShows {
		// Need to lookup the series to get TVDB ID
		lookupResults, err := sonarrClient.LookupSeries(ctx, fmt.Sprintf("trakt:%d", favorited.Show.IDs.Trakt))
		if err != nil || len(lookupResults) == 0 {
			log.Errorf("Failed to lookup show '%s (%d)' in Sonarr: %v", favorited.Show.Title, favorited.Show.Year, err)
			failed++
			continue
		}

		series := lookupResults[0]

		if series.TvdbID == 0 {
			log.Warnf("Skipping '%s (%d)' - missing TVDB ID", favorited.Show.Title, favorited.Show.Year)
			skipped++
			continue
		}

		// Request show via Jellyseerr (or simulate in dry-run mode)
		if dryRun {
			log.Infof("[DRY RUN] Would request favorited show '%s (%d)' via Jellyseerr", favorited.Show.Title, favorited.Show.Year)
			added++
		} else {
			result, err := jellyseerrClient.RequestShow(series.TvdbID)
			if err != nil {
				log.Errorf("Failed to request show '%s (%d)' via Jellyseerr: %v", favorited.Show.Title, favorited.Show.Year, err)
				failed++
				// Log failure to database
				if db != nil {
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "favorited_shows",
						MediaType: "show",
						Title:     favorited.Show.Title,
						Year:      favorited.Show.Year,
						TVDBID:    series.TvdbID,
						IMDBID:    favorited.Show.IDs.IMDB,
						Status:    "failed",
						Message:   err.Error(),
					})
				}
				continue
			}

			if result.IsAlreadyRequested() {
				log.Debugf("Show '%s (%d)' already requested in Jellyseerr", favorited.Show.Title, favorited.Show.Year)
				skipped++
			} else {
				log.Infof("Requested favorited show '%s (%d)' via Jellyseerr (Request ID: %d)", favorited.Show.Title, favorited.Show.Year, result.ID)
				added++

				// Log success to database
				if db != nil {
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "favorited_shows",
						MediaType: "show",
						Title:     favorited.Show.Title,
						Year:      favorited.Show.Year,
						TVDBID:    series.TvdbID,
						IMDBID:    favorited.Show.IDs.IMDB,
						PosterURL: fmt.Sprintf("https://www.thetvdb.com/series/%d", series.TvdbID),
						Status:    "requested",
						Message:   fmt.Sprintf("Jellyseerr request ID: %d", result.ID),
					})
				}
			}
		}
	}

	log.Infof("Favorited shows job completed (Jellyseerr mode) - Requested: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}
