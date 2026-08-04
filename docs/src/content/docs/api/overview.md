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
| 409 | Conflict - Stale revision or resource still in use |
| 500 | Internal Server Error |

## Authentication

Blockbusterr does not provide built-in authentication. It is designed for a trusted LAN and its API can change jobs, trigger downloads, and export configuration.

:::caution
Do not expose port 9090 directly to the public internet. If remote access is required, place Blockbusterr behind an authenticated reverse proxy or VPN.
:::

## Rate Limiting

The HTTP API does not impose request rate limiting. Delivery budgets configured for jobs still apply to successful automated additions and requests.

## API Sections

### [Jobs API](/api/jobs/)

Manage and trigger jobs, preview content before adding.

- Preview jobs
- Trigger jobs manually
- Get job status

### [Activity API](/api/activity/)

View and manage Activity Entries and Job Runs.

- Query title-level Activity Entries
- Inspect aggregate Job Runs
- Get activity statistics
- Clear old logs

### [Rules API](/api/rules/)

Manage reusable movie/show rules and universal title exceptions.

- Create and revise rule sets
- Inspect assignment counts
- Create job-specific rule copies

### [Configuration API](/api/config/)

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

- Explore [Jobs API](/api/jobs/)
- Manage [Rules API](/api/rules/)
- Check [Activity API](/api/activity/)
- Review [Configuration API](/api/config/)
