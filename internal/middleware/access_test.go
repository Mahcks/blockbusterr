package middleware

import (
	"encoding/base64"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestOwnerAccessAndSameOriginMutations(t *testing.T) {
	app := fiber.New()
	app.Use(OwnerAccess("test-owner-token"), SameOriginMutations())
	app.Post("/change", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	request := func(origin, token string) int {
		req := httptest.NewRequest(fiber.MethodPost, "http://blockbusterr.local/change", nil)
		req.Host = "blockbusterr.local"
		if origin != "" {
			req.Header.Set(fiber.HeaderOrigin, origin)
		}
		if token != "" {
			credentials := base64.StdEncoding.EncodeToString([]byte("blockbusterr:" + token))
			req.Header.Set(fiber.HeaderAuthorization, "Basic "+credentials)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		return resp.StatusCode
	}

	if status := request("", ""); status != fiber.StatusUnauthorized {
		t.Fatalf("anonymous mutation status = %d", status)
	}
	if status := request("http://evil.local", "test-owner-token"); status != fiber.StatusForbidden {
		t.Fatalf("cross-origin mutation status = %d", status)
	}
	if status := request("http://blockbusterr.local", "test-owner-token"); status != fiber.StatusNoContent {
		t.Fatalf("authenticated same-origin mutation status = %d", status)
	}
}

func TestSameOriginMutationsWithoutAuthentication(t *testing.T) {
	app := fiber.New()
	app.Use(SameOriginMutations())
	app.Post("/change", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })

	request := func(origin string) int {
		req := httptest.NewRequest(fiber.MethodPost, "http://blockbusterr.local/change", nil)
		req.Host = "blockbusterr.local"
		if origin != "" {
			req.Header.Set(fiber.HeaderOrigin, origin)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		return resp.StatusCode
	}

	if status := request("http://evil.local"); status != fiber.StatusForbidden {
		t.Fatalf("cross-origin mutation status = %d", status)
	}
	if status := request(""); status != fiber.StatusNoContent {
		t.Fatalf("headerless API mutation status = %d", status)
	}
}
