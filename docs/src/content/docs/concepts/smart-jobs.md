---
title: Smart Jobs
description: Let popular titles meet a lower rating threshold, with a worked setup and clear limits.
---

A **Smart Popular** job starts with a provider's popular movies or shows. It then gives titles with more votes a lower minimum rating and asks titles with fewer votes to meet a higher one. Choose it when you want mainstream hits to have an easier path into your library.

For a fixed rule such as “science-fiction rated at least 6,” use a regular **Popular** job. Smart Popular is optional; you do not need it to automate discovery.

## Create your first Smart Popular job

First connect a discovery provider and a delivery service using the [quickstart](/getting-started/quickstart/). Start with `BLOCKBUSTERR_DRY_RUN=true`: custom jobs are enabled on creation and can execute before you finish tuning them.

1. Open **Rules** and create a Movie rule set from **Empty rule set**. Name it `Smart mainstream movies`.
2. Under **Boundaries**, set a minimum of `100` votes as a starting point. Save. This excludes very small audience samples; it is a preference, not a guarantee of quality.
3. Open **Jobs → New job → Custom Job**.
4. Name the job `Popular with a flexible rating`, choose **Smart Popular**, **Movie**, and **TMDB**.
5. Assign `Smart mainstream movies`. Set **Discovery limit** to `100`, **Job delivery cap** to `5`, **Base Minimum Rating** to `6.5`, and **Adjustment Factor** to `2`.
6. Create the job, disable it while tuning, open its editor, and expand **Delivery & schedule**. Set **Custom Sync Interval** to `24h` and choose the delivery mode you configured.
7. Save and **Preview**. Read the rating decisions for both accepted and rejected titles. Restore normal delivery after reviewing all enabled jobs, then enable this job when ready.

TMDB is one supported source; Trakt and Simkl also support Smart Popular when configured. The rating and vote data come from the source, so switching providers can change the decisions.

## What those numbers mean

Popularity here means **vote-count rank within the candidates fetched for this run**. It is not box-office revenue, critic acclaim, or an absolute popularity threshold across the provider's entire catalog.

With a baseline of `6.5` and adjustment factor of `2`:

| Position by vote count in this batch | Required rating |
|---|---|
| Fewest votes | 7.5 |
| Middle | 6.5 |
| Most votes | 5.5 |

The calculation is `baseline - (percentile - 0.5) × adjustment factor`, rounded to one decimal place and limited to 0–10. A batch containing only one title uses the baseline. Changing the discovery limit can change a title's position and therefore its threshold.

Raising the factor widens the gap between the least-voted and most-voted titles. Raising the baseline makes the whole job stricter. Use a regular **Popular** job for a fixed rating threshold: in the current Smart Popular implementation, an adjustment factor of `0` falls back to `0.5`, and a baseline of `0` falls back to `6`.

## How Smart Popular interacts with rules

The adaptive threshold **replaces the assigned rule set's minimum-rating boundary for this job**. A separate minimum rating of 8 in that rule set does not impose an additional floor of 8. The saved rule set itself is not changed.

Other rules still apply: minimum votes, required genres, languages, years, runtime, content ratings, and title exceptions. Allow overrides and universal title exceptions retain their normal behavior; see [Rules](/concepts/rules/) before using them to bypass checks.

For example, add required genre `science-fiction` to make the previous job accept only science-fiction candidates. A widely voted action film with no matching genre still fails. Use **Required matches**, not **Allow overrides**, when the genre must be mandatory.

## Smart filtering versus ranked selection

Smart Popular determines whether a title passes its adaptive rating check. [Ranked selection](/concepts/jobs/#ranked-selection-cycles) compares qualifying candidates across participating jobs and chooses winners for a shared cycle. They are different settings and can be used together.

Neither setting discovers films outside the fetched source results. A little-known indie film absent from the popular chart never reaches the adaptive check. An audience score of 8.7 also does not mean critics awarded it 8.7; use a curated source when critic selection is the actual requirement.

## Tune one thing at a time

| What the preview shows | Change to try |
|---|---|
| Too many weak mainstream titles pass | Raise the baseline or reduce the adjustment factor |
| Low-vote titles dominate | Raise minimum votes in the assigned rules |
| Nothing passes | Read rejection reasons; a genre, language, or vote rule may be responsible |
| A desired title never appears | Check the source and discovery limit before relaxing rules |
| Too many titles would be delivered | Lower the job delivery cap; keep discovery broad enough to find matches |

After enabling, use **Activity → Job Runs** to inspect each execution and **Activity Entries** to understand an individual title. Preview again after changing the provider, limits, or rules.
