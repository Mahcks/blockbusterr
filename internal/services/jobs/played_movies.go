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

// RunPlayedMovies fetches most played movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunPlayedMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dryRun {
		log.Info("Starting played movies job (DRY RUN)")
	} else {
		log.Info("Starting played movies job")
	}

	// Determine mode - check job-specific mode first, then fall back to global
	mode := cfg.Jobs.PlayedMovies.Mode
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

	// Fetch played movies from Trakt
	playedMovies, err := traktClient.GetPlayedMovies(ctx, cfg.Jobs.PlayedMovies.Period, cfg.Jobs.PlayedMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch played movies from Trakt: %v", err)
		return
	}

	log.Infof("Found %d played movies from Trakt", len(playedMovies))

	// Route to appropriate handler based on mode
	if mode == "jellyseerr" {
		runPlayedMoviesJellyseerr(ctx, cfg, db, playedMovies, dryRun)
	} else {
		runPlayedMoviesDirect(ctx, cfg, db, playedMovies, dryRun)
	}
}

// runPlayedMoviesDirect adds movies directly to Radarr
func runPlayedMoviesDirect(ctx context.Context, cfg *config.Config, db *database.Database, playedMovies []integrations.PlayedMovie, dryRun bool) {
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

	for _, played := range playedMovies {
		// Skip if movie already exists in Radarr
		if existingTMDBIDs[played.Movie.IDs.TMDB] {
			log.Debugf("Skipping '%s (%d)' - already in Radarr", played.Movie.Title, played.Movie.Year)
			skipped++
			continue
		}

		// Create movie object for Radarr
		movie := integrations.RadarrMovie{
			Title:               played.Movie.Title,
			Year:                played.Movie.Year,
			TmdbID:              played.Movie.IDs.TMDB,
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
			log.Infof("[DRY RUN] Would add played movie '%s (%d)' to Radarr", played.Movie.Title, played.Movie.Year)
			added++
		} else {
			addedMovie, err := radarrClient.AddMovie(ctx, movie)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Movie '%s (%d)' already exists in Radarr", played.Movie.Title, played.Movie.Year)
					skipped++
				} else {
					log.Errorf("Failed to add movie '%s (%d)' to Radarr: %v", played.Movie.Title, played.Movie.Year, err)
					failed++
					// Log failure to database
					if db != nil {
						db.LogActivity(database.ActivityLog{
							Timestamp: time.Now(),
							JobType:   "played_movies",
							MediaType: "movie",
							Title:     played.Movie.Title,
							Year:      played.Movie.Year,
							TMDBID:    played.Movie.IDs.TMDB,
							IMDBID:    played.Movie.IDs.IMDB,
							Status:    "failed",
							Message:   err.Error(),
						})
					}
				}
				continue
			}

			log.Infof("Added played movie '%s (%d)' to Radarr (ID: %d)", addedMovie.Title, addedMovie.Year, addedMovie.ID)
			added++

			// Log success to database
			if db != nil {
				posterURL := ""
				if addedMovie.TmdbID > 0 {
					posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", addedMovie.TmdbID)
				}
				db.LogActivity(database.ActivityLog{
					Timestamp: time.Now(),
					JobType:   "played_movies",
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
		existingTMDBIDs[played.Movie.IDs.TMDB] = true
	}

	log.Infof("Played movies job completed - Added: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}

// runPlayedMoviesJellyseerr requests movies via Jellyseerr
func runPlayedMoviesJellyseerr(ctx context.Context, cfg *config.Config, db *database.Database, playedMovies []integrations.PlayedMovie, dryRun bool) {
	jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
		URL:    cfg.Jellyseerr.URL,
		APIKey: cfg.Jellyseerr.APIKey,
		UserID: cfg.Jellyseerr.UserID,
	})

	// Request movies via Jellyseerr
	added := 0
	skipped := 0
	failed := 0

	for _, played := range playedMovies {
		if played.Movie.IDs.TMDB == 0 {
			log.Warnf("Skipping '%s (%d)' - missing TMDB ID", played.Movie.Title, played.Movie.Year)
			skipped++
			continue
		}

		// Request movie via Jellyseerr (or simulate in dry-run mode)
		if dryRun {
			log.Infof("[DRY RUN] Would request played movie '%s (%d)' via Jellyseerr", played.Movie.Title, played.Movie.Year)
			added++
		} else {
			result, err := jellyseerrClient.RequestMovie(played.Movie.IDs.TMDB)
			if err != nil {
				log.Errorf("Failed to request movie '%s (%d)' via Jellyseerr: %v", played.Movie.Title, played.Movie.Year, err)
				failed++
				// Log failure to database
				if db != nil {
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "played_movies",
						MediaType: "movie",
						Title:     played.Movie.Title,
						Year:      played.Movie.Year,
						TMDBID:    played.Movie.IDs.TMDB,
						IMDBID:    played.Movie.IDs.IMDB,
						Status:    "failed",
						Message:   err.Error(),
					})
				}
				continue
			}

			if result.IsAlreadyRequested() {
				log.Debugf("Movie '%s (%d)' already requested in Jellyseerr", played.Movie.Title, played.Movie.Year)
				skipped++
			} else {
				log.Infof("Requested played movie '%s (%d)' via Jellyseerr (Request ID: %d)", played.Movie.Title, played.Movie.Year, result.ID)
				added++

				// Log success to database
				if db != nil {
					posterURL := ""
					if played.Movie.IDs.TMDB > 0 {
						posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", played.Movie.IDs.TMDB)
					}
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "played_movies",
						MediaType: "movie",
						Title:     played.Movie.Title,
						Year:      played.Movie.Year,
						TMDBID:    played.Movie.IDs.TMDB,
						IMDBID:    played.Movie.IDs.IMDB,
						PosterURL: posterURL,
						Status:    "requested",
						Message:   fmt.Sprintf("Jellyseerr request ID: %d", result.ID),
					})
				}
			}
		}
	}

	log.Infof("Played movies job completed (Jellyseerr mode) - Requested: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}
