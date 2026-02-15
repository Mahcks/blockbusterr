package database

import (
	"fmt"
	"testing"
	"time"
)

func TestCompleteJobRunRejectsInvalidStatus(t *testing.T) {
	t.Parallel()

	db, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	runID, err := db.StartJobRun("job-1", "Test Job", "movie", "jellyseerr", time.Now())
	if err != nil {
		t.Fatalf("StartJobRun() error = %v", err)
	}

	err = db.CompleteJobRun(
		runID,
		time.Now(),
		"done",
		1, 1, 1, 0, 0, 0, 0,
		"",
	)
	if err == nil {
		t.Fatal("CompleteJobRun() expected error for invalid status, got nil")
	}
}

func TestUpdateActivityLogStatusRejectsInvalidStatus(t *testing.T) {
	t.Parallel()

	db, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	err = db.LogActivity(ActivityLog{
		Timestamp: time.Now(),
		JobType:   "test",
		MediaType: "movie",
		Title:     "Validation Movie",
		Year:      2026,
		Status:    "added",
	})
	if err != nil {
		t.Fatalf("LogActivity() error = %v", err)
	}

	err = db.UpdateActivityLogStatus(1, "unknown_status", "bad")
	if err == nil {
		t.Fatal("UpdateActivityLogStatus() expected validation error for invalid status")
	}
}

func TestGetActivityDailyCountsAggregatesStatuses(t *testing.T) {
	t.Parallel()

	db, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)

	entries := []ActivityLog{
		{Timestamp: today, JobType: "test", MediaType: "movie", Title: "A", Status: "added"},
		{Timestamp: today, JobType: "test", MediaType: "movie", Title: "B", Status: "requested"},
		{Timestamp: today, JobType: "test", MediaType: "movie", Title: "C", Status: "rejected"},
		{Timestamp: yesterday, JobType: "test", MediaType: "show", Title: "D", Status: "skipped"},
	}
	for i, e := range entries {
		e.Title = fmt.Sprintf("%s-%d", e.Title, i)
		if err := db.LogActivity(e); err != nil {
			t.Fatalf("LogActivity() error = %v", err)
		}
	}

	counts, err := db.GetActivityDailyCounts(7)
	if err != nil {
		t.Fatalf("GetActivityDailyCounts() error = %v", err)
	}

	todayKey := today.Format("2006-01-02")
	yesterdayKey := yesterday.Format("2006-01-02")

	if counts[todayKey].Added != 2 {
		t.Fatalf("today added = %d, want 2", counts[todayKey].Added)
	}
	if counts[todayKey].Rejected != 1 {
		t.Fatalf("today rejected = %d, want 1", counts[todayKey].Rejected)
	}
	if counts[todayKey].Skipped != 0 {
		t.Fatalf("today skipped = %d, want 0", counts[todayKey].Skipped)
	}

	if counts[yesterdayKey].Skipped != 1 {
		t.Fatalf("yesterday skipped = %d, want 1", counts[yesterdayKey].Skipped)
	}
}
