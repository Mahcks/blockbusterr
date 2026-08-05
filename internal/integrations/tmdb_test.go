package integrations

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestTMDBCertificationEnrichment(t *testing.T) {
	client := NewTMDB(TMDBConfig{APIKey: "key"})
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.Path {
		case "/3/movie/10":
			return jsonResponse(http.StatusOK, `{"release_dates":{"results":[{"iso_3166_1":"US","release_dates":[{"certification":"PG-13"},{"certification":"PG-13"},{"certification":"R"}]},{"iso_3166_1":"GB","release_dates":[{"certification":"12"}]}]}}`), nil
		case "/3/tv/20":
			return jsonResponse(http.StatusOK, `{"content_ratings":{"results":[{"iso_3166_1":"US","rating":"TV-14"}]}}`), nil
		default:
			t.Fatalf("path = %s", request.URL.Path)
			return nil, nil
		}
	})
	movies := []Movie{{IDs: IDs{TMDB: 10}}}
	shows := []Show{{IDs: IDs{TMDB: 20}}}
	if err := client.EnrichMovieCertifications(t.Context(), movies); err != nil {
		t.Fatal(err)
	}
	if err := client.EnrichShowCertifications(t.Context(), shows); err != nil {
		t.Fatal(err)
	}
	if len(movies[0].Certifications) != 3 || movies[0].Certifications[0] != (Certification{Value: "PG-13", Country: "US", Source: "tmdb"}) {
		t.Fatalf("movie certifications = %#v", movies[0].Certifications)
	}
	if len(shows[0].Certifications) != 1 || shows[0].Certifications[0].Value != "TV-14" {
		t.Fatalf("show certifications = %#v", shows[0].Certifications)
	}
}

func TestTMDBShowEnrichmentIsBoundedAndConcurrent(t *testing.T) {
	var active, maximum int32
	client := NewTMDB(TMDBConfig{APIKey: "configured"})
	client.httpClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		current := atomic.AddInt32(&active, 1)
		for {
			observed := atomic.LoadInt32(&maximum)
			if current <= observed || atomic.CompareAndSwapInt32(&maximum, observed, current) {
				break
			}
		}
		time.Sleep(5 * time.Millisecond)
		atomic.AddInt32(&active, -1)
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"episode_run_time":[45],"external_ids":{"tvdb_id":123}}`))}, nil
	})

	shows := make([]Show, 20)
	for index := range shows {
		shows[index].IDs.TMDB = index + 1
	}
	if err := client.enrichShows(context.Background(), shows); err != nil {
		t.Fatal(err)
	}

	if maximum <= 1 || maximum > 8 {
		t.Fatalf("maximum concurrent requests = %d, want 2..8", maximum)
	}
	if shows[0].IDs.TVDB != 123 || shows[0].Runtime != 45 {
		t.Fatalf("show was not enriched: %#v", shows[0])
	}
}

func TestTMDBRecommendationsAreOneHopAndDeduplicated(t *testing.T) {
	client := NewTMDB(TMDBConfig{APIKey: "key"})
	requests := 0
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		switch request.URL.Path {
		case "/3/movie/10/recommendations":
			return jsonResponse(http.StatusOK, `{"page":1,"total_pages":1,"results":[{"id":10,"title":"Seed"},{"id":30,"title":"Shared"}]}`), nil
		case "/3/movie/20/recommendations":
			return jsonResponse(http.StatusOK, `{"page":1,"total_pages":1,"results":[{"id":30,"title":"Shared"},{"id":40,"title":"Unique"}]}`), nil
		default:
			t.Fatalf("unexpected recursive request: %s", request.URL.Path)
			return nil, nil
		}
	})

	movies, err := client.GetMovieRecommendations(t.Context(), []int{10, 20, 10}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 || len(movies) != 2 || movies[0].IDs.TMDB != 30 || movies[1].IDs.TMDB != 40 {
		t.Fatalf("requests=%d movies=%#v", requests, movies)
	}
}

func TestTMDBRecommendationsPaginateUntilLimit(t *testing.T) {
	client := NewTMDB(TMDBConfig{APIKey: "key"})
	client.httpClient.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		page := request.URL.Query().Get("page")
		if page == "1" {
			return jsonResponse(http.StatusOK, `{"page":1,"total_pages":2,"results":[{"id":10},{"id":20,"title":"First"}]}`), nil
		}
		return jsonResponse(http.StatusOK, `{"page":2,"total_pages":2,"results":[{"id":20},{"id":30,"title":"Second"}]}`), nil
	})
	movies, err := client.GetMovieRecommendations(t.Context(), []int{10}, 2)
	if err != nil || len(movies) != 2 || movies[0].IDs.TMDB != 20 || movies[1].IDs.TMDB != 30 {
		t.Fatalf("movies=%#v err=%v", movies, err)
	}
}

func TestTMDBEnrichmentFailureIsExplicit(t *testing.T) {
	client := NewTMDB(TMDBConfig{APIKey: "key"})
	client.httpClient.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusTooManyRequests, `{}`), nil
	})
	if err := client.EnrichShowCertifications(t.Context(), []Show{{IDs: IDs{TMDB: 20}}}); err == nil {
		t.Fatal("expected enrichment error")
	}
}
