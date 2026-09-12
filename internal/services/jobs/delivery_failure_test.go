package jobs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

func TestDeliveryFailureBudget(t *testing.T) {
	for _, mode := range []string{"direct", "jellyseerr"} {
		for _, media := range []string{"movie", "show"} {
			for _, outcome := range []string{"disconnect", "malformed", "gateway", "rejected"} {
				t.Run(mode+"/"+media+"/"+outcome, func(t *testing.T) {
					dir := t.TempDir()
					db, err := database.New(dir)
					if err != nil {
						t.Fatal(err)
					}
					defer func() { _ = db.Close() }()
					var posts atomic.Int32
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Type", "application/json")
						if r.Method == http.MethodPost {
							posts.Add(1)
							switch outcome {
							case "disconnect":
								conn, _, err := w.(http.Hijacker).Hijack()
								if err != nil {
									t.Error(err)
									return
								}
								_ = conn.Close()
							case "malformed":
								w.WriteHeader(http.StatusCreated)
								_, _ = w.Write([]byte(`{"id":`))
							case "gateway":
								w.WriteHeader(http.StatusBadGateway)
							case "rejected":
								w.WriteHeader(http.StatusForbidden)
							}
							return
						}
						switch r.URL.Path {
						case "/api/v3/series/lookup":
							id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Query().Get("term"), "tvdb:"))
							_, _ = fmt.Fprintf(w, `[{"title":"Test","tvdbId":%d}]`, id)
						case "/api/v3/series", "/api/v3/movie":
							_, _ = w.Write([]byte(`[]`))
						default:
							_, _ = w.Write([]byte(`{}`))
						}
					}))
					defer server.Close()
					cfg := &config.Config{}
					cfg.Radarr.URL, cfg.Sonarr.URL, cfg.Jellyseerr.URL = server.URL, server.URL, server.URL
					cfg.Jellyseerr.APIKey = "test"
					cfg.Jobs.GlobalLimitMovies, cfg.Jobs.GlobalLimitShows = 1, 1
					job := JobConfig{JobID: "failure", JobName: "Failure", Mode: mode, DeliveryLimit: 1}
					if media == "movie" {
						executor := &MovieJobExecutor{Config: cfg, Database: db}
						movies := []integrations.Movie{{Title: "Test", IDs: integrations.IDs{TMDB: 1}}, {Title: "Other", IDs: integrations.IDs{TMDB: 2}}}
						if mode == "direct" {
							if err := executor.executeMoviesDirect(t.Context(), job, movies, nil); err != nil {
								t.Fatal(err)
							}
						} else {
							executor.executeMoviesJellyseerr(t.Context(), job, movies, nil)
						}
					} else {
						executor := &ShowJobExecutor{Config: cfg, Database: db}
						shows := []integrations.Show{{Title: "Test", IDs: integrations.IDs{TMDB: 1, TVDB: 1}}, {Title: "Test", IDs: integrations.IDs{TMDB: 2, TVDB: 2}}}
						if mode == "direct" {
							if err := executor.executeShowsDirect(t.Context(), job, shows, nil); err != nil {
								t.Fatal(err)
							}
						} else {
							executor.executeShowsJellyseerr(t.Context(), job, shows, nil)
						}
					}
					wantPosts, wantUsage := int32(1), 1
					if outcome == "rejected" {
						wantPosts, wantUsage = 2, 0
					}
					if posts.Load() != wantPosts {
						t.Fatalf("POST count = %d, want %d", posts.Load(), wantPosts)
					}
					if err := db.Close(); err != nil {
						t.Fatal(err)
					}
					db, err = database.New(dir)
					if err != nil {
						t.Fatal(err)
					}
					if count, err := db.CountDeliveriesSince(media, time.Now().Add(-time.Hour)); err != nil || count != wantUsage {
						t.Fatalf("usage after restart = %d, %v; want %d", count, err, wantUsage)
					}
				})
			}
		}
	}
}

func TestCanceledDeliveryPreservesCompletedCounts(t *testing.T) {
	for _, mode := range []string{"direct", "jellyseerr"} {
		for _, media := range []string{"movie", "show", "smart_movie", "smart_show"} {
			t.Run(mode+"/"+media, func(t *testing.T) {
				db, err := database.New(t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = db.Close() }()
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				var posts atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					if r.Method == http.MethodPost {
						if posts.Add(1) == 2 {
							cancel()
							return
						}
						w.WriteHeader(http.StatusCreated)
						_, _ = w.Write([]byte(`{"id":1}`))
					} else if r.URL.Path == "/api/v3/series/lookup" {
						id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Query().Get("term"), "tvdb:"))
						_, _ = fmt.Fprintf(w, `[{"title":"Test","tvdbId":%d}]`, id)
					} else if r.URL.Path == "/api/v3/series" || r.URL.Path == "/api/v3/movie" {
						_, _ = w.Write([]byte(`[]`))
					} else {
						_, _ = w.Write([]byte(`{}`))
					}
				}))
				defer server.Close()
				cfg := &config.Config{}
				cfg.Simkl.ClientID = "test"
				cfg.Radarr.URL, cfg.Sonarr.URL, cfg.Jellyseerr.URL = server.URL, server.URL, server.URL
				cfg.Jellyseerr.APIKey = "test"
				job := JobConfig{JobID: "cancel", JobName: "Cancel", Source: "simkl", Mode: mode, MediaType: strings.TrimPrefix(media, "smart_")}
				smartJob := SmartJobConfig{JobID: job.JobID, JobName: job.JobName, Source: job.Source, Mode: job.Mode, MediaType: job.MediaType}
				if job.MediaType == "movie" {
					fetcher := func(context.Context, *DiscoveryClient, int, string) ([]integrations.Movie, error) {
						return []integrations.Movie{{Title: "Test", IDs: integrations.IDs{TMDB: 1}}, {Title: "Other", IDs: integrations.IDs{TMDB: 2}}}, nil
					}
					if media == "smart_movie" {
						err = (&SmartMovieJobExecutor{Config: cfg, Database: db}).Execute(ctx, smartJob, fetcher)
					} else {
						err = (&MovieJobExecutor{Config: cfg, Database: db}).Execute(ctx, job, fetcher)
					}
				} else {
					fetcher := func(context.Context, *DiscoveryClient, int, string) ([]integrations.Show, error) {
						return []integrations.Show{{Title: "Test", IDs: integrations.IDs{TMDB: 1, TVDB: 1}}, {Title: "Test", IDs: integrations.IDs{TMDB: 2, TVDB: 2}}}, nil
					}
					if media == "smart_show" {
						err = (&SmartShowJobExecutor{Config: cfg, Database: db}).Execute(ctx, smartJob, fetcher)
					} else {
						err = (&ShowJobExecutor{Config: cfg, Database: db}).Execute(ctx, job, fetcher)
					}
				}
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("execution error = %v", err)
				}
				runs, err := db.GetRecentJobRuns(10, "")
				if err != nil {
					t.Fatal(err)
				}
				if len(runs) != 1 || runs[0].Status != string(enums.JobRunStatusFailed) || runs[0].Added+runs[0].Requested != 1 {
					t.Fatalf("partial delivery counts lost: %+v", runs)
				}
				if posts.Load() != 2 {
					t.Fatalf("posts = %d", posts.Load())
				}
			})
		}
	}
}
