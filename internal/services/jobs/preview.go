package jobs

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/filters"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// getTMDBPosterURL fetches the actual poster URL from TMDB API
func getTMDBPosterURL(cfg *config.Config, tmdbID int, mediaType string) string {
	if cfg.TMDB.APIKey == "" || tmdbID == 0 {
		return ""
	}

	tmdbClient := integrations.NewTMDB(integrations.TMDBConfig{
		APIKey: cfg.TMDB.APIKey,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var posterURL string
	var err error

	if mediaType == "movie" {
		posterURL, err = tmdbClient.GetMoviePosterURL(ctx, tmdbID)
	} else {
		posterURL, err = tmdbClient.GetShowPosterURL(ctx, tmdbID)
	}

	if err != nil {
		// Silently fail when TMDB is not available
		return ""
	}

	return posterURL
}

// hasTMDBConfigured checks if TMDB API key is configured
func hasTMDBConfigured(cfg *config.Config) bool {
	return cfg.TMDB.APIKey != ""
}

// createMoviePreviewItem creates a PreviewItem from a Movie
func createMoviePreviewItem(cfg *config.Config, movie integrations.Movie, popularity int) PreviewItem {
	return PreviewItem{
		Title:      movie.Title,
		Year:       movie.Year,
		TMDBID:     movie.IDs.TMDB,
		IMDBID:     movie.IDs.IMDB,
		PosterURL:  getTMDBPosterURL(cfg, movie.IDs.TMDB, "movie"),
		Overview:   movie.Overview,
		Rating:     movie.Rating,
		Votes:      movie.Votes,
		Popularity: popularity,
		Genres:     movie.Genres,
		Runtime:    movie.Runtime,
	}
}

// createShowPreviewItem creates a PreviewItem from a Show
func createShowPreviewItem(cfg *config.Config, show integrations.Show, popularity int) PreviewItem {
	return PreviewItem{
		Title:      show.Title,
		Year:       show.Year,
		TMDBID:     show.IDs.TMDB,
		TVDBID:     show.IDs.TVDB,
		IMDBID:     show.IDs.IMDB,
		PosterURL:  getTMDBPosterURL(cfg, show.IDs.TMDB, "tv"),
		Overview:   show.Overview,
		Rating:     show.Rating,
		Votes:      show.Votes,
		Popularity: popularity,
		Genres:     show.Genres,
		Runtime:    show.Runtime,
	}
}

// PreviewItem represents a single item in the preview
type PreviewItem struct {
	Title         string   `json:"title"`
	Year          int      `json:"year"`
	TMDBID        int      `json:"tmdb_id,omitempty"`
	TVDBID        int      `json:"tvdb_id,omitempty"`
	IMDBID        string   `json:"imdb_id,omitempty"`
	PosterURL     string   `json:"poster_url,omitempty"`
	Overview      string   `json:"overview,omitempty"`
	Rating        float64  `json:"rating,omitempty"`
	Votes         int      `json:"votes,omitempty"`
	Popularity    int      `json:"popularity,omitempty"`
	Genres        []string `json:"genres,omitempty"`
	Runtime       int      `json:"runtime,omitempty"`
	AlreadyExists bool     `json:"already_exists"`
	FilteredOut   bool     `json:"filtered_out,omitempty"`
	FilterReason  string   `json:"filter_reason,omitempty"`
}

// PreviewResponse represents the response for a job preview
type PreviewResponse struct {
	JobName       string        `json:"job_name"`
	TotalFound    int           `json:"total_found"`
	WillAdd       int           `json:"will_add"`
	AlreadyExists int           `json:"already_exists"`
	FilteredOut   int           `json:"filtered_out"`
	Items         []PreviewItem `json:"items"`
	Mode          string        `json:"mode"`        // "direct" or "jellyseerr"
	HasPosters    bool          `json:"has_posters"` // whether TMDB is configured for poster images
}

// PreviewTrendingMovies previews what the trending movies job would add
func PreviewTrendingMovies(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "trending_movies",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.TrendingMovies.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch trending movies
	trendingMovies, err := traktClient.GetTrendingMovies(ctx, cfg.Jobs.TrendingMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch trending movies: %v", err)
		return response
	}

	response.TotalFound = len(trendingMovies)

	// Check each movie against filters and existing content
	for _, tm := range trendingMovies {
		movie := tm.Movie
		item := createMoviePreviewItem(cfg, movie, tm.Watchers)

		// Check filters
		passes, reason := filters.MoviePassesFilters(movie, cfg.Filters.Movies)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			mediaInfo, err := jellyseerrClient.GetMovieInfo(movie.IDs.TMDB)
			if err == nil && mediaInfo.HasMediaInfo() {
				item.AlreadyExists = true
				response.AlreadyExists++
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Radarr
			radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
				BaseURL: cfg.Radarr.URL,
				APIKey:  cfg.Radarr.APIKey,
			})
			existingMovies, err := radarrClient.GetMovies(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingMovies {
					if existing.TmdbID == movie.IDs.TMDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewTrendingShows previews what the trending shows job would add
func PreviewTrendingShows(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "trending_shows",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.TrendingShows.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch trending shows
	trendingShows, err := traktClient.GetTrendingShows(ctx, cfg.Jobs.TrendingShows.Limit)
	if err != nil {
		log.Errorf("Failed to fetch trending shows: %v", err)
		return response
	}

	response.TotalFound = len(trendingShows)

	// Check each show against filters and existing content
	for _, ts := range trendingShows {
		show := ts.Show
		item := createShowPreviewItem(cfg, show, ts.Watchers)

		// Check filters
		passes, reason := filters.ShowPassesFilters(show, cfg.Filters.Shows)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			if show.IDs.TMDB > 0 {
				mediaInfo, err := jellyseerrClient.GetShowInfo(show.IDs.TMDB)
				if err == nil && mediaInfo.HasMediaInfo() {
					item.AlreadyExists = true
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Sonarr
			sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
				BaseURL: cfg.Sonarr.URL,
				APIKey:  cfg.Sonarr.APIKey,
			})
			existingSeries, err := sonarrClient.GetSeries(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingSeries {
					if existing.TvdbID == show.IDs.TVDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewPopularMovies previews what the popular movies job would add
func PreviewPopularMovies(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "popular_movies",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.PopularMovies.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch popular movies
	popularMovies, err := traktClient.GetPopularMovies(ctx, cfg.Jobs.PopularMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch popular movies: %v", err)
		return response
	}

	response.TotalFound = len(popularMovies)

	// Check each movie against filters and existing content
	for i, movie := range popularMovies {
		item := createMoviePreviewItem(cfg, movie, len(popularMovies)-i) // Use reverse index as popularity

		// Check filters
		passes, reason := filters.MoviePassesFilters(movie, cfg.Filters.Movies)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			mediaInfo, err := jellyseerrClient.GetMovieInfo(movie.IDs.TMDB)
			if err == nil && mediaInfo.HasMediaInfo() {
				item.AlreadyExists = true
				response.AlreadyExists++
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Radarr
			radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
				BaseURL: cfg.Radarr.URL,
				APIKey:  cfg.Radarr.APIKey,
			})
			existingMovies, err := radarrClient.GetMovies(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingMovies {
					if existing.TmdbID == movie.IDs.TMDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewPopularShows previews what the popular shows job would add
func PreviewPopularShows(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "popular_shows",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.PopularShows.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch popular shows
	popularShows, err := traktClient.GetPopularShows(ctx, cfg.Jobs.PopularShows.Limit)
	if err != nil {
		log.Errorf("Failed to fetch popular shows: %v", err)
		return response
	}

	response.TotalFound = len(popularShows)

	// Check each show against filters and existing content
	for i, show := range popularShows {
		item := createShowPreviewItem(cfg, show, len(popularShows)-i)

		// Check filters
		passes, reason := filters.ShowPassesFilters(show, cfg.Filters.Shows)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			if show.IDs.TMDB > 0 {
				mediaInfo, err := jellyseerrClient.GetShowInfo(show.IDs.TMDB)
				if err == nil && mediaInfo.HasMediaInfo() {
					item.AlreadyExists = true
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Sonarr
			sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
				BaseURL: cfg.Sonarr.URL,
				APIKey:  cfg.Sonarr.APIKey,
			})
			existingSeries, err := sonarrClient.GetSeries(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingSeries {
					if existing.TvdbID == show.IDs.TVDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewBoxOffice previews what the box office job would add
func PreviewBoxOffice(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "box_office",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.BoxOffice.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch box office movies
	boxOfficeMovies, err := traktClient.GetBoxOfficeMovies(ctx, cfg.Jobs.BoxOffice.Limit)
	if err != nil {
		log.Errorf("Failed to fetch box office movies: %v", err)
		return response
	}

	response.TotalFound = len(boxOfficeMovies)

	// Check each movie against filters and existing content
	for _, bom := range boxOfficeMovies {
		movie := bom.Movie
		item := createMoviePreviewItem(cfg, movie, bom.Revenue)

		// Check filters
		passes, reason := filters.MoviePassesFilters(movie, cfg.Filters.Movies)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			mediaInfo, err := jellyseerrClient.GetMovieInfo(movie.IDs.TMDB)
			if err == nil && mediaInfo.HasMediaInfo() {
				item.AlreadyExists = true
				response.AlreadyExists++
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Radarr
			radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
				BaseURL: cfg.Radarr.URL,
				APIKey:  cfg.Radarr.APIKey,
			})
			existingMovies, err := radarrClient.GetMovies(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingMovies {
					if existing.TmdbID == movie.IDs.TMDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewFavoritedMovies previews what the favorited movies job would add
func PreviewFavoritedMovies(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "favorited_movies",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.FavoritedMovies.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch favorited movies
	favoritedMovies, err := traktClient.GetFavoritedMovies(ctx, cfg.Jobs.FavoritedMovies.Period, cfg.Jobs.FavoritedMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch favorited movies: %v", err)
		return response
	}

	response.TotalFound = len(favoritedMovies)

	// Check each movie against filters and existing content
	for _, fm := range favoritedMovies {
		movie := fm.Movie
		item := createMoviePreviewItem(cfg, movie, fm.UserCount)

		// Check filters
		passes, reason := filters.MoviePassesFilters(movie, cfg.Filters.Movies)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			mediaInfo, err := jellyseerrClient.GetMovieInfo(movie.IDs.TMDB)
			if err == nil && mediaInfo.HasMediaInfo() {
				item.AlreadyExists = true
				response.AlreadyExists++
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Radarr
			radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
				BaseURL: cfg.Radarr.URL,
				APIKey:  cfg.Radarr.APIKey,
			})
			existingMovies, err := radarrClient.GetMovies(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingMovies {
					if existing.TmdbID == movie.IDs.TMDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewPlayedMovies previews what the played movies job would add
func PreviewPlayedMovies(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "played_movies",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.PlayedMovies.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch played movies
	playedMovies, err := traktClient.GetPlayedMovies(ctx, cfg.Jobs.PlayedMovies.Period, cfg.Jobs.PlayedMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch played movies: %v", err)
		return response
	}

	response.TotalFound = len(playedMovies)

	// Check each movie against filters and existing content
	for _, pm := range playedMovies {
		movie := pm.Movie
		item := createMoviePreviewItem(cfg, movie, pm.WatcherCount)

		// Check filters
		passes, reason := filters.MoviePassesFilters(movie, cfg.Filters.Movies)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			mediaInfo, err := jellyseerrClient.GetMovieInfo(movie.IDs.TMDB)
			if err == nil && mediaInfo.HasMediaInfo() {
				item.AlreadyExists = true
				response.AlreadyExists++
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Radarr
			radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
				BaseURL: cfg.Radarr.URL,
				APIKey:  cfg.Radarr.APIKey,
			})
			existingMovies, err := radarrClient.GetMovies(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingMovies {
					if existing.TmdbID == movie.IDs.TMDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewWatchedMovies previews what the watched movies job would add
func PreviewWatchedMovies(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "watched_movies",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.WatchedMovies.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch watched movies
	watchedMovies, err := traktClient.GetWatchedMovies(ctx, cfg.Jobs.WatchedMovies.Period, cfg.Jobs.WatchedMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch watched movies: %v", err)
		return response
	}

	response.TotalFound = len(watchedMovies)

	// Check each movie against filters and existing content
	for _, wm := range watchedMovies {
		movie := wm.Movie
		item := createMoviePreviewItem(cfg, movie, wm.WatcherCount)

		// Check filters
		passes, reason := filters.MoviePassesFilters(movie, cfg.Filters.Movies)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			mediaInfo, err := jellyseerrClient.GetMovieInfo(movie.IDs.TMDB)
			if err == nil && mediaInfo.HasMediaInfo() {
				item.AlreadyExists = true
				response.AlreadyExists++
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Radarr
			radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
				BaseURL: cfg.Radarr.URL,
				APIKey:  cfg.Radarr.APIKey,
			})
			existingMovies, err := radarrClient.GetMovies(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingMovies {
					if existing.TmdbID == movie.IDs.TMDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewCollectedMovies previews what the collected movies job would add
func PreviewCollectedMovies(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "collected_movies",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.CollectedMovies.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch collected movies
	collectedMovies, err := traktClient.GetCollectedMovies(ctx, cfg.Jobs.CollectedMovies.Period, cfg.Jobs.CollectedMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch collected movies: %v", err)
		return response
	}

	response.TotalFound = len(collectedMovies)

	// Check each movie against filters and existing content
	for _, cm := range collectedMovies {
		movie := cm.Movie
		item := createMoviePreviewItem(cfg, movie, cm.CollectedCount)

		// Check filters
		passes, reason := filters.MoviePassesFilters(movie, cfg.Filters.Movies)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			mediaInfo, err := jellyseerrClient.GetMovieInfo(movie.IDs.TMDB)
			if err == nil && mediaInfo.HasMediaInfo() {
				item.AlreadyExists = true
				response.AlreadyExists++
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Radarr
			radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
				BaseURL: cfg.Radarr.URL,
				APIKey:  cfg.Radarr.APIKey,
			})
			existingMovies, err := radarrClient.GetMovies(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingMovies {
					if existing.TmdbID == movie.IDs.TMDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewAnticipatedMovies previews what the anticipated movies job would add
func PreviewAnticipatedMovies(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "anticipated_movies",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.AnticipatedMovies.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch anticipated movies
	anticipatedMovies, err := traktClient.GetAnticipatedMovies(ctx, cfg.Jobs.AnticipatedMovies.Limit)
	if err != nil {
		log.Errorf("Failed to fetch anticipated movies: %v", err)
		return response
	}

	response.TotalFound = len(anticipatedMovies)

	// Check each movie against filters and existing content
	for _, am := range anticipatedMovies {
		movie := am.Movie
		item := createMoviePreviewItem(cfg, movie, am.ListCount)

		// Check filters
		passes, reason := filters.MoviePassesFilters(movie, cfg.Filters.Movies)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			mediaInfo, err := jellyseerrClient.GetMovieInfo(movie.IDs.TMDB)
			if err == nil && mediaInfo.HasMediaInfo() {
				item.AlreadyExists = true
				response.AlreadyExists++
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Radarr
			radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
				BaseURL: cfg.Radarr.URL,
				APIKey:  cfg.Radarr.APIKey,
			})
			existingMovies, err := radarrClient.GetMovies(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingMovies {
					if existing.TmdbID == movie.IDs.TMDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewFavoritedShows previews what the favorited shows job would add
func PreviewFavoritedShows(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "favorited_shows",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.FavoritedShows.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch favorited shows
	favoritedShows, err := traktClient.GetFavoritedShows(ctx, cfg.Jobs.FavoritedShows.Period, cfg.Jobs.FavoritedShows.Limit)
	if err != nil {
		log.Errorf("Failed to fetch favorited shows: %v", err)
		return response
	}

	response.TotalFound = len(favoritedShows)

	// Check each show against filters and existing content
	for _, fs := range favoritedShows {
		show := fs.Show
		item := createShowPreviewItem(cfg, show, fs.UserCount)

		// Check filters
		passes, reason := filters.ShowPassesFilters(show, cfg.Filters.Shows)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			if show.IDs.TMDB > 0 {
				mediaInfo, err := jellyseerrClient.GetShowInfo(show.IDs.TMDB)
				if err == nil && mediaInfo.HasMediaInfo() {
					item.AlreadyExists = true
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Sonarr
			sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
				BaseURL: cfg.Sonarr.URL,
				APIKey:  cfg.Sonarr.APIKey,
			})
			existingSeries, err := sonarrClient.GetSeries(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingSeries {
					if existing.TvdbID == show.IDs.TVDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewPlayedShows previews what the played shows job would add
func PreviewPlayedShows(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "played_shows",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.PlayedShows.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch played shows
	playedShows, err := traktClient.GetPlayedShows(ctx, cfg.Jobs.PlayedShows.Period, cfg.Jobs.PlayedShows.Limit)
	if err != nil {
		log.Errorf("Failed to fetch played shows: %v", err)
		return response
	}

	response.TotalFound = len(playedShows)

	// Check each show against filters and existing content
	for _, ps := range playedShows {
		show := ps.Show
		item := createShowPreviewItem(cfg, show, ps.WatcherCount)

		// Check filters
		passes, reason := filters.ShowPassesFilters(show, cfg.Filters.Shows)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			if show.IDs.TMDB > 0 {
				mediaInfo, err := jellyseerrClient.GetShowInfo(show.IDs.TMDB)
				if err == nil && mediaInfo.HasMediaInfo() {
					item.AlreadyExists = true
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Sonarr
			sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
				BaseURL: cfg.Sonarr.URL,
				APIKey:  cfg.Sonarr.APIKey,
			})
			existingSeries, err := sonarrClient.GetSeries(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingSeries {
					if existing.TvdbID == show.IDs.TVDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewWatchedShows previews what the watched shows job would add
func PreviewWatchedShows(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "watched_shows",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.WatchedShows.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch watched shows
	watchedShows, err := traktClient.GetWatchedShows(ctx, cfg.Jobs.WatchedShows.Period, cfg.Jobs.WatchedShows.Limit)
	if err != nil {
		log.Errorf("Failed to fetch watched shows: %v", err)
		return response
	}

	response.TotalFound = len(watchedShows)

	// Check each show against filters and existing content
	for _, ws := range watchedShows {
		show := ws.Show
		item := createShowPreviewItem(cfg, show, ws.WatcherCount)

		// Check filters
		passes, reason := filters.ShowPassesFilters(show, cfg.Filters.Shows)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			if show.IDs.TMDB > 0 {
				mediaInfo, err := jellyseerrClient.GetShowInfo(show.IDs.TMDB)
				if err == nil && mediaInfo.HasMediaInfo() {
					item.AlreadyExists = true
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Sonarr
			sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
				BaseURL: cfg.Sonarr.URL,
				APIKey:  cfg.Sonarr.APIKey,
			})
			existingSeries, err := sonarrClient.GetSeries(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingSeries {
					if existing.TvdbID == show.IDs.TVDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewCollectedShows previews what the collected shows job would add
func PreviewCollectedShows(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "collected_shows",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.CollectedShows.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch collected shows
	collectedShows, err := traktClient.GetCollectedShows(ctx, cfg.Jobs.CollectedShows.Period, cfg.Jobs.CollectedShows.Limit)
	if err != nil {
		log.Errorf("Failed to fetch collected shows: %v", err)
		return response
	}

	response.TotalFound = len(collectedShows)

	// Check each show against filters and existing content
	for _, cs := range collectedShows {
		show := cs.Show
		item := createShowPreviewItem(cfg, show, cs.CollectedCount)

		// Check filters
		passes, reason := filters.ShowPassesFilters(show, cfg.Filters.Shows)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			if show.IDs.TMDB > 0 {
				mediaInfo, err := jellyseerrClient.GetShowInfo(show.IDs.TMDB)
				if err == nil && mediaInfo.HasMediaInfo() {
					item.AlreadyExists = true
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Sonarr
			sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
				BaseURL: cfg.Sonarr.URL,
				APIKey:  cfg.Sonarr.APIKey,
			})
			existingSeries, err := sonarrClient.GetSeries(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingSeries {
					if existing.TvdbID == show.IDs.TVDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}

// PreviewAnticipatedShows previews what the anticipated shows job would add
func PreviewAnticipatedShows(cfg *config.Config, db *database.Database) PreviewResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response := PreviewResponse{
		HasPosters: hasTMDBConfigured(cfg),
		JobName:    "anticipated_shows",
		Items:      []PreviewItem{},
	}

	// Determine mode
	mode := cfg.Jobs.AnticipatedShows.Mode
	if mode == "" {
		mode = cfg.Jobs.Mode
	}
	if mode == "" {
		mode = "direct"
	}
	response.Mode = mode

	// Create Trakt client
	traktClient := integrations.NewTrakt(integrations.TraktConfig{
		ClientID:     cfg.Trakt.ClientID,
		ClientSecret: cfg.Trakt.ClientSecret,
	})

	// Fetch anticipated shows
	anticipatedShows, err := traktClient.GetAnticipatedShows(ctx, cfg.Jobs.AnticipatedShows.Limit)
	if err != nil {
		log.Errorf("Failed to fetch anticipated shows: %v", err)
		return response
	}

	response.TotalFound = len(anticipatedShows)

	// Check each show against filters and existing content
	for _, as := range anticipatedShows {
		show := as.Show
		item := createShowPreviewItem(cfg, show, as.ListCount)

		// Check filters
		passes, reason := filters.ShowPassesFilters(show, cfg.Filters.Shows)
		if !passes {
			item.FilteredOut = true
			item.FilterReason = reason
			response.FilteredOut++
			response.Items = append(response.Items, item)
			continue
		}

		// Check if already exists (mode-specific)
		if mode == "jellyseerr" {
			jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
				URL:             cfg.Jellyseerr.URL,
				APIKey:          cfg.Jellyseerr.APIKey,
				UserID:          cfg.Jellyseerr.UserID,
				RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
				RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
			})
			if show.IDs.TMDB > 0 {
				mediaInfo, err := jellyseerrClient.GetShowInfo(show.IDs.TMDB)
				if err == nil && mediaInfo.HasMediaInfo() {
					item.AlreadyExists = true
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		} else {
			// Direct mode - check Sonarr
			sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
				BaseURL: cfg.Sonarr.URL,
				APIKey:  cfg.Sonarr.APIKey,
			})
			existingSeries, err := sonarrClient.GetSeries(ctx)
			if err == nil {
				exists := false
				for _, existing := range existingSeries {
					if existing.TvdbID == show.IDs.TVDB {
						exists = true
						break
					}
				}
				item.AlreadyExists = exists
				if exists {
					response.AlreadyExists++
				} else {
					response.WillAdd++
				}
			} else {
				response.WillAdd++
			}
		}

		response.Items = append(response.Items, item)
	}

	return response
}
