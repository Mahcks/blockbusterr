export interface MovieSettings {
    id: number;
    interval?: number; // The rate at which movies are pulled from movie databases like Trakt (in hours)
    anticipated?: number; // How many movies after every interval will grab from the anticipated list
    box_office?: number; // How many movies after every interval will grab from the box office list
    popular?: number; // How many movies after every interval will grab from the popular list
    trending?: number; // How many movies after every interval will grab from the trending list
    max_runtime?: number; // Blacklist movies with runtime longer than the specified time (in minutes)
    min_runtime?: number; // Blacklist movies with runtime shorter than the specified time (in minutes)
    min_year?: number; // Blacklist movies released before the specified year. If empty, ignore the year.
    max_year?: number; // Blacklist movies released after the specified year. If empty, use the current year.
    rotten_tomatoes?: string; // Rotten Tomatoes rating filter for movies
    allowed_countries: MovieAllowedCountry[]; // List of allowed countries
    allowed_languages: MovieAllowedLanguage[]; // List of allowed languages
    blacklisted_genres: BlacklistedGenre[]; // List of blacklisted genres
    blacklisted_title_keywords: BlacklistedTitleKeyword[]; // List of blacklisted title keywords
    blacklisted_tmdb_ids: BlacklistedTMDBID[]; // List of blacklisted TMDb IDs
}

export interface MovieAllowedCountry {
    id: number; // Primary key with auto-increment
    country_code: string; // ISO 3166-1 alpha-2 country code
}

export interface MovieAllowedLanguage {
    id: number; // Primary key with auto-increment
    language_code: string; // ISO 639-1 language code
}

export interface BlacklistedGenre {
    id: number; // Primary key with auto-increment
    genre: string; // Genre to blacklist
}

export interface BlacklistedTitleKeyword {
    id: number; // Primary key with auto-increment
    keyword: string; // Keyword to blacklist from the title of a movie
}

export interface BlacklistedTMDBID {
    id: number; // Primary key with auto-increment
    tmdb_id: number; // TMDb ID to blacklist
}