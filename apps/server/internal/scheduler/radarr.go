package scheduler

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/mahcks/blockbusterr/internal/db"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
	"github.com/mahcks/blockbusterr/internal/helpers/ombi"
	"github.com/mahcks/blockbusterr/internal/helpers/radarr"
	"github.com/mahcks/blockbusterr/internal/helpers/trakt"
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

// RadarrJobFunc defines the logic for the Radarr job
func (s Scheduler) RadarrJobFunc(gctx global.Context, helpers helpers.Helpers) {
	log.Info("[scheduler] Running movie job...")
	var err error

	// Start time tracking
	startTime := time.Now()

	ombiEnabled, err := gctx.Crate().SQL.Queries().GetSettingByKey(gctx, structures.SettingOmbiEnabled.String())
	if err != nil {
		log.Error("[radarr-job] Error getting Ombi enabled setting", "error", err)
		return
	}

	mj := radarrJob{}

	// Step 1. Get all settings from Radarr table
	mj.radarrSettings, err = gctx.Crate().SQL.Queries().GetRadarrSettings(gctx)
	if err != nil {
		log.Errorf("[radarr-job] Error getting Radarr settings: %v", err)
		return
	}

	// Step 2. Get all the settings for movies
	mj.movieSettings, err = gctx.Crate().SQL.Queries().GetMovieSettings(gctx)
	if err != nil {
		log.Error("[radarr-job] Error getting movie settings", "error", err)
		return
	}

	// Query a large number of movies from each list (e.g., 1000)
	largeMovieQueryLimit := 1000

	// Fetch Anticipated Movies
	if mj.movieSettings.Anticipated.Valid && mj.movieSettings.Anticipated.Int32 > 0 {
		params := buildTraktParamsFromSettings(mj.movieSettings, largeMovieQueryLimit)
		anticipatedMovies, err := helpers.Trakt.GetAnticipatedMovies(gctx, params)
		if err != nil {
			log.Error("[movie-job] Error fetching anticipated movies from Trakt", "error", err)
		} else {
			mj.anticipatedMovies = filterAndLimitMovies(extractMoviesFromAnticipated(anticipatedMovies), mj.movieSettings, int(mj.movieSettings.Anticipated.Int32))
		}
	}

	// Fetch BoxOffice Movies
	if mj.movieSettings.BoxOffice.Valid && mj.movieSettings.BoxOffice.Int32 > 0 {
		params := buildTraktParamsFromSettings(mj.movieSettings, largeMovieQueryLimit)
		boxOfficeMovies, err := helpers.Trakt.GetBoxOfficeMovies(gctx, params)
		if err != nil {
			log.Error("[movie-job] Error fetching box office movies from Trakt", "error", err)
		} else {
			mj.boxOfficeMovies = filterAndLimitMovies(extractMoviesFromBoxOffice(boxOfficeMovies), mj.movieSettings, int(mj.movieSettings.BoxOffice.Int32))
		}
	}

	// Fetch Popular Movies
	if mj.movieSettings.Popular.Valid && mj.movieSettings.Popular.Int32 > 0 {
		params := buildTraktParamsFromSettings(mj.movieSettings, largeMovieQueryLimit)
		popularMovies, err := helpers.Trakt.GetPopularMovies(gctx, params)
		if err != nil {
			log.Error("[movie-job] Error fetching popular movies from Trakt", "error", err)
		} else {
			mj.popularMovies = filterAndLimitMovies(extractMoviesFromPopular(popularMovies), mj.movieSettings, int(mj.movieSettings.Popular.Int32))
		}
	}

	// Fetch Trending Movies
	if mj.movieSettings.Trending.Valid && mj.movieSettings.Trending.Int32 > 0 {
		params := buildTraktParamsFromSettings(mj.movieSettings, largeMovieQueryLimit)
		trendingMovies, err := helpers.Trakt.GetTrendingMovies(gctx, params)
		if err != nil {
			log.Error("[movie-job] Error fetching trending movies from Trakt", "error", err)
		} else {
			mj.trendingMovies = filterAndLimitMovies(extractMoviesFromTrending(trendingMovies), mj.movieSettings, int(mj.movieSettings.Trending.Int32))
		}
	}

	if ombiEnabled.Value.String == "true" {
		// Ombi is enabled, use Ombi settings
		mj.ombiSettings, err = gctx.Crate().SQL.Queries().GetOmbiSettings(gctx)
		if err != nil {
			log.Error("[movie-job] Error getting Ombi settings", "error", err)
			return
		}

		// Request movies via Ombi
		requestMoviesToOmbi(helpers.Ombi, mj.anticipatedMovies, mj.ombiSettings)
		requestMoviesToOmbi(helpers.Ombi, mj.boxOfficeMovies, mj.ombiSettings)
		requestMoviesToOmbi(helpers.Ombi, mj.popularMovies, mj.ombiSettings)
		requestMoviesToOmbi(helpers.Ombi, mj.trendingMovies, mj.ombiSettings)

		log.Debug("[ombi-job] Anticipated Movies", "count", len(mj.anticipatedMovies))
		log.Debug("[ombi-job] Box Office Movies", "count", len(mj.boxOfficeMovies))
		log.Debug("[ombi-job] Popular Movies", "count", len(mj.popularMovies))
		log.Debug("[ombi-job] Trending Movies", "count", len(mj.trendingMovies))
	} else {
		// Radarr is used, get Radarr settings
		mj.radarrSettings, err = gctx.Crate().SQL.Queries().GetRadarrSettings(gctx)
		if err != nil {
			log.Error("[movie-job] Error getting Radarr settings", "error", err)
			return
		}
		// Request movies via Radarr
		requestMoviesToRadarr(helpers.Radarr, mj.anticipatedMovies, mj.radarrSettings)
		requestMoviesToRadarr(helpers.Radarr, mj.boxOfficeMovies, mj.radarrSettings)
		requestMoviesToRadarr(helpers.Radarr, mj.popularMovies, mj.radarrSettings)
		requestMoviesToRadarr(helpers.Radarr, mj.trendingMovies, mj.radarrSettings)

		log.Debug("[radarr-job] Anticipated Movies", "count", len(mj.anticipatedMovies))
		log.Debug("[radarr-job] Box Office Movies", "count", len(mj.boxOfficeMovies))
		log.Debug("[radarr-job] Popular Movies", "count", len(mj.popularMovies))
		log.Debug("[radarr-job] Trending Movies", "count", len(mj.trendingMovies))
	}

	duration := time.Since(startTime)
	durationInSeconds := float64(duration.Milliseconds()) / 1000
	roundedDuration := fmt.Sprintf("%.2f", durationInSeconds)

	log.Infof("[scheduler] Completed movie job in %v seconds!", roundedDuration)
}

// Helper function to filter and limit movies based on settings
func filterAndLimitMovies(movies []trakt.Movie, settings db.MovieSettings, limit int) []trakt.Movie {
	filteredMovies := applyAdditionalFilters(movies, settings)
	return getTopNMovies(filteredMovies, limit)
}

// Helper function to build Trakt API request parameters from the settings
func buildTraktParamsFromSettings(settings db.MovieSettings, limit int) *trakt.TraktMovieParams {
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

	// Build genre filters (if needed)
	if len(settings.BlacklistedGenres) > 0 {
		genres := []string{}
		for _, genre := range settings.BlacklistedGenres {
			genres = append(genres, genre.Genre)
		}
		// Invert the blacklisted genres to only request allowed genres
		params.Genres = strings.Join(genres, ",")
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

	// Apply year filter (if applicable)
	if settings.MinYear.Valid || settings.MaxYear.Valid {
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
	qualityProfiles, err := r.GetQualityProfiles()
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
	rootFolders, err := r.GetRootFolders()
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
func requestMoviesToOmbi(o ombi.Service, movies []trakt.Movie, ombiSettings db.OmbiSettings) {
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
		}
	}
}

func requestMoviesToRadarr(r radarr.Service, movies []trakt.Movie, radarrSettings db.RadarrSettings) {
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
		_, err := r.RequestMovie(body)
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
		}
	}
}
