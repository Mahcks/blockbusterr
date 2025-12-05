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

// RunAnticipatedMovies fetches most anticipated movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunAnticipatedMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dryRun {
		log.Info("Starting anticipated movies job (DRY RUN)")
	} else {
		log.Info("Starting anticipated movies job")
	}

	// Determine mode - check job-specific mode first, then fall back to global
	mode := cfg.Jobs.AnticipatedMovies.Mode
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

	// Fetch anticipated movies from Trakt
	anticipatedMovies, err := traktClient.GetAnticipatedMovies(ctx, cfg.Jobs.AnticipatedMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch anticipated movies from Trakt: %v", err)
		return
	}

	log.Infof("Found %d anticipated movies from Trakt", len(anticipatedMovies))

	// Route to appropriate handler based on mode
	if mode == "jellyseerr" {
		runAnticipatedMoviesJellyseerr(ctx, cfg, db, anticipatedMovies, dryRun)
	} else {
		runAnticipatedMoviesDirect(ctx, cfg, db, anticipatedMovies, dryRun)
	}
}

// runAnticipatedMoviesDirect adds movies directly to Radarr
func runAnticipatedMoviesDirect(ctx context.Context, cfg *config.Config, db *database.Database, anticipatedMovies []integrations.AnticipatedMovie, dryRun bool) {
	radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
		BaseURL: cfg.Radarr.URL,
		APIKey:  cfg.Radarr.APIKey,
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

	for _, anticipated := range anticipatedMovies {
		// Skip if movie already exists in Radarr
		if existingTMDBIDs[anticipated.Movie.IDs.TMDB] {
			log.Debugf("Skipping '%s (%d)' - already in Radarr", anticipated.Movie.Title, anticipated.Movie.Year)
			skipped++
			continue
		}

		// Create movie object for Radarr
		movie := integrations.RadarrMovie{
			Title:               anticipated.Movie.Title,
			Year:                anticipated.Movie.Year,
			TmdbID:              anticipated.Movie.IDs.TMDB,
			QualityProfileID:    cfg.Radarr.QualityProfile,
			Monitored:           true,
			MinimumAvailability: "announced",
			RootFolderPath:      cfg.Radarr.RootFolder,
			AddOptions: &integrations.RadarrAddOptions{
				SearchForMovie: true,
			},
		}

		// Add movie to Radarr (or simulate in dry-run mode)
		if dryRun {
			log.Infof("[DRY RUN] Would add anticipated movie '%s (%d)' to Radarr", anticipated.Movie.Title, anticipated.Movie.Year)
			added++
		} else {
			addedMovie, err := radarrClient.AddMovie(ctx, movie)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Movie '%s (%d)' already exists in Radarr", anticipated.Movie.Title, anticipated.Movie.Year)
					skipped++
				} else {
					log.Errorf("Failed to add movie '%s (%d)' to Radarr: %v", anticipated.Movie.Title, anticipated.Movie.Year, err)
					failed++
					// Log failure to database
					if db != nil {
						db.LogActivity(database.ActivityLog{
							Timestamp: time.Now(),
							JobType:   "anticipated_movies",
							MediaType: "movie",
							Title:     anticipated.Movie.Title,
							Year:      anticipated.Movie.Year,
							TMDBID:    anticipated.Movie.IDs.TMDB,
							IMDBID:    anticipated.Movie.IDs.IMDB,
							Status:    "failed",
							Message:   err.Error(),
						})
					}
				}
				continue
			}

			log.Infof("Added anticipated movie '%s (%d)' to Radarr (ID: %d)", addedMovie.Title, addedMovie.Year, addedMovie.ID)
			added++

			// Log success to database
			if db != nil {
				posterURL := ""
				if addedMovie.TmdbID > 0 {
					posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", addedMovie.TmdbID)
				}
				db.LogActivity(database.ActivityLog{
					Timestamp: time.Now(),
					JobType:   "anticipated_movies",
					MediaType: "movie",
					Title:     addedMovie.Title,
					Year:      addedMovie.Year,
					TMDBID:    addedMovie.TmdbID,
					IMDBID:    addedMovie.ImdbID,
					PosterURL: posterURL,
					Status:    "added",
				})
			}
		}

		// Mark as existing to avoid duplicates within this job run
		existingTMDBIDs[anticipated.Movie.IDs.TMDB] = true
	}

	log.Infof("Anticipated movies job completed - Added: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}

// runAnticipatedMoviesJellyseerr requests movies via Jellyseerr
func runAnticipatedMoviesJellyseerr(ctx context.Context, cfg *config.Config, db *database.Database, anticipatedMovies []integrations.AnticipatedMovie, dryRun bool) {
	jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
		URL:    cfg.Jellyseerr.URL,
		APIKey: cfg.Jellyseerr.APIKey,
		UserID: cfg.Jellyseerr.UserID,
	})

	// Request movies via Jellyseerr
	added := 0
	skipped := 0
	failed := 0

	for _, anticipated := range anticipatedMovies {
		if anticipated.Movie.IDs.TMDB == 0 {
			log.Warnf("Skipping '%s (%d)' - missing TMDB ID", anticipated.Movie.Title, anticipated.Movie.Year)
			skipped++
			continue
		}

		// Request movie via Jellyseerr (or simulate in dry-run mode)
		if dryRun {
			log.Infof("[DRY RUN] Would request anticipated movie '%s (%d)' via Jellyseerr", anticipated.Movie.Title, anticipated.Movie.Year)
			added++
		} else {
			result, err := jellyseerrClient.RequestMovie(anticipated.Movie.IDs.TMDB)
			if err != nil {
				log.Errorf("Failed to request movie '%s (%d)' via Jellyseerr: %v", anticipated.Movie.Title, anticipated.Movie.Year, err)
				failed++
				// Log failure to database
				if db != nil {
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "anticipated_movies",
						MediaType: "movie",
						Title:     anticipated.Movie.Title,
						Year:      anticipated.Movie.Year,
						TMDBID:    anticipated.Movie.IDs.TMDB,
						IMDBID:    anticipated.Movie.IDs.IMDB,
						Status:    "failed",
						Message:   err.Error(),
					})
				}
				continue
			}

			if result.IsAlreadyRequested() {
				log.Debugf("Movie '%s (%d)' already requested in Jellyseerr", anticipated.Movie.Title, anticipated.Movie.Year)
				skipped++
			} else {
				log.Infof("Requested anticipated movie '%s (%d)' via Jellyseerr (Request ID: %d)", anticipated.Movie.Title, anticipated.Movie.Year, result.ID)
				added++

				// Log success to database
				if db != nil {
					posterURL := ""
					if anticipated.Movie.IDs.TMDB > 0 {
						posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", anticipated.Movie.IDs.TMDB)
					}
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "anticipated_movies",
						MediaType: "movie",
						Title:     anticipated.Movie.Title,
						Year:      anticipated.Movie.Year,
						TMDBID:    anticipated.Movie.IDs.TMDB,
						IMDBID:    anticipated.Movie.IDs.IMDB,
						PosterURL: posterURL,
						Status:    "requested",
						Message:   fmt.Sprintf("Jellyseerr request ID: %d", result.ID),
					})
				}
			}
		}
	}

	log.Infof("Anticipated movies job completed (Jellyseerr mode) - Requested: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}
