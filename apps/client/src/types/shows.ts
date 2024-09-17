export interface ShowSettings {
    id: number;
    anticipated?: number; // How many shows after every interval will grab from the anticipated list
    cron_job_anticipated?: string; // Cron job for the anticipated list
    popular?: number; // How many shows after every interval will grab from the popular list
    cron_job_popular?: string; // Cron job for the popular list
    trending?: number; // How many shows after every interval will grab from the trending list
    cron_job_trending?: string; // Cron job for the trending list
    max_runtime?: number; // Blacklist shows with runtime longer than the specified time (in minutes)
    min_runtime?: number; // Blacklist shows with runtime shorter than the specified time (in minutes)
    min_year?: number; // Blacklist shows released before the specified year. If empty, ignore the year.
    max_year?: number; // Blacklist shows released after the specified year. If empty, use the current year.
    allowed_countries: ShowAllowedCountry[]; // List of allowed countries
    allowed_languages: ShowAllowedLanguage[]; // List of allowed languages
    blacklisted_genres: BlacklistedShowGenre[]; // List of blacklisted genres
    blacklisted_networks: BlacklistedNetwork[]; // List of blacklisted networks
    blacklisted_title_keywords: BlacklistedShowTitleKeyword[]; // List of blacklisted title keywords
    blacklisted_tvdb_ids: BlacklistedTVDBID[]; // List of blacklisted TVDB IDs
}

export interface ShowAllowedCountry {
    id: number; // Primary key with auto-increment
    country_code: string; // ISO 3166-1 alpha-2 country code
}

export interface ShowAllowedLanguage {
    id: number; // Primary key with auto-increment
    language_code: string; // ISO 639-1 language code
}

export interface BlacklistedShowGenre {
    id: number; // Primary key with auto-increment
    genre: string; // Genre to blacklist
}

export interface BlacklistedNetwork {
    id: number; // Primary key with auto-increment
    network: string; // Network to blacklist (e.g., 'Netflix', 'HBO')
}

export interface BlacklistedShowTitleKeyword {
    id: number; // Primary key with auto-increment
    keyword: string; // Keyword to blacklist from the title of a show
}

export interface BlacklistedTVDBID {
    id: number; // Primary key with auto-increment
    tvdb_id: number; // TVDB ID to blacklist
}