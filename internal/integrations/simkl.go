package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const simklTrendingBaseURL = "https://data.simkl.in/discover/trending"

type Simkl struct {
	clientID   string
	httpClient *http.Client
}

type SimklConfig struct {
	ClientID string
}

func NewSimkl(config SimklConfig) *Simkl {
	return &Simkl{clientID: config.ClientID, httpClient: &http.Client{Timeout: 15 * time.Second}}
}

type flexibleInt int

func (value *flexibleInt) UnmarshalJSON(data []byte) error {
	text := strings.Trim(string(data), `"`)
	if text == "" || text == "null" {
		return nil
	}
	number, err := strconv.Atoi(text)
	if err != nil {
		return err
	}
	*value = flexibleInt(number)
	return nil
}

type simklRating struct {
	Rating float64 `json:"rating"`
	Votes  int     `json:"votes"`
}

type simklItem struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	IDs   struct {
		Simkl flexibleInt `json:"simkl_id"`
		Slug  string      `json:"slug"`
		IMDB  string      `json:"imdb"`
		TMDB  flexibleInt `json:"tmdb"`
		TVDB  flexibleInt `json:"tvdb"`
	} `json:"ids"`
	ReleaseDate      string                 `json:"release_date"`
	Watched          int                    `json:"watched"`
	Ratings          map[string]simklRating `json:"ratings"`
	Country          string                 `json:"country"`
	OriginalLanguage string                 `json:"original_language"`
	Runtime          string                 `json:"runtime"`
	Overview         string                 `json:"overview"`
	Genres           []string               `json:"genres"`
	Network          string                 `json:"network"`
}

func parseSimklRuntime(value string) int {
	var hours, minutes int
	_, _ = fmt.Sscanf(value, "%dh %dm", &hours, &minutes)
	if hours == 0 {
		_, _ = fmt.Sscanf(value, "%dm", &minutes)
	}
	return hours*60 + minutes
}

func simklYear(value string) int {
	for _, part := range strings.FieldsFunc(value, func(r rune) bool { return r == '/' || r == '-' }) {
		if len(part) == 4 {
			year, _ := strconv.Atoi(part)
			return year
		}
	}
	return 0
}

func (s *Simkl) trending(ctx context.Context, mediaType, timeframe string, limit int) ([]simklItem, error) {
	if s.clientID == "" {
		return nil, fmt.Errorf("Simkl client ID is not configured")
	}
	size := 100
	if limit > size {
		size = 500
	}
	if limit > size {
		limit = size
	}
	endpoint := fmt.Sprintf("%s/%s/%s_%d.json", simklTrendingBaseURL, mediaType, timeframe, size)
	query := url.Values{"client_id": {s.clientID}, "app-name": {"blockbusterr"}, "app-version": {"1"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Simkl request: %w", err)
	}
	req.Header.Set("User-Agent", "blockbusterr/1")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Simkl request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Simkl API returned status %d", resp.StatusCode)
	}
	var items []simklItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to parse Simkl response: %w", err)
	}
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func simklIDs(item simklItem) IDs {
	return IDs{Simkl: int(item.IDs.Simkl), Slug: item.IDs.Slug, IMDB: item.IDs.IMDB, TMDB: int(item.IDs.TMDB), TVDB: int(item.IDs.TVDB)}
}

func simklScore(item simklItem) simklRating {
	if rating, ok := item.Ratings["simkl"]; ok {
		return rating
	}
	return item.Ratings["imdb"]
}

func (s *Simkl) GetMovies(ctx context.Context, timeframe string, limit int) ([]Movie, error) {
	items, err := s.trending(ctx, "movies", timeframe, limit)
	if err != nil {
		return nil, err
	}
	movies := make([]Movie, len(items))
	for i, item := range items {
		rating := simklScore(item)
		movies[i] = Movie{Title: item.Title, Year: simklYear(item.ReleaseDate), IDs: simklIDs(item), Genres: item.Genres, Language: item.OriginalLanguage, Country: item.Country, Runtime: parseSimklRuntime(item.Runtime), Overview: item.Overview, Rating: rating.Rating, Votes: rating.Votes}
	}
	return movies, nil
}

func (s *Simkl) GetShows(ctx context.Context, timeframe string, limit int) ([]Show, error) {
	items, err := s.trending(ctx, "tv", timeframe, limit)
	if err != nil {
		return nil, err
	}
	shows := make([]Show, len(items))
	for i, item := range items {
		rating := simklScore(item)
		shows[i] = Show{Title: item.Title, Year: simklYear(item.ReleaseDate), IDs: simklIDs(item), Genres: item.Genres, Language: item.OriginalLanguage, Country: item.Country, Runtime: parseSimklRuntime(item.Runtime), Network: item.Network, Overview: item.Overview, Rating: rating.Rating, Votes: rating.Votes}
	}
	return shows, nil
}
