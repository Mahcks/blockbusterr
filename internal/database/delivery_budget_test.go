package database

import (
	"testing"
	"time"
)

func TestDeliveryBudgetReservations(t *testing.T) {
	db, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	since := time.Now().Add(-24 * time.Hour)
	first, reason, err := db.TryReserveDelivery(1, "job-a", "movie", 1, 2, since)
	if err != nil || first == 0 || reason != "" {
		t.Fatalf("first reservation = (%d, %q, %v)", first, reason, err)
	}
	if _, reason, err = db.TryReserveDelivery(1, "job-a", "movie", 1, 2, since); err != nil || reason != "Job delivery limit reached" {
		t.Fatalf("per-run limit reason = %q, err = %v", reason, err)
	}
	second, reason, err := db.TryReserveDelivery(2, "job-b", "movie", 0, 2, since)
	if err != nil || second == 0 || reason != "" {
		t.Fatalf("second reservation = (%d, %q, %v)", second, reason, err)
	}
	if _, reason, err = db.TryReserveDelivery(3, "job-c", "movie", 0, 2, since); err != nil || reason != "Global delivery limit reached" {
		t.Fatalf("global limit reason = %q, err = %v", reason, err)
	}
	if err := db.ReleaseDeliveryReservation(second); err != nil {
		t.Fatal(err)
	}
	if third, reason, err := db.TryReserveDelivery(3, "job-c", "movie", 0, 2, since); err != nil || third == 0 || reason != "" {
		t.Fatalf("replacement reservation = (%d, %q, %v)", third, reason, err)
	}
}
