package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/rest"
	"github.com/mahcks/blockbusterr/internal/services"
)

var (
	Version = "dev"
	Commit  = "none"
)

func main() {
	// If Version wasn't set at build time via ldflags, use env var or default
	if Version == "dev" || Version == "" {
		if v := os.Getenv("VERSION"); v != "" {
			Version = v
		} else if Version == "" {
			Version = "dev"
		}
	}

	if Commit == "none" || Commit == "" {
		if c := os.Getenv("COMMIT"); c != "" {
			Commit = c
		}
	}

	// Set the log level based on the version
	var logLevel slog.Level
	if Version == "dev" {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	if Version == "dev" {
		slog.Info("Starting in development mode", "version", Version, "commit", Commit)
	} else {
		slog.Info("Starting in production mode", "version", Version, "commit", Commit)
	}

	cfg, err := config.New(Version)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := database.New("./data")
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	gctx, cancel := global.WithCancel(global.New(
		context.Background(),
		cfg,
		db,
		Version,
		Commit,
	))

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	done := make(chan struct{})
	wg := sync.WaitGroup{}

	// Start the job scheduler
	scheduler := services.NewScheduler(gctx.Config, db, Version)
	scheduler.Start()

	go func() {
		<-interrupt
		cancel()

		go func() {
			select {
			case <-time.After(time.Minute):
			case <-interrupt:
			}
			slog.Warn("Force shutdown - timed out")
		}()

		slog.Warn("Shutting down...")

		// Stop the scheduler
		scheduler.Stop()

		wg.Wait()

		// Close database
		if err := db.Close(); err != nil {
			slog.Error("failed to close database", "error", err)
		}

		close(done)
	}()

	wg.Go(func() {
		slog.Info("api", "status", "starting")
		if err := rest.New(gctx); err != nil {
			slog.Error("api", "status", "errored", "error", err)
			os.Exit(1)
		}
		slog.Info("api", "status", "started")
	})

	<-done
	slog.Info("Shutdown complete")
	os.Exit(0)
}
