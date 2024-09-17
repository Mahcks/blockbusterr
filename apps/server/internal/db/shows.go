package db

import (
	"context"
	"database/sql"
	"fmt"
)

type ShowSettings struct {
	ID          int           `db:"id"`          // Primary key with auto-increment
	Anticipated sql.NullInt32 `db:"anticipated"` // How many shows after every interval will grab from the anticipated list
	Popular     sql.NullInt32 `db:"popular"`     // How many shows after every interval will grab from the popular list
	Trending    sql.NullInt32 `db:"trending"`    // How many shows after every interval will grab from the trending list
	MaxRuntime  sql.NullInt32 `db:"max_runtime"` // Blacklisted shows with runtime longer than the specified time (in minutes)
	MinRuntime  sql.NullInt32 `db:"min_runtime"` // Blacklisted shows with runtime shorter than the specified time (in minutes)
	MinYear     sql.NullInt32 `db:"min_year"`    // Blacklist shows released before the specified year
	MaxYear     sql.NullInt32 `db:"max_year"`    // Blacklist shows released after the specified year

	CronJobAnticipated sql.NullString `db:"cron_job_anticipated"` // Cron expression for the anticipated list
	CronJobPopular     sql.NullString `db:"cron_job_popular"`     // Cron expression for the popular list
	CronJobTrending    sql.NullString `db:"cron_job_trending"`    // Cron expression for the trending list

	AllowedCountries         []ShowAllowedCountries         // List of allowed countries
	AllowedLanguages         []ShowAllowedLanguages         // List of allowed languages
	BlacklistedGenres        []ShowBlacklistedGenres        // List of blacklisted genres
	BlacklistedNetworks      []ShowBlacklistedNetworks      // List of blacklisted networks
	BlacklistedTitleKeywords []ShowBlacklistedTitleKeywords // List of blacklisted title keywords
	BlacklistedTVDBIDs       []ShowBlacklistedTVDBIDs       // List of blacklisted TVDB IDs
}

type ShowAllowedCountries struct {
	ID             int    `db:"id"`               // Primary key with auto-increment
	ShowSettingsID int    `db:"show_settings_id"` // Foreign key to the show settings table
	CountryCode    string `db:"country_code"`     // ISO 3166-1 alpha-2 country code
}

type ShowAllowedLanguages struct {
	ID             int    `db:"id"`               // Primary key with auto-increment
	ShowSettingsID int    `db:"show_settings_id"` // Foreign key to the show settings table
	LanguageCode   string `db:"language_code"`    // ISO 639-1 language code
}

type ShowBlacklistedGenres struct {
	ID             int    `db:"id"`               // Primary key with auto-increment
	ShowSettingsID int    `db:"show_settings_id"` // Foreign key to the show settings table
	Genre          string `db:"genre"`            // Genre to blacklist
}

type ShowBlacklistedNetworks struct {
	ID             int    `db:"id"`               // Primary key with auto-increment
	ShowSettingsID int    `db:"show_settings_id"` // Foreign key to the show settings table
	Network        string `db:"network"`          // Network to blacklist
}

type ShowBlacklistedTitleKeywords struct {
	ID             int    `db:"id"`               // Primary key with auto-increment
	ShowSettingsID int    `db:"show_settings_id"` // Foreign key to the show settings table
	Keyword        string `db:"keyword"`          // Keyword to blacklist from the title of a show
}

type ShowBlacklistedTVDBIDs struct {
	ID             int `db:"id"`               // Primary key with auto-increment
	ShowSettingsID int `db:"show_settings_id"` // Foreign key to the show settings table
	TVDBID         int `db:"tvdb_id"`          // TVDB ID to blacklist
}

func (q *Queries) GetShowSettings(ctx context.Context) (ShowSettings, error) {
	var settings ShowSettings

	// Begin a transaction
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return settings, err
	}
	defer tx.Rollback()

	// Query for the main show settings
	err = tx.QueryRowContext(ctx, `
		SELECT id, anticipated, cron_job_anticipated, popular, cron_job_popular, trending, cron_job_trending, 
		       max_runtime, min_runtime, min_year, max_year
		FROM show_settings
		LIMIT 1;
	`).Scan(
		&settings.ID,
		&settings.Anticipated,
		&settings.CronJobAnticipated,
		&settings.Popular,
		&settings.CronJobPopular,
		&settings.Trending,
		&settings.CronJobTrending,
		&settings.MaxRuntime,
		&settings.MinRuntime,
		&settings.MinYear,
		&settings.MaxYear,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return settings, fmt.Errorf("no show settings found")
		}
		return settings, err
	}

	// Fetch allowed countries
	countryQuery := `
		SELECT id, show_settings_id, country_code
		FROM show_allowed_countries
		WHERE show_settings_id = ?;
	`
	countryRows, err := tx.QueryContext(ctx, countryQuery, settings.ID)
	if err != nil {
		return settings, err
	}
	defer countryRows.Close()

	for countryRows.Next() {
		var allowedCountry ShowAllowedCountries
		if err := countryRows.Scan(&allowedCountry.ID, &allowedCountry.ShowSettingsID, &allowedCountry.CountryCode); err != nil {
			return settings, err
		}
		settings.AllowedCountries = append(settings.AllowedCountries, allowedCountry)
	}

	// Fetch allowed languages
	languageQuery := `
		SELECT id, show_settings_id, language_code
		FROM show_allowed_languages
		WHERE show_settings_id = ?;
	`
	languageRows, err := tx.QueryContext(ctx, languageQuery, settings.ID)
	if err != nil {
		return settings, err
	}
	defer languageRows.Close()

	for languageRows.Next() {
		var allowedLanguage ShowAllowedLanguages
		if err := languageRows.Scan(&allowedLanguage.ID, &allowedLanguage.ShowSettingsID, &allowedLanguage.LanguageCode); err != nil {
			return settings, err
		}
		settings.AllowedLanguages = append(settings.AllowedLanguages, allowedLanguage)
	}

	// Fetch blacklisted genres
	genreQuery := `
		SELECT id, show_settings_id, genre
		FROM show_blacklisted_genres
		WHERE show_settings_id = ?;
	`
	genreRows, err := tx.QueryContext(ctx, genreQuery, settings.ID)
	if err != nil {
		return settings, err
	}
	defer genreRows.Close()

	for genreRows.Next() {
		var blacklistedGenre ShowBlacklistedGenres
		if err := genreRows.Scan(&blacklistedGenre.ID, &blacklistedGenre.ShowSettingsID, &blacklistedGenre.Genre); err != nil {
			return settings, err
		}
		settings.BlacklistedGenres = append(settings.BlacklistedGenres, blacklistedGenre)
	}

	// Fetch blacklisted networks
	networkQuery := `
		SELECT id, show_settings_id, network
		FROM show_blacklisted_networks
		WHERE show_settings_id = ?;
	`
	networkRows, err := tx.QueryContext(ctx, networkQuery, settings.ID)
	if err != nil {
		return settings, err
	}
	defer networkRows.Close()

	for networkRows.Next() {
		var blacklistedNetwork ShowBlacklistedNetworks
		if err := networkRows.Scan(&blacklistedNetwork.ID, &blacklistedNetwork.ShowSettingsID, &blacklistedNetwork.Network); err != nil {
			return settings, err
		}
		settings.BlacklistedNetworks = append(settings.BlacklistedNetworks, blacklistedNetwork)
	}

	// Fetch blacklisted title keywords
	keywordQuery := `
		SELECT id, show_settings_id, keyword
		FROM show_blacklisted_title_keywords
		WHERE show_settings_id = ?;
	`
	keywordRows, err := tx.QueryContext(ctx, keywordQuery, settings.ID)
	if err != nil {
		return settings, err
	}
	defer keywordRows.Close()

	for keywordRows.Next() {
		var blacklistedKeyword ShowBlacklistedTitleKeywords
		if err := keywordRows.Scan(&blacklistedKeyword.ID, &blacklistedKeyword.ShowSettingsID, &blacklistedKeyword.Keyword); err != nil {
			return settings, err
		}
		settings.BlacklistedTitleKeywords = append(settings.BlacklistedTitleKeywords, blacklistedKeyword)
	}

	// Fetch blacklisted TVDB IDs
	tvdbQuery := `
		SELECT id, show_settings_id, tvdb_id
		FROM show_blacklisted_tvdb_ids
		WHERE show_settings_id = ?;
	`
	tvdbRows, err := tx.QueryContext(ctx, tvdbQuery, settings.ID)
	if err != nil {
		return settings, err
	}
	defer tvdbRows.Close()

	for tvdbRows.Next() {
		var blacklistedTVDBID ShowBlacklistedTVDBIDs
		if err := tvdbRows.Scan(&blacklistedTVDBID.ID, &blacklistedTVDBID.ShowSettingsID, &blacklistedTVDBID.TVDBID); err != nil {
			return settings, err
		}
		settings.BlacklistedTVDBIDs = append(settings.BlacklistedTVDBIDs, blacklistedTVDBID)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return settings, err
	}

	return settings, nil
}
