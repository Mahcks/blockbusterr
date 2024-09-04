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

	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/internal/rest"
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

	gctx, cancel := global.WithCancel(global.New(context.Background()))
	var err error

	{
		slog.Info("Setting up Trakt API")
		gctx.Crate().Trakt, err = trakt.Setup(gctx, trakt.SetupOptions{
			ClientID:     "1234567890",
			ClientSecret: "0987654321",
		})
		if err != nil {
			slog.Error("Error setting up Trakt API", "error", err)
			cancel()
			return
		}
		slog.Info("Trakt API setup complete")

		popularMovies, err := gctx.Crate().Trakt.GetPopularMovies(gctx, 1)
		if err != nil {
			slog.Error("Error getting popular movies", "error", err)
			cancel()
			return
		}

		fmt.Println(popularMovies)
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
