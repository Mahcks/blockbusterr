package routes

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/pkg/enums"
	"gopkg.in/yaml.v3"
)

const (
	shareableConfigVersion = 2
	maxConfigUploadSize    = 2 << 20
)

// ShareableConfig contains only portable automation settings, never credentials.
type ShareableConfig struct {
	Version int `yaml:"schema_version"`
	Scoring any `yaml:"scoring"`
	Jobs    any `yaml:"jobs"`
	Filters struct {
		Movies config.MovieFilters `yaml:"movies"`
		Shows  config.ShowFilters  `yaml:"shows"`
	} `yaml:"filters"`
	RuleSets        []config.RuleSet       `yaml:"rule_sets,omitempty"`
	TitleExceptions config.TitleExceptions `yaml:"title_exceptions,omitempty"`
}

type jobBundle struct {
	Version         int                    `yaml:"schema_version"`
	Job             config.DynamicJob      `yaml:"job"`
	RuleSet         config.RuleSet         `yaml:"rule_set"`
	TitleExceptions config.TitleExceptions `yaml:"title_exceptions,omitempty"`
}

type importedJobResponse struct {
	config.DynamicJob
	RuleSetName    string `json:"rule_set_name"`
	IDsRegenerated bool   `json:"ids_regenerated"`
}

func RegisterConfigRoutes(router fiber.Router, gctx global.Context) {
	// Export portable automation without integration credentials.
	router.Get("/config/export", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		shareableData := map[string]any{
			"schema_version":   shareableConfigVersion,
			"scoring":          cfg.Scoring,
			"jobs":             cfg.Jobs,
			"filters":          cfg.Filters,
			"rule_sets":        cfg.RuleSets,
			"title_exceptions": cfg.TitleExceptions,
		}

		// Marshal shareable config to YAML
		data, err := yaml.Marshal(shareableData)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to marshal config: %v", err),
			})
		}

		// Set headers for file download
		c.Set("Content-Type", "application/x-yaml")
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=blockbusterr-shareable-%s.yaml", time.Now().Format("2006-01-02")))

		return c.Send(data)
	})

	router.Get("/config/jobs/:id/export", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		job := cfg.GetDynamicJobByID(c.Params("id"))
		if job == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Job not found"})
		}
		rules, _, err := cfg.ResolveRuleSet(*job)
		if err != nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		data, err := yaml.Marshal(jobBundle{Version: shareableConfigVersion, Job: *job, RuleSet: *rules, TitleExceptions: cfg.TitleExceptions})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to marshal job: %v", err)})
		}
		c.Set("Content-Type", "application/x-yaml")
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=blockbusterr-job-%s.yaml", time.Now().Format("2006-01-02")))
		return c.Send(data)
	})

	// Backup full configuration (including credentials)
	router.Get("/config/backup", func(c *fiber.Ctx) error {
		cfg := gctx.Config()
		if cfg.Version == "" {
			cfg.Version = gctx.Metadata().Version
		}

		// Marshal full config to YAML
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to marshal config: %v", err),
			})
		}

		// Set headers for file download
		c.Set("Content-Type", "application/x-yaml")
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=blockbusterr-backup-%s.yaml", time.Now().Format("2006-01-02")))
		c.Set(fiber.HeaderCacheControl, "no-store, private")
		c.Set(fiber.HeaderPragma, "no-cache")
		c.Set(fiber.HeaderExpires, "0")

		return c.Send(data)
	})

	// Import a complete portable configuration while preserving credentials.
	router.Post("/config/import", func(c *fiber.Ctx) error {
		data, err := uploadedConfig(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		var candidate *config.Config
		var validationErr error
		if err := global.UpdateConfig(gctx, func(current *config.Config) error {
			candidate, validationErr = importedShareableConfig(current, data)
			if validationErr != nil {
				return validationErr
			}
			*current = *candidate
			return nil
		}); err != nil {
			if validationErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": validationErr.Error()})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to save config: %v", err),
			})
		}
		return c.JSON(fiber.Map{
			"message": "Shareable configuration imported successfully.",
			"path":    candidate.ConfigFilePath,
		})
	})

	router.Post("/config/jobs/import", func(c *fiber.Ctx) error {
		data, err := uploadedConfig(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		var candidate *config.Config
		var importedJob config.DynamicJob
		var validationErr error
		if err := global.UpdateConfig(gctx, func(current *config.Config) error {
			candidate, importedJob, validationErr = importedJobBundle(current, data)
			if validationErr != nil {
				return validationErr
			}
			*current = *candidate
			return nil
		}); err != nil {
			if validationErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": validationErr.Error()})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to save config: %v", err)})
		}
		rules, _ := candidate.RuleSetByID(importedJob.RuleSetID)
		ruleSetName := "Imported Rules"
		if rules != nil {
			ruleSetName = rules.Name
		}
		return c.Status(fiber.StatusCreated).JSON(importedJobResponse{DynamicJob: importedJob, RuleSetName: ruleSetName, IDsRegenerated: true})
	})

	// Restore full configuration backup
	router.Post("/config/restore", func(c *fiber.Ctx) error {
		data, err := uploadedConfig(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		cfg := gctx.Config()
		configPath := cfg.ConfigFilePath
		if configPath == "" {
			configPath = "./config/config.yaml"
			if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to create config directory: %v", err),
				})
			}
		}
		candidate, err := importedFullConfig(data, configPath)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		if err := global.RestoreConfig(gctx, candidate); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to restore configuration: %v", err),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Configuration restored successfully.",
			"path":    configPath,
		})
	})
}

func uploadedConfig(c *fiber.Ctx) ([]byte, error) {
	file, err := c.FormFile("config")
	if err != nil {
		return nil, fmt.Errorf("no config file provided")
	}
	if file.Size > maxConfigUploadSize {
		return nil, fmt.Errorf("configuration file exceeds the %d MiB limit", maxConfigUploadSize>>20)
	}
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer func() { _ = src.Close() }()
	data, err := io.ReadAll(io.LimitReader(src, maxConfigUploadSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}
	if len(data) > maxConfigUploadSize {
		return nil, fmt.Errorf("configuration file exceeds the %d MiB limit", maxConfigUploadSize>>20)
	}
	return data, nil
}

func importedFullConfig(data []byte, path string) (*config.Config, error) {
	var document map[string]yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("invalid YAML format: %w", err)
	}
	if len(document) == 0 || document["version"].Kind == 0 || document["jobs"].Kind == 0 {
		return nil, fmt.Errorf("file is not a Blockbusterr configuration backup")
	}
	var header struct {
		Version string `yaml:"version"`
	}
	if err := yaml.Unmarshal(data, &header); err != nil || !supportedRestoreVersion(header.Version) {
		return nil, fmt.Errorf("unsupported configuration version %q", header.Version)
	}
	candidate, err := config.Parse(data, path)
	if err != nil {
		return nil, err
	}
	if err := validatePortableAutomation(candidate); err != nil {
		return nil, err
	}
	return candidate, nil
}

func supportedRestoreVersion(version string) bool {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	if version == "dev" {
		return true
	}
	major, err := strconv.Atoi(strings.SplitN(version, ".", 2)[0])
	return err == nil && (major == 1 || major == 2)
}

func importedShareableConfig(current *config.Config, data []byte) (*config.Config, error) {
	var header struct {
		Version int `yaml:"schema_version"`
	}
	if err := yaml.Unmarshal(data, &header); err != nil {
		return nil, fmt.Errorf("invalid YAML format: %w", err)
	}
	if header.Version > shareableConfigVersion {
		return nil, fmt.Errorf("shareable configuration version %d is newer than supported version %d", header.Version, shareableConfigVersion)
	}
	var imported config.Config
	if err := yaml.Unmarshal(data, &imported); err != nil {
		return nil, fmt.Errorf("invalid YAML format: %w", err)
	}
	candidate := *current
	candidate.Scoring = imported.Scoring
	candidate.Jobs = imported.Jobs
	candidate.Filters = imported.Filters
	candidate.RuleSets = imported.RuleSets
	candidate.TitleExceptions = imported.TitleExceptions
	candidate.MigrateRuleSets()
	if err := validatePortableAutomation(&candidate); err != nil {
		return nil, err
	}
	return &candidate, nil
}

func importedJobBundle(current *config.Config, data []byte) (*config.Config, config.DynamicJob, error) {
	var bundle jobBundle
	if err := yaml.Unmarshal(data, &bundle); err != nil {
		return nil, config.DynamicJob{}, fmt.Errorf("invalid job bundle: %w", err)
	}
	if bundle.Version != shareableConfigVersion {
		return nil, config.DynamicJob{}, fmt.Errorf("unsupported job bundle version %d", bundle.Version)
	}
	candidate := *current
	rules := bundle.RuleSet
	rules.ID = uuid.NewString()
	rules.Revision = 1
	rules.Name = availableRuleSetName(&candidate, strings.TrimSpace(rules.Name))
	if err := candidate.ValidateRuleSet(rules, ""); err != nil {
		return nil, config.DynamicJob{}, fmt.Errorf("invalid imported rule set: %w", err)
	}
	job := bundle.Job
	job.ID = uuid.NewString()
	job.Enabled = false
	job.RuleSetID = rules.ID
	candidate.RuleSets = append(candidate.RuleSets, rules)
	candidate.Jobs.List = append(candidate.Jobs.List, job)
	candidate.TitleExceptions = mergeTitleExceptions(candidate.TitleExceptions, bundle.TitleExceptions)
	if err := validatePortableAutomation(&candidate); err != nil {
		return nil, config.DynamicJob{}, err
	}
	return &candidate, job, nil
}

func validatePortableAutomation(candidate *config.Config) error {
	if candidate.Jobs.GlobalLimitMovies < 0 || candidate.Jobs.GlobalLimitShows < 0 {
		return fmt.Errorf("global delivery limits cannot be negative")
	}
	if candidate.Jobs.Mode != "" && candidate.Jobs.Mode != "direct" && candidate.Jobs.Mode != "jellyseerr" {
		return fmt.Errorf("default delivery mode must be direct or jellyseerr")
	}
	if candidate.Jobs.RepeatPolicy == "" {
		candidate.Jobs.RepeatPolicy = string(enums.RepeatPolicy90Days)
	}
	if !enums.RepeatPolicy(candidate.Jobs.RepeatPolicy).IsValid(false) {
		return fmt.Errorf("repeat handling policy is invalid")
	}
	for _, job := range candidate.Jobs.List {
		if !enums.RepeatPolicy(job.RepeatPolicy).IsValid(true) {
			return fmt.Errorf("job %q has an invalid repeat handling policy", job.Name)
		}
	}
	if candidate.Jobs.GlobalPeriod == "" || candidate.Jobs.GlobalPeriod == "sync" {
		candidate.Jobs.GlobalPeriod = "daily"
	}
	if candidate.Jobs.GlobalPeriod != "daily" && candidate.Jobs.GlobalPeriod != "weekly" && candidate.Jobs.GlobalPeriod != "monthly" {
		return fmt.Errorf("global delivery period must be daily, weekly, or monthly")
	}
	if candidate.Scoring.Enabled {
		if candidate.Scoring.RatingScale <= 0 || candidate.Scoring.RecencyDays < 0 || !slices.Contains([]string{"votes", "views", "watchers"}, candidate.Scoring.PopularityMetric) {
			return fmt.Errorf("scoring normalization settings are invalid")
		}
		for _, weight := range []float64{candidate.Scoring.RatingWeight, candidate.Scoring.PopularityWeight, candidate.Scoring.RecencyWeight} {
			if weight < 0 || weight > 1 {
				return fmt.Errorf("scoring weights must be between 0 and 1")
			}
		}
		total := candidate.Scoring.RatingWeight + candidate.Scoring.PopularityWeight + candidate.Scoring.RecencyWeight
		if total < 0.999 || total > 1.001 {
			return fmt.Errorf("scoring weights must total 1.0")
		}
	}
	if candidate.Jobs.Selection.Enabled {
		if !candidate.Scoring.Enabled {
			return fmt.Errorf("ranked selection requires content scoring")
		}
		if candidate.Jobs.Selection.MovieLimit < 0 || candidate.Jobs.Selection.ShowLimit < 0 {
			return fmt.Errorf("ranked selection limits are invalid")
		}
	}
	if err := validateAutomationInterval("default job", candidate.Jobs.SyncInterval); err != nil {
		return err
	}
	if err := validateAutomationInterval("ranked selection", candidate.Jobs.Selection.SyncInterval); err != nil {
		return err
	}
	ruleIDs := make(map[string]bool, len(candidate.RuleSets))
	for _, rules := range candidate.RuleSets {
		if ruleIDs[rules.ID] {
			return fmt.Errorf("duplicate rule set ID %q", rules.ID)
		}
		ruleIDs[rules.ID] = true
		if err := candidate.ValidateRuleSet(rules, rules.ID); err != nil {
			return fmt.Errorf("invalid rule set %q: %w", rules.Name, err)
		}
	}
	providerReady := *candidate
	providerReady.Trakt.ClientID = "portable-validation"
	providerReady.TMDB.APIKey = "portable-validation"
	providerReady.Simkl.ClientID = "portable-validation"
	providerReady.MDBList.APIKey = "portable-validation"
	providerReady.Letterboxd.ExperimentalScraping = true
	providerReady.Trakt.AccessToken = "portable-validation"
	providerReady.TMDB.SessionID = "portable-validation"
	providerReady.TMDB.AccountID = 1
	jobIDs := make(map[string]bool, len(candidate.Jobs.List))
	for _, job := range candidate.Jobs.List {
		if job.ID == "" || jobIDs[job.ID] {
			return fmt.Errorf("job IDs must be present and unique")
		}
		jobIDs[job.ID] = true
		if job.Mode != "" && job.Mode != "direct" && job.Mode != "jellyseerr" {
			return fmt.Errorf("job %q has an invalid delivery mode", job.Name)
		}
		if err := validateAutomationInterval(fmt.Sprintf("job %q", job.Name), job.SyncInterval); err != nil {
			return err
		}
		if err := validateDynamicJob(&providerReady, job); err != nil {
			return fmt.Errorf("invalid job %q: %w", job.Name, err)
		}
		if _, _, err := candidate.ResolveRuleSet(job); err != nil {
			return fmt.Errorf("invalid job %q: %w", job.Name, err)
		}
	}
	for _, ids := range [][]int{
		candidate.TitleExceptions.AllowedMovieTMDBIDs,
		candidate.TitleExceptions.BlockedMovieTMDBIDs,
		candidate.TitleExceptions.AllowedShowTVDBIDs,
		candidate.TitleExceptions.BlockedShowTVDBIDs,
	} {
		for _, id := range ids {
			if id <= 0 {
				return fmt.Errorf("title exception IDs must be positive")
			}
		}
	}
	return nil
}

func validateAutomationInterval(label, interval string) error {
	if interval == "" {
		return nil
	}
	if _, _, err := config.ParseSyncInterval(interval); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	return nil
}

func availableRuleSetName(cfg *config.Config, name string) string {
	if name == "" {
		name = "Imported Rules"
	}
	base := name
	for suffix := 2; ; suffix++ {
		duplicate := false
		for _, rules := range cfg.RuleSets {
			if rules.Media == "movie" || rules.Media == "show" {
				duplicate = duplicate || strings.EqualFold(rules.Name, name)
			}
		}
		if !duplicate {
			return name
		}
		name = fmt.Sprintf("%s %d", base, suffix)
	}
}

func mergeTitleExceptions(current, imported config.TitleExceptions) config.TitleExceptions {
	merge := func(left, right []int) []int {
		seen := make(map[int]bool, len(left)+len(right))
		result := append([]int(nil), left...)
		for _, id := range left {
			seen[id] = true
		}
		for _, id := range right {
			if id > 0 && !seen[id] {
				result, seen[id] = append(result, id), true
			}
		}
		return result
	}
	current.AllowedMovieTMDBIDs = merge(current.AllowedMovieTMDBIDs, imported.AllowedMovieTMDBIDs)
	current.BlockedMovieTMDBIDs = merge(current.BlockedMovieTMDBIDs, imported.BlockedMovieTMDBIDs)
	current.AllowedShowTVDBIDs = merge(current.AllowedShowTVDBIDs, imported.AllowedShowTVDBIDs)
	current.BlockedShowTVDBIDs = merge(current.BlockedShowTVDBIDs, imported.BlockedShowTVDBIDs)
	return current
}
