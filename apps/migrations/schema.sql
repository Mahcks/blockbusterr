CREATE TABLE `settings` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `key` TEXT UNIQUE NOT NULL,
  `value` TEXT NOT NULL,
  `type` TEXT NOT NULL DEFAULT 'text',
  `updated_at` DATETIME DEFAULT 'CURRENT_TIMESTAMP'
);

INSERT INTO settings (key, value, type) VALUES ('SETUP_COMPLETE', 'false', 'boolean');

CREATE TABLE radarr (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,  -- Primary key with auto-increment
    `api_key` TEXT,                                   -- API key required to make requests to Radarr (nullable)
    `url` TEXT,                                       -- Base URL for the Radarr server (nullable)
    `minimum_availability` TEXT,                      -- Minimum availability setting (nullable)
    `quality` INTEGER,                                -- Quality profile ID to use for Radarr (nullable)
    `root_folder` INTEGER                                -- The root folder ID to use for Radarr (nullable)
);

CREATE TABLE movie_settings (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,                            -- Primary key with auto-increment
    `interval` INTEGER,                                       				 -- The rate at which movies are pulled from movie databases like Trakt
    `anticipated` INTEGER,                                    				 -- How many movies after every interval will grab from the anticipated list
    `box_office` INTEGER,                                     				 -- How many movies after every interval will grab from the box office list
    `popular` INTEGER,                                        				 -- How many movies after every interval will grab from the popular list
    `trending` INTEGER,                                       				 -- How many movies after every interval will grab from the trending list
    `max_runtime` INTEGER,                                 						 -- Blacklist movies with runtime longer than the specified time (in minutes)
    `min_runtime` INTEGER,                                 						 -- Blacklist movies with runtime shorter than the specified time (in minutes)
    `min_year` INTEGER,                                    					   -- Blacklist movies released before the specified year. If left empty/is zero, it'll ignore the year.
    `max_year` INTEGER,                                    						 -- Blacklist movies released after the specified year. If left empty, it'll be the current year
    `rotten_tomatoes` TEXT                                             -- Rotten Tomatoes rating filter for movies
);

CREATE TABLE movie_allowed_countries (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,                   			 -- Primary key with auto-increment
    `movie_settings_id` INTEGER NOT NULL,                     			 -- Reference to the movie settings
    `country_code` TEXT NOT NULL,                             			 -- Country code for allowed countries (e.g., 'us', 'gb')
    FOREIGN KEY (`movie_settings_id`) REFERENCES movie_settings(`id`)  -- Foreign key constraint
);

CREATE TABLE movie_allowed_languages (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,                   			 -- Primary key with auto-increment
    `movie_settings_id` INTEGER NOT NULL,                     			 -- Reference to the movie settings
    `language_code` TEXT NOT NULL,                            		   -- Language code for allowed languages (e.g., 'en', 'es')
    FOREIGN KEY (`movie_settings_id`) REFERENCES movie_settings(`id`)  -- Foreign key constraint
);

CREATE TABLE movie_blacklisted_genres (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,                   			 -- Primary key with auto-increment
    `movie_settings_id` INTEGER NOT NULL,                     			 -- Reference to the movie settings
    `genre` TEXT NOT NULL,                                    			 -- Genre to be blacklisted (e.g., 'anime', 'disaster')
    FOREIGN KEY (`movie_settings_id`) REFERENCES movie_settings(`id`)  -- Foreign key constraint
);

CREATE TABLE movie_blacklisted_title_keywords (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,                   			 -- Primary key with auto-increment
    `movie_settings_id` INTEGER NOT NULL,                     			 -- Reference to the movie settings
    `keyword` TEXT NOT NULL,                                  			 -- Keyword in movie title to be blacklisted (e.g., 'Barbie', 'Untitled')
    FOREIGN KEY (`movie_settings_id`) REFERENCES movie_settings(`id`)  -- Foreign key constraint
);

CREATE TABLE movie_blacklisted_tmdb_ids (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,                   			 -- Primary key with auto-increment
    `movie_settings_id` INTEGER NOT NULL,                     			 -- Reference to the movie settings
    `tmdb_id` INTEGER NOT NULL,                               			 -- TMDb ID of movies to be blacklisted
    FOREIGN KEY (`movie_settings_id`) REFERENCES movie_settings(`id`)  -- Foreign key constraint
);