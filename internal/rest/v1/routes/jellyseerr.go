package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

func (rg *RouteGroup) RegisterJellyseerrRoutes(group fiber.Router) {
	// Validate Jellyseerr connection
	validate := func(c *fiber.Ctx) error {
		cfg := rg.gctx.Config()
		url, apiKey, err := connectionCredentials(c, cfg.Jellyseerr.URL, cfg.Jellyseerr.APIKey)
		if err != nil || url == "" || apiKey == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Server URL and API key are required; enter a key when changing servers", "connected": false})
		}

		jellyseerrClient := integrations.NewJellyseerr(integrations.JellyseerrConfig{
			URL:    url,
			APIKey: apiKey,
		})

		// Test connection by getting status
		status, err := jellyseerrClient.GetStatus()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"connected": false,
				"error":     err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"connected": true,
			"message":   "Jellyseerr connection successful! Version: " + status.Version,
			"version":   status.Version,
		})
	}
	group.Get("/jellyseerr/validate", validate)
	group.Post("/jellyseerr/validate", validate)
}
