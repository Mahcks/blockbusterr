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

// TraktMovie is a struct that represents a movie from the Trakt API
type Movie struct {
	// Title of the movie
	Title string `json:"title"`
	// Year the movie was released
	Year int `json:"year"`
	// IDS of the movie
	IDs MovieIDs `json:"ids"`
	// Tagline of the movie
	Tagline string `json:"tagline,omitempty"`
	// Overview of the movie
	Overview string `json:"overview,omitempty"`
	// Released date of the movie when it was released
	Released string `json:"released,omitempty"`
	// Runtime of the movie in minutes
	Runtime int `json:"runtime,omitempty"`
	// Country where the movie was produced
	Country string `json:"country,omitempty"`
	// Status of the movie
	Status string `json:"status,omitempty"`
	// Rating of the movie
	Rating float64 `json:"rating,omitempty"`
	// Votes of the movie
	Votes int `json:"votes,omitempty"`
	// Comment count of the movie
	CommentCount int `json:"comment_count,omitempty"`
	// Trailer URL of the movie trailer
	Trailer string `json:"trailer,omitempty"`
	// Hompage is the link to the movies page
	Homepage string `json:"homepage,omitempty"`
	// UpdatedAt is the date the movie was last updated
	UpdatedAt string `json:"updated_at,omitempty"`
	// Language of the movie
	Language string `json:"language,omitempty"`
	// Languages that the movie supports
	Languages []string `json:"languages,omitempty"`
	// AvailableTranslations of the movie
	AvailableTranslations []string `json:"available_translations,omitempty"`
	// Genres of the movie
	Genres []string `json:"genres,omitempty"`
	// Certification of the movie
	Certification string `json:"certification,omitempty"`
}

// TraktMovieIDs is a struct that represents the IDs of a movie from the Trakt API
type MovieIDs struct {
	Trakt int    `json:"trakt"`
	Slug  string `json:"slug"`
	IMDB  string `json:"imdb"`
	TMDB  int    `json:"tmdb"`
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

type GetTrendingMoviesResponse []Movie

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

type GetPopularMoviesResponse []Movie

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
	ListCount int   `json:"list_count"`
	Movie     Movie `json:"movie"`
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
	Revenue int   `json:"revenue"`
	Movie   Movie `json:"movie"`
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
	WatcherCount   int   `json:"watcher_count"`
	PlayCount      int   `json:"play_count"`
	CollectedCount int   `json:"collected"`
	Movie          Movie `json:"movie"`
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
	WatcherCount   int   `json:"watcher_count"`
	PlayCount      int   `json:"play_count"`
	CollectedCount int   `json:"collected_count"`
	Movie          Movie `json:"movie"`
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
