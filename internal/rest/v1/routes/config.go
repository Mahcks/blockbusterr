package routes

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/global"
	"gopkg.in/yaml.v3"
)

// ShareableConfig contains only the shareable parts of the config (no credentials)
type ShareableConfig struct {
	Scoring any `yaml:"scoring"`
	Jobs    any `yaml:"jobs"`
	Filters struct {
		Movies config.MovieFilters `yaml:"movies"`
		Shows  config.ShowFilters  `yaml:"shows"`
	} `yaml:"filters"`
}

func RegisterConfigRoutes(router fiber.Router, gctx global.Context) {
	// Export shareable configuration (filters, jobs, scoring only)
	router.Get("/config/export", func(c *fiber.Ctx) error {
		cfg := gctx.Config()

		// Create shareable config with only filters, jobs, and scoring
		shareableData := map[string]any{
			"scoring": cfg.Scoring,
			"jobs":    cfg.Jobs,
			"filters": cfg.Filters,
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

	// Backup full configuration (including credentials)
	router.Get("/config/backup", func(c *fiber.Ctx) error {
		cfg := gctx.Config()

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

		return c.Send(data)
	})

	// Import shareable configuration (filters, jobs, scoring only)
	router.Post("/config/import", func(c *fiber.Ctx) error {
		// Get uploaded file
		file, err := c.FormFile("config")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No config file provided",
			})
		}

		// Open the uploaded file
		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to open uploaded file: %v", err),
			})
		}
		defer src.Close()

		// Read file contents
		data, err := io.ReadAll(src)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to read file: %v", err),
			})
		}

		// Parse YAML as shareable config
		var shareableConfig ShareableConfig
		if err := yaml.Unmarshal(data, &shareableConfig); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid YAML format: %v", err),
			})
		}

		// Get current config and merge shareable parts
		cfg := gctx.Config()

		// Update only the shareable sections
		if err := yaml.Unmarshal(data, &struct {
			Scoring any `yaml:"scoring"`
			Jobs    any `yaml:"jobs"`
			Filters struct {
				Movies config.MovieFilters `yaml:"movies"`
				Shows  config.ShowFilters  `yaml:"shows"`
			} `yaml:"filters"`
		}{
			Scoring: &cfg.Scoring,
			Jobs:    &cfg.Jobs,
			Filters: struct {
				Movies config.MovieFilters `yaml:"movies"`
				Shows  config.ShowFilters  `yaml:"shows"`
			}{
				Movies: cfg.Filters.Movies,
				Shows:  cfg.Filters.Shows,
			},
		}); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to parse config: %v", err),
			})
		}

		// Save the updated config
		if err := cfg.Save(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to save config: %v", err),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Shareable configuration imported successfully. Please restart the application for changes to take effect.",
			"path":    cfg.ConfigFilePath,
		})
	})

	// Restore full configuration backup
	router.Post("/config/restore", func(c *fiber.Ctx) error {
		// Get uploaded file
		file, err := c.FormFile("config")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No config file provided",
			})
		}

		// Open the uploaded file
		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to open uploaded file: %v", err),
			})
		}
		defer src.Close()

		// Read file contents
		data, err := io.ReadAll(src)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to read file: %v", err),
			})
		}

		// Parse YAML to validate it
		var newConfig config.Config
		if err := yaml.Unmarshal(data, &newConfig); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid YAML format: %v", err),
			})
		}

		// Get current config file path
		cfg := gctx.Config()
		configPath := cfg.ConfigFilePath

		if configPath == "" {
			// If no config path is set, try to create one in the default location
			configPath = "./config/config.yaml"

			// Ensure directory exists
			if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to create config directory: %v", err),
				})
			}
		}

		// Create backup of existing config if it exists
		if _, err := os.Stat(configPath); err == nil {
			backupPath := fmt.Sprintf("%s.backup-%s", configPath, time.Now().Format("2006-01-02-150405"))
			if err := os.Rename(configPath, backupPath); err != nil {
				// Just log warning, don't fail
				fmt.Printf("Warning: Failed to create backup: %v\n", err)
			}
		}

		// Write new config to file
		if err := os.WriteFile(configPath, data, 0o644); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to write config file: %v", err),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Full configuration restored successfully. Please restart the application for changes to take effect.",
			"path":    configPath,
		})
	})
}
