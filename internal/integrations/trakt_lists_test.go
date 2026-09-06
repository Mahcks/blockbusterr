package integrations

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestTraktListItemsPaginateAndPreserveMixedOrder(t *testing.T) {
	client := NewTrakt(TraktConfig{ClientID: "client", AccessToken: "secret"})
	requests := 0
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Fatal("missing OAuth header")
		}
		if !strings.Contains(request.URL.Path, "/users/max/lists/favorites/items/movie,show") {
			t.Fatalf("path = %s", request.URL.Path)
		}
		count := 100
		if request.URL.Query().Get("page") == "2" {
			count = 1
		}
		items := make([]string, count)
		for index := range items {
			id := index + 1
			if requests == 2 {
				id = 101
			}
			if id%2 == 0 {
				items[index] = fmt.Sprintf(`{"type":"show","show":{"title":"Show %d","ids":{"tmdb":%d,"tvdb":%d}}}`, id, id, id)
			} else {
				items[index] = fmt.Sprintf(`{"type":"movie","movie":{"title":"Movie %d","ids":{"tmdb":%d}}}`, id, id)
			}
		}
		return jsonResponse(http.StatusOK, "["+strings.Join(items, ",")+"]"), nil
	})
	items, err := client.GetListItems(t.Context(), "max", "favorites", false, "", 101)
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 || len(items.Movies) != 51 || len(items.Shows) != 50 || items.Movies[50].IDs.TMDB != 101 {
		t.Fatalf("requests=%d movies=%d shows=%d", requests, len(items.Movies), len(items.Shows))
	}
}

func TestTraktListLimitAppliesToRequestedMedia(t *testing.T) {
	client := NewTrakt(TraktConfig{ClientID: "client"})
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Query().Get("page") == "1" {
			items := make([]string, 100)
			for i := range items {
				items[i] = fmt.Sprintf(`{"type":"show","show":{"ids":{"tmdb":%d}}}`, i+1)
			}
			return jsonResponse(http.StatusOK, "["+strings.Join(items, ",")+"]"), nil
		}
		return jsonResponse(http.StatusOK, `[{"type":"movie","movie":{"ids":{"tmdb":201}}},{"type":"movie","movie":{"ids":{"tmdb":202}}}]`), nil
	})
	items, err := client.GetListItems(t.Context(), "", "1", false, "movie", 2)
	if err != nil || len(items.Movies) != 2 {
		t.Fatalf("movies=%d err=%v", len(items.Movies), err)
	}
}

func TestTraktRefreshIsSharedAcrossClients(t *testing.T) {
	var refreshes atomic.Int32
	current := TraktToken{AccessToken: "expired", RefreshToken: "refresh", CreatedAt: 1, ExpiresIn: 1}
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path == "/oauth/token" {
			refreshes.Add(1)
			return jsonResponse(http.StatusOK, fmt.Sprintf(`{"access_token":"fresh","refresh_token":"next","created_at":%d,"expires_in":3600}`, time.Now().Unix())), nil
		}
		return jsonResponse(http.StatusOK, `[]`), nil
	})
	newClient := func() *Trakt {
		client := NewTrakt(TraktConfig{ClientID: "client", ClientSecret: "secret", AccessToken: current.AccessToken, RefreshToken: current.RefreshToken, TokenExpires: current.ExpiresAt(), LoadToken: func() TraktToken { return current }, OnToken: func(token TraktToken) error { current = token; return nil }})
		client.httpClient.Transport = transport
		return client
	}
	clients := []*Trakt{newClient(), newClient()}
	done := make(chan error, len(clients))
	for _, client := range clients {
		go func() { _, err := client.GetListItems(t.Context(), "", "1", false, "movie", 1); done <- err }()
	}
	for range clients {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	if refreshes.Load() != 1 {
		t.Fatalf("refreshes=%d", refreshes.Load())
	}
}

func TestTraktListErrors(t *testing.T) {
	for name, response := range map[string]*http.Response{
		"rate limit": jsonResponse(http.StatusTooManyRequests, `{}`),
		"malformed":  jsonResponse(http.StatusOK, `{`),
	} {
		t.Run(name, func(t *testing.T) {
			client := NewTrakt(TraktConfig{ClientID: "client"})
			client.httpClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) { return response, nil })
			if _, err := client.GetListItems(t.Context(), "", "list", false, "", 10); err == nil {
				t.Fatal("expected error")
			}
		})
	}
	client := NewTrakt(TraktConfig{ClientID: "client"})
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) { return nil, request.Context().Err() })
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := client.GetListItems(ctx, "", "list", false, "", 10); err == nil {
		t.Fatal("expected cancellation error")
	}
}

func TestTraktConnectedWatchlist(t *testing.T) {
	client := NewTrakt(TraktConfig{ClientID: "client", AccessToken: "token"})
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/sync/watchlist/movie,show/rank/asc" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		return jsonResponse(http.StatusOK, `[{"type":"movie","movie":{"title":"Saved","ids":{"tmdb":7}}}]`), nil
	})
	items, err := client.GetListItems(t.Context(), "", "", true, "", 10)
	if err != nil || len(items.Movies) != 1 || items.Movies[0].IDs.TMDB != 7 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}
