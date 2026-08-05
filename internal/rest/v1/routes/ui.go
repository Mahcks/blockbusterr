package routes

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/services/jobs"
	"github.com/mahcks/blockbusterr/pkg/enums"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

type readinessState string

const (
	readinessReady      readinessState = "ready"
	readinessAttention  readinessState = "attention"
	readinessNotStarted readinessState = "not_started"
)

type readinessItem struct {
	State   readinessState
	Title   string
	Message string
	Href    string
}

type systemReadiness struct {
	Discovery  readinessItem
	Delivery   readinessItem
	Automation readinessItem
	Complete   bool
}

func plural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func assessReadiness(cfg *config.Config) systemReadiness {
	discoveryCount := 0
	for _, configured := range []bool{cfg.TMDB.APIKey != "", cfg.Simkl.ClientID != "", cfg.Trakt.ClientID != "", cfg.MDBList.APIKey != "", cfg.Letterboxd.ExperimentalScraping} {
		if configured {
			discoveryCount++
		}
	}
	discovery := readinessItem{State: readinessReady, Title: "Discovery ready", Message: fmt.Sprintf("%d provider%s configured", discoveryCount, plural(discoveryCount)), Href: "/config#connections"}
	if discoveryCount == 0 {
		discovery = readinessItem{State: readinessNotStarted, Title: "Connect discovery", Message: "Add TMDB, Simkl, Trakt, or MDBList", Href: "/config#connections"}
	}

	radarrReady := cfg.Radarr.URL != "" && cfg.Radarr.APIKey != ""
	sonarrReady := cfg.Sonarr.URL != "" && cfg.Sonarr.APIKey != ""
	jellyseerrReady := cfg.Jellyseerr.URL != "" && cfg.Jellyseerr.APIKey != ""
	delivery := readinessItem{State: readinessReady, Title: "Delivery ready", Href: "/config#connections"}
	if cfg.Jobs.Mode == "jellyseerr" {
		delivery.Message = "Requests go through Jellyseerr / Seerr"
		if !jellyseerrReady {
			delivery.State, delivery.Title, delivery.Message = readinessAttention, "Finish delivery setup", "Jellyseerr / Seerr needs a URL and API key"
		}
	} else {
		switch {
		case radarrReady && sonarrReady:
			delivery.Message = "Movies use Radarr; shows use Sonarr"
		case radarrReady:
			delivery.State, delivery.Title, delivery.Message = readinessAttention, "Shows need a destination", "Radarr is ready; connect Sonarr for shows"
		case sonarrReady:
			delivery.State, delivery.Title, delivery.Message = readinessAttention, "Movies need a destination", "Sonarr is ready; connect Radarr for movies"
		default:
			delivery.State, delivery.Title, delivery.Message = readinessNotStarted, "Connect delivery", "Add Radarr, Sonarr, or use Jellyseerr / Seerr"
		}
	}

	enabled, invalid := 0, 0
	for _, job := range cfg.Jobs.List {
		if !job.Enabled {
			continue
		}
		enabled++
		if !jobs.IsJobSourceConfigured(cfg, job) {
			invalid++
			continue
		}
		if _, _, err := cfg.ResolveRuleSet(job); err != nil {
			invalid++
			continue
		}
		mode := jobs.DetermineMode(job.Mode, cfg.Jobs.Mode)
		if mode == "jellyseerr" && !jellyseerrReady || mode == "direct" && job.MediaType == "movie" && !radarrReady || mode == "direct" && job.MediaType == "show" && !sonarrReady {
			invalid++
		}
	}
	automation := readinessItem{State: readinessReady, Title: "Automation ready", Message: fmt.Sprintf("%d enabled job%s", enabled, plural(enabled)), Href: "/jobs"}
	if enabled == 0 {
		automation = readinessItem{State: readinessNotStarted, Title: "Create your first job", Message: "Choose a source, rules, and delivery target", Href: "/jobs"}
	} else if invalid > 0 {
		automation = readinessItem{State: readinessAttention, Title: "Automation needs attention", Message: fmt.Sprintf("%d enabled job%s cannot run", invalid, plural(invalid)), Href: "/jobs"}
	}

	return systemReadiness{Discovery: discovery, Delivery: delivery, Automation: automation, Complete: discovery.State == readinessReady && delivery.State == readinessReady && automation.State == readinessReady}
}

// determineConfigPath finds the best location to save config file
// Prioritizes Docker data directory, falls back to local directory
func determineConfigPath() string {
	// Try /app/data first (Docker volume mount)
	if _, err := os.Stat("/app/data"); err == nil {
		return "/app/data/config.yaml"
	}

	// Try data directory (local)
	if _, err := os.Stat("./data"); err == nil {
		return "./data/config.yaml"
	}

	// Fall back to current directory
	return "./config.yaml"
}

// preserveBlankSecret keeps an existing credential unless the user supplies a replacement.
func preserveBlankSecret(c *fiber.Ctx, field string, destination *string) {
	if value := strings.TrimSpace(c.FormValue(field)); value != "" {
		*destination = value
	}
}

// RegisterUIRoutes handles web UI routes
func RegisterUIRoutes(rg *RouteGroup, app *fiber.App) {
	// Root route for web UI - show jobs page with full context
	app.Get("/", func(c *fiber.Ctx) error {
		alert := structures.AlertInfo{
			ID:      "jobs-info",
			Title:   "About Jobs",
			Content: "Jobs fetch content from the selected discovery source and add it to Radarr/Sonarr based on the sync interval. Click a job to configure its source and settings.",
			Class:   "mb-4",
		}
		cfg := rg.gctx.Config()
		discoveryDisabled := cfg.Trakt.ClientID == "" && cfg.TMDB.APIKey == "" && cfg.Simkl.ClientID == ""
		return c.Render("jobs", fiber.Map{
			"Title":             "Blockbusterr - Jobs",
			"Config":            cfg,
			"Version":           rg.gctx.Metadata().Version,
			"AlertInfo":         alert,
			"DiscoveryDisabled": discoveryDisabled,
			"Readiness":         assessReadiness(cfg),
		}, "base")
	})

	// Configuration page route
	app.Get("/config", func(c *fiber.Ctx) error {
		alert := structures.AlertInfo{
			ID:      "welcome-info",
			Title:   "Getting Started",
			Content: "Choose TMDB, Simkl, or Trakt for discovery jobs, then connect Radarr, Sonarr, Jellyseerr, or Seerr for delivery.",
			Class:   "mb-4",
		}
		cfg := rg.gctx.Config()
		budgetUsage := database.DeliveryBudgetUsage{}
		if db := rg.gctx.Database(); db != nil {
			budgetUsage, _ = db.GetDeliveryBudgetUsage(cfg.Jobs.GlobalPeriod)
		}
		return c.Render("index", fiber.Map{
			"Title":       "Blockbusterr - Configuration",
			"Config":      cfg,
			"Version":     rg.gctx.Metadata().Version,
			"AlertInfo":   alert,
			"Readiness":   assessReadiness(cfg),
			"BudgetUsage": budgetUsage,
		}, "base")
	})

	// Jobs page route (explicit)
	app.Get("/jobs", func(c *fiber.Ctx) error {
		alert := structures.AlertInfo{
			ID:      "jobs-info",
			Title:   "About Jobs",
			Content: "Jobs fetch content from the selected discovery source and add it to Radarr/Sonarr based on the sync interval. Click a job to configure its source and settings.",
			Class:   "mb-4",
		}
		cfg := rg.gctx.Config()
		discoveryDisabled := cfg.Trakt.ClientID == "" && cfg.TMDB.APIKey == "" && cfg.Simkl.ClientID == ""
		return c.Render("jobs", fiber.Map{
			"Title":             "Blockbusterr - Jobs",
			"Config":            cfg,
			"Version":           rg.gctx.Metadata().Version,
			"AlertInfo":         alert,
			"DiscoveryDisabled": discoveryDisabled,
			"Readiness":         assessReadiness(cfg),
		}, "base")
	})

	// Activity log page route
	app.Get("/activity", func(c *fiber.Ctx) error {
		cfg := rg.gctx.Config()
		return c.Render("activity", fiber.Map{
			"Title":     "Blockbusterr - Activity Log",
			"Config":    cfg,
			"Version":   rg.gctx.Metadata().Version,
			"Readiness": assessReadiness(cfg),
		}, "base")
	})

	// Config save route. Validates every value before mutating the live
	// config, so a bad submission never leaves partially-applied state in
	// memory even though the save itself failed.
	app.Post("/config/save", func(c *fiber.Ctx) error {
		cfg := rg.gctx.Config()

		var fieldErrors []string
		parseInt := func(field, formKey string, fallback int) int {
			raw := c.FormValue(formKey)
			if raw == "" {
				return fallback
			}
			val, err := strconv.Atoi(raw)
			if err != nil {
				fieldErrors = append(fieldErrors, field+" must be a whole number")
				return fallback
			}
			return val
		}
		parseNonNegativeInt := func(field, formKey string, fallback int) int {
			val := parseInt(field, formKey, fallback)
			if val < 0 {
				fieldErrors = append(fieldErrors, field+" cannot be negative")
			}
			return val
		}
		parseFloat := func(field, formKey string, fallback float64) float64 {
			raw := c.FormValue(formKey)
			if raw == "" {
				return fallback
			}
			val, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				fieldErrors = append(fieldErrors, field+" must be a number")
				return fallback
			}
			return val
		}

		radarrQualityProfile := parseInt("Radarr quality profile", "radarr.quality_profile", cfg.Radarr.QualityProfile)
		sonarrQualityProfile := parseInt("Sonarr quality profile", "sonarr.quality_profile", cfg.Sonarr.QualityProfile)
		globalLimitMovies := parseNonNegativeInt("Movie delivery limit", "jobs.global_limit_movies", cfg.Jobs.GlobalLimitMovies)
		globalLimitShows := parseNonNegativeInt("Show delivery limit", "jobs.global_limit_shows", cfg.Jobs.GlobalLimitShows)
		selectionEnabled := c.FormValue("jobs.selection.enabled") == "true"
		selectionInterval := c.FormValue("jobs.selection.sync_interval")
		if selectionInterval == "" {
			selectionInterval = "24h"
		}
		if duration, err := time.ParseDuration(selectionInterval); err != nil || duration <= 0 {
			fieldErrors = append(fieldErrors, "Ranked selection interval must be a positive duration")
		}
		selectionMovieLimit := parseNonNegativeInt("Ranked movie limit", "jobs.selection.movie_limit", cfg.Jobs.Selection.MovieLimit)
		selectionShowLimit := parseNonNegativeInt("Ranked show limit", "jobs.selection.show_limit", cfg.Jobs.Selection.ShowLimit)
		selectionMembers := map[string]bool{}
		selectionMembersSubmitted := c.FormValue("jobs.selection.members_present") == "true"
		if selectionMembersSubmitted {
			form, err := c.MultipartForm()
			if err != nil {
				fieldErrors = append(fieldErrors, "Ranked selection membership could not be read")
			}
			knownJobs := map[string]config.DynamicJob{}
			for _, job := range cfg.Jobs.List {
				knownJobs[job.ID] = job
			}
			var memberIDs []string
			if form != nil {
				memberIDs = form.Value["jobs.selection.members"]
			}
			for _, id := range memberIDs {
				job, ok := knownJobs[id]
				if !ok {
					fieldErrors = append(fieldErrors, "Ranked selection contains an unknown job")
					continue
				}
				if job.Type == "smart_popular" {
					fieldErrors = append(fieldErrors, "Adaptive smart jobs cannot join ranked selection")
					continue
				}
				selectionMembers[id] = true
			}
		}
		globalPeriod := c.FormValue("jobs.global_period")
		if globalPeriod == "" {
			globalPeriod = cfg.Jobs.GlobalPeriod
		}
		if globalPeriod == "" {
			globalPeriod = "daily"
		}
		if globalPeriod != "daily" && globalPeriod != "weekly" && globalPeriod != "monthly" {
			fieldErrors = append(fieldErrors, "Delivery limit period must be daily, weekly, or monthly")
		}
		repeatPolicy := enums.RepeatPolicy(c.FormValue("jobs.repeat_policy"))
		if repeatPolicy == enums.RepeatPolicyInherit {
			repeatPolicy = enums.RepeatPolicy(cfg.Jobs.RepeatPolicy)
		}
		if repeatPolicy == enums.RepeatPolicyInherit {
			repeatPolicy = enums.RepeatPolicy90Days
		}
		if !repeatPolicy.IsValid(false) {
			fieldErrors = append(fieldErrors, "Repeat handling policy is invalid")
		}
		scoringEnabled := c.FormValue("scoring.enabled") == "true"
		ratingWeight := parseFloat("Rating weight", "scoring.rating_weight", 0.6)
		popularityWeight := parseFloat("Popularity weight", "scoring.popularity_weight", 0.3)
		recencyWeight := parseFloat("Recency weight", "scoring.recency_weight", 0.1)
		ratingScale := parseFloat("Rating scale", "scoring.rating_scale", 10)
		recencyDays := parseNonNegativeInt("Recency window", "scoring.recency_days", 365)

		if scoringEnabled {
			total := ratingWeight + popularityWeight + recencyWeight
			if total < 0.999 || total > 1.001 {
				fieldErrors = append(fieldErrors, fmt.Sprintf("Scoring weights must total 1.0 (currently %.2f)", total))
			}
		}
		if selectionEnabled && !scoringEnabled {
			fieldErrors = append(fieldErrors, "Ranked selection requires content scoring")
		}
		if selectionEnabled {
			movieMinima, showMinima := 0, 0
			for _, job := range cfg.Jobs.List {
				participates := job.SelectionCycle
				if selectionMembersSubmitted {
					participates = selectionMembers[job.ID]
				}
				if !job.Enabled || !participates {
					continue
				}
				if job.MediaType == "show" {
					showMinima += job.MinimumPicks
				} else {
					movieMinima += job.MinimumPicks
				}
			}
			if (selectionMovieLimit > 0 && movieMinima > selectionMovieLimit) || (selectionShowLimit > 0 && showMinima > selectionShowLimit) {
				fieldErrors = append(fieldErrors, "Ranked selection limits cannot be lower than participating job minimums")
			}
		}

		if len(fieldErrors) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  strings.Join(fieldErrors, "; "),
				"fields": fieldErrors,
			})
		}

		// All values validated — safe to apply to the live config now.
		cfg.Trakt.ClientID = c.FormValue("trakt.client_id")
		cfg.Simkl.ClientID = c.FormValue("simkl.client_id")
		preserveBlankSecret(c, "trakt.client_secret", &cfg.Trakt.ClientSecret)
		preserveBlankSecret(c, "tmdb.api_key", &cfg.TMDB.APIKey)
		preserveBlankSecret(c, "mdblist.api_key", &cfg.MDBList.APIKey)
		cfg.Letterboxd.ExperimentalScraping = c.FormValue("letterboxd.experimental_scraping") == "true"
		cfg.Radarr.URL = c.FormValue("radarr.url")
		preserveBlankSecret(c, "radarr.api_key", &cfg.Radarr.APIKey)
		cfg.Radarr.RootFolder = c.FormValue("radarr.root_folder")
		cfg.Radarr.MinimumAvailability = c.FormValue("radarr.minimum_availability")
		cfg.Radarr.Monitor = c.FormValue("radarr.monitor")
		cfg.Radarr.QualityProfile = radarrQualityProfile
		cfg.Sonarr.URL = c.FormValue("sonarr.url")
		preserveBlankSecret(c, "sonarr.api_key", &cfg.Sonarr.APIKey)
		cfg.Sonarr.RootFolder = c.FormValue("sonarr.root_folder")
		cfg.Sonarr.Monitor = c.FormValue("sonarr.monitor")
		cfg.Sonarr.QualityProfile = sonarrQualityProfile
		cfg.Jellyseerr.URL = c.FormValue("jellyseerr.url")
		preserveBlankSecret(c, "jellyseerr.api_key", &cfg.Jellyseerr.APIKey)
		cfg.Jellyseerr.UserID = c.FormValue("jellyseerr.user_id")
		cfg.Jellyseerr.RequestCredentials.Email = c.FormValue("jellyseerr.request_credentials.email")
		preserveBlankSecret(c, "jellyseerr.request_credentials.password", &cfg.Jellyseerr.RequestCredentials.Password)
		if mode := c.FormValue("jobs.mode"); mode != "" {
			cfg.Jobs.Mode = mode
		}
		if syncInterval := c.FormValue("jobs.sync_interval"); syncInterval != "" {
			cfg.Jobs.SyncInterval = syncInterval
		}
		cfg.Jobs.GlobalLimitMovies = globalLimitMovies
		cfg.Jobs.GlobalLimitShows = globalLimitShows
		cfg.Jobs.GlobalPeriod = globalPeriod
		cfg.Jobs.RepeatPolicy = string(repeatPolicy)
		cfg.Jobs.Selection.Enabled = selectionEnabled
		cfg.Jobs.Selection.SyncInterval = selectionInterval
		cfg.Jobs.Selection.MovieLimit = selectionMovieLimit
		cfg.Jobs.Selection.ShowLimit = selectionShowLimit
		if selectionMembersSubmitted {
			for index := range cfg.Jobs.List {
				cfg.Jobs.List[index].SelectionCycle = selectionMembers[cfg.Jobs.List[index].ID]
				if !cfg.Jobs.List[index].SelectionCycle {
					cfg.Jobs.List[index].MinimumPicks = 0
				}
			}
		}
		cfg.Scoring.Enabled = scoringEnabled
		cfg.Scoring.RatingWeight = ratingWeight
		cfg.Scoring.PopularityWeight = popularityWeight
		cfg.Scoring.RecencyWeight = recencyWeight
		cfg.Scoring.RatingScale = ratingScale
		cfg.Scoring.RecencyDays = recencyDays
		if metric := c.FormValue("scoring.popularity_metric"); metric != "" {
			cfg.Scoring.PopularityMetric = metric
		} else if cfg.Scoring.PopularityMetric == "" {
			cfg.Scoring.PopularityMetric = "votes"
		}

		if cfg.ConfigFilePath == "" {
			cfg.ConfigFilePath = determineConfigPath()
		}
		if err := global.UpdateConfig(rg.gctx, func(candidate *config.Config) error {
			*candidate = *cfg
			return nil
		}); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save configuration: " + err.Error()})
		}

		return c.JSON(fiber.Map{"success": true, "message": "Configuration saved."})
	})

	// Jobs config save route
	app.Post("/jobs/config/save", func(c *fiber.Ctx) error {
		cfg := rg.gctx.Config()

		// Update global sync interval if provided
		if syncInterval := c.FormValue("jobs.sync_interval"); syncInterval != "" {
			cfg.Jobs.SyncInterval = syncInterval
		}

		// Update global mode if provided
		if mode := c.FormValue("jobs.mode"); mode != "" {
			cfg.Jobs.Mode = mode
		}

		// Helper to check if a field was submitted (using Fiber's FormValue which handles multipart)
		hasField := func(key string) bool {
			return c.FormValue(key) != ""
		}

		// Movies - Trending
		if hasField("jobs.trending_movies.enabled") || hasField("jobs.trending_movies.limit") {
			cfg.Jobs.TrendingMovies.Enabled = c.FormValue("jobs.trending_movies.enabled") == "on"
			if limit := c.FormValue("jobs.trending_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.TrendingMovies.Limit = val
				}
			}
			// Per-job sync interval
			if hasField("jobs.trending_movies.sync_interval") {
				cfg.Jobs.TrendingMovies.SyncInterval = c.FormValue("jobs.trending_movies.sync_interval")
			}
			// Always update mode and minimum_availability (empty string means use default)
			cfg.Jobs.TrendingMovies.Mode = c.FormValue("jobs.trending_movies.mode")
			cfg.Jobs.TrendingMovies.MinimumAvailability = c.FormValue("jobs.trending_movies.minimum_availability")
			cfg.Jobs.TrendingMovies.Monitor = c.FormValue("jobs.trending_movies.monitor")
		}

		// Movies - Popular
		if hasField("jobs.popular_movies.enabled") || hasField("jobs.popular_movies.limit") {
			cfg.Jobs.PopularMovies.Enabled = c.FormValue("jobs.popular_movies.enabled") == "on"
			if limit := c.FormValue("jobs.popular_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.PopularMovies.Limit = val
				}
			}
			// Per-job sync interval
			if hasField("jobs.popular_movies.sync_interval") {
				cfg.Jobs.PopularMovies.SyncInterval = c.FormValue("jobs.popular_movies.sync_interval")
			}
			// Always update mode and minimum_availability (empty string means use default)
			cfg.Jobs.PopularMovies.Mode = c.FormValue("jobs.popular_movies.mode")
			cfg.Jobs.PopularMovies.MinimumAvailability = c.FormValue("jobs.popular_movies.minimum_availability")
			cfg.Jobs.PopularMovies.Monitor = c.FormValue("jobs.popular_movies.monitor")
		}

		// Movies - Box Office
		if hasField("jobs.box_office.enabled") || hasField("jobs.box_office.limit") {
			cfg.Jobs.BoxOffice.Enabled = c.FormValue("jobs.box_office.enabled") == "on"
			if limit := c.FormValue("jobs.box_office.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.BoxOffice.Limit = val
				}
			}
			// Per-job sync interval
			if hasField("jobs.box_office.sync_interval") {
				cfg.Jobs.BoxOffice.SyncInterval = c.FormValue("jobs.box_office.sync_interval")
			}
			// Always update mode and minimum_availability (empty string means use default)
			cfg.Jobs.BoxOffice.Mode = c.FormValue("jobs.box_office.mode")
			cfg.Jobs.BoxOffice.MinimumAvailability = c.FormValue("jobs.box_office.minimum_availability")
			cfg.Jobs.BoxOffice.Monitor = c.FormValue("jobs.box_office.monitor")

		}

		// Movies - Favorited
		if hasField("jobs.favorited_movies.enabled") || hasField("jobs.favorited_movies.limit") || hasField("jobs.favorited_movies.period") {
			cfg.Jobs.FavoritedMovies.Enabled = c.FormValue("jobs.favorited_movies.enabled") == "on"
			if limit := c.FormValue("jobs.favorited_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.FavoritedMovies.Limit = val
				}
			}
			if period := c.FormValue("jobs.favorited_movies.period"); period != "" {
				cfg.Jobs.FavoritedMovies.Period = period
			}
			// Per-job sync interval
			if hasField("jobs.favorited_movies.sync_interval") {
				cfg.Jobs.FavoritedMovies.SyncInterval = c.FormValue("jobs.favorited_movies.sync_interval")
			}
			// Always update mode and minimum_availability (empty string means use default)
			cfg.Jobs.FavoritedMovies.Mode = c.FormValue("jobs.favorited_movies.mode")
			cfg.Jobs.FavoritedMovies.MinimumAvailability = c.FormValue("jobs.favorited_movies.minimum_availability")
			cfg.Jobs.FavoritedMovies.Monitor = c.FormValue("jobs.favorited_movies.monitor")
		}

		// Movies - Played
		if hasField("jobs.played_movies.enabled") || hasField("jobs.played_movies.limit") || hasField("jobs.played_movies.period") {
			cfg.Jobs.PlayedMovies.Enabled = c.FormValue("jobs.played_movies.enabled") == "on"
			if limit := c.FormValue("jobs.played_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.PlayedMovies.Limit = val
				}
			}
			if period := c.FormValue("jobs.played_movies.period"); period != "" {
				cfg.Jobs.PlayedMovies.Period = period
			}
			// Per-job sync interval
			if hasField("jobs.played_movies.sync_interval") {
				cfg.Jobs.PlayedMovies.SyncInterval = c.FormValue("jobs.played_movies.sync_interval")
			}
			// Always update mode and minimum_availability (empty string means use default)
			cfg.Jobs.PlayedMovies.Mode = c.FormValue("jobs.played_movies.mode")
			cfg.Jobs.PlayedMovies.MinimumAvailability = c.FormValue("jobs.played_movies.minimum_availability")
			cfg.Jobs.PlayedMovies.Monitor = c.FormValue("jobs.played_movies.monitor")
		}

		// Movies - Watched
		if hasField("jobs.watched_movies.enabled") || hasField("jobs.watched_movies.limit") || hasField("jobs.watched_movies.period") {
			cfg.Jobs.WatchedMovies.Enabled = c.FormValue("jobs.watched_movies.enabled") == "on"
			if limit := c.FormValue("jobs.watched_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.WatchedMovies.Limit = val
				}
			}
			if period := c.FormValue("jobs.watched_movies.period"); period != "" {
				cfg.Jobs.WatchedMovies.Period = period
			}
			// Per-job sync interval
			if hasField("jobs.watched_movies.sync_interval") {
				cfg.Jobs.WatchedMovies.SyncInterval = c.FormValue("jobs.watched_movies.sync_interval")
			}
			// Always update mode and minimum_availability (empty string means use default)
			cfg.Jobs.WatchedMovies.Mode = c.FormValue("jobs.watched_movies.mode")
			cfg.Jobs.WatchedMovies.MinimumAvailability = c.FormValue("jobs.watched_movies.minimum_availability")
			cfg.Jobs.WatchedMovies.Monitor = c.FormValue("jobs.watched_movies.monitor")
		}

		// Movies - Collected
		if hasField("jobs.collected_movies.enabled") || hasField("jobs.collected_movies.limit") || hasField("jobs.collected_movies.period") {
			cfg.Jobs.CollectedMovies.Enabled = c.FormValue("jobs.collected_movies.enabled") == "on"
			if limit := c.FormValue("jobs.collected_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.CollectedMovies.Limit = val
				}
			}
			if period := c.FormValue("jobs.collected_movies.period"); period != "" {
				cfg.Jobs.CollectedMovies.Period = period
			}
			// Per-job sync interval
			if hasField("jobs.collected_movies.sync_interval") {
				cfg.Jobs.CollectedMovies.SyncInterval = c.FormValue("jobs.collected_movies.sync_interval")
			}
			// Always update mode and minimum_availability (empty string means use default)
			cfg.Jobs.CollectedMovies.Mode = c.FormValue("jobs.collected_movies.mode")
			cfg.Jobs.CollectedMovies.MinimumAvailability = c.FormValue("jobs.collected_movies.minimum_availability")
			cfg.Jobs.CollectedMovies.Monitor = c.FormValue("jobs.collected_movies.monitor")
		}

		// Movies - Anticipated
		if hasField("jobs.anticipated_movies.enabled") || hasField("jobs.anticipated_movies.limit") {
			cfg.Jobs.AnticipatedMovies.Enabled = c.FormValue("jobs.anticipated_movies.enabled") == "on"
			if limit := c.FormValue("jobs.anticipated_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.AnticipatedMovies.Limit = val
				}
			}
			// Per-job sync interval
			if hasField("jobs.anticipated_movies.sync_interval") {
				cfg.Jobs.AnticipatedMovies.SyncInterval = c.FormValue("jobs.anticipated_movies.sync_interval")
			}
			// Always update mode and minimum_availability (empty string means use default)
			cfg.Jobs.AnticipatedMovies.Mode = c.FormValue("jobs.anticipated_movies.mode")
			cfg.Jobs.AnticipatedMovies.MinimumAvailability = c.FormValue("jobs.anticipated_movies.minimum_availability")
			cfg.Jobs.AnticipatedMovies.Monitor = c.FormValue("jobs.anticipated_movies.monitor")
		}

		// Shows - Trending
		if hasField("jobs.trending_shows.enabled") || hasField("jobs.trending_shows.limit") {
			cfg.Jobs.TrendingShows.Enabled = c.FormValue("jobs.trending_shows.enabled") == "on"
			if limit := c.FormValue("jobs.trending_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.TrendingShows.Limit = val
				}
			}
			// Per-job sync interval
			if hasField("jobs.trending_shows.sync_interval") {
				cfg.Jobs.TrendingShows.SyncInterval = c.FormValue("jobs.trending_shows.sync_interval")
			}
			// Always update mode and monitor (empty string means use default)
			cfg.Jobs.TrendingShows.Mode = c.FormValue("jobs.trending_shows.mode")
			cfg.Jobs.TrendingShows.Monitor = c.FormValue("jobs.trending_shows.monitor")
		}

		// Shows - Popular
		if hasField("jobs.popular_shows.enabled") || hasField("jobs.popular_shows.limit") {
			cfg.Jobs.PopularShows.Enabled = c.FormValue("jobs.popular_shows.enabled") == "on"
			if limit := c.FormValue("jobs.popular_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.PopularShows.Limit = val
				}
			}
			// Per-job sync interval
			if hasField("jobs.popular_shows.sync_interval") {
				cfg.Jobs.PopularShows.SyncInterval = c.FormValue("jobs.popular_shows.sync_interval")
			}
			// Always update mode and monitor (empty string means use default)
			cfg.Jobs.PopularShows.Mode = c.FormValue("jobs.popular_shows.mode")
			cfg.Jobs.PopularShows.Monitor = c.FormValue("jobs.popular_shows.monitor")
		}

		// Shows - Favorited
		if hasField("jobs.favorited_shows.enabled") || hasField("jobs.favorited_shows.limit") || hasField("jobs.favorited_shows.period") {
			cfg.Jobs.FavoritedShows.Enabled = c.FormValue("jobs.favorited_shows.enabled") == "on"
			if limit := c.FormValue("jobs.favorited_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.FavoritedShows.Limit = val
				}
			}
			if period := c.FormValue("jobs.favorited_shows.period"); period != "" {
				cfg.Jobs.FavoritedShows.Period = period
			}
			// Per-job sync interval
			if hasField("jobs.favorited_shows.sync_interval") {
				cfg.Jobs.FavoritedShows.SyncInterval = c.FormValue("jobs.favorited_shows.sync_interval")
			}
			// Always update mode and monitor (empty string means use default)
			cfg.Jobs.FavoritedShows.Mode = c.FormValue("jobs.favorited_shows.mode")
			cfg.Jobs.FavoritedShows.Monitor = c.FormValue("jobs.favorited_shows.monitor")
		}

		// Shows - Played
		if hasField("jobs.played_shows.enabled") || hasField("jobs.played_shows.limit") || hasField("jobs.played_shows.period") {
			cfg.Jobs.PlayedShows.Enabled = c.FormValue("jobs.played_shows.enabled") == "on"
			if limit := c.FormValue("jobs.played_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.PlayedShows.Limit = val
				}
			}
			if period := c.FormValue("jobs.played_shows.period"); period != "" {
				cfg.Jobs.PlayedShows.Period = period
			}
			// Per-job sync interval
			if hasField("jobs.played_shows.sync_interval") {
				cfg.Jobs.PlayedShows.SyncInterval = c.FormValue("jobs.played_shows.sync_interval")
			}
			// Always update mode and monitor (empty string means use default)
			cfg.Jobs.PlayedShows.Mode = c.FormValue("jobs.played_shows.mode")
			cfg.Jobs.PlayedShows.Monitor = c.FormValue("jobs.played_shows.monitor")
		}

		// Shows - Watched
		if hasField("jobs.watched_shows.enabled") || hasField("jobs.watched_shows.limit") || hasField("jobs.watched_shows.period") {
			cfg.Jobs.WatchedShows.Enabled = c.FormValue("jobs.watched_shows.enabled") == "on"
			if limit := c.FormValue("jobs.watched_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.WatchedShows.Limit = val
				}
			}
			if period := c.FormValue("jobs.watched_shows.period"); period != "" {
				cfg.Jobs.WatchedShows.Period = period
			}
			// Per-job sync interval
			if hasField("jobs.watched_shows.sync_interval") {
				cfg.Jobs.WatchedShows.SyncInterval = c.FormValue("jobs.watched_shows.sync_interval")
			}
			// Always update mode and monitor (empty string means use default)
			cfg.Jobs.WatchedShows.Mode = c.FormValue("jobs.watched_shows.mode")
			cfg.Jobs.WatchedShows.Monitor = c.FormValue("jobs.watched_shows.monitor")
		}

		// Shows - Collected
		if hasField("jobs.collected_shows.enabled") || hasField("jobs.collected_shows.limit") || hasField("jobs.collected_shows.period") {
			cfg.Jobs.CollectedShows.Enabled = c.FormValue("jobs.collected_shows.enabled") == "on"
			if limit := c.FormValue("jobs.collected_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.CollectedShows.Limit = val
				}
			}
			if period := c.FormValue("jobs.collected_shows.period"); period != "" {
				cfg.Jobs.CollectedShows.Period = period
			}
			// Per-job sync interval
			if hasField("jobs.collected_shows.sync_interval") {
				cfg.Jobs.CollectedShows.SyncInterval = c.FormValue("jobs.collected_shows.sync_interval")
			}
			// Always update mode and monitor (empty string means use default)
			cfg.Jobs.CollectedShows.Mode = c.FormValue("jobs.collected_shows.mode")
			cfg.Jobs.CollectedShows.Monitor = c.FormValue("jobs.collected_shows.monitor")
		}

		// Shows - Anticipated
		if hasField("jobs.anticipated_shows.enabled") || hasField("jobs.anticipated_shows.limit") {
			cfg.Jobs.AnticipatedShows.Enabled = c.FormValue("jobs.anticipated_shows.enabled") == "on"
			if limit := c.FormValue("jobs.anticipated_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.AnticipatedShows.Limit = val
				}
			}
			// Per-job sync interval
			if hasField("jobs.anticipated_shows.sync_interval") {
				cfg.Jobs.AnticipatedShows.SyncInterval = c.FormValue("jobs.anticipated_shows.sync_interval")
			}
			// Always update mode and monitor (empty string means use default)
			cfg.Jobs.AnticipatedShows.Mode = c.FormValue("jobs.anticipated_shows.mode")
			cfg.Jobs.AnticipatedShows.Monitor = c.FormValue("jobs.anticipated_shows.monitor")
		}

		// Movies - Smart Popular
		if hasField("jobs.smart_popular_movies.enabled") || hasField("jobs.smart_popular_movies.limit") || hasField("jobs.smart_popular_movies.base_min_rating") || hasField("jobs.smart_popular_movies.adjustment_factor") {
			cfg.Jobs.SmartPopularMovies.Enabled = c.FormValue("jobs.smart_popular_movies.enabled") == "on"
			if limit := c.FormValue("jobs.smart_popular_movies.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.SmartPopularMovies.Limit = val
				}
			}
			if baseRating := c.FormValue("jobs.smart_popular_movies.base_min_rating"); baseRating != "" {
				if val, err := strconv.ParseFloat(baseRating, 64); err == nil {
					cfg.Jobs.SmartPopularMovies.BaseMinRating = val
				}
			}
			if adjustmentFactor := c.FormValue("jobs.smart_popular_movies.adjustment_factor"); adjustmentFactor != "" {
				if val, err := strconv.ParseFloat(adjustmentFactor, 64); err == nil {
					cfg.Jobs.SmartPopularMovies.AdjustmentFactor = val
				}
			}
			// Per-job sync interval
			if hasField("jobs.smart_popular_movies.sync_interval") {
				cfg.Jobs.SmartPopularMovies.SyncInterval = c.FormValue("jobs.smart_popular_movies.sync_interval")
			}
			// Always update mode and minimum_availability (empty string means use default)
			cfg.Jobs.SmartPopularMovies.Mode = c.FormValue("jobs.smart_popular_movies.mode")
			cfg.Jobs.SmartPopularMovies.MinimumAvailability = c.FormValue("jobs.smart_popular_movies.minimum_availability")
			cfg.Jobs.SmartPopularMovies.Monitor = c.FormValue("jobs.smart_popular_movies.monitor")
		}

		// Shows - Smart Popular
		if hasField("jobs.smart_popular_shows.enabled") || hasField("jobs.smart_popular_shows.limit") || hasField("jobs.smart_popular_shows.base_min_rating") || hasField("jobs.smart_popular_shows.adjustment_factor") {
			cfg.Jobs.SmartPopularShows.Enabled = c.FormValue("jobs.smart_popular_shows.enabled") == "on"
			if limit := c.FormValue("jobs.smart_popular_shows.limit"); limit != "" {
				if val, err := strconv.Atoi(limit); err == nil {
					cfg.Jobs.SmartPopularShows.Limit = val
				}
			}
			if baseRating := c.FormValue("jobs.smart_popular_shows.base_min_rating"); baseRating != "" {
				if val, err := strconv.ParseFloat(baseRating, 64); err == nil {
					cfg.Jobs.SmartPopularShows.BaseMinRating = val
				}
			}
			if adjustmentFactor := c.FormValue("jobs.smart_popular_shows.adjustment_factor"); adjustmentFactor != "" {
				if val, err := strconv.ParseFloat(adjustmentFactor, 64); err == nil {
					cfg.Jobs.SmartPopularShows.AdjustmentFactor = val
				}
			}
			// Per-job sync interval
			if hasField("jobs.smart_popular_shows.sync_interval") {
				cfg.Jobs.SmartPopularShows.SyncInterval = c.FormValue("jobs.smart_popular_shows.sync_interval")
			}
			// Always update mode and monitor (empty string means use default)
			cfg.Jobs.SmartPopularShows.Mode = c.FormValue("jobs.smart_popular_shows.mode")
			cfg.Jobs.SmartPopularShows.Monitor = c.FormValue("jobs.smart_popular_shows.monitor")
		}

		if cfg.ConfigFilePath == "" {
			cfg.ConfigFilePath = determineConfigPath()
		}
		if err := global.UpdateConfig(rg.gctx, func(candidate *config.Config) error {
			*candidate = *cfg
			return nil
		}); err != nil {
			return c.Status(500).SendString(`
				<script>showNotification('Failed to save job configuration: ` + err.Error() + `', 'error');</script>
			`)
		}

		return c.SendString(`
			<script>
				showNotification('Job configuration saved and reloaded successfully!', 'success');
				setTimeout(function() { window.location.reload(); }, 500);
			</script>
		`)
	})

	// Restart application route
	app.Post("/config/restart", func(c *fiber.Ctx) error {
		// Reload the configuration from file
		if err := rg.gctx.ReloadConfig(); err != nil {
			return c.Status(500).SendString(`
				<div class="bg-red-900 border border-red-600 text-red-200 p-4 rounded-lg">
					Failed to reload configuration: ` + err.Error() + `
				</div>
			`)
		}

		return c.SendString(`
			<div class="bg-green-900 border border-green-600 text-green-200 p-4 rounded-lg">
				Configuration reloaded successfully! Changes are now active.
			</div>
		`)
	})

	// Filters page route
	app.Get("/filters", func(c *fiber.Ctx) error {
		return c.Render("filters", fiber.Map{
			"Title":   "Blockbusterr - Filters",
			"Config":  rg.gctx.Config(),
			"Version": rg.gctx.Metadata().Version,
		}, "base")
	})

}
