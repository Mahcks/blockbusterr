---
title: Upgrading to v2
description: Safely migrate a Blockbusterr v1 installation to jobs, reusable rules, and the v2 interface.
---

v2 keeps existing configuration readable while moving active automation to dynamic jobs and reusable rule sets. Perform the upgrade once, verify the result, then manage the installation through the v2 UI.

## Before upgrading

1. Stop Blockbusterr and copy the entire configuration and data directory.
2. Record the currently running image tag or binary version.
3. Confirm Radarr, Sonarr, and Jellyseerr or Seerr are healthy.
4. Keep the backup outside the mounted Blockbusterr directory.

The backup must include the YAML configuration and SQLite database. Do not test an upgrade against your only copy.

## Upgrade

1. Pull the v2 image or replace the binary.
2. Optionally set `BLOCKBUSTERR_AUTH_TOKEN` to a random value of at least 32 characters and save it in your password manager.
3. Ensure the mounted data directory is writable by container UID/GID `10001:10001` (for example, `sudo chown -R 10001:10001 ./data`).
4. Start Blockbusterr with the existing configuration and data mounts.
5. If authentication is enabled, sign in with username `blockbusterr` and the owner token. Open **Settings** and confirm discovery and delivery connections.
6. Open **Jobs**. If the legacy migration banner appears, review the count and choose **Upgrade jobs**.
7. Open **Rules** and review Default Movies, Default Shows, migrated job-specific rule sets, and title exceptions.
8. Preview every enabled job before running it.

The migration creates dynamic jobs, assigns media-compatible default rules, preserves supported scheduling and delivery overrides, saves the configuration, and disables migrated legacy entries. Embedded custom job filters become job-specific rule sets.

## Important changes

- Filters are presented as reusable **Rules**.
- Each job references exactly one movie or show rule set.
- Universal title exceptions are managed separately and take precedence.
- Activity is split into **Activity Entries** and **Job Runs**.
- Discovery source selection is explicit per job.
- Global Limits now enforce successful deliveries. Existing movie/show values carry forward; the obsolete `sync` period becomes a rolling 24-hour period.
- Release containers run as UID/GID `10001:10001`; `BLOCKBUSTERR_AUTH_TOKEN` optionally protects the UI and API.

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
2. Restore the backed-up configuration and database together.
3. Restore the previous image tag or binary.
4. Start the service and verify its integrations.

Do not combine a v2-written configuration with an older database or vice versa. Restore the pair from the same backup.

## After verification

Keep the backup through at least one complete schedule cycle. Once satisfied, use the v2 UI for future jobs and rules; legacy YAML remains a compatibility input, not the recommended editing model.
