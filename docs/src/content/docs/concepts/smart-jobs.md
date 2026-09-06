---
title: Smart Jobs
description: Use adaptive rating thresholds with popularity-aware discovery.
---

Smart Popular jobs adjust the rating threshold according to a candidate's popularity. Highly popular titles may clear a lower quality bar, while less-visible titles must earn a stronger rating to pass.

## Settings

| Setting | Purpose |
|---|---|
| `base_min_rating` | Baseline rating around which the adaptive threshold moves |
| `adjustment_factor` | How strongly popularity changes that threshold |
| `limit` | Maximum candidates discovered before evaluation |
| `rule_set_id` | Movie or show rules applied alongside adaptive scoring |

Configure Smart Popular from the Jobs UI. The editor stores these values in the same dynamic job model as every other discovery type.

## Rules still apply

Smart rating is one decision input, not a replacement for the assigned rule set. Genre, language, keyword, year, runtime, vote, and title-exception decisions still apply.

Avoid duplicating the adaptive threshold with an unnecessarily strict minimum-rating boundary. Use a minimum-votes boundary when you want to reject ratings based on very small samples.

## Tune safely

1. Begin with a moderate baseline and adjustment factor.
2. Preview the job several times.
3. Inspect decision details for unexpected candidates.
4. Change one setting at a time.
5. Review the Job Run distribution after enabling it.

A higher baseline is stricter everywhere. A higher adjustment factor makes popular titles more forgiving and less-popular titles stricter.

## Troubleshooting

- **Too many weak titles:** increase `base_min_rating`, add a minimum-votes boundary, or tighten the assigned rules.
- **Weak blockbusters pass:** lower `adjustment_factor` or raise `base_min_rating`.
- **Too few results:** inspect whether the assigned rule set is rejecting candidates before changing smart settings.
- **Unexpected delivery:** open the title's Activity Entry and inspect the complete decision details.
