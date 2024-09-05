#!/bin/sh

# If the database file doesn't exist, initialize it
if [ ! -f "/app/data/settings.db" ]; then
    echo "Initializing SQLite database..."
    sqlite3 /app/data/settings.db < /migrations/schema.sql
else
    echo "SQLite database already exists. Skipping initialization."
fi

# Continue with the main process
exec "$@"