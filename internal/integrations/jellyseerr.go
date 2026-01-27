package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

type JellyseerrConfig struct {
	URL    string
	APIKey string
	UserID string // Optional: request as specific user
	// Optional: Username/password for request authentication (respects user permissions)
	RequestEmail    string
	RequestPassword string
}

type Jellyseerr struct {
	config         JellyseerrConfig
	client         *http.Client
	sessionExpiry  time.Time
	sessionMutex   sync.RWMutex
	useCredentials bool
}

func NewJellyseerr(config JellyseerrConfig) *Jellyseerr {
	// Create cookie jar for session management
	jar, _ := cookiejar.New(nil)

	// Normalize URL by removing trailing slash to prevent double slashes in API paths
	if config.URL != "" {
		config.URL = strings.TrimRight(config.URL, "/")
	}

	useCredentials := config.RequestEmail != "" && config.RequestPassword != ""

	j := &Jellyseerr{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
		useCredentials: useCredentials,
	}

	// If credentials are provided, login immediately
	if useCredentials {
		if err := j.login(); err != nil {
			log.Warnf("Failed to login to Jellyseerr with credentials: %v", err)
		} else {
			log.Info("Successfully authenticated to Jellyseerr with user credentials")
		}
	}

	return j
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
	MediaID   int    `json:"mediaId"` // TMDB ID (Jellyseerr uses TMDB for TV shows, not TVDB)
	Seasons   string `json:"seasons"` // "all" or specific seasons
	UserID    *int   `json:"userId,omitempty"`
}

// RequestResponse represents the API response
type RequestResponse struct {
	ID         int    `json:"id"`
	Status     int    `json:"status"`
	MediaID    int    `json:"mediaId"`
	MediaType  string `json:"mediaType"`
	Message    string `json:"message,omitempty"`
	HTTPStatus int    `json:"-"` // HTTP status code (201 = new, 200 = existing)
}

// StatusResponse for checking Jellyseerr status
type StatusResponse struct {
	Version string `json:"version"`
	Status  string `json:"status"`
}

// LoginRequest represents the login payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

// login authenticates with Jellyseerr using username/password
func (j *Jellyseerr) login() error {
	j.sessionMutex.Lock()
	defer j.sessionMutex.Unlock()

	payload := LoginRequest{
		Email:    j.config.RequestEmail,
		Password: j.config.RequestPassword,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/auth/local", j.config.URL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := j.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute login request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	// Session cookie is automatically stored in the cookie jar
	// Set expiry to 7 days (typical session expiry)
	j.sessionExpiry = time.Now().Add(7 * 24 * time.Hour)

	log.Infof("Logged in to Jellyseerr as: %s (ID: %d)", loginResp.DisplayName, loginResp.ID)

	return nil
}

// ensureAuthenticated checks if session is valid and refreshes if needed
func (j *Jellyseerr) ensureAuthenticated() error {
	if !j.useCredentials {
		return nil // Using API key, no session management needed
	}

	j.sessionMutex.RLock()
	expired := time.Now().After(j.sessionExpiry)
	j.sessionMutex.RUnlock()

	if expired {
		log.Info("Session expired, re-authenticating with Jellyseerr")
		return j.login()
	}

	return nil
}

// doRequest performs an HTTP request with proper headers
func (j *Jellyseerr) doRequest(method, path string, body any) (*http.Response, error) {
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

	// For POST requests (creating content), use credentials if available
	// For GET requests (checking status), use API key
	if j.useCredentials && method == "POST" {
		// Ensure we have a valid session
		if err := j.ensureAuthenticated(); err != nil {
			return nil, fmt.Errorf("failed to authenticate: %w", err)
		}
		// Cookie is automatically sent via cookie jar
	} else {
		// Use API key for GET requests or when credentials not configured
		req.Header.Set("X-Api-Key", j.config.APIKey)
	}

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
	defer func() { _ = resp.Body.Close() }()

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

	// Add user ID only if using API key (not credentials)
	// When using credentials, the request is automatically made as the logged-in user
	if !j.useCredentials && j.config.UserID != "" {
		if userID, err := strconv.Atoi(j.config.UserID); err == nil {
			payload.UserID = &userID
		}
	}

	resp, err := j.doRequest("POST", "/request", payload)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)

	// Jellyseerr returns 201 for new requests, 200 for existing
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("jellyseerr returned status %d: %s", resp.StatusCode, string(body))
	}

	var result RequestResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode request response: %w", err)
	}

	// Store HTTP status to distinguish new (201) vs existing (200) requests
	result.HTTPStatus = resp.StatusCode

	return &result, nil
}

// RequestShow requests a TV show by TMDB ID
func (j *Jellyseerr) RequestShow(tmdbID int) (*RequestResponse, error) {
	payload := ShowRequest{
		MediaType: "tv",
		MediaID:   tmdbID,
		Seasons:   "all",
	}

	// Add user ID only if using API key (not credentials)
	// When using credentials, the request is automatically made as the logged-in user
	if !j.useCredentials && j.config.UserID != "" {
		if userID, err := strconv.Atoi(j.config.UserID); err == nil {
			payload.UserID = &userID
		}
	}

	resp, err := j.doRequest("POST", "/request", payload)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)

	// Jellyseerr returns 201 for new requests, 200 for existing
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("jellyseerr returned status %d: %s", resp.StatusCode, string(body))
	}

	var result RequestResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode request response: %w", err)
	}

	// Store HTTP status to distinguish new (201) vs existing (200) requests
	result.HTTPStatus = resp.StatusCode

	return &result, nil
}

// IsAlreadyRequested checks if content was already requested (not newly created)
func (result *RequestResponse) IsAlreadyRequested() bool {
	// HTTP 200 = already exists, HTTP 201 = newly created
	return result.HTTPStatus == http.StatusOK
}

// MediaInfo represents media information from Jellyseerr
type MediaInfo struct {
	MediaInfo struct {
		Status int `json:"status"`
	} `json:"mediaInfo"`
}

// HasMediaInfo checks if the media has an associated media info (requested/available)
func (m *MediaInfo) HasMediaInfo() bool {
	return m.MediaInfo.Status > 0
}

// GetMovieInfo gets information about a movie from Jellyseerr
func (j *Jellyseerr) GetMovieInfo(tmdbID int) (*MediaInfo, error) {
	path := fmt.Sprintf("/movie/%d", tmdbID)
	resp, err := j.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		// Movie not found in Jellyseerr = not requested
		return &MediaInfo{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jellyseerr returned status %d: %s", resp.StatusCode, string(body))
	}

	var mediaInfo MediaInfo
	if err := json.NewDecoder(resp.Body).Decode(&mediaInfo); err != nil {
		return nil, fmt.Errorf("failed to decode media info: %w", err)
	}

	return &mediaInfo, nil
}

// GetShowInfo gets information about a TV show from Jellyseerr
func (j *Jellyseerr) GetShowInfo(tmdbID int) (*MediaInfo, error) {
	path := fmt.Sprintf("/tv/%d", tmdbID)
	resp, err := j.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		// Show not found in Jellyseerr = not requested
		return &MediaInfo{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jellyseerr returned status %d: %s", resp.StatusCode, string(body))
	}

	var mediaInfo MediaInfo
	if err := json.NewDecoder(resp.Body).Decode(&mediaInfo); err != nil {
		return nil, fmt.Errorf("failed to decode media info: %w", err)
	}

	return &mediaInfo, nil
}
