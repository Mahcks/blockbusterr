package rest

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
	v1 "github.com/mahcks/blockbusterr/internal/rest/v1"
	"github.com/mahcks/blockbusterr/internal/scheduler"
	"github.com/mahcks/blockbusterr/internal/websocket"
	commonErrors "github.com/mahcks/blockbusterr/pkg/errors"
)

var allowedHeaders = []string{
	"Content-Type",
	"Content-Length",
	"Accept-Encoding",
	"Authorization",
	"Cookie",
}

type APIErrorResponseBodyError struct {
	StatusCode int      `json:"status_code"`
	Timestamp  int      `json:"timestamp"`
	Error      APIError `json:"error"`
	TraceID    string   `json:"trace_id,omitempty"`
}

type APIError struct {
	StatusCode int                    `json:"status_code"`
	Error      string                 `json:"error"`
	ErrorCode  int                    `json:"error_code"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

func New(gctx global.Context, hub *websocket.Hub, helpers *helpers.Helpers, scheduler *scheduler.Scheduler) error {
	if helpers == nil {
		return errors.New("helpers is nil")
	}

	app := fiber.New(fiber.Config{
		// Custom error handler for common.APIError
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			log.Error("error in fiber", "error", err)

			// Handle fiber-specific errors
			var fe *fiber.Error
			if errors.As(err, &fe) {
				return ctx.Status(fe.Code).SendString(fe.Message)
			}

			// Handle common API errors
			var ce commonErrors.APIError
			if errors.As(err, &ce) {
				ctx.Set("Content-Type", "application/json")
				ctx.Status(ce.ExpectedHTTPStatus())

				responseBody := &APIErrorResponseBodyError{
					StatusCode: ce.Code(),
					Timestamp:  int(time.Now().Unix()),
					Error: APIError{
						StatusCode: ce.ExpectedHTTPStatus(),
						Error:      ce.Message(),
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
		AllowOrigins: "http://localhost:5555,http://localhost:5173",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE",
		AllowHeaders: strings.Join(allowedHeaders, ", "),
	}))

	v1Group := app.Group("/v1")
	v1.New(gctx, hub, helpers, scheduler, v1Group)

	errCh := make(chan error)
	// Listen for connections in a separate goroutine.
	// When Listen returns, send the error (or nil if none) on errCh.
	go func() {
		if err := app.Listen(fmt.Sprintf("%v:%v", "0.0.0.0", "3000")); err != nil {
			errCh <- err
		} else {
			errCh <- nil
		}
		close(errCh)
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
