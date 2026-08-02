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
	router.Get("/rule-sets", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		items := make([]fiber.Map, 0, len(cfg.RuleSets))
		for _, rules := range cfg.RuleSets {
			items = append(items, fiber.Map{"rule_set": rules, "usage_count": cfg.RuleSetUsage(rules.ID)})
		}
		return c.JSON(items)
	})
	router.Post("/rule-sets", func(c *fiber.Ctx) error {
		var rules config.RuleSet
		if err := c.BodyParser(&rules); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
		}
		if rules.ID == "" {
			rules.ID = uuid.NewString()
		}
		rules.Revision = 1
		cfg := gctx.Config()
		if _, exists := cfg.RuleSetByID(rules.ID); exists {
			return c.Status(409).JSON(fiber.Map{"error": "Rule set ID already exists"})
		}
		if err := cfg.ValidateRuleSet(rules, ""); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		cfg.RuleSets = append(cfg.RuleSets, rules)
		if err := cfg.Save(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if err := gctx.ReloadConfig(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(201).JSON(rules)
	})
	router.Put("/rule-sets/:id", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		id := c.Params("id")
		current, ok := cfg.RuleSetByID(id)
		if !ok {
			return c.Status(404).JSON(fiber.Map{"error": "Rule set not found"})
		}
		var rules config.RuleSet
		if err := c.BodyParser(&rules); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
		}
		if rules.ID != "" && rules.ID != id {
			return c.Status(400).JSON(fiber.Map{"error": "Rule set ID cannot be changed"})
		}
		if rules.Revision != current.Revision {
			return c.Status(409).JSON(fiber.Map{"error": "Rule set changed since it was opened; reload and try again"})
		}
		rules.ID, rules.Revision = id, current.Revision+1
		if err := cfg.ValidateRuleSet(rules, id); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		*current = rules
		if err := cfg.Save(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if err := gctx.ReloadConfig(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(rules)
	})
	router.Delete("/rule-sets/:id", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		id := c.Params("id")
		if id == config.DefaultMoviesRuleSetID || id == config.DefaultShowsRuleSetID {
			return c.Status(409).JSON(fiber.Map{"error": "Default rule sets cannot be deleted"})
		}
		if count := cfg.RuleSetUsage(id); count > 0 {
			return c.Status(409).JSON(fiber.Map{"error": fmt.Sprintf("Rule set is used by %d job(s)", count)})
		}
		found := false
		for i := range cfg.RuleSets {
			if cfg.RuleSets[i].ID == id {
				cfg.RuleSets = append(cfg.RuleSets[:i], cfg.RuleSets[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return c.Status(404).JSON(fiber.Map{"error": "Rule set not found"})
		}
		if err := cfg.Save(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if err := gctx.ReloadConfig(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(204)
	})
	router.Get("/title-exceptions", func(c *fiber.Ctx) error { return c.JSON(gctx.Config().TitleExceptions) })
	router.Put("/title-exceptions", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		var exceptions config.TitleExceptions
		if err := c.BodyParser(&exceptions); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
		}
		cfg.TitleExceptions = exceptions
		if err := cfg.Save(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		if err := gctx.ReloadConfig(); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(exceptions)
	})

	router.Post("/jobs/:id/customize-rules", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		job := cfg.GetDynamicJobByID(c.Params("id"))
		if job == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Job not found"})
		}
		assigned, _, err := cfg.ResolveRuleSet(*job)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		clone := *assigned
		clone.ID = uuid.NewString()
		clone.Revision = 1
		baseName := strings.TrimSpace(job.Name) + " Rules"
		clone.Name = baseName
		for suffix := 2; ; suffix++ {
			if err := cfg.ValidateRuleSet(clone, ""); err == nil {
				break
			}
			clone.Name = fmt.Sprintf("%s %d", baseName, suffix)
		}
		previousRuleSetID := job.RuleSetID
		cfg.RuleSets = append(cfg.RuleSets, clone)
		job.RuleSetID = clone.ID
		if err := cfg.Save(); err != nil {
			cfg.RuleSets = cfg.RuleSets[:len(cfg.RuleSets)-1]
			job.RuleSetID = previousRuleSetID
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if err := gctx.ReloadConfig(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"job": job, "rule_set": clone, "usage_count": 1})
	})

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
		if job.RuleSetID == "" {
			if job.MediaType == "show" {
				job.RuleSetID = config.DefaultShowsRuleSetID
			} else {
				job.RuleSetID = config.DefaultMoviesRuleSetID
			}
		}
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
		if job.RuleSetID == "" {
			if job.MediaType == "show" {
				job.RuleSetID = config.DefaultShowsRuleSetID
			} else {
				job.RuleSetID = config.DefaultMoviesRuleSetID
			}
		}
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
			_ = jobs.RunDynamicJob(gctx, cfg, gctx.Database(), *targetJob, dryRun)
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
	if job.UseCustomFilters {
		if err := validateFilterBounds(job.Filters.Movies.BlacklistedMinYear, job.Filters.Movies.BlacklistedMaxYear, job.Filters.Movies.BlacklistedMinRuntime, job.Filters.Movies.BlacklistedMaxRuntime, job.Filters.Movies.MinRating, job.Filters.Movies.MinVotes); err != nil {
			return fmt.Errorf("movie filters: %w", err)
		}
		if err := validateFilterBounds(job.Filters.Shows.BlacklistedMinYear, job.Filters.Shows.BlacklistedMaxYear, job.Filters.Shows.BlacklistedMinRuntime, job.Filters.Shows.BlacklistedMaxRuntime, job.Filters.Shows.MinRating, job.Filters.Shows.MinVotes); err != nil {
			return fmt.Errorf("show filters: %w", err)
		}
	}
	if _, _, err := cfg.ResolveRuleSet(job); err != nil {
		return err
	}

	return nil
}

func validateFilterBounds(minYear, maxYear, minRuntime, maxRuntime int, minRating float64, minVotes int) error {
	if minYear < 0 || maxYear < 0 || minRuntime < 0 || maxRuntime < 0 || minVotes < 0 {
		return fmt.Errorf("numeric values cannot be negative")
	}
	if minRating < 0 || minRating > 10 {
		return fmt.Errorf("minimum rating must be between 0 and 10")
	}
	if minYear > 0 && maxYear > 0 && minYear > maxYear {
		return fmt.Errorf("minimum year cannot exceed maximum year")
	}
	if minRuntime > 0 && maxRuntime > 0 && minRuntime > maxRuntime {
		return fmt.Errorf("minimum runtime cannot exceed maximum runtime")
	}
	return nil
}
