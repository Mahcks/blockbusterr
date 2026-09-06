package global

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/services/jobs"
)

func TestConfigUpdatesAreCopyOnWrite(t *testing.T) {
	cfg := &config.Config{ConfigFilePath: filepath.Join(t.TempDir(), "config.yaml")}
	cfg.Jobs.GlobalLimitMovies = 1
	gctx := New(context.Background(), cfg, nil, "test", "test", nil)

	snapshot := gctx.Config()
	snapshot.Jobs.GlobalLimitMovies = 99
	if got := gctx.Config().Jobs.GlobalLimitMovies; got != 1 {
		t.Fatalf("snapshot mutation changed live config: %d", got)
	}

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			if err := UpdateConfig(gctx, func(candidate *config.Config) error {
				candidate.Jobs.GlobalLimitMovies++
				return nil
			}); err != nil {
				t.Errorf("UpdateConfig() error = %v", err)
			}
		})
	}
	wg.Wait()
	if got := gctx.Config().Jobs.GlobalLimitMovies; got != 11 {
		t.Fatalf("serialized updates = %d, want 11", got)
	}
}

func TestConfigUpdateFailureDoesNotPublish(t *testing.T) {
	cfg := &config.Config{ConfigFilePath: filepath.Join(t.TempDir(), "config.yaml")}
	cfg.Jobs.GlobalLimitMovies = 1
	gctx := New(context.Background(), cfg, nil, "test", "test", nil)
	err := UpdateConfig(gctx, func(candidate *config.Config) error {
		candidate.Jobs.GlobalLimitMovies = 99
		candidate.ConfigFilePath = filepath.Join(t.TempDir(), "missing", "config.yaml")
		return nil
	})
	if err == nil {
		t.Fatal("expected save failure")
	}
	if got := gctx.Config().Jobs.GlobalLimitMovies; got != 1 {
		t.Fatalf("failed update published value %d", got)
	}
}

type tokenTestTransport func(*http.Request) (*http.Response, error)

func (fn tokenTestTransport) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func TestTraktAdaptersRefreshFromLiveConfiguration(t *testing.T) {
	cfg := &config.Config{ConfigFilePath: filepath.Join(t.TempDir(), "config.yaml")}
	cfg.Trakt.ClientID, cfg.Trakt.ClientSecret = "client", "secret"
	cfg.Trakt.AccessToken, cfg.Trakt.RefreshToken, cfg.Trakt.TokenExpires = "expired", "original-refresh", 1
	gctx := New(context.Background(), cfg, nil, "test", "test", nil)
	snapshots := []*config.Config{gctx.Config(), gctx.Config()}
	var refreshes, pages atomic.Int32
	oldTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = oldTransport })
	http.DefaultTransport = tokenTestTransport(func(req *http.Request) (*http.Response, error) {
		body := `[{"type":"movie","movie":{"title":"Movie","ids":{"tmdb":42}}}]`
		if req.URL.Path == "/oauth/token" {
			refreshes.Add(1)
			body = fmt.Sprintf(`{"access_token":"fresh","refresh_token":"rotated","created_at":%d,"expires_in":3600}`, time.Now().Unix())
		} else {
			pages.Add(1)
			if req.URL.Query().Get("page") == "1" {
				body = "[" + strings.TrimSuffix(strings.Repeat("{},", 100), ",") + "]"
			}
			if req.Header.Get("Authorization") != "Bearer fresh" {
				t.Error("stale access token used")
			}
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	var wg sync.WaitGroup
	for _, snapshot := range snapshots {
		wg.Go(func() {
			if _, err := jobs.InspectListSource(t.Context(), snapshot, "trakt", config.ListLocator{Kind: "public_list", ListID: "1", Ordering: "source"}); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if refreshes.Load() != 1 || pages.Load() != 4 {
		t.Fatalf("refreshes=%d pages=%d", refreshes.Load(), pages.Load())
	}
	if got := gctx.Config().Trakt.RefreshToken; got != "rotated" {
		t.Fatal("refresh not persisted")
	}
}
