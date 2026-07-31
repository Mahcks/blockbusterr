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
	client.enrichShows(context.Background(), shows)

	if maximum <= 1 || maximum > 8 {
		t.Fatalf("maximum concurrent requests = %d, want 2..8", maximum)
	}
	if shows[0].IDs.TVDB != 123 || shows[0].Runtime != 45 {
		t.Fatalf("show was not enriched: %#v", shows[0])
	}
}
