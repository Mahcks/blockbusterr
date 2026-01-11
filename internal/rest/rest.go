package rest

import (
	"encoding/json"
	"errors"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	htmlEngine "github.com/gofiber/template/html/v2"
	"github.com/mahcks/blockbusterr/internal/global"
	v1 "github.com/mahcks/blockbusterr/internal/rest/v1"
	"github.com/mahcks/blockbusterr/internal/rest/v1/routes"
	apiErrors "github.com/mahcks/blockbusterr/pkg/api_errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

var allowedHeaders = []string{
	"Content-Type",
	"Content-Length",
	"Accept-Encoding",
	"Authorization",
	"Cookie",
	"X-Api-Key",
	"X-CSRF-Token",
}

func New(gctx global.Context) error {
	// Check DISABLE_UI environment variable (UI enabled by default)
	uiEnabled := true
	if disableUI := strings.ToLower(strings.TrimSpace(os.Getenv("DISABLE_UI"))); disableUI != "" {
		uiEnabled = !(disableUI == "true" || disableUI == "1" || disableUI == "yes")
	}

	// Initialize template engine with custom functions
	engine := htmlEngine.New("./web/templates", ".html")
	engine.Reload(true) // Enable template reloading in development
	engine.AddFunc("json", func(v interface{}) template.JS {
		b, _ := json.Marshal(v)
		return template.JS(b)
	})
	engine.AddFunc("mul", func(a, b float64) float64 {
		return a * b
	})

	app := fiber.New(fiber.Config{
		Views:                 engine,
		DisableStartupMessage: false,
		ServerHeader:          "Blockbusterr",
		AppName:               "Blockbusterr",
		// Custom error handler for common.APIError
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			log.Errorw("error in fiber", "error", err)

			// Handle fiber-specific errors
			var fe *fiber.Error
			if errors.As(err, &fe) {
				return ctx.Status(fe.Code).SendString(fe.Message)
			}

			// Handle common API errors
			var ce apiErrors.APIError
			if errors.As(err, &ce) {
				ctx.Set("Content-Type", "application/json")
				ctx.Status(ce.ExpectedHTTPStatus())

				responseBody := &structures.APIErrorResponseBodyError{
					StatusCode: ce.Code(),
					Timestamp:  int(time.Now().Unix()),
					Error: structures.APIError{
						StatusCode: ce.ExpectedHTTPStatus(),
						Message:    ce.Message(),
						ErrorCode:  ce.Code(),
						Details:    ce.GetFields(),
					},
				}
				return ctx.JSON(responseBody)
			}

			// Fallback error handling
			return ctx.Status(500).SendString("Internal Server Error")
		},
	})

	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     strings.Join(allowedHeaders, ", "),
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length, Content-Type",
	}))

	// Serve static files
	app.Static("/static", "./web/static")

	// Conditionally enable UI routes
	if uiEnabled {
		log.Info("Web UI is enabled")
		uiRoutes := routes.NewRouteGroup(gctx)
		routes.RegisterUIRoutes(uiRoutes, app)
	} else {
		log.Info("Web UI is disabled")
		// Provide a simple message on root route
		app.Get("/", func(c *fiber.Ctx) error {
			return c.SendString("Blockbusterr API - Web UI is disabled. Set UI_ENABLED=true environment variable to enable.")
		})
	}

	v1Group := app.Group("/v1")
	v1.New(gctx, v1Group)

	errCh := make(chan error, 1) // Buffered to prevent goroutine leak
	// Listen for connections in a separate goroutine.
	go func() {
		errCh <- app.Listen("0.0.0.0:9090")
	}()

	// Wait for the server to start or for a shutdown signal,
	// whichever comes first.
	select {
	case <-gctx.Done():
		// A shutdown signal was received before the server started,
		// so try to stop the server.
		if err := app.Shutdown(); err != nil {
			log.Error("error while shutting down server", "error", err)
		}
		return nil
	case err := <-errCh:
		// The server has exited, so return the error (if any).
		if err != nil {
			log.Error("error from server", "error", err)
			return err
		}
	}

	// Wait for a shutdown signal before stopping the server.
	<-gctx.Done()

	// Shutdown the server
	if err := app.Shutdown(); err != nil {
		log.Error("error while shutting down server", "error", err)
		return err
	}

	return nil
}
