package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
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
	// VERSION env var always wins; otherwise keep build-time value or default to dev.
	if v := strings.TrimSpace(os.Getenv("VERSION")); v != "" {
		Version = v
	} else if Version == "" {
		Version = "dev"
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
	slog.Info("config loaded", "path", cfg.ConfigFilePath)

	// Initialize database
	dataDir := resolveDataDir()
	slog.Info("database path resolved", "dir", dataDir)
	db, err := database.New(dataDir)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	baseCtx, cancel := context.WithCancel(context.Background())
	gctx := global.New(
		baseCtx,
		cfg,
		db,
		Version,
		Commit,
	)
	if gctx.Database() == nil {
		slog.Error("database is nil in global context after initialization")
		os.Exit(1)
	}

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
				slog.Warn("Force shutdown - timed out")
			case <-interrupt:
				slog.Warn("Force shutdown - second signal received")
			}
			os.Exit(1)
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

func resolveDataDir() string {
	if dataDir := strings.TrimSpace(os.Getenv("DATA_DIR")); dataDir != "" {
		return dataDir
	}

	candidates := []string{
		"./data",
		"../data",
		"/app/data",
	}
	for _, dir := range candidates {
		if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
			return dir
		}
	}

	return "./data"
}
