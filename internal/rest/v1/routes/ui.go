package routes

import (
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// determineConfigPath finds the best location to save config file
// Prioritizes Docker data directory, falls back to local directory
func determineConfigPath() string {
	// Try /app/data first (Docker volume mount)
	if _, err := os.Stat("/app/data"); err == nil {
		return "/app/data/config.yaml"
	}

	// Try data directory (local)
	if _, err := os.Stat("./data"); err == nil {
		return "./data/config.yaml"
	}

	// Fall back to current directory
	return "./config.yaml"
}

// RegisterUIRoutes handles web UI routes
func RegisterUIRoutes(rg *RouteGroup, app *fiber.App) {
	// Root route for web UI - redirect to jobs
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("jobs", fiber.Map{
			"Title":   "Blockbusterr - Jobs",
			"Config":  rg.gctx.Config(),
			"Version": rg.gctx.Metadata().Version,
		}, "base")
	})

	// Configuration page route
	app.Get("/config", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title":   "Blockbusterr - Configuration",
			"Config":  rg.gctx.Config(),
			"Version": rg.gctx.Metadata().Version,
		}, "base")
	})

	// Jobs page route (explicit)
	app.Get("/jobs", func(c *fiber.Ctx) error {
		return c.Render("jobs", fiber.Map{
			"Title":   "Blockbusterr - Jobs",
			"Config":  rg.gctx.Config(),
			"Version": rg.gctx.Metadata().Version,
		}, "base")
	})

	// Activity log page route
	app.Get("/activity", func(c *fiber.Ctx) error {
		return c.Render("activity", fiber.Map{
			"Title":   "Blockbusterr - Activity Log",
			"Config":  rg.gctx.Config(),
			"Version": rg.gctx.Metadata().Version,
		}, "base")
	})

	// Config save route
	app.Post("/config/save", func(c *fiber.Ctx) error {
		cfg := rg.gctx.Config()

		// Parse form data
		traktClientID := c.FormValue("trakt.client_id")
		traktClientSecret := c.FormValue("trakt.client_secret")
		tmdbAPIKey := c.FormValue("tmdb.api_key")
		radarrURL := c.FormValue("radarr.url")
		radarrAPIKey := c.FormValue("radarr.api_key")
		radarrQualityProfile := c.FormValue("radarr.quality_profile")
		radarrRootFolder := c.FormValue("radarr.root_folder")
		sonarrURL := c.FormValue("sonarr.url")
		sonarrAPIKey := c.FormValue("sonarr.api_key")
		sonarrQualityProfile := c.FormValue("sonarr.quality_profile")
		sonarrRootFolder := c.FormValue("sonarr.root_folder")
		jellyseerrURL := c.FormValue("jellyseerr.url")
		jellyseerrAPIKey := c.FormValue("jellyseerr.api_key")
		jellyseerrUserID := c.FormValue("jellyseerr.user_id")
		jobsMode := c.FormValue("jobs.mode")
		globalLimitMovies := c.FormValue("jobs.global_limit_movies")
		globalLimitShows := c.FormValue("jobs.global_limit_shows")
		globalPeriod := c.FormValue("jobs.global_period")

		// Update config
		cfg.Trakt.ClientID = traktClientID
		cfg.Trakt.ClientSecret = traktClientSecret
		cfg.TMDB.APIKey = tmdbAPIKey
		cfg.Radarr.URL = radarrURL
		cfg.Radarr.APIKey = radarrAPIKey
		cfg.Radarr.RootFolder = radarrRootFolder
		cfg.Sonarr.URL = sonarrURL
		cfg.Sonarr.APIKey = sonarrAPIKey
		cfg.Sonarr.RootFolder = sonarrRootFolder
		cfg.Jellyseerr.URL = jellyseerrURL
		cfg.Jellyseerr.APIKey = jellyseerrAPIKey
		cfg.Jellyseerr.UserID = jellyseerrUserID
		if jobsMode != "" {
			cfg.Jobs.Mode = jobsMode
		}
		if globalPeriod != "" {
			cfg.Jobs.GlobalPeriod = globalPeriod
		}

		// Parse quality profile IDs
		if radarrQualityProfile != "" {
			if qp, err := strconv.Atoi(radarrQualityProfile); err == nil {
				cfg.Radarr.QualityProfile = qp
			}
		}
		if sonarrQualityProfile != "" {
			if qp, err := strconv.Atoi(sonarrQualityProfile); err == nil {
				cfg.Sonarr.QualityProfile = qp
			}
		}

		// Parse global limits
		if globalLimitMovies != "" {
			if limit, err := strconv.Atoi(globalLimitMovies); err == nil {
				cfg.Jobs.GlobalLimitMovies = limit
			}
		}
		if globalLimitShows != "" {
			if limit, err := strconv.Atoi(globalLimitShows); err == nil {
				cfg.Jobs.GlobalLimitShows = limit
			}
		}

		// Save config to file (create if doesn't exist)
		if err := cfg.Save(); err != nil {
			// If config file doesn't exist, try creating it in the right location
			if cfg.ConfigFilePath == "" {
				// Try Docker data directory first, then fall back to current directory
				cfg.ConfigFilePath = determineConfigPath()
				if err := cfg.Save(); err != nil {
					return c.Status(500).SendString(`
				<script>showNotification('Failed to create configuration file: ` + err.Error() + `', 'error');</script>
			`)
				}
			} else {
				return c.Status(500).SendString(`
				<script>showNotification('Failed to save configuration: ` + err.Error() + `', 'error');</script>
			`)
			}
		}

		// Automatically reload the configuration
		if err := rg.gctx.ReloadConfig(); err != nil {
			return c.Status(500).SendString(`
				<script>showNotification('Configuration saved but failed to reload: ` + err.Error() + `', 'error');</script>
			`)
		}

		return c.SendString(`
			<script>showNotification('Configuration saved and reloaded successfully!', 'success');</script>
		`)
	})

	// Jobs config save route
	app.Post("/jobs/config/save", func(c *fiber.Ctx) error {
		cfg := rg.gctx.Config()

		// Get all form keys to detect which fields were actually submitted
		formData := c.Request().PostArgs()

		// Update global sync interval if provided
		if syncInterval := c.FormValue("jobs.sync_interval"); syncInterval != "" {
			cfg.Jobs.SyncInterval = syncInterval
		}

		// Update global mode if provided
		if mode := c.FormValue("jobs.mode"); mode != "" {
			cfg.Jobs.Mode = mode
		}

		// Helper to check if a field was submitted (even if empty)
		hasField := func(key string) bool {
			return formData.Has(key)
		}

		// Movies - Trending
		if hasField("jobs.trending_movies.enabled") || hasField("jobs.trending_movies.limit") {
			cfg.Jobs.TrendingMovies.Enabled = c.FormValue("jobs.trending_movies.enabled") == "on"
			if limit := c.FormValue("jobs.trending_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.TrendingMovies.Limit = val
				}
			}
			// Per-job sync interval and mode
			if hasField("jobs.trending_movies.sync_interval") {
				cfg.Jobs.TrendingMovies.SyncInterval = c.FormValue("jobs.trending_movies.sync_interval")
			}
			if hasField("jobs.trending_movies.mode") {
				cfg.Jobs.TrendingMovies.Mode = c.FormValue("jobs.trending_movies.mode")
			}
		}

		// Movies - Popular
		if hasField("jobs.popular_movies.enabled") || hasField("jobs.popular_movies.limit") {
			cfg.Jobs.PopularMovies.Enabled = c.FormValue("jobs.popular_movies.enabled") == "on"
			if limit := c.FormValue("jobs.popular_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.PopularMovies.Limit = val
				}
			}
			// Per-job sync interval and mode
			if hasField("jobs.popular_movies.sync_interval") {
				cfg.Jobs.PopularMovies.SyncInterval = c.FormValue("jobs.popular_movies.sync_interval")
			}
			if hasField("jobs.popular_movies.mode") {
				cfg.Jobs.PopularMovies.Mode = c.FormValue("jobs.popular_movies.mode")
			}
		} // Movies - Box Office
		if hasField("jobs.box_office.enabled") || hasField("jobs.box_office.limit") {
			cfg.Jobs.BoxOffice.Enabled = c.FormValue("jobs.box_office.enabled") == "on"
			if limit := c.FormValue("jobs.box_office.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.BoxOffice.Limit = val
				}
			}
			// Per-job sync interval and mode
			if hasField("jobs.box_office.sync_interval") {
				cfg.Jobs.BoxOffice.SyncInterval = c.FormValue("jobs.box_office.sync_interval")
			}
			if hasField("jobs.box_office.mode") {
				cfg.Jobs.BoxOffice.Mode = c.FormValue("jobs.box_office.mode")
			}
		} // Movies - Favorited
		if hasField("jobs.favorited_movies.enabled") || hasField("jobs.favorited_movies.limit") || hasField("jobs.favorited_movies.period") {
			cfg.Jobs.FavoritedMovies.Enabled = c.FormValue("jobs.favorited_movies.enabled") == "on"
			if limit := c.FormValue("jobs.favorited_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.FavoritedMovies.Limit = val
				}
			}
			if period := c.FormValue("jobs.favorited_movies.period"); period != "" {
				cfg.Jobs.FavoritedMovies.Period = period
			}
			// Per-job sync interval and mode
			if hasField("jobs.favorited_movies.sync_interval") {
				cfg.Jobs.FavoritedMovies.SyncInterval = c.FormValue("jobs.favorited_movies.sync_interval")
			}
			if hasField("jobs.favorited_movies.mode") {
				cfg.Jobs.FavoritedMovies.Mode = c.FormValue("jobs.favorited_movies.mode")
			}
		}

		// Movies - Played
		if hasField("jobs.played_movies.enabled") || hasField("jobs.played_movies.limit") || hasField("jobs.played_movies.period") {
			cfg.Jobs.PlayedMovies.Enabled = c.FormValue("jobs.played_movies.enabled") == "on"
			if limit := c.FormValue("jobs.played_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.PlayedMovies.Limit = val
				}
			}
			if period := c.FormValue("jobs.played_movies.period"); period != "" {
				cfg.Jobs.PlayedMovies.Period = period
			}
			// Per-job sync interval and mode
			if hasField("jobs.played_movies.sync_interval") {
				cfg.Jobs.PlayedMovies.SyncInterval = c.FormValue("jobs.played_movies.sync_interval")
			}
			if hasField("jobs.played_movies.mode") {
				cfg.Jobs.PlayedMovies.Mode = c.FormValue("jobs.played_movies.mode")
			}
		}

		// Movies - Watched
		if hasField("jobs.watched_movies.enabled") || hasField("jobs.watched_movies.limit") || hasField("jobs.watched_movies.period") {
			cfg.Jobs.WatchedMovies.Enabled = c.FormValue("jobs.watched_movies.enabled") == "on"
			if limit := c.FormValue("jobs.watched_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.WatchedMovies.Limit = val
				}
			}
			if period := c.FormValue("jobs.watched_movies.period"); period != "" {
				cfg.Jobs.WatchedMovies.Period = period
			}
			// Per-job sync interval and mode
			if hasField("jobs.watched_movies.sync_interval") {
				cfg.Jobs.WatchedMovies.SyncInterval = c.FormValue("jobs.watched_movies.sync_interval")
			}
			if hasField("jobs.watched_movies.mode") {
				cfg.Jobs.WatchedMovies.Mode = c.FormValue("jobs.watched_movies.mode")
			}
		}

		// Movies - Collected
		if hasField("jobs.collected_movies.enabled") || hasField("jobs.collected_movies.limit") || hasField("jobs.collected_movies.period") {
			cfg.Jobs.CollectedMovies.Enabled = c.FormValue("jobs.collected_movies.enabled") == "on"
			if limit := c.FormValue("jobs.collected_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.CollectedMovies.Limit = val
				}
			}
			if period := c.FormValue("jobs.collected_movies.period"); period != "" {
				cfg.Jobs.CollectedMovies.Period = period
			}
			// Per-job sync interval and mode
			if hasField("jobs.collected_movies.sync_interval") {
				cfg.Jobs.CollectedMovies.SyncInterval = c.FormValue("jobs.collected_movies.sync_interval")
			}
			if hasField("jobs.collected_movies.mode") {
				cfg.Jobs.CollectedMovies.Mode = c.FormValue("jobs.collected_movies.mode")
			}
		}

		// Movies - Anticipated
		if hasField("jobs.anticipated_movies.enabled") || hasField("jobs.anticipated_movies.limit") {
			cfg.Jobs.AnticipatedMovies.Enabled = c.FormValue("jobs.anticipated_movies.enabled") == "on"
			if limit := c.FormValue("jobs.anticipated_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.AnticipatedMovies.Limit = val
				}
			}
			// Per-job sync interval and mode
			if hasField("jobs.anticipated_movies.sync_interval") {
				cfg.Jobs.AnticipatedMovies.SyncInterval = c.FormValue("jobs.anticipated_movies.sync_interval")
			}
			if hasField("jobs.anticipated_movies.mode") {
				cfg.Jobs.AnticipatedMovies.Mode = c.FormValue("jobs.anticipated_movies.mode")
			}
		}

		// Shows - Trending
		if hasField("jobs.trending_shows.enabled") || hasField("jobs.trending_shows.limit") {
			cfg.Jobs.TrendingShows.Enabled = c.FormValue("jobs.trending_shows.enabled") == "on"
			if limit := c.FormValue("jobs.trending_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.TrendingShows.Limit = val
				}
			}
			// Per-job sync interval and mode
			if hasField("jobs.trending_shows.sync_interval") {
				cfg.Jobs.TrendingShows.SyncInterval = c.FormValue("jobs.trending_shows.sync_interval")
			}
			if hasField("jobs.trending_shows.mode") {
				cfg.Jobs.TrendingShows.Mode = c.FormValue("jobs.trending_shows.mode")
			}
		} // Shows - Popular
		if hasField("jobs.popular_shows.enabled") || hasField("jobs.popular_shows.limit") {
			cfg.Jobs.PopularShows.Enabled = c.FormValue("jobs.popular_shows.enabled") == "on"
			if limit := c.FormValue("jobs.popular_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.PopularShows.Limit = val
				}
			}
			// Per-job sync interval and mode
			if hasField("jobs.popular_shows.sync_interval") {
				cfg.Jobs.PopularShows.SyncInterval = c.FormValue("jobs.popular_shows.sync_interval")
			}
			if hasField("jobs.popular_shows.mode") {
				cfg.Jobs.PopularShows.Mode = c.FormValue("jobs.popular_shows.mode")
			}
		} // Shows - Favorited
		if hasField("jobs.favorited_shows.enabled") || hasField("jobs.favorited_shows.limit") || hasField("jobs.favorited_shows.period") {
			cfg.Jobs.FavoritedShows.Enabled = c.FormValue("jobs.favorited_shows.enabled") == "on"
			if limit := c.FormValue("jobs.favorited_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.FavoritedShows.Limit = val
				}
			}
			if period := c.FormValue("jobs.favorited_shows.period"); period != "" {
				cfg.Jobs.FavoritedShows.Period = period
			}
			// Per-job sync interval and mode
			if hasField("jobs.favorited_shows.sync_interval") {
				cfg.Jobs.FavoritedShows.SyncInterval = c.FormValue("jobs.favorited_shows.sync_interval")
			}
			if hasField("jobs.favorited_shows.mode") {
				cfg.Jobs.FavoritedShows.Mode = c.FormValue("jobs.favorited_shows.mode")
			}
		}

		// Shows - Played
		if hasField("jobs.played_shows.enabled") || hasField("jobs.played_shows.limit") || hasField("jobs.played_shows.period") {
			cfg.Jobs.PlayedShows.Enabled = c.FormValue("jobs.played_shows.enabled") == "on"
			if limit := c.FormValue("jobs.played_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.PlayedShows.Limit = val
				}
			}
			if period := c.FormValue("jobs.played_shows.period"); period != "" {
				cfg.Jobs.PlayedShows.Period = period
			}
			// Per-job sync interval and mode
			if hasField("jobs.played_shows.sync_interval") {
				cfg.Jobs.PlayedShows.SyncInterval = c.FormValue("jobs.played_shows.sync_interval")
			}
			if hasField("jobs.played_shows.mode") {
				cfg.Jobs.PlayedShows.Mode = c.FormValue("jobs.played_shows.mode")
			}
		}

		// Shows - Watched
		if hasField("jobs.watched_shows.enabled") || hasField("jobs.watched_shows.limit") || hasField("jobs.watched_shows.period") {
			cfg.Jobs.WatchedShows.Enabled = c.FormValue("jobs.watched_shows.enabled") == "on"
			if limit := c.FormValue("jobs.watched_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.WatchedShows.Limit = val
				}
			}
			if period := c.FormValue("jobs.watched_shows.period"); period != "" {
				cfg.Jobs.WatchedShows.Period = period
			}
			// Per-job sync interval and mode
			if hasField("jobs.watched_shows.sync_interval") {
				cfg.Jobs.WatchedShows.SyncInterval = c.FormValue("jobs.watched_shows.sync_interval")
			}
			if hasField("jobs.watched_shows.mode") {
				cfg.Jobs.WatchedShows.Mode = c.FormValue("jobs.watched_shows.mode")
			}
		}

		// Shows - Collected
		if hasField("jobs.collected_shows.enabled") || hasField("jobs.collected_shows.limit") || hasField("jobs.collected_shows.period") {
			cfg.Jobs.CollectedShows.Enabled = c.FormValue("jobs.collected_shows.enabled") == "on"
			if limit := c.FormValue("jobs.collected_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.CollectedShows.Limit = val
				}
			}
			if period := c.FormValue("jobs.collected_shows.period"); period != "" {
				cfg.Jobs.CollectedShows.Period = period
			}
			// Per-job sync interval and mode
			if hasField("jobs.collected_shows.sync_interval") {
				cfg.Jobs.CollectedShows.SyncInterval = c.FormValue("jobs.collected_shows.sync_interval")
			}
			if hasField("jobs.collected_shows.mode") {
				cfg.Jobs.CollectedShows.Mode = c.FormValue("jobs.collected_shows.mode")
			}
		}

		// Shows - Anticipated
		if hasField("jobs.anticipated_shows.enabled") || hasField("jobs.anticipated_shows.limit") {
			cfg.Jobs.AnticipatedShows.Enabled = c.FormValue("jobs.anticipated_shows.enabled") == "on"
			if limit := c.FormValue("jobs.anticipated_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.AnticipatedShows.Limit = val
				}
			}
			// Per-job sync interval and mode
			if hasField("jobs.anticipated_shows.sync_interval") {
				cfg.Jobs.AnticipatedShows.SyncInterval = c.FormValue("jobs.anticipated_shows.sync_interval")
			}
			if hasField("jobs.anticipated_shows.mode") {
				cfg.Jobs.AnticipatedShows.Mode = c.FormValue("jobs.anticipated_shows.mode")
			}
		}

		// Save config to file
		if err := cfg.Save(); err != nil {
			// If config file doesn't exist, try creating it in the right location
			if cfg.ConfigFilePath == "" {
				cfg.ConfigFilePath = determineConfigPath()
				if err := cfg.Save(); err != nil {
					return c.Status(500).SendString(`
						<script>showNotification('Failed to save job configuration: ` + err.Error() + `', 'error');</script>
					`)
				}
			} else {
				return c.Status(500).SendString(`
					<script>showNotification('Failed to save job configuration: ` + err.Error() + `', 'error');</script>
				`)
			}
		}

		// Automatically reload the configuration
		if err := rg.gctx.ReloadConfig(); err != nil {
			return c.Status(500).SendString(`
				<script>showNotification('Job configuration saved but failed to reload: ` + err.Error() + `', 'error');</script>
			`)
		}

		return c.SendString(`
			<script>
				showNotification('Job configuration saved and reloaded successfully!', 'success');
				setTimeout(function() { window.location.reload(); }, 500);
			</script>
		`)
	})

	// Restart application route
	app.Post("/config/restart", func(c *fiber.Ctx) error {
		// Reload the configuration from file
		if err := rg.gctx.ReloadConfig(); err != nil {
			return c.Status(500).SendString(`
				<div class="bg-red-900 border border-red-600 text-red-200 p-4 rounded-lg">
					Failed to reload configuration: ` + err.Error() + `
				</div>
			`)
		}

		return c.SendString(`
			<div class="bg-green-900 border border-green-600 text-green-200 p-4 rounded-lg">
				Configuration reloaded successfully! Changes are now active.
			</div>
		`)
	})

	// Filters page route
	app.Get("/filters", func(c *fiber.Ctx) error {
		return c.Render("filters", fiber.Map{
			"Title":   "Blockbusterr - Filters",
			"Config":  rg.gctx.Config(),
			"Version": rg.gctx.Metadata().Version,
		}, "base")
	})

	// Filters save route
	app.Post("/config/filters", func(c *fiber.Ctx) error {
		cfg := rg.gctx.Config()

		// Parse movie filters - get all values for multi-select fields
		cfg.Filters.Movies.AllowedCountries = bytesArrayToStrings(c.Request().PostArgs().PeekMulti("movie_allowed_countries"))
		cfg.Filters.Movies.AllowedLanguages = bytesArrayToStrings(c.Request().PostArgs().PeekMulti("movie_allowed_languages"))
		cfg.Filters.Movies.BlacklistedGenres = bytesArrayToStrings(c.Request().PostArgs().PeekMulti("movie_blacklisted_genres"))
		cfg.Filters.Movies.BlacklistedKeywords = bytesArrayToStrings(c.Request().PostArgs().PeekMulti("movie_blacklisted_keywords"))

		// Parse TMDB IDs
		idStrs := bytesArrayToStrings(c.Request().PostArgs().PeekMulti("movie_blacklisted_ids"))
		cfg.Filters.Movies.BlacklistedTMDBIds = make([]int, 0, len(idStrs))
		for _, idStr := range idStrs {
			if id, err := strconv.Atoi(idStr); err == nil {
				cfg.Filters.Movies.BlacklistedTMDBIds = append(cfg.Filters.Movies.BlacklistedTMDBIds, id)
			}
		}

		// Parse runtime and year ranges
		if val := c.FormValue("movie_min_runtime"); val != "" {
			if v, err := strconv.Atoi(val); err == nil {
				cfg.Filters.Movies.BlacklistedMinRuntime = v
			}
		} else {
			cfg.Filters.Movies.BlacklistedMinRuntime = 0
		}

		if val := c.FormValue("movie_max_runtime"); val != "" {
			if v, err := strconv.Atoi(val); err == nil {
				cfg.Filters.Movies.BlacklistedMaxRuntime = v
			}
		} else {
			cfg.Filters.Movies.BlacklistedMaxRuntime = 0
		}

		if val := c.FormValue("movie_min_year"); val != "" {
			if v, err := strconv.Atoi(val); err == nil {
				cfg.Filters.Movies.BlacklistedMinYear = v
			}
		} else {
			cfg.Filters.Movies.BlacklistedMinYear = 0
		}

		if val := c.FormValue("movie_max_year"); val != "" {
			if v, err := strconv.Atoi(val); err == nil {
				cfg.Filters.Movies.BlacklistedMaxYear = v
			}
		} else {
			cfg.Filters.Movies.BlacklistedMaxYear = 0
		}

		// Parse rating and votes filters
		if val := c.FormValue("movie_min_rating"); val != "" {
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				cfg.Filters.Movies.MinRating = v
			}
		} else {
			cfg.Filters.Movies.MinRating = 0
		}

		if val := c.FormValue("movie_min_votes"); val != "" {
			if v, err := strconv.Atoi(val); err == nil {
				cfg.Filters.Movies.MinVotes = v
			}
		} else {
			cfg.Filters.Movies.MinVotes = 0
		}

		// Parse show filters
		cfg.Filters.Shows.AllowedCountries = bytesArrayToStrings(c.Request().PostArgs().PeekMulti("show_allowed_countries"))
		cfg.Filters.Shows.AllowedLanguages = bytesArrayToStrings(c.Request().PostArgs().PeekMulti("show_allowed_languages"))
		cfg.Filters.Shows.BlacklistedGenres = bytesArrayToStrings(c.Request().PostArgs().PeekMulti("show_blacklisted_genres"))
		cfg.Filters.Shows.BlacklistedKeywords = bytesArrayToStrings(c.Request().PostArgs().PeekMulti("show_blacklisted_keywords"))
		cfg.Filters.Shows.BlacklistedNetworks = bytesArrayToStrings(c.Request().PostArgs().PeekMulti("show_blacklisted_networks"))

		// Parse TVDB IDs
		showIdStrs := bytesArrayToStrings(c.Request().PostArgs().PeekMulti("show_blacklisted_ids"))
		cfg.Filters.Shows.BlacklistedTVDBIds = make([]int, 0, len(showIdStrs))
		for _, idStr := range showIdStrs {
			if id, err := strconv.Atoi(idStr); err == nil {
				cfg.Filters.Shows.BlacklistedTVDBIds = append(cfg.Filters.Shows.BlacklistedTVDBIds, id)
			}
		}

		// Parse runtime and year ranges
		if val := c.FormValue("show_min_runtime"); val != "" {
			if v, err := strconv.Atoi(val); err == nil {
				cfg.Filters.Shows.BlacklistedMinRuntime = v
			}
		} else {
			cfg.Filters.Shows.BlacklistedMinRuntime = 0
		}

		if val := c.FormValue("show_max_runtime"); val != "" {
			if v, err := strconv.Atoi(val); err == nil {
				cfg.Filters.Shows.BlacklistedMaxRuntime = v
			}
		} else {
			cfg.Filters.Shows.BlacklistedMaxRuntime = 0
		}

		if val := c.FormValue("show_min_year"); val != "" {
			if v, err := strconv.Atoi(val); err == nil {
				cfg.Filters.Shows.BlacklistedMinYear = v
			}
		} else {
			cfg.Filters.Shows.BlacklistedMinYear = 0
		}

		if val := c.FormValue("show_max_year"); val != "" {
			if v, err := strconv.Atoi(val); err == nil {
				cfg.Filters.Shows.BlacklistedMaxYear = v
			}
		} else {
			cfg.Filters.Shows.BlacklistedMaxYear = 0
		}

		// Parse rating and votes filters
		if val := c.FormValue("show_min_rating"); val != "" {
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				cfg.Filters.Shows.MinRating = v
			}
		} else {
			cfg.Filters.Shows.MinRating = 0
		}

		if val := c.FormValue("show_min_votes"); val != "" {
			if v, err := strconv.Atoi(val); err == nil {
				cfg.Filters.Shows.MinVotes = v
			}
		} else {
			cfg.Filters.Shows.MinVotes = 0
		}

		// Save config
		if err := cfg.Save(); err != nil {
			// If config file doesn't exist, try creating it in the right location
			if cfg.ConfigFilePath == "" {
				cfg.ConfigFilePath = determineConfigPath()
				if err := cfg.Save(); err != nil {
					return c.SendString(`<div class="bg-red-500 text-white px-6 py-3 rounded-lg">Error saving filters: ` + err.Error() + `</div>`)
				}
			} else {
				return c.SendString(`<div class="bg-red-500 text-white px-6 py-3 rounded-lg">Error saving filters: ` + err.Error() + `</div>`)
			}
		}

		// Reload config
		if err := rg.gctx.ReloadConfig(); err != nil {
			return c.SendString(`<div class="bg-red-500 text-white px-6 py-3 rounded-lg">Error reloading config: ` + err.Error() + `</div>`)
		}

		return c.SendString(`<div class="bg-green-500 text-white px-6 py-3 rounded-lg">✓ Filters saved successfully</div>`)
	})
}

// Helper functions
func bytesArrayToStrings(bytesArray [][]byte) []string {
	result := make([]string, len(bytesArray))
	for i, b := range bytesArray {
		result[i] = string(b)
	}
	return result
}
