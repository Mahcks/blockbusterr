---
title: Radarr
description: Connect a movie destination, select real profiles and folders, and verify what happens after delivery.
---

Radarr manages movies. In **Direct to Radarr and Sonarr** mode, Blockbusterr adds accepted movies to Radarr using your selected quality profile, root folder, availability, and monitoring settings. Radarr handles release searching, downloads, and imports through its own configuration.

If you only want movies, you do not need Sonarr. If you want requests to go through Jellyseerr or Seerr, configure [that delivery target](/integrations/jellyseerr/) instead; its server owns the downstream Radarr settings.

## Prepare Radarr

Before connecting Blockbusterr, make sure you can add a movie in Radarr itself. Configure your library root folder, quality profiles, indexers, and download client there first. Blockbusterr does not configure those services for you.

Copy the Radarr API key from **Settings → General → Security**. You also need its server URL as reachable from the machine or container running Blockbusterr.

| Deployment | Example server URL |
| --- | --- |
| Both programs run directly on the same host | `http://localhost:7878` |
| Containers share a Docker network and Radarr's service name is `radarr` | `http://radarr:7878` |
| Radarr runs on another LAN machine | `http://192.168.1.20:7878` |

In a container, `localhost` refers to that container. A URL that works in your laptop's browser may not work from Blockbusterr. Include Radarr's configured URL base if it uses one; do not append `/api/v3`.

## Connect and select defaults

1. Start Blockbusterr with `BLOCKBUSTERR_DRY_RUN=true` while configuring delivery; follow the [quickstart](/getting-started/quickstart/) to set the environment variable and recreate your container with its existing data mount.
2. Open **Settings → Connections → Radarr**.
3. Enter **Server URL** and **API key**, then select **Test connection**.
4. Select **Reload profiles** and choose a **Quality profile** by name.
5. Select **Reload root folders** and choose the movie library folder configured in Radarr.
6. Start with **Minimum availability → Released** and **Monitor → Movie only** if those match how you normally add movies.
7. In Settings, choose **Direct to Radarr and Sonarr** as the delivery mode, then **Save changes**.
8. Create or review a movie job and inspect its preview while dry run remains on.
9. Confirm its rules, destination, and delivery cap before turning dry run off. Review all enabled jobs first, because dry run is global.

A **quality profile** is Radarr's policy for acceptable release qualities and upgrades. Its numeric ID is not its position in a list; selecting it in Blockbusterr avoids guessing IDs.

The **root folder** is the path Radarr sees, for example `/movies`. It is not necessarily the host's path, and Blockbusterr does not need the movie files mounted locally to submit additions through Radarr's API.

## Availability and monitoring

Minimum availability controls when Radarr considers a movie available. It does not enforce a video quality; Radarr's quality profile and other release settings do that.

| UI choice | Config value | Meaning |
| --- | --- | --- |
| Announced | `announced` | Use Radarr's announced availability setting |
| In cinemas | `inCinemas` | Use theatrical availability |
| Released | `released` | Use released availability |

| Monitor choice | Config value | Result |
| --- | --- | --- |
| Movie only | `movieOnly` | Monitor the selected movie |
| Movie and collection | `movieAndCollection` | Ask Radarr to monitor the movie and its collection |
| None | `none` | Add without monitoring or an initial movie search from Blockbusterr |

A collection is Radarr's collection grouping, not every movie in a shared universe. Review collection monitoring before using it because its scope can extend beyond the one title Blockbusterr selected.

Jobs can override availability and monitoring. Check an individual job if its behavior differs from the Settings defaults.

## What successful delivery means

A successful addition means Radarr accepted the movie into its management database. It does not mean a release exists, a download completed, or a playable file is already in your library. For monitored movies, Blockbusterr also asks Radarr to search when adding.

Existing movies are handled as duplicates. Removing a title from a discovery list does not delete it from Radarr. See [Jobs](/concepts/jobs/) for repeat handling when a previously delivered title is later removed from your library.

## Troubleshooting

| Problem | What to check |
| --- | --- |
| Connection refused or timed out | Check the URL from Blockbusterr's host/container, network membership, port, and firewall. |
| Unauthorized | Copy the API key from this Radarr instance; do not use the Blockbusterr or Sonarr key. |
| No profiles or folders | Create them in Radarr first, then reload the dropdowns. |
| Preview passes but nothing is added | Check dry run, job enabled state, global/per-job delivery mode, caps, duplicates, and Job Runs. |
| Movie appears but does not download | Inspect Radarr's history, queue, indexers, download client, availability, and quality decisions. |
| Wrong folder or quality | Review the saved Radarr defaults and confirm the job is delivering directly rather than through a request server. |

Use **Activity Entries** for individual delivery outcomes and **Job Runs** for the scheduled run as a whole.
