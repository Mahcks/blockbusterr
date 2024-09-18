CREATE TABLE `trakt` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `client_id` TEXT NOT NULL,
    `client_secret` TEXT NOT NULL
);

CREATE TABLE `settings` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `key` TEXT UNIQUE NOT NULL,
    `value` TEXT NOT NULL,
    `type` TEXT NOT NULL DEFAULT 'text',
    `updated_at` DATETIME DEFAULT 'CURRENT_TIMESTAMP'
);

INSERT INTO
    settings (key, value, type)
VALUES
    ('SETUP_COMPLETE', 'false', 'boolean');

INSERT INTO
    settings (key, value, type)
VALUES
    ('MODE', 'radarr-sonarr', 'text');

CREATE TABLE ombi (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    -- Primary key with auto-increment
    `api_key` TEXT,
    -- API key required to make requests to Ombi (nullable)
    `url` TEXT,
    -- Base URL for the Ombi server (nullable)
    `user_id` TEXT,
    -- User ID to use for Ombi (nullable)
    `language` TEXT,
    -- Language to use for getting movies from Ombi (nullable)
    `movie_quality` INTEGER,
    -- Quality profile ID to use for Ombi (nullable)
    `movie_root_folder` INTEGER,
    -- The root folder ID to use for Ombi (nullable)
    `show_quality` INTEGER,
    -- Quality profile ID to use for Ombi (nullable)
    `show_root_folder` INTEGER -- The root folder ID to use for Ombi (nullable)
);

CREATE TABLE radarr (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    -- Primary key with auto-increment
    `api_key` TEXT,
    -- API key required to make requests to Radarr (nullable)
    `url` TEXT,
    -- Base URL for the Radarr server (nullable)
    `minimum_availability` TEXT,
    -- Minimum availability setting (nullable)
    `quality` INTEGER,
    -- Quality profile ID to use for Radarr (nullable)
    `root_folder` INTEGER -- The root folder ID to use for Radarr (nullable)
);

INSERT INTO
    radarr (
        api_key,
        url,
        minimum_availability,
        quality,
        root_folder
    )
VALUES
    (
        null,
        null,
        'announced',
        null,
        null
    );

CREATE TABLE movie_settings (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Primary key with auto-increment
    `anticipated` INTEGER,
    -- How many movies will be pulled from the anticipated list
    `cron_job_anticipated` TEXT,
    -- Cron job expression for the anticipated list
    `box_office` INTEGER,
    -- How many movies will be pulled from the box office list
    `cron_job_box_office` TEXT,
    -- Cron job expression for the box office list
    `popular` INTEGER,
    -- How many movies will be pulled from the popular list
    `cron_job_popular` TEXT,
    -- Cron job expression for the popular list
    `trending` INTEGER,
    -- How many movies will be pulled from the trending list
    `cron_job_trending` TEXT,
    -- Cron job expression for the trending list
    `max_runtime` INTEGER,
    -- Blacklist movies with runtime longer than the specified time (in minutes)
    `min_runtime` INTEGER,
    -- Blacklist movies with runtime shorter than the specified time (in minutes)
    `min_year` INTEGER,
    -- Blacklist movies released before the specified year. If left empty/is zero, it'll ignore the year.
    `max_year` INTEGER,
    -- Blacklist movies released after the specified year. If left empty, it'll be the current year
    `rotten_tomatoes` TEXT -- Rotten Tomatoes rating filter for movies
);

INSERT INTO
    movie_settings (
        anticipated,
        cron_job_anticipated,
        box_office,
        cron_job_box_office,
        popular,
        cron_job_popular,
        trending,
        cron_job_trending,
        max_runtime,
        min_runtime,
        min_year,
        max_year,
        rotten_tomatoes
    )
VALUES
    (
        10,
        '0 0 * * 1',
        10,
        '0 0 * * *',
        5,
        '0 0 * * *',
        5,
        '0 0 * * *',
        180,
        30,
        0,
        0,
        ''
    );

CREATE TABLE movie_allowed_countries (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Primary key with auto-increment
    `movie_settings_id` INTEGER NOT NULL,
    -- Reference to the movie settings
    `country_code` TEXT NOT NULL,
    -- Country code for allowed countries (e.g., 'us', 'gb')
    FOREIGN KEY (`movie_settings_id`) REFERENCES movie_settings(`id`) -- Foreign key constraint
);

INSERT INTO
    movie_allowed_countries (movie_settings_id, country_code)
VALUES
    (1, 'us'),
    (1, 'gb');

CREATE TABLE movie_allowed_languages (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Primary key with auto-increment
    `movie_settings_id` INTEGER NOT NULL,
    -- Reference to the movie settings
    `language_code` TEXT NOT NULL,
    -- Language code for allowed languages (e.g., 'en', 'es')
    FOREIGN KEY (`movie_settings_id`) REFERENCES movie_settings(`id`) -- Foreign key constraint
);

CREATE TABLE movie_blacklisted_genres (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Primary key with auto-increment
    `movie_settings_id` INTEGER NOT NULL,
    -- Reference to the movie settings
    `genre` TEXT NOT NULL,
    -- Genre to be blacklisted (e.g., 'anime', 'disaster')
    FOREIGN KEY (`movie_settings_id`) REFERENCES movie_settings(`id`) -- Foreign key constraint
);

CREATE TABLE movie_blacklisted_title_keywords (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Primary key with auto-increment
    `movie_settings_id` INTEGER NOT NULL,
    -- Reference to the movie settings
    `keyword` TEXT NOT NULL,
    -- Keyword in movie title to be blacklisted (e.g., 'Barbie', 'Untitled')
    FOREIGN KEY (`movie_settings_id`) REFERENCES movie_settings(`id`) -- Foreign key constraint
);

CREATE TABLE movie_blacklisted_tmdb_ids (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Primary key with auto-increment
    `movie_settings_id` INTEGER NOT NULL,
    -- Reference to the movie settings
    `tmdb_id` INTEGER NOT NULL,
    -- TMDb ID of movies to be blacklisted
    FOREIGN KEY (`movie_settings_id`) REFERENCES movie_settings(`id`) -- Foreign key constraint
);

CREATE TABLE sonarr (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
    -- Primary key with auto-increment
    `api_key` TEXT,
    -- API key required to make requests to Sonarr (nullable)
    `url` TEXT,
    -- Base URL for the Sonarr server (nullable)
    `language` TEXT,
    -- Language to use for getting shows from Sonarr (nullable)
    `quality` INTEGER,
    -- Quality profile ID to use for Sonarr (nullable)
    `root_folder` INTEGER -- The root folder ID to use for Sonarr (nullable)
    `season_folder` INTEGER -- Season folder setting for Sonarr (nullable)
);

-- Table for general show settings
CREATE TABLE show_settings (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Primary key with auto-increment
    `anticipated` INTEGER,
    -- How many shows after every interval will grab from the anticipated list
    `cron_job_anticipated` TEXT,
    -- Cron job expression for the anticipated list
    `popular` INTEGER,
    -- How many shows after every interval will grab from the popular list
    `cron_job_popular` TEXT,
    -- Cron job expression for the popular list
    `trending` INTEGER,
    -- How many shows after every interval will grab from the trending list
    `cron_job_trending` TEXT,
    -- Cron job expression for the trending list
    `max_runtime` INTEGER,
    -- Blacklisted shows with runtime longer than the specified time (in minutes)
    `min_runtime` INTEGER,
    -- Blacklisted shows with runtime shorter than the specified time (in minutes)
    `min_year` INTEGER,
    -- Blacklist shows released before the specified year
    `max_year` INTEGER -- Blacklist shows released after the specified year
);

INSERT INTO
    show_settings (
        anticipated,
        cron_job_anticipated,
        popular,
        cron_job_popular,
        trending,
        cron_job_trending,
        max_runtime,
        min_runtime,
        min_year,
        max_year
    )
VALUES
    (
        10,
        '0 0 * * 1',
        10,
        '0 0 * * *',
        5,
        '0 0 * * *',
        180,
        30,
        0,
        0
    );

-- Table for allowed countries in shows settings
CREATE TABLE show_allowed_countries (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Primary key with auto-increment
    `show_settings_id` INTEGER NOT NULL,
    -- Reference to the show settings
    `country_code` TEXT NOT NULL,
    -- Country code for allowed countries (e.g., 'us', 'gb')
    FOREIGN KEY (`show_settings_id`) REFERENCES show_settings(`id`) -- Foreign key constraint
);

INSERT INTO
    show_allowed_countries (show_settings_id, country_code)
VALUES
    (1, 'us'),
    (1, 'gb'),
    (1, 'gb');

-- Table for allowed languages in shows settings
CREATE TABLE show_allowed_languages (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Primary key with auto-increment
    `show_settings_id` INTEGER NOT NULL,
    -- Reference to the show settings
    `language_code` TEXT NOT NULL,
    -- Language code for allowed languages (e.g., 'en', 'es')
    FOREIGN KEY (`show_settings_id`) REFERENCES show_settings(`id`) -- Foreign key constraint
);

INSERT INTO
    show_allowed_languages (show_settings_id, language_code)
VALUES
    (1, 'en');

-- Table for blacklisted genres in shows settings
CREATE TABLE show_blacklisted_genres (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Primary key with auto-increment
    `show_settings_id` INTEGER NOT NULL,
    -- Reference to the show settings
    `genre` TEXT NOT NULL,
    -- Genre to be blacklisted (e.g., 'animation', 'reality')
    FOREIGN KEY (`show_settings_id`) REFERENCES show_settings(`id`) -- Foreign key constraint
);

INSERT INTO
    show_blacklisted_genres (show_settings_id, genre)
VALUES
    (1, 'game-show'),
    (1, 'home-and-garden'),
    (1, 'children'),
    (1, 'anime'),
    (1, 'news'),
    (1, 'documentary'),
    (1, 'special-interest');

-- Table for blacklisted networks in shows settings
CREATE TABLE show_blacklisted_networks (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Primary key with auto-increment
    `show_settings_id` INTEGER NOT NULL,
    -- Reference to the show settings
    `network` TEXT NOT NULL,
    -- Network to be blacklisted (e.g., 'twitch', 'youtube')
    FOREIGN KEY (`show_settings_id`) REFERENCES show_settings(`id`) -- Foreign key constraint
);

INSERT INTO
    show_blacklisted_networks (show_settings_id, network)
VALUES
    (1, 'fox sports'),
    (1, 'yahoo!'),
    (1, 'espn'),
    (1, 'cartoon network'),
    (1, 'teletoon'),
    (1, 'the movie network'),
    (1, 'cbbc'),
    (1, 'cnn'),
    (1, 'reelzchannel'),
    (1, 'hallmark'),
    (1, 'nickelodeon'),
    (1, 'twitch'),
    (1, 'youtube');

-- Table for blacklisted title keywords in shows settings
CREATE TABLE show_blacklisted_title_keywords (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Primary key with auto-increment
    `show_settings_id` INTEGER NOT NULL,
    -- Reference to the show settings
    `keyword` TEXT NOT NULL,
    -- Keyword in show title to be blacklisted (e.g., 'game', 'talk')
    FOREIGN KEY (`show_settings_id`) REFERENCES show_settings(`id`) -- Foreign key constraint
);

-- Table for blacklisted TVDB IDs in shows settings
CREATE TABLE show_blacklisted_tvdb_ids (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    -- Primary key with auto-increment
    `show_settings_id` INTEGER NOT NULL,
    -- Reference to the show settings
    `tvdb_id` INTEGER NOT NULL,
    -- TVDB ID of shows to be blacklisted
    FOREIGN KEY (`show_settings_id`) REFERENCES show_settings(`id`) -- Foreign key constraint
);

CREATE TABLE IF NOT EXISTS notifications_config (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `platform` TEXT NOT NULL UNIQUE,
    `enabled` BOOLEAN NOT NULL,
    `webhook_url` TEXT NOT NULL
);

INSERT INTO
    notifications_config (`platform`, `enabled`, `webhook_url`)
VALUES
    ('discord', 0, '');

CREATE TABLE IF NOT EXISTS omdb (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `api_key` TEXT
);

-- Table to keep track of recently added media
CREATE TABLE IF NOT EXISTS recently_added (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `type` TEXT CHECK(type IN ('MOVIE', 'SHOW')),
    `title` TEXT NOT NULL,
    `year` INTEGER NOT NULL,
    `summary` TEXT NOT NULL,
    `imdb_id` TEXT NOT NULL UNIQUE,
    `poster` TEXT NOT NULL,
    `added_at` DATETIME DEFAULT CURRENT_TIMESTAMP
);