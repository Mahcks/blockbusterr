package routes

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/blockbusterr/config"
)

func TestHealthReportsRunningVersion(t *testing.T) {
	for _, storedVersion := range []string{"", "v1.5.2"} {
		app := fiber.New()
		routes := NewRouteGroup(dynamicJobsTestContext{cfg: &config.Config{Version: storedVersion}})
		app.Get("/v1/", routes.Index)
		response, err := app.Test(httptest.NewRequest("GET", "/v1/", nil))
		if err != nil {
			t.Fatal(err)
		}
		var health HealthResponse
		err = json.NewDecoder(response.Body).Decode(&health)
		_ = response.Body.Close()
		if err != nil || response.StatusCode != fiber.StatusOK || health.Version != "test" {
			t.Fatalf("stored version=%q health=%+v status=%d err=%v", storedVersion, health, response.StatusCode, err)
		}
	}
}
