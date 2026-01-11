package routes

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/global"
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
}
