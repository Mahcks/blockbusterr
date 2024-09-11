package structures

type MovieSettings struct {
	ID                       int                       `json:"id"`
	Interval                 *int                      `json:"interval,omitempty"`         // The rate at which movies are pulled from movie databases like Trakt (in hours)
	Anticipated              *int                      `json:"anticipated,omitempty"`      // How many movies after every interval will grab from the anticipated list
	BoxOffice                *int                      `json:"box_office,omitempty"`       // How many movies after every interval will grab from the box office list
	Popular                  *int                      `json:"popular,omitempty"`          // How many movies after every interval will grab from the popular list
	Trending                 *int                      `json:"trending,omitempty"`         // How many movies after every interval will grab from the trending list
	MaxRuntime               *int                      `json:"max_runtime,omitempty"`      // Blacklist movies with runtime longer than the specified time (in minutes)
	MinRuntime               *int                      `json:"min_runtime,omitempty"`      // Blacklist movies with runtime shorter than the specified time (in minutes)
	MinYear                  *int                      `json:"min_year,omitempty"`         // Blacklist movies released before the specified year. If empty, ignore the year.
	MaxYear                  *int                      `json:"max_year,omitempty"`         // Blacklist movies released after the specified year. If empty, use the current year.
	RottenTomatoes           *string                   `json:"rotten_tomatoes,omitempty"`  // Rotten Tomatoes rating filter for movies
	AllowedCountries         []MovieAllowedCountry     `json:"allowed_countries"`          // List of allowed countries
	AllowedLanguages         []MovieAllowedLanguage    `json:"allowed_languages"`          // List of allowed languages
	BlacklistedGenres        []BlacklistedGenre        `json:"blacklisted_genres"`         // List of blacklisted genres
	BlacklistedTitleKeywords []BlacklistedTitleKeyword `json:"blacklisted_title_keywords"` // List of blacklisted title keywords
	BlacklistedTMDBIDs       []BlacklistedTMDBID       `json:"blacklisted_tmdb_ids"`       // List of blacklisted TMDb IDs
}

type MovieAllowedCountry struct {
	ID          int    `json:"id"`           // Primary key with auto-increment
	CountryCode string `json:"country_code"` // ISO 3166-1 alpha-2 country code
}

type MovieAllowedLanguage struct {
	ID           int    `json:"id"`            // Primary key with auto-increment
	LanguageCode string `json:"language_code"` // ISO 639-1 language code
}

type BlacklistedGenre struct {
	ID    int    `json:"id"`    // Primary key with auto-increment
	Genre string `json:"genre"` // Genre to blacklist
}

type BlacklistedTitleKeyword struct {
	ID      int    `json:"id"`      // Primary key with auto-increment
	Keyword string `json:"keyword"` // Keyword to blacklist from the title of a movie
}

type BlacklistedTMDBID struct {
	ID     int `json:"id"`      // Primary key with auto-increment
	TMDBID int `json:"tmdb_id"` // TMDb ID to blacklist
}
