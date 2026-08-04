---
title: Jobs API
description: Create, inspect, preview, trigger, migrate, and delete v2 jobs.
---

All endpoints use `/v1` and return JSON unless noted otherwise.

## Discover capabilities

- `GET /v1/jobs/types` returns job definitions filtered to configured providers.
- `GET /v1/jobs/templates` returns versioned recipes with readiness and missing-setup details.
- `POST /v1/jobs/recipes/:id` creates a disabled recipe job and dedicated rule set.

Recommendation jobs use `type: recommendations`, `source: tmdb`, and either
`recommendation_seeds` (up to 20 TMDB IDs) or `recommendation_list` containing
a list-provider `source` and normal list locator. Show jobs may set
`series_type` to `standard`, `anime`, or `daily`.
- `GET /v1/jobs/list` returns dynamic and readable legacy jobs.
- `GET /v1/jobs/enabled` returns enabled jobs.

## Read a job

`GET /v1/jobs/dynamic/:id`

Returns one job with its resolved rule set. Legacy jobs are read-only through the dynamic API.

## Create a job

`POST /v1/jobs`

```json
{
  "name": "Trending Movies",
  "enabled": false,
  "type": "trending",
  "source": "tmdb",
  "media": "movie",
  "limit": 50,
  "delivery_limit": 5,
  "rule_set_id": "default-movies"
}
```

The server validates the type/source/media combination, discovery and delivery limits, period, schedule, delivery mode, and rule assignment. A delivery limit of `0` is unlimited.

## Update and delete

- `PUT /v1/jobs/:id` updates a dynamic job.
- `DELETE /v1/jobs/:id` deletes a dynamic job.

Legacy jobs return a conflict/error and should be migrated first.

## Preview

`POST /v1/jobs/:id/preview`

Discovers and evaluates candidates without delivering them. Preview uses the same source and assigned rule set as a live run.

`POST /v1/jobs/selection/preview`

Fetches participating jobs and returns per-job stage counts, deduplicated movie/show winners, below-cutoff candidates, merged source/job provenance, delivery-limit exclusions, and provider errors without delivering anything.

## Trigger

`POST /v1/jobs/:id/trigger`

Starts an enabled dynamic job asynchronously. Use Activity Job Runs to follow the resulting execution.

## Migrate legacy jobs

`POST /v1/jobs/migrate`

Creates dynamic equivalents for enabled legacy jobs, assigns default media rules, saves the configuration, and disables the migrated legacy entries.

## Customize rules

`POST /v1/jobs/:id/customize-rules`

Creates and assigns an independent copy of the job's effective rule set. See the [Rules API](/api/rules/).

## Legacy compatibility endpoints

The following v1 endpoints remain available for existing integrations:

- `GET /v1/jobs/status`
- `POST /v1/jobs/trigger/:job`
- `POST /v1/jobs/preview/:job`
- `GET /v1/jobs/:job/decisions`

New clients should use stable dynamic job IDs and the `/jobs/:id/...` forms.
