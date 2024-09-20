package scheduler

import (
	"context"
	"encoding/json"
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
	"github.com/mahcks/blockbusterr/internal/notifications"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

type radarrJob struct {
	gctx              global.Context
	helpers           helpers.Helpers
	ombiSettings      db.OmbiSettings
	radarrSettings    db.RadarrSettings
	movieSettings     db.MovieSettings
	anticipatedMovies []trakt.Movie
	boxOfficeMovies   []trakt.Movie
	popularMovies     []trakt.Movie
	trendingMovies    []trakt.Movie
}

// AnticipatedJobFunc fetches and processes anticipated movies
func (s Scheduler) AnticipatedJobFunc() {
	err := s.helpers.Trakt.Ping(context.Background())
	if err != nil {
		log.Warn("[Scheduler] Skipping 'Anticipated Movies' job. Trakt credentials are missing.")
		return
	}

	log.Info("[Scheduler] Starting 'Anticipated Movies' job...")
	startTime := time.Now()

	mj := radarrJob{
		gctx:    s.gctx,
		helpers: s.helpers,
	}

	largeMovieQueryLimit := 1000

	// Initialize movie settings
	if err := s.initializeMovieJob(&mj); err != nil {
		log.Error("[Scheduler] Failed to initialize 'Anticipated Movies' job. Check your settings and try again.", "error", err)
		return
	}

	// Fetch anticipated movies from Trakt
	if mj.movieSettings.Anticipated.Valid && mj.movieSettings.Anticipated.Int32 > 0 {
		params := buildTraktParamsFromSettings(mj.movieSettings, largeMovieQueryLimit, true)
		anticipatedMovies, err := s.helpers.Trakt.GetAnticipatedMovies(s.gctx, params)
		if err != nil {
			if errors.Is(err, trakt.ErrNoTraktSettings) {
				log.Warn("[Scheduler] 'Anticipated Movies' job could not be completed. Trakt Client ID is not set.")
				return
			} else {
				log.Error("[Scheduler] Error fetching 'Anticipated Movies' from Trakt.", "error", err)
				return
			}
		}

		// Process the fetched movies
		mj.anticipatedMovies = filterAndLimitMovies(extractMoviesFromAnticipated(anticipatedMovies), mj.movieSettings, int(mj.movieSettings.Anticipated.Int32))
		s.processMovies(mj.anticipatedMovies, mj)
	}

	log.Infof("[Scheduler] Completed 'Anticipated Movies' job in %.2f seconds.", time.Since(startTime).Seconds())
}

// BoxOfficeJobFunc fetches and processes box office movies
func (s Scheduler) BoxOfficeJobFunc() {
	err := s.helpers.Trakt.Ping(context.Background())
	if err != nil {
		log.Warn("[Scheduler] Skipping 'Box Office Movies' job. Trakt credentials are missing.")
		return
	}

	log.Info("[Scheduler] Starting 'Box Office Movies' job...")
	startTime := time.Now()

	mj := radarrJob{}
	largeMovieQueryLimit := 1000

	// Initialize movie settings
	if err := s.initializeMovieJob(&mj); err != nil {
		log.Error("[Scheduler] Failed to initialize 'Box Office Movies' job. Check your settings and try again.", "error", err)
		return
	}

	// Fetch box office movies from Trakt
	if mj.movieSettings.BoxOffice.Valid && mj.movieSettings.BoxOffice.Int32 > 0 {
		params := buildTraktParamsFromSettings(mj.movieSettings, largeMovieQueryLimit, false)
		boxOfficeMovies, err := s.helpers.Trakt.GetBoxOfficeMovies(s.gctx, params)
		if err != nil {
			if errors.Is(err, trakt.ErrNoTraktSettings) {
				log.Warn("[Scheduler] 'Box Office Movies' job could not be completed. Trakt Client ID is not set.")
				return
			} else {
				log.Error("[Scheduler] Error fetching 'Box Office Movies' from Trakt.", "error", err)
				return
			}
		}

		// Process the fetched movies
		mj.boxOfficeMovies = filterAndLimitMovies(extractMoviesFromBoxOffice(boxOfficeMovies), mj.movieSettings, int(mj.movieSettings.BoxOffice.Int32))
		s.processMovies(mj.boxOfficeMovies, mj)
	}

	log.Infof("[Scheduler] Completed 'Box Office Movies' job in %.2f seconds.", time.Since(startTime).Seconds())
}

// PopularJobFunc fetches and processes popular movies
func (s Scheduler) PopularJobFunc() {
	err := s.helpers.Trakt.Ping(context.Background())
	if err != nil {
		log.Warn("[Scheduler] Skipping 'Popular Movies' job. Trakt credentials are missing.")
		return
	}

	log.Info("[Scheduler] Starting 'Popular Movies' job...")
	startTime := time.Now()

	mj := radarrJob{}
	largeMovieQueryLimit := 1000

	// Initialize movie settings
	if err := s.initializeMovieJob(&mj); err != nil {
		log.Error("[Scheduler] Failed to initialize 'Popular Movies' job. Check your settings and try again.", "error", err)
		return
	}

	// Fetch popular movies from Trakt
	if mj.movieSettings.Popular.Valid && mj.movieSettings.Popular.Int32 > 0 {
		params := buildTraktParamsFromSettings(mj.movieSettings, largeMovieQueryLimit, false)
		popularMovies, err := s.helpers.Trakt.GetPopularMovies(s.gctx, params)
		if err != nil {
			if errors.Is(err, trakt.ErrNoTraktSettings) {
				log.Warn("[Scheduler] 'Popular Movies' job could not be completed. Trakt Client ID is not set.")
				return
			} else {
				log.Error("[Scheduler] Error fetching 'Popular Movies' from Trakt.", "error", err)
				return
			}
		}

		// Process the fetched movies
		mj.popularMovies = filterAndLimitMovies(extractMoviesFromPopular(popularMovies), mj.movieSettings, int(mj.movieSettings.Popular.Int32))
		s.processMovies(mj.popularMovies, mj)
	}

	log.Infof("[Scheduler] Completed 'Popular Movies' job in %.2f seconds.", time.Since(startTime).Seconds())
}

// TrendingJobFunc fetches and processes trending movies
func (s Scheduler) TrendingJobFunc() {
	err := s.helpers.Trakt.Ping(context.Background())
	if err != nil {
		log.Warn("[Scheduler] Skipping 'Trending Movies' job. Trakt credentials are missing.")
		return
	}

	log.Info("[Scheduler] Starting 'Trending Movies' job...")
	startTime := time.Now()

	mj := radarrJob{}
	largeMovieQueryLimit := 1000

	// Initialize movie settings
	if err := s.initializeMovieJob(&mj); err != nil {
		log.Error("[Scheduler] Failed to initialize 'Trending Movies' job. Check your settings and try again.", "error", err)
		return
	}

	// Fetch trending movies from Trakt
	if mj.movieSettings.Trending.Valid && mj.movieSettings.Trending.Int32 > 0 {
		params := buildTraktParamsFromSettings(mj.movieSettings, largeMovieQueryLimit, false)
		trendingMovies, err := s.helpers.Trakt.GetTrendingMovies(s.gctx, params)
		if err != nil {
			if errors.Is(err, trakt.ErrNoTraktSettings) {
				log.Warn("[Scheduler] 'Trending Movies' job could not be completed. Trakt Client ID is not set.")
				return
			} else {
				log.Error("[Scheduler] Error fetching 'Trending Movies' from Trakt.", "error", err)
				return
			}
		}

		// Process the fetched movies
		mj.trendingMovies = filterAndLimitMovies(extractMoviesFromTrending(trendingMovies), mj.movieSettings, int(mj.movieSettings.Trending.Int32))
		s.processMovies(mj.trendingMovies, mj)
	}

	log.Infof("[Scheduler] Completed 'Trending Movies' job in %.2f seconds.", time.Since(startTime).Seconds())
}

// initializeMovieJob handles common setup logic for all movie jobs
func (s Scheduler) initializeMovieJob(mj *radarrJob) error {
	gctx := s.gctx

	// Get all settings for Radarr
	var err error
	mj.radarrSettings, err = gctx.Crate().SQL.Queries().GetRadarrSettings(gctx)
	if err != nil {
		return fmt.Errorf("[Scheduler] Error fetching Radarr settings: %w", err)
	}

	// Get all settings for movies
	mj.movieSettings, err = gctx.Crate().SQL.Queries().GetMovieSettings(gctx)
	if err != nil {
		return fmt.Errorf("[Scheduler] Error fetching movie settings: %w", err)
	}

	return nil
}

// processMovies handles the logic for processing and sending movie requests to Ombi or Radarr
func (s Scheduler) processMovies(movies []trakt.Movie, mj radarrJob) {
	gctx := s.gctx
	helpers := s.helpers

	// Check if Ombi is enabled
	ombiEnabled, err := gctx.Crate().SQL.Queries().GetSettingByKey(gctx, structures.SettingMode.String())
	if err != nil {
		log.Error("[Scheduler] Error fetching Ombi enabled setting.", "error", err)
		return
	}

	// If Ombi is enabled, request movies via Ombi
	if ombiEnabled.Value.String == "ombi" {
		mj.ombiSettings, err = gctx.Crate().SQL.Queries().GetOmbiSettings(gctx)
		if err != nil {
			if errors.Is(err, db.ErrNoOmbiSettings) {
				log.Warn("[Scheduler] Skipping Ombi job. Ombi settings are not configured.")
				return
			}

			log.Error("[Scheduler] Error fetching Ombi settings.", "error", err)
			return
		}

		// Request movies via Ombi
		requestMoviesToOmbi(s.gctx, helpers, s.notifications, movies, mj.ombiSettings)
	} else {
		// Otherwise, use Radarr to request movies
		requestMoviesToRadarr(s.gctx, helpers, s.notifications, movies, mj.radarrSettings)
	}
}

// Helper function to filter and limit movies based on settings
func filterAndLimitMovies(movies []trakt.Movie, settings db.MovieSettings, limit int) []trakt.Movie {
	filteredMovies := applyAdditionalFilters(movies, settings)
	return getTopNMovies(filteredMovies, limit)
}

// Helper function to build Trakt API request parameters from the settings
func buildTraktParamsFromSettings(settings db.MovieSettings, limit int, isAnticipated bool) *trakt.TraktMovieParams {
	params := &trakt.TraktMovieParams{}
	params.Extended = "full"
	params.Limit = limit

	// Apply allowed countries filter
	if len(settings.AllowedCountries) > 0 {
		countries := make([]string, len(settings.AllowedCountries))
		for i, country := range settings.AllowedCountries {
			countries[i] = country.CountryCode
		}
		params.Countries = strings.Join(countries, ",")
	}

	// Apply allowed languages filter
	if len(settings.AllowedLanguages) > 0 {
		languages := make([]string, len(settings.AllowedLanguages))
		for i, language := range settings.AllowedLanguages {
			languages[i] = language.LanguageCode
		}
		params.Languages = strings.Join(languages, ",")
	}

	// Apply runtime filter
	if settings.MinRuntime.Valid || settings.MaxRuntime.Valid {
		params.Runtime = fmt.Sprintf("%d-%d", settings.MinRuntime.Int32, settings.MaxRuntime.Int32)
	}

	// Apply year filter for anticipated or regular items
	if isAnticipated {
		minYear := time.Now().Year()
		maxYear := minYear + 10
		if settings.MinYear.Valid {
			minYear = int(settings.MinYear.Int32)
		}
		params.Years = fmt.Sprintf("%d-%d", minYear, maxYear)
	} else {
		minYear := 0
		maxYear := time.Now().Year()
		if settings.MinYear.Valid {
			minYear = int(settings.MinYear.Int32)
		}
		if settings.MaxYear.Valid {
			maxYear = int(settings.MaxYear.Int32)
		}
		params.Years = fmt.Sprintf("%d-%d", minYear, maxYear)
	}

	return params
}

// Additional filtering logic for movies
func applyAdditionalFilters(movies []trakt.Movie, settings db.MovieSettings) []trakt.Movie {
	filteredMovies := []trakt.Movie{}

	// Build blacklisted genres, keywords, and TMDb IDs from settings
	blacklistedGenres := make(map[string]bool)
	for _, genre := range settings.BlacklistedGenres {
		blacklistedGenres[strings.ToLower(genre.Genre)] = true
	}

	blacklistedKeywords := make(map[string]bool)
	for _, keyword := range settings.BlacklistedTitleKeywords {
		blacklistedKeywords[strings.ToLower(keyword.Keyword)] = true
	}

	blacklistedTMDBIDs := make(map[int]bool)
	for _, tmdbID := range settings.BlacklistedTMDBIDs {
		blacklistedTMDBIDs[tmdbID.TMDBID] = true
	}

	// Apply filters to each movie
	for _, movie := range movies {
		if blacklistedTMDBIDs[movie.IDs.TMDB] {
			continue
		}

		isBlacklisted := false
		for _, genre := range movie.Genres {
			if blacklistedGenres[strings.ToLower(genre)] {
				isBlacklisted = true
				break
			}
		}
		if isBlacklisted {
			continue
		}

		for keyword := range blacklistedKeywords {
			if strings.Contains(strings.ToLower(movie.Title), keyword) {
				isBlacklisted = true
				break
			}
		}
		if isBlacklisted {
			continue
		}

		filteredMovies = append(filteredMovies, movie)
	}

	return filteredMovies
}

func getTopNMovies(movies []trakt.Movie, limit int) []trakt.Movie {
	if len(movies) > limit {
		return movies[:limit]
	}
	return movies
}

// Extract movies from various Trakt types
func extractMoviesFromTrending(trendingMovies []trakt.TrendingMovie) []trakt.Movie {
	var movies []trakt.Movie
	for _, trendingMovie := range trendingMovies {
		movies = append(movies, trendingMovie.Movie)
	}
	return movies
}

func extractMoviesFromPopular(popularMovies []trakt.Movie) []trakt.Movie {
	return popularMovies
}

func extractMoviesFromAnticipated(anticipatedMovies []trakt.TraktAnticipatedMovie) []trakt.Movie {
	var movies []trakt.Movie
	for _, anticipated := range anticipatedMovies {
		movies = append(movies, anticipated.Movie)
	}
	return movies
}

func extractMoviesFromBoxOffice(boxOfficeMovies []trakt.TraktBoxOfficeMovie) []trakt.Movie {
	var movies []trakt.Movie
	for _, boxOffice := range boxOfficeMovies {
		movies = append(movies, boxOffice.Movie)
	}
	return movies
}

func fetchRadarrSettings(r radarr.Service, radarrSettings db.RadarrSettings) (int, string, error) {
	// Fetch quality profiles from Radarr
	qualityProfiles, err := r.GetQualityProfiles(nil, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get Radarr quality profiles: %w", err)
	}

	// Fetch root folders from Radarr
	rootFolders, err := r.GetRootFolders(nil, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get Radarr root folders: %w", err)
	}

	// Match quality profile and root folder from settings
	var qualityProfileID int
	for _, profile := range qualityProfiles {
		if profile.ID == int(radarrSettings.Quality.Int32) {
			qualityProfileID = profile.ID
			break
		}
	}

	var rootFolderPath string
	for _, folder := range rootFolders {
		if folder.ID == int(radarrSettings.RootFolder.Int32) {
			rootFolderPath = folder.Path
			break
		}
	}

	if qualityProfileID == 0 || rootFolderPath == "" {
		return 0, "", fmt.Errorf("invalid quality profile or root folder")
	}

	return qualityProfileID, rootFolderPath, nil
}

// Request movies to Ombi
func requestMoviesToOmbi(gctx global.Context, helpers helpers.Helpers, notifications *notifications.NotificationManager, movies []trakt.Movie, ombiSettings db.OmbiSettings) {
	for _, movie := range movies {
		body := ombi.RequestMovieBody{
			TheMovieDBID: movie.IDs.TMDB,
			LanguageCode: "en",  // Adjust as needed
			Is4KRequest:  false, // Adjust if needed for 4K
		}

		if ombiSettings.UserID.Valid {
			body.RequestOnBehalf = ombiSettings.UserID.String
		}

		if ombiSettings.MovieRootFolder.Valid {
			rootFolder := fmt.Sprintf("%d", ombiSettings.MovieRootFolder.Int32)
			body.RootFolderOverride = &rootFolder
		}

		if ombiSettings.MovieQuality.Valid {
			qualityProfile := fmt.Sprintf("%d", ombiSettings.MovieQuality.Int32)
			body.QualityPathOverride = &qualityProfile
		}

		_, err := helpers.Ombi.RequestMovie(body)
		if err != nil {
			if errors.Is(err, ombi.ErrMovieAlreadyRequested) {
				log.Warnf("[Ombi Job] Skipping '%s' - already requested.", movie.Title)
			} else {
				log.Errorf("[Ombi Job] Failed to request movie '%s': %v", movie.Title, err)
			}
		} else {
			log.Infof("[Ombi Job] Movie '%s' successfully requested.", movie.Title)

			// Fetch and store movie poster
			media, err := helpers.OMDb.GetMedia(context.Background(), movie.IDs.IMDB)
			if err != nil {
				log.Errorf("[Ombi Job] Failed to fetch movie poster for '%s': %v", movie.Title, err)
				continue
			}

			// Add movie to recently added list
			err = gctx.Crate().SQL.Queries().AddToRecentlyAddedMedia(context.Background(), "MOVIE", media.Title, movie.Year, media.Plot, media.IMDBID, media.Poster)
			if err != nil {
				log.Errorf("[Ombi Job] Failed to add movie '%s' to recently added list: %v", movie.Title, err)
			}

			// Send notification
			moviePayload, err := json.Marshal(movie)
			if err != nil {
				log.Errorf("[Ombi Job] Failed to marshal movie payload for notifications: %v", err)
			}
			err = notifications.SendNotification(structures.MOVIEADDEDALERT, moviePayload)
			if err != nil {
				log.Errorf("[Ombi Job] Failed to send notification for movie '%s': %v", movie.Title, err)
			}
		}
	}
}

// Request movies to Radarr
func requestMoviesToRadarr(gctx global.Context, helpers helpers.Helpers, notifications *notifications.NotificationManager, movies []trakt.Movie, radarrSettings db.RadarrSettings) {
	qualityProfileID, rootFolderPath, err := fetchRadarrSettings(helpers.Radarr, radarrSettings)
	if err != nil {
		log.Error("[Radarr Job] Failed to retrieve Radarr settings.", "error", err)
		return
	}

	for _, movie := range movies {
		body := radarr.RequestMovieBody{
			Title:               movie.Title,
			TMDBID:              movie.IDs.TMDB,
			Monitored:           true,
			QualityProfileID:    qualityProfileID,
			RootFolderPath:      rootFolderPath,
			MinimumAvailability: radarrSettings.MinimumAvailability.String, // Adjust as needed
		}

		body.AddOptions.SearchForMovie = true

		_, err := helpers.Radarr.RequestMovie(nil, nil, body)
		if err != nil {
			if errors.Is(err, radarr.ErrMovieAlreadyExists) {
				log.Warnf("[Radarr Job] Skipping '%s' - already exists in Radarr.", movie.Title)
			} else {
				log.Errorf("[Radarr Job] Failed to request movie '%s': %v", movie.Title, err)
			}
		} else {
			log.Infof("[Radarr Job] Movie '%s' successfully requested.", movie.Title)

			// Fetch and store movie poster
			media, err := helpers.OMDb.GetMedia(context.Background(), movie.IDs.IMDB)
			if err != nil {
				log.Errorf("[Radarr Job] Failed to fetch movie poster for '%s': %v", movie.Title, err)
				continue
			}

			// Add movie to recently added list
			err = gctx.Crate().SQL.Queries().AddToRecentlyAddedMedia(context.Background(), "MOVIE", media.Title, movie.Year, media.Plot, media.IMDBID, media.Poster)
			if err != nil {
				log.Errorf("[Radarr Job] Failed to add movie '%s' to recently added list: %v", movie.Title, err)
			}

			// Send notification
			moviePayload, err := json.Marshal(movie)
			if err != nil {
				log.Errorf("[Radarr Job] Failed to marshal movie payload for notifications: %v", err)
			}
			err = notifications.SendNotification(structures.MOVIEADDEDALERT, moviePayload)
			if err != nil {
				log.Errorf("[Radarr Job] Failed to send notification for movie '%s': %v", movie.Title, err)
			}
		}
	}
}
