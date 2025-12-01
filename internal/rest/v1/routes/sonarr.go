package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

// RegisterSonarrRoutes registers Sonarr-related API routes
func RegisterSonarrRoutes(rg *RouteGroup, group fiber.Router) {
	sonarr := group.Group("/sonarr")

	// GET /v1/sonarr/validate - Test connection
	sonarr.Get("/validate", rg.ValidateSonarr)

	// GET /v1/sonarr/series - List all series
	sonarr.Get("/series", rg.GetSonarrSeries)

	// GET /v1/sonarr/quality-profiles - Get quality profiles
	sonarr.Get("/quality-profiles", rg.GetSonarrQualityProfiles)

	// GET /v1/sonarr/root-folders - Get root folders
	sonarr.Get("/root-folders", rg.GetSonarrRootFolders)

	// GET /v1/sonarr/lookup?term=breaking+bad - Lookup series
	sonarr.Get("/lookup", rg.LookupSonarrSeries)

	// POST /v1/sonarr/series - Add a series
	sonarr.Post("/series", rg.AddSonarrSeries)
}

// ValidateSonarr validates the Sonarr API connection
func (rg *RouteGroup) ValidateSonarr(c *fiber.Ctx) error {
	// Check if we're in test mode (testing form values)
	testMode := c.Query("test") == "true"

	var url, apiKey string

	if testMode {
		// In test mode, require query parameters
		url = c.Query("url")
		apiKey = c.Query("api_key")
	} else {
		// Normal mode, use saved config
		cfg := rg.gctx.Config()
		url = cfg.Sonarr.URL
		apiKey = cfg.Sonarr.APIKey
	}

	if url == "" || apiKey == "" {
		return c.Status(400).JSON(fiber.Map{
			"error":     "Sonarr URL or API key is not configured",
			"connected": false,
		})
	}

	sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
		BaseURL: url,
		APIKey:  apiKey,
	})

	status, err := sonarrClient.GetSystemStatus(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":     err.Error(),
			"connected": false,
		})
	}

	return c.JSON(fiber.Map{
		"message":   "Sonarr API connection successful",
		"connected": true,
		"version":   status.Version,
	})
}

// GetSonarrSeries returns all series in Sonarr
func (rg *RouteGroup) GetSonarrSeries(c *fiber.Ctx) error {
	cfg := rg.gctx.Config()

	sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
		BaseURL: cfg.Sonarr.URL,
		APIKey:  cfg.Sonarr.APIKey,
	})

	series, err := sonarrClient.GetSeries(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  series,
		"count": len(series),
	})
}

// GetSonarrQualityProfiles returns quality profiles
func (rg *RouteGroup) GetSonarrQualityProfiles(c *fiber.Ctx) error {
	cfg := rg.gctx.Config()

	sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
		BaseURL: cfg.Sonarr.URL,
		APIKey:  cfg.Sonarr.APIKey,
	})

	profiles, err := sonarrClient.GetQualityProfiles(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  profiles,
		"count": len(profiles),
	})
}

// GetSonarrRootFolders returns root folders
func (rg *RouteGroup) GetSonarrRootFolders(c *fiber.Ctx) error {
	cfg := rg.gctx.Config()

	sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
		BaseURL: cfg.Sonarr.URL,
		APIKey:  cfg.Sonarr.APIKey,
	})

	folders, err := sonarrClient.GetRootFolders(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  folders,
		"count": len(folders),
	})
}

// LookupSonarrSeries looks up a series by term
func (rg *RouteGroup) LookupSonarrSeries(c *fiber.Ctx) error {
	term := c.Query("term")
	if term == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "term parameter is required",
		})
	}

	cfg := rg.gctx.Config()

	sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
		BaseURL: cfg.Sonarr.URL,
		APIKey:  cfg.Sonarr.APIKey,
	})

	series, err := sonarrClient.LookupSeries(c.Context(), term)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":  series,
		"count": len(series),
	})
}

// AddSonarrSeries adds a series to Sonarr
func (rg *RouteGroup) AddSonarrSeries(c *fiber.Ctx) error {
	var series integrations.SonarrSeries
	if err := c.BodyParser(&series); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	cfg := rg.gctx.Config()

	sonarrClient := integrations.NewSonarr(integrations.SonarrConfig{
		BaseURL: cfg.Sonarr.URL,
		APIKey:  cfg.Sonarr.APIKey,
	})

	addedSeries, err := sonarrClient.AddSeries(c.Context(), series)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"data":    addedSeries,
		"message": "Series added successfully",
	})
}
