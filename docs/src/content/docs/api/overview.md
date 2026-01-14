---
title: API Overview
description: RESTful API for managing jobs and viewing activity
---

Blockbusterr provides a comprehensive RESTful API for managing jobs, viewing activity, and previewing content.

## Base URL

```
http://localhost:9090/v1
```

Replace `localhost:9090` with your Blockbusterr instance URL.

## Response Format

All endpoints return JSON unless otherwise specified.

**Success Response:**
```json
{
  "data": { ... },
  "count": 10
}
```

**Error Response:**
```json
{
  "error": "Error message here"
}
```

## Status Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 400 | Bad Request - Invalid parameters |
| 404 | Not Found - Resource doesn't exist |
| 500 | Internal Server Error |

## Authentication

Currently, no authentication is required for API access.

:::caution
It's recommended to run Blockbusterr behind a reverse proxy with authentication if exposing to the internet.
:::

## Rate Limiting

Currently, no rate limiting is enforced. Please use the API responsibly.

## API Sections

### [Jobs API](/blockbusterr/api/jobs/)

Manage and trigger jobs, preview content before adding.

- Preview jobs
- Trigger jobs manually
- Get job status

### [Activity API](/blockbusterr/api/activity/)

View and manage activity logs.

- Get activity logs
- Get activity statistics
- Clear old logs

### [Configuration API](/blockbusterr/api/config/)

Manage integrations and test connections.

- Validate Radarr/Sonarr connections
- Get quality profiles and root folders
- Fetch metadata (genres, countries, languages)

## Quick Examples

### Preview a Job

```bash
curl -X POST http://localhost:9090/v1/jobs/preview/trending-movies
```

### Trigger a Job

```bash
curl -X POST http://localhost:9090/v1/jobs/trigger/popular-shows
```

### Get Recent Activity

```bash
curl "http://localhost:9090/v1/activity/logs?limit=50"
```

### Get Trending Movies from Trakt

```bash
curl "http://localhost:9090/v1/trakt/trending/movies?limit=10"
```

## Next Steps

- Explore [Jobs API](/blockbusterr/api/jobs/)
- Check [Activity API](/blockbusterr/api/activity/)
- Review [Configuration API](/blockbusterr/api/config/)
