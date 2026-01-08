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

// RunBoxOffice fetches box office movies from Trakt and adds them to Radarr or requests via Jellyseerr
func RunBoxOffice(cfg *config.Config, db *database.Database, dryRun bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if dryRun {
		log.Info("Starting box office job (DRY RUN)")
	} else {
		log.Info("Starting box office job")
	}

	// Determine mode - check job-specific mode first, then fall back to global
	mode := cfg.Jobs.BoxOffice.Mode
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

	// Fetch box office movies from Trakt
	boxOfficeMovies, err := traktClient.GetBoxOfficeMovies(ctx, cfg.Jobs.BoxOffice.Limit)
	if err != nil {
		log.Errorf("Failed to fetch box office movies from Trakt: %v", err)
		return
	}

	log.Infof("Found %d box office movies from Trakt", len(boxOfficeMovies))

	// Apply filters - extract Movie slice, filter, and convert back
	movies := make([]integrations.Movie, len(boxOfficeMovies))
	for i, bom := range boxOfficeMovies {
		movies[i] = bom.Movie
	}
	filteredMovies := filters.FilterMovies(movies, cfg.Filters.Movies)
	if len(filteredMovies) < len(movies) {
		log.Infof("Filtered out %d movies, %d remaining", len(movies)-len(filteredMovies), len(filteredMovies))
	}
	// Convert back to BoxOfficeMovie slice
	filteredBoxOffice := make([]integrations.BoxOfficeMovie, 0, len(filteredMovies))
	for _, movie := range filteredMovies {
		for _, bom := range boxOfficeMovies {
			if bom.Movie.IDs.TMDB == movie.IDs.TMDB {
				filteredBoxOffice = append(filteredBoxOffice, bom)
				break
			}
		}
	}
	boxOfficeMovies = filteredBoxOffice

	// Route to appropriate handler based on mode
	if mode == "jellyseerr" {
		runBoxOfficeJellyseerr(ctx, cfg, db, boxOfficeMovies, dryRun)
	} else {
		runBoxOfficeDirect(ctx, cfg, db, boxOfficeMovies, dryRun)
	}
}

// runBoxOfficeDirect adds movies directly to Radarr
func runBoxOfficeDirect(ctx context.Context, cfg *config.Config, db *database.Database, boxOfficeMovies []integrations.BoxOfficeMovie, dryRun bool) {
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
					if db != nil {
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
				}
				continue
			}

			log.Infof("Added box office movie '%s (%d)' to Radarr (ID: %d, Revenue: $%d)", addedMovie.Title, addedMovie.Year, addedMovie.ID, boxOffice.Revenue)
			added++

			// Log successful activity
			if db != nil {
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
		}

		// Mark as existing to avoid duplicates within this job run
		existingTMDBIDs[boxOffice.Movie.IDs.TMDB] = true
	}

	log.Infof("Box office job completed - Added: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}

// runBoxOfficeJellyseerr requests movies via Jellyseerr
func runBoxOfficeJellyseerr(ctx context.Context, cfg *config.Config, db *database.Database, boxOfficeMovies []integrations.BoxOfficeMovie, dryRun bool) {
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

	for _, boxOffice := range boxOfficeMovies {
		if boxOffice.Movie.IDs.TMDB == 0 {
			log.Warnf("Skipping '%s (%d)' - missing TMDB ID", boxOffice.Movie.Title, boxOffice.Movie.Year)
			skipped++
			continue
		}

		// Check if movie already exists in Jellyseerr
		mediaInfo, err := jellyseerrClient.GetMovieInfo(boxOffice.Movie.IDs.TMDB)
		if err != nil {
			log.Debugf("Failed to check Jellyseerr status for '%s (%d)': %v", boxOffice.Movie.Title, boxOffice.Movie.Year, err)
		} else if mediaInfo.HasMediaInfo() {
			log.Debugf("Skipping '%s (%d)' - already requested/available in Jellyseerr", boxOffice.Movie.Title, boxOffice.Movie.Year)
			skipped++
			continue
		}

		// Request movie via Jellyseerr (or simulate in dry-run mode)
		if dryRun {
			log.Infof("[DRY RUN] Would request box office movie '%s (%d)' via Jellyseerr (Revenue: $%d)", boxOffice.Movie.Title, boxOffice.Movie.Year, boxOffice.Revenue)
			added++
		} else {
			result, err := jellyseerrClient.RequestMovie(boxOffice.Movie.IDs.TMDB)
			if err != nil {
				log.Errorf("Failed to request movie '%s (%d)' via Jellyseerr: %v", boxOffice.Movie.Title, boxOffice.Movie.Year, err)
				failed++
				// Log failure to database
				if db != nil {
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

			if result.IsAlreadyRequested() {
				log.Debugf("Movie '%s (%d)' already requested in Jellyseerr", boxOffice.Movie.Title, boxOffice.Movie.Year)
				skipped++
			} else {
				log.Infof("Requested box office movie '%s (%d)' via Jellyseerr (Request ID: %d, Revenue: $%d)", boxOffice.Movie.Title, boxOffice.Movie.Year, result.ID, boxOffice.Revenue)
				added++

				// Log success to database
				if db != nil {
					posterURL := ""
					if boxOffice.Movie.IDs.TMDB > 0 {
						posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", boxOffice.Movie.IDs.TMDB)
					}
					db.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   "box_office",
						MediaType: "movie",
						Title:     boxOffice.Movie.Title,
						Year:      boxOffice.Movie.Year,
						TMDBID:    boxOffice.Movie.IDs.TMDB,
						IMDBID:    boxOffice.Movie.IDs.IMDB,
						PosterURL: posterURL,
						Status:    "requested",
						Message:   fmt.Sprintf("Jellyseerr request ID: %d", result.ID),
					})
				}
			}
		}
	}

	log.Infof("Box office job completed (Jellyseerr mode) - Requested: %d, Skipped: %d, Failed: %d", added, skipped, failed)
}
