---
title: Configuration
description: Configure v2 providers, delivery integrations, jobs, scoring, and reusable rules.
---

Blockbusterr reads `config/config.yaml` in release builds and `config/config.dev.yaml` in development. The web UI writes the same model and is recommended for normal administration.

Configuration files contain API keys and other credentials. Blockbusterr writes them with owner-only permissions; keep downloaded configuration backups private as well. The Settings page backup contains YAML configuration only—it does not include SQLite Activity history or repeat state. Back up the complete `data` directory when you need a recoverable installation snapshot.

## Minimal v2 example

```yaml
version: "2.0.0"

tmdb:
  api_key: your_tmdb_api_key

radarr:
  url: http://radarr:7878
  api_key: your_radarr_api_key
  quality_profile: 1
  root_folder: /movies
  minimum_availability: announced
  monitor: movieOnly

jobs:
  sync_interval: 24h
  mode: direct
  repeat_policy: 90_days
  global_limit_movies: 20
  global_limit_shows: 10
  global_period: weekly
  list:
    - id: trending-movies
      name: Trending Movies
      enabled: true
      type: trending
      source: tmdb
      media: movie
      limit: 20
      delivery_limit: 5
      sync_interval: 6h
      rule_set_id: default-movies

rule_sets:
  - id: default-movies
    name: Default Movies
    media: movie
    revision: 1
    movies:
      allowed_languages: [en]
      blacklisted_genres: [documentary]
      blacklisted_keywords: []
      blacklisted_tmdb_ids: []
      blacklisted_min_runtime: 0
      blacklisted_max_runtime: 180
      blacklisted_min_year: 2000
      blacklisted_max_year: 0
      min_rating: 6.5
      min_votes: 1000
```

You only need one configured discovery provider and the integrations required by enabled jobs. Trakt is optional when all enabled jobs use TMDB, Simkl, or MDBList.

## Discovery providers

| Field | Purpose |
|---|---|
| `trakt.client_id` | Enables Trakt discovery |
| `trakt.client_secret` | Optional Trakt application secret |
| `tmdb.api_key` | Enables TMDB discovery and poster enrichment |
| `simkl.client_id` | Enables Simkl discovery |
| `mdblist.api_key` | Enables MDBList lists and the account watchlist |

The Jobs editor and `GET /v1/jobs/types` report which sources each job type supports.

## Delivery integrations

| Integration | Required and optional fields |
|---|---|
| Radarr | `url`, `api_key`, `quality_profile`, `root_folder`, `minimum_availability`, `monitor` |
| Sonarr | `url`, `api_key`, `quality_profile`, `root_folder`, `monitor` |
| Jellyseerr / Seerr | `url`, `api_key`, optional `user_id`, optional request email/password |

Direct movie jobs require Radarr; direct show jobs require Sonarr. Jellyseerr-mode jobs require Jellyseerr or Seerr. There is no separate integration `enabled` flag.

## Jobs

Jobs created in the UI are stored under `jobs.list`.

| Field | Description |
|---|---|
| `id` | Stable unique identifier |
| `name` | Display name |
| `enabled` | Whether the scheduler may run the job |
| `type` | Discovery strategy such as `trending`, `popular`, or `anticipated` |
| `source` | Supported discovery provider |
| `media` | `movie` or `show` |
| `limit` | Maximum candidates fetched |
| `delivery_limit` | Maximum successful additions or requests per run; `0` is unlimited |
| `period` | Period for time-based job types |
| `sync_interval` | Go duration or five-field cron override |
| `mode` | Optional `direct` or `jellyseerr` override |
| `minimum_availability` | Optional Radarr override |
| `monitor` | Optional Radarr/Sonarr override |
| `base_min_rating` | Smart Popular baseline rating |
| `adjustment_factor` | Smart Popular adjustment |
| `repeat_policy` | Optional repeat-handling override; empty uses the global default |
| `rule_set_id` | Assigned media-compatible rule set |

The global `jobs.sync_interval` and `jobs.mode` values are defaults. Job-level values override them.

## Delivery budgets

Delivery budgets are enforced immediately before Blockbusterr calls Radarr, Sonarr, Jellyseerr, or Seerr:

- `jobs.global_limit_movies` caps successful movie deliveries across all jobs.
- `jobs.global_limit_shows` caps successful show deliveries across all jobs.
- `jobs.global_period` uses a rolling `daily` (24-hour), `weekly` (7-day), or `monthly` (30-day) window.
- Each job's `delivery_limit` caps successful deliveries within that run.

Zero means unlimited. Only successful additions and requests consume a slot. Rejected, skipped, duplicate, failed, and previewed titles do not count. When both limits apply, the first exhausted budget skips that delivery and records the reason in Activity Entries.

## Repeat handling

Blockbusterr always skips titles currently present in Radarr, Sonarr, Jellyseerr, or Seerr. `jobs.repeat_policy` controls when a title becomes eligible after Blockbusterr previously delivered it and it is later removed. The recommended default is `90_days`; alternatives are `immediate`, `30_days`, `180_days`, and `never`. Each job may override the global policy. Dry runs and previews never start a cooldown.

## Rule sets

Rule sets live under `rule_sets`. A movie rule set contains `movies`; a show rule set contains `shows`. IDs must be unique, names must be unique within a media type, and a job cannot reference a rule set for the other media type.

Blockbusterr maintains `default-movies` and `default-shows`. Use the Rules UI to create, duplicate, assign, and revise policies instead of editing YAML manually.

## Title exceptions

Universal exceptions are stored separately because they apply before the assigned rule set:

```yaml
title_exceptions:
  allowed_movie_tmdb_ids: []
  blocked_movie_tmdb_ids: []
  allowed_show_tvdb_ids: []
  blocked_show_tvdb_ids: []
```

## Scoring

```yaml
scoring:
  enabled: true
  rating_weight: 0.6
  popularity_weight: 0.3
  recency_weight: 0.1
  rating_scale: 10
  popularity_metric: votes
  recency_days: 365
```

Weights must total `1.0`. Scoring ranks candidates; the assigned rule set determines eligibility.

## Legacy compatibility

v1 `filters.movies`, `filters.shows`, embedded custom filters, and named job sections remain readable during upgrade. Blockbusterr migrates them into default or job-specific rule sets deterministically. Avoid maintaining both formats after migration; use the v2 UI and keep a backup of the original configuration.

The old aggregation-era Global Limits implementation was replaced by delivery-time enforcement. Existing `global_limit_movies`, `global_limit_shows`, and daily/weekly/monthly `global_period` values remain compatible. The old `sync` period is normalized to `daily` because it did not represent a shared enforceable window.

## Environment variables

Viper maps nested keys by replacing dots with underscores. For example, `TMDB_API_KEY` overrides `tmdb.api_key` and `SIMKL_CLIENT_ID` overrides `simkl.client_id`.

- `CONFIG_PATH`: additional directory searched for configuration
- `DATA_DIR`: SQLite database directory
- `DISABLE_UI`: disable UI routes when `true`, `1`, or `yes`
- `VERSION` and `COMMIT`: displayed build metadata overrides
