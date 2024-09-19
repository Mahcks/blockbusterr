#!/bin/sh

DB_PATH="/app/data/settings.db"
MIGRATIONS_PATH="/migrations/schema.sql"
LOG_FILE="/app/data/init.log"

# Check if the log file exists to avoid repeated initialization
if [ ! -f "$DB_PATH" ]; then
    echo "$(date): SQLite database not found. Initializing..." | tee -a "$LOG_FILE"
    sqlite3 "$DB_PATH" < "$MIGRATIONS_PATH"
    echo "$(date): Database initialization complete." | tee -a "$LOG_FILE"
else
    echo "$(date): SQLite database already exists. Skipping initialization." | tee -a "$LOG_FILE"
fi

# Continue with the main process
exec "$@"
