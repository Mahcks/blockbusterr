package jobs

import (
	"fmt"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

func effectiveRepeatPolicy(jobPolicy, globalPolicy string) enums.RepeatPolicy {
	policy := enums.RepeatPolicy(jobPolicy)
	if policy == enums.RepeatPolicyInherit {
		policy = enums.RepeatPolicy(globalPolicy)
	}
	if !policy.IsValid(false) {
		return enums.RepeatPolicy90Days
	}
	return policy
}

func repeatCooldown(policy enums.RepeatPolicy) time.Duration {
	switch policy {
	case enums.RepeatPolicy30Days:
		return 30 * 24 * time.Hour
	case enums.RepeatPolicy90Days:
		return 90 * 24 * time.Hour
	case enums.RepeatPolicy180Days:
		return 180 * 24 * time.Hour
	default:
		return 0
	}
}

func repeatSkipReason(cfg *config.Config, db *database.Database, jobPolicy, mediaType string, tmdbID, tvdbID int, now time.Time) (string, error) {
	policy := effectiveRepeatPolicy(jobPolicy, cfg.Jobs.RepeatPolicy)
	if policy == enums.RepeatPolicyImmediate || db == nil {
		return "", nil
	}
	deliveredAt, found, err := db.LatestSuccessfulDelivery(mediaType, tmdbID, tvdbID)
	if err != nil || !found {
		return "", err
	}
	return repeatSkipReasonForDelivery(policy, deliveredAt, now), nil
}

func repeatSkipReasonForDelivery(policy enums.RepeatPolicy, deliveredAt, now time.Time) string {
	if policy == enums.RepeatPolicyNever {
		return "Previously delivered; repeat handling is set to never"
	}
	remaining := time.Until(deliveredAt.Add(repeatCooldown(policy)))
	if !now.IsZero() {
		remaining = deliveredAt.Add(repeatCooldown(policy)).Sub(now)
	}
	if remaining <= 0 {
		return ""
	}
	days := int((remaining + 24*time.Hour - 1) / (24 * time.Hour))
	unit := "days"
	if days == 1 {
		unit = "day"
	}
	return fmt.Sprintf("Previously delivered; eligible again in %d %s", days, unit)
}
