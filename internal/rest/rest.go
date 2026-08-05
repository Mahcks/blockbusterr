package rest

import (
	"errors"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/logger"
	htmlEngine "github.com/gofiber/template/html/v2"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/middleware"
	v1 "github.com/mahcks/blockbusterr/internal/rest/v1"
	"github.com/mahcks/blockbusterr/internal/rest/v1/routes"
)

const shutdownTimeout = 5 * time.Second

func New(gctx global.Context) error {
	// Check DISABLE_UI environment variable (UI enabled by default)
	uiEnabled := true
	if disableUI := strings.ToLower(strings.TrimSpace(os.Getenv("DISABLE_UI"))); disableUI != "" {
		uiEnabled = disableUI != "true" && disableUI != "1" && disableUI != "yes"
	}

	// Initialize template engine with custom functions
	engine := htmlEngine.New("./web/templates", ".html")
	engine.Reload(true) // Enable template reloading in development
	engine.AddFunc("safeHTML", func(s string) template.HTML {
		return template.HTML(s)
	})
	engine.AddFunc("mul", func(a, b float64) float64 {
		return a * b
	})
	engine.AddFunc("add", func(a, b int) int {
		return a + b
	})
	engine.AddFunc("sub", func(a, b int) int {
		return a - b
	})
	engine.AddFunc("contains", func(s, substr string) bool {
		return strings.Contains(s, substr)
	})

	app := fiber.New(fiber.Config{
		Views:                 engine,
		ReadBufferSize:        16 * 1024,
		DisableStartupMessage: false,
		ServerHeader:          "Blockbusterr",
		AppName:               "Blockbusterr",
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			log.Errorw("error in fiber", "error", err)

			// Handle fiber-specific errors
			var fe *fiber.Error
			if errors.As(err, &fe) {
				return ctx.Status(fe.Code).SendString(fe.Message)
			}

			// Fallback error handling
			return ctx.Status(500).SendString("Internal Server Error")
		},
	})

	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	// Security headers
	app.Use(middleware.SecureHeaders())

	// Serve static files
	app.Static("/static", "./web/static")
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.Redirect("/static/favicon.svg", fiber.StatusPermanentRedirect)
	})

	ownerToken := strings.TrimSpace(os.Getenv("BLOCKBUSTERR_AUTH_TOKEN"))
	app.Use(middleware.SameOriginMutations())
	if ownerToken == "" {
		log.Warn("Owner authentication is disabled; keep Blockbusterr on a trusted network")
	} else {
		if len(ownerToken) < 32 {
			return errors.New("BLOCKBUSTERR_AUTH_TOKEN must contain at least 32 characters")
		}
		app.Use(middleware.OwnerAccess(ownerToken))
	}

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

	// Register config routes at root level (not under /v1)
	routes.RegisterConfigRoutes(app, gctx)

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
		if err := app.ShutdownWithTimeout(shutdownTimeout); err != nil {
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
	if err := app.ShutdownWithTimeout(shutdownTimeout); err != nil {
		log.Error("error while shutting down server", "error", err)
		return err
	}

	return nil
}
