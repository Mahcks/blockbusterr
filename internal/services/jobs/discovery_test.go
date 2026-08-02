package jobs

import (
	"testing"

	"github.com/mahcks/blockbusterr/config"
)

func TestNewDiscoveryClient(t *testing.T) {
	cfg := &config.Config{}
	cfg.Trakt.ClientID = "configured"
	cfg.TMDB.APIKey = "configured"
	cfg.Simkl.ClientID = "configured"
	for _, source := range []string{"trakt", "tmdb", "simkl"} {
		client, err := NewDiscoveryClient(cfg, source)
		if err != nil {
			t.Fatalf("source %s: %v", source, err)
		}
		if client.Source() != source {
			t.Fatalf("source = %s, want %s", client.Source(), source)
		}
	}
	if _, err := NewDiscoveryClient(cfg, "other"); err == nil {
		t.Fatal("expected invalid source error")
	}
}

func TestNewDiscoveryClientRejectsUnconfiguredProvider(t *testing.T) {
	if _, err := NewDiscoveryClient(&config.Config{}, "simkl"); err == nil {
		t.Fatal("expected missing Simkl credentials error")
	}
}

func TestDynamicExecutorReturnsProviderConfigurationError(t *testing.T) {
	executor := &DynamicJobExecutor{Config: &config.Config{}}
	err := executor.Execute(t.Context(), config.DynamicJob{Type: "trending", Source: "simkl", MediaType: "movie", Limit: 1})
	if err == nil {
		t.Fatal("expected provider configuration error")
	}
}

func TestConfigForJobUsesAssignedRuleSet(t *testing.T) {
	cfg := &config.Config{}
	cfg.Filters.Movies.MinRating = 7
	cfg.Filters.Movies.BlacklistedTMDBIds = []int{1}
	cfg.MigrateRuleSets()
	job := config.DynamicJob{MediaType: "movie"}
	effective, err := configForJob(cfg, job)
	if err != nil {
		t.Fatal(err)
	}
	if got := effective.Filters.Movies.MinRating; got != 7 {
		t.Fatalf("inherited minimum rating = %v, want 7", got)
	}
	cfg.TitleExceptions.BlockedMovieTMDBIDs = []int{2}
	effective, err = configForJob(cfg, job)
	if err != nil {
		t.Fatal(err)
	}
	blocked := effective.Filters.Movies.BlacklistedTMDBIds
	if len(blocked) != 2 || blocked[0] != 1 || blocked[1] != 2 {
		t.Fatalf("effective blocked TMDB IDs = %v, want [1 2]", blocked)
	}
	if cfg.Filters.Movies.MinRating != 7 {
		t.Fatal("custom filters mutated the global config")
	}
}

func TestSupportsSource(t *testing.T) {
	if !SupportsSource("trending", "tmdb") || !SupportsSource("watched", "simkl") {
		t.Fatal("expected supported source")
	}
	if SupportsSource("collected", "tmdb") {
		t.Fatal("TMDB must not support collected jobs")
	}
}
