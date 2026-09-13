---
title: Configuration
description: Set up connections, save configuration, control additions, and test automation before enabling it.
---

Use **Settings** for connections and automation defaults, **Rules** for eligibility, and **Jobs** for individual discovery schedules. Most installations do not need hand-written YAML.

Blockbusterr reads `config.yaml` in release builds and `config.dev.yaml` in development. It searches the directory named by `CONFIG_PATH` first, then the standard locations, including `./config` and container configuration directories. `CONFIG_PATH` is a **directory**, not a filename. The UI saves to the configuration file the application loaded.

Keep that directory on persistent storage and writable by the container's application user. An environment-only installation can start without a configuration file, but cannot save UI changes because it has no file path to write to. For a UI-managed installation, create the file before starting; see [Installation](/getting-started/installation/).

Configuration files contain API keys and other credentials. Blockbusterr writes them with owner-only permissions; keep downloaded configuration backups private as well. The Settings page backup contains YAML configuration only—it does not include SQLite Activity history or repeat state. Back up the complete `data` directory when you need a recoverable installation snapshot.

## Start with a small, reviewable setup

1. Configure one discovery source under **Settings → Connections** and test it. You do not need every provider.
2. Configure Radarr for movies or Sonarr for shows. Select the actual quality profile and root folder from that service.
3. Under **Settings → Automation**, choose Direct mode and modest movie/show delivery caps. Start with a few additions per week while tuning.
4. Create a recipe job. Recipes start disabled, so you can inspect its rule set and Preview first.
5. Review accepted and rejected titles, then enable the job. Check **Activity Entries** and **Job Runs** after its first live execution.

Custom job creation creates an enabled job. If you need the whole installation to simulate delivery while you configure custom jobs, set `BLOCKBUSTERR_DRY_RUN=true` and recreate the container with its existing data mount first; see [Dry run](#dry-run-and-preview).

## Example configuration

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
      enabled: false
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

The example job is deliberately disabled. Replace the API keys, URL, quality-profile ID, and root folder before using it. The profile ID `1` is only an example: it must exist in your Radarr instance. `/movies` is the path Radarr knows, which may differ from a host filesystem path. When both services run in Docker, `http://radarr:7878` works only if that service name is reachable on a shared network; `localhost` inside Blockbusterr refers to its own container.

This job examines up to 20 trending movie candidates every six hours, then applies the rule set: English, not documentary, year 2000 onward, runtime up to 180 minutes when known, rating at least 6.5, and at least 1,000 votes. It can deliver at most five per run, within the global rolling cap of 20 movies per seven days. It does not promise five matches or downloads per run. Enable it through Jobs after Preview.

You only need one configured discovery provider and the integrations required by enabled jobs. Trakt is optional when all enabled jobs use TMDB, Simkl, or MDBList.

## Discovery providers

| Field | Purpose |
|---|---|
| `trakt.client_id` | Enables Trakt discovery; provider account/API eligibility still applies |
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

The root folder and quality profile tell Radarr/Sonarr where and how to manage a title; they do not create an indexer or download client. Configure those in Radarr/Sonarr. A successful Blockbusterr delivery means the downstream service accepted the addition/request, not that a playable file has downloaded.

Radarr's **Minimum availability** controls when it considers a movie available for downloading; it is separate from Blockbusterr's release-year filter. The monitor setting controls what Radarr/Sonarr monitors. Jellyseerr/Seerr can add a further approval step depending on the requesting user's permissions. See [Integration modes](/concepts/integration-modes/).

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

The global `jobs.sync_interval` and `jobs.mode` values are defaults. Job-level values override them. A job's `rule_set_id` selects its entire rule set; it does not add another layer on top of Default Movies or Default Shows.

A duration such as `6h` runs on startup and repeats at that interval. A five-field cron expression, such as `0 3 * * *`, waits for its next scheduled occurrence. Check your deployment's timezone before relying on a wall-clock schedule. An enabled duration job can run again when you restart Blockbusterr; repeat checks and global caps still apply.

List jobs also store a `list` locator with provider identifiers. Recommendations use `recommendation_seeds` and/or `recommendation_list`. Use the Jobs editor to construct these rather than putting a full URL into an identifier field. [Jobs](/concepts/jobs/) explains each discovery type.

## Delivery budgets

Delivery budgets are enforced immediately before Blockbusterr calls Radarr, Sonarr, Jellyseerr, or Seerr:

- `jobs.global_limit_movies` caps successful movie deliveries across all jobs.
- `jobs.global_limit_shows` caps successful show deliveries across all jobs.
- `jobs.global_period` uses a rolling `daily` (24-hour), `weekly` (7-day), or `monthly` (30-day) window.
- Each job's `delivery_limit` caps successful deliveries within that run.

Zero means unlimited. Successful additions and requests consume a slot. Rejected, skipped, duplicate, and previewed titles do not count. A lost response, timeout, malformed success response, or server error may follow a successful delivery, so its reservation is retained conservatively. Check the downstream service before retrying; a failed Activity Entry does not prove the title was rejected. Rolling reservations expire with the configured window. When both limits apply, the first exhausted budget skips that delivery and records the reason in Activity Entries.

### Candidate limits, delivery caps, and disk space

These limits answer different questions:

| Control | Example | Effect |
|---|---|---|
| Discovery limit (`limit`) | 100 | Examine up to 100 candidates from the source before filtering |
| Job delivery cap (`delivery_limit`) | 3 | Deliver no more than three successful additions/requests from that job run |
| Global movie cap | 10 weekly | Share ten movie delivery slots across all jobs in a rolling seven-day window |
| Ranked-selection movie slots | 5 | Choose up to five winners for a shared cycle, subject to the other caps |

For example, 100 candidates might become 12 rule matches, eight already present titles, and four new eligible titles. A job cap of three permits at most three additions; if only one global slot remains, at most one can be delivered. A single-job Preview is useful for checking eligibility, but its accepted count is not a forecast that all those titles will fit the live delivery caps.

These are **title-count limits**, not gigabyte limits or free-space checks. One show delivery can lead Sonarr to download many episodes. Choose sensible quality profiles, monitoring, download-client settings, and disk limits in the downstream services. Blockbusterr does not reclaim disk space by removing titles that disappear from a source list.

## Dry run and Preview

**Preview** evaluates a selected job without delivering its candidates. It still reads discovery providers and checks available integration state, so working credentials are needed. It does not consume delivery slots or start repeat cooldowns.

For a deployment-wide trial, set this environment variable and recreate the container with its existing data mount:

```yaml
# Add to the Blockbusterr service's existing Docker Compose environment section.
environment:
  BLOCKBUSTERR_DRY_RUN: "true"
```

Dry run simulates job delivery instead of posting additions/requests. It is independent of whether a job is enabled: schedules and discovery can still run. Remove the variable or set it to `"false"` and recreate the container when you are ready for live delivery. For Compose, use `docker compose up -d blockbusterr` after changing the environment; a plain restart does not apply new container environment values. Development builds identified as `dev` always use dry run, even when this variable is false.

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

Weights must total `1.0`. Scoring ranks candidates; the assigned rule set determines eligibility. A high score cannot rescue a rule rejection. The rating component uses the source's candidate rating; setting `rating_scale: 10` does not switch it to IMDb.

Leave ranked selection off while setting up your first job. Enable shared cycles when several jobs should compete for a combined set of winners; then configure cycle slots and optional job minimum picks under **Settings → Automation**. See [Ranked selection cycles](/concepts/jobs/#ranked-selection-cycles).

## Legacy compatibility

v1 `filters.movies`, `filters.shows`, embedded custom filters, and named job sections remain readable during upgrade. Blockbusterr migrates them into default or job-specific rule sets deterministically. Avoid maintaining both formats after migration; use the v2 UI and keep a backup of the original configuration.

The old aggregation-era Global Limits implementation was replaced by delivery-time enforcement. Existing `global_limit_movies`, `global_limit_shows`, and daily/weekly/monthly `global_period` values remain compatible. The old `sync` period is normalized to `daily` because it did not represent a shared enforceable window.

## Environment variables

Environment overrides use uppercase YAML field names joined with underscores. For example, `TMDB_API_KEY` overrides `tmdb.api_key` and `SIMKL_CLIENT_ID` overrides `simkl.client_id`.

- `CONFIG_PATH`: directory searched first for configuration
- `DATA_DIR`: SQLite database directory
- `DISABLE_UI`: disable UI routes when `true`, `1`, or `yes`
- `BLOCKBUSTERR_DRY_RUN`: simulate delivery in a release build when `true`
- `VERSION` and `COMMIT`: build metadata overrides


Environment values override the loaded YAML at startup. If a setting keeps changing back after a restart, check the container environment as well as the file. Changing a variable in Compose requires recreating the container with that environment; editing the running UI does not change the Compose file.

## Save, back up, and restore

Use the Settings configuration backup before a large change. For a complete recovery, preserve both the loaded configuration directory and the SQLite data directory: configuration stores connections, jobs, and rules; SQLite stores Activity history and delivery/repeat state. Stop the application before copying a live database directory so its files are consistent.

If editing YAML by hand, stop Blockbusterr first, edit the loaded file, and restart. This avoids a UI save overwriting your manual edit. Do not paste a whole example over an existing configuration just to change one field. Keep API keys private when sharing configuration for troubleshooting.

| Problem | Check |
|---|---|
| UI cannot save | The loaded config file exists, its directory is writable, and the persistent volume is mounted correctly |
| Saved value changes on restart | An environment variable overrides that field, or a different configuration directory is loaded |
| Radarr/Sonarr connection fails | Container DNS/network, URL and port, API key, and whether `localhost` points at the wrong container |
| Preview accepts titles but no live additions appear | Job enabled state, dry run, duplicate/repeat decisions, job/global caps, and downstream approval |
| Rules reject every list item | Whether the list integration supplies the required metadata; see [Rules](/concepts/rules/#missing-metadata-is-handled-differently-by-field) |
