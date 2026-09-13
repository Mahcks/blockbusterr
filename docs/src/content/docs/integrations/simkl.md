---
title: Simkl
description: Configure optional Simkl discovery and understand its feed sizes, periods, and ratings.
---

Simkl is an alternative discovery source for movie and show **Trending**, **Popular**, **Smart Popular**, and **Watched** jobs. Use it when you want another source of audience interest alongside, or instead of, TMDB. You do not need Trakt for Simkl jobs.

Blockbusterr's Simkl integration reads discovery feeds. It does not currently import your Simkl watchlist, synchronize your watch history, or support Simkl as a list source.

## Connect Simkl

Start Blockbusterr with `BLOCKBUSTERR_DRY_RUN=true` before creating custom jobs or changing delivery; follow the [quickstart dry-run setup](/getting-started/quickstart/). Custom jobs are created enabled; recipes start disabled. Keep dry run on while reviewing sources and previews, then review all enabled jobs before turning it off.

1. Sign in to Simkl and open [Developer Settings](https://simkl.com/settings/developer/).
2. Create an application and copy its **client ID**.
3. Open **Settings → Connections → Simkl** in Blockbusterr.
4. Enter the client ID and select **Save changes**.
5. Open **Jobs → New job → Custom Job**. Choose movies or shows, a supported discovery type, and **Simkl** as the source.
6. Assign a rule set and review the preview while dry run remains on. Review its cap and destination before allowing delivery.

There is no separate Simkl **Test connection** button in Settings. Use a preview to verify discovery. Review [Simkl's API rules](https://api.simkl.org/api-rules) for attribution and usage requirements; Blockbusterr identifies Simkl-sourced results in the app.

## What the jobs fetch

| Job | Feed used by Blockbusterr |
| --- | --- |
| Trending | Simkl's current trending feed |
| Popular and Smart Popular | Monthly feed |
| Watched | Weekly or monthly feed, according to the chosen period |

Feeds return at most **500 candidates**. Increasing a job's content limit beyond that does not create a full-catalog search. Smart Popular changes how candidates are evaluated; it does not expand Simkl's underlying feed.

For example, create a weekly **Watched** movie job with a content limit of `50`, a rating threshold suited to your taste, and a small job delivery cap. Preview it to see which movies pass before allowing direct delivery to Radarr.

## Ratings and metadata

Blockbusterr uses a Simkl rating when the feed supplies one and otherwise falls back to its IMDb rating entry. Do not describe the ordinary rating filter as a guaranteed IMDb-only threshold for every Simkl title.

Available genres, runtime, language, country, and IDs depend on the returned items. Missing metadata can make a candidate fail a rule. Inspect preview reasons before relaxing rules, especially if the rule exists to enforce a firm preference.

## Troubleshooting

- **Simkl is missing from sources:** save its client ID and choose Trending, Popular, Smart Popular, or Watched. It is not a source for Box office, Anticipated, or List or Watchlist.
- **Wrong period:** use weekly or monthly for Watched jobs. Other providers' periods are not necessarily supported by Simkl.
- **Fewer than the content limit:** feeds are capped, and filtering/duplicate checks further reduce the result.
- **Connection error:** confirm that the Blockbusterr host can reach Simkl discovery servers, including `data.simkl.in`.
- **Sonarr identity error:** a title needs a verified TVDB identity for direct Sonarr delivery. Preview explains titles that cannot be resolved.

See [Jobs](/concepts/jobs/) for scheduling and [Rules](/concepts/rules/) for interpreting filter results.
