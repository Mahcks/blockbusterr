---
title: Letterboxd
description: Use opt-in public Letterboxd lists and watchlists, with clear limits on reliability and metadata.
---

Blockbusterr can read **public Letterboxd movie lists and public watchlists** through an opt-in experimental scraper. It does not use your Letterboxd login and cannot access private lists.

Choose this only if you want to follow a public Letterboxd source and accept that page changes or blocking can interrupt discovery. For a provider API workflow, use [MDBList](/integrations/mdblist/) and its supported external-list import instead.

Letterboxd has an API, but access is selective; its [current access policy](https://letterboxd.com/api-beta/access/) excludes private or personal projects. Do not spend time looking for a generally available personal API key for this integration.

## Set up a public list

1. Start Blockbusterr with `BLOCKBUSTERR_DRY_RUN=true` before creating a custom job; see the [quickstart](/getting-started/quickstart/) for how to set this environment variable and recreate your container with its existing data mount. Custom jobs are created enabled.
2. Open **Settings → Connections → Letterboxd**.
3. Select **Enable experimental scraping**, then **Save changes**.
4. Open **Jobs → New job → Custom Job** and choose **List or Watchlist**, **Movie**, and **Letterboxd**.
5. Choose **Public list** and enter the account name in **Owner or member** and the list slug in **List ID or slug**.
6. Select **Check source**. Save the job and review its preview while dry run remains on.
7. Check the rule set and delivery cap before allowing delivery. Turning off global dry run affects all enabled jobs.

For the example URL `https://letterboxd.com/alex/list/weekend-movies/`, enter owner `alex` and slug `weekend-movies`. These are format examples, not a real recommended list.

For a public watchlist such as `https://letterboxd.com/alex/watchlist/`, choose **Personal watchlist**, enter owner `alex`, and leave the list ID empty. The UI's “Personal” label does not grant access to private content.

## Metadata and rules

The scraper uses embedded TMDB identities to identify films; it does not guess matches from titles. It imports limited candidate data, such as title, year where available, and TMDB ID. It does not provide ordinary ratings, vote counts, or genre fields for your rules.

Use a dedicated rule set suited to this limited metadata. Do not apply a minimum rating, minimum votes, or required genre unless you have verified the needed data in the actual preview. A TMDB connection does not automatically turn a Letterboxd list into a fully enriched TMDB discovery feed.

Removing an item from the Letterboxd list does not remove media already delivered to Radarr or the request server.

## Troubleshooting

- **Letterboxd is absent from the selector:** save the experimental setting and select a movie List or Watchlist job.
- **Private, missing, or blocked source:** open it while signed out of Letterboxd. The integration cannot use your account to bypass access restrictions.
- **Source used to work:** the website may have changed or blocked scraping. Pause the job and consider MDBList's import support.
- **Missing identity error:** the page does not provide a usable TMDB ID. Blockbusterr fails instead of guessing another movie.
- **Every result is filtered:** inspect the assigned rule set for metadata that this source does not supply.
