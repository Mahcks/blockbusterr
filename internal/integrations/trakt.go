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
	TraktAPIBaseURL = "https://api.trakt.tv"
	TraktAPIVersion = "2"
)

// Trakt is the client for interacting with Trakt API
type Trakt struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// TraktConfig holds configuration for Trakt client
type TraktConfig struct {
	ClientID     string
	ClientSecret string
}

// NewTrakt creates a new Trakt API client
func NewTrakt(config TraktConfig) *Trakt {
	return &Trakt{
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request to the Trakt API
func (t *Trakt) doRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", TraktAPIBaseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", TraktAPIVersion)
	req.Header.Set("trakt-api-key", t.clientID)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
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

// Movie represents a Trakt movie
type Movie struct {
	Title    string   `json:"title"`
	Year     int      `json:"year"`
	IDs      IDs      `json:"ids"`
	Genres   []string `json:"genres"`
	Language string   `json:"language"`
	Country  string   `json:"country"`
	Runtime  int      `json:"runtime"` // in minutes
}

// Show represents a Trakt TV show
type Show struct {
	Title    string   `json:"title"`
	Year     int      `json:"year"`
	IDs      IDs      `json:"ids"`
	Genres   []string `json:"genres"`
	Language string   `json:"language"`
	Country  string   `json:"country"`
	Runtime  int      `json:"runtime"` // in minutes
	Network  string   `json:"network"`
}

// IDs contains various IDs for a media item
type IDs struct {
	Trakt int    `json:"trakt"`
	Slug  string `json:"slug"`
	IMDB  string `json:"imdb"`
	TMDB  int    `json:"tmdb"`
	TVDB  int    `json:"tvdb"`
}

// GetTrendingMovies returns trending movies
func (t *Trakt) GetTrendingMovies(ctx context.Context, limit int) ([]TrendingMovie, error) {
	endpoint := fmt.Sprintf("/movies/trending?limit=%d", limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var movies []TrendingMovie
	if err := json.NewDecoder(resp.Body).Decode(&movies); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return movies, nil
}

// GetTrendingShows returns trending TV shows
func (t *Trakt) GetTrendingShows(ctx context.Context, limit int) ([]TrendingShow, error) {
	endpoint := fmt.Sprintf("/shows/trending?limit=%d", limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var shows []TrendingShow
	if err := json.NewDecoder(resp.Body).Decode(&shows); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return shows, nil
}

// GetPopularMovies returns popular movies
func (t *Trakt) GetPopularMovies(ctx context.Context, limit int) ([]Movie, error) {
	endpoint := fmt.Sprintf("/movies/popular?limit=%d", limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var movies []Movie
	if err := json.NewDecoder(resp.Body).Decode(&movies); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return movies, nil
}

// GetPopularShows returns popular TV shows
func (t *Trakt) GetPopularShows(ctx context.Context, limit int) ([]Show, error) {
	endpoint := fmt.Sprintf("/shows/popular?limit=%d", limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var shows []Show
	if err := json.NewDecoder(resp.Body).Decode(&shows); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return shows, nil
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var results []BoxOfficeMovie
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetFavoritedMovies returns the most favorited movies
func (t *Trakt) GetFavoritedMovies(ctx context.Context, period string, limit int) ([]FavoritedMovie, error) {
	endpoint := fmt.Sprintf("/movies/favorited/%s?limit=%d", period, limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var results []FavoritedMovie
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetPlayedMovies returns the most played movies
func (t *Trakt) GetPlayedMovies(ctx context.Context, period string, limit int) ([]PlayedMovie, error) {
	endpoint := fmt.Sprintf("/movies/played/%s?limit=%d", period, limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var results []PlayedMovie
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetWatchedMovies returns the most watched movies
func (t *Trakt) GetWatchedMovies(ctx context.Context, period string, limit int) ([]WatchedMovie, error) {
	endpoint := fmt.Sprintf("/movies/watched/%s?limit=%d", period, limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var results []WatchedMovie
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetCollectedMovies returns the most collected movies
func (t *Trakt) GetCollectedMovies(ctx context.Context, period string, limit int) ([]CollectedMovie, error) {
	endpoint := fmt.Sprintf("/movies/collected/%s?limit=%d", period, limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var results []CollectedMovie
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetAnticipatedMovies returns the most anticipated movies
func (t *Trakt) GetAnticipatedMovies(ctx context.Context, limit int) ([]AnticipatedMovie, error) {
	endpoint := fmt.Sprintf("/movies/anticipated?limit=%d", limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var results []AnticipatedMovie
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetFavoritedShows returns the most favorited TV shows
func (t *Trakt) GetFavoritedShows(ctx context.Context, period string, limit int) ([]FavoritedShow, error) {
	endpoint := fmt.Sprintf("/shows/favorited/%s?limit=%d", period, limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var results []FavoritedShow
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetPlayedShows returns the most played TV shows
func (t *Trakt) GetPlayedShows(ctx context.Context, period string, limit int) ([]PlayedShow, error) {
	endpoint := fmt.Sprintf("/shows/played/%s?limit=%d", period, limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var results []PlayedShow
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetWatchedShows returns the most watched TV shows
func (t *Trakt) GetWatchedShows(ctx context.Context, period string, limit int) ([]WatchedShow, error) {
	endpoint := fmt.Sprintf("/shows/watched/%s?limit=%d", period, limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var results []WatchedShow
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetCollectedShows returns the most collected TV shows
func (t *Trakt) GetCollectedShows(ctx context.Context, period string, limit int) ([]CollectedShow, error) {
	endpoint := fmt.Sprintf("/shows/collected/%s?limit=%d", period, limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var results []CollectedShow
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}

// GetAnticipatedShows returns the most anticipated TV shows
func (t *Trakt) GetAnticipatedShows(ctx context.Context, limit int) ([]AnticipatedShow, error) {
	endpoint := fmt.Sprintf("/shows/anticipated?limit=%d", limit)

	resp, err := t.doRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var results []AnticipatedShow
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	var networks []Network
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return networks, nil
}
