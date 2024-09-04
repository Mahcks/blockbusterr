package structures

// TraktMovie is a struct that represents a movie from the Trakt API
type TraktMovie struct {
	Title string        `json:"title"`
	Year  int           `json:"year"`
	IDs   TraktMovieIDs `json:"ids"`
}

// TraktMovieIDs is a struct that represents the IDs of a movie from the Trakt API
type TraktMovieIDs struct {
	Trakt int    `json:"trakt"`
	Slug  string `json:"slug"`
	IMDB  string `json:"imdb"`
	TMDB  int    `json:"tmdb"`
}
