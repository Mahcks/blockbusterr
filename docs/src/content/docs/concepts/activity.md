---
title: Activity and troubleshooting
description: Understand why a movie or show was added, rejected, skipped, or failed, and find the setting to check next.
---

Start with **Activity** when a job does something unexpected. You usually do not need to read Docker logs to find out why a particular movie failed a rule.

## Which view should I open?

| View | Use it to answer… |
| --- | --- |
| **Activity Entries** | What happened to this movie or show? Which rule or delivery check stopped it? |
| **Job Runs** | Did my job run? How many candidates did it find, and did an error stop it? |
| **Preview** on a job | What would the current source and rules produce before I run it? |

Use the job, status, and title filters to narrow down the history. Open an entry's details rather than relying on its color or summary alone. If a source failed before returning titles, the useful error may be in **Job Runs**, even when there are no Activity Entries for that execution.

## What do the outcomes mean?

| Outcome | Meaning | What to check next |
| --- | --- | --- |
| Added | The destination accepted the addition, or dry-run simulated one | Radarr/Sonarr for monitoring, searches, and download progress; read the entry message for `[DRY RUN]` |
| Requested | A request was created, or dry-run simulated one | Jellyseerr/Seerr for its approval and fulfillment status |
| Rejected | The title did not pass its rules | Open the rule checks to see the failed condition |
| Skipped | No addition was made, for example because it already exists, repeat handling prevents it, or a delivery cap was reached | Read the reason; a skipped title does not necessarily need a fix |
| Blocked | The title was stopped by a blocking decision | Read the detailed reason and relevant title exception or rule |
| Failed | A fetch, lookup, or delivery operation could not complete | Check the error, credentials, connection, and provider availability |

Dry-run can record **Added** and **Requested** entries with a `[DRY RUN]` message. Those are simulated outcomes. They do not prove that anything was sent to another app.

A **Completed** Job Run means the execution finished. It does not mean every candidate passed, or that every delivered title has downloaded. A run with zero additions can be working correctly.

## Walk through a result

Suppose a job has a discovery limit of 50, a minimum rating of 6, and a delivery cap of 3.

1. The source returns 50 movies.
2. Twelve pass the rule set.
3. Four of those are already in Radarr, leaving eight eligible candidates.
4. The job delivers three and skips the rest because of its cap.

The other five were not necessarily bad matches. They may be considered again if a later run discovers them and has capacity. They are **not a guaranteed queue**: the source list can change before the next run.

## “Why didn't this movie get added?”

### It does not appear in Preview

Check the **source, job type, and discovery limit** first. Rules only inspect the candidates fetched by that job.

For example, a science-fiction rule on a Trending job will not find a niche sci-fi film that is outside the fetched trending batch. Increasing the discovery limit can widen that batch, but it does not turn it into a search of the entire catalog. A supported list or a TMDB Recommendations job may be a better source for the title you have in mind.

Adding a movie to **Always allow** does not fetch it. The exception applies only if a job discovers that movie.

### It appears but is rejected

Open its rule checks and compare the actual value with your boundary. Common causes:

- The rating comes from the source and differs from the IMDb score you looked up.
- The title has too few votes, even though its rating is high.
- The original language is not one of your required languages.
- A required genre does not match the supplied value. For TMDB sci-fi, use `science-fiction`.
- A title contains a blocked keyword. Keyword checks use substrings, not whole words.
- An age certification is missing and you selected **Reject** for unknown ratings.
- The source did not supply the rating, votes, or genres needed by your rules.

For **MDBList**, apply rating and genre criteria upstream and use an empty Blockbusterr rule set for those already-filtered results. Its current importer does not fill those fields. See [Rules](/concepts/rules/) and the [MDBList guide](/integrations/mdblist/).

### It passes but is skipped

Read the skip reason. Check whether the title is already in the destination, was previously delivered and is covered by repeat handling, or lost out to a job cap, Global Limit, or ranked-selection cutoff.

An **Always allow** exception bypasses rule rejection; it does not bypass existing-library checks, repeat handling, delivery limits, or missing destination IDs.

### It says Added, but there is no file

First look for `[DRY RUN]` in the entry message. If the delivery was real, open Radarr or Sonarr:

- Is the title monitored?
- Is its minimum availability satisfied?
- Is there a release matching the quality profile?
- Is automatic searching enabled where you want it?
- Do the indexers and download client work?

Blockbusterr sends the title to the destination. The destination handles finding and downloading files. See [Radarr](/integrations/radarr/) or [Sonarr](/integrations/sonarr/).

### It says Requested, but nobody was asked to approve it

Request mode uses Jellyseerr/Seerr's permissions. A request may be approved automatically if the requesting identity has that permission. Configure the request identity and approval permissions there if you want a review step. See [delivery modes](/concepts/integration-modes/).

## “Preview looked different from the live run”

A regular job preview evaluates a fetched batch against the current rules and existing/repeat checks. Its eligible count does **not** simulate every live delivery cap or the competition with other jobs.

Other things can change between preview and execution:

- The provider's trending or popular list changes.
- Another job uses the remaining shared delivery budget.
- A title is added to Radarr or Sonarr elsewhere.
- Someone edits a shared rule set.
- A provider or destination becomes unavailable.

Use the **ranked-selection preview** when comparing participating jobs together. Even that preview is a snapshot, not a reservation of future delivery slots.

## “My job isn't running”

1. Check whether it is enabled. Recipes start disabled; Custom Job currently starts enabled.
2. Check its delivery connection and source credentials. Only supported, configured sources appear in the job editor.
3. Check the schedule. `24h` is an interval, while `0 9 * * *` means 09:00 each day in the configured timezone. Duration jobs can run on startup; cron jobs wait for their next scheduled time.
4. If **Include in ranked selection** is on and ranked selection is enabled globally, the job follows the shared selection schedule. It does not run on its own schedule.
5. Look in **Job Runs** for failures, then check container logs if no run explains the problem.

```bash
docker logs --tail 100 blockbusterr
```

## “The connection test fails”

| Symptom | First thing to check |
| --- | --- |
| Connection refused or timeout | Correct URL and port; both containers can reach each other |
| URL uses `localhost` | Inside Blockbusterr's container this points to Blockbusterr itself; use the other service's reachable address |
| Unauthorized / 401 / 403 | Correct service's API key, access permissions, and any reverse-proxy authentication |
| Provider not available in Jobs | Save its credentials, then choose a supported job type and media type |
| List not found | Correct owner and list ID/slug, public visibility or required account authorization; do not paste a full web URL into the ID field |
| Profile or folder is missing | Create it in Radarr/Sonarr, then load the options again in Blockbusterr |
| Rate-limit response / 429 | Wait for the provider's quota to recover; reduce unnecessary runs and previews |

Some providers are checked by making a job preview rather than a separate connection-test button. Follow the guide for the provider you are using.

## “A deleted movie keeps coming back”

A movie can still be on the source list after you delete it from Radarr. Check **Repeat handling** on the job and the global default. Options include a cooldown or **Never re-add**.

Cooldowns are measured from Blockbusterr's **last successful delivery**, not from the moment you delete the movie. They depend on Blockbusterr having recorded that delivery. Removing a title from a source list also does not delete it from your library.

Normal activity cleanup preserves delivery memory, which is stored separately. **Clear all** also offers an explicit **Also clear delivery memory** option; using it removes the history needed to prevent re-adds. See [Jobs](/concepts/jobs/) for repeat examples.

## Before asking for help

Include the Blockbusterr version, job type and source, whether dry-run is enabled, and the relevant Preview or Job Run error. Say what you expected and what actually happened. For example:

> TMDB Popular movie job, rule set requires science-fiction and rating 6. Preview rejects a movie for missing genre. I expected it to pass because TMDB lists it as science fiction.

Remove API keys, tokens, and other credentials from anything you share. A short error and the relevant settings are usually more useful than a full configuration dump.
