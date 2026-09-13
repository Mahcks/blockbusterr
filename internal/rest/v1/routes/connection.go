package routes

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// connectionCredentials accepts draft credentials without saving them. A hidden
// saved key may only be reused for the same server, never a replacement URL.
func connectionCredentials(c *fiber.Ctx, savedURL, savedKey string) (string, string, error) {
	c.Set(fiber.HeaderCacheControl, "no-store")
	if c.Method() != fiber.MethodPost {
		return savedURL, savedKey, nil
	}
	var request struct {
		URL    string `json:"url"`
		APIKey string `json:"api_key"`
	}
	if err := c.BodyParser(&request); err != nil {
		return "", "", fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	url, key := strings.TrimSpace(request.URL), strings.TrimSpace(request.APIKey)
	if key == "" && url == savedURL {
		key = savedKey
	}
	if url == "" || key == "" {
		return "", "", fiber.NewError(fiber.StatusBadRequest, "Server URL and API key are required; enter a key when changing servers")
	}
	return url, key, nil
}
