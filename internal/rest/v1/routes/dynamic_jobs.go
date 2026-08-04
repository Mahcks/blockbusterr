package routes

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/internal/services/jobs"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

// AddDynamicJobsRoutes adds the dynamic job management API endpoints
func AddDynamicJobsRoutes(router fiber.Router, gctx global.Context) {
	router.Get("/mdblist/validate", func(c *fiber.Ctx) error {
		apiKey := c.Query("api_key")
		if apiKey == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "API key is required"})
		}
		if err := integrations.NewMDBList(integrations.MDBListConfig{APIKey: apiKey}).Validate(c.Context()); err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"connected": true, "message": "Connected to MDBList."})
	})

	router.Post("/jobs/lists/inspect", func(c *fiber.Ctx) error {
		var request struct {
			Source string             `json:"source"`
			List   config.ListLocator `json:"list"`
		}
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
		}
		if err := jobs.ValidateListSourceLocator(request.Source, request.List); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		inspection, err := jobs.InspectListSource(c.Context(), gctx.Config(), request.Source, request.List)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(inspection)
	})

	registerAccountAuthRoutes(router, gctx)

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
		config.ApplyRuleSetDefaults(&rules)
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
		config.ApplyRuleSetDefaults(&rules)
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
		return c.JSON(jobs.GetAvailableJobTypes(providers, jobs.AvailableListSources(cfg)))
	})

	// Get all pre-configured job templates
	router.Get("/jobs/templates", func(c *fiber.Ctx) error {
		return c.JSON(jobs.GetAllTemplates(gctx.Config()))
	})
	router.Post("/jobs/selection/preview", func(c *fiber.Ctx) error {
		preview, err := jobs.PreviewSelectionCycle(gctx.Config(), gctx.Database())
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(preview)
	})

	router.Post("/jobs/recipes/:id", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		recipe, ok := jobs.GetTemplate(c.Params("id"))
		if !ok {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Recipe not found"})
		}
		for _, available := range jobs.GetAllTemplates(cfg) {
			if available.ID == recipe.ID && !available.Ready {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": available.Missing})
			}
		}

		candidate := *cfg
		candidate.RuleSets = append([]config.RuleSet(nil), cfg.RuleSets...)
		candidate.Jobs.List = append([]config.DynamicJob(nil), cfg.Jobs.List...)
		ruleSetID := config.DefaultMoviesRuleSetID
		if recipe.MediaType == "show" {
			ruleSetID = config.DefaultShowsRuleSetID
		}
		if !recipe.DefaultRules {
			rules := config.RuleSet{ID: uuid.NewString(), Name: availableRecipeRuleName(&candidate, recipe.RuleSetName, recipe.MediaType), Media: recipe.MediaType, Revision: 1}
			if recipe.Movies != nil {
				filters := *recipe.Movies
				rules.Movies = &filters
			}
			if recipe.Shows != nil {
				filters := *recipe.Shows
				rules.Shows = &filters
			}
			config.ApplyRuleSetDefaults(&rules)
			if err := candidate.ValidateRuleSet(rules, ""); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Recipe rules are invalid: " + err.Error()})
			}
			candidate.RuleSets = append(candidate.RuleSets, rules)
			ruleSetID = rules.ID
		}
		job := config.DynamicJob{ID: uuid.NewString(), Name: recipe.Name, Enabled: false, Type: recipe.Type, Source: recipe.Source, MediaType: recipe.MediaType, Limit: recipe.Limit, DeliveryLimit: recipe.DeliveryLimit, Period: recipe.Period, SyncInterval: recipe.SyncInterval, Mode: recipe.Mode, SeriesType: recipe.SeriesType, RuleSetID: ruleSetID}
		if recipe.List != nil {
			locator := *recipe.List
			job.List = &locator
		}
		if err := validateDynamicJob(&candidate, job); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Recipe job is invalid: " + err.Error()})
		}
		candidate.Jobs.List = append(candidate.Jobs.List, job)
		if err := candidate.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save recipe: " + err.Error()})
		}
		*cfg = candidate
		if err := gctx.ReloadConfig(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Recipe saved but configuration failed to reload: " + err.Error()})
		}
		return c.Status(fiber.StatusCreated).JSON(job)
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
		if cfg.Jobs.Selection.Enabled && targetJob.SelectionCycle {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "This job is managed by ranked selection and cannot run independently",
			})
		}
		if !jobs.IsJobSourceConfigured(cfg, *targetJob) {
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
		jobsBefore := cfg.Jobs
		jobsBefore.List = append([]config.DynamicJob(nil), cfg.Jobs.List...)

		migratedJobs, err := cfg.MigrateLegacyJobs()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to migrate jobs: " + err.Error(),
			})
		}

		// Save config to persist changes
		if err := cfg.Save(); err != nil {
			cfg.Jobs = jobsBefore
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

		message := "No legacy jobs to migrate"
		if len(migratedJobs) > 0 {
			message = fmt.Sprintf("Successfully migrated %d legacy jobs", len(migratedJobs))
		}
		return c.JSON(fiber.Map{
			"message": message,
			"count":   len(migratedJobs),
			"jobs":    migratedJobs,
		})
	})
}

func registerAccountAuthRoutes(router fiber.Router, gctx global.Context) {
	router.Post("/auth/trakt/device", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		if cfg.Trakt.ClientID == "" || cfg.Trakt.ClientSecret == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Save the Trakt client ID and secret first"})
		}
		code, err := integrations.NewTrakt(integrations.TraktConfig{ClientID: cfg.Trakt.ClientID, ClientSecret: cfg.Trakt.ClientSecret}).StartDeviceAuth(c.Context())
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(code)
	})
	router.Post("/auth/trakt/device/poll", func(c *fiber.Ctx) error {
		var request struct {
			DeviceCode string `json:"device_code"`
		}
		if err := c.BodyParser(&request); err != nil || request.DeviceCode == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Device code is required"})
		}
		cfg := gctx.Config()
		token, status, err := integrations.NewTrakt(integrations.TraktConfig{ClientID: cfg.Trakt.ClientID, ClientSecret: cfg.Trakt.ClientSecret}).PollDeviceAuth(c.Context(), request.DeviceCode)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}
		if status != fiber.StatusOK {
			return c.Status(status).JSON(fiber.Map{"connected": false})
		}
		previousAccess, previousRefresh, previousExpiry := cfg.Trakt.AccessToken, cfg.Trakt.RefreshToken, cfg.Trakt.TokenExpires
		cfg.Trakt.AccessToken, cfg.Trakt.RefreshToken, cfg.Trakt.TokenExpires = token.AccessToken, token.RefreshToken, token.ExpiresAt()
		if err := cfg.Save(); err != nil {
			cfg.Trakt.AccessToken, cfg.Trakt.RefreshToken, cfg.Trakt.TokenExpires = previousAccess, previousRefresh, previousExpiry
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"connected": true})
	})
	router.Post("/auth/tmdb/start", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		if cfg.TMDB.APIKey == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Save the TMDB API key first"})
		}
		token, err := integrations.NewTMDB(integrations.TMDBConfig{APIKey: cfg.TMDB.APIKey}).CreateRequestToken(c.Context())
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"request_token": token.RequestToken, "authorize_url": "https://www.themoviedb.org/authenticate/" + token.RequestToken, "expires_at": token.ExpiresAt})
	})
	router.Post("/auth/tmdb/complete", func(c *fiber.Ctx) error {
		var request struct {
			RequestToken string `json:"request_token"`
		}
		if err := c.BodyParser(&request); err != nil || request.RequestToken == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Request token is required"})
		}
		cfg := gctx.Config()
		client := integrations.NewTMDB(integrations.TMDBConfig{APIKey: cfg.TMDB.APIKey})
		session, err := client.CreateSession(c.Context(), request.RequestToken)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}
		account, err := client.GetAccount(c.Context(), session.SessionID)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}
		previousSession, previousAccount := cfg.TMDB.SessionID, cfg.TMDB.AccountID
		cfg.TMDB.SessionID, cfg.TMDB.AccountID = session.SessionID, account.ID
		if err := cfg.Save(); err != nil {
			cfg.TMDB.SessionID, cfg.TMDB.AccountID = previousSession, previousAccount
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"connected": true, "username": account.Username})
	})
	router.Delete("/auth/:provider", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		previousTrakt := cfg.Trakt
		previousTMDB := cfg.TMDB
		switch c.Params("provider") {
		case "trakt":
			cfg.Trakt.AccessToken, cfg.Trakt.RefreshToken, cfg.Trakt.TokenExpires = "", "", 0
		case "tmdb":
			cfg.TMDB.SessionID, cfg.TMDB.AccountID = "", 0
		default:
			return c.SendStatus(fiber.StatusNotFound)
		}
		if err := cfg.Save(); err != nil {
			cfg.Trakt, cfg.TMDB = previousTrakt, previousTMDB
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(fiber.StatusNoContent)
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
	if job.DeliveryLimit < 0 {
		return fmt.Errorf("delivery limit cannot be negative")
	}
	if job.MinimumPicks < 0 {
		return fmt.Errorf("minimum picks cannot be negative")
	}
	if !job.SelectionCycle && job.MinimumPicks > 0 {
		return fmt.Errorf("minimum picks require ranked selection")
	}
	if job.SelectionCycle {
		if job.Type == "smart_popular" {
			return fmt.Errorf("adaptive smart jobs cannot join ranked selection")
		}
		if cfg.Jobs.Selection.Enabled && !cfg.Scoring.Enabled {
			return fmt.Errorf("ranked selection requires content scoring")
		}
		if cfg.Jobs.Selection.Enabled {
			capacity, minima := cfg.Jobs.Selection.MovieLimit, job.MinimumPicks
			if job.MediaType == "show" {
				capacity = cfg.Jobs.Selection.ShowLimit
			}
			for _, existing := range cfg.Jobs.List {
				if existing.ID != job.ID && existing.Enabled && existing.SelectionCycle && existing.MediaType == job.MediaType {
					minima += existing.MinimumPicks
				}
			}
			if capacity <= 0 || minima > capacity {
				return fmt.Errorf("ranked %s minimum picks require capacity %d or greater", job.MediaType, minima)
			}
		}
	}
	if !enums.RepeatPolicy(job.RepeatPolicy).IsValid(true) {
		return fmt.Errorf("repeat handling policy is invalid")
	}

	// Check media type support
	if !jobs.SupportsMediaType(job.Type, job.MediaType) {
		return fmt.Errorf("%s jobs do not support media type '%s'", job.Type, job.MediaType)
	}
	if job.Type == string(enums.JobTypeList) {
		if job.List == nil {
			return fmt.Errorf("list locator is required")
		}
		if err := jobs.ValidateListSourceLocator(job.Source, *job.List); err != nil {
			return err
		}
		if !slices.Contains(jobs.AvailableListSources(cfg), job.Source) {
			return fmt.Errorf("%s list adapter is unavailable", job.Source)
		}
		if enums.ListKind(job.List.Kind) == enums.ListKindWatchlist {
			switch job.Source {
			case "trakt":
				if cfg.Trakt.AccessToken == "" && job.List.Owner == "" {
					return fmt.Errorf("connect a Trakt account or provide a public watchlist owner")
				}
			case "tmdb":
				if cfg.TMDB.SessionID == "" || cfg.TMDB.AccountID == 0 {
					return fmt.Errorf("connect a TMDB account to use its watchlist")
				}
			case "mdblist":
				if cfg.MDBList.APIKey == "" {
					return fmt.Errorf("configure an MDBList API key to use its watchlist")
				}
			}
		}
	} else {
		if job.Type == string(enums.JobTypeRecommendations) {
			if len(job.RecommendationSeeds) == 0 && job.RecommendationList == nil {
				return fmt.Errorf("TMDB seed IDs or a provider seed list are required")
			}
			if len(job.RecommendationSeeds) > 20 {
				return fmt.Errorf("recommendation jobs support at most 20 seed IDs")
			}
			for _, seed := range job.RecommendationSeeds {
				if seed <= 0 {
					return fmt.Errorf("TMDB seed IDs must be positive numbers")
				}
			}
			if job.RecommendationList != nil {
				if err := jobs.ValidateListSourceLocator(job.RecommendationList.Source, job.RecommendationList.List); err != nil {
					return fmt.Errorf("recommendation seed list: %w", err)
				}
				if !slices.Contains(jobs.AvailableListSources(cfg), job.RecommendationList.Source) {
					return fmt.Errorf("%s seed-list adapter is unavailable", job.RecommendationList.Source)
				}
			}
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
	}
	if job.SeriesType != "" && !slices.Contains([]string{"standard", "daily", "anime"}, job.SeriesType) {
		return fmt.Errorf("sonarr series type must be standard, daily, or anime")
	}
	if job.MediaType != "show" && job.SeriesType != "" {
		return fmt.Errorf("sonarr series type is only valid for show jobs")
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

func availableRecipeRuleName(cfg *config.Config, base, media string) string {
	name := base
	for suffix := 2; ; suffix++ {
		available := true
		for _, rules := range cfg.RuleSets {
			if rules.Media == media && strings.EqualFold(rules.Name, name) {
				available = false
				break
			}
		}
		if available {
			return name
		}
		name = fmt.Sprintf("%s %d", base, suffix)
	}
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
