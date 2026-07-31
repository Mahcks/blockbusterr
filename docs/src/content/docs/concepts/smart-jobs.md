---
title: Smart Jobs
description: Adaptive rating thresholds based on popularity
---

import { Aside } from '@astrojs/starlight/components';

Smart Jobs use adaptive rating thresholds that adjust based on content popularity, allowing you to get mainstream blockbusters without sacrificing quality for lesser-known content.

## The Problem

Traditional static rating filters have a fundamental issue:

**Popular blockbusters** with thousands of votes often have "average" ratings (6-7/10) even when they're great movies. Think Marvel films, big-budget action movies, or mainstream hits - they appeal to everyone, including critics who vote them down.

**Hidden gems** with fewer votes need higher ratings to prove their quality. A movie with only 200 votes scoring 7.5 is likely genuinely good.

**With static filters, you face a dilemma:**
- Set `min_rating: 7.0` → Miss great blockbusters rated 6.5-6.9
- Set `min_rating: 6.0` → Get mediocre low-popularity content

## The Solution

Smart Jobs **automatically adjust** rating requirements based on popularity percentile:

- **High popularity** (90th percentile) → Requires **lower rating** (6.2)
- **Medium popularity** (50th percentile) → Requires **baseline rating** (7.0)
- **Low popularity** (10th percentile) → Requires **higher rating** (7.8)

<Aside type="tip">
  Think of it as "innocent until proven guilty" for popular content, and "prove yourself" for obscure content.
</Aside>

## How It Works

### The Formula

```
adjusted_threshold = base_min_rating - ((percentile - 0.5) × adjustment_factor)
```

Where:
- `base_min_rating` - Your baseline quality requirement (e.g., 7.0)
- `percentile` - Item's popularity percentile (0.0 to 1.0)
- `adjustment_factor` - How aggressively to adjust (0.5 to 3.0)

### Example Calculation

With `base_min_rating: 7.0` and `adjustment_factor: 2.0`:

| Popularity Percentile | Calculation | Required Rating |
|----------------------|-------------|-----------------|
| 90th (very popular) | 7.0 - ((0.9 - 0.5) × 2.0) | **6.2** |
| 70th (popular) | 7.0 - ((0.7 - 0.5) × 2.0) | **6.6** |
| 50th (average) | 7.0 - ((0.5 - 0.5) × 2.0) | **7.0** |
| 30th (less popular) | 7.0 - ((0.3 - 0.5) × 2.0) | **7.4** |
| 10th (obscure) | 7.0 - ((0.1 - 0.5) × 2.0) | **7.8** |

## Configuration

### Basic Configuration

```yaml
jobs:
  smart_popular_movies:
    enabled: true
    sync_interval: "0 */6 * * *"
    limit: 20
    base_min_rating: 7.0      # Baseline rating requirement
    adjustment_factor: 2.0     # How aggressive (default: 1.0)
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `base_min_rating` | float | `7.0` | Baseline rating for average popularity |
| `adjustment_factor` | float | `1.0` | Aggressiveness (0.5 = mild, 2.0 = aggressive, 3.0 = very aggressive) |
| `limit` | integer | `10` | Maximum items to add |
| `sync_interval` | string | global default | Duration or cron schedule |

### Global Filters

Smart jobs use the same global movie or show filters as every other job. Per-job filters are not supported.

```yaml
smart_popular_movies:
  enabled: true
  limit: 20
  base_min_rating: 7.0
  adjustment_factor: 2.0
```

<Aside type="caution">
  Don't use `min_rating` filter with Smart Jobs - it's handled by the adaptive system.
</Aside>

## Adjustment Factor Guide

The `adjustment_factor` controls how aggressively ratings are adjusted.

### Conservative (0.5 - 1.0)

Subtle adjustments, closer to static filtering.

```yaml
adjustment_factor: 0.5
# 90th percentile: 7.0 - ((0.9 - 0.5) × 0.5) = 6.8
# 10th percentile: 7.0 - ((0.1 - 0.5) × 0.5) = 7.2
```

**Use when:** You trust your baseline rating and want minor flexibility.

### Moderate (1.0 - 1.5)

Balanced adjustments, good for most use cases.

```yaml
adjustment_factor: 1.0
# 90th percentile: 7.0 - ((0.9 - 0.5) × 1.0) = 6.6
# 10th percentile: 7.0 - ((0.1 - 0.5) × 1.0) = 7.4
```

**Use when:** You want a balanced approach (recommended starting point).

### Aggressive (2.0 - 2.5)

Significant adjustments, prioritizes mainstream content.

```yaml
adjustment_factor: 2.0
# 90th percentile: 7.0 - ((0.9 - 0.5) × 2.0) = 6.2
# 10th percentile: 7.0 - ((0.1 - 0.5) × 2.0) = 7.8
```

**Use when:** You want to capture popular blockbusters with "average" ratings.

### Very Aggressive (3.0+)

Maximum adjustments, heavily favors popularity.

```yaml
adjustment_factor: 3.0
# 90th percentile: 7.0 - ((0.9 - 0.5) × 3.0) = 5.8
# 10th percentile: 7.0 - ((0.1 - 0.5) × 3.0) = 8.2
```

**Use when:** You primarily want mainstream content and trust popular opinion.

## Available Smart Jobs

Currently, two smart jobs are available:

### smart_popular_movies

Adaptive thresholds for popular movies.

```yaml
smart_popular_movies:
  enabled: true
  sync_interval: "0 6 * * *"
  limit: 20
  base_min_rating: 7.0
  adjustment_factor: 2.0
```

### smart_popular_shows

Adaptive thresholds for popular TV shows.

```yaml
smart_popular_shows:
  enabled: true
  sync_interval: "0 12 * * *"
  limit: 15
  base_min_rating: 7.0
  adjustment_factor: 1.5
```

## Real-World Examples

### Example 1: Mainstream Movies

**Goal:** Get popular blockbusters without quality compromise.

```yaml
smart_popular_movies:
  enabled: true
  limit: 25
  base_min_rating: 7.0
  adjustment_factor: 2.0
```

**Results:**
- Dune: Part Two (8.7 rating, 95th percentile) ✓ Added (requires 6.1)
- Fast X (6.8 rating, 88th percentile) ✓ Added (requires 6.3)
- Unknown Indie (7.2 rating, 15th percentile) ✗ Filtered (requires 7.7)

### Example 2: Balanced Discovery

**Goal:** Mix of popular and quality content.

```yaml
smart_popular_movies:
  enabled: true
  limit: 20
  base_min_rating: 6.8
  adjustment_factor: 1.0
```

**Results:** Moderate adjustment curve, good mix of mainstream and quality.

### Example 3: Quality First

**Goal:** High-quality content with some flexibility for popular items.

```yaml
smart_popular_movies:
  enabled: true
  limit: 15
  base_min_rating: 7.5
  adjustment_factor: 0.8
```

**Results:** Strict baseline with minor adjustments for popularity.

## Comparing Regular vs. Smart Jobs

### Regular Popular Job (Static Filter)

```yaml
popular_movies:
  enabled: true
  limit: 20
```

**Result:** Misses blockbusters rated 6.5-6.9, regardless of popularity.

### Smart Popular Job (Adaptive Filter)

```yaml
smart_popular_movies:
  enabled: true
  limit: 20
  base_min_rating: 7.0
  adjustment_factor: 2.0  # Adaptive threshold
```

**Result:** Gets blockbusters rated 6.2+ if popular, requires 7.8+ for obscure content.

## Preview Smart Jobs

Always preview smart jobs to understand the adaptive behavior:

```bash
curl -X POST http://localhost:9090/v1/jobs/preview/smart-popular-movies
```

The response shows:
- Each item's popularity percentile
- Calculated rating threshold for that item
- Whether it passed or failed

## Best Practices

### Start with Defaults

Begin with moderate settings and adjust based on results:

```yaml
base_min_rating: 7.0
adjustment_factor: 1.0
```

### Use min_votes

Set the global movie or show `min_votes` threshold to improve rating reliability:

```yaml
filters:
  movies:
    min_votes: 1000
```

### Preview Extensively

Smart jobs behave differently than static filters. Preview multiple times to understand the results.

### Monitor Activity

Check what gets added and adjust `base_min_rating` or `adjustment_factor` as needed.

### Combine with Other Filters

Smart rating is just one filter. Use genre, language, and keyword filters as normal:

```yaml
smart_popular_movies:
  base_min_rating: 7.0
  adjustment_factor: 2.0
```

## Troubleshooting

### Getting Too Much Low-Quality Content?

**Increase base_min_rating:**
```yaml
base_min_rating: 7.5  # Up from 7.0
```

**Or decrease adjustment_factor:**
```yaml
adjustment_factor: 1.0  # Down from 2.0
```

### Missing Popular Blockbusters?

**Increase adjustment_factor:**
```yaml
adjustment_factor: 2.5  # Up from 2.0
```

**Or decrease base_min_rating:**
```yaml
base_min_rating: 6.5  # Down from 7.0
```

### Too Many Obscure Items?

**Increase the global movie vote threshold:**
```yaml
filters:
  movies:
    min_votes: 2000
```

## Next Steps

- Learn about [Regular Jobs](/concepts/jobs/)
- Configure [Advanced Filters](/concepts/filters/)
- Understand [Integration Modes](/concepts/integration-modes/)
