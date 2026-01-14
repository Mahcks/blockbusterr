---
title: Sonarr Integration
description: Configure Sonarr for automatic TV show management
---

Sonarr integration allows Blockbusterr to automatically add TV shows to your library.

## Prerequisites

- Running Sonarr instance (v3 or v4)
- Sonarr API key
- Network access from Blockbusterr to Sonarr

## Configuration

```yaml
sonarr:
  enabled: true
  url: "http://sonarr:8989"
  api_key: "your_sonarr_api_key"
  quality_profile_id: 1
  root_folder: "/tv"
  search_on_add: true
  monitored: true
  season_folder: true
  series_type: "standard"
```

## Configuration Options

| Option | Values | Description |
|--------|--------|-------------|
| `series_type` | `standard`, `daily`, `anime` | Type of series |
| `season_folder` | `true`, `false` | Use season subdirectories |
| `monitored` | `true`, `false` | Monitor new series |
| `search_on_add` | `true`, `false` | Search immediately |

## Getting API Key

1. Open Sonarr web interface
2. Go to **Settings** → **General**
3. Scroll to **Security** section
4. Copy the **API Key**

## Testing Connection

```bash
curl "http://localhost:9090/v1/sonarr/validate"
```

## Docker Networking

Use container names:

```yaml
sonarr:
  url: "http://sonarr:8989"  # Container name
```

## Next Steps

- Configure [Radarr](/blockbusterr/integrations/radarr/) for movies
- Set up [TV show jobs](/blockbusterr/concepts/jobs/)
- Learn about [integration modes](/blockbusterr/concepts/integration-modes/)
