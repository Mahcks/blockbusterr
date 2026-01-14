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
- `total_found` - Total items found from Trakt
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
- `rating` - Trakt rating (0-10)
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

- Explore [Activity API](/blockbusterr/api/activity/)
- Review [Configuration API](/blockbusterr/api/config/)
- Learn about [Jobs](/blockbusterr/concepts/jobs/)
