package trakt

import (
	"context"
	"database/sql"

	"github.com/dghubble/sling"
	"github.com/mahcks/blockbusterr/internal/global"
	"github.com/mahcks/blockbusterr/pkg/errors"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

type Service interface {
	GetTrendingMovies(ctx context.Context, params *GetTrendingMoviesParams) (GetTrendingMoviesResponse, error)
	GetPopularMovies(ctx context.Context, params *GetPopularMoviesParams) (GetPopularMoviesResponse, error)
	GetAnticipatedMovies(ctx context.Context, params *GetAnticipatedMoviesParams) ([]TraktAnticipatedMovie, error)
	GetBoxOfficeMovies(ctx context.Context, params *GetBoxOfficeMoviesParams) ([]TraktBoxOfficeMovie, error)
	GetMostWatchedMovies(ctx context.Context, params *GetMostWatchedMoviesParams) ([]TraktMostWatchedMovie, error)
	GetMostPlayedMovies(ctx context.Context, params *GetMostPlayedMoviesParams) (GetMostPlayedMoviesResponse, error)
}

type traktService struct {
	gctx global.Context
	base *sling.Sling
}

func (t *traktService) FetchClientIDFromDB(ctx context.Context) (string, error) {
	var clientID string

	// Use parameterized query with context to prevent SQL injection
	query := `SELECT value FROM settings WHERE key = ?`

	err := t.gctx.Crate().SQL.DB().QueryRowContext(ctx, query, structures.SettingTraktClientID).Scan(&clientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.ErrMissingEnvironmentVariable().SetDetail("%s is missing", structures.SettingTraktClientID.String())
		}
		return "", err
	}
	return clientID, nil
}

type GetTrendingMoviesParams struct {
	// Either `full` or `metadata`
	Extended string `url:"extended,omitempty"`
}

type GetTrendingMoviesResponse []structures.TraktMovie

func (t *traktService) GetTrendingMovies(ctx context.Context, params *GetTrendingMoviesParams) (GetTrendingMoviesResponse, error) {
	clientID, err := t.FetchClientIDFromDB(ctx)
	if err != nil {
		return nil, err
	}

	var response GetTrendingMoviesResponse
	_, err = t.base.New().Set("trakt-api-key", clientID).QueryStruct(params).Get("/movies/trending").ReceiveSuccess(&response)
	if err != nil {
		return nil, err
	}

	return response, err
}

type GetPopularMoviesParams struct {
	// Either `full` or `metadata`
	Extended string `url:"extended,omitempty"`
}

type GetPopularMoviesResponse []structures.TraktMovie

func (t *traktService) GetPopularMovies(ctx context.Context, params *GetPopularMoviesParams) (GetPopularMoviesResponse, error) {
	clientID, err := t.FetchClientIDFromDB(ctx)
	if err != nil {
		return nil, err
	}
	var response GetPopularMoviesResponse
	_, err = t.base.New().Set("trakt-api-key", clientID).QueryStruct(params).Get("/movies/popular").ReceiveSuccess(&response)
	if err != nil {
		return nil, err
	}

	return response, err
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
	clientID, err := t.FetchClientIDFromDB(ctx)
	if err != nil {
		return nil, err
	}

	var movies []TraktAnticipatedMovie
	_, err = t.base.New().Set("trakt-api-key", clientID).QueryStruct(params).Get("/movies/anticipated").ReceiveSuccess(&movies)
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
	clientID, err := t.FetchClientIDFromDB(ctx)
	if err != nil {
		return nil, err
	}

	var movies []TraktBoxOfficeMovie
	_, err = t.base.New().Set("trakt-api-key", clientID).QueryStruct(params).Get("/movies/boxoffice").ReceiveSuccess(&movies)
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
	clientID, err := t.FetchClientIDFromDB(ctx)
	if err != nil {
		return nil, err
	}

	var movies []TraktMostWatchedMovie
	_, err = t.base.New().Set("trakt-api-key", clientID).QueryStruct(params).Get("/movies/watched").ReceiveSuccess(&movies)
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
	clientID, err := t.FetchClientIDFromDB(ctx)
	if err != nil {
		return GetMostPlayedMoviesResponse{}, err
	}

	var response GetMostPlayedMoviesResponse

	var movies []MostPlayedMovie
	_, err = t.base.New().Set("trakt-api-key", clientID).QueryStruct(params).Get("/movies/played").ReceiveSuccess(&movies)
	if err != nil {
		return GetMostPlayedMoviesResponse{}, err
	}

	response.Movies = movies

	return response, err
}
