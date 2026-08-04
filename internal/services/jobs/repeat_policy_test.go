package jobs

import (
	"strings"
	"testing"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
	"github.com/mahcks/blockbusterr/pkg/enums"
)

func TestRepeatPolicyUsesSuccessfulDeliveryHistory(t *testing.T) {
	db, err := database.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Now()
	for _, log := range []database.ActivityLog{
		{Timestamp: now.Add(-2 * time.Hour), MediaType: "movie", Title: "Dry run", TMDBID: 10, Status: string(enums.ActivityStatusAdded), Message: "[DRY RUN] Would be added to Radarr"},
		{Timestamp: now.Add(-10 * 24 * time.Hour), MediaType: "movie", Title: "Delivered", TMDBID: 20, Status: string(enums.ActivityStatusAdded)},
	} {
		if err := db.LogActivity(log); err != nil {
			t.Fatal(err)
		}
	}

	cfg := &config.Config{}
	cfg.Jobs.RepeatPolicy = string(enums.RepeatPolicy90Days)
	if reason, err := repeatSkipReason(cfg, db, "", "movie", 10, 0, now); err != nil || reason != "" {
		t.Fatalf("dry run history blocked delivery: reason=%q err=%v", reason, err)
	}
	if reason, err := repeatSkipReason(cfg, db, "", "movie", 20, 0, now); err != nil || !strings.Contains(reason, "80 days") {
		t.Fatalf("cooldown result: reason=%q err=%v", reason, err)
	}
	if reason, err := repeatSkipReason(cfg, db, string(enums.RepeatPolicyImmediate), "movie", 20, 0, now); err != nil || reason != "" {
		t.Fatalf("immediate override blocked delivery: reason=%q err=%v", reason, err)
	}
	if reason, err := repeatSkipReason(cfg, db, string(enums.RepeatPolicyNever), "movie", 20, 0, now); err != nil || !strings.Contains(reason, "never") {
		t.Fatalf("never override result: reason=%q err=%v", reason, err)
	}
}
