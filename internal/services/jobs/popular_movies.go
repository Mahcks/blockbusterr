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

// RunPopularMovies fetches popular movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunPopularMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dryRun {
		log.Info("Starting popular movies job (DRY RUN)")
	} else {
		log.Info("Starting popular movies job")
	}

	// Determine mode
	mode := cfg.Jobs.Mode
	if mode == "" {
		mode = "direct" // Default to direct if not specified
	}

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch popular movies from Trakt
	popularMovies, err := traktClient.GetPopularMovies(ctx, cfg.Jobs.PopularMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch popular movies from Trakt: %v", err)
		return
	}

	log.Infof("Found %d popular movies from Trakt", len(popularMovies))

	// Route to appropriate handler based on mode
	if mode == "jellyseerr" {
		runPopularMoviesJellyseerr(ctx, cfg, db, popularMovies, dryRun)
	} else {
		runPopularMoviesDirect(ctx, cfg, db, popularMovies, dryRun)
	}
}

// runPopularMoviesDirect adds movies directly to Radarr
func runPopularMoviesDirect(ctx context.Context, cfg *config.Config, db *database.Database, popularMovies []integrations.Movie, dryRun bool) {
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

	for _, popular := range popularMovies {
		// Skip if movie already exists in Radarr
		if existingTMDBIDs[popular.IDs.TMDB] {
			log.Debugf("Skipping '%s (%d)' - already in Radarr", popular.Title, popular.Year)
			skipped++
			continue
		}

		// Create movie object for Radarr
		movie := integrations.RadarrMovie{
			Title:               popular.Title,
			Year:                popular.Year,
			TmdbID:              popular.IDs.TMDB,
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
			log.Infof("[DRY RUN] Would add popular movie '%s (%d)' to Radarr", popular.Title, popular.Year)
			added++
		} else {
			addedMovie, err := radarrClient.AddMovie(ctx, movie)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Movie '%s (%d)' already exists in Radarr", popular.Title, popular.Year)
					skipped++
				} else {
					log.Errorf("Failed to add movie '%s (%d)' to Radarr: %v", popular.Title, popular.Year, err)
					failed++
					// Log failed activity
					if db != nil {
						db.LogActivity(database.ActivityLog{
							Timestamp: time.Now(),
							JobType:   "popular_movies",
							MediaType: "movie",
							Title:     popular.Title,
							Year:      popular.Year,
							TMDBID:    popular.IDs.TMDB,
							IMDBID:    popular.IDs.IMDB,
							Status:    "failed",
							Message:   err.Error(),
						})
					}
				}
				continue
			}

			log.Infof("Added popular movie '%s (%d)' to Radarr (ID: %d)", addedMovie.Title, addedMovie.Year, addedMovie.ID)
			added++

			// Log successful activity
			if db != nil {
				posterURL := ""
				if addedMovie.TmdbID > 0 {
					posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", addedMovie.TmdbID)
				}
				db.LogActivity(database.ActivityLog{
					Timestamp: time.Now(),
					JobType:   "popular_movies",
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
		existingTMDBIDs[popular.IDs.TMDB] = true
	}

	log.Infof("Popular movies job completed - Added: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}

// runPopularMoviesJellyseerr requests movies via Jellyseerr
func runPopularMoviesJellyseerr(ctx context.Context, cfg *config.Config, db *database.Database, popularMovies []integrations.Movie, dryRun bool) {
	jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
		URL:    cfg.Jellyseerr.URL,
		APIKey: cfg.Jellyseerr.APIKey,
		UserID: cfg.Jellyseerr.UserID,
	})

	// Request movies via Jellyseerr
	added := 0
	skipped := 0
	failed := 0

	for _, popular := range popularMovies {
		if popular.IDs.TMDB == 0 {
			log.Warnf("Skipping '%s (%d)' - missing TMDB ID", popular.Title, popular.Year)
			skipped++
			continue
		}

		// Request movie via Jellyseerr (or simulate in dry-run mode)
		if dryRun {
			log.Infof("[DRY RUN] Would request popular movie '%s (%d)' via Jellyseerr", popular.Title, popular.Year)
			added++
		} else {
			result, err := jellyseerrClient.RequestMovie(popular.IDs.TMDB)
			if err != nil {
				log.Errorf("Failed to request movie '%s (%d)' via Jellyseerr: %v", popular.Title, popular.Year, err)
				failed++
				// Log failure to database
				if db != nil {
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "popular_movies",
						MediaType: "movie",
						Title:     popular.Title,
						Year:      popular.Year,
						TMDBID:    popular.IDs.TMDB,
						IMDBID:    popular.IDs.IMDB,
						Status:    "failed",
						Message:   err.Error(),
					})
				}
				continue
			}

			if result.IsAlreadyRequested() {
				log.Debugf("Movie '%s (%d)' already requested in Jellyseerr", popular.Title, popular.Year)
				skipped++
			} else {
				log.Infof("Requested popular movie '%s (%d)' via Jellyseerr (Request ID: %d)", popular.Title, popular.Year, result.ID)
				added++

				// Log success to database
				if db != nil {
					posterURL := ""
					if popular.IDs.TMDB > 0 {
						posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", popular.IDs.TMDB)
					}
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "popular_movies",
						MediaType: "movie",
						Title:     popular.Title,
						Year:      popular.Year,
						TMDBID:    popular.IDs.TMDB,
						IMDBID:    popular.IDs.IMDB,
						PosterURL: posterURL,
						Status:    "requested",
						Message:   fmt.Sprintf("Jellyseerr request ID: %d", result.ID),
					})
				}
			}
		}
	}

	log.Infof("Popular movies job completed (Jellyseerr mode) - Requested: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}
