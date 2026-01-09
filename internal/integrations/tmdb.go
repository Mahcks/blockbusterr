package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	TMDBAPIBaseURL   = "https://api.themoviedb.org/3"
	TMDBImageBaseURL = "https://image.tmdb.org/t/p/w500"
)

// TMDB is the client for interacting with TMDB API
type TMDB struct {
	apiKey     string
	httpClient *http.Client
}

// TMDBConfig holds configuration for TMDB client
type TMDBConfig struct {
	APIKey string
}

// NewTMDB creates a new TMDB API client
func NewTMDB(config TMDBConfig) *TMDB {
	return &TMDB{
		apiKey: config.APIKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TMDBMovie represents a movie from TMDB
type TMDBMovie struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}

// TMDBShow represents a TV show from TMDB
type TMDBShow struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}

// GetMoviePosterURL fetches the poster URL for a movie by TMDB ID
func (t *TMDB) GetMoviePosterURL(ctx context.Context, tmdbID int) (string, error) {
	if t.apiKey == "" {
		return "", nil // Return empty if no API key configured
	}

	url := fmt.Sprintf("%s/movie/%d?api_key=%s", TMDBAPIBaseURL, tmdbID, t.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var movie TMDBMovie
	if err := json.Unmarshal(body, &movie); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if movie.PosterPath != "" {
		return TMDBImageBaseURL + movie.PosterPath, nil
	}

	return "", nil
}

// GetShowPosterURL fetches the poster URL for a TV show by TMDB ID
func (t *TMDB) GetShowPosterURL(ctx context.Context, tmdbID int) (string, error) {
	if t.apiKey == "" {
		return "", nil // Return empty if no API key configured
	}

	url := fmt.Sprintf("%s/tv/%d?api_key=%s", TMDBAPIBaseURL, tmdbID, t.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var show TMDBShow
	if err := json.Unmarshal(body, &show); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if show.PosterPath != "" {
		return TMDBImageBaseURL + show.PosterPath, nil
	}

	return "", nil
}
