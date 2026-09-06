---
title: MDBList
description: Discover movies and shows from MDBList lists and your watchlist.
---

MDBList is an optional list provider. It supports public user lists, official
lists, numeric list IDs, and the API-key owner's watchlist. It is also the
recommended bridge for lists imported from Letterboxd or IMDb.

## Configure MDBList

1. Create a free API key in [MDBList Preferences](https://mdblist.com/preferences/).
2. Open **Settings → Connections → MDBList**.
3. Save the API key and select **Test connection**.
4. Create a **List or Watchlist** job and select MDBList.
5. Use **Check source**, preview the job, then enable it.

For a public user list, enter its username as the owner and its list name or
slug as the identifier. For a numeric list, enter only the list ID. For an
official list, enter its official slug without an owner. A watchlist needs no
owner or list ID because it belongs to the API-key account.

The free plan currently includes 1,000 API requests per day. Blockbusterr uses
cursor pagination and only fetches lists when previewing or running a job.

## Letterboxd

Letterboxd API access is selective and is not currently granted for private or
personal projects. Blockbusterr offers opt-in experimental scraping for public
lists and public watchlists, but it may be blocked or stop working without
notice. It uses Letterboxd's embedded TMDB links and refuses title-only guesses.
Private content is unsupported. Import the Letterboxd list into MDBList when
reliability matters.

Source-list synchronization is add-only. Removing an item from MDBList does not
delete media already delivered to Radarr, Sonarr, Jellyseerr, or Seerr.
