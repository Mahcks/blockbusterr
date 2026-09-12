package jobs

import (
	"context"
	"fmt"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/integrations"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

// DiscoveryClient is the small provider boundary used by existing job fetchers.
type DiscoveryClient struct {
	provider enums.DiscoveryProvider
	trakt    *integrations.Trakt
	tmdb     *integrations.TMDB
	simkl    *integrations.Simkl
	list     ListSource
}

func newListDiscoveryClient(source string, adapter ListSource) *DiscoveryClient {
	return &DiscoveryClient{provider: enums.DiscoveryProvider(source), list: adapter}
}

func NewDiscoveryClient(cfg *config.Config, source string) (*DiscoveryClient, error) {
	if source == "" {
		source = string(enums.DiscoveryProviderTrakt)
	}
	provider := enums.DiscoveryProvider(source)
	if !provider.Valid() {
		return nil, fmt.Errorf("unsupported discovery provider: %s", source)
	}
	if !IsProviderConfigured(cfg, source) {
		return nil, fmt.Errorf("%s credentials are not configured", provider)
	}
	client := &DiscoveryClient{provider: provider}
	switch provider {
	case enums.DiscoveryProviderTrakt:
		client.trakt = integrations.NewTrakt(integrations.TraktConfig{ClientID: cfg.Trakt.ClientID, ClientSecret: cfg.Trakt.ClientSecret})
	case enums.DiscoveryProviderTMDB:
		client.tmdb = integrations.NewTMDB(integrations.TMDBConfig{APIKey: cfg.TMDB.APIKey})
	case enums.DiscoveryProviderSimkl:
		client.simkl = integrations.NewSimkl(integrations.SimklConfig{ClientID: cfg.Simkl.ClientID})
	}
	return client, nil
}

// IsProviderConfigured reports whether a discovery source has the credential
// required to make requests. Empty sources retain the legacy Trakt default.
func IsProviderConfigured(cfg *config.Config, source string) bool {
	switch enums.DiscoveryProvider(source) {
	case "", enums.DiscoveryProviderTrakt:
		return cfg.Trakt.ClientID != ""
	case enums.DiscoveryProviderTMDB:
		return cfg.TMDB.APIKey != ""
	case enums.DiscoveryProviderSimkl:
		return cfg.Simkl.ClientID != ""
	default:
		return false
	}
}

func (d *DiscoveryClient) Source() string { return string(d.provider) }

func (d *DiscoveryClient) GetListMovies(ctx context.Context, locator config.ListLocator, limit int) ([]integrations.Movie, error) {
	result, err := d.getList(ctx, locator, string(enums.MediaTypeMovie), limit)
	return result.Movies, err
}

func (d *DiscoveryClient) GetListShows(ctx context.Context, locator config.ListLocator, limit int) ([]integrations.Show, error) {
	result, err := d.getList(ctx, locator, string(enums.MediaTypeShow), limit)
	return result.Shows, err
}

func (d *DiscoveryClient) getList(ctx context.Context, locator config.ListLocator, mediaType string, limit int) (ListResult, error) {
	if err := ValidateListSourceLocator(d.Source(), locator); err != nil {
		return ListResult{}, err
	}
	if d.list == nil {
		return ListResult{}, fmt.Errorf("%s list adapter is unavailable", d.provider)
	}
	result, err := d.list.FetchList(ctx, locator, mediaType, limit)
	if err != nil {
		return ListResult{}, err
	}
	if result.Source == "" {
		result.Source = d.Source()
	}
	return normalizeListResult(result), nil
}

func unsupported(provider enums.DiscoveryProvider, jobType string) error {
	return fmt.Errorf("%s does not support %s jobs", provider, jobType)
}

func wrapTrendingMovies(movies []integrations.Movie) []integrations.TrendingMovie {
	result := make([]integrations.TrendingMovie, len(movies))
	for i, movie := range movies {
		result[i] = integrations.TrendingMovie{Movie: movie}
	}
	return result
}

func wrapTrendingShows(shows []integrations.Show) []integrations.TrendingShow {
	result := make([]integrations.TrendingShow, len(shows))
	for i, show := range shows {
		result[i] = integrations.TrendingShow{Show: show}
	}
	return result
}

func simklTimeframe(period string) string {
	switch period {
	case "monthly", "yearly", "all":
		return "month"
	case "weekly":
		return "week"
	default:
		return "today"
	}
}

func (d *DiscoveryClient) GetTrendingMovies(ctx context.Context, limit int) ([]integrations.TrendingMovie, error) {
	switch d.provider {
	case enums.DiscoveryProviderTrakt:
		return d.trakt.GetTrendingMovies(ctx, limit)
	case enums.DiscoveryProviderTMDB:
		movies, err := d.tmdb.GetTrendingMovies(ctx, limit)
		return wrapTrendingMovies(movies), err
	case enums.DiscoveryProviderSimkl:
		movies, err := d.simkl.GetMovies(ctx, "today", limit)
		return wrapTrendingMovies(movies), err
	default:
		return nil, unsupported(d.provider, "trending movie")
	}
}

func (d *DiscoveryClient) GetPopularMovies(ctx context.Context, limit int) ([]integrations.Movie, error) {
	switch d.provider {
	case enums.DiscoveryProviderTrakt:
		return d.trakt.GetPopularMovies(ctx, limit)
	case enums.DiscoveryProviderTMDB:
		return d.tmdb.GetPopularMovies(ctx, limit)
	case enums.DiscoveryProviderSimkl:
		return d.simkl.GetMovies(ctx, "month", limit)
	default:
		return nil, unsupported(d.provider, "popular movie")
	}
}

func (d *DiscoveryClient) GetMovieRecommendations(ctx context.Context, seeds []int, limit int) ([]integrations.Movie, error) {
	if d.provider != enums.DiscoveryProviderTMDB {
		return nil, unsupported(d.provider, "movie recommendations")
	}
	return d.tmdb.GetMovieRecommendations(ctx, seeds, limit)
}

func (d *DiscoveryClient) GetTrendingShows(ctx context.Context, limit int) ([]integrations.TrendingShow, error) {
	switch d.provider {
	case enums.DiscoveryProviderTrakt:
		return d.trakt.GetTrendingShows(ctx, limit)
	case enums.DiscoveryProviderTMDB:
		shows, err := d.tmdb.GetTrendingShows(ctx, limit)
		return wrapTrendingShows(shows), err
	case enums.DiscoveryProviderSimkl:
		shows, err := d.simkl.GetShows(ctx, "today", limit)
		return wrapTrendingShows(shows), err
	default:
		return nil, unsupported(d.provider, "trending show")
	}
}

func (d *DiscoveryClient) GetPopularShows(ctx context.Context, limit int) ([]integrations.Show, error) {
	switch d.provider {
	case enums.DiscoveryProviderTrakt:
		return d.trakt.GetPopularShows(ctx, limit)
	case enums.DiscoveryProviderTMDB:
		return d.tmdb.GetPopularShows(ctx, limit)
	case enums.DiscoveryProviderSimkl:
		return d.simkl.GetShows(ctx, "month", limit)
	default:
		return nil, unsupported(d.provider, "popular show")
	}
}

func (d *DiscoveryClient) GetShowRecommendations(ctx context.Context, seeds []int, limit int) ([]integrations.Show, error) {
	if d.provider != enums.DiscoveryProviderTMDB {
		return nil, unsupported(d.provider, "show recommendations")
	}
	return d.tmdb.GetShowRecommendations(ctx, seeds, limit)
}

func (d *DiscoveryClient) GetWatchedMovies(ctx context.Context, period string, limit int) ([]integrations.WatchedMovie, error) {
	if d.provider == enums.DiscoveryProviderTrakt {
		return d.trakt.GetWatchedMovies(ctx, period, limit)
	}
	if d.provider == enums.DiscoveryProviderSimkl {
		movies, err := d.simkl.GetMovies(ctx, simklTimeframe(period), limit)
		result := make([]integrations.WatchedMovie, len(movies))
		for i, movie := range movies {
			result[i].Movie = movie
		}
		return result, err
	}
	return nil, unsupported(d.provider, "most watched movie")
}

func (d *DiscoveryClient) GetWatchedShows(ctx context.Context, period string, limit int) ([]integrations.WatchedShow, error) {
	if d.provider == enums.DiscoveryProviderTrakt {
		return d.trakt.GetWatchedShows(ctx, period, limit)
	}
	if d.provider == enums.DiscoveryProviderSimkl {
		shows, err := d.simkl.GetShows(ctx, simklTimeframe(period), limit)
		result := make([]integrations.WatchedShow, len(shows))
		for i, show := range shows {
			result[i].Show = show
		}
		return result, err
	}
	return nil, unsupported(d.provider, "most watched show")
}

func (d *DiscoveryClient) GetAnticipatedMovies(ctx context.Context, limit int) ([]integrations.AnticipatedMovie, error) {
	if d.provider != enums.DiscoveryProviderTrakt {
		return nil, unsupported(d.provider, "anticipated movie")
	}
	return d.trakt.GetAnticipatedMovies(ctx, limit)
}
func (d *DiscoveryClient) GetCollectedMovies(ctx context.Context, period string, limit int) ([]integrations.CollectedMovie, error) {
	if d.provider != enums.DiscoveryProviderTrakt {
		return nil, unsupported(d.provider, "collected movie")
	}
	return d.trakt.GetCollectedMovies(ctx, period, limit)
}
func (d *DiscoveryClient) GetFavoritedMovies(ctx context.Context, period string, limit int) ([]integrations.FavoritedMovie, error) {
	if d.provider != enums.DiscoveryProviderTrakt {
		return nil, unsupported(d.provider, "favorited movie")
	}
	return d.trakt.GetFavoritedMovies(ctx, period, limit)
}
func (d *DiscoveryClient) GetPlayedMovies(ctx context.Context, period string, limit int) ([]integrations.PlayedMovie, error) {
	if d.provider != enums.DiscoveryProviderTrakt {
		return nil, unsupported(d.provider, "played movie")
	}
	return d.trakt.GetPlayedMovies(ctx, period, limit)
}
func (d *DiscoveryClient) GetBoxOfficeMovies(ctx context.Context, limit int) ([]integrations.BoxOfficeMovie, error) {
	if d.provider != enums.DiscoveryProviderTrakt {
		return nil, unsupported(d.provider, "box office")
	}
	return d.trakt.GetBoxOfficeMovies(ctx, limit)
}
func (d *DiscoveryClient) GetAnticipatedShows(ctx context.Context, limit int) ([]integrations.AnticipatedShow, error) {
	if d.provider != enums.DiscoveryProviderTrakt {
		return nil, unsupported(d.provider, "anticipated show")
	}
	return d.trakt.GetAnticipatedShows(ctx, limit)
}
func (d *DiscoveryClient) GetCollectedShows(ctx context.Context, period string, limit int) ([]integrations.CollectedShow, error) {
	if d.provider != enums.DiscoveryProviderTrakt {
		return nil, unsupported(d.provider, "collected show")
	}
	return d.trakt.GetCollectedShows(ctx, period, limit)
}
func (d *DiscoveryClient) GetFavoritedShows(ctx context.Context, period string, limit int) ([]integrations.FavoritedShow, error) {
	if d.provider != enums.DiscoveryProviderTrakt {
		return nil, unsupported(d.provider, "favorited show")
	}
	return d.trakt.GetFavoritedShows(ctx, period, limit)
}
func (d *DiscoveryClient) GetPlayedShows(ctx context.Context, period string, limit int) ([]integrations.PlayedShow, error) {
	if d.provider != enums.DiscoveryProviderTrakt {
		return nil, unsupported(d.provider, "played show")
	}
	return d.trakt.GetPlayedShows(ctx, period, limit)
}
