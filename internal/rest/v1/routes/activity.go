package routes

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/internal/services/jobs"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

type activityLogGroup = database.ActivityLogGroup

func activityDateBounds(dateRange string, now time.Time) (*time.Time, *time.Time) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	switch dateRange {
	case "today":
		return &today, nil
	case "yesterday":
		start := today.AddDate(0, 0, -1)
		return &start, &today
	case "week":
		start := now.AddDate(0, 0, -7)
		return &start, nil
	case "month":
		start := now.AddDate(0, 0, -30)
		return &start, nil
	default:
		return nil, nil
	}
}

func RegisterActivityRoutes(router fiber.Router, gctx global.Context) {
	// Debug endpoint for runtime activity/config/database diagnosis
	router.Get("/activity/debug", func(c *fiber.Ctx) error {
		db := gctx.Database()
		cfg := gctx.Config()

		response := fiber.Map{
			"db_nil":       db == nil,
			"config_path":  cfg.ConfigFilePath,
			"app_version":  gctx.Metadata().Version,
			"app_commit":   gctx.Metadata().Commit,
			"runtime_time": time.Now().Format(time.RFC3339),
		}

		if db != nil {
			debugInfo, err := db.GetActivityDebug()
			if err != nil {
				response["debug_error"] = err.Error()
			} else {
				response["db"] = debugInfo
			}
		}

		return c.JSON(response)
	})

	// Get recent activity logs
	router.Get("/activity/logs", func(c *fiber.Ctx) error {
		const (
			defaultPageSize = 50
			maxPageSize     = 200
			maxLegacyLimit  = 10000
		)

		db := gctx.Database()
		if db == nil {
			if c.Get("HX-Request") == "true" {
				return c.Status(fiber.StatusServiceUnavailable).SendString("Database unavailable")
			}
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Database unavailable",
			})
		}

		// Get pagination params
		page := 1
		if pageStr := c.Query("page"); pageStr != "" {
			if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
				page = parsedPage
			}
		}

		pageSize := defaultPageSize
		if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
			if parsedSize, err := strconv.Atoi(pageSizeStr); err == nil && parsedSize > 0 && parsedSize <= maxPageSize {
				pageSize = parsedSize
			}
		}

		// Legacy support for limit param
		if limitStr := c.Query("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				pageSize = min(parsedLimit, maxLegacyLimit)
			}
		}

		// Get filter params
		status := c.Query("status")   // "added", "failed", "rejected", "requested", or empty for all
		mediaType := c.Query("media") // "movie", "show", or empty for all
		jobType := c.Query("job")     // job type filter or empty for all
		language := strings.TrimSpace(c.Query("language"))
		search := c.Query("search")        // search by title
		dateRange := c.Query("date_range") // "today", "yesterday", "week", "month"
		runIDStr := c.Query("run_id")
		sortField := c.Query("sort")
		order := c.Query("order")

		if status != "" {
			parsedStatus, ok := enums.ParseActivityStatus(status)
			if !ok {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("invalid status: %s", status),
				})
			}
			status = string(parsedStatus)
		}
		if mediaType != "" {
			parsedMediaType, ok := enums.ParseMediaType(mediaType)
			if !ok {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("invalid media type: %s", mediaType),
				})
			}
			mediaType = string(parsedMediaType)
		}

		var runID *int64
		if runIDStr != "" {
			parsedRunID, err := strconv.ParseInt(runIDStr, 10, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("invalid run_id: %s", runIDStr),
				})
			}
			runID = &parsedRunID
		}

		// Optional grouping (dedupe) for cleaner activity views
		dedupe := true
		if dedupeStr := strings.ToLower(strings.TrimSpace(c.Query("dedupe"))); dedupeStr != "" {
			dedupe = dedupeStr != "false" && dedupeStr != "0" && dedupeStr != "no"
		}

		start, end := activityDateBounds(dateRange, time.Now())
		result, err := db.QueryActivity(database.ActivityQuery{
			Status: status, MediaType: mediaType, Job: jobType, Language: language,
			Search: search, Start: start, End: end, RunID: runID, Dedupe: dedupe,
			Sort: sortField, Order: order, Limit: pageSize, Offset: (page - 1) * pageSize,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve activity logs"})
		}
		sources := make(map[string]string)
		for _, job := range gctx.Config().GetAllJobs() {
			sources[job.ID] = job.Source
		}
		for index := range result.Groups {
			result.Groups[index].Log.Source = sources[result.Groups[index].Log.JobID]
			for historyIndex := range result.Groups[index].History {
				result.Groups[index].History[historyIndex].Source = sources[result.Groups[index].History[historyIndex].JobID]
			}
		}

		totalRecords := result.Total
		totalPages := (totalRecords + pageSize - 1) / pageSize
		if totalPages == 0 {
			totalPages = 1
		}
		page = result.Offset/pageSize + 1

		// Check if this is an HTMX request (wants HTML)
		if c.Get("HX-Request") == "true" {
			// For HTMX, return the table with pagination controls
			return c.Render("activity_table", fiber.Map{
				"Logs":         result.Groups,
				"Page":         page,
				"PageSize":     pageSize,
				"TotalRecords": totalRecords,
				"TotalPages":   totalPages,
			})
		}

		// Otherwise return JSON (for API clients)
		return c.JSON(fiber.Map{
			"logs":          result.Groups,
			"grouped":       dedupe,
			"page":          page,
			"page_size":     pageSize,
			"total_records": totalRecords,
			"total_pages":   totalPages,
		})
	})

	router.Get("/activity/languages", func(c *fiber.Ctx) error {
		languages, err := gctx.Database().GetActivityLanguages()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve activity languages"})
		}
		return c.JSON(fiber.Map{"languages": languages})
	})

	// Get recent job runs timeline
	router.Get("/activity/runs", func(c *fiber.Ctx) error {
		db := gctx.Database()
		if db == nil {
			return c.JSON(fiber.Map{
				"runs": []database.JobRun{},
			})
		}
		limit := 100
		if limitStr := c.Query("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 500 {
				limit = parsedLimit
			}
		}

		jobID := strings.TrimSpace(c.Query("job_id"))
		runs, err := db.GetRecentJobRuns(limit, jobID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve job runs",
			})
		}
		cycles, err := db.GetRecentSelectionCycles(20)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve selection cycles"})
		}

		return c.JSON(fiber.Map{
			"runs":   runs,
			"cycles": cycles,
		})
	})

	// Get activity chart data
	router.Get("/activity/chart", func(c *fiber.Ctx) error {
		db := gctx.Database()
		if db == nil {
			return c.JSON(fiber.Map{
				"labels": []string{}, "added": []int{}, "requested": []int{}, "would_add": []int{}, "would_request": []int{},
				"rejected": []int{}, "skipped": []int{}, "failed": []int{}, "dry_run": jobs.DryRunEnabled(gctx.Metadata().Version),
			})
		}

		chartData := make(map[string]any)
		labels := make([]string, 0, 7)
		added := make([]int, 0, 7)
		requested := make([]int, 0, 7)
		wouldAdd := make([]int, 0, 7)
		wouldRequest := make([]int, 0, 7)
		rejected := make([]int, 0, 7)
		skipped := make([]int, 0, 7)
		failed := make([]int, 0, 7)

		countsByDay, err := db.GetActivityDailyCounts(7)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve chart data",
			})
		}

		now := time.Now()
		for i := 6; i >= 0; i-- {
			day := now.AddDate(0, 0, -i)
			labels = append(labels, day.Format("1/2"))

			key := day.Format("2006-01-02")
			if counts, ok := countsByDay[key]; ok {
				added = append(added, counts.Added)
				requested = append(requested, counts.Requested)
				wouldAdd = append(wouldAdd, counts.WouldAdd)
				wouldRequest = append(wouldRequest, counts.WouldRequest)
				rejected = append(rejected, counts.Rejected)
				skipped = append(skipped, counts.Skipped)
				failed = append(failed, counts.Failed)
				continue
			}
			added = append(added, 0)
			requested = append(requested, 0)
			wouldAdd = append(wouldAdd, 0)
			wouldRequest = append(wouldRequest, 0)
			rejected = append(rejected, 0)
			skipped = append(skipped, 0)
			failed = append(failed, 0)
		}

		chartData["labels"] = labels
		chartData["added"] = added
		chartData["requested"] = requested
		chartData["would_add"] = wouldAdd
		chartData["would_request"] = wouldRequest
		chartData["rejected"] = rejected
		chartData["skipped"] = skipped
		chartData["failed"] = failed
		chartData["dry_run"] = jobs.DryRunEnabled(gctx.Metadata().Version)

		return c.JSON(chartData)
	})

	// Get activity statistics
	router.Get("/activity/stats", func(c *fiber.Ctx) error {
		db := gctx.Database()
		if db == nil {
			return c.JSON(fiber.Map{
				"total_added":    0,
				"total_failed":   0,
				"total_rejected": 0,
				"total_skipped":  0,
				"total_movies":   0,
				"total_shows":    0,
				"added_last_24h": 0,
			})
		}

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
		if db == nil {
			return c.JSON(fiber.Map{
				"total_rejected": 0,
				"breakdown":      map[string]int{},
			})
		}

		totalRejected, reasonCounts, err := db.GetRejectionBreakdown()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve rejection data",
			})
		}

		return c.JSON(fiber.Map{
			"total_rejected": totalRejected,
			"breakdown":      reasonCounts,
		})
	})

	// Clear old logs (admin endpoint)
	router.Delete("/activity/logs", func(c *fiber.Ctx) error {
		db := gctx.Database()
		if db == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Database not initialized",
			})
		}

		if c.Query("scope") == "all" {
			if c.Query("confirm") != "CLEAR" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Type CLEAR to confirm deletion"})
			}
			count, err := db.ClearActivityHistory(c.QueryBool("clear_delivery_memory", false))
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to clear activity history"})
			}
			return c.JSON(fiber.Map{"message": "Activity history cleared successfully", "count": count})
		}

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
		if db == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Database not initialized",
			})
		}
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
		err = db.UpdateActivityLogStatus(id, string(enums.ActivityStatusAdded), "Manually added by user")
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

	// Block media (add to blocklist)
	router.Post("/activity/:id/block", func(c *fiber.Ctx) error {
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
		if db == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Database not initialized",
			})
		}
		activityLog, err := db.GetActivityLogByID(id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve activity log",
			})
		}

		if activityLog == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Activity log not found",
			})
		}

		// Add to blocklist in config
		switch activityLog.MediaType {
		case string(enums.MediaTypeMovie):
			if activityLog.TMDBID == 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "No TMDB ID available for this movie",
				})
			}

			if err := global.UpdateConfig(gctx, func(candidate *config.Config) error {
				if !slices.Contains(candidate.TitleExceptions.BlockedMovieTMDBIDs, activityLog.TMDBID) {
					candidate.TitleExceptions.BlockedMovieTMDBIDs = append(candidate.TitleExceptions.BlockedMovieTMDBIDs, activityLog.TMDBID)
				}
				return nil
			}); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to save config: %v", err)})
			}

			log.Infof("Blocked movie '%s' (TMDB ID: %d) - added to blocklist", activityLog.Title, activityLog.TMDBID)

		case string(enums.MediaTypeShow):
			if activityLog.TVDBID == 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "No TVDB ID available for this show",
				})
			}

			if err := global.UpdateConfig(gctx, func(candidate *config.Config) error {
				if !slices.Contains(candidate.TitleExceptions.BlockedShowTVDBIDs, activityLog.TVDBID) {
					candidate.TitleExceptions.BlockedShowTVDBIDs = append(candidate.TitleExceptions.BlockedShowTVDBIDs, activityLog.TVDBID)
				}
				return nil
			}); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to save config: %v", err)})
			}

			log.Infof("Blocked show '%s' (TVDB ID: %d) - added to blocklist", activityLog.Title, activityLog.TVDBID)
		}

		// Update the activity log
		err = db.UpdateActivityLogStatus(id, string(enums.ActivityStatusBlocked), "Blocked by user")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update activity log",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": fmt.Sprintf("%s has been blocked and will not be added in future runs", activityLog.Title),
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
		case string(enums.MediaTypeMovie):
			if activityLog.TMDBID == 0 {
				return fmt.Errorf("no TMDB ID available for movie")
			}
			_, err := jellyseerrClient.RequestMovie(activityLog.TMDBID)
			if err != nil {
				return fmt.Errorf("failed to request movie via Jellyseerr: %w", err)
			}
			log.Infof("Manually requested movie '%s' via Jellyseerr (TMDB ID: %d)", activityLog.Title, activityLog.TMDBID)
		case string(enums.MediaTypeShow):
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
		case string(enums.MediaTypeMovie):
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
		case string(enums.MediaTypeShow):
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
