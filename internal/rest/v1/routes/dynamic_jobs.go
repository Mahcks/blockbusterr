package routes

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/services/jobs"
)

// AddDynamicJobsRoutes adds the dynamic job management API endpoints
func AddDynamicJobsRoutes(router fiber.Router, gctx global.Context) {
	// Get all available job type definitions (for UI dropdowns)
	router.Get("/jobs/types", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		providers := make([]string, 0, 3)
		if cfg.Trakt.ClientID != "" {
			providers = append(providers, "trakt")
		}
		if cfg.TMDB.APIKey != "" {
			providers = append(providers, "tmdb")
		}
		if cfg.Simkl.ClientID != "" {
			providers = append(providers, "simkl")
		}
		return c.JSON(jobs.GetAvailableJobTypes(providers))
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
		cfg := gctx.Config()
		if err := validateDynamicJob(cfg, job); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

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
		cfg := gctx.Config()
		if err := validateDynamicJob(cfg, job); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

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
		if !jobs.IsProviderConfigured(cfg, targetJob.Source) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": fmt.Sprintf("%s is not configured", targetJob.Source),
			})
		}

		// Execute job asynchronously
		go func() {
			dryRun := gctx.Metadata().Version == "dev"
			_ = jobs.RunDynamicJob(cfg, gctx.Database(), *targetJob, dryRun)
		}()

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message": fmt.Sprintf("Job '%s' queued", targetJob.Name),
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

		preview, err := jobs.PreviewDynamicJob(cfg, gctx.Database(), *targetJob)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}
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
func validateDynamicJob(cfg *config.Config, job config.DynamicJob) error {
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
	if job.Source == "" {
		job.Source = "trakt"
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
	if !jobs.SupportsSource(job.Type, job.Source) {
		return fmt.Errorf("%s jobs do not support the %s source", job.Type, job.Source)
	}
	if !jobs.IsProviderConfigured(cfg, job.Source) {
		return fmt.Errorf("%s is not configured", job.Source)
	}
	if job.Source == "simkl" && job.Limit > 500 {
		return fmt.Errorf("simkl jobs cannot exceed 500 items")
	}

	// Check period requirement
	if typeDef.RequiresPeriod && job.Period == "" {
		return fmt.Errorf("%s jobs require a period (weekly, monthly, yearly, or all)", job.Type)
	}

	// Validate period value if provided
	if job.Period != "" {
		validPeriods := []string{"weekly", "monthly", "yearly", "all"}
		if !slices.Contains(validPeriods, job.Period) {
			return fmt.Errorf("invalid period: %s (must be weekly, monthly, yearly, or all)", job.Period)
		}
	}
	if job.Source == "simkl" && job.Type == "watched" && job.Period != "weekly" && job.Period != "monthly" {
		return fmt.Errorf("simkl most watched jobs support weekly or monthly periods")
	}

	return nil
}
