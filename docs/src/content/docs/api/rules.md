---
title: Rules API
description: Manage reusable rule sets, title exceptions, and job-specific rule copies.
---

All endpoints use the `/v1` base path and return JSON unless noted otherwise. See [Rules](/concepts/rules/) for what each field actually does and the exact formats providers report (genre slugs, country/language codes, TMDB-vs-TVDB IDs) — this page only covers the HTTP contract.

## List rule sets

`GET /v1/rule-sets`

Returns an array; each entry wraps the rule set with how many jobs use it:

```json
[
  {
    "rule_set": {
      "id": "default-movies",
      "name": "Default Movies",
      "media": "movie",
      "revision": 3,
      "movies": {
        "allowed_countries": [],
        "allowed_languages": [],
        "blacklisted_genres": [],
        "blacklisted_keywords": [],
        "blacklisted_tmdb_ids": [],
        "blacklisted_min_runtime": 0,
        "blacklisted_max_runtime": 0,
        "blacklisted_min_year": 0,
        "blacklisted_max_year": 0,
        "min_rating": 0,
        "min_votes": 0,
        "unknown_certification": "allow"
      }
    },
    "usage_count": 4
  }
]
```

Every field inside `movies`/`shows` is always present in the response — there's no `omitempty` shrinking here. A list that has never been populated serializes as `null`, not `[]`; once you've saved it with at least an empty array, it stays `[]`. Treat `null` and `[]` as the same thing ("no restriction") when reading this API. The only field that's ever fully absent is the top-level `movies` or `shows` key itself — a show rule set's response has no `movies` key at all, and vice versa.

## Create a rule set

`POST /v1/rule-sets`

Send `media` (`movie` or `show`) and exactly the matching filter payload — a movie rule set must include `movies` and must not include `shows`, and vice versa.

```json
{
  "name": "Family Movies",
  "media": "movie",
  "movies": {
    "allowed_countries": [],
    "allowed_languages": [],
    "blacklisted_countries": [],
    "blacklisted_languages": [],
    "blacklisted_genres": [],
    "blacklisted_keywords": [],
    "blacklisted_tmdb_ids": [],
    "blacklisted_min_runtime": 0,
    "blacklisted_max_runtime": 0,
    "blacklisted_min_year": 0,
    "blacklisted_max_year": 0,
    "required_genres": ["family"],
    "min_rating": 5,
    "min_votes": 0,
    "certification_country": "US",
    "allowed_certifications": ["G", "PG", "PG-13"],
    "blocked_certifications": [],
    "unknown_certification": "reject"
  }
}
```

```json
{
  "id": "8f9c1e2a-...",
  "name": "Family Movies",
  "media": "movie",
  "revision": 1,
  "movies": { "...": "as sent, with defaults applied" }
}
```

The server generates a UUID for `id` when you omit it and always starts `revision` at `1` — any `revision` you send is ignored on create. `unknown_certification` defaults to `"allow"` when omitted. Returns `201 Created`.

Validation runs before the write, so a bad payload never reaches disk. Notable rules (see [Rules](/concepts/rules/) for the reasoning behind each):

- `id` and a non-blank `name` are required; `media` must be exactly `movie` or `show`.
- The filter payload matching `media` is required, and the other one must be absent (a `movie` rule set with a `shows` key set is rejected).
- No numeric field may be negative; `min_rating` and `allow_min_rating` must be between `0` and `10`; a configured min year/runtime cannot exceed its max.
- If either certification list is non-empty, `certification_country` must be exactly two characters.
- `unknown_certification` must be `allow` or `reject` (Fiber will accept an omitted value, but an unrecognized string is rejected).
- The `name` must be unique among rule sets of the same `media` — creating a second movie rule set named `Family Movies` fails.

## Update a rule set

`PUT /v1/rule-sets/:id`

Send the complete rule set — this replaces the document, it does not merge — including the `revision` you last read.

```json
{
  "id": "8f9c1e2a-...",
  "name": "Family Movies",
  "media": "movie",
  "revision": 1,
  "movies": { "...": "the complete, updated filter payload" }
}
```

A successful update ignores the `revision` you sent for storage purposes (only for comparison) and returns the saved rule set with `revision` incremented by one. Sending a `revision` that doesn't match the server's current value returns `409 Conflict` with `{"error": "Rule set changed since it was opened; reload and try again"}` — re-fetch the rule set and reapply your change rather than retrying blindly, since something else changed it in the meantime. Changing `id` in the body (when non-empty) returns `400`.

Default rule sets (`default-movies`, `default-shows`) can be edited but not deleted.

## Delete a rule set

`DELETE /v1/rule-sets/:id`

Returns `204 No Content` on success.

- A default rule set (`default-movies` or `default-shows`) returns `409 Conflict`: `{"error": "Default rule sets cannot be deleted"}`.
- A rule set still assigned to one or more jobs returns `409 Conflict`: `{"error": "Rule set is used by N job(s)"}` — reassign those jobs first.
- An `:id` that doesn't exist at all currently returns `500 Internal Server Error` with `{"error": "rule set not found"}`, not `404` — check for this string rather than relying on the status code if you need to distinguish "already gone" from a real server error.

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

Movie IDs are TMDB IDs; show IDs are TheTVDB IDs (not TMDB show IDs) — see [Rules](/concepts/rules/#blockedallowed-title-ids--movies-use-tmdb-shows-use-tvdb). `PUT` replaces the complete document with no partial-update option and no validation of the IDs themselves (a made-up ID is stored without error, it just never matches anything). Read the current document first if you're changing only one list, so you don't drop the others.

## Make a job-specific copy

`POST /v1/jobs/:id/customize-rules`

Clones the job's effective (currently assigned) rule set, names the copy `"<job name> Rules"` (appending a number if that name is taken), assigns the new copy to the job, and returns both:

```json
{
  "job": { "...": "the job, now pointing at the new rule_set_id" },
  "rule_set": { "...": "the cloned rule set, revision 1" },
  "usage_count": 1
}
```

Returns `201 Created`. This endpoint is for dynamic jobs; it returns `404` if the job ID doesn't exist, and `409` if the job's currently assigned rule set can't be resolved (for example, it was deleted out from under the job).

## Errors

| Status | Meaning |
|---|---|
| `400` | Invalid JSON body, or a rule set failed validation (see the list under Create) |
| `404` | Job not found (`customize-rules`), or rule set not found on `PUT` |
| `409` | Duplicate rule set ID, stale `revision` on `PUT`, a protected default or in-use rule set on `DELETE`, or an unresolvable rule set on `customize-rules` |
| `500` | The configuration failed to save or reload, or — on `DELETE` only — the `:id` did not exist |
