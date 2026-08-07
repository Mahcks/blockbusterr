---
title: Upgrading to v2
description: Safely migrate a Blockbusterr v1 installation to jobs, reusable rules, and the v2 interface.
---

v2 upgrades a supported v1 installation on first startup. It preserves the original configuration, converts enabled legacy jobs to dynamic jobs and reusable rules, upgrades SQLite, and then starts normally. No migration button or manual ownership change is required for the standard Docker installation.

## Before upgrading

1. Stop Blockbusterr and copy the entire configuration and data directory.
2. Record the currently running image tag or binary version.
3. Confirm Radarr, Sonarr, and Jellyseerr or Seerr are healthy.
4. Keep the backup outside the mounted Blockbusterr directory.

The backup must include the YAML configuration and SQLite database. Do not test an upgrade against your only copy.

## Upgrade

1. Pull the v2 image or replace the binary.
2. Optionally set `BLOCKBUSTERR_AUTH_TOKEN` to a random value of at least 32 characters and save it in your password manager.
3. Docker users should leave the container user unset. The v2 entrypoint will repair ownership left by root-running v1 containers only for the mounted data directory and Blockbusterr's known writable files, then immediately run the application as UID/GID `10001:10001`.
4. Start Blockbusterr with the existing configuration and data mounts.
5. If authentication is enabled, sign in with username `blockbusterr` and the owner token. Open **Settings** and confirm discovery and delivery connections.
6. Open **Jobs** and confirm the enabled v1 jobs appear with their schedules and delivery settings.
7. Open **Rules** and review Default Movies, Default Shows, migrated job-specific rule sets, and title exceptions.
8. Preview every enabled job before running it manually.

Before changing data, the Docker entrypoint saves the stopped v1 SQLite database and configuration under `data/backups/pre-v2/`. Configuration migration also preserves the original file as `config.yaml.backup`. The migration creates dynamic jobs, assigns media-compatible default rules, preserves supported scheduling and delivery overrides, saves the configuration atomically, and disables migrated legacy entries. Embedded custom job filters become job-specific rule sets.

If backup, ownership repair, configuration migration, or database initialization fails, Blockbusterr exits before starting automation and leaves the original data available for recovery.

## Important changes

- Filters are presented as reusable **Rules**.
- Each job references exactly one movie or show rule set.
- Universal title exceptions are managed separately and take precedence.
- Activity is split into **Activity Entries** and **Job Runs**.
- Discovery source selection is explicit per job.
- Global Limits now enforce successful deliveries. Existing movie/show values carry forward; the obsolete `sync` period becomes a rolling 24-hour period.
- Release containers run as UID/GID `10001:10001`; `BLOCKBUSTERR_AUTH_TOKEN` optionally protects the UI and API.
- The Docker entrypoint performs the one-time v1 ownership repair without changing file modes or granting broad permissions. If you override the container `user`, the entrypoint cannot perform that repair; stop Blockbusterr and run `sudo chown -R 10001:10001 <mounted-data-directory>` on the Docker host instead.
- Cron schedules wait for their next scheduled time after startup. Duration schedules retain their existing run-on-start behavior.

## Verify the upgrade

For each enabled job:

1. Confirm type, source, media, limit, period, and assigned rules.
2. Confirm any schedule or delivery-mode override.
3. Run Preview and inspect accepted and rejected decisions.
4. Run one job manually.
5. Confirm the Job Run completes and its Activity Entries match the delivery target.

Also test configuration backup/download after migration.

## Roll back

1. Stop Blockbusterr.
2. Restore the externally backed-up configuration and database together. If needed, the automatic pre-upgrade copies are under `data/backups/pre-v2/` and `data/config.yaml.backup`.
3. Restore the previous image tag or binary.
4. Start the service and verify its integrations.

Do not combine a v2-written configuration with an older database or vice versa. Restore the pair from the same backup.

## After verification

Keep the backup through at least one complete schedule cycle. Once satisfied, use the v2 UI for future jobs and rules; legacy YAML remains a compatibility input, not the recommended editing model.
