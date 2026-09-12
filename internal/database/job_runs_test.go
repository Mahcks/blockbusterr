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

func TestActivityLanguageFiltering(t *testing.T) {
	t.Parallel()
	db, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	for _, language := range []string{"en", "fr", ""} {
		if err := db.LogActivity(ActivityLog{Timestamp: time.Now(), JobType: "test", MediaType: "movie", Title: language, Language: language, Status: "added"}); err != nil {
			t.Fatal(err)
		}
	}
	logs, err := db.GetRecentActivityFiltered(10, "", "", "", "fr")
	if err != nil || len(logs) != 1 || logs[0].Language != "fr" {
		t.Fatalf("filtered logs = %#v, err = %v", logs, err)
	}
	languages, err := db.GetActivityLanguages()
	if err != nil || len(languages) != 2 || languages[0] != "en" || languages[1] != "fr" {
		t.Fatalf("languages = %v, err = %v", languages, err)
	}
}

func TestActivityJobFilteringUsesStableID(t *testing.T) {
	t.Parallel()
	db, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	for _, entry := range []ActivityLog{
		{Timestamp: time.Now(), JobID: "box-office", JobType: "Box Office", MediaType: "movie", Title: "A", Status: "added"},
		{Timestamp: time.Now(), JobID: "smart-popular", JobType: "Smart Popular Movies", MediaType: "movie", Title: "B", Status: "added"},
	} {
		if err := db.LogActivity(entry); err != nil {
			t.Fatal(err)
		}
	}
	logs, err := db.GetRecentActivityFiltered(10, "", "", "box-office", "")
	if err != nil || len(logs) != 1 || logs[0].JobID != "box-office" {
		t.Fatalf("filtered logs = %#v, err = %v", logs, err)
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
		{Timestamp: today, JobType: "test", MediaType: "movie", Title: "D", Status: "failed"},
		{Timestamp: today, JobType: "test", MediaType: "movie", Title: "E", Status: "added", Message: "[DRY RUN] Would be added to Radarr"},
		{Timestamp: today, JobType: "test", MediaType: "movie", Title: "F", Status: "requested", Message: "[DRY RUN] Would be requested via Jellyseerr"},
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

	if counts[todayKey].Added != 1 || counts[todayKey].Requested != 1 {
		t.Fatalf("today delivery = %#v, want 1 added and 1 requested", counts[todayKey])
	}
	if counts[todayKey].WouldAdd != 1 || counts[todayKey].WouldRequest != 1 {
		t.Fatalf("today dry-run delivery = %#v, want 1 would add and 1 would request", counts[todayKey])
	}
	if counts[todayKey].Rejected != 1 {
		t.Fatalf("today rejected = %d, want 1", counts[todayKey].Rejected)
	}
	if counts[todayKey].Skipped != 0 {
		t.Fatalf("today skipped = %d, want 0", counts[todayKey].Skipped)
	}
	if counts[todayKey].Failed != 1 {
		t.Fatalf("today failed = %d, want 1", counts[todayKey].Failed)
	}
	stats, err := db.GetActivityStats()
	if err != nil || stats["total_added"] != 2 {
		t.Fatalf("delivery stats = %#v, error = %v; dry-run deliveries must not count", stats, err)
	}

	if counts[yesterdayKey].Skipped != 1 {
		t.Fatalf("yesterday skipped = %d, want 1", counts[yesterdayKey].Skipped)
	}
}
