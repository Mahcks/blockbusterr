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
  enabled: true
  url: "http://radarr:7878"
  api_key: "your_radarr_api_key"
  quality_profile_id: 1
  root_folder: "/movies"
  search_on_add: true
  monitored: true
  minimum_availability: "released"
```

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
