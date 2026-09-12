package integrations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestTraktPaginationKeepsOffsets(t *testing.T) {
	for _, limit := range []int{1, 100, 101, 150, 250} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			client := NewTrakt(TraktConfig{ClientID: "client"})
			client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
				page, _ := strconv.Atoi(request.URL.Query().Get("page"))
				size, _ := strconv.Atoi(request.URL.Query().Get("limit"))
				items := make([]string, size)
				for i := range items {
					items[i] = fmt.Sprintf(`{"ids":{"tmdb":%d}}`, (page-1)*size+i+1)
				}
				return jsonResponse(http.StatusOK, "["+strings.Join(items, ",")+"]"), nil
			})
			movies, err := client.GetPopularMovies(t.Context(), limit)
			if err != nil || len(movies) != limit {
				t.Fatalf("movies=%d err=%v", len(movies), err)
			}
			for i, movie := range movies {
				if movie.IDs.TMDB != i+1 {
					t.Fatalf("item %d has ID %d", i, movie.IDs.TMDB)
				}
			}
		})
	}
}

func TestTraktRefreshRejectsIncompleteToken(t *testing.T) {
	for _, body := range []string{`null`, `{}`, `{"access_token":"fresh"}`, `{"access_token":"fresh","refresh_token":"next","expires_in":0}`} {
		for _, device := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/device=%t", body, device), func(t *testing.T) {
				saved := false
				client := NewTrakt(TraktConfig{ClientID: "client", AccessToken: "expired", RefreshToken: "refresh", TokenExpires: 1, OnToken: func(TraktToken) error { saved = true; return nil }})
				client.httpClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) { return jsonResponse(http.StatusOK, body), nil })
				var err error
				if device {
					_, _, err = client.PollDeviceAuth(t.Context(), "device")
				} else {
					_, err = client.accessTokenForRequest(t.Context())
				}
				if err == nil || saved || client.accessToken != "expired" || client.refreshToken != "refresh" {
					t.Fatal("incomplete token was accepted or replaced existing credentials")
				}
			})
		}
	}
}

func TestProviderPaginationFailureDiscardsEarlierCandidates(t *testing.T) {
	for _, provider := range []string{"trakt", "trakt-list", "tmdb", "mdblist"} {
		for _, failure := range []string{"unauthorized", "rate-limit", "malformed", "timeout"} {
			t.Run(provider+"/"+failure, func(t *testing.T) {
				calls := 0
				transport := roundTripFunc(func(*http.Request) (*http.Response, error) {
					calls++
					if calls == 1 {
						switch provider {
						case "trakt":
							return jsonResponse(http.StatusOK, "["+strings.TrimSuffix(strings.Repeat(`{"ids":{"tmdb":1}},`, 100), ",")+"]"), nil
						case "trakt-list":
							return jsonResponse(http.StatusOK, "["+strings.TrimSuffix(strings.Repeat(`{"type":"movie","movie":{"ids":{"tmdb":1}}},`, 100), ",")+"]"), nil
						case "tmdb":
							return jsonResponse(http.StatusOK, `{"page":1,"total_pages":2,"results":[{"id":1}]}`), nil
						case "mdblist":
							return jsonResponse(http.StatusOK, `{"movies":[{"id":1}],"next_cursor":"next"}`), nil
						}
					}
					switch failure {
					case "unauthorized":
						return jsonResponse(http.StatusUnauthorized, `{}`), nil
					case "rate-limit":
						return jsonResponse(http.StatusTooManyRequests, `{}`), nil
					case "malformed":
						return jsonResponse(http.StatusOK, `{`), nil
					default:
						return nil, context.DeadlineExceeded
					}
				})
				var movies []Movie
				var err error
				switch provider {
				case "trakt", "trakt-list":
					client := NewTrakt(TraktConfig{ClientID: "client"})
					client.httpClient.Transport = transport
					if provider == "trakt" {
						movies, err = client.GetPopularMovies(t.Context(), 150)
					} else {
						var result TraktListItems
						result, err = client.GetListItems(t.Context(), "", "list", false, "movie", 150)
						movies = result.Movies
					}
				case "tmdb":
					client := NewTMDB(TMDBConfig{APIKey: "key"})
					client.httpClient.Transport = transport
					movies, err = client.GetPopularMovies(t.Context(), 150)
				case "mdblist":
					client := NewMDBList(MDBListConfig{APIKey: "key"})
					client.httpClient.Transport = transport
					var result MDBListItems
					result, err = client.GetListItems(t.Context(), "", "1", false, "movie", 150)
					movies = result.Movies
				}
				if calls != 2 || err == nil || len(movies) != 0 {
					t.Fatalf("calls=%d movies=%d err=%v", calls, len(movies), err)
				}
				if failure == "timeout" && !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("lost timeout cause: %v", err)
				}
			})
		}
	}
}

func TestTraktRefreshFailurePreservesCredentialsAndSkipsDiscovery(t *testing.T) {
	for _, failure := range []string{"unauthorized", "rate-limit", "malformed", "timeout"} {
		t.Run(failure, func(t *testing.T) {
			saved := false
			client := NewTrakt(TraktConfig{ClientID: "client", AccessToken: "expired", RefreshToken: "refresh", TokenExpires: 1, OnToken: func(TraktToken) error { saved = true; return nil }})
			calls := 0
			client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
				calls++
				if request.URL.Path != "/oauth/token" {
					t.Fatal("discovery ran after failed refresh")
				}
				switch failure {
				case "unauthorized":
					return jsonResponse(http.StatusUnauthorized, `{}`), nil
				case "rate-limit":
					return jsonResponse(http.StatusTooManyRequests, `{}`), nil
				case "malformed":
					return jsonResponse(http.StatusOK, `{`), nil
				default:
					return nil, context.DeadlineExceeded
				}
			})
			items, err := client.GetListItems(t.Context(), "", "list", false, "movie", 1)
			if calls != 1 || err == nil || len(items.Movies) != 0 || saved || client.accessToken != "expired" || client.refreshToken != "refresh" {
				t.Fatal("failed refresh changed credentials or continued discovery")
			}
			if failure == "timeout" && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("lost timeout cause: %v", err)
			}
		})
	}
}
