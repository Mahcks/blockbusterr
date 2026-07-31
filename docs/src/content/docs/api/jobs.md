---
title: Jobs API
description: API endpoints for managing and previewing jobs
---

The Jobs API allows you to preview content, trigger jobs manually, and check job status.

## Preview Job

Preview what content a job would add without actually adding anything.

**Endpoint:** `POST /v1/jobs/preview/{job-name}`

**Parameters:**
- `job-name` (path, required) - Job identifier (e.g., `trending-movies`, `popular-shows`)

**Available Jobs:**
- `trending-movies`, `trending-shows`
- `popular-movies`, `popular-shows`
- `smart-popular-movies`, `smart-popular-shows`
- `box-office`
- `favorited-movies`, `favorited-shows`
- `played-movies`, `played-shows`
- `watched-movies`, `watched-shows`
- `collected-movies`, `collected-shows`
- `anticipated-movies`, `anticipated-shows`

**Example Request:**

```bash
curl -X POST http://localhost:9090/v1/jobs/preview/trending-movies
```

**Example Response:**

```json
{
  "job_name": "trending_movies",
  "total_found": 25,
  "will_add": 10,
  "already_exists": 8,
  "filtered_out": 7,
  "mode": "direct",
  "has_posters": true,
  "items": [
    {
      "title": "Dune: Part Two",
      "year": 2024,
      "tmdb_id": 693134,
      "imdb_id": "tt15239678",
      "poster_url": "https://image.tmdb.org/t/p/w500/...",
      "overview": "Follow the mythic journey of Paul Atreides...",
      "rating": 8.7,
      "votes": 5234,
      "popularity": 95000,
      "genres": ["Science Fiction", "Adventure"],
      "runtime": 166,
      "already_exists": false,
      "filtered_out": false,
      "filter_reason": ""
    }
  ]
}
```

**Response Fields:**
- `job_name` - Name of the job
- `total_found` - Total items returned by the selected discovery source
- `will_add` - Items that would be added (pass filters, not in library)
- `already_exists` - Items already in your library
- `filtered_out` - Items rejected by filters
- `mode` - Integration mode (`direct` or `jellyseerr`)
- `has_posters` - Whether TMDB poster images are available
- `items` - Array of preview items

**Item Fields:**
- `title` - Content title
- `year` - Release year
- `tmdb_id` - TMDB ID (movies and shows)
- `tvdb_id` - TVDB ID (shows only)
- `imdb_id` - IMDB ID
- `poster_url` - Poster image URL (if TMDB configured)
- `overview` - Content description
- `rating` - Provider rating when the selected source supplies one
- `votes` - Number of votes
- `popularity` - Popularity score
- `genres` - Array of genre names
- `runtime` - Runtime in minutes
- `already_exists` - Whether already in library
- `filtered_out` - Whether rejected by filters
- `filter_reason` - Reason for rejection (if filtered)

## Trigger Job

Manually trigger a job to run immediately.

**Endpoint:** `POST /v1/jobs/trigger/{job-name}`

**Parameters:**
- `job-name` (path, required) - Job identifier

**Example Request:**

```bash
curl -X POST http://localhost:9090/v1/jobs/trigger/trending-movies
```

**Example Response:**

```json
{
  "message": "Job 'trending-movies' triggered successfully",
  "job": "trending-movies"
}
```

## Job Status

Get the status of all configured jobs.

**Endpoint:** `GET /v1/jobs/status`

**Example Request:**

```bash
curl http://localhost:9090/v1/jobs/status
```

**Example Response:**

```json
{
  "jobs": [
    {
      "name": "trending_movies",
      "enabled": true,
      "last_run": "2026-01-08T15:30:00Z",
      "next_run": "2026-01-08T16:30:00Z",
      "sync_interval": "1h",
      "mode": "direct"
    },
    {
      "name": "popular_shows",
      "enabled": false,
      "last_run": null,
      "next_run": null,
      "sync_interval": "6h",
      "mode": "direct"
    }
  ]
}
```

## Dynamic Job Management

*New in v1.4.0*

### List All Jobs

Get all configured jobs (both dynamic and legacy).

**Endpoint:** `GET /v1/jobs/list`

**Example Request:**

```bash
curl http://localhost:9090/v1/jobs/list
```

**Example Response:**

```json
[
  {
    "id": "my-trending-movies",
    "name": "My Trending Movies",
    "enabled": true,
    "type": "trending",
    "source": "trakt",
    "media": "movie",
    "limit": 50,
    "sync_interval": "6h"
  },
  {
    "id": "legacy_popular_movies",
    "name": "Popular Movies",
    "enabled": true,
    "type": "popular",
    "source": "trakt",
    "media": "movie",
    "limit": 20
  }
]
```

**Job Fields:**
- `id` - Unique identifier (user-provided or auto-generated UUID)
- `name` - Display name
- `enabled` - Whether the job runs on schedule
- `type` - Job type: `trending`, `popular`, `watched`, `collected`, `favorited`, `played`, `anticipated`, `box_office`, `smart_popular`
- `source` - Discovery source. Supported values depend on the job type and configured providers.
- `media` - Media type: `movie` or `show`
- `limit` - Number of items to fetch
- `period` - Time period for watched/collected/favorited/played: `weekly`, `monthly`, `yearly`, `all`
- `sync_interval` - Custom sync interval (e.g., `6h`, `12h`, `24h`)
- `mode` - Execution mode: `direct` or `jellyseerr`
- `minimum_availability` - For Radarr: `announced`, `in_cinemas`, `released`
- `monitor` - Monitor setting for Radarr/Sonarr
- `base_min_rating` - For smart jobs: base rating threshold
- `adjustment_factor` - For smart jobs: popularity adjustment factor

### Get Job Types

Get all available job type definitions for building UI dropdowns.

**Endpoint:** `GET /v1/jobs/types`

**Example Response:**

```json
{
  "trending": {
    "type": "trending",
    "name": "Trending",
    "description": "Currently being watched and talked about",
    "source": "trakt",
    "sources": ["trakt", "tmdb", "simkl"],
    "supported_media": ["movie", "show"],
    "requires_period": false,
    "is_smart_job": false,
    "default_limit": 50,
    "max_limit": 1000
  },
  "box_office": {
    "type": "box_office",
    "name": "Box Office",
    "description": "Top weekend box office movies (Trakt returns max 10)",
    "source": "trakt",
    "supported_media": ["movie"],
    "requires_period": false,
    "is_smart_job": false,
    "default_limit": 10,
    "max_limit": 10
  }
}
```

### Get Templates

Get pre-configured job templates for quick setup.

**Endpoint:** `GET /v1/jobs/templates`

**Example Response:**

```json
[
  {
    "name": "Weekly Trending Movies",
    "description": "Top 10 trending movies this week",
    "type": "trending",
    "media": "movie",
    "limit": 10,
    "category": "Movies"
  },
  {
    "name": "Weekly Watched Shows",
    "description": "Most watched shows this week",
    "type": "watched",
    "media": "show",
    "limit": 20,
    "period": "weekly",
    "category": "TV Shows"
  }
]
```

### Create Job

Create a new dynamic job.

**Endpoint:** `POST /v1/jobs`

**Request Body:**

```json
{
  "id": "my-custom-job",
  "name": "My Custom Job",
  "type": "trending",
  "source": "tmdb",
  "media": "movie",
  "enabled": true,
  "limit": 50,
  "sync_interval": "6h"
}
```

**Notes:**
- `id` is optional - a UUID will be generated if not provided
- `type`, `media`, and `name` are required
- `limit` defaults to the job type's `default_limit` if not specified

**Example Response:**

```json
{
  "id": "my-custom-job",
  "name": "My Custom Job",
  "enabled": true,
  "type": "trending",
  "source": "tmdb",
  "media": "movie",
  "limit": 50,
  "sync_interval": "6h"
}
```

### Update Job

Update an existing dynamic job.

**Endpoint:** `PUT /v1/jobs/:id`

**Parameters:**
- `id` (path, required) - Job ID

**Request Body:** Same fields as Create Job

**Note:** Legacy jobs (IDs starting with `legacy_`) cannot be updated through this endpoint.

### Delete Job

Delete a dynamic job.

**Endpoint:** `DELETE /v1/jobs/:id`

**Parameters:**
- `id` (path, required) - Job ID

**Example Request:**

```bash
curl -X DELETE http://localhost:9090/v1/jobs/my-custom-job
```

**Note:** Legacy jobs cannot be deleted - disable them in your config instead.

### Trigger Dynamic Job

Run a dynamic job immediately.

**Endpoint:** `POST /v1/jobs/:id/trigger`

**Parameters:**
- `id` (path, required) - Job ID

**Example Response:**

```json
{
  "message": "Job 'My Custom Job' triggered successfully"
}
```

### Preview Dynamic Job

Preview what a dynamic job would add without executing.

**Endpoint:** `POST /v1/jobs/:id/preview`

**Parameters:**
- `id` (path, required) - Job ID

### Migrate Legacy Jobs

Convert legacy config jobs to dynamic format.

**Endpoint:** `POST /v1/jobs/migrate`

**Example Request:**

```bash
curl -X POST http://localhost:9090/v1/jobs/migrate
```

**Example Response:**

```json
{
  "message": "Successfully migrated 5 legacy jobs",
  "count": 5,
  "jobs": [
    {
      "id": "trending_movies",
      "name": "Trending Movies",
      "enabled": true,
      "type": "trending",
      "source": "trakt",
      "media": "movie",
      "limit": 20
    }
  ]
}
```

## Advanced Examples

### Preview with jq Filtering

Get only items that will be added:

```bash
curl -s -X POST http://localhost:9090/v1/jobs/preview/trending-movies \
  | jq '.items[] | select(.will_add == true) | {title, year, rating}'
```

### Get Filtered Out Items

See what was rejected and why:

```bash
curl -s -X POST http://localhost:9090/v1/jobs/preview/trending-movies \
  | jq '.items[] | select(.filtered_out == true) | {title, filter_reason}'
```

### Check All Job Statuses

```bash
curl -s http://localhost:9090/v1/jobs/status | jq '.jobs[] | {name, enabled, last_run}'
```

## Error Responses

**Job Not Found (404):**
```json
{
  "error": "Job not found: invalid-job-name"
}
```

**Job Disabled (400):**
```json
{
  "error": "Job 'trending-movies' is not enabled"
}
```

**Execution Error (500):**
```json
{
  "error": "Failed to fetch trending movies: connection timeout"
}
```

## Next Steps

- Explore [Activity API](/api/activity/)
- Review [Configuration API](/api/config/)
- Learn about [Jobs](/concepts/jobs/)
