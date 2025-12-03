package routes

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

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

		// Update config
		cfg.Trakt.ClientID = traktClientID
		cfg.Trakt.ClientSecret = traktClientSecret
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

		// Save config to file (create if doesn't exist)
		if err := cfg.Save(); err != nil {
			// If config file doesn't exist, try creating it
			if cfg.ConfigFilePath == "" {
				cfg.ConfigFilePath = "./config.yaml"
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

		// Update sync interval if provided
		if syncInterval := c.FormValue("jobs.sync_interval"); syncInterval != "" {
			cfg.Jobs.SyncInterval = syncInterval
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
		}

		// Movies - Popular
		if hasField("jobs.popular_movies.enabled") || hasField("jobs.popular_movies.limit") {
			cfg.Jobs.PopularMovies.Enabled = c.FormValue("jobs.popular_movies.enabled") == "on"
			if limit := c.FormValue("jobs.popular_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.PopularMovies.Limit = val
				}
			}
		}

		// Movies - Box Office
		if hasField("jobs.box_office.enabled") || hasField("jobs.box_office.limit") {
			cfg.Jobs.BoxOffice.Enabled = c.FormValue("jobs.box_office.enabled") == "on"
			if limit := c.FormValue("jobs.box_office.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.BoxOffice.Limit = val
				}
			}
		}

		// Movies - Favorited
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
		}

		// Movies - Anticipated
		if hasField("jobs.anticipated_movies.enabled") || hasField("jobs.anticipated_movies.limit") {
			cfg.Jobs.AnticipatedMovies.Enabled = c.FormValue("jobs.anticipated_movies.enabled") == "on"
			if limit := c.FormValue("jobs.anticipated_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.AnticipatedMovies.Limit = val
				}
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
		}

		// Shows - Popular
		if hasField("jobs.popular_shows.enabled") || hasField("jobs.popular_shows.limit") {
			cfg.Jobs.PopularShows.Enabled = c.FormValue("jobs.popular_shows.enabled") == "on"
			if limit := c.FormValue("jobs.popular_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.PopularShows.Limit = val
				}
			}
		}

		// Shows - Favorited
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
		}

		// Shows - Anticipated
		if hasField("jobs.anticipated_shows.enabled") || hasField("jobs.anticipated_shows.limit") {
			cfg.Jobs.AnticipatedShows.Enabled = c.FormValue("jobs.anticipated_shows.enabled") == "on"
			if limit := c.FormValue("jobs.anticipated_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.AnticipatedShows.Limit = val
				}
			}
		}

		// Save config to file
		if err := cfg.Save(); err != nil {
			return c.Status(500).SendString(`
				<script>showNotification('Failed to save job configuration: ` + err.Error() + `', 'error');</script>
			`)
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
}
