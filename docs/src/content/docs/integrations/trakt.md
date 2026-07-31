---
title: Trakt Integration
description: Configure Trakt for fetching trending and popular content
---

Trakt is one supported discovery source, providing trending, popular, and Trakt-specific list data.

## Overview

Trakt is optional. Existing Trakt jobs remain supported, while dynamic jobs can use TMDB or Simkl where the selected job type supports them.

## Getting Trakt Credentials

### Step 1: Create a Trakt Application

1. Go to [trakt.tv/oauth/applications](https://trakt.tv/oauth/applications)
2. Click **"New Application"**
3. Fill out the form:
   - **Name**: Blockbusterr
   - **Description**: Media automation tool
   - **Redirect URI**: `http://localhost:9090/auth/trakt/callback`
   - **Permissions**: Check all boxes
4. Click **"Save App"**

### Step 2: Copy Credentials

After creating the application, you'll receive:
- **Client ID** - Copy this
- **Client Secret** - Copy this

### Step 3: Configure Blockbusterr

**Via Configuration File:**

```yaml
trakt:
  client_id: "your_client_id_here"
  client_secret: "your_client_secret_here"
```

**Via Web UI:**

1. Open `http://localhost:9090`
2. Go to **Configuration** tab
3. Find **Trakt** section
4. Paste Client ID and Client Secret
5. Click **Save**
6. Click **"Authorize with Trakt"** to complete OAuth flow

### Step 4: Authorize

Complete the OAuth authorization:

1. Click **"Authorize with Trakt"** in the UI
2. You'll be redirected to Trakt
3. Log in and approve the authorization
4. You'll be redirected back to Blockbusterr

The `access_token` and `refresh_token` will be automatically saved.

## Configuration

```yaml
trakt:
  client_id: "your_client_id"
  client_secret: "your_client_secret"
  access_token: ""     # Auto-populated after OAuth
  refresh_token: ""    # Auto-populated after OAuth
```

## Testing

Verify Trakt integration:

```bash
curl "http://localhost:9090/v1/trakt/trending/movies?limit=5"
```

You should see trending movies data.

## Troubleshooting

**"Invalid client credentials"**
- Verify Client ID and Client Secret are correct
- Ensure no extra spaces

**"Access token expired"**
- Blockbusterr automatically refreshes tokens
- If issues persist, re-authorize through the UI

**"Rate limit exceeded"**
- Trakt has API rate limits
- Reduce job frequency if hitting limits

## Next Steps

- Configure [Radarr integration](/integrations/radarr/)
- Configure [Sonarr integration](/integrations/sonarr/)
- Set up your [first job](/getting-started/quickstart/)
