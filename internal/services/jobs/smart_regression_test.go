package jobs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/internal/filters"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

func TestSmartDeliveryPreservesJobSettings(t *testing.T) {
	for _, media := range []string{"movie", "show"} {
		t.Run(media, func(t *testing.T) {
			db, err := database.New(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = db.Close() })
			if err := db.LogActivity(database.ActivityLog{Timestamp: time.Now().Add(-200 * 24 * time.Hour), MediaType: media, TMDBID: 42, TVDBID: 123, Title: "Removed", Status: string(enums.ActivityStatusAdded)}); err != nil {
				t.Fatal(err)
			}
			posts := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodPost {
					posts++
					if media == "show" {
						var series integrations.SonarrSeries
						if err := json.NewDecoder(r.Body).Decode(&series); err != nil {
							t.Error(err)
						}
						if series.SeriesType != "anime" || series.AddOptions == nil || series.AddOptions.Monitor != "future" {
							t.Errorf("lost show settings: %+v", series)
						}
					}
					_, _ = w.Write([]byte(`{"id":1}`))
					return
				}
				if r.URL.Path == "/api/v3/series/lookup" {
					_, _ = w.Write([]byte(`[{"title":"Removed","tvdbId":123,"tmdbId":42}]`))
					return
				}
				_, _ = w.Write([]byte(`[]`))
			}))
			t.Cleanup(server.Close)
			cfg := &config.Config{}
			cfg.Simkl.ClientID = "test"
			cfg.Radarr.URL = server.URL
			cfg.Sonarr.URL = server.URL
			cfg.Jobs.RepeatPolicy = string(enums.RepeatPolicy90Days)
			job := config.DynamicJob{ID: "smart", Name: "Smart", Type: "smart_popular", Source: "simkl", MediaType: media, Mode: "direct", Limit: 1, Monitor: "future", SeriesType: "anime", RepeatPolicy: string(enums.RepeatPolicyNever)}
			executor := &DynamicJobExecutor{Config: cfg, Database: db, Movies: []integrations.Movie{{Title: "Removed", IDs: integrations.IDs{TMDB: 42}, Rating: 8}}, Shows: []integrations.Show{{Title: "Removed", IDs: integrations.IDs{TVDB: 123, TMDB: 42}, Rating: 8}}}
			if err := executor.Execute(t.Context(), job); err != nil {
				t.Fatal(err)
			}
			if posts != 0 {
				t.Fatal("never override was lost")
			}
			job.RepeatPolicy = string(enums.RepeatPolicyImmediate)
			if err := executor.Execute(t.Context(), job); err != nil {
				t.Fatal(err)
			}
			if posts != 1 {
				t.Fatalf("immediate override: posts=%d", posts)
			}
		})
	}
}

func TestSmartPreviewMatchesAdaptiveRules(t *testing.T) {
	cfg := &config.Config{}
	cfg.Filters.Movies.MinRating = 7
	cfg.Filters.Shows.MinRating = 7
	cfg.Filters.Movies.AllowGenres = []string{"drama"}
	cfg.Filters.Shows.AllowGenres = []string{"drama"}
	cfg.TitleExceptions.AllowedMovieTMDBIDs = []int{3}
	cfg.TitleExceptions.AllowedShowTVDBIDs = []int{13}
	// TVDB/TMDB overlap deliberately: an incorrect TMDB percentile lookup uses another show's rank.
	movies := []integrations.Movie{{Title: "Popular", IDs: integrations.IDs{TMDB: 1}, Rating: 5.8, Votes: 100}, {Title: "Unpopular", IDs: integrations.IDs{TMDB: 2}, Rating: 6.1, Votes: 1}, {Title: "Allowed", IDs: integrations.IDs{TMDB: 3}, Rating: 1, Votes: 10}, {Title: "Override", IDs: integrations.IDs{TMDB: 4}, Rating: 1, Votes: 20, Genres: []string{"drama"}}}
	shows := make([]integrations.Show, len(movies))
	for i, m := range movies {
		shows[i] = integrations.Show{Title: m.Title, IDs: integrations.IDs{TVDB: i + 11, TMDB: 12 - i}, Rating: m.Rating, Votes: m.Votes, Genres: m.Genres}
	}
	job := config.DynamicJob{Type: "smart_popular", BaseMinRating: 6, AdjustmentFactor: 1}
	moviePreview, showPreview := PreviewResponse{}, PreviewResponse{}
	if err := previewMovies(t.Context(), cfg, nil, job, movies, &moviePreview); err != nil {
		t.Fatal(err)
	}
	if err := previewShows(t.Context(), cfg, nil, job, shows, &showPreview); err != nil {
		t.Fatal(err)
	}
	smart := SmartJobConfig{BaseMinRating: 6, AdjustmentFactor: 1}
	_, _, movieDecisions := (&SmartMovieJobExecutor{Config: cfg}).evaluateMoviesWithAdaptiveFilters(t.Context(), movies, filters.CalculateMoviePopularityPercentiles(movies), smart)
	_, _, showDecisions := (&SmartShowJobExecutor{Config: cfg}).evaluateShowsWithAdaptiveFilters(t.Context(), shows, filters.CalculateShowPopularityPercentiles(shows), smart)
	for i, want := range []bool{true, false, true, true} {
		if movieDecisions[i].PassedFilters != want || moviePreview.Items[i].FilteredOut == want {
			t.Errorf("movie %s preview/run disagree with %v", movies[i].Title, want)
		}
		if showDecisions[i].PassedFilters != want || showPreview.Items[i].FilteredOut == want {
			t.Errorf("show %s preview/run disagree with %v", shows[i].Title, want)
		}
	}
}
