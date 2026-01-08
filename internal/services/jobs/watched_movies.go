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

// RunWatchedMovies fetches most watched movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunWatchedMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dryRun {
		log.Info("Starting watched movies job (DRY RUN)")
	} else {
		log.Info("Starting watched movies job")
	}

	// Determine mode - check job-specific mode first, then fall back to global
	mode := cfg.Jobs.WatchedMovies.Mode
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

	// Fetch watched movies from Trakt
	watchedMovies, err := traktClient.GetWatchedMovies(ctx, cfg.Jobs.WatchedMovies.Period, cfg.Jobs.WatchedMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch watched movies from Trakt: %v", err)
		return
	}

	log.Infof("Found %d watched movies from Trakt", len(watchedMovies))

	// Apply filters - extract Movie slice, filter, and convert back
	movies := make([]integrations.Movie, len(watchedMovies))
	for i, wm := range watchedMovies {
		movies[i] = wm.Movie
	}
	filteredMovies := filters.FilterMovies(movies, cfg.Filters.Movies)
	if len(filteredMovies) < len(movies) {
		log.Infof("Filtered out %d movies, %d remaining", len(movies)-len(filteredMovies), len(filteredMovies))
	}
	// Convert back to WatchedMovie slice
	filteredWatched := make([]integrations.WatchedMovie, 0, len(filteredMovies))
	for _, movie := range filteredMovies {
		for _, wm := range watchedMovies {
			if wm.Movie.IDs.TMDB == movie.IDs.TMDB {
				filteredWatched = append(filteredWatched, wm)
				break
			}
		}
	}
	watchedMovies = filteredWatched

	// Route to appropriate handler based on mode
	if mode == "jellyseerr" {
		runWatchedMoviesJellyseerr(ctx, cfg, db, watchedMovies, dryRun)
	} else {
		runWatchedMoviesDirect(ctx, cfg, db, watchedMovies, dryRun)
	}
}

// runWatchedMoviesDirect adds movies directly to Radarr
func runWatchedMoviesDirect(ctx context.Context, cfg *config.Config, db *database.Database, watchedMovies []integrations.WatchedMovie, dryRun bool) {
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

	for _, watched := range watchedMovies {
		// Skip if movie already exists in Radarr
		if existingTMDBIDs[watched.Movie.IDs.TMDB] {
			log.Debugf("Skipping '%s (%d)' - already in Radarr", watched.Movie.Title, watched.Movie.Year)
			skipped++
			continue
		}

		// Create movie object for Radarr
		movie := integrations.RadarrMovie{
			Title:               watched.Movie.Title,
			Year:                watched.Movie.Year,
			TmdbID:              watched.Movie.IDs.TMDB,
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
			log.Infof("[DRY RUN] Would add watched movie '%s (%d)' to Radarr", watched.Movie.Title, watched.Movie.Year)
			added++
		} else {
			addedMovie, err := radarrClient.AddMovie(ctx, movie)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Movie '%s (%d)' already exists in Radarr", watched.Movie.Title, watched.Movie.Year)
					skipped++
				} else {
					log.Errorf("Failed to add movie '%s (%d)' to Radarr: %v", watched.Movie.Title, watched.Movie.Year, err)
					failed++
					// Log failure to database
					if db != nil {
						db.LogActivity(database.ActivityLog{
							Timestamp: time.Now(),
							JobType:   "watched_movies",
							MediaType: "movie",
							Title:     watched.Movie.Title,
							Year:      watched.Movie.Year,
							TMDBID:    watched.Movie.IDs.TMDB,
							IMDBID:    watched.Movie.IDs.IMDB,
							Status:    "failed",
							Message:   err.Error(),
						})
					}
				}
				continue
			}

			log.Infof("Added watched movie '%s (%d)' to Radarr (ID: %d)", addedMovie.Title, addedMovie.Year, addedMovie.ID)
			added++

			// Log success to database
			if db != nil {
				posterURL := ""
				if addedMovie.TmdbID > 0 {
					posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", addedMovie.TmdbID)
				}
				db.LogActivity(database.ActivityLog{
					Timestamp: time.Now(),
					JobType:   "watched_movies",
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
		existingTMDBIDs[watched.Movie.IDs.TMDB] = true
	}

	log.Infof("Watched movies job completed - Added: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}

// runWatchedMoviesJellyseerr requests movies via Jellyseerr
func runWatchedMoviesJellyseerr(ctx context.Context, cfg *config.Config, db *database.Database, watchedMovies []integrations.WatchedMovie, dryRun bool) {
	jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
		URL:    cfg.Jellyseerr.URL,
		APIKey: cfg.Jellyseerr.APIKey,
		UserID: cfg.Jellyseerr.UserID,
		RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
		RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
	})

	// Request movies via Jellyseerr
	added := 0
	skipped := 0
	failed := 0

	for _, watched := range watchedMovies {
		if watched.Movie.IDs.TMDB == 0 {
			log.Warnf("Skipping '%s (%d)' - missing TMDB ID", watched.Movie.Title, watched.Movie.Year)
			skipped++
			continue
		}

		// Request movie via Jellyseerr (or simulate in dry-run mode)
		if dryRun {
			log.Infof("[DRY RUN] Would request watched movie '%s (%d)' via Jellyseerr", watched.Movie.Title, watched.Movie.Year)
			added++
		} else {
			result, err := jellyseerrClient.RequestMovie(watched.Movie.IDs.TMDB)
			if err != nil {
				log.Errorf("Failed to request movie '%s (%d)' via Jellyseerr: %v", watched.Movie.Title, watched.Movie.Year, err)
				failed++
				// Log failure to database
				if db != nil {
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "watched_movies",
						MediaType: "movie",
						Title:     watched.Movie.Title,
						Year:      watched.Movie.Year,
						TMDBID:    watched.Movie.IDs.TMDB,
						IMDBID:    watched.Movie.IDs.IMDB,
						Status:    "failed",
						Message:   err.Error(),
					})
				}
				continue
			}

			if result.IsAlreadyRequested() {
				log.Debugf("Movie '%s (%d)' already requested in Jellyseerr", watched.Movie.Title, watched.Movie.Year)
				skipped++
			} else {
				log.Infof("Requested watched movie '%s (%d)' via Jellyseerr (Request ID: %d)", watched.Movie.Title, watched.Movie.Year, result.ID)
				added++

				// Log success to database
				if db != nil {
					posterURL := ""
					if watched.Movie.IDs.TMDB > 0 {
						posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", watched.Movie.IDs.TMDB)
					}
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "watched_movies",
						MediaType: "movie",
						Title:     watched.Movie.Title,
						Year:      watched.Movie.Year,
						TMDBID:    watched.Movie.IDs.TMDB,
						IMDBID:    watched.Movie.IDs.IMDB,
						PosterURL: posterURL,
						Status:    "requested",
						Message:   fmt.Sprintf("Jellyseerr request ID: %d", result.ID),
					})
				}
			}
		}
	}

	log.Infof("Watched movies job completed (Jellyseerr mode) - Requested: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}
