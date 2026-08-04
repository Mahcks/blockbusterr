package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type TMDBListItems struct {
	Name   string
	Movies []Movie
	Shows  []Show
}

type TMDBRequestToken struct {
	Success      bool   `json:"success"`
	RequestToken string `json:"request_token"`
	ExpiresAt    string `json:"expires_at"`
}

type TMDBSession struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id"`
}

type TMDBAccount struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

type tmdbMixedResult struct {
	MediaType     string   `json:"media_type"`
	ID            int      `json:"id"`
	Title         string   `json:"title"`
	Name          string   `json:"name"`
	ReleaseDate   string   `json:"release_date"`
	FirstAirDate  string   `json:"first_air_date"`
	GenreIDs      []int    `json:"genre_ids"`
	Language      string   `json:"original_language"`
	OriginCountry []string `json:"origin_country"`
	Overview      string   `json:"overview"`
	VoteAverage   float64  `json:"vote_average"`
	VoteCount     int      `json:"vote_count"`
}

func (t *TMDB) GetListItems(ctx context.Context, listID string, watchlist bool, limit int) (TMDBListItems, error) {
	if watchlist {
		return t.getWatchlist(ctx, limit)
	}
	result := TMDBListItems{}
	for page := 1; len(result.Movies)+len(result.Shows) < limit; page++ {
		var response struct {
			Name       string            `json:"name"`
			Page       int               `json:"page"`
			TotalPages int               `json:"total_pages"`
			Results    []tmdbMixedResult `json:"results"`
		}
		if err := t.getFrom(ctx, TMDBAPIv4BaseURL, "/list/"+url.PathEscape(listID), url.Values{"page": {strconv.Itoa(page)}, "language": {"en-US"}}, &response); err != nil {
			return TMDBListItems{}, err
		}
		result.Name = response.Name
		appendTMDBMixed(&result, response.Results, limit)
		if page >= response.TotalPages || len(response.Results) == 0 {
			break
		}
	}
	t.enrichShows(ctx, result.Shows)
	return result, nil
}

func (t *TMDB) getWatchlist(ctx context.Context, limit int) (TMDBListItems, error) {
	if t.sessionID == "" || t.accountID <= 0 {
		return TMDBListItems{}, fmt.Errorf("TMDB account is not authorized")
	}
	result := TMDBListItems{Name: "TMDB Watchlist"}
	for _, media := range []string{"movies", "tv"} {
		for page := 1; ; page++ {
			var response struct {
				Page       int               `json:"page"`
				TotalPages int               `json:"total_pages"`
				Results    []tmdbMixedResult `json:"results"`
			}
			query := url.Values{"page": {strconv.Itoa(page)}, "language": {"en-US"}, "sort_by": {"created_at.asc"}, "session_id": {t.sessionID}}
			if err := t.getFrom(ctx, TMDBAPIBaseURL, fmt.Sprintf("/account/%d/watchlist/%s", t.accountID, media), query, &response); err != nil {
				return TMDBListItems{}, err
			}
			for index := range response.Results {
				if media == "tv" {
					response.Results[index].MediaType = "tv"
				} else {
					response.Results[index].MediaType = "movie"
				}
			}
			appendTMDBMixed(&result, response.Results, limit)
			mediaFull := (media == "movies" && len(result.Movies) >= limit) || (media == "tv" && len(result.Shows) >= limit)
			if page >= response.TotalPages || len(response.Results) == 0 || mediaFull {
				break
			}
		}
	}
	t.enrichShows(ctx, result.Shows)
	return result, nil
}

func appendTMDBMixed(result *TMDBListItems, items []tmdbMixedResult, limit int) {
	for _, item := range items {
		switch item.MediaType {
		case "movie":
			if len(result.Movies) < limit {
				result.Movies = append(result.Movies, Movie{Title: item.Title, Year: yearFromDate(item.ReleaseDate), IDs: IDs{TMDB: item.ID}, Genres: genreNames(item.GenreIDs, tmdbMovieGenres), Language: item.Language, Overview: item.Overview, Rating: item.VoteAverage, Votes: item.VoteCount})
			}
		case "tv":
			if len(result.Shows) < limit {
				result.Shows = append(result.Shows, Show{Title: item.Name, Year: yearFromDate(item.FirstAirDate), IDs: IDs{TMDB: item.ID}, Genres: genreNames(item.GenreIDs, tmdbShowGenres), Language: item.Language, Country: joinCountry(item.OriginCountry), Overview: item.Overview, Rating: item.VoteAverage, Votes: item.VoteCount})
			}
		}
	}
}

func joinCountry(countries []string) string {
	if len(countries) == 0 {
		return ""
	}
	return strings.ToLower(countries[0])
}

func (t *TMDB) CreateRequestToken(ctx context.Context) (TMDBRequestToken, error) {
	var token TMDBRequestToken
	err := t.getFrom(ctx, TMDBAPIBaseURL, "/authentication/token/new", nil, &token)
	return token, err
}

func (t *TMDB) CreateSession(ctx context.Context, requestToken string) (TMDBSession, error) {
	var session TMDBSession
	err := t.post(ctx, "/authentication/session/new", url.Values{"api_key": {t.apiKey}}, map[string]string{"request_token": requestToken}, &session)
	return session, err
}

func (t *TMDB) GetAccount(ctx context.Context, sessionID string) (TMDBAccount, error) {
	var account TMDBAccount
	err := t.getFrom(ctx, TMDBAPIBaseURL, "/account", url.Values{"session_id": {sessionID}}, &account)
	return account, err
}

func (t *TMDB) getFrom(ctx context.Context, baseURL, endpoint string, query url.Values, target any) error {
	if query == nil {
		query = url.Values{}
	}
	if t.apiKey == "" {
		return fmt.Errorf("TMDB API key is not configured")
	}
	query.Set("api_key", t.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return err
	}
	return t.do(req, target)
}

func (t *TMDB) post(ctx context.Context, endpoint string, query url.Values, payload, target any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TMDBAPIBaseURL+endpoint+"?"+query.Encode(), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return t.do(req, target)
}

func (t *TMDB) do(req *http.Request, target any) error {
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("TMDB request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to parse TMDB response: %w", err)
	}
	return nil
}
