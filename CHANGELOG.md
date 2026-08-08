# Changelog

All notable changes to Blockbusterr are documented here. The project follows semantic versioning.

## [Unreleased]

## [2.0.0-beta.6] - 2026-08-08

### Changed

- Updated Fiber and frontend dependencies to patched releases and cleared open dependency alerts.
- Added deployment guidance for common self-hosted container platforms.

## [2.0.0-beta.5] - 2026-08-07

### Added

- Added an unattended v1-to-v2 upgrade with automatic backups, ownership repair, and fail-closed migration checks.
- Added CI coverage that upgrades the published v1.5.0 image into the candidate v2 image.

## [2.0.0-beta.4] - 2026-08-05

### Added

- Added instance-wide dry-run mode across scheduled, manual, ranked-selection, and legacy execution paths.

### Changed

- Upgraded Go and restored lint, security, dependency, asset, documentation, and race-test release gates.

## [2.0.0-beta.3] - 2026-08-05

### Fixed

- Serialized job execution and made ranked selection deliver its reviewed snapshot.
- Made configuration restore atomic and delivery memory independent from Activity retention.
- Hardened provider pagination, identity, enrichment failures, duplicate classification, and token refresh.

## [2.0.0-beta.2] - 2026-08-04

### Fixed

- Cron jobs now wait for their next occurrence instead of running at startup.
- Activity job filtering now uses enabled jobs and stable job IDs.

## [2.0.0-beta.1] - 2026-08-04

### Added

- Reusable movie and show rule sets with assignment counts and revisions.
- Per-job rule assignment and one-click job-specific rule copies.
- Universal movie and show title exceptions.
- Country, language, genre, certification, network, keyword, provider ID, year, runtime, rating, and vote rules.
- Typed list and watchlist jobs for TMDB, Trakt, MDBList, and experimental public Letterboxd lists.
- Curated recipes and recommendations for common movie and show discovery jobs.
- Dynamic jobs with explicit TMDB, Simkl, or Trakt discovery sources where supported.
- Redesigned Jobs, Rules, Activity, Job Runs, and Settings workflows.
- Structured Job Run flow and outcome distribution with title-level decision details.
- Ranked selection cycles that compare candidates across participating jobs before delivery.
- Preview-first job creation, provider inspection, and disabled-by-default imports and recipes.
- Repeat handling for titles removed by another application.
- Optional owner authentication with `BLOCKBUSTERR_AUTH_TOKEN`.
- Readiness checks, local compiled assets, configuration backup/restore, complete log clearing, and improved shutdown behavior.
- Automatic migration of legacy jobs and embedded custom filters.
- Enforced per-job and global delivery budgets with rolling movie/show periods.
- Versioned portable job bundles that include their reusable rules and import disabled with fresh IDs.

### Changed

- Filters are now presented as reusable Rules while the existing documentation URL remains compatible.
- Activity terminology is standardized as Activity Entries and Job Runs.
- The application no longer requires Trakt when enabled jobs use another configured provider.
- Direct Radarr/Sonarr additions and Jellyseerr/Seerr requests remain distinct in Job Runs while combined totals are labeled Delivered.
- The settings and job editors now expose delivery mode, rule assignment, safety limits, and provider readiness more clearly.
- Documentation is versioned so stable v1 and beta v2 guidance can coexist.
- Documentation now describes Blockbusterr as the discovery and decision layer in a media automation stack.
- Sonarr lookup results now require matching provider identity before a show is treated as already present.

### Fixed

- Graceful shutdown is bounded so repeated interrupts cannot leave Blockbusterr hanging indefinitely.
- Ranked-selection participation is stored per job instead of leaking between job editors.
- Activity empty states now reflect existing Activity Entries and Job Runs correctly.
- Legacy v1 jobs and custom filters migrate into the v2 job and Rules model.
- Provider readiness warnings identify the affected enabled jobs instead of reporting misleading totals.

### Removed

- Nonfunctional aggregation-era Global Limits code, replaced by delivery-time enforcement.
- Unused template showcase components, frontend definitions, API-error wrappers, and duplicate documentation assets.

### Upgrade notes

- Back up the configuration and database together before upgrading.
- Review migrated jobs and rule assignments, then preview every enabled job.
- The beta image is published as `ghcr.io/mahcks/blockbusterr:v2.0.0-beta.1` and `ghcr.io/mahcks/blockbusterr:latest-beta`; `latest` remains on v1 until the stable release.
- See the [v2 upgrade guide](https://blockbusterr.dev/v2/getting-started/upgrading-to-v2/).

[Unreleased]: https://github.com/Mahcks/blockbusterr/compare/v2.0.0-beta.6...HEAD
[2.0.0-beta.6]: https://github.com/Mahcks/blockbusterr/compare/v2.0.0-beta.5...v2.0.0-beta.6
[2.0.0-beta.5]: https://github.com/Mahcks/blockbusterr/compare/v2.0.0-beta.4...v2.0.0-beta.5
[2.0.0-beta.4]: https://github.com/Mahcks/blockbusterr/compare/v2.0.0-beta.3...v2.0.0-beta.4
[2.0.0-beta.3]: https://github.com/Mahcks/blockbusterr/compare/v2.0.0-beta.2...v2.0.0-beta.3
[2.0.0-beta.2]: https://github.com/Mahcks/blockbusterr/compare/v2.0.0-beta.1...v2.0.0-beta.2
[2.0.0-beta.1]: https://github.com/Mahcks/blockbusterr/compare/v1.5.0...v2.0.0-beta.1
