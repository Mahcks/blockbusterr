---
title: Activity API
description: API endpoints for viewing and managing activity logs
---

The Activity API provides access to activity logs and statistics about content added to your library.

## Get Activity Logs

Retrieve recent activity logs with optional filtering.

**Endpoint:** `GET /v1/activity/logs`

**Query Parameters:**
- `limit` (optional) - Number of logs to return (default: 50, max: 500)
- `media_type` (optional) - Filter by type: `movie` or `show`
- `job_type` (optional) - Filter by job name (e.g., `trending_movies`)
- `status` (optional) - Filter by status: `added`, `requested`, `failed`

**Example Requests:**

```bash
# Get last 100 activity logs
curl "http://localhost:9090/v1/activity/logs?limit=100"

# Get only movies
curl "http://localhost:9090/v1/activity/logs?media_type=movie"

# Get trending movies only
curl "http://localhost:9090/v1/activity/logs?job_type=trending_movies"

# Get failed adds
curl "http://localhost:9090/v1/activity/logs?status=failed"
```

**Example Response:**

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
    },
    {
      "id": 124,
      "timestamp": "2026-01-08T15:31:12Z",
      "job_type": "popular_shows",
      "media_type": "show",
      "title": "The Last of Us",
      "year": 2023,
      "tmdb_id": 100088,
      "tvdb_id": 392256,
      "poster_url": "https://www.themoviedb.org/tv/100088",
      "status": "requested"
    }
  ],
  "count": 100
}
```

**Response Fields:**
- `id` - Activity log ID
- `timestamp` - When the content was added (ISO 8601)
- `job_type` - Which job added it
- `media_type` - `movie` or `show`
- `title` - Content title
- `year` - Release year
- `tmdb_id` - TMDB ID (movies and shows)
- `tvdb_id` - TVDB ID (shows only)
- `imdb_id` - IMDB ID
- `poster_url` - Link to content page
- `status` - `added` (direct mode) or `requested` (Jellyseerr mode)

## Get Activity Statistics

Get statistics about content added over different time periods.

**Endpoint:** `GET /v1/activity/stats`

**Example Request:**

```bash
curl http://localhost:9090/v1/activity/stats
```

**Example Response:**

```json
{
  "total_items": 1523,
  "movies_added": 892,
  "shows_added": 631,
  "last_24h": 45,
  "last_7d": 312,
  "last_30d": 891,
  "by_job": {
    "trending_movies": 234,
    "popular_movies": 189,
    "smart_popular_movies": 156,
    "trending_shows": 178,
    "popular_shows": 142
  }
}
```

**Response Fields:**
- `total_items` - Total content added all-time
- `movies_added` - Total movies added
- `shows_added` - Total shows added
- `last_24h` - Items added in last 24 hours
- `last_7d` - Items added in last 7 days
- `last_30d` - Items added in last 30 days
- `by_job` - Breakdown by job type

## Clear Activity Logs

Delete activity logs older than specified days.

**Endpoint:** `DELETE /v1/activity/logs`

**Query Parameters:**
- `days` (required) - Delete logs older than N days

**Example Request:**

```bash
# Delete logs older than 30 days
curl -X DELETE "http://localhost:9090/v1/activity/logs?days=30"
```

**Example Response:**

```json
{
  "message": "Deleted 234 activity logs older than 30 days",
  "deleted_count": 234
}
```

## Add Media Anyway (Manual Override)

Manually add media that was rejected by filters. This allows you to override filter decisions for specific items shown in the activity log.

**Endpoint:** `POST /v1/activity/:id/add-anyway`

**Path Parameters:**
- `id` (required) - Activity log entry ID

**Behavior:**
- In `jellyseerr` mode: Creates a request in Jellyseerr
- In `direct` mode: Adds directly to Radarr (movies) or Sonarr (shows)
- Updates the activity log status to "added" with message "Manually added by user"

**Example Request:**

```bash
# Manually add a rejected movie (using activity log ID)
curl -X POST "http://localhost:9090/v1/activity/123/add-anyway"
```

**Example Response:**

```json
{
  "success": true,
  "message": "Stranger Things has been added successfully"
}
```

**Error Responses:**

```json
{
  "error": "Activity log not found"
}
```

```json
{
  "error": "Failed to add media: no TMDB ID available for movie"
}
```

**Use Case:**

This endpoint is useful when a filter incorrectly rejects content. For example, if "Stranger Things" is rejected because it contains the "horror" genre, but you want to add it anyway:

1. View the rejected item in the Activity Log UI
2. Click the "Add Anyway" button (available for `rejected` status items)
3. The content is added to Jellyseerr or *arr based on your mode
4. The activity log updates in place to show "Added" status

## Advanced Examples

### Get Today's Activity

```bash
curl -s "http://localhost:9090/v1/activity/logs?limit=500" \
  | jq '.logs[] | select(.timestamp >= "'$(date -u -d '24 hours ago' +%Y-%m-%dT%H:%M:%S)'Z")'
```

### Count by Media Type

```bash
curl -s "http://localhost:9090/v1/activity/stats" \
  | jq '{movies: .movies_added, shows: .shows_added}'
```

### Get Failed Adds

```bash
curl -s "http://localhost:9090/v1/activity/logs?status=failed&limit=100" \
  | jq '.logs[] | {title, year, job_type}'
```

### Export to CSV

```bash
curl -s "http://localhost:9090/v1/activity/logs?limit=500" \
  | jq -r '.logs[] | [.timestamp, .title, .year, .media_type, .job_type] | @csv' \
  > activity.csv
```

## Error Responses

**Invalid Limit (400):**
```json
{
  "error": "Limit must be between 1 and 500"
}
```

**Invalid Media Type (400):**
```json
{
  "error": "Invalid media_type. Must be 'movie' or 'show'"
}
```

**Missing Required Parameter (400):**
```json
{
  "error": "Parameter 'days' is required"
}
```

## Next Steps

- Explore [Jobs API](/blockbusterr/api/jobs/)
- Review [Configuration API](/blockbusterr/api/config/)
- Learn about [Integration Modes](/blockbusterr/concepts/integration-modes/)
