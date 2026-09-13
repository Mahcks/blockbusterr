---
title: Sonarr
description: Connect a show destination, choose monitoring, and understand identity and download behavior.
---

Sonarr manages television series. In **Direct to Radarr and Sonarr** mode, Blockbusterr adds accepted shows using your selected Sonarr profile, root folder, and monitoring policy. Sonarr then handles episode searching, downloads, and imports.

You do not need Radarr for a shows-only setup. To route requests through Jellyseerr or Seerr instead, use the [request integration](/integrations/jellyseerr/); configure its Sonarr destination inside that server.

## Prepare Sonarr

Configure a series root folder, a quality profile, indexers, and a download client in Sonarr. Verify you can add a series there before using Blockbusterr.

Get its API key from **Settings → General → Security**. Choose a URL reachable from Blockbusterr, such as `http://sonarr:8989` when both containers share a Docker network. Use the service name actually defined in your deployment. `localhost` inside Blockbusterr's container does not point to a separate Sonarr container.

Include Sonarr's URL base if one is configured. Enter the server URL, not a URL ending in `/api/v3`.

## Connect and select defaults

1. Start Blockbusterr with `BLOCKBUSTERR_DRY_RUN=true` while setting up delivery; follow the [quickstart](/getting-started/quickstart/) to set the environment variable and recreate your container with its existing data mount.
2. Open **Settings → Connections → Sonarr**.
3. Enter **Server URL** and **API key**, then select **Test connection**.
4. Select **Reload profiles** and choose the **Quality profile** you use for shows.
5. Select **Reload root folders** and choose the series library path as Sonarr sees it, such as `/tv`.
6. Choose **Monitor** according to how much of a newly added show you want.
7. Select **Direct to Radarr and Sonarr** as the delivery mode, then **Save changes**.
8. Review a show job's preview, rule set, monitoring, and delivery cap while dry run is on.
9. Review all enabled jobs before disabling global dry run. Custom jobs are created enabled, so creating one can otherwise allow scheduled delivery immediately.

Profiles come from Sonarr; IDs are not list positions. The root folder must exist in Sonarr, and does not have to be mounted in Blockbusterr.

## Choose monitoring deliberately

Adding a show can involve many episodes. For a small first trial, consider **First season** or **Pilot episode**, then check how Sonarr configured that show.

| UI choice | Config value | Intended monitoring scope |
| --- | --- | --- |
| All episodes | `all` | The show's episodes broadly |
| Future episodes | `future` | Episodes that have not aired |
| Missing episodes | `missing` | Episodes missing from the library |
| Existing episodes | `existing` | Episodes with existing files |
| First season | `firstSeason` | First season |
| Last season | `lastSeason` | Last season |
| Latest season | `latestSeason` | Latest season according to Sonarr |
| Pilot episode | `pilot` | Pilot |
| Recent episodes | `recent` | Sonarr's recent-episode scope |
| Monitor specials | `monitorSpecials` | Sonarr's specials-monitoring option |
| Unmonitor specials | `unmonitorSpecials` | Sonarr's specials-unmonitoring option |
| None | `none` | No episodes monitored |

These are values passed to Sonarr. Monitoring and the initial missing-episode search are separate operations; do not use **None** as a substitute for Blockbusterr's dry run. Use dry run if you want no delivery at all while experimenting.

Individual jobs can override the monitoring default and choose a Sonarr series type: `standard`, `anime`, or `daily`. Select the appropriate type for the series naming conventions you use.

## Show identity and delivery

Direct Sonarr delivery requires a verified **TVDB identity**. Different discovery providers can identify shows differently; Blockbusterr resolves supported identities and rejects unresolved or mismatched results rather than adding a similarly named show.

TMDB-sourced shows need the relevant identity resolution to succeed. A candidate can satisfy your content rules and still fail delivery because it has no usable TVDB mapping. Read the preview or Activity Entry for that distinction.

A successful addition means Sonarr accepted the series. It does not mean every requested episode is available or downloaded. Removing the series from an upstream list does not delete it from Sonarr.

## Troubleshooting

- **Connection error:** check the container network, URL, port, URL base, and API key.
- **Empty dropdowns:** configure profiles/root folders in Sonarr, then reload them in Blockbusterr.
- **Identity error:** inspect the show's provider IDs and preview reason. Do not fix it by guessing a different series title.
- **Too many episodes requested:** check the global and per-job Monitor choices and inspect the series in Sonarr. A job delivery cap limits shows, not episode counts.
- **No download after addition:** check Sonarr's monitored seasons, search results, queue, indexers, download client, and quality profile.
- **Nothing delivered:** inspect dry run, enabled state, scheduling, duplicates, caps, and the selected delivery mode in **Job Runs** and **Activity Entries**.
