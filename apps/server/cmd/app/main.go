package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/helpers"
	"github.com/mahcks/blockbusterr/internal/rest"
	"github.com/mahcks/blockbusterr/internal/services/sqlite"
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

	gctx, cancel := global.WithCancel(global.New(context.Background()))
	var err error

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

	// Initialize helpers
	helpersInstance, err := helpers.SetupHelpers(gctx)
	if err != nil {
		log.Fatalf("Failed to initialize helpers: %v", err)
	}

	radarrReq, err := helpersInstance.Radarr.GetQualityProfiles()
	if err != nil {
		log.Fatalf("Failed to get root folders: %v", err)
	}

	// Pretty print the JSON with indentation
	jsonData, err := json.MarshalIndent(radarrReq, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(string(jsonData))

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
		if err := rest.New(gctx, helpersInstance); err != nil {
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
