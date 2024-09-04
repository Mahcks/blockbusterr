package trakt

import (
	"context"

	"github.com/dghubble/sling"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

type Service interface {
	GetTrendingMovies(ctx context.Context, params *GetTrendingMoviesParams) ([]structures.TraktMovie, error)
	GetPopularMovies(ctx context.Context, params *GetPopularMoviesParams) ([]structures.TraktMovie, error)
	GetAnticipatedMovies(ctx context.Context, params *GetAnticipatedMoviesParams) ([]TraktAnticipatedMovie, error)
	GetBoxOfficeMovies(ctx context.Context, params *GetBoxOfficeMoviesParams) ([]TraktBoxOfficeMovie, error)
	GetMostWatchedMovies(ctx context.Context, params *GetMostWatchedMoviesParams) ([]TraktMostWatchedMovie, error)
	GetMostPlayedMovies(ctx context.Context, params *GetMostPlayedMoviesParams) (GetMostPlayedMoviesResponse, error)
}

type traktService struct {
	base *sling.Sling
}

type GetTrendingMoviesParams struct {
	// Either `full` or `metadata`
	Extended string `url:"extended,omitempty"`
}

func (t *traktService) GetTrendingMovies(ctx context.Context, params *GetTrendingMoviesParams) ([]structures.TraktMovie, error) {
	var movies []structures.TraktMovie
	_, err := t.base.New().QueryStruct(params).Get("/movies/trending").ReceiveSuccess(&movies)
	if err != nil {
		return nil, err
	}

	return movies, err
}

type GetPopularMoviesParams struct {
	// Either `full` or `metadata`
	Extended string `url:"extended,omitempty"`
}

func (t *traktService) GetPopularMovies(ctx context.Context, params *GetPopularMoviesParams) ([]structures.TraktMovie, error) {
	var movies []structures.TraktMovie
	_, err := t.base.New().QueryStruct(params).Get("/movies/popular").ReceiveSuccess(&movies)
	if err != nil {
		return nil, err
	}

	return movies, err
}

type GetAnticipatedMoviesParams struct {
	// Either `full` or `metadata`
	Extended string `url:"extended,omitempty"`
}

type TraktAnticipatedMovie struct {
	ListCount int                   `json:"list_count"`
	Movie     structures.TraktMovie `json:"movie"`
}

func (t *traktService) GetAnticipatedMovies(ctx context.Context, params *GetAnticipatedMoviesParams) ([]TraktAnticipatedMovie, error) {
	var movies []TraktAnticipatedMovie
	_, err := t.base.New().QueryStruct(params).Get("/movies/anticipated").ReceiveSuccess(&movies)
	if err != nil {
		return nil, err
	}

	return movies, err
}

type GetBoxOfficeMoviesParams struct {
	// Either `full` or `metadata`
	Extended string `url:"extended,omitempty"`
}

type TraktBoxOfficeMovie struct {
	Revenue int                   `json:"revenue"`
	Movie   structures.TraktMovie `json:"movie"`
}

func (t *traktService) GetBoxOfficeMovies(ctx context.Context, params *GetBoxOfficeMoviesParams) ([]TraktBoxOfficeMovie, error) {
	var movies []TraktBoxOfficeMovie
	_, err := t.base.New().QueryStruct(params).Get("/movies/boxoffice").ReceiveSuccess(&movies)
	if err != nil {
		return nil, err
	}

	return movies, err
}

type GetMostWatchedMoviesParams struct {
	// Either `full` or `metadata`
	Extended string `url:"extended,omitempty"`
	Period   string `url:"period,omitempty"`
}

type TraktMostWatchedMovie struct {
	WatcherCount   int                   `json:"watcher_count"`
	PlayCount      int                   `json:"play_count"`
	CollectedCount int                   `json:"collected"`
	Movie          structures.TraktMovie `json:"movie"`
}

func (t *traktService) GetMostWatchedMovies(ctx context.Context, params *GetMostWatchedMoviesParams) ([]TraktMostWatchedMovie, error) {
	var movies []TraktMostWatchedMovie
	_, err := t.base.New().QueryStruct(params).Get("/movies/watched").ReceiveSuccess(&movies)
	if err != nil {
		return nil, err
	}

	return movies, err
}

type GetMostPlayedMoviesParams struct {
	// Either `full` or `metadata`
	Extended string `url:"extended,omitempty"`
	Period   string `url:"period,omitempty"`
}

type MostPlayedMovie struct {
	WatcherCount   int                   `json:"watcher_count"`
	PlayCount      int                   `json:"play_count"`
	CollectedCount int                   `json:"collected_count"`
	Movie          structures.TraktMovie `json:"movie"`
}

type GetMostPlayedMoviesResponse struct {
	Movies []MostPlayedMovie `json:"movies"`
}

func (t *traktService) GetMostPlayedMovies(ctx context.Context, params *GetMostPlayedMoviesParams) (GetMostPlayedMoviesResponse, error) {
	var response GetMostPlayedMoviesResponse

	var movies []MostPlayedMovie
	_, err := t.base.New().QueryStruct(params).Get("/movies/played").ReceiveSuccess(&movies)
	if err != nil {
		return GetMostPlayedMoviesResponse{}, err
	}

	response.Movies = movies

	return response, err
}
