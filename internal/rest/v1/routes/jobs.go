package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/services/jobs"
)

func AddJobsRoutes(router fiber.Router, gctx global.Context) {
	router.Get("/jobs/status", func(c *fiber.Ctx) error {
		cfg := gctx.Config()

		status := map[string]interface{}{
			"sync_interval": cfg.Jobs.SyncInterval,
			"jobs": map[string]interface{}{
				"trending_movies": map[string]interface{}{
					"enabled": cfg.Jobs.TrendingMovies.Enabled,
					"limit":   cfg.Jobs.TrendingMovies.Limit,
				},
				"trending_shows": map[string]interface{}{
					"enabled": cfg.Jobs.TrendingShows.Enabled,
					"limit":   cfg.Jobs.TrendingShows.Limit,
				},
				"popular_movies": map[string]interface{}{
					"enabled": cfg.Jobs.PopularMovies.Enabled,
					"limit":   cfg.Jobs.PopularMovies.Limit,
				},
				"popular_shows": map[string]interface{}{
					"enabled": cfg.Jobs.PopularShows.Enabled,
					"limit":   cfg.Jobs.PopularShows.Limit,
				},
				"box_office": map[string]interface{}{
					"enabled": cfg.Jobs.BoxOffice.Enabled,
					"limit":   cfg.Jobs.BoxOffice.Limit,
				},
				"favorited_movies": map[string]interface{}{
					"enabled": cfg.Jobs.FavoritedMovies.Enabled,
					"limit":   cfg.Jobs.FavoritedMovies.Limit,
					"period":  cfg.Jobs.FavoritedMovies.Period,
				},
				"played_movies": map[string]interface{}{
					"enabled": cfg.Jobs.PlayedMovies.Enabled,
					"limit":   cfg.Jobs.PlayedMovies.Limit,
					"period":  cfg.Jobs.PlayedMovies.Period,
				},
				"watched_movies": map[string]interface{}{
					"enabled": cfg.Jobs.WatchedMovies.Enabled,
					"limit":   cfg.Jobs.WatchedMovies.Limit,
					"period":  cfg.Jobs.WatchedMovies.Period,
				},
				"collected_movies": map[string]interface{}{
					"enabled": cfg.Jobs.CollectedMovies.Enabled,
					"limit":   cfg.Jobs.CollectedMovies.Limit,
					"period":  cfg.Jobs.CollectedMovies.Period,
				},
				"anticipated_movies": map[string]interface{}{
					"enabled": cfg.Jobs.AnticipatedMovies.Enabled,
					"limit":   cfg.Jobs.AnticipatedMovies.Limit,
				},
				"favorited_shows": map[string]interface{}{
					"enabled": cfg.Jobs.FavoritedShows.Enabled,
					"limit":   cfg.Jobs.FavoritedShows.Limit,
					"period":  cfg.Jobs.FavoritedShows.Period,
				},
				"played_shows": map[string]interface{}{
					"enabled": cfg.Jobs.PlayedShows.Enabled,
					"limit":   cfg.Jobs.PlayedShows.Limit,
					"period":  cfg.Jobs.PlayedShows.Period,
				},
				"watched_shows": map[string]interface{}{
					"enabled": cfg.Jobs.WatchedShows.Enabled,
					"limit":   cfg.Jobs.WatchedShows.Limit,
					"period":  cfg.Jobs.WatchedShows.Period,
				},
				"collected_shows": map[string]interface{}{
					"enabled": cfg.Jobs.CollectedShows.Enabled,
					"limit":   cfg.Jobs.CollectedShows.Limit,
					"period":  cfg.Jobs.CollectedShows.Period,
				},
				"anticipated_shows": map[string]interface{}{
					"enabled": cfg.Jobs.AnticipatedShows.Enabled,
					"limit":   cfg.Jobs.AnticipatedShows.Limit,
				},
			},
		}

		return c.JSON(status)
	})

	router.Post("/jobs/trigger/:job", func(c *fiber.Ctx) error {
		jobName := c.Params("job")
		cfg := gctx.Config()
		// Manual triggers should always add content (not dry-run)
		dryRun := false

		switch jobName {
		case "trending-movies":
			if !cfg.Jobs.TrendingMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Trending movies job is disabled",
				})
			}
			go jobs.RunTrendingMovies(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Trending movies job triggered",
			})

		case "trending-shows":
			if !cfg.Jobs.TrendingShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Trending shows job is disabled",
				})
			}
			go jobs.RunTrendingShows(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Trending shows job triggered",
			})

		case "popular-movies":
			if !cfg.Jobs.PopularMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Popular movies job is disabled",
				})
			}
			go jobs.RunPopularMovies(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Popular movies job triggered",
			})

		case "popular-shows":
			if !cfg.Jobs.PopularShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Popular shows job is disabled",
				})
			}
			go jobs.RunPopularShows(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Popular shows job triggered",
			})

		case "box-office":
			if !cfg.Jobs.BoxOffice.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Box office job is disabled",
				})
			}
			go jobs.RunBoxOffice(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Box office job triggered",
			})

		case "favorited-movies":
			if !cfg.Jobs.FavoritedMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Favorited movies job is disabled",
				})
			}
			go jobs.RunFavoritedMovies(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Favorited movies job triggered",
			})

		case "played-movies":
			if !cfg.Jobs.PlayedMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Played movies job is disabled",
				})
			}
			go jobs.RunPlayedMovies(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Played movies job triggered",
			})

		case "watched-movies":
			if !cfg.Jobs.WatchedMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Watched movies job is disabled",
				})
			}
			go jobs.RunWatchedMovies(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Watched movies job triggered",
			})

		case "collected-movies":
			if !cfg.Jobs.CollectedMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Collected movies job is disabled",
				})
			}
			go jobs.RunCollectedMovies(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Collected movies job triggered",
			})

		case "anticipated-movies":
			if !cfg.Jobs.AnticipatedMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Anticipated movies job is disabled",
				})
			}
			go jobs.RunAnticipatedMovies(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Anticipated movies job triggered",
			})

		case "favorited-shows":
			if !cfg.Jobs.FavoritedShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Favorited shows job is disabled",
				})
			}
			go jobs.RunFavoritedShows(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Favorited shows job triggered",
			})

		case "played-shows":
			if !cfg.Jobs.PlayedShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Played shows job is disabled",
				})
			}
			go jobs.RunPlayedShows(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Played shows job triggered",
			})

		case "watched-shows":
			if !cfg.Jobs.WatchedShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Watched shows job is disabled",
				})
			}
			go jobs.RunWatchedShows(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Watched shows job triggered",
			})

		case "collected-shows":
			if !cfg.Jobs.CollectedShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Collected shows job is disabled",
				})
			}
			go jobs.RunCollectedShows(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Collected shows job triggered",
			})

		case "anticipated-shows":
			if !cfg.Jobs.AnticipatedShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Anticipated shows job is disabled",
				})
			}
			go jobs.RunAnticipatedShows(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Anticipated shows job triggered",
			})

		default:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Job not found",
			})
		}
	})

	router.Post("/jobs/preview/:job", func(c *fiber.Ctx) error {
		jobName := c.Params("job")
		cfg := gctx.Config()

		switch jobName {
		case "trending-movies":
			if !cfg.Jobs.TrendingMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Trending movies job is disabled",
				})
			}
			preview := jobs.PreviewTrendingMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "trending-shows":
			if !cfg.Jobs.TrendingShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Trending shows job is disabled",
				})
			}
			preview := jobs.PreviewTrendingShows(cfg, gctx.Database())
			return c.JSON(preview)

		case "popular-movies":
			if !cfg.Jobs.PopularMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Popular movies job is disabled",
				})
			}
			preview := jobs.PreviewPopularMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "popular-shows":
			if !cfg.Jobs.PopularShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Popular shows job is disabled",
				})
			}
			preview := jobs.PreviewPopularShows(cfg, gctx.Database())
			return c.JSON(preview)

		case "box-office":
			if !cfg.Jobs.BoxOffice.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Box office job is disabled",
				})
			}
			preview := jobs.PreviewBoxOffice(cfg, gctx.Database())
			return c.JSON(preview)

		case "favorited-movies":
			if !cfg.Jobs.FavoritedMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Favorited movies job is disabled",
				})
			}
			preview := jobs.PreviewFavoritedMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "played-movies":
			if !cfg.Jobs.PlayedMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Played movies job is disabled",
				})
			}
			preview := jobs.PreviewPlayedMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "watched-movies":
			if !cfg.Jobs.WatchedMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Watched movies job is disabled",
				})
			}
			preview := jobs.PreviewWatchedMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "collected-movies":
			if !cfg.Jobs.CollectedMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Collected movies job is disabled",
				})
			}
			preview := jobs.PreviewCollectedMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "anticipated-movies":
			if !cfg.Jobs.AnticipatedMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Anticipated movies job is disabled",
				})
			}
			preview := jobs.PreviewAnticipatedMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "favorited-shows":
			if !cfg.Jobs.FavoritedShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Favorited shows job is disabled",
				})
			}
			preview := jobs.PreviewFavoritedShows(cfg, gctx.Database())
			return c.JSON(preview)

		case "played-shows":
			if !cfg.Jobs.PlayedShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Played shows job is disabled",
				})
			}
			preview := jobs.PreviewPlayedShows(cfg, gctx.Database())
			return c.JSON(preview)

		case "watched-shows":
			if !cfg.Jobs.WatchedShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Watched shows job is disabled",
				})
			}
			preview := jobs.PreviewWatchedShows(cfg, gctx.Database())
			return c.JSON(preview)

		case "collected-shows":
			if !cfg.Jobs.CollectedShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Collected shows job is disabled",
				})
			}
			preview := jobs.PreviewCollectedShows(cfg, gctx.Database())
			return c.JSON(preview)

		case "anticipated-shows":
			if !cfg.Jobs.AnticipatedShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Anticipated shows job is disabled",
				})
			}
			preview := jobs.PreviewAnticipatedShows(cfg, gctx.Database())
			return c.JSON(preview)

		default:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Job not found",
			})
		}
	})
}
