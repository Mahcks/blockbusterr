---
title: Configuration
description: Configuration reference for Blockbusterr
---

Blockbusterr reads YAML from `config/config.yaml` in releases and `config/config.dev.yaml` for a development build. The Web UI writes the same configuration model.

## Complete example

```yaml
version: "1.5.0"

# Configure at least one discovery provider.
trakt:
  client_id: your_trakt_client_id
  client_secret: your_trakt_client_secret

tmdb:
  api_key: your_tmdb_api_key

simkl:
  client_id: your_simkl_client_id

radarr:
  url: http://radarr:7878
  api_key: your_radarr_api_key
  quality_profile: 1
  root_folder: /movies
  minimum_availability: announced
  monitor: movieOnly

sonarr:
  url: http://sonarr:8989
  api_key: your_sonarr_api_key
  quality_profile: 1
  root_folder: /tv
  monitor: all

jellyseerr:
  url: http://jellyseerr:5055
  api_key: your_jellyseerr_api_key
  user_id: ""

jobs:
  sync_interval: 24h
  mode: direct
  list:
    - id: trending-movies
      name: Trending Movies
      enabled: true
      type: trending
      source: tmdb
      media: movie
      limit: 20
      sync_interval: 6h
      minimum_availability: announced
      monitor: movieOnly

filters:
  movies:
    allowed_countries: []
    allowed_languages: [en]
    blacklisted_genres: [documentary]
    blacklisted_keywords: []
    blacklisted_tmdb_ids: []
    blacklisted_min_runtime: 0
    blacklisted_max_runtime: 180
    blacklisted_min_year: 0
    blacklisted_max_year: 0
    min_rating: 6.5
    min_votes: 1000
  shows:
    allowed_countries: []
    allowed_languages: [en]
    blacklisted_genres: []
    blacklisted_keywords: []
    blacklisted_networks: []
    blacklisted_tvdb_ids: []
    blacklisted_min_runtime: 0
    blacklisted_max_runtime: 0
    blacklisted_min_year: 0
    blacklisted_max_year: 0
    min_rating: 7.0
    min_votes: 500
```

You only need credentials for providers and delivery integrations you use. Trakt is optional when all enabled jobs use TMDB or Simkl.

## Discovery providers

| Field | Purpose |
|---|---|
| `trakt.client_id` | Enables Trakt job sources |
| `trakt.client_secret` | Optional Trakt application secret |
| `tmdb.api_key` | Enables TMDB job sources and poster enrichment |
| `simkl.client_id` | Enables Simkl job sources |

Each job type advertises its supported sources in the Jobs UI and `GET /v1/jobs/types`.

## Delivery integrations

Radarr and Sonarr are used by `direct` jobs. Jellyseerr is used by `jellyseerr` jobs. An integration is considered configured when its required URL and API key are present; there is no separate `enabled` property.

| Integration | Fields |
|---|---|
| Radarr | `url`, `api_key`, `quality_profile`, `root_folder`, `minimum_availability`, `monitor` |
| Sonarr | `url`, `api_key`, `quality_profile`, `root_folder`, `monitor` |
| Jellyseerr | `url`, `api_key`, optional `user_id`, optional `request_credentials.email` and `request_credentials.password` |

## Dynamic jobs

Jobs created in the UI are stored under `jobs.list`.

| Field | Description |
|---|---|
| `id` | Stable unique job identifier |
| `name` | Display name |
| `enabled` | Whether the scheduler may run the job |
| `type` | `trending`, `popular`, `watched`, `collected`, `favorited`, `played`, `anticipated`, `box_office`, or `smart_popular` |
| `source` | Discovery provider supported by the selected type |
| `media` | `movie` or `show` |
| `limit` | Maximum candidates fetched |
| `period` | Required by time-based job types: `weekly`, `monthly`, `yearly`, or `all` |
| `sync_interval` | Go duration such as `6h`, or a five-field cron expression |
| `mode` | Optional `direct` or `jellyseerr` override |
| `minimum_availability` | Optional Radarr override |
| `monitor` | Optional Radarr/Sonarr override |
| `base_min_rating` | Smart Popular baseline rating |
| `adjustment_factor` | Smart Popular popularity adjustment |

Legacy named job sections remain readable for backward compatibility. Use the Jobs UI for new configurations.

## Filters and scoring

Filters are global per media type; per-job filters are not supported. See [Filters & Scoring](/concepts/filters/) for the complete behavior.

## Environment variables

Viper maps nested keys by replacing dots with underscores. For example, `TMDB_API_KEY` overrides `tmdb.api_key` and `SIMKL_CLIENT_ID` overrides `simkl.client_id`.

Application-level environment variables are:

- `CONFIG_PATH`: additional directory to search for the config file
- `DATA_DIR`: database directory
- `DISABLE_UI`: disable UI routes when `true`, `1`, or `yes`
- `VERSION` and `COMMIT`: displayed build metadata overrides
