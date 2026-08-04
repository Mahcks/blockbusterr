package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	TMDBAPIBaseURL   = "https://api.themoviedb.org/3"
	TMDBAPIv4BaseURL = "https://api.themoviedb.org/4"
	TMDBImageBaseURL = "https://image.tmdb.org/t/p/w500"
)

// TMDB is the client for interacting with TMDB API
type TMDB struct {
	apiKey     string
	sessionID  string
	accountID  int
	httpClient *http.Client
}

// TMDBConfig holds configuration for TMDB client
type TMDBConfig struct {
	APIKey    string
	SessionID string
	AccountID int
}

// NewTMDB creates a new TMDB API client
func NewTMDB(config TMDBConfig) *TMDB {
	return &TMDB{
		apiKey:    config.APIKey,
		sessionID: config.SessionID,
		accountID: config.AccountID,
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

type tmdbListResponse[T any] struct {
	Page       int `json:"page"`
	TotalPages int `json:"total_pages"`
	Results    []T `json:"results"`
}

type tmdbMovieResult struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	ReleaseDate      string  `json:"release_date"`
	GenreIDs         []int   `json:"genre_ids"`
	OriginalLanguage string  `json:"original_language"`
	Overview         string  `json:"overview"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
}

type tmdbShowResult struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	FirstAirDate     string   `json:"first_air_date"`
	GenreIDs         []int    `json:"genre_ids"`
	OriginCountry    []string `json:"origin_country"`
	OriginalLanguage string   `json:"original_language"`
	Overview         string   `json:"overview"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
}

type tmdbShowDetails struct {
	EpisodeRunTime []int `json:"episode_run_time"`
	Networks       []struct {
		Name string `json:"name"`
	} `json:"networks"`
	ExternalIDs struct {
		IMDB string `json:"imdb_id"`
		TVDB int    `json:"tvdb_id"`
	} `json:"external_ids"`
}

var tmdbMovieGenres = map[int]string{12: "adventure", 14: "fantasy", 16: "animation", 18: "drama", 27: "horror", 28: "action", 35: "comedy", 36: "history", 37: "western", 53: "thriller", 80: "crime", 99: "documentary", 878: "science-fiction", 9648: "mystery", 10402: "music", 10749: "romance", 10751: "family", 10752: "war", 10770: "tv-movie"}
var tmdbShowGenres = map[int]string{16: "animation", 18: "drama", 35: "comedy", 37: "western", 80: "crime", 99: "documentary", 9648: "mystery", 10751: "family", 10759: "action-adventure", 10762: "kids", 10763: "news", 10764: "reality", 10765: "sci-fi-fantasy", 10766: "soap", 10767: "talk", 10768: "war-politics"}

func (t *TMDB) get(ctx context.Context, endpoint string, query url.Values, target any) error {
	if t.apiKey == "" {
		return fmt.Errorf("TMDB API key is not configured")
	}
	query.Set("api_key", t.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, TMDBAPIBaseURL+endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return fmt.Errorf("failed to create TMDB request: %w", err)
	}
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("TMDB request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to parse TMDB response: %w", err)
	}
	return nil
}

func yearFromDate(value string) int {
	if len(value) < 4 {
		return 0
	}
	year, _ := strconv.Atoi(value[:4])
	return year
}

func genreNames(ids []int, names map[int]string) []string {
	genres := make([]string, 0, len(ids))
	for _, id := range ids {
		if name := names[id]; name != "" {
			genres = append(genres, name)
		}
	}
	return genres
}

func (t *TMDB) getMovies(ctx context.Context, endpoint string, limit int) ([]Movie, error) {
	result := make([]Movie, 0, limit)
	for page := 1; len(result) < limit; page++ {
		var response tmdbListResponse[tmdbMovieResult]
		if err := t.get(ctx, endpoint, url.Values{"page": {strconv.Itoa(page)}, "language": {"en-US"}}, &response); err != nil {
			return nil, err
		}
		for _, item := range response.Results {
			result = append(result, Movie{Title: item.Title, Year: yearFromDate(item.ReleaseDate), IDs: IDs{TMDB: item.ID}, Genres: genreNames(item.GenreIDs, tmdbMovieGenres), Language: item.OriginalLanguage, Overview: item.Overview, Rating: item.VoteAverage, Votes: item.VoteCount})
			if len(result) == limit {
				break
			}
		}
		if page >= response.TotalPages || len(response.Results) == 0 {
			break
		}
	}
	return result, nil
}

func (t *TMDB) getShows(ctx context.Context, endpoint string, limit int) ([]Show, error) {
	result := make([]Show, 0, limit)
	for page := 1; len(result) < limit; page++ {
		var response tmdbListResponse[tmdbShowResult]
		if err := t.get(ctx, endpoint, url.Values{"page": {strconv.Itoa(page)}, "language": {"en-US"}}, &response); err != nil {
			return nil, err
		}
		for _, item := range response.Results {
			show := Show{Title: item.Name, Year: yearFromDate(item.FirstAirDate), IDs: IDs{TMDB: item.ID}, Genres: genreNames(item.GenreIDs, tmdbShowGenres), Language: item.OriginalLanguage, Country: strings.ToLower(strings.Join(item.OriginCountry, ",")), Overview: item.Overview, Rating: item.VoteAverage, Votes: item.VoteCount}
			result = append(result, show)
			if len(result) == limit {
				break
			}
		}
		if page >= response.TotalPages || len(response.Results) == 0 {
			break
		}
	}
	t.enrichShows(ctx, result)
	return result, nil
}

func (t *TMDB) enrichShows(ctx context.Context, shows []Show) {
	semaphore := make(chan struct{}, 8)
	var waitGroup sync.WaitGroup
	for index := range shows {
		waitGroup.Go(func() {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			var details tmdbShowDetails
			if err := t.get(ctx, fmt.Sprintf("/tv/%d", shows[index].IDs.TMDB), url.Values{"append_to_response": {"external_ids"}, "language": {"en-US"}}, &details); err != nil {
				return
			}
			shows[index].IDs.IMDB = details.ExternalIDs.IMDB
			shows[index].IDs.TVDB = details.ExternalIDs.TVDB
			if len(details.EpisodeRunTime) > 0 {
				shows[index].Runtime = details.EpisodeRunTime[0]
			}
			if len(details.Networks) > 0 {
				shows[index].Network = details.Networks[0].Name
			}
		})
	}
	waitGroup.Wait()
}

func (t *TMDB) GetTrendingMovies(ctx context.Context, limit int) ([]Movie, error) {
	return t.getMovies(ctx, "/trending/movie/day", limit)
}

func (t *TMDB) GetPopularMovies(ctx context.Context, limit int) ([]Movie, error) {
	return t.getMovies(ctx, "/movie/popular", limit)
}

func (t *TMDB) GetTrendingShows(ctx context.Context, limit int) ([]Show, error) {
	return t.getShows(ctx, "/trending/tv/day", limit)
}

func (t *TMDB) GetPopularShows(ctx context.Context, limit int) ([]Show, error) {
	return t.getShows(ctx, "/tv/popular", limit)
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
	defer func() { _ = resp.Body.Close() }()

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
	defer func() { _ = resp.Body.Close() }()

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
