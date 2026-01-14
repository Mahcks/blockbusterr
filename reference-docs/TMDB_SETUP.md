 fet # TMDB API Integration for Poster Images

## Overview
The job preview feature supports high-quality poster images from The Movie Database (TMDB). **The TMDB API key is completely optional** - without it, you'll see a clean list view instead of poster images.

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

1. **Create a TMDB Account**
   - Go to https://www.themoviedb.org/signup
   - Sign up for a free account

2. **Request an API Key**
   - Once logged in, go to https://www.themoviedb.org/settings/api
   - Click "Request an API Key"
   - Select "Developer" when asked for the type of API key
   - Fill out the form with your application details:
     - **Application Name**: Blockbusterr (or your preferred name)
     - **Application URL**: Can use localhost or your deployment URL
     - **Application Summary**: Brief description (e.g., "Personal media automation tool")
   - Accept the terms and submit

3. **Copy Your API Key**
   - Once approved (usually instant), you'll receive an API Key (v3 auth)
   - Copy this key

4. **Add to Configuration**
   - Open `config/config.yaml`
   - Find the `tmdb` section:
     ```yaml
     tmdb:
       api_key: "YOUR_API_KEY_HERE"
     ```
   - Replace `"YOUR_API_KEY_HERE"` with your actual API key
   - Save the file and restart the application

## Features

With TMDB integration enabled, the job preview will:
- Display actual movie and TV show posters
- Show high-quality images (500px width)
- Automatically fall back to styled placeholders if no poster is available
- Cache poster URLs for faster loading

## Troubleshooting

**Posters not loading?**
- Verify your API key is correct in `config.yaml`
- Check that the TMDB service is accessible from your network
- Look for TMDB-related errors in the application logs
- The system will gracefully fall back to a list view if TMDB is unavailable

**Prefer list view over posters?**
- Simply remove or leave the `api_key` empty in your configuration
- The application will automatically switch to list view mode
- All functionality remains the same, just without poster images
