package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Radarr client for movie management
type Radarr struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// RadarrConfig holds configuration for Radarr client
type RadarrConfig struct {
	BaseURL string
	APIKey  string
}

// NewRadarr creates a new Radarr API client
func NewRadarr(config RadarrConfig) *Radarr {
	baseURL := config.BaseURL

	// Validate and normalize URL
	if baseURL != "" {
		// Check if URL has a scheme
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			// Default to http:// if no scheme provided
			baseURL = "http://" + baseURL
		}

		// Validate URL structure
		if _, err := url.Parse(baseURL); err != nil {
			// Invalid URL, keep as-is but will fail on first request
			baseURL = config.BaseURL
		}

		// Remove trailing slash
		baseURL = strings.TrimRight(baseURL, "/")
	}

	return &Radarr{
		baseURL: baseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request to the Radarr API
func (r *Radarr) doRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	if r.baseURL == "" {
		return nil, errors.New("Radarr base URL is not configured")
	}

	url := fmt.Sprintf("%s/api/v3%s", r.baseURL, endpoint)

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", r.apiKey)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// Movie represents a Radarr movie
type RadarrMovie struct {
	ID                  int               `json:"id,omitempty"`
	Title               string            `json:"title"`
	Year                int               `json:"year"`
	TmdbID              int               `json:"tmdbId"`
	ImdbID              string            `json:"imdbId,omitempty"`
	TitleSlug           string            `json:"titleSlug,omitempty"`
	Path                string            `json:"path,omitempty"`
	QualityProfileID    int               `json:"qualityProfileId"`
	Monitored           bool              `json:"monitored"`
	MinimumAvailability string            `json:"minimumAvailability"`
	RootFolderPath      string            `json:"rootFolderPath"`
	Tags                []int             `json:"tags,omitempty"`
	AddOptions          *RadarrAddOptions `json:"addOptions,omitempty"`
}

// RadarrAddOptions specifies options when adding a movie
type RadarrAddOptions struct {
	SearchForMovie bool `json:"searchForMovie"`
}

// QualityProfile represents a Radarr quality profile
type QualityProfile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// RootFolder represents a Radarr root folder
type RootFolder struct {
	ID         int    `json:"id"`
	Path       string `json:"path"`
	FreeSpace  int64  `json:"freeSpace"`
	TotalSpace int64  `json:"totalSpace"`
	Accessible bool   `json:"accessible"`
}

// SystemStatus represents Radarr system status
type SystemStatus struct {
	Version string `json:"version"`
	AppName string `json:"appName"`
}

// GetSystemStatus returns Radarr system status
func (r *Radarr) GetSystemStatus(ctx context.Context) (*SystemStatus, error) {
	resp, err := r.doRequest(ctx, http.MethodGet, "/system/status", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var status SystemStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}

// GetQualityProfiles returns available quality profiles
func (r *Radarr) GetQualityProfiles(ctx context.Context) ([]QualityProfile, error) {
	resp, err := r.doRequest(ctx, http.MethodGet, "/qualityprofile", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var profiles []QualityProfile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return profiles, nil
}

// GetRootFolders returns available root folders
func (r *Radarr) GetRootFolders(ctx context.Context) ([]RootFolder, error) {
	resp, err := r.doRequest(ctx, http.MethodGet, "/rootfolder", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var folders []RootFolder
	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return folders, nil
}

// AddMovie adds a movie to Radarr
func (r *Radarr) AddMovie(ctx context.Context, movie RadarrMovie) (*RadarrMovie, error) {
	resp, err := r.doRequest(ctx, http.MethodPost, "/movie", movie)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var addedMovie RadarrMovie
	if err := json.NewDecoder(resp.Body).Decode(&addedMovie); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &addedMovie, nil
}

// GetMovies returns all movies in Radarr
func (r *Radarr) GetMovies(ctx context.Context) ([]RadarrMovie, error) {
	resp, err := r.doRequest(ctx, http.MethodGet, "/movie", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var movies []RadarrMovie
	if err := json.NewDecoder(resp.Body).Decode(&movies); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return movies, nil
}

// LookupMovie searches for a movie by TMDB ID or title
func (r *Radarr) LookupMovie(ctx context.Context, term string) ([]RadarrMovie, error) {
	endpoint := fmt.Sprintf("/movie/lookup?term=%s", term)

	resp, err := r.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var movies []RadarrMovie
	if err := json.NewDecoder(resp.Body).Decode(&movies); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return movies, nil
}
