package trakt

import (
	"context"

	"github.com/dghubble/sling"
	"github.com/mahcks/blockbusterr/pkg/structures"
)

type Service interface {
	GetPopularMovies(ctx context.Context, page int) ([]structures.TraktMovie, error)
}

type traktService struct {
	base *sling.Sling
}

type PopularMovies struct {
	Movies []structures.TraktMovie
}

func (t *traktService) GetPopularMovies(ctx context.Context, page int) ([]structures.TraktMovie, error) {
	var movies []structures.TraktMovie
	_, err := t.base.New().
		Set("X-Pagination-Page", "1").
		Set("X-Pagination-Limit", "10").
		Set("X-Pagination-Page-Count", "10").
		Set("X-Pagination-Item-Count", "100").
		Get("/movies/popular").ReceiveSuccess(&movies)
	return movies, err
}
