# Blockbusterr

**Automate your media library with smart content discovery from Trakt.**

Blockbusterr automatically adds trending, popular, and highly-rated movies and TV shows from Trakt to your Radarr/Sonarr library. Configure once, let it run on a schedule, and never manually manage Trakt lists again.

[![GitHub release](https://img.shields.io/github/v/release/Mahcks/blockbusterr)](https://github.com/Mahcks/blockbusterr/releases)
[![GitHub Packages](https://img.shields.io/badge/GitHub%20Packages-ghcr.io-blue)](https://github.com/mahcks/blockbusterr/pkgs/container/blockbusterr)
[![License](https://img.shields.io/github/license/Mahcks/blockbusterr)](LICENSE)
[![Documentation](https://img.shields.io/badge/docs-live-blue)](https://mahcks.github.io/blockbusterr/)

---

## Documentation

**Full documentation available at [mahcks.github.io/blockbusterr](https://mahcks.github.io/blockbusterr/)**

- [Quick Start Guide](https://mahcks.github.io/blockbusterr/getting-started/quickstart/)
- [Installation Methods](https://mahcks.github.io/blockbusterr/getting-started/installation/)
- [Configuration](https://mahcks.github.io/blockbusterr/getting-started/configuration/)
- [Jobs Overview](https://mahcks.github.io/blockbusterr/concepts/jobs/)
- [Filters & Scoring](https://mahcks.github.io/blockbusterr/concepts/filters/)
- [Real-World Examples](https://mahcks.github.io/blockbusterr/examples/use-cases/)
- [API Reference](https://mahcks.github.io/blockbusterr/api/overview/)

---

## Quick Start

**Get running in 60 seconds:**

```bash
docker run -d \
  --name blockbusterr \
  -p 9090:9090 \
  -v $(pwd)/data:/app/data \
  ghcr.io/mahcks/blockbusterr:latest
```

Then open `http://localhost:9090` and configure your services.

**[→ Full Quick Start Guide](https://mahcks.github.io/blockbusterr/getting-started/quickstart/)**

---

## Features

- **17 Job Types** - Trending, popular, anticipated, favorited, box office, and more
- **Smart Filtering** - Genre, certification, runtime, year, language, country, keywords
- **Weighted Scoring** - Combine IMDb, Trakt, TMDB, and Metacritic ratings
- **Two Integration Modes** - Direct to Radarr/Sonarr or via Jellyseerr for approval
- **Activity Tracking** - See what was added, when, and why with visual logs
- **Job Preview** - Test configurations before enabling to avoid surprises
- **Unified Dashboard** - Manage movies and TV shows in one place
- **Smart Jobs** - Adaptive scoring that balances popularity with quality

---

## 📸 Screenshots

### Dashboard & Configuration
![Configuration](reference-docs/images/configuration.png)

### Job Preview & Management
![Jobs Preview](reference-docs/images/jobs_preview.png)

### Activity Log
![Activity Log](reference-docs/images/activity_log_preview.png)

---

## Why Blockbusterr?

**The Problem:** Managing Trakt lists in Radarr/Sonarr is tedious - configure each list separately, no filtering, no limits, no unified tracking.

**The Solution:** One dashboard to rule them all. Set filters once, configure jobs with limits, see everything that's added in one activity log.

| Feature | Manual Trakt Lists | Blockbusterr |
|---------|-------------------|--------------|
| Setup | Configure in each *arr app | One unified dashboard |
| Limits | All or nothing | Top N items per job |
| Filters | None | Genre, rating, runtime, language, etc. |
| Activity Log | Check each app separately | Unified log with posters |
| Preview | No preview capability | Test before enabling |

---

## Use Cases

- **Family Server** - Block R-rated content, require G/PG/PG-13 only
- **Quality Curator** - Only add movies scoring 80+, minimum 50k IMDb votes
- **Genre Specialist** - Sci-fi and fantasy only, no comedies or romance
- **Completionist** - Add everything trending with minimal filtering

**[→ See Real-World Examples](https://mahcks.github.io/blockbusterr/examples/use-cases/)**

---

## Installation

**Docker Compose** (Recommended):

```yaml
version: '3.8'
services:
  blockbusterr:
    image: ghcr.io/mahcks/blockbusterr:latest
    container_name: blockbusterr
    ports:
      - "9090:9090"
    volumes:
      - ./data:/app/data
    environment:
      - TZ=America/New_York
    restart: unless-stopped
```

**Other methods:** Binary, from source, with full stack → **[Installation Guide](https://mahcks.github.io/blockbusterr/getting-started/installation/)**

---

## Configuration

Blockbusterr can be configured via:
- **Web UI** - `http://localhost:9090` (easiest)
- **config.yaml** - Mount as volume or edit in container
- **Environment Variables** - For Docker deployments

**[→ Configuration Guide](https://mahcks.github.io/blockbusterr/getting-started/configuration/)**

---

## How It Works

1. **Jobs run on schedule** (cron) - e.g., "Trending Movies" every 6 hours
2. **Fetch content from Trakt** - Trending, popular, anticipated, etc.
3. **Apply filters** - Genre, rating, runtime, certification, language
4. **Calculate scores** - Weighted average of IMDb/Trakt/TMDB ratings
5. **Check threshold** - Only content scoring above threshold proceeds
6. **Add to library** - Direct to Radarr/Sonarr or create Jellyseerr request
7. **Log activity** - Track what was added with posters and metadata

**[→ Learn About Jobs](https://mahcks.github.io/blockbusterr/concepts/jobs/)** | **[→ Filters & Scoring](https://mahcks.github.io/blockbusterr/concepts/filters/)**

---

## Contributing

Contributions welcome! Please open an issue or PR.

**Development:**

```bash
git clone https://github.com/mahcks/blockbusterr.git
cd blockbusterr
make dev
```

---

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

## Support

If Blockbusterr saves you time, consider giving it a star!

[![Star History Chart](https://api.star-history.com/svg?repos=Mahcks/blockbusterr&type=Date)](https://star-history.com/#Mahcks/blockbusterr&Date)


**Made for the *arr community** 

