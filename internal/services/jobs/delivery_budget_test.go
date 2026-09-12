package jobs

import (
	"testing"
	"time"

	"github.com/mahcks/blockbusterr/config"
	"github.com/mahcks/blockbusterr/internal/database"
)

func TestDeliveryBudgetEnforcesGlobalAndPerRunLimits(t *testing.T) {
	db, err := database.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	cfg := &config.Config{}
	cfg.Jobs.GlobalLimitMovies = 1
	cfg.Jobs.GlobalPeriod = "daily"
	first := newDeliveryBudget(cfg, db, "job-a", 1, 0, false)
	reservationID, allowed, reason := first.reserve("movie")
	if !allowed || reservationID == 0 || reason != "" {
		t.Fatalf("first reservation = (%d, %t, %q)", reservationID, allowed, reason)
	}
	second := newDeliveryBudget(cfg, db, "job-b", 2, 0, false)
	if _, allowed, reason = second.reserve("movie"); allowed || reason != "Global delivery limit reached" {
		t.Fatalf("global limit = (%t, %q)", allowed, reason)
	}
	first.release(reservationID)
	if _, allowed, reason = second.reserve("movie"); !allowed || reason != "" {
		t.Fatalf("released global slot = (%t, %q)", allowed, reason)
	}

	cfg.Jobs.GlobalLimitMovies = 0
	dryRun := newDeliveryBudget(cfg, db, "job-c", 3, 1, true)
	if _, allowed, _ = dryRun.reserve("movie"); !allowed {
		t.Fatal("first dry-run delivery was blocked")
	}
	if _, allowed, reason = dryRun.reserve("movie"); allowed || reason != "Job delivery limit reached" {
		t.Fatalf("per-run limit = (%t, %q)", allowed, reason)
	}
}

func TestUnlimitedDryRunDoesNotConsumeDeliveryBudget(t *testing.T) {
	db, err := database.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	budget := newDeliveryBudget(&config.Config{}, db, "dry", 1, 0, true)
	if _, allowed, reason := budget.reserve("movie"); !allowed {
		t.Fatalf("dry run blocked: %s", reason)
	}
	if count, err := db.CountDeliveriesSince("movie", time.Now().Add(-time.Hour)); err != nil || count != 0 {
		t.Fatalf("dry run consumed budget: %d, %v", count, err)
	}
}
