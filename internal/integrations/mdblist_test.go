package integrations

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestMDBListPaginatesAndNormalizes(t *testing.T) {
	client := NewMDBList(MDBListConfig{APIKey: "secret"})
	requests := 0
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.URL.Query().Get("apikey") != "secret" || request.URL.Path != "/lists/max/weekend/items" {
			t.Fatalf("request URL = %s", request.URL.String())
		}
		if requests == 1 {
			return jsonResponse(http.StatusOK, `{"movies":[{"id":10,"title":"Movie","imdb_id":"tt10","release_year":2025}],"pagination":{"next_cursor":"next"}}`), nil
		}
		if request.URL.Query().Get("cursor") != "next" {
			t.Fatal("missing cursor")
		}
		return jsonResponse(http.StatusOK, `{"shows":[{"id":20,"title":"Show","tvdb_id":30,"release_year":2024}],"pagination":{}}`), nil
	})
	items, err := client.GetListItems(t.Context(), "max", "weekend", false, "", 10)
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 || items.Movies[0].IDs.TMDB != 10 || items.Shows[0].IDs.TMDB != 20 || items.Shows[0].IDs.TVDB != 30 {
		t.Fatalf("requests=%d items=%#v", requests, items)
	}
}

func TestMDBListWatchlistAndErrors(t *testing.T) {
	client := NewMDBList(MDBListConfig{APIKey: "key"})
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/watchlist/items" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		return jsonResponse(http.StatusOK, `{"movies":[],"shows":[]}`), nil
	})
	if _, err := client.GetListItems(t.Context(), "", "", true, "", 10); err != nil {
		t.Fatal(err)
	}

	for name, response := range map[string]*http.Response{
		"private":    jsonResponse(http.StatusForbidden, `{}`),
		"rate limit": jsonResponse(http.StatusTooManyRequests, `{}`),
		"malformed":  jsonResponse(http.StatusOK, `{`),
	} {
		t.Run(name, func(t *testing.T) {
			client := NewMDBList(MDBListConfig{APIKey: "key"})
			client.httpClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) { return response, nil })
			if _, err := client.GetListItems(t.Context(), "", "1", false, "", 10); err == nil {
				t.Fatal("expected error")
			}
		})
	}

	client = NewMDBList(MDBListConfig{APIKey: "key"})
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) { return nil, request.Context().Err() })
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := client.GetListItems(ctx, "", "1", false, "", 10); err == nil || !strings.Contains(err.Error(), "canceled") {
		t.Fatalf("cancellation error = %v", err)
	}
}

func TestMDBListLimitAppliesToRequestedMedia(t *testing.T) {
	client := NewMDBList(MDBListConfig{APIKey: "key"})
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Query().Get("cursor") == "" {
			return jsonResponse(http.StatusOK, `{"shows":[{"id":1}],"next_cursor":"next"}`), nil
		}
		return jsonResponse(http.StatusOK, `{"movies":[{"id":2},{"id":3}]}`), nil
	})
	items, err := client.GetListItems(t.Context(), "", "1", false, "movie", 2)
	if err != nil || len(items.Movies) != 2 {
		t.Fatalf("movies=%d err=%v", len(items.Movies), err)
	}
}
