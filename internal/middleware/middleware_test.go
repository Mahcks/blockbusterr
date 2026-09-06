package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSecureHeadersIncludesRestrictiveCSP(t *testing.T) {
	app := fiber.New()
	app.Use(SecureHeaders())
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	policy := resp.Header.Get("Content-Security-Policy")
	if !strings.Contains(policy, "script-src 'self'") || strings.Contains(policy, "unsafe-eval") {
		t.Fatalf("unsafe content security policy: %q", policy)
	}
}
