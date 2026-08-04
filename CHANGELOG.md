# Changelog

All notable changes to Blockbusterr are documented here. The project follows semantic versioning.

## [Unreleased]

### Added

- Reusable movie and show rule sets with assignment counts and revisions.
- Per-job rule assignment and one-click job-specific rule copies.
- Universal movie and show title exceptions.
- Dynamic jobs with explicit TMDB, Simkl, or Trakt discovery sources where supported.
- Redesigned Jobs, Rules, Activity, Job Runs, and Settings workflows.
- Structured Job Run flow and outcome distribution with title-level decision details.
- Readiness checks, local compiled assets, configuration backup/restore, and improved shutdown behavior.
- Automatic migration of legacy jobs and embedded custom filters.
- Enforced per-job and global delivery budgets with rolling movie/show periods.
- Versioned portable job bundles that include their reusable rules and import disabled with fresh IDs.
- A typed list/watchlist job contract that reuses the existing preview, rules, Activity, and delivery pipeline.

### Changed

- Filters are now presented as reusable Rules while the existing documentation URL remains compatible.
- Activity terminology is standardized as Activity Entries and Job Runs.
- The application no longer requires Trakt when enabled jobs use another configured provider.
- Documentation now describes Blockbusterr as the discovery and decision layer in a media automation stack.
- Sonarr lookup results now require matching provider identity before a show is treated as already present.

### Removed

- Nonfunctional aggregation-era Global Limits code, replaced by delivery-time enforcement.
- Unused template showcase components, frontend definitions, API-error wrappers, and duplicate documentation assets.

### Upgrade notes

- Back up the configuration and database together before upgrading.
- Review migrated jobs and rule assignments, then preview every enabled job.
- See the [v2 upgrade guide](https://blockbusterr.dev/getting-started/upgrading-to-v2/).

[Unreleased]: https://github.com/Mahcks/blockbusterr/compare/v1.5.0...HEAD
