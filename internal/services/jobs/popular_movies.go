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

// RunPopularMovies fetches popular movies from Trakt and adds them to Radarr
func RunPopularMovies(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dryRun {
		log.Info("Starting popular movies job (DRY RUN)")
	} else {
		log.Info("Starting popular movies job")
	}

	// Create clients
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
		BaseURL: cfg.Radarr.URL,
		APIKey:  cfg.Radarr.APIKey,
	})

	// Fetch popular movies from Trakt
	popularMovies, err := traktClient.GetPopularMovies(ctx, cfg.Jobs.PopularMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch popular movies from Trakt: %v", err)
		return
	}

	log.Infof("Found %d popular movies from Trakt", len(popularMovies))

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

			log.Infof("Added popular movie '%s (%d)' to Radarr (ID: %d)", addedMovie.Title, addedMovie.Year, addedMovie.ID)
			added++
			// Log successful activity
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

		// Mark as existing to avoid duplicates within this job run
		existingTMDBIDs[popular.IDs.TMDB] = true
	}

	log.Infof("Popular movies job completed - Added: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}
