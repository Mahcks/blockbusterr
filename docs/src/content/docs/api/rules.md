---
title: Rules API
description: Manage reusable rule sets, title exceptions, and job-specific rule copies.
---

All endpoints use the `/v1` base path and return JSON unless noted otherwise.

## List rule sets

`GET /v1/rule-sets`

Returns each rule set with its `usage_count`.

## Create a rule set

`POST /v1/rule-sets`

```json
{
  "name": "Recent Movies",
  "media": "movie",
  "movies": {
    "blacklisted_min_year": 2020,
    "min_rating": 6.5,
    "min_votes": 500
  }
}
```

The server generates an ID when omitted and starts the revision at `1`.

## Update a rule set

`PUT /v1/rule-sets/:id`

Send the complete rule set and its current `revision`. A successful update increments the revision. A stale revision returns `409 Conflict` so one editor cannot silently overwrite another.

Default rule sets can be edited but not deleted.

## Delete a rule set

`DELETE /v1/rule-sets/:id`

Returns `204 No Content`. Assigned rule sets and built-in defaults return `409 Conflict`; reassign dependent jobs first.

## Title exceptions

- `GET /v1/title-exceptions`
- `PUT /v1/title-exceptions`

```json
{
  "allowed_movie_tmdb_ids": [603],
  "blocked_movie_tmdb_ids": [],
  "allowed_show_tvdb_ids": [],
  "blocked_show_tvdb_ids": [12345]
}
```

The update replaces the complete exception document. Read it first when performing a partial administrative change.

## Make a job-specific copy

`POST /v1/jobs/:id/customize-rules`

Clones the job's effective rule set, assigns the new copy to the job, and returns the updated job and new rule set. This endpoint is for dynamic jobs.

## Errors

| Status | Meaning |
|---|---|
| `400` | Invalid payload or media-specific fields |
| `404` | Job or rule set not found |
| `409` | Duplicate/stale rule, protected default, assigned rule, or incompatible state |
| `500` | Configuration could not be saved or reloaded |
