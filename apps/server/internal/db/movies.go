package db

import (
	"context"
	"database/sql"
	"fmt"
)

type MovieSettings struct {
	ID             int            `db:"id"`              // Primary key with auto-increment
	Interval       sql.NullInt32  `db:"interval"`        // The rate at which movies are pulled from movie databases like Trakt (in hours)
	Anticipated    sql.NullInt32  `db:"anticipated"`     // How many movies after every interval will grab from the anticipated list
	BoxOffice      sql.NullInt32  `db:"box_office"`      // How many movies after every interval will grab from the box office list
	Popular        sql.NullInt32  `db:"popular"`         // How many movies after every interval will grab from the popular list
	Trending       sql.NullInt32  `db:"trending"`        // How many movies after every interval will grab from the trending list
	MaxRuntime     sql.NullInt32  `db:"max_runtime"`     // Blacklist movies with runtime longer than the specified time (in minutes)
	MinRuntime     sql.NullInt32  `db:"min_runtime"`     // Blacklist movies with runtime shorter than the specified time (in minutes)
	MinYear        sql.NullInt32  `db:"min_year"`        // Blacklist movies released before the specified year. If left empty/is zero, it'll ignore the year.
	MaxYear        sql.NullInt32  `db:"max_year"`        // Blacklist movies released after the specified year. If left empty/is zero, it'll be the current year
	RottenTomatoes sql.NullString `db:"rotten_tomatoes"` // Rotten Tomatoes rating filter for movies

	AllowedCountries         []MovieAllowedCountries    // List of allowed countries
	AllowedLanguages         []MovieAllowedLanguages    // List of allowed languages
	BlacklistedGenres        []BlacklistedGenres        // List of blacklisted genres
	BlacklistedTitleKeywords []BlacklistedTitleKeywords // List of blacklisted title keywords
	BlacklistedTMDBIDs       []BlacklistedTMDBIDs       // List of blacklisted TMDb IDs    // Comma-separated list of blacklisted TMDb IDs
}

type MovieAllowedCountries struct {
	ID              int    `db:"id"`                // Primary key with auto-increment
	MovieSettingsID int    `db:"movie_settings_id"` // Foreign key to the movie settings table
	CountryCode     string `db:"country_code"`      // ISO 3166-1 alpha-2 country code
}

type MovieAllowedLanguages struct {
	ID              int    `db:"id"`                // Primary key with auto-increment
	MovieSettingsID int    `db:"movie_settings_id"` // Foreign key to the movie settings table
	LanguageCode    string `db:"language_code"`     // ISO 639-1 language code
}

type BlacklistedGenres struct {
	ID              int    `db:"id"`                // Primary key with auto-increment
	MovieSettingsID int    `db:"movie_settings_id"` // Foreign key to the movie settings table
	Genre           string `db:"genre"`             // Genre to blacklist
}

type BlacklistedTitleKeywords struct {
	ID              int    `db:"id"`                // Primary key with auto-increment
	MovieSettingsID int    `db:"movie_settings_id"` // Foreign key to the movie settings table
	Keyword         string `db:"keyword"`           // Keyword to blacklist from the title of a movie
}

type BlacklistedTMDBIDs struct {
	ID              int `db:"id"`                // Primary key with auto-increment
	MovieSettingsID int `db:"movie_settings_id"` // Foreign key to the movie settings table
	TMDBID          int `db:"tmdb_id"`           // TMDb ID to blacklist
}

func (q *Queries) GetMovieSettings(ctx context.Context) (MovieSettings, error) {
	var settings MovieSettings

	// Begin a transaction
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return settings, err
	}
	defer tx.Rollback()

	// Query for the main movie settings
	err = tx.QueryRowContext(ctx, `
		SELECT id, interval, anticipated, box_office, popular, trending, 
		       max_runtime, min_runtime, min_year, max_year, rotten_tomatoes
		FROM movie_settings
		LIMIT 1;
	`).Scan(
		&settings.ID,
		&settings.Interval,
		&settings.Anticipated,
		&settings.BoxOffice,
		&settings.Popular,
		&settings.Trending,
		&settings.MaxRuntime,
		&settings.MinRuntime,
		&settings.MinYear,
		&settings.MaxYear,
		&settings.RottenTomatoes,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return settings, fmt.Errorf("no movie settings found")
		}
		return settings, err
	}

	// Fetch allowed countries
	countryQuery := `
		SELECT id, movie_settings_id, country_code
		FROM movie_allowed_countries
		WHERE movie_settings_id = ?;
	`
	countryRows, err := tx.QueryContext(ctx, countryQuery, settings.ID)
	if err != nil {
		return settings, err
	}
	defer countryRows.Close()

	for countryRows.Next() {
		var allowedCountry MovieAllowedCountries
		if err := countryRows.Scan(&allowedCountry.ID, &allowedCountry.MovieSettingsID, &allowedCountry.CountryCode); err != nil {
			return settings, err
		}
		settings.AllowedCountries = append(settings.AllowedCountries, allowedCountry)
	}

	// Fetch allowed languages
	languageQuery := `
		SELECT id, movie_settings_id, language_code
		FROM movie_allowed_languages
		WHERE movie_settings_id = ?;
	`
	languageRows, err := tx.QueryContext(ctx, languageQuery, settings.ID)
	if err != nil {
		return settings, err
	}
	defer languageRows.Close()

	for languageRows.Next() {
		var allowedLanguage MovieAllowedLanguages
		if err := languageRows.Scan(&allowedLanguage.ID, &allowedLanguage.MovieSettingsID, &allowedLanguage.LanguageCode); err != nil {
			return settings, err
		}
		settings.AllowedLanguages = append(settings.AllowedLanguages, allowedLanguage)
	}

	// Fetch blacklisted genres
	genreQuery := `
		SELECT id, movie_settings_id, genre
		FROM movie_blacklisted_genres
		WHERE movie_settings_id = ?;
	`
	genreRows, err := tx.QueryContext(ctx, genreQuery, settings.ID)
	if err != nil {
		return settings, err
	}
	defer genreRows.Close()

	for genreRows.Next() {
		var blacklistedGenre BlacklistedGenres
		if err := genreRows.Scan(&blacklistedGenre.ID, &blacklistedGenre.MovieSettingsID, &blacklistedGenre.Genre); err != nil {
			return settings, err
		}
		settings.BlacklistedGenres = append(settings.BlacklistedGenres, blacklistedGenre)
	}

	// Fetch blacklisted title keywords
	keywordQuery := `
		SELECT id, movie_settings_id, keyword
		FROM movie_blacklisted_title_keywords
		WHERE movie_settings_id = ?;
	`
	keywordRows, err := tx.QueryContext(ctx, keywordQuery, settings.ID)
	if err != nil {
		return settings, err
	}
	defer keywordRows.Close()

	for keywordRows.Next() {
		var blacklistedKeyword BlacklistedTitleKeywords
		if err := keywordRows.Scan(&blacklistedKeyword.ID, &blacklistedKeyword.MovieSettingsID, &blacklistedKeyword.Keyword); err != nil {
			return settings, err
		}
		settings.BlacklistedTitleKeywords = append(settings.BlacklistedTitleKeywords, blacklistedKeyword)
	}

	// Fetch blacklisted TMDb IDs
	tmdbQuery := `
		SELECT id, movie_settings_id, tmdb_id
		FROM movie_blacklisted_tmdb_ids
		WHERE movie_settings_id = ?;
	`
	tmdbRows, err := tx.QueryContext(ctx, tmdbQuery, settings.ID)
	if err != nil {
		return settings, err
	}
	defer tmdbRows.Close()

	for tmdbRows.Next() {
		var blacklistedTMDBID BlacklistedTMDBIDs
		if err := tmdbRows.Scan(&blacklistedTMDBID.ID, &blacklistedTMDBID.MovieSettingsID, &blacklistedTMDBID.TMDBID); err != nil {
			return settings, err
		}
		settings.BlacklistedTMDBIDs = append(settings.BlacklistedTMDBIDs, blacklistedTMDBID)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return settings, err
	}

	return settings, nil
}

func (q *Queries) GetMovieInterval(ctx context.Context) (sql.NullInt32, error) {
	var interval sql.NullInt32

	err := q.db.QueryRowContext(ctx, `
		SELECT interval
		FROM movie_settings
		LIMIT 1;
	`).Scan(&interval)
	if err != nil {
		if err == sql.ErrNoRows {
			return interval, fmt.Errorf("no movie interval found")
		}
		return interval, err
	}

	return interval, nil
}
