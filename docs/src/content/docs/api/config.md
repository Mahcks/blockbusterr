---
title: Configuration API
description: API endpoints for testing integrations and fetching metadata
---

The Configuration API allows you to test connections and retrieve metadata for configuring filters.

## Validate Radarr Connection

Test connection to Radarr instance.

**Endpoint:** `GET /v1/radarr/validate`

**Query Parameters:**
- `test` (optional) - Set to `true` to test custom credentials
- `url` (required if test=true) - Radarr URL
- `api_key` (required if test=true) - Radarr API key

**Example Requests:**

```bash
# Test configured Radarr
curl "http://localhost:9090/v1/radarr/validate"

# Test custom credentials
curl "http://localhost:9090/v1/radarr/validate?test=true&url=http://radarr:7878&api_key=YOUR_KEY"
```

**Example Response:**

```json
{
  "valid": true,
  "message": "Connection successful",
  "version": "4.7.5"
}
```

## Get Radarr Quality Profiles

Retrieve available quality profiles from Radarr.

**Endpoint:** `GET /v1/radarr/quality-profiles`

**Example Request:**

```bash
curl "http://localhost:9090/v1/radarr/quality-profiles"
```

**Example Response:**

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

## Get Radarr Root Folders

Retrieve available root folders from Radarr.

**Endpoint:** `GET /v1/radarr/root-folders`

**Example Request:**

```bash
curl "http://localhost:9090/v1/radarr/root-folders"
```

**Example Response:**

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

## Sonarr Endpoints

Similar endpoints are available for Sonarr:

- `GET /v1/sonarr/validate`
- `GET /v1/sonarr/quality-profiles`
- `GET /v1/sonarr/root-folders`

Usage is identical to Radarr endpoints.

## Get Trakt Metadata

Retrieve metadata for configuring filters.

### Get Movie Genres

**Endpoint:** `GET /v1/trakt/genres/movies`

**Example Response:**

```json
{
  "data": [
    {"slug": "action", "name": "Action"},
    {"slug": "adventure", "name": "Adventure"},
    {"slug": "animation", "name": "Animation"}
  ]
}
```

### Get TV Show Genres

**Endpoint:** `GET /v1/trakt/genres/shows`

### Get Countries

**Endpoints:**
- `GET /v1/trakt/countries/movies`
- `GET /v1/trakt/countries/shows`

**Example Response:**

```json
{
  "data": [
    {"code": "us", "name": "United States"},
    {"code": "gb", "name": "United Kingdom"}
  ]
}
```

### Get Languages

**Endpoints:**
- `GET /v1/trakt/languages/movies`
- `GET /v1/trakt/languages/shows`

### Get Networks

**Endpoint:** `GET /v1/trakt/networks`

**Example Response:**

```json
{
  "data": [
    {"slug": "hbo", "name": "HBO"},
    {"slug": "netflix", "name": "Netflix"}
  ]
}
```

## Get Trakt Content

Fetch content directly from Trakt.

### Trending Content

**Endpoints:**
- `GET /v1/trakt/trending/movies?limit=10`
- `GET /v1/trakt/trending/shows?limit=10`

**Query Parameters:**
- `limit` (optional) - Number of items (default: 10)

### Popular Content

**Endpoints:**
- `GET /v1/trakt/popular/movies?limit=10`
- `GET /v1/trakt/popular/shows?limit=10`

## Error Responses

**Connection Failed (500):**
```json
{
  "error": "Failed to connect to Radarr: connection refused"
}
```

**Invalid API Key (401):**
```json
{
  "error": "Invalid API key"
}
```

**Not Configured (400):**
```json
{
  "error": "Radarr is not configured"
}
```

## Next Steps

- Explore [Jobs API](/blockbusterr/api/jobs/)
- Check [Activity API](/blockbusterr/api/activity/)
- Learn about [Configuration](/blockbusterr/getting-started/configuration/)
