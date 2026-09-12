package jobs

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

// ListResult is the provider-neutral output consumed by the existing job pipeline.
type ListResult struct {
	Source   string
	Name     string
	Movies   []integrations.Movie
	Shows    []integrations.Show
	Warnings []string
}

type ListInspection struct {
	Source   string   `json:"source"`
	Name     string   `json:"name,omitempty"`
	Movies   int      `json:"movies"`
	Shows    int      `json:"shows"`
	Warnings []string `json:"warnings,omitempty"`
}

type ListSource interface {
	FetchList(context.Context, config.ListLocator, string, int) (ListResult, error)
}

type ListSourceRegistry map[string]ListSource

type ListSourceFactory func(*config.Config) (ListSource, error)

var listSourceFactories = map[string]ListSourceFactory{}

// RegisterListSource registers only list-capable adapters; chart providers remain separate.
func RegisterListSource(provider string, factory ListSourceFactory) {
	listSourceFactories[provider] = factory
}

func AvailableListSources(cfg *config.Config) []string {
	providers := make([]string, 0, len(listSourceFactories))
	for provider, factory := range listSourceFactories {
		if _, err := factory(cfg); err == nil {
			providers = append(providers, provider)
		}
	}
	slices.Sort(providers)
	return providers
}

func IsJobSourceConfigured(cfg *config.Config, job config.DynamicJob) bool {
	if job.Type == string(enums.JobTypeList) {
		return slices.Contains(AvailableListSources(cfg), job.Source)
	}
	if job.Type == string(enums.JobTypeRecommendations) && job.RecommendationList != nil {
		return IsProviderConfigured(cfg, job.Source) && slices.Contains(AvailableListSources(cfg), job.RecommendationList.Source)
	}
	return IsProviderConfigured(cfg, job.Source)
}

func configuredListSource(cfg *config.Config, provider string) (ListSource, error) {
	factory := listSourceFactories[provider]
	if factory == nil {
		return nil, fmt.Errorf("%s list adapter is unavailable", provider)
	}
	return factory(cfg)
}

func InspectListSource(ctx context.Context, cfg *config.Config, provider string, locator config.ListLocator) (ListInspection, error) {
	if err := ValidateListSourceLocator(provider, locator); err != nil {
		return ListInspection{}, err
	}
	adapter, err := configuredListSource(cfg, provider)
	if err != nil {
		return ListInspection{}, err
	}
	result, err := adapter.FetchList(ctx, locator, "", 5)
	if err != nil {
		return ListInspection{}, err
	}
	result = normalizeListResult(result)
	return ListInspection{Source: provider, Name: result.Name, Movies: len(result.Movies), Shows: len(result.Shows), Warnings: result.Warnings}, nil
}

func ValidateListSourceLocator(provider string, locator config.ListLocator) error {
	if err := ValidateListLocator(locator); err != nil {
		return err
	}
	if provider == "letterboxd" && strings.TrimSpace(locator.Owner) == "" {
		return fmt.Errorf("letterboxd requires a public member name")
	}
	return nil
}

func ValidateListLocator(locator config.ListLocator) error {
	kind := enums.ListKind(locator.Kind)
	if !kind.Valid() {
		return fmt.Errorf("list kind must be public_list or watchlist")
	}
	if !enums.ListOrder(locator.Ordering).Valid() {
		return fmt.Errorf("list ordering must be source or rank")
	}
	for _, field := range []struct{ name, value string }{{"owner", locator.Owner}, {"list ID", locator.ListID}, {"slug", locator.Slug}} {
		if strings.Contains(field.value, "://") || strings.ContainsAny(field.value, "/?#") {
			return fmt.Errorf("%s must be a provider identifier, not a URL", field.name)
		}
	}
	if kind == enums.ListKindPublicList && strings.TrimSpace(locator.ListID) == "" && strings.TrimSpace(locator.Slug) == "" {
		return fmt.Errorf("public lists require a list ID or slug")
	}
	return nil
}

func normalizeListResult(result ListResult) ListResult {
	movieIDs, showIDs := map[string]bool{}, map[string]bool{}
	movies := make([]integrations.Movie, 0, len(result.Movies))
	shows := make([]integrations.Show, 0, len(result.Shows))
	for _, movie := range result.Movies {
		key := integrations.MovieKey(movie.IDs)
		if key == "" || movieIDs[key] {
			continue
		}
		movieIDs[key] = true
		movies = append(movies, movie)
	}
	for _, show := range result.Shows {
		key := integrations.ShowKey(show.IDs)
		if key == "" || showIDs[key] {
			continue
		}
		showIDs[key] = true
		shows = append(shows, show)
	}
	result.Movies, result.Shows = movies, shows
	return result
}
