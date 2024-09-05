package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/rest"
	"github.com/mahcks/blockbusterr/internal/services/sqlite"
	"github.com/mahcks/blockbusterr/internal/services/trakt"
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
	var logger *slog.Logger
	if version == "dev" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	} else {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	slog.SetDefault(logger)

	slog.Info("Starting the application", "version", version, "timestamp", Timestamp)

	// Load configuration
	cfg, err := config.New(Version, time.Now())
	if err != nil {
		slog.Error("Error loading config", "error", err)
		os.Exit(1)
	}

	gctx, cancel := global.WithCancel(global.New(context.Background(), cfg))

	{
		slog.Info("Setting up SQLite database")
		gctx.Crate().SQL, err = sqlite.Setup(gctx)
		if err != nil {
			slog.Error("Error setting up SQLite database", "error", err)
			cancel()
			return
		}
		slog.Info("SQLite database setup complete")
	}

	{
		slog.Info("Setting up Trakt API")
		gctx.Crate().Trakt, err = trakt.Setup(gctx, trakt.SetupOptions{
			ClientID:     gctx.Config().Trakt.ClientID,
			ClientSecret: gctx.Config().Trakt.ClientSecret,
		})
		if err != nil {
			slog.Error("Error setting up Trakt API", "error", err)
			cancel()
			return
		}
		slog.Info("Trakt API setup complete")
	}

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
			fmt.Println("force shutdown")
		}()

		fmt.Println("shutting down")

		wg.Wait()

		close(done)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()

		slog.Info("Starting API server")
		if err := rest.New(gctx); err != nil {
			slog.Error("Error starting API server", "error", err)
			cancel()
			return
		}
		slog.Info("API server stopped")
	}()

	<-done

	slog.Info("Application stopped")
	os.Exit(0)
}
