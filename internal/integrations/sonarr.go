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

// Sonarr client for TV show management
type Sonarr struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// SonarrConfig holds configuration for Sonarr client
type SonarrConfig struct {
	BaseURL string
	APIKey  string
}

// NewSonarr creates a new Sonarr API client
func NewSonarr(config SonarrConfig) *Sonarr {
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

	return &Sonarr{
		baseURL: baseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request to the Sonarr API
func (s *Sonarr) doRequest(ctx context.Context, method, endpoint string, body any) (*http.Response, error) {
	if s.baseURL == "" {
		return nil, errors.New("Sonarr base URL is not configured")
	}

	url := fmt.Sprintf("%s/api/v3%s", s.baseURL, endpoint)

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
	req.Header.Set("X-Api-Key", s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// Series represents a Sonarr TV series
type SonarrSeries struct {
	ID                int               `json:"id,omitempty"`
	Title             string            `json:"title"`
	Year              int               `json:"year,omitempty"`
	TvdbID            int               `json:"tvdbId"`
	TmdbID            int               `json:"tmdbId,omitempty"`
	ImdbID            string            `json:"imdbId,omitempty"`
	TitleSlug         string            `json:"titleSlug,omitempty"`
	Path              string            `json:"path,omitempty"`
	QualityProfileID  int               `json:"qualityProfileId"`
	LanguageProfileID int               `json:"languageProfileId,omitempty"`
	Monitored         bool              `json:"monitored"`
	SeriesType        string            `json:"seriesType,omitempty"`
	SeasonFolder      bool              `json:"seasonFolder"`
	RootFolderPath    string            `json:"rootFolderPath"`
	Tags              []int             `json:"tags,omitempty"`
	AddOptions        *SonarrAddOptions `json:"addOptions,omitempty"`
}

// SonarrAddOptions specifies options when adding a series
type SonarrAddOptions struct {
	SearchForMissingEpisodes bool `json:"searchForMissingEpisodes"`
}

// QualityProfile represents a Sonarr quality profile
type SonarrQualityProfile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// RootFolder represents a Sonarr root folder
type SonarrRootFolder struct {
	ID         int    `json:"id"`
	Path       string `json:"path"`
	FreeSpace  int64  `json:"freeSpace"`
	TotalSpace int64  `json:"totalSpace"`
	Accessible bool   `json:"accessible"`
}

// SystemStatus represents Sonarr system status
type SonarrSystemStatus struct {
	Version string `json:"version"`
	AppName string `json:"appName"`
}

// GetSystemStatus returns Sonarr system status
func (s *Sonarr) GetSystemStatus(ctx context.Context) (*SonarrSystemStatus, error) {
	resp, err := s.doRequest(ctx, http.MethodGet, "/system/status", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var status SonarrSystemStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}

// GetQualityProfiles returns available quality profiles
func (s *Sonarr) GetQualityProfiles(ctx context.Context) ([]SonarrQualityProfile, error) {
	resp, err := s.doRequest(ctx, http.MethodGet, "/qualityprofile", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var profiles []SonarrQualityProfile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return profiles, nil
}

// GetRootFolders returns available root folders
func (s *Sonarr) GetRootFolders(ctx context.Context) ([]SonarrRootFolder, error) {
	resp, err := s.doRequest(ctx, http.MethodGet, "/rootfolder", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var folders []SonarrRootFolder
	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return folders, nil
}

// AddSeries adds a TV series to Sonarr
func (s *Sonarr) AddSeries(ctx context.Context, series SonarrSeries) (*SonarrSeries, error) {
	resp, err := s.doRequest(ctx, http.MethodPost, "/series", series)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var addedSeries SonarrSeries
	if err := json.NewDecoder(resp.Body).Decode(&addedSeries); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &addedSeries, nil
}

// GetSeries returns all series in Sonarr
func (s *Sonarr) GetSeries(ctx context.Context) ([]SonarrSeries, error) {
	resp, err := s.doRequest(ctx, http.MethodGet, "/series", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var series []SonarrSeries
	if err := json.NewDecoder(resp.Body).Decode(&series); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return series, nil
}

// LookupSeries searches for a TV series by TVDB ID or title
func (s *Sonarr) LookupSeries(ctx context.Context, term string) ([]SonarrSeries, error) {
	endpoint := fmt.Sprintf("/series/lookup?term=%s", url.QueryEscape(term))

	resp, err := s.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var series []SonarrSeries
	if err := json.NewDecoder(resp.Body).Decode(&series); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return series, nil
}
