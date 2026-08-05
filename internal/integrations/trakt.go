package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	TraktAPIBaseURL      = "https://api.trakt.tv"
	TraktAPIVersion      = "2"
	TraktDefaultPageSize = 100 // Max items per page for most Trakt endpoints
)

// Trakt is the client for interacting with Trakt API
type Trakt struct {
	clientID     string
	clientSecret string
	accessToken  string
	refreshToken string
	tokenExpires int64
	onToken      func(TraktToken) error
	loadToken    func() TraktToken
	httpClient   *http.Client
}

// ponytail: one process-wide lock is sufficient; use per-account locks only if refresh contention becomes measurable.
var traktTokenMu sync.Mutex

// TraktConfig holds configuration for Trakt client
type TraktConfig struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
	TokenExpires int64
	OnToken      func(TraktToken) error
	LoadToken    func() TraktToken
}

// NewTrakt creates a new Trakt API client
func NewTrakt(config TraktConfig) *Trakt {
	return &Trakt{
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		accessToken:  config.AccessToken,
		refreshToken: config.RefreshToken,
		tokenExpires: config.TokenExpires,
		onToken:      config.OnToken,
		loadToken:    config.LoadToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request to the Trakt API
func (t *Trakt) doRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	accessToken, err := t.accessTokenForRequest(ctx)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s%s", TraktAPIBaseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", TraktAPIVersion)
	req.Header.Set("trakt-api-key", t.clientID)
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// doRequestPaginated performs paginated HTTP requests to the Trakt API
// It fetches multiple pages if needed to reach the requested limit
func (t *Trakt) doRequestPaginated(ctx context.Context, endpoint string, limit int) ([][]byte, error) {
	var allResults [][]byte

	page := 1
	remaining := limit
	pageSize := TraktDefaultPageSize

	for remaining > 0 {
		// Calculate how many items to request this page
		requestLimit := min(remaining, pageSize)

		// Build paginated endpoint
		separator := "?"
		if strings.Contains(endpoint, "?") {
			separator = "&"
		}
		paginatedEndpoint := fmt.Sprintf("%s%spage=%d&limit=%d", endpoint, separator, page, requestLimit)

		resp, err := t.doRequest(ctx, http.MethodGet, paginatedEndpoint, nil)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			err := newAPIError("Trakt", resp)
			_ = resp.Body.Close()
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		allResults = append(allResults, body)

		// Check if we got fewer items than requested (end of results)
		var items []json.RawMessage
		if err := json.Unmarshal(body, &items); err != nil {
			return nil, fmt.Errorf("failed to check response length: %w", err)
		}

		if len(items) < requestLimit {
			// No more items available
			break
		}

		remaining -= len(items)
		page++
	}

	return allResults, nil
}

// TrendingMovie represents a trending movie from Trakt
type TrendingMovie struct {
	Watchers int   `json:"watchers"`
	Movie    Movie `json:"movie"`
}

// TrendingShow represents a trending TV show from Trakt
type TrendingShow struct {
	Watchers int  `json:"watchers"`
	Show     Show `json:"show"`
}

// BoxOfficeMovie represents a box office movie from Trakt
type BoxOfficeMovie struct {
	Revenue int   `json:"revenue"`
	Movie   Movie `json:"movie"`
}

// FavoritedMovie represents a favorited movie from Trakt
type FavoritedMovie struct {
	UserCount int   `json:"user_count"`
	Movie     Movie `json:"movie"`
}

// PlayedMovie represents a played movie from Trakt
type PlayedMovie struct {
	WatcherCount   int   `json:"watcher_count"`
	PlayCount      int   `json:"play_count"`
	CollectedCount int   `json:"collected_count"`
	Movie          Movie `json:"movie"`
}

// WatchedMovie represents a watched movie from Trakt
type WatchedMovie struct {
	WatcherCount   int   `json:"watcher_count"`
	PlayCount      int   `json:"play_count"`
	CollectedCount int   `json:"collected_count"`
	Movie          Movie `json:"movie"`
}

// CollectedMovie represents a collected movie from Trakt
type CollectedMovie struct {
	WatcherCount   int   `json:"watcher_count"`
	PlayCount      int   `json:"play_count"`
	CollectedCount int   `json:"collected_count"`
	Movie          Movie `json:"movie"`
}

// AnticipatedMovie represents an anticipated movie from Trakt
type AnticipatedMovie struct {
	ListCount int   `json:"list_count"`
	Movie     Movie `json:"movie"`
}

// FavoritedShow represents a favorited TV show from Trakt
type FavoritedShow struct {
	UserCount int  `json:"user_count"`
	Show      Show `json:"show"`
}

// PlayedShow represents a played TV show from Trakt
type PlayedShow struct {
	WatcherCount   int  `json:"watcher_count"`
	PlayCount      int  `json:"play_count"`
	CollectedCount int  `json:"collected_count"`
	CollectorCount int  `json:"collector_count"`
	Show           Show `json:"show"`
}

// WatchedShow represents a watched TV show from Trakt
type WatchedShow struct {
	WatcherCount   int  `json:"watcher_count"`
	PlayCount      int  `json:"play_count"`
	CollectedCount int  `json:"collected_count"`
	CollectorCount int  `json:"collector_count"`
	Show           Show `json:"show"`
}

// CollectedShow represents a collected TV show from Trakt
type CollectedShow struct {
	WatcherCount   int  `json:"watcher_count"`
	PlayCount      int  `json:"play_count"`
	CollectedCount int  `json:"collected_count"`
	CollectorCount int  `json:"collector_count"`
	Show           Show `json:"show"`
}

// AnticipatedShow represents an anticipated TV show from Trakt
type AnticipatedShow struct {
	ListCount int  `json:"list_count"`
	Show      Show `json:"show"`
}

// GetTrendingMovies returns trending movies with pagination support
func (t *Trakt) GetTrendingMovies(ctx context.Context, limit int) ([]TrendingMovie, error) {
	endpoint := "/movies/trending?extended=full"

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allMovies []TrendingMovie
	for _, page := range pages {
		var movies []TrendingMovie
		if err := json.Unmarshal(page, &movies); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allMovies = append(allMovies, movies...)
	}

	// Trim to exact limit if we got more
	if len(allMovies) > limit {
		allMovies = allMovies[:limit]
	}

	return allMovies, nil
}

// GetTrendingShows returns trending TV shows with pagination support
func (t *Trakt) GetTrendingShows(ctx context.Context, limit int) ([]TrendingShow, error) {
	endpoint := "/shows/trending?extended=full"

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allShows []TrendingShow
	for _, page := range pages {
		var shows []TrendingShow
		if err := json.Unmarshal(page, &shows); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allShows = append(allShows, shows...)
	}

	if len(allShows) > limit {
		allShows = allShows[:limit]
	}

	return allShows, nil
}

// GetPopularMovies returns popular movies with pagination support
func (t *Trakt) GetPopularMovies(ctx context.Context, limit int) ([]Movie, error) {
	endpoint := "/movies/popular?extended=full"

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allMovies []Movie
	for _, page := range pages {
		var movies []Movie
		if err := json.Unmarshal(page, &movies); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allMovies = append(allMovies, movies...)
	}

	if len(allMovies) > limit {
		allMovies = allMovies[:limit]
	}

	return allMovies, nil
}

// GetPopularShows returns popular TV shows with pagination support
func (t *Trakt) GetPopularShows(ctx context.Context, limit int) ([]Show, error) {
	endpoint := "/shows/popular?extended=full"

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allShows []Show
	for _, page := range pages {
		var shows []Show
		if err := json.Unmarshal(page, &shows); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allShows = append(allShows, shows...)
	}

	if len(allShows) > limit {
		allShows = allShows[:limit]
	}

	return allShows, nil
}

// SearchResult represents a search result from Trakt
type SearchResult struct {
	Type  string  `json:"type"`
	Score float64 `json:"score"`
	Movie *Movie  `json:"movie,omitempty"`
	Show  *Show   `json:"show,omitempty"`
}

// Search performs a search across Trakt content
func (t *Trakt) Search(ctx context.Context, query string, searchType string, limit int) ([]SearchResult, error) {
	endpoint := fmt.Sprintf("/search/%s?query=%s&limit=%d", searchType, query, limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, newAPIError("Trakt", resp)
	}

	var results []SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetBoxOfficeMovies returns current box office movies
func (t *Trakt) GetBoxOfficeMovies(ctx context.Context, limit int) ([]BoxOfficeMovie, error) {
	endpoint := fmt.Sprintf("/movies/boxoffice?extended=full&limit=%d", limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, newAPIError("Trakt", resp)
	}

	var results []BoxOfficeMovie
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetFavoritedMovies returns the most favorited movies with pagination support
func (t *Trakt) GetFavoritedMovies(ctx context.Context, period string, limit int) ([]FavoritedMovie, error) {
	endpoint := fmt.Sprintf("/movies/favorited/%s?extended=full", period)

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allResults []FavoritedMovie
	for _, page := range pages {
		var results []FavoritedMovie
		if err := json.Unmarshal(page, &results); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

// GetPlayedMovies returns the most played movies with pagination support
func (t *Trakt) GetPlayedMovies(ctx context.Context, period string, limit int) ([]PlayedMovie, error) {
	endpoint := fmt.Sprintf("/movies/played/%s?extended=full", period)

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allResults []PlayedMovie
	for _, page := range pages {
		var results []PlayedMovie
		if err := json.Unmarshal(page, &results); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

// GetWatchedMovies returns the most watched movies with pagination support
func (t *Trakt) GetWatchedMovies(ctx context.Context, period string, limit int) ([]WatchedMovie, error) {
	endpoint := fmt.Sprintf("/movies/watched/%s?extended=full", period)

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allResults []WatchedMovie
	for _, page := range pages {
		var results []WatchedMovie
		if err := json.Unmarshal(page, &results); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

// GetCollectedMovies returns the most collected movies with pagination support
func (t *Trakt) GetCollectedMovies(ctx context.Context, period string, limit int) ([]CollectedMovie, error) {
	endpoint := fmt.Sprintf("/movies/collected/%s?extended=full", period)

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allResults []CollectedMovie
	for _, page := range pages {
		var results []CollectedMovie
		if err := json.Unmarshal(page, &results); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

// GetAnticipatedMovies returns the most anticipated movies with pagination support
func (t *Trakt) GetAnticipatedMovies(ctx context.Context, limit int) ([]AnticipatedMovie, error) {
	endpoint := "/movies/anticipated?extended=full"

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allResults []AnticipatedMovie
	for _, page := range pages {
		var results []AnticipatedMovie
		if err := json.Unmarshal(page, &results); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

// GetFavoritedShows returns the most favorited TV shows with pagination support
func (t *Trakt) GetFavoritedShows(ctx context.Context, period string, limit int) ([]FavoritedShow, error) {
	endpoint := fmt.Sprintf("/shows/favorited/%s?extended=full", period)

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allResults []FavoritedShow
	for _, page := range pages {
		var results []FavoritedShow
		if err := json.Unmarshal(page, &results); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

// GetPlayedShows returns the most played TV shows with pagination support
func (t *Trakt) GetPlayedShows(ctx context.Context, period string, limit int) ([]PlayedShow, error) {
	endpoint := fmt.Sprintf("/shows/played/%s?extended=full", period)

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allResults []PlayedShow
	for _, page := range pages {
		var results []PlayedShow
		if err := json.Unmarshal(page, &results); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

// GetWatchedShows returns the most watched TV shows with pagination support
func (t *Trakt) GetWatchedShows(ctx context.Context, period string, limit int) ([]WatchedShow, error) {
	endpoint := fmt.Sprintf("/shows/watched/%s?extended=full", period)

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allResults []WatchedShow
	for _, page := range pages {
		var results []WatchedShow
		if err := json.Unmarshal(page, &results); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

// GetCollectedShows returns the most collected TV shows with pagination support
func (t *Trakt) GetCollectedShows(ctx context.Context, period string, limit int) ([]CollectedShow, error) {
	endpoint := fmt.Sprintf("/shows/collected/%s?extended=full", period)

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allResults []CollectedShow
	for _, page := range pages {
		var results []CollectedShow
		if err := json.Unmarshal(page, &results); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

// GetAnticipatedShows returns the most anticipated TV shows with pagination support
func (t *Trakt) GetAnticipatedShows(ctx context.Context, limit int) ([]AnticipatedShow, error) {
	endpoint := "/shows/anticipated?extended=full"

	pages, err := t.doRequestPaginated(ctx, endpoint, limit)
	if err != nil {
		return nil, err
	}

	var allResults []AnticipatedShow
	for _, page := range pages {
		var results []AnticipatedShow
		if err := json.Unmarshal(page, &results); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

// Validate checks if the Trakt client credentials are valid
func (t *Trakt) Validate(ctx context.Context) error {
	if t.clientID == "" {
		return fmt.Errorf("client ID is required")
	}

	// Try to fetch trending movies as a validation check
	_, err := t.GetTrendingMovies(ctx, 1)
	return err
}

// Language represents a language from Trakt
type Language struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// Genre represents a genre from Trakt
type Genre struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Country represents a country from Trakt
type Country struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// Network represents a TV network from Trakt
type Network struct {
	Name    string `json:"name"`
	Country string `json:"country"`
}

// GetLanguages returns available languages for movies or shows
func (t *Trakt) GetLanguages(ctx context.Context, mediaType string) ([]Language, error) {
	endpoint := fmt.Sprintf("/languages/%s", mediaType)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, newAPIError("Trakt", resp)
	}

	var languages []Language
	if err := json.NewDecoder(resp.Body).Decode(&languages); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return languages, nil
}

// GetGenres returns available genres for movies or shows
func (t *Trakt) GetGenres(ctx context.Context, mediaType string) ([]Genre, error) {
	endpoint := fmt.Sprintf("/genres/%s", mediaType)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, newAPIError("Trakt", resp)
	}

	var genres []Genre
	if err := json.NewDecoder(resp.Body).Decode(&genres); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return genres, nil
}

// GetCountries returns available countries for movies or shows
func (t *Trakt) GetCountries(ctx context.Context, mediaType string) ([]Country, error) {
	endpoint := fmt.Sprintf("/countries/%s", mediaType)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, newAPIError("Trakt", resp)
	}

	var countries []Country
	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return countries, nil
}

// GetNetworks returns available TV networks
func (t *Trakt) GetNetworks(ctx context.Context) ([]Network, error) {
	endpoint := "/networks"

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, newAPIError("Trakt", resp)
	}

	var networks []Network
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return networks, nil
}
