package jobs

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
)

type fakeListSource struct {
	result ListResult
	err    error
	calls  int
	limit  int
}

func (source *fakeListSource) FetchList(ctx context.Context, _ config.ListLocator, limit int) (ListResult, error) {
	source.calls++
	source.limit = limit
	if err := ctx.Err(); err != nil {
		return ListResult{}, err
	}
	return source.result, source.err
}

func TestListSourceUsesSharedPreviewAndExecutionPipeline(t *testing.T) {
	radarr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	t.Cleanup(radarr.Close)

	cfg := &config.Config{}
	cfg.Radarr.URL, cfg.Radarr.APIKey = radarr.URL, "test"
	cfg.Jobs.Mode = "direct"
	cfg.Filters.Movies.BlacklistedMinYear = 3000
	job := config.DynamicJob{ID: "list-job", Name: "My List", Type: "list", Source: "fake", MediaType: "movie", Limit: 50, RuleSetID: config.DefaultMoviesRuleSetID, List: &config.ListLocator{Kind: "public_list", ListID: "favorites", Ordering: "source"}}
	source := &fakeListSource{result: ListResult{Movies: []integrations.Movie{
		{Title: "Keep", Year: 2025, IDs: integrations.IDs{TMDB: 10}},
		{Title: "Duplicate", IDs: integrations.IDs{TMDB: 10}},
		{Title: "Invalid"},
	}, Shows: []integrations.Show{{Title: "Mixed", IDs: integrations.IDs{TVDB: 20}}}}}
	registry := ListSourceRegistry{"fake": source}

	preview, err := previewDynamicJob(cfg, nil, job, registry)
	if err != nil {
		t.Fatal(err)
	}
	if preview.TotalFound != 1 || preview.FilteredOut != 1 {
		t.Fatalf("preview = found %d, filtered %d; want 1, 1", preview.TotalFound, preview.FilteredOut)
	}
	executor := DynamicJobExecutor{Config: cfg, DryRun: true, ListSources: registry}
	if err := executor.Execute(t.Context(), job); err != nil {
		t.Fatal(err)
	}
	if source.calls != 2 {
		t.Fatalf("source calls = %d; want one preview and one execution call", source.calls)
	}
	if source.limit != 50 {
		t.Fatalf("source limit = %d; want 50", source.limit)
	}
}

func TestListSourceRejectsInvalidLocatorBeforeFetch(t *testing.T) {
	source := &fakeListSource{}
	executor := DynamicJobExecutor{Config: &config.Config{}, DryRun: true, ListSources: ListSourceRegistry{"fake": source}}
	job := config.DynamicJob{Name: "Unsafe", Type: "list", Source: "fake", MediaType: "movie", List: &config.ListLocator{Kind: "public_list", ListID: "https://example.com/list"}}
	if err := executor.Execute(t.Context(), job); err == nil {
		t.Fatal("expected invalid locator error")
	}
	if source.calls != 0 {
		t.Fatalf("source called %d times", source.calls)
	}
}

func TestListSourceRequiresLetterboxdOwner(t *testing.T) {
	locator := config.ListLocator{Kind: "public_list", ListID: "favorites"}
	if err := ValidateListSourceLocator("letterboxd", locator); err == nil {
		t.Fatal("expected Letterboxd owner error")
	}
	if err := ValidateListSourceLocator("trakt", locator); err != nil {
		t.Fatalf("Trakt locator rejected: %v", err)
	}
}

func TestListSourcePropagatesProviderAndCancellationErrors(t *testing.T) {
	providerErr := errors.New("provider unavailable")
	source := &fakeListSource{err: providerErr}
	client := newListDiscoveryClient("fake", source)
	locator := config.ListLocator{Kind: "public_list", Slug: "favorites"}
	if _, err := client.GetListMovies(t.Context(), locator, 10); !errors.Is(err, providerErr) {
		t.Fatalf("provider error = %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := client.GetListMovies(ctx, locator, 10); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}
}

func TestListSourceAcceptsEmptyList(t *testing.T) {
	client := newListDiscoveryClient("fake", &fakeListSource{})
	movies, err := client.GetListMovies(t.Context(), config.ListLocator{Kind: "public_list", ListID: "empty"}, 10)
	if err != nil || len(movies) != 0 {
		t.Fatalf("empty list = %v, %v", movies, err)
	}
}

func TestListSourceRegistryOnlyReportsConfiguredAdapters(t *testing.T) {
	RegisterListSource("fake", func(cfg *config.Config) (ListSource, error) {
		if cfg.Trakt.ClientID == "" {
			return nil, errors.New("not configured")
		}
		return &fakeListSource{}, nil
	})
	t.Cleanup(func() { delete(listSourceFactories, "fake") })
	if got := AvailableListSources(&config.Config{}); len(got) != 0 {
		t.Fatalf("unconfigured sources = %v", got)
	}
	cfg := &config.Config{}
	cfg.Trakt.ClientID = "configured"
	cfg.MDBList.APIKey = "configured"
	cfg.Letterboxd.ExperimentalScraping = true
	if got := AvailableListSources(cfg); len(got) != 4 || got[0] != "fake" || got[1] != "letterboxd" || got[2] != "mdblist" || got[3] != "trakt" {
		t.Fatalf("configured sources = %v", got)
	}
}
