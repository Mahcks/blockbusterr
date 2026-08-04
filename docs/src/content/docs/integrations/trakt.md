---
title: Trakt
description: Configure Trakt as an optional Blockbusterr discovery provider.
---

Trakt is one of Blockbusterr's discovery sources. It is optional when every enabled job uses TMDB or Simkl.

## Create credentials

1. Sign in to Trakt and create an API application.
2. Copy the client ID and, when provided, client secret.
3. Enter them under **Settings → Discovery → Trakt**.
4. Test and save the connection.

Blockbusterr uses application credentials for public discovery endpoints; it does not require an interactive user OAuth session for the supported jobs.

```yaml
trakt:
  client_id: your_client_id
  client_secret: your_client_secret
```

## Supported discovery

Trakt supplies several trending, popular, watched, collected, played, favorited, anticipated, and metadata paths. Availability depends on the selected media and job type; the Jobs UI only shows valid combinations.

## Troubleshooting

- Confirm the client ID is copied without surrounding whitespace.
- Confirm the Blockbusterr container can reach `api.trakt.tv`.
- Check Jobs after saving; Trakt should appear only for supported types.
- Use Preview to verify discovery before enabling delivery.
