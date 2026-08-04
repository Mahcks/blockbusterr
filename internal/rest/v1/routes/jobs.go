package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/services/jobs"
)

func AddJobsRoutes(router fiber.Router, gctx global.Context) {
	router.Get("/jobs/status", func(c *fiber.Ctx) error {
		cfg := gctx.Config()

		status := map[string]any{
			"sync_interval": cfg.Jobs.SyncInterval,
			"jobs": map[string]any{
				"trending_movies": map[string]any{
					"enabled": cfg.Jobs.TrendingMovies.Enabled,
					"limit":   cfg.Jobs.TrendingMovies.Limit,
				},
				"trending_shows": map[string]any{
					"enabled": cfg.Jobs.TrendingShows.Enabled,
					"limit":   cfg.Jobs.TrendingShows.Limit,
				},
				"popular_movies": map[string]any{
					"enabled": cfg.Jobs.PopularMovies.Enabled,
					"limit":   cfg.Jobs.PopularMovies.Limit,
				},
				"popular_shows": map[string]any{
					"enabled": cfg.Jobs.PopularShows.Enabled,
					"limit":   cfg.Jobs.PopularShows.Limit,
				},
				"box_office": map[string]any{
					"enabled": cfg.Jobs.BoxOffice.Enabled,
					"limit":   cfg.Jobs.BoxOffice.Limit,
				},
				"favorited_movies": map[string]any{
					"enabled": cfg.Jobs.FavoritedMovies.Enabled,
					"limit":   cfg.Jobs.FavoritedMovies.Limit,
					"period":  cfg.Jobs.FavoritedMovies.Period,
				},
				"played_movies": map[string]any{
					"enabled": cfg.Jobs.PlayedMovies.Enabled,
					"limit":   cfg.Jobs.PlayedMovies.Limit,
					"period":  cfg.Jobs.PlayedMovies.Period,
				},
				"watched_movies": map[string]any{
					"enabled": cfg.Jobs.WatchedMovies.Enabled,
					"limit":   cfg.Jobs.WatchedMovies.Limit,
					"period":  cfg.Jobs.WatchedMovies.Period,
				},
				"collected_movies": map[string]any{
					"enabled": cfg.Jobs.CollectedMovies.Enabled,
					"limit":   cfg.Jobs.CollectedMovies.Limit,
					"period":  cfg.Jobs.CollectedMovies.Period,
				},
				"anticipated_movies": map[string]any{
					"enabled": cfg.Jobs.AnticipatedMovies.Enabled,
					"limit":   cfg.Jobs.AnticipatedMovies.Limit,
				},
				"favorited_shows": map[string]any{
					"enabled": cfg.Jobs.FavoritedShows.Enabled,
					"limit":   cfg.Jobs.FavoritedShows.Limit,
					"period":  cfg.Jobs.FavoritedShows.Period,
				},
				"played_shows": map[string]any{
					"enabled": cfg.Jobs.PlayedShows.Enabled,
					"limit":   cfg.Jobs.PlayedShows.Limit,
					"period":  cfg.Jobs.PlayedShows.Period,
				},
				"watched_shows": map[string]any{
					"enabled": cfg.Jobs.WatchedShows.Enabled,
					"limit":   cfg.Jobs.WatchedShows.Limit,
					"period":  cfg.Jobs.WatchedShows.Period,
				},
				"collected_shows": map[string]any{
					"enabled": cfg.Jobs.CollectedShows.Enabled,
					"limit":   cfg.Jobs.CollectedShows.Limit,
					"period":  cfg.Jobs.CollectedShows.Period,
				},
				"anticipated_shows": map[string]any{
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
		// Development triggers must never mutate a real media stack.
		dryRun := gctx.Metadata().Version == "dev"

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

		case "smart-popular-movies":
			if !cfg.Jobs.SmartPopularMovies.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Smart popular movies job is disabled",
				})
			}
			go jobs.RunSmartPopularMovies(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Smart popular movies job triggered",
			})

		case "smart-popular-shows":
			if !cfg.Jobs.SmartPopularShows.Enabled {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Smart popular shows job is disabled",
				})
			}
			go jobs.RunSmartPopularShows(cfg, gctx.Database(), dryRun)
			return c.JSON(fiber.Map{
				"message": "Smart popular shows job triggered",
			})

		default:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Job not found",
			})
		}
	})

	// Get decision details for a specific job's last run
	router.Get("/jobs/:job/decisions", func(c *fiber.Ctx) error {
		jobName := c.Params("job")
		db := gctx.Database()

		// Convert hyphenated job name to underscore format for database
		dbJobName := jobs.HyphenToUnderscore(jobName)

		// Get recent logs for this job to build decision summary
		logs, err := db.GetRecentActivityFiltered(100, "", "", dbJobName, "")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve job decisions",
			})
		}

		if len(logs) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "No decisions found for this job",
			})
		}

		// Group by run (use timestamp clustering - items within 5 minutes are same run)
		// For simplicity, just return the most recent run (last 20 items)
		recentLogs := logs
		if len(logs) > 20 {
			recentLogs = logs[:20]
		}

		// Build decision summary
		summary := map[string]int{
			"total":     len(recentLogs),
			"added":     0,
			"requested": 0,
			"skipped":   0,
			"rejected":  0,
			"failed":    0,
		}

		decisions := make([]map[string]any, 0, len(recentLogs))
		for _, log := range recentLogs {
			// Update summary stats
			summary[log.Status]++

			decision := map[string]any{
				"title":      log.Title,
				"year":       log.Year,
				"media_type": log.MediaType,
				"tmdb_id":    log.TMDBID,
				"tvdb_id":    log.TVDBID,
				"imdb_id":    log.IMDBID,
				"poster_url": log.PosterURL,
				"score":      log.Score,
				"rank":       log.Rank,
				"status":     log.Status,
				"message":    log.Message,
				"timestamp":  log.Timestamp,
			}

			// Parse filter details if available
			if log.FilterDetails != "" {
				filterChecks := jobs.FilterChecksFromJSON(log.FilterDetails)
				decision["filter_checks"] = filterChecks
			}

			decisions = append(decisions, decision)
		}

		return c.JSON(fiber.Map{
			"job":       jobName,
			"run_time":  recentLogs[0].Timestamp,
			"summary":   summary,
			"decisions": decisions,
		})
	})

	router.Post("/jobs/preview/:job", func(c *fiber.Ctx) error {
		jobName := c.Params("job")
		cfg := gctx.Config()

		switch jobName {
		case "trending-movies":
			preview := jobs.PreviewTrendingMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "trending-shows":
			preview := jobs.PreviewTrendingShows(cfg, gctx.Database())
			return c.JSON(preview)

		case "popular-movies":
			preview := jobs.PreviewPopularMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "popular-shows":
			preview := jobs.PreviewPopularShows(cfg, gctx.Database())
			return c.JSON(preview)

		case "box-office":
			preview := jobs.PreviewBoxOffice(cfg, gctx.Database())
			return c.JSON(preview)

		case "favorited-movies":
			preview := jobs.PreviewFavoritedMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "played-movies":
			preview := jobs.PreviewPlayedMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "watched-movies":
			preview := jobs.PreviewWatchedMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "collected-movies":
			preview := jobs.PreviewCollectedMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "anticipated-movies":
			preview := jobs.PreviewAnticipatedMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "favorited-shows":
			preview := jobs.PreviewFavoritedShows(cfg, gctx.Database())
			return c.JSON(preview)

		case "played-shows":
			preview := jobs.PreviewPlayedShows(cfg, gctx.Database())
			return c.JSON(preview)

		case "watched-shows":
			preview := jobs.PreviewWatchedShows(cfg, gctx.Database())
			return c.JSON(preview)

		case "collected-shows":
			preview := jobs.PreviewCollectedShows(cfg, gctx.Database())
			return c.JSON(preview)

		case "anticipated-shows":
			preview := jobs.PreviewAnticipatedShows(cfg, gctx.Database())
			return c.JSON(preview)

		case "smart-popular-movies":
			preview := jobs.PreviewSmartPopularMovies(cfg, gctx.Database())
			return c.JSON(preview)

		case "smart-popular-shows":
			preview := jobs.PreviewSmartPopularShows(cfg, gctx.Database())
			return c.JSON(preview)

		default:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Job not found",
			})
		}
	})
}
