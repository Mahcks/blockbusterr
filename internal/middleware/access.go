package middleware

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
)

// OwnerAccess protects the single-owner UI and API with one deployment token.
func OwnerAccess(token string) fiber.Handler {
	return basicauth.New(basicauth.Config{
		Users: map[string]string{"blockbusterr": token},
		Realm: "Blockbusterr",
	})
}

// SameOriginMutations rejects browser writes initiated by another origin.
// Requests without browser origin headers remain available to authenticated API clients.
func SameOriginMutations() fiber.Handler {
	return func(c *fiber.Ctx) error {
		switch c.Method() {
		case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
			return c.Next()
		}

		source := c.Get(fiber.HeaderOrigin)
		if source == "" {
			source = c.Get(fiber.HeaderReferer)
		}
		if source == "" {
			return c.Next()
		}
		u, err := url.Parse(source)
		if err != nil || !strings.EqualFold(u.Host, string(c.Context().Host())) {
			return fiber.ErrForbidden
		}
		return c.Next()
	}
}
