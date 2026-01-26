---
title: Radarr Integration
description: Configure Radarr for automatic movie management
---

Radarr integration allows Blockbusterr to automatically add movies to your library.

## Prerequisites

- Running Radarr instance (v3 or v4)
- Radarr API key
- Network access from Blockbusterr to Radarr

## Configuration

```yaml
radarr:
  url: "http://radarr:7878"
  api_key: "your_radarr_api_key"
  quality_profile: 1
  root_folder: "/movies"
  minimum_availability: "released"
  monitor: "movieOnly"
```

## Monitor Options

The `monitor` setting controls how movies are monitored when added to Radarr.

| Value | Description |
|-------|-------------|
| `movieOnly` | Monitor only the movie itself (default) |
| `movieAndCollection` | Monitor the movie and its entire collection (e.g., all MCU movies) |
| `none` | Add the movie but don't monitor it for downloads |

:::tip
Use `movieAndCollection` if you want Radarr to automatically grab other movies in a franchise when you add one from the collection.
:::

## Minimum Availability Options

The `minimum_availability` setting determines when Radarr considers a movie available for download.

| Value | Description |
|-------|-------------|
| `announced` | As soon as the movie is announced |
| `inCinemas` | When the movie is released in theaters |
| `released` | When the movie is released on physical/digital media (default, recommended) |

:::caution
Setting `announced` or `inCinemas` may result in lower quality releases or CAM rips. Use `released` for best quality.
:::

## Getting API Key

1. Open Radarr web interface
2. Go to **Settings** → **General**
3. Scroll to **Security** section
4. Copy the **API Key**

## Finding Quality Profile ID

**Via API:**
```bash
curl "http://localhost:9090/v1/radarr/quality-profiles"
```

**Via Radarr UI:**
1. Go to **Settings** → **Profiles**
2. Note the profile name you want to use
3. The ID corresponds to the order (1, 2, 3, etc.)

## Finding Root Folder

**Via API:**
```bash
curl "http://localhost:9090/v1/radarr/root-folders"
```

**Via Radarr UI:**
1. Go to **Settings** → **Media Management**
2. Check **Root Folders** section

## Testing Connection

```bash
curl "http://localhost:9090/v1/radarr/validate"
```

## Docker Networking

If using Docker, use container names:

```yaml
radarr:
  url: "http://radarr:7878"  # Container name, not localhost
```

Ensure both containers are on the same Docker network.

## Next Steps

- Configure [Sonarr](/blockbusterr/integrations/sonarr/) for TV shows
- Set up [movie jobs](/blockbusterr/concepts/jobs/)
- Learn about [integration modes](/blockbusterr/concepts/integration-modes/)
