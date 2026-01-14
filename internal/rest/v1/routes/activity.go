package routes

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

func RegisterActivityRoutes(router fiber.Router, gctx global.Context) {
	// Get recent activity logs
	router.Get("/activity/logs", func(c *fiber.Ctx) error {
		db := gctx.Database()

		// Get limit from query params (default 50)
		limit := 50
		if limitStr := c.Query("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		// Get filter params
		status := c.Query("status")        // "added", "failed", "rejected", "requested", or empty for all
		mediaType := c.Query("media")      // "movie", "show", or empty for all
		jobType := c.Query("job")          // job type filter or empty for all
		search := c.Query("search")        // search by title
		dateRange := c.Query("date_range") // "today", "yesterday", "week", "month"
		_ = c.Query("sort")                // Reserved for future use
		_ = c.Query("order")               // Reserved for future use

		logs, err := db.GetRecentActivityFiltered(limit, status, mediaType, jobType)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve activity logs",
			})
		}

		// Apply search filter
		if search != "" {
			filtered := []database.ActivityLog{}
			searchLower := strings.ToLower(search)
			for _, log := range logs {
				if strings.Contains(strings.ToLower(log.Title), searchLower) {
					filtered = append(filtered, log)
				}
			}
			logs = filtered
		}

		// Apply date range filter
		if dateRange != "" {
			filtered := []database.ActivityLog{}
			now := time.Now()
			var cutoff time.Time

			switch dateRange {
			case "today":
				cutoff = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			case "yesterday":
				yesterday := now.AddDate(0, 0, -1)
				cutoff = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, now.Location())
			case "week":
				cutoff = now.AddDate(0, 0, -7)
			case "month":
				cutoff = now.AddDate(0, 0, -30)
			}

			for _, log := range logs {
				if log.Timestamp.After(cutoff) {
					filtered = append(filtered, log)
				}
			}
			logs = filtered
		}

		// Check if this is an HTMX request (wants HTML)
		if c.Get("HX-Request") == "true" {
			return c.Render("activity_table", logs)
		}

		// Otherwise return JSON (for API clients)
		return c.JSON(logs)
	})

	// Get activity chart data
	router.Get("/activity/chart", func(c *fiber.Ctx) error {
		db := gctx.Database()

		// Get last 7 days of activity
		chartData := make(map[string]interface{})
		labels := []string{}
		added := []int{}
		rejected := []int{}
		skipped := []int{}

		now := time.Now()
		for i := 6; i >= 0; i-- {
			date := now.AddDate(0, 0, -i)
			dateStr := date.Format("1/2")
			labels = append(labels, dateStr)

			// Get logs for this day
			dayLogs, err := db.GetRecentActivityFiltered(10000, "", "", "")
			if err != nil {
				continue
			}

			dayAdded := 0
			dayRejected := 0
			daySkipped := 0

			dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
			dayEnd := dayStart.AddDate(0, 0, 1)

			for _, log := range dayLogs {
				if log.Timestamp.After(dayStart) && log.Timestamp.Before(dayEnd) {
					switch log.Status {
					case "added", "requested":
						dayAdded++
					case "rejected":
						dayRejected++
					case "skipped":
						daySkipped++
					}
				}
			}

			added = append(added, dayAdded)
			rejected = append(rejected, dayRejected)
			skipped = append(skipped, daySkipped)
		}

		chartData["labels"] = labels
		chartData["added"] = added
		chartData["rejected"] = rejected
		chartData["skipped"] = skipped

		return c.JSON(chartData)
	})

	// Get activity statistics
	router.Get("/activity/stats", func(c *fiber.Ctx) error {
		db := gctx.Database()

		stats, err := db.GetActivityStats()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve activity statistics",
			})
		}

		return c.JSON(stats)
	})

	// Get rejection reasons breakdown
	router.Get("/activity/rejection-breakdown", func(c *fiber.Ctx) error {
		db := gctx.Database()

		// Get recent rejected items to analyze filter reasons
		logs, err := db.GetRecentActivityFiltered(1000, "rejected", "", "")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve rejection data",
			})
		}

		// Count rejection reasons from messages
		reasonCounts := make(map[string]int)
		for _, log := range logs {
			if log.Message != "" {
				// Parse the reason from message
				reason := log.Message
				// Categorize common reasons
				if contains := func(s, substr string) bool {
					return len(s) >= len(substr) && (s[:len(substr)] == substr || len(s) > len(substr) && s[len(s)-len(substr):] == substr || func() bool {
						for i := 0; i <= len(s)-len(substr); i++ {
							if s[i:i+len(substr)] == substr {
								return true
							}
						}
						return false
					}())
				}; contains(reason, "rating") {
					reasonCounts["Low Rating"]++
				} else if contains(reason, "country") {
					reasonCounts["Wrong Country"]++
				} else if contains(reason, "language") {
					reasonCounts["Wrong Language"]++
				} else if contains(reason, "genre") {
					reasonCounts["Blacklisted Genre"]++
				} else if contains(reason, "keyword") {
					reasonCounts["Blacklisted Keyword"]++
				} else if contains(reason, "runtime") {
					reasonCounts["Runtime Out of Range"]++
				} else if contains(reason, "year") {
					reasonCounts["Year Out of Range"]++
				} else if contains(reason, "votes") {
					reasonCounts["Insufficient Votes"]++
				} else if contains(reason, "network") {
					reasonCounts["Blacklisted Network"]++
				} else {
					reasonCounts["Other"]++
				}
			}
		}

		return c.JSON(fiber.Map{
			"total_rejected": len(logs),
			"breakdown":      reasonCounts,
		})
	})

	// Clear old logs (admin endpoint)
	router.Delete("/activity/logs", func(c *fiber.Ctx) error {
		db := gctx.Database()

		// Get days from query params (default 30)
		days := 30
		if daysStr := c.Query("days"); daysStr != "" {
			if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
				days = parsedDays
			}
		}

		count, err := db.ClearOldLogs(days)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to clear old logs",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Old logs cleared successfully",
			"count":   count,
		})
	})

	// Add media anyway (manual override for rejected items)
	router.Post("/activity/:id/add-anyway", func(c *fiber.Ctx) error {
		// Get the activity log ID from the URL
		idStr := c.Params("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid activity log ID",
			})
		}

		// Retrieve the activity log entry
		db := gctx.Database()
		log, err := db.GetActivityLogByID(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve activity log",
			})
		}

		if log == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Activity log not found",
			})
		}

		// Call the manual add logic
		err = addMediaManually(c.Context(), gctx, log)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to add media: %v", err),
			})
		}

		// Update the activity log status
		err = db.UpdateActivityLogStatus(id, "added", "Manually added by user")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update activity log",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": fmt.Sprintf("%s has been added successfully", log.Title),
		})
	})
}

// addMediaManually handles adding media manually via Jellyseerr or direct *arr integration
func addMediaManually(ctx context.Context, gctx global.Context, activityLog *database.ActivityLog) error {
	cfg := gctx.Config()
	mode := cfg.Jobs.Mode

	// Determine which integration to use based on mode
	if mode == "jellyseerr" || mode == "" {
		// Use Jellyseerr
		jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
			URL:             cfg.Jellyseerr.URL,
			APIKey:          cfg.Jellyseerr.APIKey,
			UserID:          cfg.Jellyseerr.UserID,
			RequestEmail:    cfg.Jellyseerr.RequestCredentials.Email,
			RequestPassword: cfg.Jellyseerr.RequestCredentials.Password,
		})

		switch activityLog.MediaType {
		case "movie":
			if activityLog.TMDBID == 0 {
				return fmt.Errorf("no TMDB ID available for movie")
			}
			_, err := jellyseerrClient.RequestMovie(activityLog.TMDBID)
			if err != nil {
				return fmt.Errorf("failed to request movie via Jellyseerr: %w", err)
			}
			log.Infof("Manually requested movie '%s' via Jellyseerr (TMDB ID: %d)", activityLog.Title, activityLog.TMDBID)
		case "show":
			if activityLog.TMDBID == 0 {
				return fmt.Errorf("no TMDB ID available for show")
			}
			_, err := jellyseerrClient.RequestShow(activityLog.TMDBID)
			if err != nil {
				return fmt.Errorf("failed to request show via Jellyseerr: %w", err)
			}
			log.Infof("Manually requested show '%s' via Jellyseerr (TMDB ID: %d)", activityLog.Title, activityLog.TMDBID)
		}
	} else {
		// Direct mode - add to *arr
		switch activityLog.MediaType {
		case "movie":
			// Add to Radarr
			radarrClient := integrations.NewRadarr(integrations.RadarrConfig{
				BaseURL: cfg.Radarr.URL,
				APIKey:  cfg.Radarr.APIKey,
			})

			if activityLog.TMDBID == 0 {
				return fmt.Errorf("no TMDB ID available for movie")
			}

			// Look up the movie in Radarr to get full details
			movies, err := radarrClient.LookupMovie(ctx, fmt.Sprintf("tmdb:%d", activityLog.TMDBID))
			if err != nil {
				return fmt.Errorf("failed to lookup movie in Radarr: %w", err)
			}

			if len(movies) == 0 {
				return fmt.Errorf("movie not found in Radarr lookup")
			}

			movie := movies[0]
			movie.QualityProfileID = cfg.Radarr.QualityProfile
			movie.RootFolderPath = cfg.Radarr.RootFolder
			movie.Monitored = true
			movie.MinimumAvailability = "released"
			movie.AddOptions = &integrations.RadarrAddOptions{
				SearchForMovie: true,
			}

			_, err = radarrClient.AddMovie(ctx, movie)
			if err != nil {
				return fmt.Errorf("failed to add movie to Radarr: %w", err)
			}
			log.Infof("Manually added movie '%s' to Radarr (TMDB ID: %d)", activityLog.Title, activityLog.TMDBID)
		case "show":
			// Add to Sonarr
			sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
				BaseURL: cfg.Sonarr.URL,
				APIKey:  cfg.Sonarr.APIKey,
			})

			if activityLog.TVDBID == 0 {
				return fmt.Errorf("no TVDB ID available for show")
			}

			// Look up the show in Sonarr to get full details
			series, err := sonarrClient.LookupSeries(ctx, fmt.Sprintf("tvdb:%d", activityLog.TVDBID))
			if err != nil {
				return fmt.Errorf("failed to lookup show in Sonarr: %w", err)
			}

			if len(series) == 0 {
				return fmt.Errorf("show not found in Sonarr lookup")
			}

			show := series[0]
			show.QualityProfileID = cfg.Sonarr.QualityProfile
			show.RootFolderPath = cfg.Sonarr.RootFolder
			show.Monitored = true
			show.SeasonFolder = true
			show.AddOptions = &integrations.SonarrAddOptions{
				SearchForMissingEpisodes: true,
			}

			_, err = sonarrClient.AddSeries(ctx, show)
			if err != nil {
				return fmt.Errorf("failed to add show to Sonarr: %w", err)
			}
			log.Infof("Manually added show '%s' to Sonarr (TVDB ID: %d)", activityLog.Title, activityLog.TVDBID)
		}
	}

	return nil
}
