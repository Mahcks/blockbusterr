package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const MDBListAPIBaseURL = "https://api.mdblist.com"

type MDBList struct {
	apiKey     string
	httpClient *http.Client
}

type MDBListItems struct {
	Name   string
	Movies []Movie
	Shows  []Show
}

type MDBListConfig struct{ APIKey string }

func NewMDBList(config MDBListConfig) *MDBList {
	return &MDBList{apiKey: config.APIKey, httpClient: &http.Client{Timeout: 15 * time.Second}}
}

func (client *MDBList) GetListItems(ctx context.Context, owner, identifier string, watchlist bool, mediaType string, limit int) (MDBListItems, error) {
	if client.apiKey == "" {
		return MDBListItems{}, fmt.Errorf("MDBList API key is not configured")
	}
	endpoint := "/watchlist/items"
	name := "MDBList Watchlist"
	if !watchlist {
		name = identifier
		if owner != "" {
			endpoint = "/lists/" + url.PathEscape(owner) + "/" + url.PathEscape(identifier) + "/items"
		} else if _, err := strconv.Atoi(identifier); err == nil {
			endpoint = "/lists/" + url.PathEscape(identifier) + "/items"
		} else {
			endpoint = "/lists/official/" + url.PathEscape(identifier) + "/items"
		}
	}

	result := MDBListItems{Name: name}
	cursor := ""
	for {
		pageSize := min(limit, 1000)
		query := url.Values{"apikey": {client.apiKey}, "limit": {strconv.Itoa(pageSize)}}
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		var response struct {
			Movies     []mdbListItem `json:"movies"`
			Shows      []mdbListItem `json:"shows"`
			NextCursor string        `json:"next_cursor"`
			Pagination struct {
				NextCursor string `json:"next_cursor"`
			} `json:"pagination"`
		}
		if err := client.get(ctx, endpoint, query, &response); err != nil {
			return MDBListItems{}, err
		}
		appendMDBListItems(&result, response.Movies, response.Shows, limit)
		if listMediaCount(mediaType, len(result.Movies), len(result.Shows)) >= limit {
			break
		}
		if response.NextCursor == "" {
			response.NextCursor = response.Pagination.NextCursor
		}
		if response.NextCursor == "" || response.NextCursor == cursor || len(response.Movies)+len(response.Shows) == 0 {
			break
		}
		cursor = response.NextCursor
	}
	keepRequestedListMedia(mediaType, &result.Movies, &result.Shows)
	return result, nil
}

func (client *MDBList) Validate(ctx context.Context) error {
	return client.get(ctx, "/user", url.Values{"apikey": {client.apiKey}}, &struct{}{})
}

type mdbListItem struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	IMDB     string `json:"imdb_id"`
	TVDB     int    `json:"tvdb_id"`
	Language string `json:"language"`
	Country  string `json:"country"`
	Year     int    `json:"release_year"`
	IDs      struct {
		IMDB string `json:"imdb"`
		TMDB int    `json:"tmdb"`
		TVDB int    `json:"tvdb"`
	} `json:"ids"`
}

func appendMDBListItems(result *MDBListItems, movies, shows []mdbListItem, limit int) {
	for _, item := range movies {
		if len(result.Movies) >= limit {
			break
		}
		if item.IDs.TMDB == 0 {
			item.IDs.TMDB = item.ID
		}
		if item.IDs.IMDB == "" {
			item.IDs.IMDB = item.IMDB
		}
		result.Movies = append(result.Movies, Movie{Title: item.Title, Year: item.Year, Language: item.Language, Country: item.Country, IDs: IDs{TMDB: item.IDs.TMDB, IMDB: item.IDs.IMDB}})
	}
	for _, item := range shows {
		if len(result.Shows) >= limit {
			break
		}
		if item.IDs.TMDB == 0 {
			item.IDs.TMDB = item.ID
		}
		if item.IDs.TVDB == 0 {
			item.IDs.TVDB = item.TVDB
		}
		if item.IDs.IMDB == "" {
			item.IDs.IMDB = item.IMDB
		}
		result.Shows = append(result.Shows, Show{Title: item.Title, Year: item.Year, Language: item.Language, Country: item.Country, IDs: IDs{TMDB: item.IDs.TMDB, TVDB: item.IDs.TVDB, IMDB: item.IDs.IMDB}})
	}
}

func (client *MDBList) get(ctx context.Context, endpoint string, query url.Values, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, MDBListAPIBaseURL+endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return err
	}
	response, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("MDBList request failed: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("MDBList API returned status %d", response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to parse MDBList response: %w", err)
	}
	return nil
}
