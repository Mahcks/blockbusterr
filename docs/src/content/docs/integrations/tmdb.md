---
title: TMDB Integration
description: Configure TMDB discovery and poster images
---

import { Aside } from '@astrojs/starlight/components';

TMDB provides trending and popular discovery jobs plus poster images. It is optional when another discovery source is configured.

TMDB also supports public mixed-media lists and the connected account's movie
or TV watchlist. Save the API key first, then use **Connect account** in
Settings when a personal watchlist job is needed. Public lists do not require
account authorization. Session credentials stay in the private configuration
and are excluded from shareable exports.

## Display Modes

### Without TMDB (List View)

When no TMDB API key is configured, the preview shows a simple, clean list with:
- Title, year, and rating
- Popularity metrics
- Genre information
- Status badges (NEW/EXISTS/FILTERED)

### With TMDB (Poster Grid View)

When TMDB is configured, you get beautiful poster images:
- High-quality movie and TV show posters (500px width)
- Visual grid layout
- Hover effects and details
- Automatic fallback to styled placeholders if a poster is unavailable

## Getting a TMDB API Key

### Step 1: Create a TMDB Account

1. Go to [themoviedb.org/signup](https://www.themoviedb.org/signup)
2. Sign up for a free account
3. Verify your email address

### Step 2: Request an API Key

1. Once logged in, go to [themoviedb.org/settings/api](https://www.themoviedb.org/settings/api)
2. Click **"Request an API Key"**
3. Select **"Developer"** when asked for the type of API key
4. Fill out the application form:
   - **Application Name**: Blockbusterr (or your preferred name)
   - **Application URL**: Can use `http://localhost:9090` or your deployment URL
   - **Application Summary**: Brief description (e.g., "Personal media automation tool")
5. Accept the terms and submit

<Aside type="tip">
  API keys are usually approved instantly for Developer accounts.
</Aside>

### Step 3: Copy Your API Key

1. Once approved, you'll receive an **API Key (v3 auth)**
2. Copy this key - you'll need it for configuration

### Step 4: Add to Configuration

Add the API key to your Blockbusterr configuration:

**Via Configuration File:**

Edit `config/config.yaml`:

```yaml
tmdb:
  api_key: "your_api_key_here"
```

**Via Web UI:**

1. Open `http://localhost:9090`
2. Go to **Configuration** tab
3. Find the **TMDB** section
4. Paste your API key
5. Click **Save**

**Via Environment Variable:**

```bash
TMDB_API_KEY=your_api_key_here
```

### Step 5: Restart Blockbusterr

Restart the application to apply the changes:

```bash
docker restart blockbusterr
```

## Verifying TMDB Integration

### Test the Preview

1. Go to the **Jobs** tab
2. Click **Preview** on any job
3. If configured correctly, you'll see poster images
4. If not configured, you'll see a clean list view (both work fine!)

### Check Logs

Look for TMDB initialization messages:

```bash
docker logs blockbusterr | grep -i tmdb
```

You should see:
```
INFO TMDB API key configured - poster images enabled
```

## Features

With TMDB integration enabled:

- **High-Quality Posters**: 500px width images from TMDB
- **Visual Grid**: Clean poster grid layout with hover effects
- **Automatic Caching**: Poster URLs cached for faster loading
- **Graceful Fallback**: Styled placeholders if poster unavailable
- **No Performance Impact**: Fetched asynchronously, doesn't slow down jobs

## Troubleshooting

### Posters Not Loading?

**Check API Key:**
- Verify it's correctly entered in configuration
- Ensure no extra spaces or quotes
- Try copying the key again from TMDB

**Check Network Access:**
```bash
docker exec blockbusterr curl -s "https://api.themoviedb.org/3/configuration?api_key=YOUR_KEY"
```

Should return JSON configuration data.

**Check Logs:**
```bash
docker logs blockbusterr | grep -i tmdb
```

Look for error messages like:
- "Invalid API key"
- "TMDB API request failed"
- "Rate limit exceeded"

### Rate Limiting

TMDB has rate limits:
- **40 requests per 10 seconds**
- **1000 requests per day** (for free accounts)

Blockbusterr respects these limits and caches aggressively to minimize API calls.

If you hit rate limits:
- Reduce preview frequency
- Consider upgrading to TMDB paid tier (not usually necessary)

### Prefer List View?

If you prefer the list view over posters:

1. Remove the API key from configuration
2. Or set it to an empty string:
```yaml
tmdb:
  api_key: ""
```

The application automatically switches to list view mode when TMDB is not configured.

## Configuration Options

```yaml
tmdb:
  api_key: "your_api_key_here"  # TMDB API key (optional)
```

<Aside type="note">
  A TMDB key is required only for jobs whose discovery source is set to TMDB.
</Aside>

## Privacy & Data

### What Data is Sent?

When TMDB is configured, Blockbusterr sends:
- TMDB IDs (for movies)
- TVDB IDs (for TV shows)
- Your API key (authentication)

### What Data is Received?

Blockbusterr receives:
- Poster image URLs
- Backdrop image URLs (not currently used)

### Data Storage

- Poster URLs are cached in the local database
- No personal data is sent to TMDB
- TMDB cannot see what you're adding to your library

## Alternative: No TMDB

You don't need TMDB to use Blockbusterr. The list view provides all the same functionality:

**List View Includes:**
- ✅ Title and year
- ✅ Rating and votes
- ✅ Popularity metrics
- ✅ Genres
- ✅ Runtime
- ✅ Overview/description
- ✅ Status badges (Will Add / Exists / Filtered)
- ✅ Filter reasons

The only difference is visual - no poster images.

## TMDB API Documentation

For more information about TMDB API:
- [TMDB API Documentation](https://developers.themoviedb.org/3)
- [TMDB API Settings](https://www.themoviedb.org/settings/api)
- [TMDB Terms of Use](https://www.themoviedb.org/terms-of-use)

## Next Steps

- Learn about [Job Previews](/concepts/jobs/#preview-before-enabling)
- Configure your [first job](/getting-started/quickstart/#step-3-configure-your-first-job)
- Explore [other integrations](/integrations/trakt/)
