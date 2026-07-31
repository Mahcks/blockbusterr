package jobs

import (
	"context"

	"github.com/mahcks/blockbusterr/internal/integrations"
)

// Movie Fetchers

func fetchPopularMovies(ctx context.Context, discovery *DiscoveryClient, limit int, _ string) ([]integrations.Movie, error) {
	return discovery.GetPopularMovies(ctx, limit)
}

func fetchAnticipatedMovies(ctx context.Context, discovery *DiscoveryClient, limit int, _ string) ([]integrations.Movie, error) {
	anticipatedMovies, err := discovery.GetAnticipatedMovies(ctx, limit)
	if err != nil {
		return nil, err
	}
	movies := make([]integrations.Movie, len(anticipatedMovies))
	for i, am := range anticipatedMovies {
		movies[i] = am.Movie
	}
	return movies, nil
}

func fetchWatchedMovies(ctx context.Context, discovery *DiscoveryClient, limit int, period string) ([]integrations.Movie, error) {
	watchedMovies, err := discovery.GetWatchedMovies(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	movies := make([]integrations.Movie, len(watchedMovies))
	for i, wm := range watchedMovies {
		movies[i] = wm.Movie
	}
	return movies, nil
}

func fetchCollectedMovies(ctx context.Context, discovery *DiscoveryClient, limit int, period string) ([]integrations.Movie, error) {
	collectedMovies, err := discovery.GetCollectedMovies(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	movies := make([]integrations.Movie, len(collectedMovies))
	for i, cm := range collectedMovies {
		movies[i] = cm.Movie
	}
	return movies, nil
}

func fetchFavoritedMovies(ctx context.Context, discovery *DiscoveryClient, limit int, period string) ([]integrations.Movie, error) {
	favoritedMovies, err := discovery.GetFavoritedMovies(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	movies := make([]integrations.Movie, len(favoritedMovies))
	for i, fm := range favoritedMovies {
		movies[i] = fm.Movie
	}
	return movies, nil
}

func fetchPlayedMovies(ctx context.Context, discovery *DiscoveryClient, limit int, period string) ([]integrations.Movie, error) {
	playedMovies, err := discovery.GetPlayedMovies(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	movies := make([]integrations.Movie, len(playedMovies))
	for i, pm := range playedMovies {
		movies[i] = pm.Movie
	}
	return movies, nil
}

func fetchBoxOfficeMovies(ctx context.Context, discovery *DiscoveryClient, limit int, _ string) ([]integrations.Movie, error) {
	boxOfficeMovies, err := discovery.GetBoxOfficeMovies(ctx, limit)
	if err != nil {
		return nil, err
	}

	// Trakt API always returns 10 box office movies regardless of limit parameter
	// Slice the results to respect the configured limit
	actualLimit := len(boxOfficeMovies)
	if limit > 0 && limit < actualLimit {
		actualLimit = limit
	}

	movies := make([]integrations.Movie, actualLimit)
	for i := 0; i < actualLimit; i++ {
		movies[i] = boxOfficeMovies[i].Movie
	}
	return movies, nil
}

// Show Fetchers

func fetchTrendingShows(ctx context.Context, discovery *DiscoveryClient, limit int, _ string) ([]integrations.Show, error) {
	trendingShows, err := discovery.GetTrendingShows(ctx, limit)
	if err != nil {
		return nil, err
	}
	shows := make([]integrations.Show, len(trendingShows))
	for i, ts := range trendingShows {
		shows[i] = ts.Show
	}
	return shows, nil
}

func fetchPopularShows(ctx context.Context, discovery *DiscoveryClient, limit int, _ string) ([]integrations.Show, error) {
	return discovery.GetPopularShows(ctx, limit)
}

func fetchAnticipatedShows(ctx context.Context, discovery *DiscoveryClient, limit int, _ string) ([]integrations.Show, error) {
	anticipatedShows, err := discovery.GetAnticipatedShows(ctx, limit)
	if err != nil {
		return nil, err
	}
	shows := make([]integrations.Show, len(anticipatedShows))
	for i, as := range anticipatedShows {
		shows[i] = as.Show
	}
	return shows, nil
}

func fetchWatchedShows(ctx context.Context, discovery *DiscoveryClient, limit int, period string) ([]integrations.Show, error) {
	watchedShows, err := discovery.GetWatchedShows(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	shows := make([]integrations.Show, len(watchedShows))
	for i, ws := range watchedShows {
		shows[i] = ws.Show
	}
	return shows, nil
}

func fetchCollectedShows(ctx context.Context, discovery *DiscoveryClient, limit int, period string) ([]integrations.Show, error) {
	collectedShows, err := discovery.GetCollectedShows(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	shows := make([]integrations.Show, len(collectedShows))
	for i, cs := range collectedShows {
		shows[i] = cs.Show
	}
	return shows, nil
}

func fetchFavoritedShows(ctx context.Context, discovery *DiscoveryClient, limit int, period string) ([]integrations.Show, error) {
	favoritedShows, err := discovery.GetFavoritedShows(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	shows := make([]integrations.Show, len(favoritedShows))
	for i, fs := range favoritedShows {
		shows[i] = fs.Show
	}
	return shows, nil
}

func fetchPlayedShows(ctx context.Context, discovery *DiscoveryClient, limit int, period string) ([]integrations.Show, error) {
	playedShows, err := discovery.GetPlayedShows(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	shows := make([]integrations.Show, len(playedShows))
	for i, ps := range playedShows {
		shows[i] = ps.Show
	}
	return shows, nil
}
