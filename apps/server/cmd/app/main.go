package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/charmbracelet/log"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
	"github.com/mahcks/blockbusterr/internal/notifications"
	"github.com/mahcks/blockbusterr/internal/notifications/discord"
	"github.com/mahcks/blockbusterr/internal/rest"
	"github.com/mahcks/blockbusterr/internal/scheduler"
	"github.com/mahcks/blockbusterr/internal/services/sqlite"
	"github.com/mahcks/blockbusterr/internal/websocket"
)

var (
	Version   = "dev"
	Timestamp = "unknown"
)

func main() {
	Timestamp = time.Now().Format(time.RFC3339)

	version := os.Getenv("VERSION")
	if version == "" {
		version = Version
	} else {
		Version = version
	}

	// Intialize the logger depending on the version of the app
	var logger *log.Logger
	if version == "dev" {
		// Enable debug logging in development, and use colorful output
		logger = log.NewWithOptions(os.Stdout, log.Options{
			Level:           log.DebugLevel,
			ReportCaller:    true,
			ReportTimestamp: true,
		})
	} else {
		// Production logger with JSON output and info level
		logger = log.NewWithOptions(os.Stdout, log.Options{
			Level:           log.InfoLevel,
			Formatter:       log.JSONFormatter,
			ReportTimestamp: true,
		})
	}

	// Set the logger as the default logger
	log.SetDefault(logger)

	log.Info("Starting the application", "version", version, "timestamp", Timestamp)

	gctx, cancel := global.WithCancel(global.New(context.Background()))
	var err error

	// Initialize websocket hub
	hub := websocket.NewHub(gctx)
	go hub.Run()

	{
		log.Info("Setting up SQLite database")
		gctx.Crate().SQL, err = sqlite.Setup(gctx)
		if err != nil {
			log.Error("Error setting up SQLite database", "error", err)
			cancel()
			return
		}
		log.Info("SQLite database setup complete")
	}

	// Initialize notifications
	notficationsInstance, err := notifications.Setup(gctx)
	if err != nil {
		log.Fatalf("Failed to initialize notifications: %v", err)
	}

	notficationsInstance.Discord.SendDiscordEmbed("https://discord.com/api/webhooks/1282948906157609021/oocIcgHqGCK-HP7BY1Fy0D45nDEO3ELICmr6N_QegMBNlNC9Lbp_iAu2cWOhHKeybVNR", discord.Embed{
		Title:       "Test",
		Description: "This is a test",
		Color:       0x00FF00,
		Fields: []discord.EmbedField{
			{
				Name:   "Field 1",
				Value:  "Value 1",
				Inline: true,
			},
		},
	})

	// Initialize helpers
	helpersInstance, err := helpers.SetupHelpers(gctx)
	if err != nil {
		log.Fatalf("Failed to initialize helpers: %v", err)
	}

	// Setup the scheduler
	schedulerInstance := scheduler.Setup(gctx, *helpersInstance)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})

	wg := sync.WaitGroup{}

	go func() {
		<-interrupt
		cancel()

		go func() {
			// If interrupt signal is not handled in 1 minute or interrupted once again, force shutdown
			select {
			case <-time.After(time.Minute):
			case <-interrupt:
			}
			log.Warn("Force shutdown")
		}()

		log.Warn("Shutting down...")

		wg.Wait()

		close(done)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		log.Info("Starting API server")
		if err := rest.New(gctx, hub, helpersInstance, schedulerInstance); err != nil {
			log.Error("Error starting API server", "error", err)
			cancel()
			return
		}
		log.Info("API server stopped")
	}()

	<-done

	log.Info("Application stopped")
	os.Exit(0)
}
