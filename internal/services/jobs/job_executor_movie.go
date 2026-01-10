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

// MovieJobExecutor handles execution of movie jobs with unified logic
type MovieJobExecutor struct {
	Config   *config.Config
	Database *database.Database
	DryRun   bool
}

// Execute runs a movie job with the given configuration and fetcher function
func (e *MovieJobExecutor) Execute(
	ctx context.Context,
	jobConfig JobConfig,
	fetcher MovieFetcher,
) {
	if e.DryRun {
		log.Infof("Starting %s job (DRY RUN)", jobConfig.JobName)
	} else {
		log.Infof("Starting %s job", jobConfig.JobName)
	}

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     e.Config.Trakt.ClientID,
		ClientSecret: e.Config.Trakt.ClientSecret,
	})

	// Fetch movies using the provided fetcher
	movies, err := fetcher(ctx, traktClient, jobConfig.Limit, jobConfig.Period)
	if err != nil {
		log.Errorf("Failed to fetch %s from Trakt: %v", jobConfig.JobName, err)
		return
	}

	log.Infof("Found %d movies from Trakt for %s", len(movies), jobConfig.JobName)

	// Apply filters
	filteredMovies := filters.FilterMovies(movies, e.Config.Filters.Movies)
	if len(filteredMovies) < len(movies) {
		log.Infof("Filtered out %d movies, %d remaining", len(movies)-len(filteredMovies), len(filteredMovies))
	}

	// Calculate scores and ranks if scoring is enabled
	scoreMap := ScoreAndRankMovies(filteredMovies, e.Config)

	// Route to appropriate handler based on mode
	if jobConfig.Mode == "jellyseerr" {
		e.executeMoviesJellyseerr(ctx, jobConfig, filteredMovies, scoreMap)
	} else {
		e.executeMoviesDirect(ctx, jobConfig, filteredMovies, scoreMap)
	}
}

// executeMoviesDirect adds movies directly to Radarr
func (e *MovieJobExecutor) executeMoviesDirect(
	ctx context.Context,
	jobConfig JobConfig,
	movies []integrations.Movie,
	scoreMap map[int]ScoreInfo,
) {
	radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
		BaseURL: e.Config.Radarr.URL,
		APIKey:  e.Config.Radarr.APIKey,
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

	for _, movie := range movies {
		// Skip if movie already exists in Radarr
		if existingTMDBIDs[movie.IDs.TMDB] {
			log.Debugf("Skipping '%s (%d)' - already in Radarr", movie.Title, movie.Year)
			skipped++
			continue
		}

		// Create movie object for Radarr
		radarrMovie := integrations.RadarrMovie{
			Title:               movie.Title,
			Year:                movie.Year,
			TmdbID:              movie.IDs.TMDB,
			QualityProfileID:    e.Config.Radarr.QualityProfile,
			Monitored:           true,
			MinimumAvailability: "announced",
			RootFolderPath:      e.Config.Radarr.RootFolder,
			AddOptions: &integrations.RadarrAddOptions{
				SearchForMovie: true,
			},
		}

		// Add movie to Radarr (or simulate in dry-run mode)
		if e.DryRun {
			log.Infof("[DRY RUN] Would add %s movie '%s (%d)' to Radarr", jobConfig.JobName, movie.Title, movie.Year)
			added++
		} else {
			addedMovie, err := radarrClient.AddMovie(ctx, radarrMovie)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") {
					log.Debugf("Movie '%s (%d)' already exists in Radarr", movie.Title, movie.Year)
					skipped++
				} else {
					log.Errorf("Failed to add movie '%s (%d)' to Radarr: %v", movie.Title, movie.Year, err)
					failed++
					// Log failure to database
					if e.Database != nil {
						scoreInfo := scoreMap[movie.IDs.TMDB]
						e.Database.LogActivity(database.ActivityLog{
							Timestamp: time.Now(),
							JobType:   jobConfig.JobName,
							MediaType: "movie",
							Title:     movie.Title,
							Year:      movie.Year,
							TMDBID:    movie.IDs.TMDB,
							IMDBID:    movie.IDs.IMDB,
							Score:     scoreInfo.Score,
							Rank:      scoreInfo.Rank,
							Status:    "failed",
							Message:   err.Error(),
						})
					}
				}
				continue
			}

			log.Infof("Added %s movie '%s (%d)' to Radarr (ID: %d)", jobConfig.JobName, addedMovie.Title, addedMovie.Year, addedMovie.ID)
			added++

			// Log success to database
			if e.Database != nil {
				posterURL := ""
				if addedMovie.TmdbID > 0 {
					posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", addedMovie.TmdbID)
				}
				scoreInfo := scoreMap[addedMovie.TmdbID]
				e.Database.LogActivity(database.ActivityLog{
					Timestamp: time.Now(),
					JobType:   jobConfig.JobName,
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
		existingTMDBIDs[movie.IDs.TMDB] = true
	}

	log.Infof("%s job completed - Added: %d, Skipped: %d, Failed: %d", jobConfig.JobName, added, skipped, failed)
}

// executeMoviesJellyseerr requests movies via Jellyseerr
func (e *MovieJobExecutor) executeMoviesJellyseerr(
	ctx context.Context,
	jobConfig JobConfig,
	movies []integrations.Movie,
	scoreMap map[int]ScoreInfo,
) {
	jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
		URL:             e.Config.Jellyseerr.URL,
		APIKey:          e.Config.Jellyseerr.APIKey,
		UserID:          e.Config.Jellyseerr.UserID,
		RequestEmail:    e.Config.Jellyseerr.RequestCredentials.Email,
		RequestPassword: e.Config.Jellyseerr.RequestCredentials.Password,
	})

	// Request new movies via Jellyseerr
	requested := 0
	skipped := 0
	failed := 0

	for _, movie := range movies {
		// Check if movie already exists in Jellyseerr
		mediaInfo, err := jellyseerrClient.GetMovieInfo(movie.IDs.TMDB)
		if err != nil {
			log.Debugf("Failed to check Jellyseerr status for '%s (%d)': %v", movie.Title, movie.Year, err)
		} else if mediaInfo.HasMediaInfo() {
			log.Debugf("Skipping '%s (%d)' - already requested/available in Jellyseerr", movie.Title, movie.Year)
			skipped++
			continue
		}

		// Request movie via Jellyseerr (or simulate in dry-run mode)
		if e.DryRun {
			log.Infof("[DRY RUN] Would request %s movie '%s (%d)' via Jellyseerr", jobConfig.JobName, movie.Title, movie.Year)
			requested++
		} else {
			result, err := jellyseerrClient.RequestMovie(movie.IDs.TMDB)
			if err != nil {
				// Check if it's a duplicate error
				if strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "exists") || strings.Contains(err.Error(), "requested") {
					log.Debugf("Movie '%s (%d)' already requested in Jellyseerr", movie.Title, movie.Year)
					skipped++
				} else {
					log.Errorf("Failed to request movie '%s (%d)' via Jellyseerr: %v", movie.Title, movie.Year, err)
					failed++
					// Log failure to database
					if e.Database != nil {
						scoreInfo := scoreMap[movie.IDs.TMDB]
						e.Database.LogActivity(database.ActivityLog{
							Timestamp: time.Now(),
							JobType:   jobConfig.JobName,
							MediaType: "movie",
							Title:     movie.Title,
							Year:      movie.Year,
							TMDBID:    movie.IDs.TMDB,
							IMDBID:    movie.IDs.IMDB,
							Score:     scoreInfo.Score,
							Rank:      scoreInfo.Rank,
							Status:    "failed",
							Message:   err.Error(),
						})
					}
				}
				continue
			}

			if result.IsAlreadyRequested() {
				log.Debugf("Movie '%s (%d)' already requested in Jellyseerr", movie.Title, movie.Year)
				skipped++
			} else {
				log.Infof("Requested %s movie '%s (%d)' via Jellyseerr (Request ID: %d)", jobConfig.JobName, movie.Title, movie.Year, result.ID)
				requested++

				// Log success to database
				if e.Database != nil {
					posterURL := ""
					if movie.IDs.TMDB > 0 {
						posterURL = fmt.Sprintf("https://www.themoviedb.org/movie/%d", movie.IDs.TMDB)
					}
					scoreInfo := scoreMap[movie.IDs.TMDB]
					e.Database.LogActivity(database.ActivityLog{
						Timestamp: time.Now(),
						JobType:   jobConfig.JobName,
						MediaType: "movie",
						Title:     movie.Title,
						Year:      movie.Year,
						TMDBID:    movie.IDs.TMDB,
						IMDBID:    movie.IDs.IMDB,
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

	log.Infof("%s job completed - Requested: %d, Skipped: %d, Failed: %d", jobConfig.JobName, requested, skipped, failed)
}
