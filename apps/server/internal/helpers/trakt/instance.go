package trakt

import (
	"context"
	"fmt"

	"github.com/dghubble/sling"
	"github.com/mahcks/blockbusterr/internal/global"
)

type Service interface {
	GetTrendingMovies(ctx context.Context, params *TraktMovieParams) (GetTrendingMoviesResponse, error)
	GetPopularMovies(ctx context.Context, params *TraktMovieParams) (GetPopularMoviesResponse, error)
	GetAnticipatedMovies(ctx context.Context, params *TraktMovieParams) ([]TraktAnticipatedMovie, error)
	GetBoxOfficeMovies(ctx context.Context, params *TraktMovieParams) ([]TraktBoxOfficeMovie, error)
	GetMostWatchedMovies(ctx context.Context, params *TraktMovieParams) ([]TraktMostWatchedMovie, error)
	GetMostPlayedMovies(ctx context.Context, params *TraktMovieParams) (GetMostPlayedMoviesResponse, error)

	GetAnticipatedShows(ctx context.Context, params *TraktMovieParams) (GetAnticipatedShowsResponse, error)
	GetPopularShows(ctx context.Context, params *TraktMovieParams) (GetPopularShowsResponse, error)
	GetTrendingShows(ctx context.Context, params *TraktMovieParams) (GetTrendingShowsResponse, error)

	GetListItems(ctx context.Context, parms *GetListItemsParams) (GetListItemsResponse, error)
}

type traktService struct {
	gctx global.Context
	base *sling.Sling
}

// Movie params for every Trakt API movie request
type TraktMovieParams struct {
	// Either `full` or `metadata`
	Extended string `url:"extended,omitempty"`

	// 4 digit year or range of years. (Example: 2007 or 2007-2016)
	Years string `url:"years,omitempty"`
	// Genre slugs.
	Genres string `url:"genres,omitempty"`
	// Language codes (ISO 639-1)
	Languages string `url:"languages,omitempty"`
	// Country codes (ISO 3166-1)
	Countries string `url:"countries,omitempty"`
	// Minimum and maximum length of runtime in minutes. (Example: 90-120)
	Runtime string `url:"runtime,omitempty"`

	// Pagination
	Page  int `url:"page,omitempty"`
	Limit int `url:"limit,omitempty"`
}

// Movie is a struct that represents a movie from the Trakt API
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

// MovieIDs is a struct that represents the IDs of a movie from the Trakt API
type MovieIDs struct {
	Trakt int    `json:"trakt"`
	Slug  string `json:"slug"`
	IMDB  string `json:"imdb"`
	TMDB  int    `json:"tmdb"`
}

type Show struct {
	Title                 string   `json:"title"`
	Year                  int      `json:"year"`
	IDs                   ShowIDs  `json:"ids"`
	Tagline               string   `json:"tagline,omitempty"`
	Overview              string   `json:"overview,omitempty"`
	FirstAired            string   `json:"first_aired,omitempty"`
	Airs                  ShowAirs `json:"airs,omitempty"`
	Runtime               int      `json:"runtime,omitempty"`
	Certification         string   `json:"certification,omitempty"`
	Network               string   `json:"network,omitempty"`
	Country               string   `json:"country,omitempty"`
	Trailer               string   `json:"trailer,omitempty"`
	Homepage              string   `json:"homepage,omitempty"`
	Status                string   `json:"status,omitempty"`
	Rating                float64  `json:"rating,omitempty"`
	Votes                 int      `json:"votes,omitempty"`
	CommentCount          int      `json:"comment_count,omitempty"`
	UpdatedAt             string   `json:"updated_at,omitempty"`
	Language              string   `json:"language,omitempty"`
	Languages             []string `json:"languages,omitempty"`
	AvailableTranslations []string `json:"available_translations,omitempty"`
	Genres                []string `json:"genres,omitempty"`
	AiredEpisodes         int      `json:"aired_episodes,omitempty"`
}

type ShowIDs struct {
	Trakt int    `json:"trakt"`
	Slug  string `json:"slug"`
	TVDB  int    `json:"tvdb"`
	IMDB  string `json:"imdb"`
	TMDB  int    `json:"tmdb"`
}

type ShowAirs struct {
	Day      string `json:"day"`
	Time     string `json:"time"`
	Timezone string `json:"timezone"`
}

func (t *traktService) FetchClientIDFromDB(ctx context.Context) (string, error) {
	// Use parameterized query with context to prevent SQL injection
	traktSetings, err := t.gctx.Crate().SQL.Queries().GetTraktSettings(ctx)
	if err != nil {
		return "", err
	}

	return traktSetings.ClientID, nil
}

type GetTrendingMoviesResponse []TrendingMovie

type TrendingMovie struct {
	Watchers int   `json:"watchers"`
	Movie    Movie `json:"movie"`
}

func (t *traktService) GetTrendingMovies(ctx context.Context, params *TraktMovieParams) (GetTrendingMoviesResponse, error) {
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

type GetPopularMoviesResponse []Movie

func (t *traktService) GetPopularMovies(ctx context.Context, params *TraktMovieParams) (GetPopularMoviesResponse, error) {
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

type TraktAnticipatedMovie struct {
	ListCount int   `json:"list_count"`
	Movie     Movie `json:"movie"`
}

func (t *traktService) GetAnticipatedMovies(ctx context.Context, params *TraktMovieParams) ([]TraktAnticipatedMovie, error) {
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

type TraktBoxOfficeMovie struct {
	Revenue int   `json:"revenue"`
	Movie   Movie `json:"movie"`
}

func (t *traktService) GetBoxOfficeMovies(ctx context.Context, params *TraktMovieParams) ([]TraktBoxOfficeMovie, error) {
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

type TraktMostWatchedMovie struct {
	WatcherCount   int   `json:"watcher_count"`
	PlayCount      int   `json:"play_count"`
	CollectedCount int   `json:"collected"`
	Movie          Movie `json:"movie"`
}

func (t *traktService) GetMostWatchedMovies(ctx context.Context, params *TraktMovieParams) ([]TraktMostWatchedMovie, error) {
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

type MostPlayedMovie struct {
	WatcherCount   int   `json:"watcher_count"`
	PlayCount      int   `json:"play_count"`
	CollectedCount int   `json:"collected_count"`
	Movie          Movie `json:"movie"`
}

type GetMostPlayedMoviesResponse struct {
	Movies []MostPlayedMovie `json:"movies"`
}

func (t *traktService) GetMostPlayedMovies(ctx context.Context, params *TraktMovieParams) (GetMostPlayedMoviesResponse, error) {
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

type GetAnticipatedShowsResponse []AnticipatedShow

type AnticipatedShow struct {
	ListCount int  `json:"list_count"`
	Show      Show `json:"show"`
}

func (t *traktService) GetAnticipatedShows(ctx context.Context, params *TraktMovieParams) (GetAnticipatedShowsResponse, error) {
	clientID, err := t.FetchClientIDFromDB(ctx)
	if err != nil {
		return GetAnticipatedShowsResponse{}, err
	}

	var response GetAnticipatedShowsResponse
	res, err := t.base.New().Set("trakt-api-key", clientID).QueryStruct(params).Get("/shows/anticipated").ReceiveSuccess(&response)
	if err != nil {
		return GetAnticipatedShowsResponse{}, err
	}

	if res.StatusCode != 200 {
		return GetAnticipatedShowsResponse{}, fmt.Errorf("failed to get anticipated shows: %v", res.Status)
	}

	return response, err
}

type GetPopularShowsResponse []Show

func (t *traktService) GetPopularShows(ctx context.Context, params *TraktMovieParams) (GetPopularShowsResponse, error) {
	clientID, err := t.FetchClientIDFromDB(ctx)
	if err != nil {
		return GetPopularShowsResponse{}, err
	}

	var response GetPopularShowsResponse
	_, err = t.base.New().Set("trakt-api-key", clientID).QueryStruct(params).Get("/shows/popular").ReceiveSuccess(&response)
	if err != nil {
		return GetPopularShowsResponse{}, err
	}

	return response, err
}

type GetTrendingShowsResponse []TrendingShow

type TrendingShow struct {
	Watchers int  `json:"watchers"`
	Show     Show `json:"show"`
}

func (t *traktService) GetTrendingShows(ctx context.Context, params *TraktMovieParams) (GetTrendingShowsResponse, error) {
	clientID, err := t.FetchClientIDFromDB(ctx)
	if err != nil {
		return GetTrendingShowsResponse{}, err
	}

	var response GetTrendingShowsResponse
	res, err := t.base.New().Set("trakt-api-key", clientID).QueryStruct(params).Get("/shows/trending").ReceiveSuccess(&response)
	if err != nil {
		return GetTrendingShowsResponse{}, err
	}

	if res.StatusCode != 200 {
		return GetTrendingShowsResponse{}, fmt.Errorf("failed to get trending shows: %v", res.Status)
	}

	return response, err
}

type GetListItemsParams struct {
	ListID    int    `url:"id"`
	MediaType string `url:"type"`
}

type GetListItemsResponse []ListItem

type ListItem struct {
	Rake     int     `json:"rank"`
	ID       int     `json:"id"`
	ListedAt string  `json:"listed_at"`
	Notes    *string `json:"notes"`
	Type     string  `json:"type"`
	Movie    Movie   `json:"movie"`
}

func (t *traktService) GetListItems(ctx context.Context, params *GetListItemsParams) (GetListItemsResponse, error) {
	clientID, err := t.FetchClientIDFromDB(ctx)
	if err != nil {
		return GetListItemsResponse{}, err
	}

	var response GetListItemsResponse
	res, err := t.base.New().Set("trakt-api-key", clientID).QueryStruct(params).Get("/shows/trending").ReceiveSuccess(&response)
	if err != nil {
		return GetListItemsResponse{}, err
	}

	if res.StatusCode != 200 {
		return GetListItemsResponse{}, fmt.Errorf("failed to get trending shows: %v", res.Status)
	}

	return response, nil
}
