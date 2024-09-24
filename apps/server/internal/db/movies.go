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

	// Cron fields for individual list scheduling
	CronAnticipated sql.NullString `db:"cron_job_anticipated"` // Cron expression for the anticipated list
	CronBoxOffice   sql.NullString `db:"cron_job_box_office"`  // Cron expression for the box office list
	CronPopular     sql.NullString `db:"cron_job_popular"`     // Cron expression for the popular list
	CronTrending    sql.NullString `db:"cron_job_trending"`    // Cron expression for the trending list

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

	// Query for the main movie settings including the cron fields
	err = tx.QueryRowContext(ctx, `
		SELECT id, anticipated, box_office, popular, trending,
		       max_runtime, min_runtime, min_year, max_year, rotten_tomatoes,
		       cron_job_anticipated, cron_job_box_office, cron_job_popular, cron_job_trending
		FROM movie_settings
		LIMIT 1;
	`).Scan(
		&settings.ID,
		&settings.Anticipated,
		&settings.BoxOffice,
		&settings.Popular,
		&settings.Trending,
		&settings.MaxRuntime,
		&settings.MinRuntime,
		&settings.MinYear,
		&settings.MaxYear,
		&settings.RottenTomatoes,
		&settings.CronAnticipated,
		&settings.CronBoxOffice,
		&settings.CronPopular,
		&settings.CronTrending,
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

func (q *Queries) UpdateMovieSettings(ctx context.Context, anticipated, boxOffice, popular, trending, maxRuntime, minRuntime, minYear, maxYear sql.NullInt32, rottenTomatoes sql.NullString, cronAnticipated, cronBoxOffice, cronPopular, cronTrending sql.NullString) error {
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE movie_settings
		SET anticipated = $1, box_office = $2, popular = $3, trending = $4,
		    max_runtime = $5, min_runtime = $6, min_year = $7, max_year = $8, rotten_tomatoes = $9,
		    cron_job_anticipated = $10, cron_job_box_office = $11, cron_job_popular = $12, cron_job_trending = $13
		WHERE id = 1;
	`, anticipated, boxOffice, popular, trending, maxRuntime, minRuntime, minYear, maxYear, rottenTomatoes, cronAnticipated, cronBoxOffice, cronPopular, cronTrending)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
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
