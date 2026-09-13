package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func TestConnectionOptionsUseDraftAndSavedCredentials(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "test-secret" || r.URL.RawQuery != "" {
			t.Error("credentials must use the API key header")
		}
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "qualityprofile") {
			_, _ = w.Write([]byte(`[{"id":7,"name":"HD"}]`))
		} else if strings.Contains(r.URL.Path, "rootfolder") {
			_, _ = w.Write([]byte(`[{"id":1,"path":"/media"}]`))
		} else {
			_, _ = w.Write([]byte(`{"version":"1.0.0"}`))
		}
	}))
	defer upstream.Close()
	for _, saved := range []bool{false, true} {
		cfg := &config.Config{}
		if saved {
			cfg.Radarr.URL, cfg.Radarr.APIKey = upstream.URL, "test-secret"
			cfg.Sonarr.URL, cfg.Sonarr.APIKey = upstream.URL, "test-secret"
			cfg.Jellyseerr.URL, cfg.Jellyseerr.APIKey = upstream.URL, "test-secret"
		}
		gctx := global.New(context.Background(), cfg, nil, "test", "test", nil)
		app := fiber.New()
		routes := NewRouteGroup(gctx)
		RegisterRadarrRoutes(routes, app.Group("/v1"))
		RegisterSonarrRoutes(routes, app.Group("/v1"))
		routes.RegisterJellyseerrRoutes(app.Group("/v1"))
		for _, service := range []string{"radarr", "sonarr", "jellyseerr"} {
			paths := []string{"validate"}
			if service != "jellyseerr" {
				paths = append(paths, "quality-profiles", "root-folders")
			}
			for _, path := range paths {
				for _, scenario := range []struct {
					name, url, key string
					status         int
				}{
					{"draft", upstream.URL, "test-secret", 200},
					{"hidden", upstream.URL, "", map[bool]int{true: 200, false: 400}[saved]},
					{"different server", "http://127.0.0.1:1", "", 400},
				} {
					t.Run(fmt.Sprintf("%s/%s/saved=%t/%s", service, path, saved, scenario.name), func(t *testing.T) {
						body, err := json.Marshal(map[string]string{"url": scenario.url, "api_key": scenario.key})
						if err != nil {
							t.Fatal(err)
						}
						req := httptest.NewRequest(http.MethodPost, "/v1/"+service+"/"+path, bytes.NewReader(body))
						req.Header.Set("Content-Type", "application/json")
						resp, err := app.Test(req, -1)
						if err != nil {
							t.Fatal(err)
						}
						defer func() { _ = resp.Body.Close() }()
						if resp.StatusCode != scenario.status {
							t.Fatalf("status=%d want=%d", resp.StatusCode, scenario.status)
						}
						if resp.Header.Get("Cache-Control") != "no-store" {
							t.Error("missing no-store")
						}
					})
				}
			}
		}
		if !saved && (cfg.Radarr.APIKey != "" || cfg.Sonarr.APIKey != "" || cfg.Jellyseerr.APIKey != "") {
			t.Error("draft request persisted credentials")
		}
	}
}
