# API Documentation

## Overview

Blockbusterr provides a RESTful API for managing jobs, viewing activity, and previewing content.

Base URL: `http://localhost:9090/v1`

All endpoints return JSON unless otherwise specified.

---

## Jobs API

### Preview Job

Preview what content a job would add without actually adding anything.

**Endpoint:** `POST /v1/jobs/preview/{job-name}`

**Parameters:**
- `job-name` (path) - Job identifier (e.g., `trending-movies`, `popular-shows`)

**Available Jobs:**
- `trending-movies`
- `trending-shows`
- `popular-movies`
- `popular-shows`
- `box-office`
- `favorited-movies`
- `favorited-shows`
- `played-movies`
- `played-shows`
- `watched-movies`
- `watched-shows`
- `collected-movies`
- `collected-shows`
- `anticipated-movies`
- `anticipated-shows`

**Response:**

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
    },
    {
      "title": "Low Budget Horror",
      "year": 2024,
      "tmdb_id": 123456,
      "rating": 4.2,
      "votes": 150,
      "genres": ["Horror"],
      "runtime": 82,
      "already_exists": false,
      "filtered_out": true,
      "filter_reason": "Rating 4.2 below minimum threshold"
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
- `already_exists` - Boolean indicating if in library
- `filtered_out` - Boolean indicating if rejected by filters
- `filter_reason` - Reason for rejection (if filtered)

**Example Request:**

```bash
curl -X POST http://localhost:9090/v1/jobs/preview/trending-movies
```

**Example with jq:**

```bash
curl -s -X POST http://localhost:9090/v1/jobs/preview/trending-movies | jq '.items[] | select(.will_add == true) | {title, year, rating}'
```

---

### Trigger Job

Manually trigger a job to run immediately.

**Endpoint:** `POST /v1/jobs/trigger/{job-name}`

**Example:**

```bash
curl -X POST http://localhost:9090/v1/jobs/trigger/trending-movies
```

**Response:**

```json
{
  "message": "Job 'trending-movies' triggered successfully",
  "job": "trending-movies"
}
```

---

### Job Status

Get the status of all configured jobs.

**Endpoint:** `GET /v1/jobs/status`

**Response:**

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
    }
  ]
}
```

---

## Activity API

### Activity Logs

Get recent activity logs.

**Endpoint:** `GET /v1/activity/logs`

**Query Parameters:**
- `limit` (optional) - Number of logs to return (default: 50, max: 500)
- `media_type` (optional) - Filter by type: `movie` or `show`
- `job_type` (optional) - Filter by job name
- `status` (optional) - Filter by status: `added`, `requested`, `failed`

**Example:**

```bash
curl "http://localhost:9090/v1/activity/logs?limit=100&media_type=movie"
```

**Response:**

```json
{
  "logs": [
    {
      "id": 123,
      "timestamp": "2026-01-08T15:30:45Z",
      "job_type": "trending_movies",
      "media_type": "movie",
      "title": "Dune: Part Two",
      "year": 2024,
      "tmdb_id": 693134,
      "imdb_id": "tt15239678",
      "poster_url": "https://www.themoviedb.org/movie/693134",
      "status": "added"
    }
  ],
  "count": 100
}
```

---

### Activity Stats

Get statistics about activity.

**Endpoint:** `GET /v1/activity/stats`

**Response:**

```json
{
  "total_items": 1523,
  "movies_added": 892,
  "shows_added": 631,
  "last_24h": 45,
  "last_7d": 312
}
```

---

### Clear Activity Logs

Delete activity logs older than specified days.

**Endpoint:** `DELETE /v1/activity/logs`

**Query Parameters:**
- `days` (required) - Delete logs older than N days

**Example:**

```bash
curl -X DELETE "http://localhost:9090/v1/activity/logs?days=30"
```

---

## Trakt API

### Trending Content

Get trending content from Trakt.

**Endpoints:**
- `GET /v1/trakt/trending/movies?limit=10`
- `GET /v1/trakt/trending/shows?limit=10`

**Query Parameters:**
- `limit` (optional) - Number of items (default: 10)

**Response:**

```json
{
  "data": [
    {
      "watchers": 95000,
      "movie": {
        "title": "Dune: Part Two",
        "year": 2024,
        "ids": {
          "trakt": 334536,
          "slug": "dune-part-two-2024",
          "imdb": "tt15239678",
          "tmdb": 693134
        },
        "genres": ["science-fiction", "adventure"],
        "rating": 8.7,
        "votes": 5234,
        "overview": "Follow the mythic journey..."
      }
    }
  ],
  "count": 10
}
```

---

### Popular Content

Get popular content from Trakt.

**Endpoints:**
- `GET /v1/trakt/popular/movies?limit=10`
- `GET /v1/trakt/popular/shows?limit=10`

---

### Metadata

Get available genres, countries, languages, and networks for filtering.

**Endpoints:**
- `GET /v1/trakt/genres/movies` - Movie genres
- `GET /v1/trakt/genres/shows` - TV show genres
- `GET /v1/trakt/countries/movies` - Movie countries
- `GET /v1/trakt/countries/shows` - TV show countries
- `GET /v1/trakt/languages/movies` - Movie languages
- `GET /v1/trakt/languages/shows` - TV show languages
- `GET /v1/trakt/networks` - TV networks

**Example:**

```bash
curl http://localhost:9090/v1/trakt/genres/movies
```

**Response:**

```json
{
  "data": [
    {"slug": "action", "name": "Action"},
    {"slug": "adventure", "name": "Adventure"},
    {"slug": "animation", "name": "Animation"}
  ]
}
```

---

## Radarr/Sonarr API

### Validate Connection

Test connection to Radarr or Sonarr.

**Endpoints:**
- `GET /v1/radarr/validate?test=true&url=http://radarr:7878&api_key=KEY`
- `GET /v1/sonarr/validate?test=true&url=http://sonarr:8989&api_key=KEY`

**Query Parameters:**
- `test` (optional) - Set to `true` to test custom credentials
- `url` (required if test=true) - Server URL
- `api_key` (required if test=true) - API key

---

### Get Quality Profiles

Get available quality profiles.

**Endpoints:**
- `GET /v1/radarr/quality-profiles`
- `GET /v1/sonarr/quality-profiles`

**Response:**

```json
{
  "data": [
    {"id": 1, "name": "Any"},
    {"id": 4, "name": "HD-1080p"},
    {"id": 6, "name": "Ultra-HD"}
  ],
  "count": 3
}
```

---

### Get Root Folders

Get available root folders.

**Endpoints:**
- `GET /v1/radarr/root-folders`
- `GET /v1/sonarr/root-folders`

**Response:**

```json
{
  "data": [
    {
      "id": 1,
      "path": "/movies",
      "freeSpace": 5368709120,
      "totalSpace": 10737418240,
      "accessible": true
    }
  ],
  "count": 1
}
```

---

## Error Responses

All endpoints return standard HTTP status codes and error messages:

**400 Bad Request:**
```json
{
  "error": "Invalid job name"
}
```

**404 Not Found:**
```json
{
  "error": "Job not found"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Failed to fetch trending movies: connection timeout"
}
```

---

## Rate Limiting

Currently no rate limiting is enforced. Use responsibly.

---

## Authentication

Currently no authentication is required. It's recommended to run Blockbusterr behind a reverse proxy with authentication if exposing to the internet.
