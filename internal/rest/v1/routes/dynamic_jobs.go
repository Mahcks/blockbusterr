package routes

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/services/jobs"
)

// AddDynamicJobsRoutes adds the dynamic job management API endpoints
func AddDynamicJobsRoutes(router fiber.Router, gctx global.Context) {
	// Get all available job type definitions (for UI dropdowns)
	router.Get("/jobs/types", func(c *fiber.Ctx) error {
		return c.JSON(jobs.GetAllJobTypes())
	})

	// Get all pre-configured job templates
	router.Get("/jobs/templates", func(c *fiber.Ctx) error {
		return c.JSON(jobs.GetAllTemplates())
	})

	// Get all configured jobs (dynamic + legacy)
	router.Get("/jobs/list", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		allJobs := cfg.GetAllJobs()
		return c.JSON(allJobs)
	})

	// Get only enabled jobs
	router.Get("/jobs/enabled", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		enabledJobs := cfg.GetEnabledJobs()
		return c.JSON(enabledJobs)
	})

	// Get a specific job by ID
	router.Get("/jobs/dynamic/:id", func(c *fiber.Ctx) error {
		jobID := c.Params("id")
		cfg := gctx.Config()

		// Search in all jobs (dynamic + legacy)
		allJobs := cfg.GetAllJobs()
		for _, job := range allJobs {
			if job.ID == jobID {
				return c.JSON(job)
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Job not found",
		})
	})

	// Create a new dynamic job
	router.Post("/jobs", func(c *fiber.Ctx) error {
		var job config.DynamicJob
		if err := c.BodyParser(&job); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		// Generate ID if not provided
		if job.ID == "" {
			job.ID = uuid.NewString()
		}

		// Validate the job
		if err := validateDynamicJob(job); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		cfg := gctx.Config()
		if err := cfg.AddDynamicJob(job); err != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Save config to persist changes
		if err := cfg.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save config: " + err.Error(),
			})
		}

		// Reload config to apply changes
		if err := gctx.ReloadConfig(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to reload config: " + err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(job)
	})

	// Update an existing dynamic job
	router.Put("/jobs/:id", func(c *fiber.Ctx) error {
		jobID := c.Params("id")

		// Prevent updating legacy jobs through this endpoint
		if strings.HasPrefix(jobID, "legacy_") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot update legacy jobs through this endpoint. Use the legacy configuration instead.",
			})
		}

		var job config.DynamicJob
		if err := c.BodyParser(&job); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		// Enforce immutable ID
		if job.ID != "" && job.ID != jobID {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Job ID cannot be changed",
			})
		}
		job.ID = jobID

		// Validate the job
		if err := validateDynamicJob(job); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		cfg := gctx.Config()
		if err := cfg.UpdateDynamicJob(job); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Save config to persist changes
		if err := cfg.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save config: " + err.Error(),
			})
		}

		// Reload config to apply changes
		if err := gctx.ReloadConfig(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to reload config: " + err.Error(),
			})
		}

		return c.JSON(job)
	})

	// Delete a dynamic job
	router.Delete("/jobs/:id", func(c *fiber.Ctx) error {
		jobID := c.Params("id")

		// Prevent deletion of legacy jobs
		if strings.HasPrefix(jobID, "legacy_") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot delete legacy jobs. Disable them in the configuration instead.",
			})
		}

		cfg := gctx.Config()
		if err := cfg.DeleteDynamicJob(jobID); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Save config to persist changes
		if err := cfg.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save config: " + err.Error(),
			})
		}

		// Reload config to apply changes
		if err := gctx.ReloadConfig(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to reload config: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Job deleted successfully",
		})
	})

	// Trigger a dynamic job by ID
	router.Post("/jobs/:id/trigger", func(c *fiber.Ctx) error {
		jobID := c.Params("id")
		cfg := gctx.Config()

		// Find the job
		allJobs := cfg.GetAllJobs()
		var targetJob *config.DynamicJob
		for _, job := range allJobs {
			if job.ID == jobID {
				j := job // Create a copy to avoid loop variable issues
				targetJob = &j
				break
			}
		}

		if targetJob == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Job not found",
			})
		}

		if !targetJob.Enabled {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Job is disabled",
			})
		}

		// Execute job asynchronously
		go func() {
			dryRun := gctx.Metadata().Version == "dev"
			_ = jobs.RunDynamicJob(cfg, gctx.Database(), *targetJob, dryRun)
		}()

		return c.JSON(fiber.Map{
			"message": fmt.Sprintf("Job '%s' triggered successfully", targetJob.Name),
		})
	})

	// Preview a dynamic job by ID
	router.Post("/jobs/:id/preview", func(c *fiber.Ctx) error {
		jobID := c.Params("id")
		cfg := gctx.Config()

		// Find the job
		allJobs := cfg.GetAllJobs()
		var targetJob *config.DynamicJob
		for _, job := range allJobs {
			if job.ID == jobID {
				j := job
				targetJob = &j
				break
			}
		}

		if targetJob == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Job not found",
			})
		}

		// Use existing preview functions based on job type
		preview := previewDynamicJob(cfg, gctx.Database(), *targetJob)
		return c.JSON(preview)
	})

	// Migrate legacy jobs to dynamic list
	router.Post("/jobs/migrate", func(c *fiber.Ctx) error {
		cfg := gctx.Config()

		migratedJobs, err := cfg.MigrateLegacyJobs()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to migrate jobs: " + err.Error(),
			})
		}

		if len(migratedJobs) == 0 {
			return c.JSON(fiber.Map{
				"message": "No legacy jobs to migrate",
				"count":   0,
			})
		}

		// Save config to persist changes
		if err := cfg.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save config: " + err.Error(),
			})
		}

		// Reload config to apply changes
		if err := gctx.ReloadConfig(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to reload config: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": fmt.Sprintf("Successfully migrated %d legacy jobs", len(migratedJobs)),
			"count":   len(migratedJobs),
			"jobs":    migratedJobs,
		})
	})
}

// validateDynamicJob validates a DynamicJob configuration
func validateDynamicJob(job config.DynamicJob) error {
	if job.Type == "" {
		return fmt.Errorf("job type is required")
	}
	if job.MediaType == "" {
		return fmt.Errorf("media type is required")
	}
	if job.MediaType != "movie" && job.MediaType != "show" {
		return fmt.Errorf("media type must be 'movie' or 'show'")
	}
	if job.Name == "" {
		return fmt.Errorf("job name is required")
	}

	// Validate against registry
	typeDef, ok := jobs.GetJobTypeDefinition(job.Type)
	if !ok {
		return fmt.Errorf("unknown job type: %s", job.Type)
	}

	// Set default limit if not specified
	if job.Limit <= 0 {
		job.Limit = typeDef.DefaultLimit
	}

	// Check max limit
	if job.Limit > typeDef.MaxLimit {
		return fmt.Errorf("limit cannot exceed %d for %s jobs", typeDef.MaxLimit, job.Type)
	}

	// Check media type support
	if !jobs.SupportsMediaType(job.Type, job.MediaType) {
		return fmt.Errorf("%s jobs do not support media type '%s'", job.Type, job.MediaType)
	}

	// Check period requirement
	if typeDef.RequiresPeriod && job.Period == "" {
		return fmt.Errorf("%s jobs require a period (weekly, monthly, yearly, or all)", job.Type)
	}

	// Validate period value if provided
	if job.Period != "" {
		validPeriods := []string{"weekly", "monthly", "yearly", "all"}
		isValid := false
		for _, p := range validPeriods {
			if job.Period == p {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid period: %s (must be weekly, monthly, yearly, or all)", job.Period)
		}
	}

	return nil
}

// previewDynamicJob returns a preview for a dynamic job
// This routes to the appropriate existing preview function based on job type
func previewDynamicJob(cfg *config.Config, db *database.Database, job config.DynamicJob) interface{} {
	typeDef, ok := jobs.GetJobTypeDefinition(job.Type)
	if !ok {
		return fiber.Map{
			"error": "Unknown job type: " + job.Type,
		}
	}

	effectiveLimit := job.Limit
	if effectiveLimit <= 0 {
		effectiveLimit = typeDef.DefaultLimit
	}
	if effectiveLimit > typeDef.MaxLimit {
		effectiveLimit = typeDef.MaxLimit
	}

	cfgCopy := *cfg

	switch job.Type {
	case "trending":
		if job.MediaType == "movie" {
			cfgCopy.Jobs.TrendingMovies.Limit = effectiveLimit
			cfgCopy.Jobs.TrendingMovies.Mode = job.Mode
			return jobs.PreviewTrendingMovies(&cfgCopy, db)
		}
		cfgCopy.Jobs.TrendingShows.Limit = effectiveLimit
		cfgCopy.Jobs.TrendingShows.Mode = job.Mode
		return jobs.PreviewTrendingShows(&cfgCopy, db)
	case "popular":
		if job.MediaType == "movie" {
			cfgCopy.Jobs.PopularMovies.Limit = effectiveLimit
			cfgCopy.Jobs.PopularMovies.Mode = job.Mode
			return jobs.PreviewPopularMovies(&cfgCopy, db)
		}
		cfgCopy.Jobs.PopularShows.Limit = effectiveLimit
		cfgCopy.Jobs.PopularShows.Mode = job.Mode
		return jobs.PreviewPopularShows(&cfgCopy, db)
	case "watched":
		if job.MediaType == "movie" {
			cfgCopy.Jobs.WatchedMovies.Limit = effectiveLimit
			cfgCopy.Jobs.WatchedMovies.Period = job.Period
			cfgCopy.Jobs.WatchedMovies.Mode = job.Mode
			return jobs.PreviewWatchedMovies(&cfgCopy, db)
		}
		cfgCopy.Jobs.WatchedShows.Limit = effectiveLimit
		cfgCopy.Jobs.WatchedShows.Period = job.Period
		cfgCopy.Jobs.WatchedShows.Mode = job.Mode
		return jobs.PreviewWatchedShows(&cfgCopy, db)
	case "collected":
		if job.MediaType == "movie" {
			cfgCopy.Jobs.CollectedMovies.Limit = effectiveLimit
			cfgCopy.Jobs.CollectedMovies.Period = job.Period
			cfgCopy.Jobs.CollectedMovies.Mode = job.Mode
			return jobs.PreviewCollectedMovies(&cfgCopy, db)
		}
		cfgCopy.Jobs.CollectedShows.Limit = effectiveLimit
		cfgCopy.Jobs.CollectedShows.Period = job.Period
		cfgCopy.Jobs.CollectedShows.Mode = job.Mode
		return jobs.PreviewCollectedShows(&cfgCopy, db)
	case "favorited":
		if job.MediaType == "movie" {
			cfgCopy.Jobs.FavoritedMovies.Limit = effectiveLimit
			cfgCopy.Jobs.FavoritedMovies.Period = job.Period
			cfgCopy.Jobs.FavoritedMovies.Mode = job.Mode
			return jobs.PreviewFavoritedMovies(&cfgCopy, db)
		}
		cfgCopy.Jobs.FavoritedShows.Limit = effectiveLimit
		cfgCopy.Jobs.FavoritedShows.Period = job.Period
		cfgCopy.Jobs.FavoritedShows.Mode = job.Mode
		return jobs.PreviewFavoritedShows(&cfgCopy, db)
	case "played":
		if job.MediaType == "movie" {
			cfgCopy.Jobs.PlayedMovies.Limit = effectiveLimit
			cfgCopy.Jobs.PlayedMovies.Period = job.Period
			cfgCopy.Jobs.PlayedMovies.Mode = job.Mode
			return jobs.PreviewPlayedMovies(&cfgCopy, db)
		}
		cfgCopy.Jobs.PlayedShows.Limit = effectiveLimit
		cfgCopy.Jobs.PlayedShows.Period = job.Period
		cfgCopy.Jobs.PlayedShows.Mode = job.Mode
		return jobs.PreviewPlayedShows(&cfgCopy, db)
	case "anticipated":
		if job.MediaType == "movie" {
			cfgCopy.Jobs.AnticipatedMovies.Limit = effectiveLimit
			cfgCopy.Jobs.AnticipatedMovies.Mode = job.Mode
			return jobs.PreviewAnticipatedMovies(&cfgCopy, db)
		}
		cfgCopy.Jobs.AnticipatedShows.Limit = effectiveLimit
		cfgCopy.Jobs.AnticipatedShows.Mode = job.Mode
		return jobs.PreviewAnticipatedShows(&cfgCopy, db)
	case "box_office":
		cfgCopy.Jobs.BoxOffice.Limit = effectiveLimit
		cfgCopy.Jobs.BoxOffice.Mode = job.Mode
		return jobs.PreviewBoxOffice(&cfgCopy, db)
	case "smart_popular":
		baseMinRating := job.BaseMinRating
		if baseMinRating == 0 {
			baseMinRating = 6.0
		}
		adjustmentFactor := job.AdjustmentFactor
		if adjustmentFactor == 0 {
			adjustmentFactor = 0.5
		}

		if job.MediaType == "movie" {
			cfgCopy.Jobs.SmartPopularMovies.Limit = effectiveLimit
			cfgCopy.Jobs.SmartPopularMovies.Mode = job.Mode
			cfgCopy.Jobs.SmartPopularMovies.BaseMinRating = baseMinRating
			cfgCopy.Jobs.SmartPopularMovies.AdjustmentFactor = adjustmentFactor
			return jobs.PreviewSmartPopularMovies(&cfgCopy, db)
		}
		cfgCopy.Jobs.SmartPopularShows.Limit = effectiveLimit
		cfgCopy.Jobs.SmartPopularShows.Mode = job.Mode
		cfgCopy.Jobs.SmartPopularShows.BaseMinRating = baseMinRating
		cfgCopy.Jobs.SmartPopularShows.AdjustmentFactor = adjustmentFactor
		return jobs.PreviewSmartPopularShows(&cfgCopy, db)
	default:
		return fiber.Map{
			"error": "Preview not supported for job type: " + job.Type,
		}
	}
}
