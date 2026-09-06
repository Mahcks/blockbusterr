---
title: Activity API
description: Query Activity Entries, Job Runs, charts, statistics, and title actions.
---

The Activity API exposes individual title decisions and aggregate job executions.

## Activity Entries

`GET /v1/activity/logs`

Supported query parameters include limit, page, search, status, media type, job type, language, date range, and sort options used by the Activity UI.

```bash
curl "http://localhost:9090/v1/activity/logs?limit=50&status=rejected"
```

Each entry contains its job identity, media identity, outcome, decision message, optional score/rank, provider IDs, poster metadata, and structured filter details when available.

## Job Runs

`GET /v1/activity/runs`

Returns executions with start/finish timing and found, passed, added, requested, rejected, skipped, and failed totals. The response also includes recent ranked-selection cycle summaries when that opt-in feature is enabled.

The UI reports direct Radarr/Sonarr additions as **Added**, Jellyseerr/Seerr submissions as **Requested**, and their combined global total as **Delivered**. The API preserves the separate `added` and `requested` fields.

## Supporting data

- `GET /v1/activity/languages`
- `GET /v1/activity/chart`
- `GET /v1/activity/stats`
- `GET /v1/activity/rejection-breakdown`

These endpoints power the Activity filters, overview metrics, trend chart, and rejection summaries.

## Retention

`DELETE /v1/activity/logs?days=30`

Deletes Activity Entries older than the requested retention period. This operation is irreversible unless the database is restored from backup.

## Add anyway

`POST /v1/activity/:id/add-anyway`

Attempts delivery for an eligible rejected entry using the configured integration mode and updates the recorded outcome when successful.

## Block a title

`POST /v1/activity/:id/block`

Adds the title's provider ID to the universal block exceptions for its media type. Future jobs evaluate that exception before their assigned rule set.

## Debug

`GET /v1/activity/debug`

Returns installation diagnostics for maintainers. Avoid exposing it outside a trusted network.
