---
title: Simkl Integration
description: Configure Simkl trending and most-watched discovery
---

Simkl is an optional discovery source for Trending, Popular, Smart Popular, and Most Watched movie and TV jobs.

## Configuration

Create an application in [Simkl Developer Settings](https://simkl.com/settings/developer/) and copy its client ID:

```yaml
simkl:
  client_id: "your_simkl_client_id"
```

Select **Simkl** as the Discovery Source on a supported dynamic job. Simkl Most Watched jobs support weekly and monthly periods, and Simkl feeds return up to 500 items.

Blockbusterr labels Simkl-sourced previews and activity entries to satisfy Simkl's attribution requirements. Review the [Simkl API rules](https://api.simkl.org/api-rules) before commercial deployment.
