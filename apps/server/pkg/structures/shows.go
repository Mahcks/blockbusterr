package structures

type ShowSettings struct {
	ID                       int                           `json:"id"`
	Interval                 *int                          `json:"interval,omitempty"`         // The rate at which shows are pulled from show databases like Trakt (in hours)
	Anticipated              *int                          `json:"anticipated,omitempty"`      // How many shows after every interval will grab from the anticipated list
	Popular                  *int                          `json:"popular,omitempty"`          // How many shows after every interval will grab from the popular list
	Trending                 *int                          `json:"trending,omitempty"`         // How many shows after every interval will grab from the trending list
	MaxRuntime               *int                          `json:"max_runtime,omitempty"`      // Blacklist shows with runtime longer than the specified time (in minutes)
	MinRuntime               *int                          `json:"min_runtime,omitempty"`      // Blacklist shows with runtime shorter than the specified time (in minutes)
	MinYear                  *int                          `json:"min_year,omitempty"`         // Blacklist shows released before the specified year. If empty, ignore the year.
	MaxYear                  *int                          `json:"max_year,omitempty"`         // Blacklist shows released after the specified year. If empty, use the current year.
	AllowedCountries         []ShowAllowedCountry          `json:"allowed_countries"`          // List of allowed countries
	AllowedLanguages         []ShowAllowedLanguage         `json:"allowed_languages"`          // List of allowed languages
	BlacklistedGenres        []BlacklistedShowGenre        `json:"blacklisted_genres"`         // List of blacklisted genres
	BlacklistedNetworks      []BlacklistedNetwork          `json:"blacklisted_networks"`       // List of blacklisted networks
	BlacklistedTitleKeywords []BlacklistedShowTitleKeyword `json:"blacklisted_title_keywords"` // List of blacklisted title keywords
	BlacklistedTVDBIDs       []BlacklistedTVDBID           `json:"blacklisted_tvdb_ids"`       // List of blacklisted TVDB IDs
}

type ShowAllowedCountry struct {
	ID          int    `json:"id"`           // Primary key with auto-increment
	CountryCode string `json:"country_code"` // ISO 3166-1 alpha-2 country code
}

type ShowAllowedLanguage struct {
	ID           int    `json:"id"`            // Primary key with auto-increment
	LanguageCode string `json:"language_code"` // ISO 639-1 language code
}

type BlacklistedShowGenre struct {
	ID    int    `json:"id"`    // Primary key with auto-increment
	Genre string `json:"genre"` // Genre to blacklist
}

type BlacklistedNetwork struct {
	ID      int    `json:"id"`      // Primary key with auto-increment
	Network string `json:"network"` // Network to blacklist (e.g., 'Netflix', 'HBO')
}

type BlacklistedShowTitleKeyword struct {
	ID      int    `json:"id"`      // Primary key with auto-increment
	Keyword string `json:"keyword"` // Keyword to blacklist from the title of a show
}

type BlacklistedTVDBID struct {
	ID     int `json:"id"`      // Primary key with auto-increment
	TVDBID int `json:"tvdb_id"` // TVDB ID to blacklist
}
