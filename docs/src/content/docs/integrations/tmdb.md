---
title: TMDB
description: Start with free movie and show discovery, connect a watchlist, and understand which ratings your rules use.
---

TMDB (The Movie Database) is a good first discovery provider: it finds trending and popular movies and shows, supplies posters, and supports public lists, personal watchlists, and recommendations. You can run Blockbusterr with TMDB and Radarr alone for movies, or TMDB and Sonarr for shows.

TMDB's API is [free for noncommercial use with attribution](https://developer.themoviedb.org/docs/faq). You do not need Trakt or an MDBList subscription for TMDB jobs. Other providers remain optional.

## Connect TMDB

Start Blockbusterr with `BLOCKBUSTERR_DRY_RUN=true` before creating custom jobs or changing delivery; follow the [quickstart dry-run setup](/getting-started/quickstart/). Custom jobs are created enabled; recipes start disabled. Keep dry run on while reviewing sources and previews, then review all enabled jobs before turning it off.

1. Create an account at [TMDB](https://www.themoviedb.org/).
2. On a desktop browser, open [account API settings](https://www.themoviedb.org/settings/api) and apply for a developer API key for your personal use. Describe what you are actually building or running.
3. Copy the **API key**, rather than the longer API Read Access Token. Blockbusterr's current integration uses the key.
4. In Blockbusterr, open **Settings → Connections → TMDB** and paste it into **API key**.
5. Select **Save changes**.
6. Open **Jobs → New job** and choose a TMDB recipe, such as **Balanced Trending** for movies. Review its preview before enabling it.

TMDB does not have a separate **Test connection** button in Settings. A successful job preview confirms that Blockbusterr can fetch its discovery data.

The API key field appears blank after saving because saved credentials are hidden. Leaving it blank keeps the existing key.

## Choose a discovery method

| Method | Use it for | What it does not promise |
| --- | --- | --- |
| Trending | Titles receiving attention now | Box-office revenue or an IMDb popularity ranking |
| Popular | A broader pool of popular titles | Every title in a genre or a complete catalog scan |
| Smart Popular | Popular titles evaluated with an adaptive rating threshold | A critics-only score |
| List or Watchlist | An explicit collection you want to follow | Automatic access to a private list without authorization |
| Recommendations | Suggestions from seed movies or shows you choose | Recursive discovery from every new suggestion |

The **Science-Fiction Discovery** recipe starts a TMDB movie job with genre rules. Review its rating and vote requirements for your taste. These rules filter the candidates fetched by the job; they do not turn the job into a search of every science-fiction movie on TMDB. Increase the content limit or use a dedicated list when you need more coverage.

For recommendations, use up to 20 explicit TMDB seed IDs or a supported provider seed list. Blockbusterr fetches one round of recommendations and excludes the seed titles themselves. See [Jobs](/concepts/jobs/).

## Follow a public list

For a TMDB list URL such as `https://www.themoviedb.org/list/12345`, the identifier is `12345`. This is an example of the URL format, not a recommended list.

1. Create a **Custom Job** with type **List or Watchlist**, source **TMDB**, and the media type you want.
2. Choose **Public list**.
3. Leave **Owner or member** empty and enter the numeric identifier in **List ID or slug**.
4. Select **Check source** to confirm access and inspect a sample.
5. Assign a rule set, save the job, and preview it while dry run is on. Check its delivery settings before allowing live delivery.

A mixed movie/show list needs separate movie and show jobs if you want both kinds delivered.

## Connect your personal watchlist

1. Save your TMDB API key first.
2. In **Settings → Connections → TMDB**, select **Connect account for watchlists**.
3. Approve access in the TMDB tab.
4. Return to Blockbusterr and select **Complete connection**.
5. Create a **List or Watchlist** job, choose **TMDB → Personal watchlist**, and leave the owner and list identifier empty. You can also start with the **Personal Watchlist** recipe.
6. Check the source and preview while dry run remains on. Review its cap and destination before allowing delivery.

Connecting an account is optional for trending, popular, and public-list jobs. Removing an item from your watchlist does not remove it from Radarr or Sonarr after delivery.

## Understand the metadata

TMDB job ratings and vote counts are **TMDB user ratings and votes**. A minimum rating of `5` is a TMDB threshold, not IMDb 5/10. TMDB movie genres include `science-fiction`; TV uses the broader `sci-fi-fantasy` genre.

Adding a TMDB key also enables poster lookup and specific metadata enrichment used by features such as regional certifications and show identity resolution. It does not populate every missing field from other list providers. In particular, do not expect an MDBList job to gain rating, vote, or genre metadata merely because TMDB is configured.

## Troubleshooting

| Problem | What to check |
| --- | --- |
| TMDB is absent from the source selector | Save the API key, then select a job type supported by TMDB. Box office and anticipated are Trakt-specific. |
| Preview reports an authentication error | Use the API key, remove accidental whitespace, and confirm it is still valid in TMDB settings. |
| Watchlist access fails | Complete account authorization; a saved API key alone is not a connected account. |
| A show cannot be added to Sonarr | Check the preview's identity error. Direct Sonarr delivery requires a verified TVDB identity, and some titles cannot be mapped. |
| Few titles pass | Inspect rejection reasons and the content limit. A strict rule cannot discover a title outside the fetched candidates. |
| Posters or requests time out | Check that the Blockbusterr host/container can reach TMDB and its image servers. |

Next, connect [Radarr](/integrations/radarr/) or [Sonarr](/integrations/sonarr/) and follow the [quickstart](/getting-started/quickstart/).
