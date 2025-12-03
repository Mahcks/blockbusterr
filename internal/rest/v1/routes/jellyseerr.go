package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

func (rg *RouteGroup) RegisterJellyseerrRoutes(group fiber.Router) {
	// Validate Jellyseerr connection
	group.Get("/jellyseerr/validate", func(c *fiber.Ctx) error {
		// Check if test mode
		testMode := c.Query("test") == "true"

		var url, apiKey string

		if testMode {
			// Test mode: use form values
			url = c.Query("url")
			apiKey = c.Query("api_key")

			if url == "" || apiKey == "" {
				return c.Status(400).JSON(fiber.Map{
					"error": "Jellyseerr URL and API key are required",
				})
			}
		} else {
			// Normal mode, use saved config
			cfg := rg.gctx.Config()
			url = cfg.Jellyseerr.URL
			apiKey = cfg.Jellyseerr.APIKey

			if url == "" || apiKey == "" {
				return c.Status(400).JSON(fiber.Map{
					"connected": false,
					"error":     "Jellyseerr URL or API key is not configured",
				})
			}
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
	})
}
