---
title: Trakt
description: Set up optional Trakt discovery, understand the VIP credential requirement, and connect lists or watchlists.
---

Choose Trakt if you want its box-office, anticipated, watched, collected, played, or favorited discovery, or already follow Trakt lists. It also supports trending, popular, and Smart Popular jobs. **Trakt is optional**: TMDB, Simkl, and MDBList jobs can run without it.

## Account cost and credentials

Trakt made [creating API applications a VIP feature on July 30, 2026](https://github.com/trakt/trakt-web/pull/3057). Blockbusterr currently asks you to supply your own application's client ID and secret, so do not assume a new Trakt setup is free. The ability to connect to an existing community app is a separate feature from creating these credentials. Check Trakt's current account terms before subscribing.

If you want to avoid this dependency, start with [TMDB](/integrations/tmdb/) for trending/popular discovery or [MDBList](/integrations/mdblist/) for lists. TMDB popularity is an approximation of public interest, not a substitute for a box-office chart.

## Connect Trakt

Start Blockbusterr with `BLOCKBUSTERR_DRY_RUN=true` before creating custom jobs or changing delivery; follow the [quickstart dry-run setup](/getting-started/quickstart/). Custom jobs are created enabled; recipes start disabled. Keep dry run on while reviewing sources and previews, then review all enabled jobs before turning it off.

1. Sign in to Trakt and open [API applications](https://app.trakt.tv/settings/apps/api).
2. Register an application for your Blockbusterr installation and copy its **client ID** and **client secret**.
3. Open **Settings → Connections → Trakt** in Blockbusterr and enter both values.
4. Select **Test connection**, then **Save changes**.
5. Open **Jobs → New job → Custom Job**, choose your media and discovery type, and select **Trakt** as the source.
6. Assign rules and review the preview while dry run remains on. Check the cap and destination before allowing delivery.

Public discovery uses the client ID. The Settings connection test expects both fields, and account authorization needs the application credentials. You do not have to connect a personal account to use public discovery or public lists.

## Pick a job

| Goal | Discovery type | Scope |
| --- | --- | --- |
| Keep up with current releases in theaters | Box office, movies | The provider's current box-office feed, not a custom worldwide revenue threshold |
| Follow what people are watching | Watched or Played | A selected supported period |
| Follow upcoming interest | Anticipated | Upcoming titles may have few or no ratings |
| Find broadly discussed titles | Trending or Popular | Provider-ranked candidates, then your local rules |
| Follow a curator | List or Watchlist | Only titles present in that list |

The Jobs editor shows valid combinations for the selected media type. Standard rating and vote rules use the Trakt values supplied by these feeds; they are not IMDb scores or critic-review counts.

For example, a **Box office** movie job with minimum rating `5` filters current box-office candidates using their Trakt rating. It does not mean “worldwide gross above a chosen amount and IMDb above 5.” Preview to see what is actually included.

## Follow a list or watchlist

Create a **List or Watchlist** job with source **Trakt** and enter identifiers rather than the complete web URL.

| Source | List kind | Owner or member | List ID or slug |
| --- | --- | --- | --- |
| `trakt.tv/users/alex/lists/weekend-movies` | Public list | `alex` | `weekend-movies` |
| A global numeric list ID | Public list | Leave empty | The numeric ID |
| Another user's public watchlist | Personal watchlist | Their username | Leave empty |
| Your connected account's watchlist | Personal watchlist | Leave empty | Leave empty |

The first row illustrates the URL format; it is not a suggested real list. Use **Check source** to verify access, then preview to check the assigned rules and delivery settings.

For your connected watchlist, first save the credentials and select **Connect account for watchlists** under Trakt in Settings. Open the displayed activation link, enter the device code, and wait for **Account connected**. Private content requires authorization; entering a username does not grant access.

List synchronization only adds accepted media. Removing a source item does not delete an already delivered movie or show.

## Troubleshooting

- **Cannot create credentials:** check the VIP requirement above. Blockbusterr cannot bypass Trakt account restrictions.
- **Authentication or client-not-found error:** verify the application still exists and replace revoked credentials. Reconnect your account if its authorization expired.
- **Public list works but personal watchlist fails:** complete account authorization, then leave the watchlist owner empty to use that connected account.
- **Source unavailable:** save credentials and choose a compatible job type. Account authorization alone does not configure the discovery client.
- **Small or empty results:** inspect the source limit, Trakt's account/list limits, and rule rejection reasons separately. A successful connection does not guarantee that a particular list is accessible.
- **Network error:** verify access to `api.trakt.tv` from the machine or container running Blockbusterr.
