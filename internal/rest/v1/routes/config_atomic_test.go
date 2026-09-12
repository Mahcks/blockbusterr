package routes

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/global"
)

// Inject a token refresh between the handler's initial read and its atomic update.
type interveningConfigContext struct{ global.Context }

func (c interveningConfigContext) UpdateConfig(update func(*config.Config) error) error {
	if err := global.UpdateConfig(c.Context, func(cfg *config.Config) error { cfg.Trakt.AccessToken = "refreshed-token"; return nil }); err != nil {
		return err
	}
	return global.UpdateConfig(c.Context, update)
}

func TestConfigMutationsPreserveInterveningTokenRefresh(t *testing.T) {
	for _, path := range []string{"/config/save", "/jobs/config/save", "/config/import", "/config/jobs/import", "/jobs/recipes/balanced-trending-movies"} {
		t.Run(path, func(t *testing.T) {
			cfg := portableTestConfig(t)
			cfg.Jobs.SyncInterval = "1h"
			cfg.Radarr.URL = "http://radarr.invalid"
			ctx := interveningConfigContext{global.New(context.Background(), cfg, nil, "test", "", nil)}
			app := fiber.New()
			RegisterConfigRoutes(app, ctx)
			RegisterUIRoutes(NewRouteGroup(ctx), app)
			AddDynamicJobsRoutes(app, ctx)
			var status int
			if path == "/config/import" || path == "/config/jobs/import" {
				exportPath := "/config/export"
				if path == "/config/jobs/import" {
					exportPath = "/config/jobs/job-1/export"
				}
				resp, err := app.Test(httptest.NewRequest("GET", exportPath, nil), -1)
				if err != nil {
					t.Fatal(err)
				}
				data, err := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if err != nil {
					t.Fatal(err)
				}
				resp = uploadConfig(t, app, path, data)
				status = resp.StatusCode
				_ = resp.Body.Close()
			} else {
				req := httptest.NewRequest("POST", path, strings.NewReader("jobs.sync_interval=1h"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				resp, err := app.Test(req, -1)
				if err != nil {
					t.Fatal(err)
				}
				status = resp.StatusCode
				if status >= 400 {
					body, _ := io.ReadAll(resp.Body)
					t.Log(string(body))
				}
				_ = resp.Body.Close()
			}
			if status < 200 || status >= 300 {
				t.Fatalf("status=%d", status)
			}
			if got := ctx.Config().Trakt.AccessToken; got != "refreshed-token" {
				t.Fatalf("token refresh overwritten: %q", got)
			}
		})
	}
}

func TestSettingsRejectUnsafeSchedules(t *testing.T) {
	for _, interval := range []string{"-1s", "0s", "0 0 31 2 *"} {
		app, cfg := newConfigSaveTestApp(t)
		rec, _ := postConfigSave(t, app, map[string]string{"jobs.sync_interval": interval})
		if rec.Code != fiber.StatusBadRequest || cfg.Jobs.SyncInterval != "1h" {
			t.Fatalf("interval=%q status=%d saved=%q", interval, rec.Code, cfg.Jobs.SyncInterval)
		}
		if err := validateDynamicJob(cfg, config.DynamicJob{SyncInterval: interval}); err == nil {
			t.Fatalf("dynamic job accepted %q", interval)
		}
	}
}
