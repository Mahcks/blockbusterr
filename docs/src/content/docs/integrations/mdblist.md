---
title: MDBList
description: Follow public lists and watchlists, configure list identifiers, and handle missing ratings and genres correctly.
---

MDBList is useful when you already have a curated list or want to build a list using criteria such as IMDb ratings or critics' scores. Blockbusterr fetches its members, applies your assigned rules, and delivers accepted titles to Radarr, Sonarr, or Jellyseerr / Seerr.

**You do not need Trakt credentials for Blockbusterr's MDBList integration.** It connects directly to MDBList using an API key. TMDB and Simkl remain useful alternatives for native trending/popular discovery.

## Before creating a job

Start Blockbusterr with `BLOCKBUSTERR_DRY_RUN=true` before creating a custom job; see the [quickstart](/getting-started/quickstart/) for how to set this environment variable and recreate your container with its existing data mount. Custom jobs are created enabled; a schedule may otherwise deliver titles before you review them. Keep dry run on until you have checked the source, rules, and destination. See [Jobs](/concepts/jobs/) for scheduling and delivery limits.

## Connect MDBList

1. Create an MDBList account and get a free API key from [Preferences](https://mdblist.com/preferences/).
2. Open **Settings → Connections → MDBList** in Blockbusterr.
3. Enter the key, select **Test connection**, then **Save changes**.
4. Open **Jobs → New job → Custom Job**.
5. Choose **List or Watchlist**, **MDBList**, and either movies or shows.
6. Enter the list fields below and select **Check source**.
7. Assign an appropriate rule set, save the job, and review its preview while dry run remains enabled.
8. Review delivery caps and the destination before turning dry run off. That change allows all enabled jobs to deliver, not only this list job.

## Enter the list correctly

A slug is the readable name at the end of a list URL. For the example `https://mdblist.com/lists/alex/weekend-movies/`, the owner is `alex` and the slug is `weekend-movies`. This example demonstrates the format, not a real recommended list.

| Source | List kind | Owner or member | List ID or slug |
| --- | --- | --- | --- |
| A user's public list | Public list | Username, such as `alex` | Slug, such as `weekend-movies` |
| A numeric list ID | Public list | Leave empty | Numeric ID |
| An official MDBList list | Public list | Leave empty | Official slug |
| Your MDBList watchlist | Personal watchlist | Leave empty | Leave empty |

Do not paste the complete URL into **List ID or slug**. Your watchlist belongs to the account that issued the API key; it does not use a separate OAuth connection. A mixed-media list needs separate movie and show jobs to deliver both kinds.

**Check source** verifies access and shows a sample. It does not mean the entire list will pass your rules or fit the delivery cap.

## Put rating and genre criteria in the right place

The current MDBList list adapter imports title, year, IDs, and available language/country. It **does not import ratings, vote counts, genres, runtime, or critic-review counts** into Blockbusterr's ordinary candidate fields. Configuring a TMDB API key does not automatically fill those fields for MDBList list jobs.

This means copying a rule set with minimum rating `7`, minimum votes `100`, or a required genre onto an MDBList job can reject its entries even though those titles satisfy the criteria on MDBList.

For a list already filtered on MDBList:

1. Set the IMDb, critics' score, genre, or other selection criteria on the MDBList side.
2. Create a dedicated Blockbusterr rule set for this list rather than weakening your default movie rules.
3. Set rating and minimum-vote requirements to `0`, leave required genres empty, and remove runtime or other requirements that depend on metadata the list does not provide.
4. Retain checks that make sense for the supplied data and your intended policy. Review language/country restrictions too, since their presence varies.
5. Preview and inspect rejection reasons. Do not enable delivery until the results match what you intended.

A generic `8.7` rating does not represent critics agreeing about a film. If your goal is acclaim from a small group of critics, use an upstream critics-score filter or a curator whose selection you trust. IMDb user votes, critic counts, and percentage-positive critic scores are different measures.

## Example: three separate movie tastes

Keep these as separate jobs so each can have its own source, rules, schedule, and cap:

| Goal | Practical source | Where to filter |
| --- | --- | --- |
| Current popular movies | Native TMDB Trending/Popular | Blockbusterr rating rules; this is popularity, not box-office revenue |
| Critically acclaimed independent films | An MDBList you configure or a trusted curator's list | MDBList criteria or the curator's selection; do not claim low vote count proves a film is independent |
| Science-fiction movies | Native TMDB Science-Fiction Discovery recipe, or a dedicated MDBList | Genre/rating rules in Blockbusterr for TMDB; upstream for MDBList |

MDBList supports imported external lists, including sources such as IMDb and Letterboxd. Consult [its list types](https://docs.mdblist.com/docs/list_types) to distinguish imported, dynamic, and static lists. A manually copied static list does not automatically gain its source's future updates.

## Free limits and refresh behavior

As checked in September 2026, MDBList's free plan includes **4 dynamic lists, 4 static lists, 1 external list, 10,000 items per list, and 1,000 API requests per day**. Newly created free lists may take 30 minutes to populate. See the [current plan table](https://docs.mdblist.com/docs/supporter) because service limits can change.

Blockbusterr fetches lists when checking, previewing, or running them; pagination can require multiple requests. A successful connection test does not grant unlimited API access.

Free MDBList lists can stop updating after 120 days without account activity. The provider documents that direct Radarr list use counts as activity; do not assume Blockbusterr API reads do. Log in periodically or check MDBList's current policy.

Removing an item from a source list never deletes media Blockbusterr already delivered. A job follows additions; it is not a mirror that removes your library's other titles.

## Troubleshooting

| Problem | What to check |
| --- | --- |
| Source check fails | Verify API key, public access, owner, and slug/ID separately. |
| New list is empty | Wait for MDBList to populate it and open the list on MDBList itself. |
| Every candidate fails rating/votes/genre | Use a dedicated rule set appropriate to the adapter's missing metadata, as described above. |
| API returns a rate-limit error | Reduce preview frequency, list sizes, or job cadence; inspect the account's API usage. |
| Source stopped changing | Check the upstream list's refresh and free-account activity status. |
| A show fails delivery | Check that the entry has a resolvable identity for the selected delivery target. |

For direct experimental Letterboxd access, see [Letterboxd](/integrations/letterboxd/).
