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