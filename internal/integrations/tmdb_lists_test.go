package integrations

import (
	"context"
	"net/http"
	"testing"
)

func TestTMDBPublicListPaginatesMixedMedia(t *testing.T) {
	client := NewTMDB(TMDBConfig{APIKey: "key"})
	pages := 0
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path == "/3/tv/20" {
			return jsonResponse(http.StatusOK, `{"episode_run_time":[42],"external_ids":{"tvdb_id":200}}`), nil
		}
		pages++
		if request.URL.Query().Get("api_key") != "key" {
			t.Fatal("missing API key")
		}
		if pages == 1 {
			return jsonResponse(http.StatusOK, `{"name":"Weekend","page":1,"total_pages":2,"results":[{"media_type":"movie","id":10,"title":"Movie","release_date":"2025-01-01"}]}`), nil
		}
		return jsonResponse(http.StatusOK, `{"name":"Weekend","page":2,"total_pages":2,"results":[{"media_type":"tv","id":20,"name":"Show","first_air_date":"2024-01-01"}]}`), nil
	})
	items, err := client.GetListItems(t.Context(), "55", false, 10)
	if err != nil {
		t.Fatal(err)
	}
	if items.Name != "Weekend" || len(items.Movies) != 1 || len(items.Shows) != 1 || items.Shows[0].IDs.TVDB != 200 {
		t.Fatalf("items = %#v", items)
	}
}

func TestTMDBWatchlistRequiresAuthorization(t *testing.T) {
	client := NewTMDB(TMDBConfig{APIKey: "key"})
	if _, err := client.GetListItems(t.Context(), "", true, 10); err == nil {
		t.Fatal("expected authorization error")
	}
}

func TestTMDBListErrors(t *testing.T) {
	for name, response := range map[string]*http.Response{
		"private":    jsonResponse(http.StatusUnauthorized, `{}`),
		"rate limit": jsonResponse(http.StatusTooManyRequests, `{}`),
		"malformed":  jsonResponse(http.StatusOK, `{`),
	} {
		t.Run(name, func(t *testing.T) {
			client := NewTMDB(TMDBConfig{APIKey: "key"})
			client.httpClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) { return response, nil })
			if _, err := client.GetListItems(t.Context(), "1", false, 10); err == nil {
				t.Fatal("expected error")
			}
		})
	}
	client := NewTMDB(TMDBConfig{APIKey: "key"})
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) { return nil, request.Context().Err() })
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := client.GetListItems(ctx, "1", false, 10); err == nil {
		t.Fatal("expected cancellation error")
	}
}

func TestTMDBConnectedWatchlist(t *testing.T) {
	client := NewTMDB(TMDBConfig{APIKey: "key", SessionID: "session", AccountID: 4})
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path == "/3/account/4/watchlist/movies" {
			return jsonResponse(http.StatusOK, `{"page":1,"total_pages":1,"results":[{"id":7,"title":"Saved"}]}`), nil
		}
		if request.URL.Path == "/3/account/4/watchlist/tv" {
			return jsonResponse(http.StatusOK, `{"page":1,"total_pages":1,"results":[]}`), nil
		}
		t.Fatalf("path = %s", request.URL.Path)
		return nil, nil
	})
	items, err := client.GetListItems(t.Context(), "", true, 10)
	if err != nil || len(items.Movies) != 1 || items.Movies[0].IDs.TMDB != 7 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}
