---
title: Configuration
description: Complete configuration reference for Blockbusterr
---

Blockbusterr can be configured via a YAML configuration file or through the Web UI.

## Configuration File

The configuration file is located at `config/config.yaml`. Here's a complete example:

```yaml
server:
  host: "0.0.0.0"
  port: 9090

database:
  path: "data/blockbusterr.db"

# Trakt Integration (Required)
trakt:
  client_id: "your_trakt_client_id"
  client_secret: "your_trakt_client_secret"
  access_token: ""  # Auto-populated after OAuth
  refresh_token: ""  # Auto-populated after OAuth

# TMDB Integration (Optional - for poster images)
tmdb:
  api_key: "your_tmdb_api_key"

# Radarr Integration (Optional)
radarr:
  enabled: true
  url: "http://radarr:7878"
  api_key: "your_radarr_api_key"
  quality_profile_id: 1
  root_folder: "/movies"
  search_on_add: true
  monitored: true
  minimum_availability: "released"  # announced, inCinemas, released, preDB

# Sonarr Integration (Optional)
sonarr:
  enabled: true
  url: "http://sonarr:8989"
  api_key: "your_sonarr_api_key"
  quality_profile_id: 1
  root_folder: "/tv"
  search_on_add: true
  monitored: true
  season_folder: true
  series_type: "standard"  # standard, daily, anime

# Jellyseerr Integration (Optional)
jellyseerr:
  enabled: false
  url: "http://jellyseerr:5055"
  api_key: "your_jellyseerr_api_key"

# Integration Mode
integration:
  mode: "direct"  # direct or jellyseerr

# Jobs Configuration
jobs:
  trending_movies:
    enabled: true
    schedule: "0 */6 * * *"  # Every 6 hours
    limit: 20
    filters:
      min_rating: 6.5
      min_votes: 1000
      max_runtime: 180
      exclude_genres: ["Documentary"]
      
  popular_shows:
    enabled: true
    schedule: "0 0 * * *"  # Daily at midnight
    limit: 15
    filters:
      min_rating: 7.0
      min_votes: 500
```

## Server Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `server.host` | string | `"0.0.0.0"` | Host to bind the server to |
| `server.port` | int | `9090` | Port to run the server on |

## Database Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `database.path` | string | `"data/blockbusterr.db"` | Path to SQLite database file |

## Trakt Configuration

Trakt is required for fetching trending, popular, and other list-based content.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `trakt.client_id` | string | Yes | Trakt API client ID |
| `trakt.client_secret` | string | Yes | Trakt API client secret |
| `trakt.access_token` | string | Auto | OAuth access token (auto-populated) |
| `trakt.refresh_token` | string | Auto | OAuth refresh token (auto-populated) |

### Getting Trakt Credentials

1. Go to [Trakt API Applications](https://trakt.tv/oauth/applications)
2. Create a new application
3. Set redirect URI to `http://localhost:9090/auth/trakt/callback`
4. Copy the Client ID and Client Secret

## TMDB Configuration (Optional)

TMDB provides poster images for the preview feature.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tmdb.api_key` | string | No | TMDB API key for poster images |

See [TMDB Setup Guide](/blockbusterr/integrations/tmdb/) for details.

## Radarr Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `radarr.enabled` | bool | `false` | Enable Radarr integration |
| `radarr.url` | string | - | Radarr instance URL |
| `radarr.api_key` | string | - | Radarr API key |
| `radarr.quality_profile_id` | int | `1` | Quality profile ID to use |
| `radarr.root_folder` | string | - | Root folder path |
| `radarr.search_on_add` | bool | `true` | Automatically search for movie |
| `radarr.monitored` | bool | `true` | Monitor the movie |
| `radarr.minimum_availability` | string | `"released"` | Minimum availability: `announced`, `inCinemas`, `released`, `preDB` |

## Sonarr Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sonarr.enabled` | bool | `false` | Enable Sonarr integration |
| `sonarr.url` | string | - | Sonarr instance URL |
| `sonarr.api_key` | string | - | Sonarr API key |
| `sonarr.quality_profile_id` | int | `1` | Quality profile ID to use |
| `sonarr.root_folder` | string | - | Root folder path |
| `sonarr.search_on_add` | bool | `true` | Automatically search for episodes |
| `sonarr.monitored` | bool | `true` | Monitor the series |
| `sonarr.season_folder` | bool | `true` | Use season folders |
| `sonarr.series_type` | string | `"standard"` | Series type: `standard`, `daily`, `anime` |

## Jellyseerr Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `jellyseerr.enabled` | bool | `false` | Enable Jellyseerr integration |
| `jellyseerr.url` | string | - | Jellyseerr instance URL |
| `jellyseerr.api_key` | string | - | Jellyseerr API key |

## Integration Mode

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `integration.mode` | string | `"direct"` | Integration mode: `direct` (Radarr/Sonarr) or `jellyseerr` |

See [Integration Modes](/blockbusterr/concepts/integration-modes/) for details.

## Job Configuration

Each job can have the following configuration:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Enable or disable the job |
| `schedule` | string | - | Cron schedule (e.g., `"0 */6 * * *"`) |
| `limit` | int | `10` | Maximum items to fetch |
| `filters` | object | - | Filter configuration (see below) |
| `scoring` | object | - | Scoring configuration (see below) |

### Available Jobs

- `trending_movies`, `trending_shows`
- `popular_movies`, `popular_shows`
- `box_office`
- `anticipated_movies`, `anticipated_shows`
- `favorited_movies`, `favorited_shows`
- `played_movies`, `played_shows`
- `watched_movies`, `watched_shows`
- `collected_movies`, `collected_shows`

## Filter Configuration

Filters allow you to control what content gets added. All filter fields are optional.

```yaml
filters:
  # Rating filters
  min_rating: 6.5        # Minimum Trakt rating (0-10)
  max_rating: 10.0       # Maximum Trakt rating
  
  # Vote filters
  min_votes: 1000        # Minimum number of votes
  
  # Runtime filters
  min_runtime: 60        # Minimum runtime in minutes
  max_runtime: 180       # Maximum runtime in minutes
  
  # Release date filters
  min_year: 2020         # Minimum release year
  max_year: 2026         # Maximum release year
  
  # Genre filters
  include_genres: ["Action", "Sci-Fi"]
  exclude_genres: ["Documentary", "Reality"]
  
  # Keyword filters
  include_keywords: ["superhero"]
  exclude_keywords: ["christmas", "hallmark"]
  
  # Language filters
  include_languages: ["en", "fr"]
  exclude_languages: ["hi"]
  
  # Country filters
  include_countries: ["us", "gb"]
  exclude_countries: ["in"]
  
  # Status filters (shows only)
  allowed_statuses: ["continuing", "returning series"]
```

See [Filters & Scoring](/blockbusterr/concepts/filters/) for detailed explanation.

## Cron Schedule Format

Jobs use standard cron syntax for scheduling:

```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
* * * * *
```

Examples:
- `"0 */6 * * *"` - Every 6 hours
- `"0 0 * * *"` - Daily at midnight
- `"0 12 * * 1"` - Every Monday at noon
- `"*/30 * * * *"` - Every 30 minutes

## Environment Variables

Override configuration using environment variables:

```bash
SERVER_PORT=9090
DATABASE_PATH=/app/data/blockbusterr.db
TRAKT_CLIENT_ID=your_client_id
TRAKT_CLIENT_SECRET=your_client_secret
RADARR_URL=http://radarr:7878
RADARR_API_KEY=your_api_key
```

Environment variables take precedence over configuration file values.
