package integrations

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJellyseerrRequestHonorsCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewJellyseerr(JellyseerrConfig{URL: server.URL, APIKey: "test"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetMovieInfoContext(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("GetMovieInfoContext() error = %v, want context canceled", err)
	}
}
