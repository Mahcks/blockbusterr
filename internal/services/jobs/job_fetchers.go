package jobs

import (
	"context"

	"github.com/mahcks/blockbusterr/internal/integrations"
)

// Movie Fetchers

func fetchTrendingMovies(ctx context.Context, trakt *integrations.Trakt, limit int, _ string) ([]integrations.Movie, error) {
	trendingMovies, err := trakt.GetTrendingMovies(ctx, limit)
	if err != nil {
		return nil, err
	}
	movies := make([]integrations.Movie, len(trendingMovies))
	for i, tm := range trendingMovies {
		movies[i] = tm.Movie
	}
	return movies, nil
}

func fetchPopularMovies(ctx context.Context, trakt *integrations.Trakt, limit int, _ string) ([]integrations.Movie, error) {
	return trakt.GetPopularMovies(ctx, limit)
}

func fetchAnticipatedMovies(ctx context.Context, trakt *integrations.Trakt, limit int, _ string) ([]integrations.Movie, error) {
	anticipatedMovies, err := trakt.GetAnticipatedMovies(ctx, limit)
	if err != nil {
		return nil, err
	}
	movies := make([]integrations.Movie, len(anticipatedMovies))
	for i, am := range anticipatedMovies {
		movies[i] = am.Movie
	}
	return movies, nil
}

func fetchWatchedMovies(ctx context.Context, trakt *integrations.Trakt, limit int, period string) ([]integrations.Movie, error) {
	watchedMovies, err := trakt.GetWatchedMovies(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	movies := make([]integrations.Movie, len(watchedMovies))
	for i, wm := range watchedMovies {
		movies[i] = wm.Movie
	}
	return movies, nil
}

func fetchCollectedMovies(ctx context.Context, trakt *integrations.Trakt, limit int, period string) ([]integrations.Movie, error) {
	collectedMovies, err := trakt.GetCollectedMovies(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	movies := make([]integrations.Movie, len(collectedMovies))
	for i, cm := range collectedMovies {
		movies[i] = cm.Movie
	}
	return movies, nil
}

func fetchFavoritedMovies(ctx context.Context, trakt *integrations.Trakt, limit int, period string) ([]integrations.Movie, error) {
	favoritedMovies, err := trakt.GetFavoritedMovies(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	movies := make([]integrations.Movie, len(favoritedMovies))
	for i, fm := range favoritedMovies {
		movies[i] = fm.Movie
	}
	return movies, nil
}

func fetchPlayedMovies(ctx context.Context, trakt *integrations.Trakt, limit int, period string) ([]integrations.Movie, error) {
	playedMovies, err := trakt.GetPlayedMovies(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	movies := make([]integrations.Movie, len(playedMovies))
	for i, pm := range playedMovies {
		movies[i] = pm.Movie
	}
	return movies, nil
}

func fetchBoxOfficeMovies(ctx context.Context, trakt *integrations.Trakt, limit int, _ string) ([]integrations.Movie, error) {
	boxOfficeMovies, err := trakt.GetBoxOfficeMovies(ctx, limit)
	if err != nil {
		return nil, err
	}
	movies := make([]integrations.Movie, len(boxOfficeMovies))
	for i, bom := range boxOfficeMovies {
		movies[i] = bom.Movie
	}
	return movies, nil
}

// Show Fetchers

func fetchTrendingShows(ctx context.Context, trakt *integrations.Trakt, limit int, _ string) ([]integrations.Show, error) {
	trendingShows, err := trakt.GetTrendingShows(ctx, limit)
	if err != nil {
		return nil, err
	}
	shows := make([]integrations.Show, len(trendingShows))
	for i, ts := range trendingShows {
		shows[i] = ts.Show
	}
	return shows, nil
}

func fetchPopularShows(ctx context.Context, trakt *integrations.Trakt, limit int, _ string) ([]integrations.Show, error) {
	return trakt.GetPopularShows(ctx, limit)
}

func fetchAnticipatedShows(ctx context.Context, trakt *integrations.Trakt, limit int, _ string) ([]integrations.Show, error) {
	anticipatedShows, err := trakt.GetAnticipatedShows(ctx, limit)
	if err != nil {
		return nil, err
	}
	shows := make([]integrations.Show, len(anticipatedShows))
	for i, as := range anticipatedShows {
		shows[i] = as.Show
	}
	return shows, nil
}

func fetchWatchedShows(ctx context.Context, trakt *integrations.Trakt, limit int, period string) ([]integrations.Show, error) {
	watchedShows, err := trakt.GetWatchedShows(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	shows := make([]integrations.Show, len(watchedShows))
	for i, ws := range watchedShows {
		shows[i] = ws.Show
	}
	return shows, nil
}

func fetchCollectedShows(ctx context.Context, trakt *integrations.Trakt, limit int, period string) ([]integrations.Show, error) {
	collectedShows, err := trakt.GetCollectedShows(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	shows := make([]integrations.Show, len(collectedShows))
	for i, cs := range collectedShows {
		shows[i] = cs.Show
	}
	return shows, nil
}

func fetchFavoritedShows(ctx context.Context, trakt *integrations.Trakt, limit int, period string) ([]integrations.Show, error) {
	favoritedShows, err := trakt.GetFavoritedShows(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	shows := make([]integrations.Show, len(favoritedShows))
	for i, fs := range favoritedShows {
		shows[i] = fs.Show
	}
	return shows, nil
}

func fetchPlayedShows(ctx context.Context, trakt *integrations.Trakt, limit int, period string) ([]integrations.Show, error) {
	playedShows, err := trakt.GetPlayedShows(ctx, period, limit)
	if err != nil {
		return nil, err
	}
	shows := make([]integrations.Show, len(playedShows))
	for i, ps := range playedShows {
		shows[i] = ps.Show
	}
	return shows, nil
}
