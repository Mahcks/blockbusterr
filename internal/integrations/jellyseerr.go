package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type JellyseerrConfig struct {
	URL    string
	APIKey string
	UserID string // Optional: request as specific user
}

type Jellyseerr struct {
	config JellyseerrConfig
	client *http.Client
}

func NewJellyseerr(config JellyseerrConfig) *Jellyseerr {
	return &Jellyseerr{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MovieRequest represents a movie request payload
type MovieRequest struct {
	MediaType string `json:"mediaType"`
	MediaID   int    `json:"mediaId"` // TMDB ID
	UserID    *int   `json:"userId,omitempty"`
}

// ShowRequest represents a TV show request payload
type ShowRequest struct {
	MediaType string `json:"mediaType"`
	MediaID   int    `json:"mediaId"` // TVDB ID
	Seasons   string `json:"seasons"` // "all" or specific seasons
	UserID    *int   `json:"userId,omitempty"`
}

// RequestResponse represents the API response
type RequestResponse struct {
	ID        int    `json:"id"`
	Status    int    `json:"status"`
	MediaID   int    `json:"mediaId"`
	MediaType string `json:"mediaType"`
	Message   string `json:"message,omitempty"`
}

// StatusResponse for checking Jellyseerr status
type StatusResponse struct {
	Version string `json:"version"`
	Status  string `json:"status"`
}

// doRequest performs an HTTP request with proper headers
func (j *Jellyseerr) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := fmt.Sprintf("%s/api/v1%s", j.config.URL, path)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Api-Key", j.config.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// GetStatus checks if Jellyseerr is accessible
func (j *Jellyseerr) GetStatus() (*StatusResponse, error) {
	resp, err := j.doRequest("GET", "/status", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jellyseerr returned status %d: %s", resp.StatusCode, string(body))
	}

	var status StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode status response: %w", err)
	}

	return &status, nil
}

// RequestMovie requests a movie by TMDB ID
func (j *Jellyseerr) RequestMovie(tmdbID int) (*RequestResponse, error) {
	payload := MovieRequest{
		MediaType: "movie",
		MediaID:   tmdbID,
	}

	// Add user ID if configured
	if j.config.UserID != "" {
		if userID, err := strconv.Atoi(j.config.UserID); err == nil {
			payload.UserID = &userID
		}
	}

	resp, err := j.doRequest("POST", "/request", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Jellyseerr returns 201 for new requests, 200 for existing
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("jellyseerr returned status %d: %s", resp.StatusCode, string(body))
	}

	var result RequestResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode request response: %w", err)
	}

	return &result, nil
}

// RequestShow requests a TV show by TVDB ID
func (j *Jellyseerr) RequestShow(tvdbID int) (*RequestResponse, error) {
	payload := ShowRequest{
		MediaType: "tv",
		MediaID:   tvdbID,
		Seasons:   "all",
	}

	// Add user ID if configured
	if j.config.UserID != "" {
		if userID, err := strconv.Atoi(j.config.UserID); err == nil {
			payload.UserID = &userID
		}
	}

	resp, err := j.doRequest("POST", "/request", payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Jellyseerr returns 201 for new requests, 200 for existing
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("jellyseerr returned status %d: %s", resp.StatusCode, string(body))
	}

	var result RequestResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode request response: %w", err)
	}

	return &result, nil
}

// IsAlreadyRequested checks if content is already requested or available
func (result *RequestResponse) IsAlreadyRequested() bool {
	// Status codes: 1 = pending, 2 = approved, 3 = declined, 4 = available
	return result.Status == 1 || result.Status == 2 || result.Status == 4
}
