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

// RunBoxOffice fetches box office movies from Trakt and adds them to Radarr
func RunBoxOffice(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dryRun {
		log.Info("Starting box office job (DRY RUN)")
	} else {
		log.Info("Starting box office job")
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

	// Fetch box office movies from Trakt
	boxOfficeMovies, err := traktClient.GetBoxOfficeMovies(ctx, cfg.Jobs.BoxOffice.Limit)
	if err != nil {
		log.Errorf("Failed to fetch box office movies from Trakt: %v", err)
		return
	}

	log.Infof("Found %d box office movies from Trakt", len(boxOfficeMovies))

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

	for _, boxOffice := range boxOfficeMovies {
		// Skip if movie already exists in Radarr
		if existingTMDBIDs[boxOffice.Movie.IDs.TMDB] {
			log.Debugf("Skipping '%s (%d)' - already in Radarr", boxOffice.Movie.Title, boxOffice.Movie.Year)
			skipped++
			continue
		}

		// Create movie object for Radarr
		movie := integrations.RadarrMovie{
			Title:               boxOffice.Movie.Title,
			Year:                boxOffice.Movie.Year,
			TmdbID:              boxOffice.Movie.IDs.TMDB,
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
			log.Infof("[DRY RUN] Would add box office movie '%s (%d)' to Radarr (Revenue: $%d)", boxOffice.Movie.Title, boxOffice.Movie.Year, boxOffice.Revenue)
			added++
		} else {
			addedMovie, err := radarrClient.AddMovie(ctx, movie)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Movie '%s (%d)' already exists in Radarr", boxOffice.Movie.Title, boxOffice.Movie.Year)
					skipped++
				} else {
					log.Errorf("Failed to add movie '%s (%d)' to Radarr: %v", boxOffice.Movie.Title, boxOffice.Movie.Year, err)
					failed++
					// Log failed activity
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "box_office",
						MediaType: "movie",
						Title:     boxOffice.Movie.Title,
						Year:      boxOffice.Movie.Year,
						TMDBID:    boxOffice.Movie.IDs.TMDB,
						IMDBID:    boxOffice.Movie.IDs.IMDB,
						Status:    "failed",
						Message:   err.Error(),
					})
				}
				continue
			}

			log.Infof("Added box office movie '%s (%d)' to Radarr (ID: %d, Revenue: $%d)", addedMovie.Title, addedMovie.Year, addedMovie.ID, boxOffice.Revenue)
			added++
			// Log successful activity
			posterURL := ""
			if addedMovie.TmdbID > 0 {
				posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", addedMovie.TmdbID)
			}
			db.LogActivity(database.ActivityLog{
				Timestamp: time.Now(),
				JobType:   "box_office",
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
		existingTMDBIDs[boxOffice.Movie.IDs.TMDB] = true
	}

	log.Infof("Box office job completed - Added: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}
