package jobs

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

func TestMatchingSonarrSeriesRejectsUnrelatedFirstResult(t *testing.T) {
	show := integrations.Show{Title: "Fallout", Year: 2024, IDs: integrations.IDs{TVDB: 416744, TMDB: 106379}}
	results := []integrations.SonarrSeries{
		{ID: 107, Title: "Trakt: The Series", Year: 2019, TvdbID: 123},
		{Title: "Fallout", Year: 2024, TvdbID: 416744, TmdbID: 106379},
	}

	series, ok := matchingSonarrSeries(show, results)
	if !ok || series.TvdbID != show.IDs.TVDB || series.ID != 0 {
		t.Fatalf("matched %+v, ok=%t", series, ok)
	}
}

func TestMatchingSonarrSeriesRequiresIdentityOrTitleAndYear(t *testing.T) {
	show := integrations.Show{Title: "The Pitt", Year: 2025, IDs: integrations.IDs{TVDB: 449139}}
	results := []integrations.SonarrSeries{{ID: 107, Title: "Trakt: The Series", Year: 2019, TvdbID: 123}}

	if series, ok := matchingSonarrSeries(show, results); ok {
		t.Fatalf("unexpected match: %+v", series)
	}
}

func TestShowDeliveryPreservesRepeatIdentity(t *testing.T) {
	for _, mode := range []string{"direct", "jellyseerr"} {
		for _, tvdbID := range []int{0, 123} {
			t.Run(mode+fmt.Sprint(tvdbID), func(t *testing.T) {
				db, err := database.New(t.TempDir())
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = db.Close() })
				var posts atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					switch {
					case r.Method == http.MethodPost:
						posts.Add(1)
						w.WriteHeader(http.StatusCreated)
						_, _ = w.Write([]byte(`{"id":1,"title":"Delivered","tvdbId":123}`))
					case r.URL.Path == "/api/v3/series/lookup":
						_, _ = w.Write([]byte(`[{"title":"Delivered","tvdbId":123}]`))
					case r.URL.Path == "/api/v3/series":
						_, _ = w.Write([]byte(`[]`))
					default:
						_, _ = w.Write([]byte(`{}`))
					}
				}))
				t.Cleanup(server.Close)
				cfg := &config.Config{}
				cfg.Sonarr.URL = server.URL
				cfg.Jellyseerr.URL = server.URL
				cfg.Jellyseerr.APIKey = "test"
				job := JobConfig{JobID: "repeat", JobName: "Repeat", MediaType: "show", Mode: mode, RepeatPolicy: string(enums.RepeatPolicyNever)}
				show := integrations.Show{Title: "Delivered", IDs: integrations.IDs{TMDB: 42, TVDB: tvdbID}}
				executor := &ShowJobExecutor{Config: cfg, Database: db}
				for range 2 {
					if mode == "direct" {
						if err := executor.executeShowsDirect(t.Context(), job, []integrations.Show{show}, nil); err != nil {
							t.Fatal(err)
						}
					} else {
						executor.executeShowsJellyseerr(t.Context(), job, []integrations.Show{show}, nil)
					}
				}
				if posts.Load() != 1 {
					t.Fatalf("delivered %d times despite never repeat policy", posts.Load())
				}
				preview := PreviewResponse{}
				if err := previewShows(t.Context(), cfg, db, config.DynamicJob{Mode: mode, RepeatPolicy: job.RepeatPolicy}, []integrations.Show{show}, &preview); err != nil {
					t.Fatal(err)
				}
				if len(preview.Items) != 1 || !preview.Items[0].RepeatBlocked {
					t.Fatalf("preview lost delivery identity: %+v", preview)
				}
			})
		}
	}
}

func TestMatchingSonarrSeriesRejectsConflictingIdentity(t *testing.T) {
	for _, ids := range []integrations.IDs{{TVDB: 123}, {TMDB: 42}, {TVDB: 123, TMDB: 42}} {
		show := integrations.Show{Title: "Same title", Year: 2024, IDs: ids}
		if series, ok := matchingSonarrSeries(show, []integrations.SonarrSeries{{Title: show.Title, Year: show.Year, TvdbID: 456, TmdbID: 99}}); ok {
			t.Fatalf("matched conflicting IDs: %+v", series)
		}
	}
}
