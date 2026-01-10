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

// RunFavoritedMovies fetches favorited movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunFavoritedMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dryRun {
		log.Info("Starting favorited movies job (DRY RUN)")
	} else {
		log.Info("Starting favorited movies job")
	}

	// Determine mode - check job-specific mode first, then fall back to global
	mode := cfg.Jobs.FavoritedMovies.Mode
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

	// Fetch favorited movies from Trakt
	favoritedMovies, err := traktClient.GetFavoritedMovies(ctx, cfg.Jobs.FavoritedMovies.Period, cfg.Jobs.FavoritedMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch favorited movies from Trakt: %v", err)
		return
	}

	log.Infof("Found %d favorited movies from Trakt", len(favoritedMovies))

	// Apply filters - extract Movie slice, filter, and convert back
	movies := make([]integrations.Movie, len(favoritedMovies))
	for i, fm := range favoritedMovies {
		movies[i] = fm.Movie
	}
	filteredMovies := filters.FilterMovies(movies, cfg.Filters.Movies)
	if len(filteredMovies) < len(movies) {
		log.Infof("Filtered out %d movies, %d remaining", len(movies)-len(filteredMovies), len(filteredMovies))
	}
	// Convert back to FavoritedMovie slice
	filteredFavorited := make([]integrations.FavoritedMovie, 0, len(filteredMovies))
	for _, movie := range filteredMovies {
		for _, fm := range favoritedMovies {
			if fm.Movie.IDs.TMDB == movie.IDs.TMDB {
				filteredFavorited = append(filteredFavorited, fm)
				break
			}
		}
	}
	favoritedMovies = filteredFavorited

	// Calculate scores and ranks if scoring is enabled
	scoreMap := ScoreAndRankMovies(filteredMovies, cfg)

	// Route to appropriate handler based on mode
	if mode == "jellyseerr" {
		runFavoritedMoviesJellyseerr(ctx, cfg, db, favoritedMovies, scoreMap, dryRun)
	} else {
		runFavoritedMoviesDirect(ctx, cfg, db, favoritedMovies, scoreMap, dryRun)
	}
}

// runFavoritedMoviesDirect adds movies directly to Radarr
func runFavoritedMoviesDirect(ctx context.Context, cfg *config.Config, db *database.Database, favoritedMovies []integrations.FavoritedMovie, scoreMap map[int]struct {
	Score float64
	Rank  int
}, dryRun bool,
) {
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

	for _, favorited := range favoritedMovies {
		// Skip if movie already exists in Radarr
		if existingTMDBIDs[favorited.Movie.IDs.TMDB] {
			log.Debugf("Skipping '%s (%d)' - already in Radarr", favorited.Movie.Title, favorited.Movie.Year)
			skipped++
			continue
		}

		// Create movie object for Radarr
		movie := integrations.RadarrMovie{
			Title:               favorited.Movie.Title,
			Year:                favorited.Movie.Year,
			TmdbID:              favorited.Movie.IDs.TMDB,
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
			log.Infof("[DRY RUN] Would add favorited movie '%s (%d)' to Radarr", favorited.Movie.Title, favorited.Movie.Year)
			added++
		} else {
			addedMovie, err := radarrClient.AddMovie(ctx, movie)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Movie '%s (%d)' already exists in Radarr", favorited.Movie.Title, favorited.Movie.Year)
					skipped++
				} else {
					log.Errorf("Failed to add movie '%s (%d)' to Radarr: %v", favorited.Movie.Title, favorited.Movie.Year, err)
					failed++
					// Log failure to database
					if db != nil {
						scoreInfo := scoreMap[favorited.Movie.IDs.TMDB]
						db.LogActivity(database.ActivityLog{
							Timestamp: time.Now(),
							JobType:   "favorited_movies",
							MediaType: "movie",
							Title:     favorited.Movie.Title,
							Year:      favorited.Movie.Year,
							TMDBID:    favorited.Movie.IDs.TMDB,
							IMDBID:    favorited.Movie.IDs.IMDB,
							Score:     scoreInfo.Score,
							Rank:      scoreInfo.Rank,
							Status:    "failed",
							Message:   err.Error(),
						})
					}
				}
				continue
			}

			log.Infof("Added favorited movie '%s (%d)' to Radarr (ID: %d)", addedMovie.Title, addedMovie.Year, addedMovie.ID)
			added++

			// Log success to database
			if db != nil {
				posterURL := ""
				if addedMovie.TmdbID > 0 {
					posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", addedMovie.TmdbID)
				}
				scoreInfo := scoreMap[addedMovie.TmdbID]
				db.LogActivity(database.ActivityLog{
					Timestamp: time.Now(),
					JobType:   "favorited_movies",
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
			}
		}

		// Mark as existing to avoid duplicates within this job run
		existingTMDBIDs[favorited.Movie.IDs.TMDB] = true
	}

	log.Infof("Favorited movies job completed - Added: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}

// runFavoritedMoviesJellyseerr requests movies via Jellyseerr
func runFavoritedMoviesJellyseerr(ctx context.Context, cfg *config.Config, db *database.Database, favoritedMovies []integrations.FavoritedMovie, scoreMap map[int]struct {
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

	// Request movies via Jellyseerr
	added := 0
	skipped := 0
	failed := 0

	for _, favorited := range favoritedMovies {
		if favorited.Movie.IDs.TMDB == 0 {
			log.Warnf("Skipping '%s (%d)' - missing TMDB ID", favorited.Movie.Title, favorited.Movie.Year)
			skipped++
			continue
		}

		// Request movie via Jellyseerr (or simulate in dry-run mode)
		if dryRun {
			log.Infof("[DRY RUN] Would request favorited movie '%s (%d)' via Jellyseerr", favorited.Movie.Title, favorited.Movie.Year)
			added++
		} else {
			result, err := jellyseerrClient.RequestMovie(favorited.Movie.IDs.TMDB)
			if err != nil {
				log.Errorf("Failed to request movie '%s (%d)' via Jellyseerr: %v", favorited.Movie.Title, favorited.Movie.Year, err)
				failed++
				// Log failure to database
				if db != nil {
					scoreInfo := scoreMap[favorited.Movie.IDs.TMDB]
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "favorited_movies",
						MediaType: "movie",
						Title:     favorited.Movie.Title,
						Year:      favorited.Movie.Year,
						TMDBID:    favorited.Movie.IDs.TMDB,
						IMDBID:    favorited.Movie.IDs.IMDB,
						Score:     scoreInfo.Score,
						Rank:      scoreInfo.Rank,
						Status:    "failed",
						Message:   err.Error(),
					})
				}
				continue
			}

			if result.IsAlreadyRequested() {
				log.Debugf("Movie '%s (%d)' already requested in Jellyseerr", favorited.Movie.Title, favorited.Movie.Year)
				skipped++
			} else {
				log.Infof("Requested favorited movie '%s (%d)' via Jellyseerr (Request ID: %d)", favorited.Movie.Title, favorited.Movie.Year, result.ID)
				added++

				// Log success to database
				if db != nil {
					posterURL := ""
					if favorited.Movie.IDs.TMDB > 0 {
						posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", favorited.Movie.IDs.TMDB)
					}
					scoreInfo := scoreMap[favorited.Movie.IDs.TMDB]
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "favorited_movies",
						MediaType: "movie",
						Title:     favorited.Movie.Title,
						Year:      favorited.Movie.Year,
						TMDBID:    favorited.Movie.IDs.TMDB,
						IMDBID:    favorited.Movie.IDs.IMDB,
						PosterURL: posterURL,
						Score:     scoreInfo.Score,
						Rank:      scoreInfo.Rank,
						Status:    "requested",
						Message:   fmt.Sprintf("Jellyseerr request ID: %d", result.ID),
					})
				}
			}
		}
	}

	log.Infof("Favorited movies job completed (Jellyseerr mode) - Requested: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}
