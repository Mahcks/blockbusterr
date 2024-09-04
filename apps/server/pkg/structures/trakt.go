package structures

// TraktMovie is a struct that represents a movie from the Trakt API
type TraktMovie struct {
	// Title of the movie
	Title string `json:"title"`
	// Year the movie was released
	Year int `json:"year"`
	// IDS of the movie
	IDs TraktMovieIDs `json:"ids"`
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
type TraktMovieIDs struct {
	Trakt int    `json:"trakt"`
	Slug  string `json:"slug"`
	IMDB  string `json:"imdb"`
	TMDB  int    `json:"tmdb"`
}
