package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/mahcks/blockbusterr/internal/db"
	"github.com/mahcks/blockbusterr/internal/helpers/ombi"
	"github.com/mahcks/blockbusterr/internal/helpers/radarr"
	"github.com/mahcks/blockbusterr/internal/helpers/trakt"
	"github.com/mahcks/blockbusterr/internal/notifications"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

type radarrJob struct {
	ombiSettings   db.OmbiSettings
	radarrSettings db.RadarrSettings
	movieSettings  db.MovieSettings

	anticipatedMovies []trakt.Movie
	boxOfficeMovies   []trakt.Movie
	popularMovies     []trakt.Movie
	trendingMovies    []trakt.Movie
}

// AnticipatedJobFunc fetches and processes anticipated movies
func (s Scheduler) AnticipatedJobFunc() {
	log.Info("[scheduler] Running Anticipated Movies job...")
	mj := radarrJob{}
	gctx := s.gctx
	helpers := s.helpers

	largeMovieQueryLimit := 1000

	// Get movie settings and handle errors
	if err := s.initializeMovieJob(&mj); err != nil {
		log.Error("[scheduler] Failed to initialize movie job", "error", err)
		return
	}

	// Fetch Anticipated Movies
	if mj.movieSettings.Anticipated.Valid && mj.movieSettings.Anticipated.Int32 > 0 {
		params := buildTraktParamsFromSettings(mj.movieSettings, largeMovieQueryLimit, true)
		anticipatedMovies, err := helpers.Trakt.GetAnticipatedMovies(gctx, params)
		if err != nil {
			log.Error("[movie-job] Error fetching anticipated movies from Trakt", "error", err)
		} else {
			mj.anticipatedMovies = filterAndLimitMovies(extractMoviesFromAnticipated(anticipatedMovies), mj.movieSettings, int(mj.movieSettings.Anticipated.Int32))
			s.processMovies(mj.anticipatedMovies, mj)
		}
	}
}

// BoxOfficeJobFunc fetches and processes box office movies
func (s Scheduler) BoxOfficeJobFunc() {
	log.Info("[scheduler] Running Box Office Movies job...")
	mj := radarrJob{}
	gctx := s.gctx
	helpers := s.helpers

	largeMovieQueryLimit := 1000

	// Get movie settings and handle errors
	if err := s.initializeMovieJob(&mj); err != nil {
		log.Error("[scheduler] Failed to initialize movie job", "error", err)
		return
	}

	// Fetch Box Office Movies
	if mj.movieSettings.BoxOffice.Valid && mj.movieSettings.BoxOffice.Int32 > 0 {
		params := buildTraktParamsFromSettings(mj.movieSettings, largeMovieQueryLimit, false)
		boxOfficeMovies, err := helpers.Trakt.GetBoxOfficeMovies(gctx, params)
		if err != nil {
			log.Error("[movie-job] Error fetching box office movies from Trakt", "error", err)
		} else {
			mj.boxOfficeMovies = filterAndLimitMovies(extractMoviesFromBoxOffice(boxOfficeMovies), mj.movieSettings, int(mj.movieSettings.BoxOffice.Int32))
			s.processMovies(mj.boxOfficeMovies, mj)
		}
	}
}

// PopularJobFunc fetches and processes popular movies
func (s Scheduler) PopularJobFunc() {
	log.Info("[scheduler] Running Popular Movies job...")
	mj := radarrJob{}
	gctx := s.gctx
	helpers := s.helpers

	largeMovieQueryLimit := 1000

	// Get movie settings and handle errors
	if err := s.initializeMovieJob(&mj); err != nil {
		log.Error("[scheduler] Failed to initialize movie job", "error", err)
		return
	}

	// Fetch Popular Movies
	if mj.movieSettings.Popular.Valid && mj.movieSettings.Popular.Int32 > 0 {
		params := buildTraktParamsFromSettings(mj.movieSettings, largeMovieQueryLimit, false)
		popularMovies, err := helpers.Trakt.GetPopularMovies(gctx, params)
		if err != nil {
			log.Error("[movie-job] Error fetching popular movies from Trakt", "error", err)
		} else {
			mj.popularMovies = filterAndLimitMovies(extractMoviesFromPopular(popularMovies), mj.movieSettings, int(mj.movieSettings.Popular.Int32))
			s.processMovies(mj.popularMovies, mj)
		}
	}
}

// TrendingJobFunc fetches and processes trending movies
func (s Scheduler) TrendingJobFunc() {
	log.Info("[scheduler] Running Trending Movies job...")
	mj := radarrJob{}
	gctx := s.gctx
	helpers := s.helpers

	largeMovieQueryLimit := 1000

	// Get movie settings and handle errors
	if err := s.initializeMovieJob(&mj); err != nil {
		log.Error("[scheduler] Failed to initialize movie job", "error", err)
		return
	}

	// Fetch Trending Movies
	if mj.movieSettings.Trending.Valid && mj.movieSettings.Trending.Int32 > 0 {
		params := buildTraktParamsFromSettings(mj.movieSettings, largeMovieQueryLimit, false)
		trendingMovies, err := helpers.Trakt.GetTrendingMovies(gctx, params)
		if err != nil {
			log.Error("[movie-job] Error fetching trending movies from Trakt", "error", err)
		} else {
			mj.trendingMovies = filterAndLimitMovies(extractMoviesFromTrending(trendingMovies), mj.movieSettings, int(mj.movieSettings.Trending.Int32))
			s.processMovies(mj.trendingMovies, mj)
		}
	}
}

// initializeMovieJob handles common setup logic for all movie jobs
func (s Scheduler) initializeMovieJob(mj *radarrJob) error {
	gctx := s.gctx

	// Get all settings for Radarr
	var err error
	mj.radarrSettings, err = gctx.Crate().SQL.Queries().GetRadarrSettings(gctx)
	if err != nil {
		return fmt.Errorf("error getting Radarr settings: %w", err)
	}

	// Get all settings for movies
	mj.movieSettings, err = gctx.Crate().SQL.Queries().GetMovieSettings(gctx)
	if err != nil {
		return fmt.Errorf("error getting movie settings: %w", err)
	}

	return nil
}

// processMovies handles the logic for processing and sending movie requests to Ombi or Radarr
func (s Scheduler) processMovies(movies []trakt.Movie, mj radarrJob) {
	gctx := s.gctx
	helpers := s.helpers

	// Check if Ombi is enabled
	ombiEnabled, err := gctx.Crate().SQL.Queries().GetSettingByKey(gctx, structures.SettingOmbiEnabled.String())
	if err != nil {
		log.Error("[scheduler] Error getting Ombi enabled setting", "error", err)
		return
	}

	// If Ombi is enabled, request movies via Ombi
	if ombiEnabled.Value.String == "true" {
		mj.ombiSettings, err = gctx.Crate().SQL.Queries().GetOmbiSettings(gctx)
		if err != nil {
			log.Error("[scheduler] Error getting Ombi settings", "error", err)
			return
		}

		// Request movies via Ombi
		requestMoviesToOmbi(helpers.Ombi, s.notifications, movies, mj.ombiSettings)
	} else {
		// Otherwise, use Radarr to request movies
		requestMoviesToRadarr(helpers.Radarr, s.notifications, movies, mj.radarrSettings)
	}
}

// Helper function to filter and limit movies based on settings
func filterAndLimitMovies(movies []trakt.Movie, settings db.MovieSettings, limit int) []trakt.Movie {
	filteredMovies := applyAdditionalFilters(movies, settings)
	return getTopNMovies(filteredMovies, limit)
}

// Helper function to build Trakt API request parameters from the settings
// Helper function to build Trakt API request parameters from the settings
func buildTraktParamsFromSettings(settings db.MovieSettings, limit int, isAnticipated bool) *trakt.TraktMovieParams {
	params := &trakt.TraktMovieParams{}
	params.Extended = "full"

	// Set the limit based on the settings
	params.Limit = limit

	// Build allowed countries string (comma-separated)
	allowedCountries := []string{}
	for _, country := range settings.AllowedCountries {
		allowedCountries = append(allowedCountries, country.CountryCode)
	}
	if len(allowedCountries) > 0 {
		params.Countries = strings.Join(allowedCountries, ",")
	}

	// Build allowed languages string (comma-separated)
	allowedLanguages := []string{}
	for _, language := range settings.AllowedLanguages {
		allowedLanguages = append(allowedLanguages, language.LanguageCode)
	}
	if len(allowedLanguages) > 0 {
		params.Languages = strings.Join(allowedLanguages, ",")
	}

	// Apply runtime filter
	if settings.MinRuntime.Valid || settings.MaxRuntime.Valid {
		minRuntime := 0
		maxRuntime := 0

		if settings.MinRuntime.Valid {
			minRuntime = int(settings.MinRuntime.Int32)
		}
		if settings.MaxRuntime.Valid {
			maxRuntime = int(settings.MaxRuntime.Int32)
		}

		params.Runtime = fmt.Sprintf("%d-%d", minRuntime, maxRuntime)
	}

	// Apply year filter (with custom logic for anticipated items)
	if isAnticipated {
		// For anticipated shows or movies, allow future years
		minYear := time.Now().Year()
		maxYear := minYear + 10 // Set a future year limit

		if settings.MinYear.Valid {
			minYear = int(settings.MinYear.Int32)
		}
		// Don't apply MaxYear for anticipated items
		params.Years = fmt.Sprintf("%d-%d", minYear, maxYear)
	} else {
		// Regular logic for non-anticipated shows/movies
		minYear := 0
		maxYear := 0

		if settings.MinYear.Valid {
			minYear = int(settings.MinYear.Int32)
		}
		if settings.MaxYear.Valid {
			maxYear = int(settings.MaxYear.Int32)
		} else {
			maxYear = time.Now().Year() // Default max year is the current year
		}

		params.Years = fmt.Sprintf("%d-%d", minYear, maxYear)
	}

	// Return the configured parameters
	return params
}

func applyAdditionalFilters(movies []trakt.Movie, settings db.MovieSettings) []trakt.Movie {
	filteredMovies := []trakt.Movie{}

	// Build blacklisted genres, keywords, and TMDb IDs from settings
	blacklistedGenres := map[string]bool{}
	for _, genre := range settings.BlacklistedGenres {
		blacklistedGenres[strings.ToLower(genre.Genre)] = true
	}

	blacklistedKeywords := map[string]bool{}
	for _, keyword := range settings.BlacklistedTitleKeywords {
		blacklistedKeywords[strings.ToLower(keyword.Keyword)] = true
	}

	blacklistedTMDBIDs := map[int]bool{}
	for _, tmdbID := range settings.BlacklistedTMDBIDs {
		blacklistedTMDBIDs[tmdbID.TMDBID] = true
	}

	// Loop through the movies and filter them
	for _, movie := range movies {
		// Filter by TMDb ID
		if blacklistedTMDBIDs[movie.IDs.TMDB] {
			continue
		}

		// Filter by genres
		blacklisted := false
		for _, genre := range movie.Genres {
			if blacklistedGenres[strings.ToLower(genre)] {
				blacklisted = true
				break
			}
		}
		if blacklisted {
			continue
		}

		// Filter by title keywords
		for keyword := range blacklistedKeywords {
			if strings.Contains(strings.ToLower(movie.Title), keyword) {
				blacklisted = true
				break
			}
		}
		if blacklisted {
			continue
		}

		// Add the movie to the filtered list if it passed all checks
		filteredMovies = append(filteredMovies, movie)
	}

	return filteredMovies
}

func getTopNMovies(movies []trakt.Movie, limit int) []trakt.Movie {
	if len(movies) <= limit {
		return movies
	}
	return movies[:limit]
}

// Extract Movies from TrendingMovies
func extractMoviesFromTrending(trendingMovies []trakt.TrendingMovie) []trakt.Movie {
	movies := []trakt.Movie{}
	for _, trendingMovie := range trendingMovies {
		movies = append(movies, trendingMovie.Movie)
	}
	return movies
}

// Extract Movies from PopularMovies
func extractMoviesFromPopular(popularMovies []trakt.Movie) []trakt.Movie {
	return popularMovies
}

// Extract Movies from AnticipatedMovies
func extractMoviesFromAnticipated(anticipatedMovies []trakt.TraktAnticipatedMovie) []trakt.Movie {
	movies := []trakt.Movie{}
	for _, anticipated := range anticipatedMovies {
		movies = append(movies, anticipated.Movie)
	}
	return movies
}

// Extract Movies from BoxOfficeMovies
func extractMoviesFromBoxOffice(boxOfficeMovies []trakt.TraktBoxOfficeMovie) []trakt.Movie {
	movies := []trakt.Movie{}
	for _, boxOffice := range boxOfficeMovies {
		movies = append(movies, boxOffice.Movie)
	}
	return movies
}

func fetchRadarrSettings(r radarr.Service, radarrSettings db.RadarrSettings) (int, string, error) {
	// Get quality profiles from Radarr
	qualityProfiles, err := r.GetQualityProfiles(nil, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get Radarr quality profiles: %w", err)
	}

	// Find the quality profile matching the stored quality ID
	var qualityProfileID int
	for _, profile := range qualityProfiles {
		if profile.ID == int(radarrSettings.Quality.Int32) {
			qualityProfileID = profile.ID
			break
		}
	}
	if qualityProfileID == 0 {
		return 0, "", fmt.Errorf("quality profile ID %v not found", radarrSettings.Quality.Int32)
	}

	// Get root folders from Radarr
	rootFolders, err := r.GetRootFolders(nil, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get Radarr root folders: %w", err)
	}

	// Find the root folder matching the stored root folder ID
	var rootFolderPath string
	for _, folder := range rootFolders {
		if folder.ID == int(radarrSettings.RootFolder.Int32) {
			rootFolderPath = folder.Path
			break
		}
	}
	if rootFolderPath == "" {
		return 0, "", fmt.Errorf("root folder ID %v not found", radarrSettings.RootFolder.Int32)
	}

	// Return the matched quality profile ID and root folder path
	return qualityProfileID, rootFolderPath, nil
}

// Request movies to Ombi
func requestMoviesToOmbi(o ombi.Service, notifications *notifications.NotificationManager, movies []trakt.Movie, ombiSettings db.OmbiSettings) {
	for _, movie := range movies {

		body := ombi.RequestMovieBody{
			TheMovieDBID:        movie.IDs.TMDB,
			LanguageCode:        "en",  // Adjust language code as needed
			RootFolderOverride:  nil,   // Override if needed
			QualityPathOverride: nil,   // Override if needed
			Is4KRequest:         false, // Modify if needed for 4K movies
		}

		// Set the request on behalf of a specific user if configured
		if ombiSettings.UserID.Valid || ombiSettings.UserID.String != "" {
			body.RequestOnBehalf = ombiSettings.UserID.String
		}

		// Set the quality profile override if configured
		if ombiSettings.MovieRootFolder.Valid || ombiSettings.MovieRootFolder.Int32 != 0 {
			rootFolder := fmt.Sprintf("%d", ombiSettings.MovieRootFolder.Int32)
			body.RootFolderOverride = &rootFolder
		}

		// Set the quality profile override if configured
		if ombiSettings.MovieQuality.Valid || ombiSettings.MovieQuality.Int32 != 0 {
			qualityProfile := fmt.Sprintf("%d", ombiSettings.MovieQuality.Int32)
			body.QualityPathOverride = &qualityProfile
		}

		_, err := o.RequestMovie(body)
		if err != nil {
			if errors.Is(err, ombi.ErrMovieAlreadyRequested) {
				// Log a warning if the movie already exists in Radarr
				log.Warnf(`[ombi-job] Skipping "%s" as it was already requested...`, movie.Title)
			} else {
				// Log an error for any other issues
				log.Errorf("[ombi-job] Failed to request movie %s: %v", movie.Title, err)
			}
		} else {
			// Log a success message if the movie was added successfully
			log.Infof("[ombi-job] Movie requested successfully: %s", movie.Title)
			// Send notification that movie was added
			moviePayload, err := json.Marshal(movie)
			if err != nil {
				log.Errorf("[ombi-job] Failed to marshal movie payload: %v", err)
				continue
			}

			err = notifications.SendNotification(structures.MOVIEADDEDALERT, moviePayload)
			if err != nil {
				log.Errorf("[ombi-job] Failed to send notification for movie %s: %v", movie.Title, err)
			}
		}
	}
}

func requestMoviesToRadarr(r radarr.Service, notifications *notifications.NotificationManager, movies []trakt.Movie, radarrSettings db.RadarrSettings) {
	// Fetch quality profile and root folder from Radarr
	qualityProfileID, rootFolderPath, err := fetchRadarrSettings(r, radarrSettings)
	if err != nil {
		log.Error("[radarr-job: radarr] Error fetching Radarr settings", "error", err)
		return
	}

	for _, movie := range movies {
		// Default minimum availability to "released"
		minimumAvailability := "released"
		if radarrSettings.MinimumAvailability.Valid {
			minimumAvailability = radarrSettings.MinimumAvailability.String
		}

		// Prepare the request body
		body := radarr.RequestMovieBody{
			Title:               movie.Title,
			TMDBID:              movie.IDs.TMDB,
			Monitored:           true,
			QualityProfileID:    qualityProfileID,
			RootFolderPath:      rootFolderPath,
			MinimumAvailability: minimumAvailability, // Set default or customize this
		}

		// Set additional options for Radarr
		body.AddOptions.SearchForMovie = true

		// Make the request to Radarr
		_, err := r.RequestMovie(nil, nil, body)
		if err != nil {
			if errors.Is(err, radarr.ErrMovieAlreadyExists) {
				// Log a warning if the movie already exists in Radarr
				log.Warnf(`[radarr-job] Skipping "%s" as it already exists in Radarr...`, movie.Title)
			} else {
				// Log an error for any other issues
				log.Errorf("[radarr-job] Failed to request movie %s: %v", movie.Title, err)
			}
		} else {
			// Log a success message if the movie was added successfully
			log.Infof("[radarr-job] Movie requested successfully: %s", movie.Title)

			// Send notification that movie was added
			moviePayload, err := json.Marshal(movie)
			if err != nil {
				log.Errorf("[radarr-job] Failed to marshal movie payload: %v", err)
				continue
			}

			err = notifications.SendNotification(structures.MOVIEADDEDALERT, moviePayload)
			if err != nil {
				log.Errorf("[radarr-job] Failed to send notification for movie %s: %v", movie.Title, err)
			}
		}
	}
}
