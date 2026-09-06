package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

func (rg *RouteGroup) RegisterJellyseerrRoutes(group fiber.Router) {
	// Validate Jellyseerr connection
	validate := func(c *fiber.Ctx) error {
		var url, apiKey string
		if c.Method() == fiber.MethodPost {
			var request struct {
				URL    string `json:"url"`
				APIKey string `json:"api_key"`
			}
			if err := c.BodyParser(&request); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body", "connected": false})
			}
			url, apiKey = request.URL, request.APIKey
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
		c.Set(fiber.HeaderCacheControl, "no-store")
		if url == "" || apiKey == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Jellyseerr URL and API key are required", "connected": false})
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
