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
  url: "http://sonarr:8989"
  api_key: "your_sonarr_api_key"
  quality_profile: 1
  root_folder: "/tv"
  monitor: "all"
```

## Monitor Options

The `monitor` setting controls which episodes Sonarr will monitor and download when adding a new series.

| Value | Description |
|-------|-------------|
| `all` | Monitor all episodes (default) |
| `future` | Only monitor future episodes (not yet aired) |
| `missing` | Monitor episodes that are missing from your library |
| `existing` | Monitor episodes you already have |
| `firstSeason` | Only monitor the first season |
| `lastSeason` | Only monitor the last season |
| `latestSeason` | Only monitor the latest/current season |
| `pilot` | Only monitor the pilot episode |
| `recent` | Monitor recent episodes |
| `monitorSpecials` | Include special episodes in monitoring |
| `unmonitorSpecials` | Exclude special episodes from monitoring |
| `none` | Add the series but don't monitor any episodes |

:::tip
For new shows still airing, use `future` or `latestSeason` to avoid downloading older seasons you may not want.
:::

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

- Configure [Radarr](/integrations/radarr/) for movies
- Set up [TV show jobs](/concepts/jobs/)
- Learn about [integration modes](/concepts/integration-modes/)
