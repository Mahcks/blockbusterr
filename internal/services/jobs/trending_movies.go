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

// RunTrendingMovies fetches trending movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunTrendingMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dryRun {
		log.Info("Starting trending movies job (DRY RUN)")
	} else {
		log.Info("Starting trending movies job")
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

	// Fetch trending movies from Trakt
	trendingMovies, err := traktClient.GetTrendingMovies(ctx, cfg.Jobs.TrendingMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch trending movies from Trakt: %v", err)
		return
	}

	log.Infof("Found %d trending movies from Trakt", len(trendingMovies))

	// Route to appropriate handler based on mode
	if mode == "jellyseerr" {
		runTrendingMoviesJellyseerr(ctx, cfg, db, trendingMovies, dryRun)
	} else {
		runTrendingMoviesDirect(ctx, cfg, db, trendingMovies, dryRun)
	}
}

// runTrendingMoviesDirect adds movies directly to Radarr
func runTrendingMoviesDirect(ctx context.Context, cfg *config.Config, db *database.Database, trendingMovies []integrations.TrendingMovie, dryRun bool) {
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

	for _, trending := range trendingMovies {
		// Skip if movie already exists in Radarr
		if existingTMDBIDs[trending.Movie.IDs.TMDB] {
			log.Debugf("Skipping '%s (%d)' - already in Radarr", trending.Movie.Title, trending.Movie.Year)
			skipped++
			continue
		}

		// Create movie object for Radarr
		movie := integrations.RadarrMovie{
			Title:               trending.Movie.Title,
			Year:                trending.Movie.Year,
			TmdbID:              trending.Movie.IDs.TMDB,
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
			log.Infof("[DRY RUN] Would add trending movie '%s (%d)' to Radarr", trending.Movie.Title, trending.Movie.Year)
			added++
		} else {
			addedMovie, err := radarrClient.AddMovie(ctx, movie)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Movie '%s (%d)' already exists in Radarr", trending.Movie.Title, trending.Movie.Year)
					skipped++
				} else {
					log.Errorf("Failed to add movie '%s (%d)' to Radarr: %v", trending.Movie.Title, trending.Movie.Year, err)
					failed++
					// Log failure to database
					if db != nil {
						db.LogActivity(database.ActivityLog{
							Timestamp: time.Now(),
							JobType:   "trending_movies",
							MediaType: "movie",
							Title:     trending.Movie.Title,
							Year:      trending.Movie.Year,
							TMDBID:    trending.Movie.IDs.TMDB,
							IMDBID:    trending.Movie.IDs.IMDB,
							Status:    "failed",
							Message:   err.Error(),
						})
					}
				}
				continue
			}

			log.Infof("Added trending movie '%s (%d)' to Radarr (ID: %d)", addedMovie.Title, addedMovie.Year, addedMovie.ID)
			added++

			// Log success to database
			if db != nil {
				posterURL := ""
				if addedMovie.TmdbID > 0 {
					posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", addedMovie.TmdbID)
				}
				db.LogActivity(database.ActivityLog{
					Timestamp: time.Now(),
					JobType:   "trending_movies",
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
		existingTMDBIDs[trending.Movie.IDs.TMDB] = true
	}

	log.Infof("Trending movies job completed - Added: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}

// runTrendingMoviesJellyseerr requests movies via Jellyseerr
func runTrendingMoviesJellyseerr(ctx context.Context, cfg *config.Config, db *database.Database, trendingMovies []integrations.TrendingMovie, dryRun bool) {
	jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
		URL:    cfg.Jellyseerr.URL,
		APIKey: cfg.Jellyseerr.APIKey,
		UserID: cfg.Jellyseerr.UserID,
	})

	// Request movies via Jellyseerr
	added := 0
	skipped := 0
	failed := 0

	for _, trending := range trendingMovies {
		if trending.Movie.IDs.TMDB == 0 {
			log.Warnf("Skipping '%s (%d)' - missing TMDB ID", trending.Movie.Title, trending.Movie.Year)
			skipped++
			continue
		}

		// Request movie via Jellyseerr (or simulate in dry-run mode)
		if dryRun {
			log.Infof("[DRY RUN] Would request trending movie '%s (%d)' via Jellyseerr", trending.Movie.Title, trending.Movie.Year)
			added++
		} else {
			result, err := jellyseerrClient.RequestMovie(trending.Movie.IDs.TMDB)
			if err != nil {
				log.Errorf("Failed to request movie '%s (%d)' via Jellyseerr: %v", trending.Movie.Title, trending.Movie.Year, err)
				failed++
				// Log failure to database
				if db != nil {
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "trending_movies",
						MediaType: "movie",
						Title:     trending.Movie.Title,
						Year:      trending.Movie.Year,
						TMDBID:    trending.Movie.IDs.TMDB,
						IMDBID:    trending.Movie.IDs.IMDB,
						Status:    "failed",
						Message:   err.Error(),
					})
				}
				continue
			}

			if result.IsAlreadyRequested() {
				log.Debugf("Movie '%s (%d)' already requested in Jellyseerr", trending.Movie.Title, trending.Movie.Year)
				skipped++
			} else {
				log.Infof("Requested trending movie '%s (%d)' via Jellyseerr (Request ID: %d)", trending.Movie.Title, trending.Movie.Year, result.ID)
				added++

				// Log success to database
				if db != nil {
					posterURL := ""
					if trending.Movie.IDs.TMDB > 0 {
						posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", trending.Movie.IDs.TMDB)
					}
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "trending_movies",
						MediaType: "movie",
						Title:     trending.Movie.Title,
						Year:      trending.Movie.Year,
						TMDBID:    trending.Movie.IDs.TMDB,
						IMDBID:    trending.Movie.IDs.IMDB,
						PosterURL: posterURL,
						Status:    "requested",
						Message:   fmt.Sprintf("Jellyseerr request ID: %d", result.ID),
					})
				}
			}
		}
	}

	log.Infof("Trending movies job completed (Jellyseerr mode) - Requested: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}
