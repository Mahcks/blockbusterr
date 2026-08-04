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
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

type activityLogGroup struct {
	Log     database.ActivityLog
	Count   int
	History []database.ActivityLog
}

func activityLogGroupKey(log database.ActivityLog) string {
	idPart := ""
	switch log.MediaType {
	case string(enums.MediaTypeMovie):
		if log.TMDBID > 0 {
			idPart = fmt.Sprintf("tmdb:%d", log.TMDBID)
		}
	case string(enums.MediaTypeShow):
		if log.TVDBID > 0 {
			idPart = fmt.Sprintf("tvdb:%d", log.TVDBID)
		} else if log.TMDBID > 0 {
			idPart = fmt.Sprintf("tmdb:%d", log.TMDBID)
		}
	}
	if idPart == "" {
		idPart = strings.ToLower(log.Title) + ":" + strconv.Itoa(log.Year)
	}

	parts := []string{log.MediaType, idPart, log.Status, log.JobType}
	if log.RunID > 0 {
		parts = append(parts, fmt.Sprintf("run:%d", log.RunID))
	}
	return strings.Join(parts, "|")
}

func groupActivityLogs(logs []database.ActivityLog) []activityLogGroup {
	if len(logs) == 0 {
		return []activityLogGroup{}
	}

	groups := make([]activityLogGroup, 0, len(logs))
	indexByKey := make(map[string]int)

	for _, log := range logs {
		key := activityLogGroupKey(log)
		if idx, ok := indexByKey[key]; ok {
			groups[idx].Count++
			groups[idx].History = append(groups[idx].History, log)
			continue
		}

		indexByKey[key] = len(groups)
		groups = append(groups, activityLogGroup{
			Log:     log,
			Count:   1,
			History: []database.ActivityLog{log},
		})
	}

	return groups
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
			maxFetchLimit   = 50000
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
		_ = c.Query("sort")  // Reserved for future use
		_ = c.Query("order") // Reserved for future use

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

		// Fetch more logs than needed to apply filters and calculate total
		fetchLimit := min(pageSize*100, maxFetchLimit) // Fetch enough for filtering
		logs, err := db.GetRecentActivityFiltered(fetchLimit, status, mediaType, jobType, language)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve activity logs",
			})
		}
		sources := make(map[string]string)
		for _, job := range gctx.Config().GetAllJobs() {
			sources[job.ID] = job.Source
		}
		for index := range logs {
			logs[index].Source = sources[logs[index].JobID]
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
			var cutoffStart *time.Time
			var cutoffEnd *time.Time

			switch dateRange {
			case "today":
				start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				cutoffStart = &start
			case "yesterday":
				yesterday := now.AddDate(0, 0, -1)
				start := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, now.Location())
				end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				cutoffStart = &start
				cutoffEnd = &end
			case "week":
				start := now.AddDate(0, 0, -7)
				cutoffStart = &start
			case "month":
				start := now.AddDate(0, 0, -30)
				cutoffStart = &start
			}

			for _, log := range logs {
				if cutoffStart != nil && log.Timestamp.Before(*cutoffStart) {
					continue
				}
				if cutoffEnd != nil && !log.Timestamp.Before(*cutoffEnd) {
					continue
				}
				filtered = append(filtered, log)
			}
			logs = filtered
		}

		// Optional filter by run ID
		if runIDStr != "" {
			runID, err := strconv.ParseInt(runIDStr, 10, 64)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("invalid run_id: %s", runIDStr),
				})
			}
			filtered := []database.ActivityLog{}
			for _, log := range logs {
				if log.RunID == runID {
					filtered = append(filtered, log)
				}
			}
			logs = filtered
		}

		// Optional grouping (dedupe) for cleaner activity views
		dedupe := true
		if dedupeStr := strings.ToLower(strings.TrimSpace(c.Query("dedupe"))); dedupeStr != "" {
			dedupe = dedupeStr != "false" && dedupeStr != "0" && dedupeStr != "no"
		}

		var groupedLogs []activityLogGroup
		if dedupe {
			groupedLogs = groupActivityLogs(logs)
		} else {
			groupedLogs = make([]activityLogGroup, 0, len(logs))
			for _, log := range logs {
				groupedLogs = append(groupedLogs, activityLogGroup{
					Log:     log,
					Count:   1,
					History: []database.ActivityLog{log},
				})
			}
		}

		// Calculate pagination (on grouped results)
		totalRecords := len(groupedLogs)
		totalPages := (totalRecords + pageSize - 1) / pageSize
		if totalPages == 0 {
			totalPages = 1
		}
		if page > totalPages {
			page = totalPages
		}

		// Apply pagination
		startIdx := (page - 1) * pageSize
		endIdx := startIdx + pageSize
		if startIdx >= totalRecords {
			startIdx = 0
		}
		if endIdx > totalRecords {
			endIdx = totalRecords
		}

		paginatedLogs := groupedLogs
		if totalRecords > 0 {
			paginatedLogs = groupedLogs[startIdx:endIdx]
		}

		// Check if this is an HTMX request (wants HTML)
		if c.Get("HX-Request") == "true" {
			// For HTMX, return the table with pagination controls
			return c.Render("activity_table", fiber.Map{
				"Logs":         paginatedLogs,
				"Page":         page,
				"PageSize":     pageSize,
				"TotalRecords": totalRecords,
				"TotalPages":   totalPages,
			})
		}

		// Otherwise return JSON (for API clients)
		return c.JSON(fiber.Map{
			"logs":          paginatedLogs,
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
				"labels":   []string{},
				"added":    []int{},
				"rejected": []int{},
				"skipped":  []int{},
			})
		}

		chartData := make(map[string]any)
		labels := make([]string, 0, 7)
		added := make([]int, 0, 7)
		rejected := make([]int, 0, 7)
		skipped := make([]int, 0, 7)

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
				rejected = append(rejected, counts.Rejected)
				skipped = append(skipped, counts.Skipped)
				continue
			}
			added = append(added, 0)
			rejected = append(rejected, 0)
			skipped = append(skipped, 0)
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

		// Get all rejected count for accurate total
		stats, err := db.GetActivityStats()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve rejection data",
			})
		}
		totalRejected, _ := stats["total_rejected"].(int)

		// Get recent rejected items to analyze filter reasons (sample up to 1000)
		logs, err := db.GetRecentActivityFiltered(1000, string(enums.ActivityStatusRejected), "", "", "")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve rejection data",
			})
		}

		// Count rejection reasons from messages
		reasonCounts := make(map[string]int)
		for _, log := range logs {
			if log.Message != "" {
				reason := strings.ToLower(log.Message)
				// Categorize common reasons
				if strings.Contains(reason, "certification") {
					reasonCounts["Content Rating"]++
				} else if strings.Contains(reason, "rating") {
					reasonCounts["Low Rating"]++
				} else if strings.Contains(reason, "country") {
					reasonCounts["Wrong Country"]++
				} else if strings.Contains(reason, "language") {
					reasonCounts["Wrong Language"]++
				} else if strings.Contains(reason, "genre") {
					reasonCounts["Blacklisted Genre"]++
				} else if strings.Contains(reason, "keyword") {
					reasonCounts["Blacklisted Keyword"]++
				} else if strings.Contains(reason, "runtime") {
					reasonCounts["Runtime Out of Range"]++
				} else if strings.Contains(reason, "year") {
					reasonCounts["Year Out of Range"]++
				} else if strings.Contains(reason, "votes") {
					reasonCounts["Insufficient Votes"]++
				} else if strings.Contains(reason, "network") {
					reasonCounts["Blacklisted Network"]++
				} else {
					reasonCounts["Other"]++
				}
			}
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
			count, err := db.ClearActivityHistory()
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
		cfg := gctx.Config()
		switch activityLog.MediaType {
		case string(enums.MediaTypeMovie):
			if activityLog.TMDBID == 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "No TMDB ID available for this movie",
				})
			}

			// Check if already blocked
			alreadyBlocked := slices.Contains(cfg.TitleExceptions.BlockedMovieTMDBIDs, activityLog.TMDBID)

			if !alreadyBlocked {
				cfg.TitleExceptions.BlockedMovieTMDBIDs = append(cfg.TitleExceptions.BlockedMovieTMDBIDs, activityLog.TMDBID)
				if err := cfg.Save(); err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": fmt.Sprintf("Failed to save config: %v", err),
					})
				}
			}

			log.Infof("Blocked movie '%s' (TMDB ID: %d) - added to blocklist", activityLog.Title, activityLog.TMDBID)

		case string(enums.MediaTypeShow):
			if activityLog.TVDBID == 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "No TVDB ID available for this show",
				})
			}

			// Check if already blocked
			alreadyBlocked := slices.Contains(cfg.TitleExceptions.BlockedShowTVDBIDs, activityLog.TVDBID)

			if !alreadyBlocked {
				cfg.TitleExceptions.BlockedShowTVDBIDs = append(cfg.TitleExceptions.BlockedShowTVDBIDs, activityLog.TVDBID)
				if err := cfg.Save(); err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": fmt.Sprintf("Failed to save config: %v", err),
					})
				}
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
