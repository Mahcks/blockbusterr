package routes

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/global"
)

func TestRadarrValidationAcceptsCredentialsInPostBody(t *testing.T) {
	const apiKey = "test-secret"
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("upstream query contains credentials: %q", r.URL.RawQuery)
		}
		if got := r.Header.Get("X-Api-Key"); got != apiKey {
			t.Errorf("API key header = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"1.0.0"}`))
	}))
	defer upstream.Close()

	gctx := global.New(context.Background(), &config.Config{}, nil, "test", "test", nil)
	app := fiber.New()
	RegisterRadarrRoutes(NewRouteGroup(gctx), app.Group("/v1"))
	body := `{"url":"` + upstream.URL + `","api_key":"` + apiKey + `"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/radarr/validate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if cache := resp.Header.Get("Cache-Control"); !strings.Contains(cache, "no-store") {
		t.Fatalf("Cache-Control = %q", cache)
	}
}
