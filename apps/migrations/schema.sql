CREATE TABLE `settings` (
  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
  `key` TEXT UNIQUE NOT NULL,
  `value` TEXT NOT NULL,
  `type` TEXT NOT NULL DEFAULT 'text',
  `updated_at` DATETIME DEFAULT 'CURRENT_TIMESTAMP'
);

INSERT INTO settings (key, value, type) VALUES ('SETUP_COMPLETE', 'false', 'boolean');